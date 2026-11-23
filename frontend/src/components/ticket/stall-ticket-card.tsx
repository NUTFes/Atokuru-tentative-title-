"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { TicketStatus } from "@/components/ticket/ticket-card";
import { useT } from "@/lib/i18n";

export type StallTicketCardProps = {
  /** 整理番号 */
  orderNumber: number;
  /** 備考（なければ表示しない） */
  note?: string;
  /** 整理券の状態 */
  status: TicketStatus;
  /** 「調理完了にする」を押したときの処理 */
  onMarkReady: () => void;
  /** 「受け取り済みにする」を押したときの処理 */
  onMarkReceived: () => void;
};

const badgeVariant = {
  waiting: "secondary",
  ready: "default",
  received: "outline",
} as const;

/** 店舗画面の整理券カード。状態を進めるボタンを持つ。見た目だけを担当する */
export function StallTicketCard({
  orderNumber,
  note,
  status,
  onMarkReady,
  onMarkReceived,
}: StallTicketCardProps) {
  const { t } = useT();

  return (
    <Card size="sm">
      <CardHeader>
        <CardTitle className="text-3xl font-bold tabular-nums">{orderNumber}</CardTitle>
        <CardAction>
          <Badge variant={badgeVariant[status]}>{t.ticketStatus[status]}</Badge>
        </CardAction>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        {note && (
          <p className="text-sm">
            <span className="text-muted-foreground mr-2 text-xs">{t.ticket.note}</span>
            {note}
          </p>
        )}
        {status === "waiting" && (
          <Button className="w-full" onClick={onMarkReady}>
            {t.stall.tickets.markReady}
          </Button>
        )}
        {status === "ready" && (
          <Button variant="outline" className="w-full" onClick={onMarkReceived}>
            {t.stall.tickets.markReceived}
          </Button>
        )}
      </CardContent>
    </Card>
  );
}
