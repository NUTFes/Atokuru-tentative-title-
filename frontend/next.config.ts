import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // 静的エクスポート。`npm run build` で out/ に HTML・CSS・JS が出力され、Go（Echo）から配信する。
  // そのため API Routes・Server Actions・SSR は使えない。
  output: "export",
  // /map を out/map/index.html として出力し、Echo の静的配信でそのまま返せるようにする（#31）。
  trailingSlash: true,
  // 静的エクスポートでは Next.js の画像最適化サーバーが動かないので、元の画像をそのまま使う。
  images: {
    unoptimized: true,
  },
  turbopack: {
    rules: {
      "*.css": {
        loaders: ["@tailwindcss/turbopack"],
        as: "*.css",
      },
    },
  },
};

export default nextConfig;
