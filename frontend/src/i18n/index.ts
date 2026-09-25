import { ref } from "vue";
import enUS from "./en-US";
import zhCN from "./zh-CN";
import zhHK from "./zh-HK";

export const locales = { "zh-CN": zhCN, "zh-HK": zhHK, "en-US": enUS };
export type Locale = keyof typeof locales;
export const localeOptions: { code: Locale; labelKey: "language.zhCN" | "language.zhHK" | "language.enUS" }[] = [
  { code: "zh-CN", labelKey: "language.zhCN" },
  { code: "zh-HK", labelKey: "language.zhHK" },
  { code: "en-US", labelKey: "language.enUS" },
];

const cookieName = "pab_locale";
const validLocale = (value?: string): value is Locale => Boolean(value && value in locales);
const readCookie = () => document.cookie.split("; ").find((entry) => entry.startsWith(`${cookieName}=`))?.split("=")[1];
const initialLocale = (): Locale => {
  const stored = readCookie();
  if (validLocale(stored)) return stored;
  return navigator.language.toLowerCase().startsWith("zh-hk") || navigator.language.toLowerCase().startsWith("zh-tw") ? "zh-HK" : navigator.language.toLowerCase().startsWith("en") ? "en-US" : "zh-CN";
};

export const locale = ref<Locale>(initialLocale());

export function t(path: string, values: Record<string, string | number> = {}) {
  const translation = path.split(".").reduce<unknown>((current, key) => current && typeof current === "object" ? (current as Record<string, unknown>)[key] : undefined, locales[locale.value]);
  if (typeof translation !== "string") return path;
  return translation.replace(/\{(\w+)\}/g, (_, key: string) => String(values[key] ?? `{${key}}`));
}

export function setLocale(value: Locale) {
  locale.value = value;
  document.documentElement.lang = value;
  document.cookie = `${cookieName}=${value}; Max-Age=31536000; Path=/; SameSite=Lax`;
}

document.documentElement.lang = locale.value;
export const useI18n = () => ({ locale, localeOptions, setLocale, t });
