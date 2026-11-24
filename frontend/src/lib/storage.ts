// localStorage の読み書きはこのファイルだけで行う。
// プライベートモードなどで localStorage が使えないときは、保存せずに続ける。

// キーの "atokuru:" は、アプリの名前が t0m0kuru に変わる前のもの。変えると来場者が保存した整理券などが消えるので、そのままにする。
const LOCALE_KEY = "atokuru:locale";

function read(key: string): string | null {
  try {
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
}

function write(key: string, value: string): void {
  try {
    window.localStorage.setItem(key, value);
  } catch {
    // 保存できなくても画面はそのまま動かす
  }
}

/** 保存した言語を返す。保存していなければ null */
export function loadLocale(): string | null {
  return read(LOCALE_KEY);
}

/** 言語を保存する */
export function saveLocale(locale: string): void {
  write(LOCALE_KEY, locale);
}

const INSTALL_GUIDE_DISMISSED_KEY = "atokuru:install-guide-dismissed";

/** ホーム画面に追加の案内を閉じたか */
export function loadInstallGuideDismissed(): boolean {
  return read(INSTALL_GUIDE_DISMISSED_KEY) === "1";
}

/** ホーム画面に追加の案内を閉じたことを保存する */
export function saveInstallGuideDismissed(): void {
  write(INSTALL_GUIDE_DISMISSED_KEY, "1");
}

const TICKET_TOKENS_KEY = "atokuru:ticket-tokens";

/** 保存した整理券の public_token の一覧（新しい順） */
export function loadTicketTokens(): string[] {
  try {
    const value: unknown = JSON.parse(read(TICKET_TOKENS_KEY) ?? "[]");
    return Array.isArray(value) ? value.filter((v): v is string => typeof v === "string") : [];
  } catch {
    return [];
  }
}

/** 整理券の public_token を先頭に足す。すでにあれば false を返す */
export function addTicketToken(token: string): boolean {
  const tokens = loadTicketTokens();
  if (tokens.includes(token)) return false;
  write(TICKET_TOKENS_KEY, JSON.stringify([token, ...tokens]));
  return true;
}

/** 整理券の public_token を一覧から消す */
export function removeTicketToken(token: string): void {
  write(TICKET_TOKENS_KEY, JSON.stringify(loadTicketTokens().filter((t) => t !== token)));
}

const STALL_ID_KEY = "atokuru:stall-id";

/** 模擬店画面でログインしている店の ID。ログインしていなければ null */
export function loadStallId(): number | null {
  const id = Number(read(STALL_ID_KEY));
  return Number.isInteger(id) && id > 0 ? id : null;
}

/** 模擬店画面でログインした店の ID を保存する */
export function saveStallId(id: number): void {
  write(STALL_ID_KEY, String(id));
}

/** 模擬店画面からログアウトする */
export function clearStallId(): void {
  try {
    window.localStorage.removeItem(STALL_ID_KEY);
  } catch {
    // 消せなくても画面はそのまま動かす
  }
}

const PUSH_REGISTERED_KEY = "atokuru:push-registered-tokens";

/** 通知の購読をサーバーに登録した整理券の public_token の一覧 */
export function loadPushRegisteredTokens(): string[] {
  try {
    const value: unknown = JSON.parse(read(PUSH_REGISTERED_KEY) ?? "[]");
    return Array.isArray(value) ? value.filter((v): v is string => typeof v === "string") : [];
  } catch {
    return [];
  }
}

/** 通知の購読をサーバーに登録した整理券を覚える（同じ整理券に何度も登録しないため） */
export function savePushRegisteredToken(token: string): void {
  const tokens = loadPushRegisteredTokens();
  if (!tokens.includes(token)) write(PUSH_REGISTERED_KEY, JSON.stringify([...tokens, token]));
}
