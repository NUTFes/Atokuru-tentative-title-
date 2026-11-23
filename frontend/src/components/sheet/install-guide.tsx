"use client";

import { ShareIcon, XIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useT } from "@/lib/i18n";

export type InstallGuideProps = {
  /** 案内を出すか */
  open: boolean;
  /** 閉じるボタンを押したときの処理 */
  onDismiss: () => void;
};

/** iPhone で「ホーム画面に追加」を案内する。見た目だけを担当する */
export function InstallGuide({ open, onDismiss }: InstallGuideProps) {
  const { t } = useT();
  if (!open) return null;

  // 文言の {share} の位置に共有ボタンのアイコンを入れる
  const [before, after = ""] = t.pwa.installBody.split("{share}");

  return (
    <div
      role="dialog"
      aria-labelledby="install-guide-title"
      className="bg-background fixed inset-x-0 top-0 z-50 mx-auto max-w-md p-4 pt-[calc(1rem+env(safe-area-inset-top))]"
    >
      <div className="flex gap-3 rounded-xl border p-4 shadow-lg">
        <div className="flex-1 text-sm">
          <p id="install-guide-title" className="mb-1 font-bold">
            {t.pwa.installTitle}
          </p>
          <p className="text-muted-foreground">
            {before}
            <ShareIcon className="inline size-4 align-text-bottom" aria-hidden />
            {after}
          </p>
        </div>
        <Button variant="ghost" size="icon" onClick={onDismiss} aria-label={t.pwa.dismiss}>
          <XIcon aria-hidden />
        </Button>
      </div>
    </div>
  );
}
