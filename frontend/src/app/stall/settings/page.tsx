"use client";

import { LogOutIcon } from "lucide-react";
import { useRouter } from "next/navigation";

import { StallBottomNav } from "@/components/nav/stall-bottom-nav";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { locales, useT } from "@/lib/i18n";
import { clearStallId } from "@/lib/storage";
import { useStallId } from "@/lib/use-stall-id";

export default function StallSettingsPage() {
  // 言語はアプリ全体で持つので、useT() の locale と setLocale を使う
  const { t, locale, setLocale } = useT();
  const router = useRouter();
  // ログインしていなければ、ログインページに移る
  useStallId();

  return (
    <main className="mx-auto flex min-h-svh max-w-md flex-col gap-4 p-4 pb-24">
      <h1 className="text-2xl font-bold">{t.stall.settings.title}</h1>

      <Card>
        <CardHeader>
          <CardTitle>{t.stall.settings.language}</CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-2">
          {locales.map((lang) => (
            <Button
              key={lang}
              variant={lang === locale ? "default" : "outline"}
              aria-pressed={lang === locale}
              onClick={() => setLocale(lang)}
            >
              {t.language[lang]}
            </Button>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t.stall.settings.account}</CardTitle>
        </CardHeader>
        <CardContent>
          <Button
            variant="destructive"
            className="w-full"
            onClick={() => {
              clearStallId();
              router.replace("/stall/login/");
            }}
          >
            <LogOutIcon aria-hidden />
            {t.stall.settings.logout}
          </Button>
        </CardContent>
      </Card>

      <StallBottomNav
        ticketsHref="/stall/tickets"
        statusHref="/stall/status"
        settingsHref="/stall/settings"
        activeHref="/stall/settings"
      />
    </main>
  );
}
