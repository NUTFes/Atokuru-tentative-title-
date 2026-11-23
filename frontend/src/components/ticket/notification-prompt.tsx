"use client";

import { BellIcon, BellOffIcon, BellRingIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useT } from "@/lib/i18n";

export type NotificationPromptProps = {
  /** 通知の許可の状態 */
  permission: NotificationPermission;
  /** 「通知で受け取る」を押したときの処理。許可を求めるので、必ずボタンから呼ぶ */
  onEnable: () => void;
  /** 許可を求めている途中か */
  busy?: boolean;
};

/** 整理券の呼び出しを通知で受け取るためのボタンと案内。見た目だけを担当する */
export function NotificationPrompt({ permission, onEnable, busy = false }: NotificationPromptProps) {
  const { t } = useT();

  if (permission === "granted") {
    return (
      <p className="text-muted-foreground flex items-center gap-2 text-sm">
        <BellRingIcon className="size-4" aria-hidden />
        {t.notification.enabled}
      </p>
    );
  }

  if (permission === "denied") {
    return (
      <p className="text-muted-foreground flex items-center gap-2 text-sm">
        <BellOffIcon className="size-4" aria-hidden />
        {t.notification.denied}
      </p>
    );
  }

  return (
    <Button variant="outline" className="w-full" onClick={onEnable} disabled={busy}>
      <BellIcon aria-hidden />
      {t.notification.enable}
    </Button>
  );
}
