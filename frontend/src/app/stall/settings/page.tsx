"use client";

import { LogOutIcon, PlusIcon, Trash2Icon } from "lucide-react";
import { useRouter } from "next/navigation";
import { useCallback, useState } from "react";

import { StallBottomNav } from "@/components/nav/stall-bottom-nav";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { api, type CommentTemplate } from "@/lib/api/client";
import { format, locales, useT } from "@/lib/i18n";
import { clearStallId } from "@/lib/storage";
import { useOnVisible } from "@/lib/use-on-visible";
import { useStallId } from "@/lib/use-stall-id";

/** 定型文の長さの上限。整理券の備考と同じ（openapi.yaml） */
const TEMPLATE_MAX_LENGTH = 200;

type TemplatesState =
  | { kind: "loading" }
  | { kind: "error" }
  | { kind: "loaded"; templates: CommentTemplate[] };

export default function StallSettingsPage() {
  // 言語はアプリ全体で持つので、useT() の locale と setLocale を使う
  const { t, locale, setLocale } = useT();
  const router = useRouter();
  // ログインしていなければ、ログインページに移る
  const stallId = useStallId();

  const [templates, setTemplates] = useState<TemplatesState>({ kind: "loading" });
  const [draft, setDraft] = useState("");
  // 登録は翻訳に数秒かかるので、そのあいだはボタンを押せないようにする
  const [adding, setAdding] = useState(false);
  const [deletingId, setDeletingId] = useState<number | null>(null);
  // 409 などで返ってきた message はそのまま出す。通信の失敗は null 以外の固定文
  const [templateError, setTemplateError] = useState<string | null>(null);

  const loadTemplates = useCallback(async () => {
    if (stallId === null) return;
    try {
      const { data } = await api.GET("/stalls/{stallId}/comment-templates", {
        params: { path: { stallId } },
      });
      setTemplates(data ? { kind: "loaded", templates: data } : { kind: "error" });
    } catch (e) {
      console.error(e);
      setTemplates({ kind: "error" });
    }
  }, [stallId]);

  // ページを開いたときと、アプリに戻ってきたときに取り直す
  useOnVisible(loadTemplates);

  const addTemplate = async () => {
    const comment = draft.trim();
    if (stallId === null || comment === "") return;
    setAdding(true);
    setTemplateError(null);
    try {
      const { data, error } = await api.POST("/stalls/{stallId}/comment-templates", {
        params: { path: { stallId } },
        body: { comment },
      });
      if (data) {
        setDraft("");
        // 登録した順に並ぶので、最後に足す
        setTemplates((prev) =>
          prev.kind === "loaded" ? { kind: "loaded", templates: [...prev.templates, data] } : prev,
        );
      } else {
        // 409（同じ定型文がある・30件を超える）などは、サーバーの message をそのまま出す
        setTemplateError(error?.message ?? t.common.error);
      }
    } catch (e) {
      console.error(e);
      setTemplateError(t.common.networkError);
    }
    setAdding(false);
  };

  const deleteTemplate = async (templateId: number) => {
    setDeletingId(templateId);
    setTemplateError(null);
    try {
      const { response, error } = await api.DELETE("/comment-templates/{templateId}", {
        params: { path: { templateId } },
      });
      if (response.ok) {
        setTemplates((prev) =>
          prev.kind === "loaded"
            ? { kind: "loaded", templates: prev.templates.filter((x) => x.id !== templateId) }
            : prev,
        );
      } else {
        setTemplateError(error?.message ?? t.common.error);
      }
    } catch (e) {
      console.error(e);
      setTemplateError(t.common.networkError);
    }
    setDeletingId(null);
  };

  const ts = t.stall.settings.templates;

  return (
    <main className="mx-auto flex min-h-svh max-w-md flex-col gap-4 p-4 pb-24">
      <h1 className="text-2xl font-bold">{t.stall.settings.title}</h1>

      <Card>
        <CardHeader>
          <CardTitle>{ts.title}</CardTitle>
          <CardDescription>{ts.description}</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          <form
            className="flex gap-2"
            onSubmit={(e) => {
              e.preventDefault();
              void addTemplate();
            }}
          >
            <Input
              value={draft}
              maxLength={TEMPLATE_MAX_LENGTH}
              placeholder={ts.placeholder}
              aria-label={ts.title}
              disabled={adding}
              onChange={(e) => setDraft(e.target.value)}
            />
            <Button type="submit" disabled={adding || stallId === null || draft.trim() === ""}>
              <PlusIcon aria-hidden />
              {adding ? ts.adding : ts.add}
            </Button>
          </form>

          {templateError && (
            <p role="alert" className="text-destructive text-sm">
              {templateError}
            </p>
          )}

          {templates.kind === "loading" && (
            <p className="text-muted-foreground text-sm">{t.common.loading}</p>
          )}

          {templates.kind === "error" && (
            <div role="alert" className="flex flex-col items-start gap-2">
              <p className="text-destructive text-sm">{t.common.networkError}</p>
              <Button variant="outline" size="sm" onClick={() => void loadTemplates()}>
                {t.common.retry}
              </Button>
            </div>
          )}

          {templates.kind === "loaded" &&
            (templates.templates.length === 0 ? (
              <p className="text-muted-foreground text-sm">{ts.empty}</p>
            ) : (
              <ul className="divide-y">
                {templates.templates.map((template) => (
                  <li key={template.id} className="flex items-center gap-2 py-2">
                    <div className="min-w-0 flex-1">
                      <p className="break-words">{template.comment}</p>
                      {/* 訳が変なときに店が気づけるように、英語の訳を小さく出す */}
                      {template.translations.en && (
                        <p className="text-muted-foreground text-xs break-words">
                          {format(ts.translation, { text: template.translations.en })}
                        </p>
                      )}
                    </div>
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label={format(ts.deleteLabel, { comment: template.comment })}
                      disabled={deletingId === template.id}
                      onClick={() => void deleteTemplate(template.id)}
                    >
                      <Trash2Icon aria-hidden />
                    </Button>
                  </li>
                ))}
              </ul>
            ))}
        </CardContent>
      </Card>

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
