<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { setColorScheme } from "mdui/functions/setColorScheme.js";
import { setTheme } from "mdui/functions/setTheme.js";
import { request, type Settings } from "./api";
import AppFooter from "./components/AppFooter.vue";
import { useI18n } from "./i18n";

const route = useRoute();
const router = useRouter();
const { locale, localeOptions, setLocale, t } = useI18n();
const siteName = ref("");
const copyrightName = ref("");
type ThemeMode = "auto" | "light" | "dark";
const themeCookie = "pab_theme";
function readTheme(): ThemeMode {
  const value = document.cookie.split("; ").find((entry) => entry.startsWith(`${themeCookie}=`))?.split("=")[1];
  return value === "light" || value === "dark" || value === "auto" ? value : "auto";
}
const themeMode = ref<ThemeMode>(readTheme());
const colorPreference = window.matchMedia("(prefers-color-scheme: dark)");
const dark = computed(() => themeMode.value === "dark" || (themeMode.value === "auto" && colorPreference.matches));
function applyTheme() { setTheme(dark.value ? "dark" : "light"); }
function setThemeMode(value: ThemeMode) {
  themeMode.value = value;
  document.cookie = `${themeCookie}=${value}; Max-Age=31536000; Path=/; SameSite=Lax`;
  applyTheme();
}
colorPreference.addEventListener("change", applyTheme);
applyTheme();
const selected = computed(() => route.path.startsWith("/explore") ? "explore" : route.path.startsWith("/admin") ? "admin" : "ask");
const navigate = (path: string) => void router.push(path);
onMounted(async () => {
  try {
    const settings = await request<Settings>("/api/settings");
    siteName.value = settings.site_name;
    copyrightName.value = settings.copyright_name;
    setColorScheme(settings.primary_color);
  } catch { /* Backend can be started after Vite during development. */ }
});
function toggleTheme() {
  const modes: ThemeMode[] = ["auto", "light", "dark"];
  setThemeMode(modes[(modes.indexOf(themeMode.value) + 1) % modes.length]);
}
</script>

<template>
  <mdui-layout>
    <mdui-top-app-bar class="app-topbar">
      <mdui-top-app-bar-title>{{ siteName }}</mdui-top-app-bar-title>
      <div class="grow" />
      <mdui-button-icon :aria-label="t('nav.search')" @click="navigate('/search')"><mdui-icon-search /></mdui-button-icon>
      <mdui-dropdown placement="bottom-end"><mdui-button-icon slot="trigger" :aria-label="t('language.label')"><mdui-icon-language /></mdui-button-icon><mdui-menu><mdui-menu-item v-for="option in localeOptions" :key="option.code" :selected="locale === option.code" @click="setLocale(option.code)">{{ t(option.labelKey) }}</mdui-menu-item></mdui-menu></mdui-dropdown>
      <mdui-button-icon :aria-label="t('nav.themeMode', { mode: t(`nav.theme${themeMode[0].toUpperCase()}${themeMode.slice(1)}`) })" @click="toggleTheme"><mdui-icon-brightness-auto v-if="themeMode === 'auto'" /><mdui-icon-brightness-2 v-else-if="themeMode === 'dark'" /><mdui-icon-brightness-7 v-else /></mdui-button-icon>
    </mdui-top-app-bar>

    <mdui-navigation-rail class="app-sidebar desktop-navigation" :value="selected" alignment="center">
      <mdui-navigation-rail-item value="ask" @click="navigate('/ask')"><mdui-icon-add-comment--outlined slot="icon" /><mdui-icon-add-comment slot="active-icon" />{{ t('nav.ask') }}</mdui-navigation-rail-item>
      <mdui-navigation-rail-item value="explore" @click="navigate('/explore')"><mdui-icon-question-answer--outlined slot="icon" /><mdui-icon-question-answer slot="active-icon" />{{ t('nav.explore') }}</mdui-navigation-rail-item>
      <mdui-navigation-rail-item value="admin" @click="navigate('/admin')"><mdui-icon-admin-panel-settings--outlined slot="icon" /><mdui-icon-admin-panel-settings slot="active-icon" />{{ t('nav.admin') }}</mdui-navigation-rail-item>
    </mdui-navigation-rail>

    <mdui-layout-main>
      <router-view />
      <AppFooter :copyright-name="copyrightName" />
    </mdui-layout-main>

    <mdui-navigation-bar class="mobile-navigation" :value="selected">
      <mdui-navigation-bar-item value="ask" @click="navigate('/ask')"><mdui-icon-add-comment--outlined slot="icon" /><mdui-icon-add-comment slot="active-icon" />{{ t('nav.ask') }}</mdui-navigation-bar-item>
      <mdui-navigation-bar-item value="explore" @click="navigate('/explore')"><mdui-icon-question-answer--outlined slot="icon" /><mdui-icon-question-answer slot="active-icon" />{{ t('nav.explore') }}</mdui-navigation-bar-item>
      <mdui-navigation-bar-item value="admin" @click="navigate('/admin')"><mdui-icon-admin-panel-settings--outlined slot="icon" /><mdui-icon-admin-panel-settings slot="active-icon" />{{ t('nav.admin') }}</mdui-navigation-bar-item>
    </mdui-navigation-bar>
  </mdui-layout>
</template>
