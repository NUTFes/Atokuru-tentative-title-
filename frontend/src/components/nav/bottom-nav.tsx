import Link from "next/link";
import type { LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";

export type BottomNavItem = {
  /** 遷移先の URL */
  href: string;
  /** 表示する文言 */
  label: string;
  /** 表示するアイコン */
  icon: LucideIcon;
};

export type BottomNavProps = {
  /** 並べる項目。遷移先は呼び出す側が決める */
  items: BottomNavItem[];
  /** 今いるページの href。一致する項目を強調する */
  activeHref: string;
};

/** 画面下部に固定するナビゲーションバー。見た目だけを担当する */
export function BottomNav({ items, activeHref }: BottomNavProps) {
  return (
    <nav className="bg-background fixed inset-x-0 bottom-0 z-40 border-t pb-[env(safe-area-inset-bottom)]">
      <ul className="mx-auto flex max-w-md">
        {items.map(({ href, label, icon: Icon }) => {
          const active = href === activeHref;
          return (
            <li key={href} className="flex-1">
              <Link
                href={href}
                aria-current={active ? "page" : undefined}
                className={cn(
                  "flex flex-col items-center gap-1 py-2 text-xs",
                  active ? "text-foreground font-bold" : "text-muted-foreground",
                )}
              >
                <Icon className="size-6" aria-hidden />
                {label}
              </Link>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
