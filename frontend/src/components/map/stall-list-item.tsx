"use client";

import { ExternalLinkIcon } from "lucide-react";

import type { BusinessStatus } from "@/components/map/business-status-selector";
import { format, useT } from "@/lib/i18n";
import { cn } from "@/lib/utils";

export type Congestion = "red" | "yellow" | "green" | "gray";

export type StallListItemProps = {
  /** 店番号（A-01 など） */
  stallNumber: string;
  /** 屋台の名前 */
  name: string;
  /** 混雑度 */
  congestion: Congestion;
  /** 営業状況 */
  businessStatus: BusinessStatus;
  /** 待っている人数（status が waiting の整理券の数） */
  waitingCount: number;
  /** HP のリンク。ない店は null */
  hpLink: string | null;
};

const dotClass: Record<Congestion, string> = {
  red: "bg-red-500",
  yellow: "bg-yellow-400",
  green: "bg-green-500",
  gray: "bg-gray-400",
};

/** マップの屋台1店分。混雑度の色・営業状況・待ち人数・HP リンクを出す。見た目だけを担当する */
export function StallListItem({
  stallNumber,
  name,
  congestion,
  businessStatus,
  waitingCount,
  hpLink,
}: StallListItemProps) {
  const { t } = useT();
  const isOpen = businessStatus === "open";

  return (
    <div className="flex items-start gap-3 rounded-xl border p-3">
      <span
        className={cn("mt-1.5 size-3 shrink-0 rounded-full", dotClass[congestion])}
        aria-label={`${t.congestion.label}：${t.congestion[congestion]}`}
        role="img"
      />
      <div className="min-w-0 flex-1">
        <p className="text-muted-foreground text-xs">{stallNumber}</p>
        <p className="line-clamp-2 font-bold break-words">{name}</p>
        <p className="text-muted-foreground mt-1 text-sm">
          {isOpen
            ? `${t.congestion[congestion]}・${format(t.map.waitingCount, { count: waitingCount })}`
            : t.businessStatus[businessStatus]}
        </p>
        {hpLink && (
          <a
            href={hpLink}
            target="_blank"
            rel="noopener noreferrer"
            className="text-primary mt-1 inline-flex items-center gap-1 text-sm underline-offset-4 hover:underline"
          >
            {t.map.homepage}
            <ExternalLinkIcon className="size-3.5" aria-hidden />
          </a>
        )}
      </div>
    </div>
  );
}
