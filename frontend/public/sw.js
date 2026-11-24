// t0m0kuru の Service Worker。
// 目的は Web Push の通知を受け取ること。
// 古い画面が残る事故を避けるため、キャッシュはしない（fetch のハンドラを置かない）。
// Echo から Cache-Control: no-cache で返るので、更新するとすぐに新しい sw.js が使われる。

self.addEventListener("install", () => {
  // 新しい sw.js を、開いているタブが閉じるのを待たずに有効にする
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(self.clients.claim());
});

// 辞書を読めなかったときに使う文言（ja.json の notification と同じ）
const FALLBACK = {
  title: "{stallName}　{number} 番",
  ready: "調理が完了しました。お店で受け取ってください",
  near: "あと {count} 人であなたの番です。お店の近くでお待ちください",
};

function format(template, values) {
  return template.replace(/\{(\w+)\}/g, (match, key) => (key in values ? String(values[key]) : match));
}

/**
 * 通知の文言を、端末の言語の辞書から読む。
 * 辞書は src/locales を public/locales にコピーしたもの（npm run copy-locales）。
 */
async function loadMessages() {
  const locale = (self.navigator.language || "ja").toLowerCase().startsWith("ja") ? "ja" : "en";
  try {
    const response = await fetch(`/locales/${locale}.json`);
    const dict = await response.json();
    return dict.notification ?? FALLBACK;
  } catch {
    return FALLBACK;
  }
}

// サーバーから届いた通知を出す。中身は仕様書の「通知（Web Push）」の JSON
// { type, public_token, stall_name, stall_number, order_number, people_ahead, url }
self.addEventListener("push", (event) => {
  event.waitUntil(
    (async () => {
      let data;
      try {
        data = event.data.json();
      } catch {
        return;
      }
      const messages = await loadMessages();
      const title = format(messages.title, { stallName: data.stall_name, number: data.order_number });
      const body =
        data.type === "near" ? format(messages.near, { count: data.people_ahead }) : messages.ready;
      await self.registration.showNotification(title, {
        body,
        icon: "/icons/icon-192.png",
        badge: "/icons/icon-192.png",
        // 同じ整理券の通知は1つにまとめる（「あと3人」のあとに「調理完了」が来たら置き換える）
        tag: data.public_token,
        renotify: true,
        data: { url: data.url || "/" },
      });
    })(),
  );
});

// 通知を押したら整理券の画面を開く。すでに t0m0kuru を開いていれば、そのタブに移る
self.addEventListener("notificationclick", (event) => {
  event.notification.close();
  const url = new URL(event.notification.data?.url || "/", self.location.origin).href;
  event.waitUntil(
    (async () => {
      const windows = await self.clients.matchAll({ type: "window", includeUncontrolled: true });
      const existing = windows.find((client) => new URL(client.url).origin === self.location.origin);
      if (existing) {
        await existing.focus();
        return existing.navigate(url);
      }
      return self.clients.openWindow(url);
    })(),
  );
});
