"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { format, useT } from "@/lib/i18n";

export type TicketStatus = "waiting" | "ready" | "received";

export type TicketCardProps = {
  /** 屋台の名前 */
  stallName: string;
  /** 整理番号 */
  orderNumber: number;
  /** 自分の前に何人いるか。waiting のときだけ表示するので、それ以外では省略できる */
  aheadCount?: number;
  /** 整理券の状態 */
  status: TicketStatus;
  /** ボタンを押したときの処理。渡さなければボタンを出さない */
  onAction?: () => void;
  /** 備考（メニューなど）。訳があれば訳したものを渡す。なければ表示しない */
  note?: string | null;
};

const badgeVariant = {
  waiting: "secondary",
  ready: "default",
  received: "outline",
} as const;

/**
 * 整理券のカード。見た目だけを担当し、データと callback は props で受け取る。
 * ほかのコンポーネントを作るときの手本。
 */
export function TicketCard({
  stallName,
  orderNumber,
  aheadCount = 0,
  status,
  onAction,
  note,
}: TicketCardProps) {
  const { t } = useT();

  return (
    <Card className={status === "received" ? "opacity-60" : undefined}>
      <CardHeader>
        <CardTitle>{stallName}</CardTitle>
        <CardAction>
          <Badge variant={badgeVariant[status]}>{t.ticketStatus[status]}</Badge>
        </CardAction>
      </CardHeader>
      <CardContent className="flex items-end justify-between gap-4">
        <div>
          <p className="text-muted-foreground text-xs">{t.ticket.orderNumber}</p>
          <p className="text-4xl font-bold tabular-nums">{orderNumber}</p>
        </div>
        <p className="text-sm">
          {status === "waiting" && format(t.ticket.peopleAhead, { count: aheadCount })}
          {status === "ready" && t.ticket.readyMessage}
          {status === "received" && t.ticket.receivedMessage}
        </p>
      </CardContent>
      {note && (
        <CardContent className="text-sm">
          <span className="text-muted-foreground mr-2 text-xs">{t.ticket.note}</span>
          <span className="break-words">{note}</span>
        </CardContent>
      )}
      {onAction && (
        <CardFooter>
          <Button variant="outline" className="w-full" onClick={onAction}>
            {t.ticket.detail}
          </Button>
        </CardFooter>
      )}
    </Card>
  );
}
