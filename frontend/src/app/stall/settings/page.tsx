"use client";

import { LogOutIcon } from "lucide-react";
import { useRouter } from "next/navigation";
import { useCallback, useState } from "react";

import { StallBottomNav } from "@/components/nav/stall-bottom-nav";
import { CommentTemplateEditor } from "@/components/sheet/comment-template-editor";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { api, type CommentTemplate } from "@/lib/api/client";
import { locales, useT } from "@/lib/i18n";
import { clearStallId } from "@/lib/storage";
import { useOnVisible } from "@/lib/use-on-visible";
import { useStallId } from "@/lib/use-stall-id";

export default function StallSettingsPage() {
  // 言語はアプリ全体で持つので、useT() の locale と setLocale を使う
  const { t, locale, setLocale } = useT();
  const router = useRouter();
  // ログインしていなければ、ログインページに移る
  const stallId = useStallId();
  const [templates, setTemplates] = useState<CommentTemplate[]>([]);
  const [busy, setBusy] = useState(false);
  const [templateError, setTemplateError] = useState<string | null>(null);

  const loadTemplates = useCallback(async () => {
    if (stallId === null) return;
    try {
      const { data } = await api.GET("/stalls/{stallId}/comment-templates", {
        params: { path: { stallId } },
      });
      if (data) setTemplates(data);
    } catch (e) {
      console.error(e);
    }
  }, [stallId]);

  // ページを開いたときと、アプリに戻ってきたときに取り直す
  useOnVisible(loadTemplates);

  const addTemplate = async (comment: string): Promise<boolean> => {
    if (stallId === null) return false;
    setBusy(true);
    setTemplateError(null);
    let added = false;
    try {
      const { data, response } = await api.POST("/stalls/{stallId}/comment-templates", {
        params: { path: { stallId } },
        body: { comment },
      });
      if (data) {
        setTemplates((current) => [...current, data]);
        added = true;
      } else {
        setTemplateError(response.status === 409 ? t.stall.templates.conflict : t.common.error);
      }
    } catch (e) {
      console.error(e);
      setTemplateError(t.common.networkError);
    }
    setBusy(false);
    return added;
  };

  const deleteTemplate = async (templateId: number) => {
    setBusy(true);
    setTemplateError(null);
    try {
      const { response } = await api.DELETE("/comment-templates/{templateId}", {
        params: { path: { templateId } },
      });
      // 404 はもう消えているので、一覧から外してよい
      if (response.ok || response.status === 404) {
        setTemplates((current) => current.filter((template) => template.id !== templateId));
      } else {
        setTemplateError(t.common.error);
      }
    } catch (e) {
      console.error(e);
      setTemplateError(t.common.networkError);
    }
    setBusy(false);
  };

  return (
    <main className="mx-auto flex min-h-svh max-w-md flex-col gap-4 p-4 pb-24">
      <h1 className="text-2xl font-bold">{t.stall.settings.title}</h1>

      <CommentTemplateEditor
        templates={templates}
        onAdd={addTemplate}
        onDelete={(id) => void deleteTemplate(id)}
        busy={busy || stallId === null}
        error={templateError}
      />

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
