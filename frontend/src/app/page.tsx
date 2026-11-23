"use client";

import { useCallback, useEffect, useState } from "react";

import { UserBottomNav } from "@/components/nav/user-bottom-nav";
import { TicketCard } from "@/components/ticket/ticket-card";
import { Button } from "@/components/ui/button";
import { api, type TicketDetail } from "@/lib/api/client";
import { useT } from "@/lib/i18n";
import { addTicketToken, loadTicketTokens, removeTicketToken } from "@/lib/storage";
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

export default function Home() {
  const { t } = useT();
  const [state, setState] = useState<LoadState>({ kind: "loading" });
  const [notice, setNotice] = useState<Notice>(null);

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
                />
              </li>
            ))}
          </ul>
        ))}

      <UserBottomNav homeHref="/" mapHref="/map" activeHref="/" />
    </main>
  );
}
