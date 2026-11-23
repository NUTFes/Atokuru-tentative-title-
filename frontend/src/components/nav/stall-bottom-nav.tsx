"use client";

import { EllipsisIcon, StoreIcon, TicketIcon } from "lucide-react";

import { BottomNav } from "@/components/nav/bottom-nav";
import { useT } from "@/lib/i18n";

export type StallBottomNavProps = {
  /** 整理券一覧の URL */
  ticketsHref: string;
  /** 営業状況の URL */
  statusHref: string;
  /** その他の URL */
  settingsHref: string;
  /** 今いるページの href */
  activeHref: string;
};

/** 店舗画面のボトムナビ。整理券一覧・営業状況・その他を切り替える */
export function StallBottomNav({
  ticketsHref,
  statusHref,
  settingsHref,
  activeHref,
}: StallBottomNavProps) {
  const { t } = useT();

  return (
    <BottomNav
      activeHref={activeHref}
      items={[
        { href: ticketsHref, label: t.nav.stall.tickets, icon: TicketIcon },
        { href: statusHref, label: t.nav.stall.status, icon: StoreIcon },
        { href: settingsHref, label: t.nav.stall.settings, icon: EllipsisIcon },
      ]}
    />
  );
}
