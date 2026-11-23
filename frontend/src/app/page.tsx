"use client";

import { TicketCard } from "@/components/ticket/ticket-card";

// #10 の確認用。固定値を渡してカードを表示する（#20 でホーム画面を作るときに置き換える）
export default function Home() {
  return (
    <main className="mx-auto flex min-h-svh max-w-md flex-col gap-4 p-4">
      <TicketCard
        stallName="たこ焼き屋"
        orderNumber={12}
        aheadCount={3}
        status="waiting"
        onAction={() => console.log("action")}
      />
      <TicketCard stallName="焼きそば屋" orderNumber={5} status="ready" />
      <TicketCard stallName="クレープ屋" orderNumber={2} status="received" />
    </main>
  );
}
