"use client";

import { CheckIcon } from "lucide-react";

import { cn } from "@/lib/utils";
import { useT } from "@/lib/i18n";

export type BusinessStatus = "open" | "preparing" | "sold_out" | "closed";

export type BusinessStatusSelectorProps = {
  /** 今の営業状況 */
  value: BusinessStatus;
  /** 営業状況を選んだときの処理 */
  onChange: (value: BusinessStatus) => void;
  /** 変更中か（true のあいだは選べない） */
  disabled?: boolean;
};

const statuses: BusinessStatus[] = ["open", "preparing", "sold_out", "closed"];

// 営業中だけ色を付け、それ以外はマップで灰になることが分かるようにする
const dotClass: Record<BusinessStatus, string> = {
  open: "bg-green-500",
  preparing: "bg-gray-400",
  sold_out: "bg-gray-400",
  closed: "bg-gray-400",
};

/** 4つの営業状況から1つを選ぶ。見た目だけを担当する */
export function BusinessStatusSelector({
  value,
  onChange,
  disabled = false,
}: BusinessStatusSelectorProps) {
  const { t } = useT();

  return (
    <div role="radiogroup" aria-label={t.stall.status.title} className="flex flex-col gap-3">
      {statuses.map((status) => {
        const selected = status === value;
        return (
          <button
            key={status}
            type="button"
            role="radio"
            aria-checked={selected}
            disabled={disabled}
            onClick={() => onChange(status)}
            className={cn(
              "flex items-center gap-3 rounded-xl border p-4 text-left transition-colors disabled:opacity-50",
              selected ? "border-primary bg-primary/5 ring-primary ring-1" : "hover:bg-muted",
            )}
          >
            <span className={cn("size-3 shrink-0 rounded-full", dotClass[status])} aria-hidden />
            <span className="flex-1">
              <span className="block font-bold">{t.businessStatus[status]}</span>
              <span className="text-muted-foreground block text-sm">{t.stall.status.hints[status]}</span>
            </span>
            {selected && <CheckIcon className="text-primary size-5" aria-hidden />}
          </button>
        );
      })}
    </div>
  );
}
