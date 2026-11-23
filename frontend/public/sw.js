// アトクルの Service Worker。
// 目的は Web Push の通知を受け取ること（#63 で push と notificationclick を足す）。
// 古い画面が残る事故を避けるため、キャッシュはしない（fetch のハンドラを置かない）。
// Echo から Cache-Control: no-cache で返るので、更新するとすぐに新しい sw.js が使われる。

self.addEventListener("install", () => {
  // 新しい sw.js を、開いているタブが閉じるのを待たずに有効にする
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(self.clients.claim());
});
