"use client";

import { ZoomInIcon } from "lucide-react";
import { useCallback, useState } from "react";

import { useT } from "@/lib/i18n";

export type VenueMapProps = {
  /** 全体マップの画像の場所 */
  src: string;
  /** 画像の幅と高さ（読み込む前に場所を空けておき、画面がずれないようにする） */
  width: number;
  height: number;
};

/**
 * 会場の全体マップの画像。幅いっぱいに縮めて出し、押すと画像だけを開いて拡大できる
 * （ブラウザのピンチで拡大する）。画像を読めなかったときは何も出さない。見た目だけを担当する。
 */
export function VenueMap({ src, width, height }: VenueMapProps) {
  const { t } = useT();
  const [failed, setFailed] = useState(false);
  // 静的エクスポートでは React の準備より先に画像を読み込むので、onError を逃すことがある。
  // 準備ができた時点で、すでに読み込みに失敗していないかも確かめる
  const checkLoaded = useCallback((img: HTMLImageElement | null) => {
    if (img?.complete && img.naturalWidth === 0) setFailed(true);
  }, []);

  if (failed) return null;

  return (
    <figure className="flex flex-col gap-1">
      <a href={src} target="_blank" rel="noopener" className="block overflow-hidden rounded-xl border">
        {/* 静的エクスポートでは next/image の最適化が使えないので、<img> をそのまま使う */}
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img
          ref={checkLoaded}
          src={src}
          width={width}
          height={height}
          alt={t.map.imageAlt}
          className="h-auto w-full"
          onError={() => setFailed(true)}
        />
      </a>
      <figcaption className="text-muted-foreground flex items-center gap-1 text-xs">
        <ZoomInIcon className="size-3.5" aria-hidden />
        {t.map.imageHint}
      </figcaption>
    </figure>
  );
}
