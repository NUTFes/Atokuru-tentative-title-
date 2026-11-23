"use client";

import { useState } from "react";

import { IssueTicketSheet } from "@/components/sheet/issue-ticket-sheet";
import { IssuedTicketModal } from "@/components/sheet/issued-ticket-modal";
import { Button } from "@/components/ui/button";
import { useT } from "@/lib/i18n";

// #14 の確認用。固定値でボトムシートとモーダルを表示する（#23 で整理券一覧ページを作るときに置き換える）
export default function StallTicketsPage() {
  const { t } = useT();
  const [sheetOpen, setSheetOpen] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);

  return (
    <main className="mx-auto flex min-h-svh max-w-md flex-col gap-4 p-4">
      <Button onClick={() => setSheetOpen(true)}>{t.stall.tickets.issue}</Button>
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
    </main>
  );
}
