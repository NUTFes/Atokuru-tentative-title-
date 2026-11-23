import type { Metadata } from "next";
import "./globals.css";

import { I18nProvider } from "@/lib/i18n";

export const metadata: Metadata = {
  title: "アトクル",
  description: "学園祭の模擬店の整理券アプリ",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="ja" className="h-full antialiased">
      <body className="min-h-full flex flex-col">
        <I18nProvider>{children}</I18nProvider>
      </body>
    </html>
  );
}
