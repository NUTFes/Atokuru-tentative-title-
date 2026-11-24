import Image from "next/image";

export type VenueMapProps = {
  /** 画像の場所（public/ からのパス）。例："/map.png" */
  src: string;
  /** 画像の代わりの文言 */
  alt: string;
  /** 画像の元の幅（px）。縦横比を決めて、読み込み中に画面がずれないようにする */
  width: number;
  /** 画像の元の高さ（px） */
  height: number;
  /** 画像の下に出す、拡大のしかたの案内 */
  hint?: string;
};

/**
 * 会場の全体マップ。見た目だけを担当する。
 * スマホの幅いっぱいに縮めて出し、押すと画像だけを新しいタブで開く（ブラウザのピンチで拡大できる）。
 */
export function VenueMap({ src, alt, width, height, hint }: VenueMapProps) {
  return (
    <figure className="flex flex-col gap-1">
      <a
        href={src}
        target="_blank"
        rel="noopener"
        className="focus-visible:ring-ring/50 block overflow-hidden rounded-xl border outline-none focus-visible:ring-3"
      >
        {/* 静的エクスポートなので、next.config.ts の images.unoptimized で元の画像をそのまま使う */}
        <Image src={src} alt={alt} width={width} height={height} className="h-auto w-full" priority />
      </a>
      {hint && <figcaption className="text-muted-foreground text-xs">{hint}</figcaption>}
    </figure>
  );
}
