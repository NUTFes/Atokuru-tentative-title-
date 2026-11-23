// Web Push の購読まわり。API は呼ばない（API を呼ぶのは app/ の中だけ）。

/** この端末・ブラウザで Web Push を使えるか。iPhone はホーム画面から開いたときだけ使える */
export function isPushSupported(): boolean {
  return (
    "serviceWorker" in navigator && "PushManager" in window && "Notification" in window
  );
}

/** 通知の許可の状態 */
export function notificationPermission(): NotificationPermission {
  return Notification.permission;
}

/** base64url の VAPID 公開鍵を、pushManager.subscribe に渡せる形にする */
function base64UrlToUint8Array(base64Url: string): Uint8Array<ArrayBuffer> {
  const padding = "=".repeat((4 - (base64Url.length % 4)) % 4);
  const base64 = (base64Url + padding).replace(/-/g, "+").replace(/_/g, "/");
  const raw = atob(base64);
  const bytes = new Uint8Array(new ArrayBuffer(raw.length));
  for (let i = 0; i < raw.length; i++) bytes[i] = raw.charCodeAt(i);
  return bytes;
}

/**
 * 通知の許可を求める。iPhone はボタンを押したときでないと許可を出せないので、
 * ボタンの onClick から呼ぶ。
 */
export async function requestNotificationPermission(): Promise<NotificationPermission> {
  return Notification.requestPermission();
}

/**
 * 購読して、サーバーに送る形（PushSubscription.toJSON()）を返す。
 * すでに購読していれば、それを使う。許可されていなければ null。
 */
export async function subscribePush(vapidPublicKey: string): Promise<PushSubscriptionJSON | null> {
  if (Notification.permission !== "granted") return null;
  const registration = await navigator.serviceWorker.ready;
  const subscription =
    (await registration.pushManager.getSubscription()) ??
    (await registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: base64UrlToUint8Array(vapidPublicKey),
    }));
  return subscription.toJSON();
}
