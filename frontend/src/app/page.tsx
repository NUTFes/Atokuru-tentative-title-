"use client";

import { useCallback, useEffect, useState } from "react";

import { UserBottomNav } from "@/components/nav/user-bottom-nav";
import { NotificationPrompt } from "@/components/ticket/notification-prompt";
import { TicketCard } from "@/components/ticket/ticket-card";
import { Button } from "@/components/ui/button";
import { api, type TicketDetail } from "@/lib/api/client";
import { type Locale, useT } from "@/lib/i18n";
import {
  isPushSupported,
  notificationPermission,
  requestNotificationPermission,
  subscribePush,
} from "@/lib/push";
import {
  addTicketToken,
  loadPushRegisteredTokens,
  loadTicketTokens,
  removeTicketToken,
  savePushRegisteredToken,
} from "@/lib/storage";
import { useOnVisible } from "@/lib/use-on-visible";

type Notice = "added" | "alreadyAdded" | "notFound" | null;

type LoadState =
  | { kind: "loading" }
  | { kind: "error" }
  | { kind: "loaded"; tickets: TicketDetail[] };

/** 整理券を1枚取る。見つからなければ null、通信に失敗したら throw */
async function fetchTicket(publicToken: string): Promise<TicketDetail | null> {
  const { data, response } = await api.GET("/tickets/{publicToken}", {
    params: { path: { publicToken } },
  });
  if (data) return data;
  if (response.status === 404) return null;
  throw new Error(`GET /tickets/{publicToken} が ${response.status} を返しました`);
}

/** 通知に使う VAPID 公開鍵を取る。通知を使えない（サーバーに鍵がない・端末が対応していない）ときは null */
async function fetchVapidPublicKey(): Promise<string | null> {
  if (!isPushSupported()) return null;
  try {
    const { data } = await api.GET("/push/vapid-public-key");
    return data?.public_key ?? null;
  } catch (error) {
    console.error(error);
    return null;
  }
}

/**
 * まだ登録していない整理券に、通知の購読を登録する。
 * 1つの端末で複数の整理券を持てるので、整理券ごとに登録する。受け取り済みには登録できない。
 */
async function registerPush(vapidPublicKey: string, tickets: TicketDetail[]): Promise<void> {
  const registered = loadPushRegisteredTokens();
  const targets = tickets.filter(
    (ticket) => ticket.status !== "received" && !registered.includes(ticket.public_token),
  );
  if (targets.length === 0) return;
  const subscription = await subscribePush(vapidPublicKey);
  if (!subscription?.endpoint || !subscription.keys?.p256dh || !subscription.keys.auth) return;
  const body = {
    endpoint: subscription.endpoint,
    expirationTime: subscription.expirationTime ?? null,
    keys: { p256dh: subscription.keys.p256dh, auth: subscription.keys.auth },
  };
  for (const ticket of targets) {
    const { response } = await api.POST("/tickets/{publicToken}/push-subscriptions", {
      params: { path: { publicToken: ticket.public_token } },
      body,
    });
    if (response.ok) savePushRegisteredToken(ticket.public_token);
  }
}

/**
 * 整理券のカードに出す備考を選ぶ。
 * 今の言語が日本語以外で訳があれば訳を出し、元の日本語を添える。訳がなければ日本語のまま出す。
 */
function ticketNote(
  ticket: TicketDetail,
  locale: Locale,
): { note?: string; noteOriginal?: string } {
  if (ticket.comment === null) return {};
  const translated = locale === "ja" ? undefined : ticket.comment_translations[locale];
  return translated
    ? { note: translated, noteOriginal: ticket.comment }
    : { note: ticket.comment };
}

export default function Home() {
  const { t, locale } = useT();
  const [state, setState] = useState<LoadState>({ kind: "loading" });
  const [notice, setNotice] = useState<Notice>(null);
  // 通知：使えないときは null で、ボタンを出さない
  const [vapidPublicKey, setVapidPublicKey] = useState<string | null>(null);
  const [permission, setPermission] = useState<NotificationPermission>("default");
  const [enabling, setEnabling] = useState(false);

  useEffect(() => {
    void fetchVapidPublicKey().then((key) => {
      setVapidPublicKey(key);
      if (key) setPermission(notificationPermission());
    });
  }, []);

  // すでに許可されていれば、新しく保存した整理券にも購読を登録する
  useEffect(() => {
    if (!vapidPublicKey || permission !== "granted" || state.kind !== "loaded") return;
    registerPush(vapidPublicKey, state.tickets).catch((error) => console.error(error));
  }, [vapidPublicKey, permission, state]);

  const enableNotification = async () => {
    setEnabling(true);
    try {
      setPermission(await requestNotificationPermission());
    } catch (error) {
      console.error(error);
    }
    setEnabling(false);
  };

  // QR コードの URL（/?token=...）で開いたら、整理券を保存して URL から token を消す
  useEffect(() => {
    const url = new URL(window.location.href);
    const token = url.searchParams.get("token");
    if (!token) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect -- URL はブラウザでしか読めないので、マウント後に反映する
    setNotice(addTicketToken(token) ? "added" : "alreadyAdded");
    url.searchParams.delete("token");
    window.history.replaceState(null, "", url.pathname + url.search);
  }, []);

  const load = useCallback(async () => {
    try {
      const tokens = loadTicketTokens();
      const results = await Promise.all(tokens.map(fetchTicket));
      // 見つからない整理券（消された・token の打ち間違い）は一覧から外す
      const missing = tokens.filter((_, i) => results[i] === null);
      if (missing.length > 0) {
        missing.forEach(removeTicketToken);
        setNotice("notFound");
      }
      setState({
        kind: "loaded",
        tickets: results.filter((ticket): ticket is TicketDetail => ticket !== null),
      });
    } catch (error) {
      console.error(error);
      setState({ kind: "error" });
    }
  }, []);

  // ページを開いたときと、アプリに戻ってきたときに取り直す
  useOnVisible(load);

  return (
    <main className="mx-auto flex min-h-svh max-w-md flex-col gap-4 p-4 pb-24">
      <h1 className="text-2xl font-bold">{t.home.title}</h1>

      {notice && (
        <p role="status" className="bg-muted rounded-lg px-3 py-2 text-sm">
          {t.home[notice]}
        </p>
      )}

      {state.kind === "loading" && <p className="text-muted-foreground">{t.common.loading}</p>}

      {state.kind === "error" && (
        <div role="alert" className="flex flex-col items-start gap-2">
          <p className="text-destructive text-sm">{t.common.networkError}</p>
          <Button variant="outline" onClick={() => void load()}>
            {t.common.retry}
          </Button>
        </div>
      )}

      {vapidPublicKey &&
        state.kind === "loaded" &&
        state.tickets.some((ticket) => ticket.status !== "received") && (
          <NotificationPrompt
            permission={permission}
            onEnable={() => void enableNotification()}
            busy={enabling}
          />
        )}

      {state.kind === "loaded" &&
        (state.tickets.length === 0 ? (
          <div className="text-muted-foreground flex flex-1 flex-col items-center justify-center gap-2 text-center">
            <p className="font-bold">{t.home.empty}</p>
            <p className="text-sm">{t.home.emptyHint}</p>
          </div>
        ) : (
          <ul className="flex flex-col gap-4">
            {state.tickets.map((ticket) => (
              <li key={ticket.public_token}>
                <TicketCard
                  stallName={ticket.stall_name}
                  orderNumber={ticket.order_number}
                  aheadCount={ticket.people_ahead}
                  status={ticket.status}
                  {...ticketNote(ticket, locale)}
                />
              </li>
            ))}
          </ul>
        ))}

      <UserBottomNav homeHref="/" mapHref="/map" activeHref="/" />
    </main>
  );
}
