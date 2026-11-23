"use client";

import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";

import { InstallGuide } from "@/components/sheet/install-guide";
import { isIOS, isStandalone, registerServiceWorker } from "@/lib/pwa";
import { loadInstallGuideDismissed, saveInstallGuideDismissed } from "@/lib/storage";

/**
 * Service Worker を登録し、iPhone でまだホーム画面に追加していなければ案内を出す。
 * iPhone はホーム画面に追加しないと Web Push が届かないため。
 * 案内は来場者の画面だけに出す（模擬店の画面では通知を使わない）。
 */
export function PwaSetup() {
  const pathname = usePathname();
  const [needsGuide, setNeedsGuide] = useState(false);

  useEffect(() => {
    registerServiceWorker();
    // eslint-disable-next-line react-hooks/set-state-in-effect -- 端末の情報はブラウザでしか分からないので、マウント後に反映する
    setNeedsGuide(isIOS() && !isStandalone() && !loadInstallGuideDismissed());
  }, []);

  const isStallPage = pathname.startsWith("/stall");

  return (
    <InstallGuide
      open={needsGuide && !isStallPage}
      onDismiss={() => {
        saveInstallGuideDismissed();
        setNeedsGuide(false);
      }}
    />
  );
}
