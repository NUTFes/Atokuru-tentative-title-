"use client";

import { PlusIcon } from "lucide-react";
import { useState } from "react";

import { StallBottomNav } from "@/components/nav/stall-bottom-nav";
import { IssueTicketSheet } from "@/components/sheet/issue-ticket-sheet";
import { IssuedTicketModal } from "@/components/sheet/issued-ticket-modal";
import { StallTicketCard, type StallTicketCardProps } from "@/components/ticket/stall-ticket-card";
import { Button } from "@/components/ui/button";
import { useT } from "@/lib/i18n";

type StallTicket = Pick<StallTicketCardProps, "orderNumber" | "note" | "status"> & {
  publicToken: string;
};

// API との繋ぎこみはまだしないので、固定値を表示する
const tickets: StallTicket[] = [
  { publicToken: "sample-10", orderNumber: 10, status: "ready" },
  { publicToken: "sample-11", orderNumber: 11, status: "waiting", note: "ソース抜き" },
  { publicToken: "sample-12", orderNumber: 12, status: "waiting" },
];

export default function StallTicketsPage() {
  const { t } = useT();
  const [sheetOpen, setSheetOpen] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);

  return (
    <main className="mx-auto flex min-h-svh max-w-md flex-col gap-4 p-4 pb-[calc(10rem+env(safe-area-inset-bottom))]">
      <h1 className="text-2xl font-bold">{t.stall.tickets.title}</h1>
      {tickets.length === 0 ? (
        <p className="text-muted-foreground flex flex-1 items-center justify-center text-center">
          {t.stall.tickets.empty}
        </p>
      ) : (
        <ul className="grid grid-cols-2 gap-3">
          {tickets.map(({ publicToken, ...ticket }) => (
            <li key={publicToken}>
              <StallTicketCard
                {...ticket}
                onMarkReady={() => console.log("ready", publicToken)}
                onMarkReceived={() => console.log("received", publicToken)}
              />
            </li>
          ))}
        </ul>
      )}

      <div className="fixed inset-x-0 bottom-[calc(5rem+env(safe-area-inset-bottom))] z-30 mx-auto max-w-md px-4">
        <Button size="lg" className="w-full shadow-lg" onClick={() => setSheetOpen(true)}>
          <PlusIcon aria-hidden />
          {t.stall.tickets.issue}
        </Button>
      </div>

      <IssueTicketSheet
        open={sheetOpen}
        onOpenChange={setSheetOpen}
        onIssue={(note) => {
          console.log("issue", note);
          setSheetOpen(false);
          setModalOpen(true);
        }}
      />
      <IssuedTicketModal
        open={modalOpen}
        onOpenChange={setModalOpen}
        orderNumber={13}
        qrValue="https://example.com/ticket?token=sample-public-token"
      />
      <StallBottomNav
        ticketsHref="/stall/tickets"
        statusHref="/stall/status"
        settingsHref="/stall/settings"
        activeHref="/stall/tickets"
      />
    </main>
  );
}
