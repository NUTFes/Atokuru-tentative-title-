"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { api } from "@/lib/api/client";
import { useT } from "@/lib/i18n";
import { saveStallId } from "@/lib/storage";

type ErrorKind = "notFound" | "network" | null;

export default function StallLoginPage() {
  const { t } = useT();
  const router = useRouter();
  const [stallId, setStallId] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<ErrorKind>(null);

  const id = Number(stallId);
  const valid = Number.isInteger(id) && id > 0;

  const submit = async () => {
    if (!valid) return;
    setSubmitting(true);
    setError(null);
    try {
      // パスワードは使わず、店が存在するかだけを確かめる（#47）
      const { data, response } = await api.GET("/stalls/{stallId}", {
        params: { path: { stallId: id } },
      });
      if (data) {
        saveStallId(data.id);
        router.push("/stall/tickets/");
        return;
      }
      setError(response.status === 404 || response.status === 400 ? "notFound" : "network");
    } catch (e) {
      console.error(e);
      setError("network");
    }
    setSubmitting(false);
  };

  return (
    <main className="mx-auto flex min-h-svh max-w-md flex-col justify-center gap-6 p-4">
      <div>
        <h1 className="text-2xl font-bold">{t.stall.login.title}</h1>
        <p className="text-muted-foreground text-sm">{t.stall.login.description}</p>
      </div>
      <form
        className="flex flex-col gap-4"
        onSubmit={(e) => {
          e.preventDefault();
          void submit();
        }}
      >
        <div className="flex flex-col gap-2">
          <Label htmlFor="stall-id">{t.stall.login.stallId}</Label>
          <Input
            id="stall-id"
            inputMode="numeric"
            autoComplete="off"
            value={stallId}
            onChange={(e) => setStallId(e.target.value.replace(/\D/g, ""))}
            className="h-11 text-lg"
          />
        </div>
        {error && (
          <p role="alert" className="text-destructive text-sm">
            {error === "notFound" ? t.stall.login.notFound : t.common.networkError}
          </p>
        )}
        <Button type="submit" size="lg" disabled={!valid || submitting}>
          {t.stall.login.submit}
        </Button>
      </form>
    </main>
  );
}
