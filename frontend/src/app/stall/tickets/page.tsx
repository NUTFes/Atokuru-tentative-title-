"use client";

import { PlusIcon } from "lucide-react";
import { useCallback, useState } from "react";

import { StallBottomNav } from "@/components/nav/stall-bottom-nav";
import { IssueTicketSheet } from "@/components/sheet/issue-ticket-sheet";
import { IssuedTicketModal } from "@/components/sheet/issued-ticket-modal";
import { StallTicketCard } from "@/components/ticket/stall-ticket-card";
import { Button } from "@/components/ui/button";
import { api, type Ticket } from "@/lib/api/client";
import { format, useT } from "@/lib/i18n";
import { useOnVisible } from "@/lib/use-on-visible";
import { useStallId } from "@/lib/use-stall-id";

type LoadState = { kind: "loading" } | { kind: "error" } | { kind: "loaded"; tickets: Ticket[] };

type ErrorKind = "notOpen" | "network" | null;

/** 来場者が QR コードで開く URL。id は連番で推測されやすいので、public_token を使う */
function ticketUrl(publicToken: string): string {
  return `${window.location.origin}/?token=${encodeURIComponent(publicToken)}`;
}

export default function StallTicketsPage() {
  const { t } = useT();
  const stallId = useStallId();
  const [state, setState] = useState<LoadState>({ kind: "loading" });
  const [error, setError] = useState<ErrorKind>(null);
  const [sheetOpen, setSheetOpen] = useState(false);
  const [issuing, setIssuing] = useState(false);
  const [issued, setIssued] = useState<Ticket | null>(null);
  const [templates, setTemplates] = useState<string[]>([]);

  const load = useCallback(async () => {
    if (stallId === null) return;
    try {
      // 整理券は10万件規模になるので、待っているもの（調理中・調理完了）だけを取る
      const { data } = await api.GET("/stalls/{stallId}/tickets", {
        params: { path: { stallId }, query: { status: ["waiting", "ready"] } },
      });
      setState(data ? { kind: "loaded", tickets: data } : { kind: "error" });
    } catch (e) {
      console.error(e);
      setState({ kind: "error" });
    }
  }, [stallId]);

  // 発券のシートに出す備考の定型文。取れなくても発券はできるので、エラーは出さない
  const loadTemplates = useCallback(async () => {
    if (stallId === null) return;
    try {
      const { data } = await api.GET("/stalls/{stallId}/comment-templates", {
        params: { path: { stallId } },
      });
      if (data) setTemplates(data.map((template) => template.comment));
    } catch (e) {
      console.error(e);
    }
  }, [stallId]);

  // ページを開いたとき（ログイン中の店の ID を読んだあと）と、アプリに戻ってきたときに取り直す
  useOnVisible(load);
  useOnVisible(loadTemplates);

  const issue = async (note: string) => {
    if (stallId === null) return;
    setIssuing(true);
    setError(null);
    try {
      const { data, response } = await api.POST("/stalls/{stallId}/tickets", {
        params: { path: { stallId } },
        body: { comment: note === "" ? null : note },
      });
      if (data) {
        // シートを閉じて（備考はここで消える）、QR コードのモーダルを出す
        setSheetOpen(false);
        setIssued(data);
        void load();
      } else {
        setError(response.status === 409 ? "notOpen" : "network");
      }
    } catch (e) {
      console.error(e);
      setError("network");
    }
    setIssuing(false);
  };

  const changeStatus = async (ticketId: number, status: Ticket["status"]) => {
    setError(null);
    try {
      const { data } = await api.PATCH("/tickets/{ticketId}/status", {
        params: { path: { ticketId } },
        body: { status },
      });
      if (!data) setError("network");
    } catch (e) {
      console.error(e);
      setError("network");
    }
    void load();
  };

  const waitingCount =
    state.kind === "loaded" ? state.tickets.filter((ticket) => ticket.status === "waiting").length : 0;

  return (
    <main className="mx-auto flex min-h-svh max-w-md flex-col gap-4 p-4 pb-[calc(10rem+env(safe-area-inset-bottom))]">
      <div className="flex items-baseline justify-between">
        <h1 className="text-2xl font-bold">{t.stall.tickets.title}</h1>
        {state.kind === "loaded" && (
          <p className="text-muted-foreground text-sm">
            {format(t.stall.tickets.waitingCount, { count: waitingCount })}
          </p>
        )}
      </div>

      {error && (
        <p role="alert" className="text-destructive text-sm">
          {error === "notOpen" ? t.stall.tickets.notOpen : t.common.networkError}
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
          <p className="text-muted-foreground flex flex-1 items-center justify-center text-center">
            {t.stall.tickets.empty}
          </p>
        ) : (
          <ul className="grid grid-cols-2 gap-3">
            {state.tickets.map((ticket) => (
              <li key={ticket.id}>
                <StallTicketCard
                  orderNumber={ticket.order_number}
                  note={ticket.comment ?? undefined}
                  status={ticket.status}
                  onMarkReady={() => void changeStatus(ticket.id, "ready")}
                  onMarkReceived={() => void changeStatus(ticket.id, "received")}
                />
              </li>
            ))}
          </ul>
        ))}

      <div className="fixed inset-x-0 bottom-[calc(5rem+env(safe-area-inset-bottom))] z-30 mx-auto max-w-md px-4">
        <Button
          size="lg"
          className="w-full shadow-lg"
          disabled={stallId === null}
          onClick={() => {
            setError(null);
            setSheetOpen(true);
          }}
        >
          <PlusIcon aria-hidden />
          {t.stall.tickets.issue}
        </Button>
      </div>

      <IssueTicketSheet
        open={sheetOpen}
        onOpenChange={setSheetOpen}
        onIssue={(note) => void issue(note)}
        issuing={issuing}
        templates={templates}
      />
      {issued && (
        <IssuedTicketModal
          open
          onOpenChange={(open) => {
            if (!open) setIssued(null);
          }}
          orderNumber={issued.order_number}
          qrValue={ticketUrl(issued.public_token)}
        />
      )}
      <StallBottomNav
        ticketsHref="/stall/tickets"
        statusHref="/stall/status"
        settingsHref="/stall/settings"
        activeHref="/stall/tickets"
      />
    </main>
  );
}
