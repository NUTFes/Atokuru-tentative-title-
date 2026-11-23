"use client";

import { MapIcon, TicketIcon } from "lucide-react";

import { BottomNav } from "@/components/nav/bottom-nav";
import { useT } from "@/lib/i18n";

export type UserBottomNavProps = {
  /** ホーム（整理券一覧）の URL */
  homeHref: string;
  /** マップの URL */
  mapHref: string;
  /** 今いるページの href */
  activeHref: string;
};

/** ユーザー画面のボトムナビ。ホームとマップを切り替える */
export function UserBottomNav({ homeHref, mapHref, activeHref }: UserBottomNavProps) {
  const { t } = useT();

  return (
    <BottomNav
      activeHref={activeHref}
      items={[
        { href: homeHref, label: t.nav.user.home, icon: TicketIcon },
        { href: mapHref, label: t.nav.user.map, icon: MapIcon },
      ]}
    />
  );
}
