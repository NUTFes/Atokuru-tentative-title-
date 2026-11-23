import { StallBottomNav } from "@/components/nav/stall-bottom-nav";

// #17 の確認用。ページの中身は別の Issue で作る
export default function StallStatusPage() {
  return (
    <main className="min-h-svh pb-20">
      <StallBottomNav
        ticketsHref="/stall/tickets"
        statusHref="/stall/status"
        settingsHref="/stall/settings"
        activeHref="/stall/status"
      />
    </main>
  );
}
