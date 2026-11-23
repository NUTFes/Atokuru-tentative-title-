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
