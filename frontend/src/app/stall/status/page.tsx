"use client";

import { useState } from "react";

import { StallBottomNav } from "@/components/nav/stall-bottom-nav";
import {
  BusinessStatusSelector,
  type BusinessStatus,
} from "@/components/map/business-status-selector";
import { useT } from "@/lib/i18n";

export default function StallStatusPage() {
  const { t } = useT();
  // API との繋ぎこみはまだしないので、固定値から始めて画面の中だけで切り替える
  const [status, setStatus] = useState<BusinessStatus>("open");

  return (
    <main className="mx-auto flex min-h-svh max-w-md flex-col gap-4 p-4 pb-24">
      <div>
        <h1 className="text-2xl font-bold">{t.stall.status.title}</h1>
        <p className="text-muted-foreground text-sm">{t.stall.status.description}</p>
      </div>
      <BusinessStatusSelector value={status} onChange={setStatus} />
      <StallBottomNav
        ticketsHref="/stall/tickets"
        statusHref="/stall/status"
        settingsHref="/stall/settings"
        activeHref="/stall/status"
      />
    </main>
  );
}
