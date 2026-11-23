import type { Metadata, Viewport } from "next";
import "./globals.css";

import { PwaSetup } from "@/app/pwa-setup";
import { I18nProvider } from "@/lib/i18n";

export const metadata: Metadata = {
  title: "アトクル",
  description: "学園祭の模擬店の整理券アプリ",
  manifest: "/manifest.json",
  icons: {
    icon: "/icons/icon-192.png",
    apple: "/icons/apple-touch-icon.png",
  },
  appleWebApp: {
    capable: true,
    title: "アトクル",
    statusBarStyle: "default",
  },
};

export const viewport: Viewport = {
  themeColor: "#ea580c",
  // iPhone のホームバーやノッチの下まで画面を広げ、env(safe-area-inset-*) で余白をとる
  viewportFit: "cover",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="ja" className="h-full antialiased">
      <body className="min-h-full flex flex-col">
        <I18nProvider>
          {children}
          <PwaSetup />
        </I18nProvider>
      </body>
    </html>
  );
}
