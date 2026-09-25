export type Status = "pending" | "answered" | "published";
export type Question = { id: string; nickname: string; content: string; answer?: string; status: Status; created_at: string; answered_at?: string; image_filename?: string };
export type Settings = { site_name: string; primary_color: string; copyright_name: string; top_bar_opacity: number; navigation_opacity: number; card_opacity: number; max_upload_kb: number };
export type Session = { id: string; created_at: string; expires_at: string; current: boolean };

export async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, init);
  if (!response.ok) {
    const body = await response.json().catch(() => ({})) as { error?: string };
    throw new Error(body.error ?? t("error.requestFailed"));
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

export const formatTime = (value?: string) => value ? new Intl.DateTimeFormat(locale.value, { dateStyle: "medium", timeStyle: "short" }).format(new Date(value)) : "";
import { locale, t } from "./i18n";
