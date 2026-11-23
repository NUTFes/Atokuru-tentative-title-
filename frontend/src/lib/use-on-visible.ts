"use client";

import { useEffect, useRef } from "react";

/**
 * ページを開いたときと、アプリに戻ってきたとき（visibilitychange で表示に戻ったとき）に
 * load を呼ぶ。定期的なポーリングはしない（AGENTS.md の決まり）。
 *
 * load は useCallback で包んで渡す。load が変わったとき（たとえば模擬店の ID を
 * localStorage から読み終えたとき）にも、もう一度呼ぶ。
 */
export function useOnVisible(load: () => void | Promise<void>): void {
  const ref = useRef(load);

  useEffect(() => {
    ref.current = load;
    void load();
  }, [load]);

  useEffect(() => {
    const onChange = () => {
      if (document.visibilityState === "visible") void ref.current();
    };
    document.addEventListener("visibilitychange", onChange);
    return () => document.removeEventListener("visibilitychange", onChange);
  }, []);
}
