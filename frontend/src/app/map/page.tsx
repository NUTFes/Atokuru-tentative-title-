import { UserBottomNav } from "@/components/nav/user-bottom-nav";

// #16 の確認用。マップの中身は別の Issue で作る
export default function MapPage() {
  return (
    <main className="min-h-svh pb-20">
      <UserBottomNav homeHref="/" mapHref="/map" activeHref="/map" />
    </main>
  );
}
