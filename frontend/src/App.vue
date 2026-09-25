<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { setColorScheme } from "mdui/functions/setColorScheme.js";
import { setTheme } from "mdui/functions/setTheme.js";
import { request, type Settings } from "./api";
import { useI18n } from "./i18n";

const route = useRoute();
const router = useRouter();
const { locale, localeOptions, setLocale, t } = useI18n();
const siteName = ref("个人提问箱");
const dark = ref(false);
const selected = computed(() => route.path.startsWith("/explore") ? "explore" : route.path.startsWith("/admin") ? "admin" : "ask");
const navigate = (path: string) => void router.push(path);
onMounted(async () => {
  try {
    const settings = await request<Settings>("/api/settings");
    siteName.value = settings.site_name;
    setColorScheme(settings.primary_color);
  } catch { /* Backend can be started after Vite during development. */ }
});
function toggleTheme() { dark.value = !dark.value; setTheme(dark.value ? "dark" : "light"); }
</script>

<template>
  <mdui-layout>
    <mdui-top-app-bar class="app-topbar">
      <mdui-top-app-bar-title>{{ siteName }}</mdui-top-app-bar-title>
      <div class="grow" />
      <mdui-button-icon :aria-label="t('nav.search')" @click="navigate('/search')"><mdui-icon-search /></mdui-button-icon>
      <mdui-dropdown placement="bottom-end"><mdui-button-icon slot="trigger" :aria-label="t('language.label')"><mdui-icon-language /></mdui-button-icon><mdui-menu><mdui-menu-item v-for="option in localeOptions" :key="option.code" :selected="locale === option.code" @click="setLocale(option.code)">{{ t(option.labelKey) }}</mdui-menu-item></mdui-menu></mdui-dropdown>
      <mdui-button-icon :aria-label="t('nav.toggleTheme')" @click="toggleTheme"><mdui-icon-dark-mode v-if="!dark" /><mdui-icon-light-mode v-else /></mdui-button-icon>
    </mdui-top-app-bar>

    <mdui-navigation-rail class="app-sidebar desktop-navigation" :value="selected" alignment="center">
      <mdui-navigation-rail-item value="ask" @click="navigate('/ask')"><mdui-icon-add-comment slot="icon" />{{ t('nav.ask') }}</mdui-navigation-rail-item>
      <mdui-navigation-rail-item value="explore" @click="navigate('/explore')"><mdui-icon-question-answer slot="icon" />{{ t('nav.explore') }}</mdui-navigation-rail-item>
      <mdui-navigation-rail-item value="admin" @click="navigate('/admin')"><mdui-icon-admin-panel-settings slot="icon" />{{ t('nav.admin') }}</mdui-navigation-rail-item>
    </mdui-navigation-rail>

    <mdui-layout-main>
      <router-view />
    </mdui-layout-main>

    <mdui-navigation-bar class="mobile-navigation" :value="selected">
      <mdui-navigation-bar-item value="ask" @click="navigate('/ask')"><mdui-icon-add-comment slot="icon" />{{ t('nav.ask') }}</mdui-navigation-bar-item>
      <mdui-navigation-bar-item value="explore" @click="navigate('/explore')"><mdui-icon-question-answer slot="icon" />{{ t('nav.explore') }}</mdui-navigation-bar-item>
      <mdui-navigation-bar-item value="admin" @click="navigate('/admin')"><mdui-icon-admin-panel-settings slot="icon" />{{ t('nav.admin') }}</mdui-navigation-bar-item>
    </mdui-navigation-bar>
  </mdui-layout>
</template>
