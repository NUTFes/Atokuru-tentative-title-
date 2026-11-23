"use client";

import { UserBottomNav } from "@/components/nav/user-bottom-nav";
import { TicketCard, type TicketCardProps } from "@/components/ticket/ticket-card";
import { useT } from "@/lib/i18n";

type HomeTicket = Omit<TicketCardProps, "onAction"> & { publicToken: string };

// API とローカルストレージとの繋ぎこみはまだしないので、固定値を表示する
const tickets: HomeTicket[] = [
  { publicToken: "sample-1", stallName: "たこ焼き屋", orderNumber: 12, aheadCount: 3, status: "waiting" },
  { publicToken: "sample-2", stallName: "焼きそば屋", orderNumber: 5, status: "ready" },
  { publicToken: "sample-3", stallName: "クレープ屋", orderNumber: 2, status: "received" },
];

export default function Home() {
  const { t } = useT();

  return (
    <main className="mx-auto flex min-h-svh max-w-md flex-col gap-4 p-4 pb-24">
      <h1 className="text-2xl font-bold">{t.home.title}</h1>
      {tickets.length === 0 ? (
        <div className="text-muted-foreground flex flex-1 flex-col items-center justify-center gap-2 text-center">
          <p className="font-bold">{t.home.empty}</p>
          <p className="text-sm">{t.home.emptyHint}</p>
        </div>
      ) : (
        <ul className="flex flex-col gap-4">
          {tickets.map(({ publicToken, ...ticket }) => (
            <li key={publicToken}>
              <TicketCard {...ticket} onAction={() => console.log("open", publicToken)} />
            </li>
          ))}
        </ul>
      )}
      <UserBottomNav homeHref="/" mapHref="/map" activeHref="/" />
    </main>
  );
}
