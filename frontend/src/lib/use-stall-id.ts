"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { loadStallId } from "@/lib/storage";

/**
 * 模擬店画面でログインしている店の ID を返す。読み込むまでは null。
 * ログインしていなければ、ログインページに移る。
 */
export function useStallId(): number | null {
  const router = useRouter();
  const [stallId, setStallId] = useState<number | null>(null);

  useEffect(() => {
    const id = loadStallId();
    if (id === null) {
      router.replace("/stall/login/");
      return;
    }
    // eslint-disable-next-line react-hooks/set-state-in-effect -- localStorage はブラウザでしか読めないので、マウント後に反映する
    setStallId(id);
  }, [router]);

  return stallId;
}
