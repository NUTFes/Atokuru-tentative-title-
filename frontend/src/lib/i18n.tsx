"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";

import { loadLocale, saveLocale } from "@/lib/storage";
import en from "@/locales/en.json";
import ja from "@/locales/ja.json";

export type Dict = typeof ja;
export type Locale = "ja" | "en";

// en のキーが ja と違うと、ここで型エラーになる
const dicts: Record<Locale, Dict> = { ja, en };

export const locales = Object.keys(dicts) as Locale[];

function isLocale(value: string | null): value is Locale {
  return value !== null && value in dicts;
}

/** 保存した言語 → ブラウザの言語 → 日本語 の順に決める */
function detectLocale(): Locale {
  const saved = loadLocale();
  if (isLocale(saved)) return saved;
  return navigator.language.toLowerCase().startsWith("ja") ? "ja" : "en";
}

type I18nValue = {
  t: Dict;
  locale: Locale;
  setLocale: (locale: Locale) => void;
};

const I18nContext = createContext<I18nValue | null>(null);

export function I18nProvider({ children }: { children: React.ReactNode }) {
  // 静的エクスポートでは日本語で書き出し、開いたあとに端末の言語に切り替える
  const [locale, setLocaleState] = useState<Locale>("ja");

  useEffect(() => {
    const detected = detectLocale();
    // eslint-disable-next-line react-hooks/set-state-in-effect -- localStorage はブラウザでしか読めないので、マウント後に反映する
    setLocaleState(detected);
  }, []);

  useEffect(() => {
    document.documentElement.lang = locale;
  }, [locale]);

  const setLocale = useCallback((next: Locale) => {
    saveLocale(next);
    setLocaleState(next);
  }, []);

  const value = useMemo(() => ({ t: dicts[locale], locale, setLocale }), [locale, setLocale]);

  return <I18nContext value={value}>{children}</I18nContext>;
}

/** 今の言語の辞書を受け取る。I18nProvider の中で呼ぶ */
export function useT(): I18nValue {
  const value = useContext(I18nContext);
  if (!value) throw new Error("useT は I18nProvider の中で呼んでください");
  return value;
}

/** "あなたの前に {count} 人" の {count} を埋める */
export function format(template: string, values: Record<string, string | number>): string {
  return template.replace(/\{(\w+)\}/g, (match, key: string) =>
    key in values ? String(values[key]) : match,
  );
}
