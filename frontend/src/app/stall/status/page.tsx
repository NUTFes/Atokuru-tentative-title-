"use client";

import { useCallback, useState } from "react";

import { StallBottomNav } from "@/components/nav/stall-bottom-nav";
import {
  BusinessStatusSelector,
  type BusinessStatus,
} from "@/components/map/business-status-selector";
import { Button } from "@/components/ui/button";
import { api } from "@/lib/api/client";
import { useT } from "@/lib/i18n";
import { useOnVisible } from "@/lib/use-on-visible";
import { useStallId } from "@/lib/use-stall-id";

type Message = "updated" | "error" | null;

export default function StallStatusPage() {
  const { t } = useT();
  const stallId = useStallId();
  const [status, setStatus] = useState<BusinessStatus | null>(null);
  const [loadFailed, setLoadFailed] = useState(false);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<Message>(null);

  const load = useCallback(async () => {
    if (stallId === null) return;
    try {
      const { data } = await api.GET("/stalls/{stallId}", { params: { path: { stallId } } });
      if (data) {
        setStatus(data.business_status);
        setLoadFailed(false);
      } else {
        setLoadFailed(true);
      }
    } catch (e) {
      console.error(e);
      setLoadFailed(true);
    }
  }, [stallId]);

  // ページを開いたときと、アプリに戻ってきたときに取り直す
  useOnVisible(load);

  const change = async (next: BusinessStatus) => {
    if (stallId === null || next === status) return;
    setSaving(true);
    setMessage(null);
    try {
      const { data } = await api.PATCH("/stalls/{stallId}/business-status", {
        params: { path: { stallId } },
        body: { business_status: next },
      });
      if (data) {
        setStatus(data.business_status);
        setMessage("updated");
      } else {
        setMessage("error");
      }
    } catch (e) {
      console.error(e);
      setMessage("error");
    }
    setSaving(false);
  };

  return (
    <main className="mx-auto flex min-h-svh max-w-md flex-col gap-4 p-4 pb-24">
      <div>
        <h1 className="text-2xl font-bold">{t.stall.status.title}</h1>
        <p className="text-muted-foreground text-sm">{t.stall.status.description}</p>
      </div>

      {message && (
        <p
          role={message === "error" ? "alert" : "status"}
          className={message === "error" ? "text-destructive text-sm" : "bg-muted rounded-lg px-3 py-2 text-sm"}
        >
          {message === "error" ? t.common.networkError : t.stall.status.updated}
        </p>
      )}

      {loadFailed && status === null ? (
        <div role="alert" className="flex flex-col items-start gap-2">
          <p className="text-destructive text-sm">{t.common.networkError}</p>
          <Button variant="outline" onClick={() => void load()}>
            {t.common.retry}
          </Button>
        </div>
      ) : status === null ? (
        <p className="text-muted-foreground">{t.common.loading}</p>
      ) : (
        <BusinessStatusSelector value={status} onChange={(next) => void change(next)} disabled={saving} />
      )}

      <StallBottomNav
        ticketsHref="/stall/tickets"
        statusHref="/stall/status"
        settingsHref="/stall/settings"
        activeHref="/stall/status"
      />
    </main>
  );
}
