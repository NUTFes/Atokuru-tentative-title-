"use client";

import { useCallback, useState } from "react";

import { StallListItem } from "@/components/map/stall-list-item";
import { VenueMap } from "@/components/map/venue-map";
import { UserBottomNav } from "@/components/nav/user-bottom-nav";
import { Button } from "@/components/ui/button";
import { api, type StallSummary } from "@/lib/api/client";
import { useT } from "@/lib/i18n";
import { useOnVisible } from "@/lib/use-on-visible";

type LoadState =
  | { kind: "loading" }
  | { kind: "error" }
  | { kind: "loaded"; stalls: StallSummary[] };

export default function MapPage() {
  const { t } = useT();
  const [state, setState] = useState<LoadState>({ kind: "loading" });

  const load = useCallback(async () => {
    try {
      const { data } = await api.GET("/stalls");
      setState(data ? { kind: "loaded", stalls: data } : { kind: "error" });
    } catch (error) {
      console.error(error);
      setState({ kind: "error" });
    }
  }, []);

  // ページを開いたときと、アプリに戻ってきたときに取り直す
  useOnVisible(load);

  return (
    <main className="mx-auto flex min-h-svh max-w-md flex-col gap-4 p-4 pb-24">
      <h1 className="text-2xl font-bold">{t.map.title}</h1>

      <VenueMap src="/map.png" alt={t.map.venueAlt} width={1200} height={864} hint={t.map.zoomHint} />

      {state.kind === "loading" && <p className="text-muted-foreground">{t.common.loading}</p>}

      {state.kind === "error" && (
        <div role="alert" className="flex flex-col items-start gap-2">
          <p className="text-destructive text-sm">{t.common.networkError}</p>
          <Button variant="outline" onClick={() => void load()}>
            {t.common.retry}
          </Button>
        </div>
      )}

      {state.kind === "loaded" && (
        <ul className="flex flex-col gap-2">
          {state.stalls.map((stall) => (
            <li key={stall.id}>
              <StallListItem
                stallNumber={stall.stall_number}
                name={stall.name}
                congestion={stall.congestion}
                businessStatus={stall.business_status}
                waitingCount={stall.waiting_count}
                hpLink={stall.hp_link}
              />
            </li>
          ))}
        </ul>
      )}

      <UserBottomNav homeHref="/" mapHref="/map" activeHref="/map" />
    </main>
  );
}
