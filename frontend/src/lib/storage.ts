// localStorage の読み書きはこのファイルだけで行う。
// プライベートモードなどで localStorage が使えないときは、保存せずに続ける。

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
