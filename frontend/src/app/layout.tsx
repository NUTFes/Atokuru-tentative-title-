import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "アトクル",
  description: "学園祭の模擬店の整理券アプリ",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="ja" className="h-full antialiased">
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
