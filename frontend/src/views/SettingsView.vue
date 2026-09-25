<script setup lang="ts">
import { onMounted, ref } from "vue";
import { setColorScheme } from "mdui/functions/setColorScheme.js";
import { request, type Settings } from "../api";
import { useI18n } from "../i18n";

const { t } = useI18n();
const settings = ref<Settings | null>(null);
const message = ref("");
onMounted(async () => { try { settings.value = await request<Settings>("/api/settings"); } catch (e) { message.value = e instanceof Error ? e.message : t("settings.loadFailed"); } });
function update(key: keyof Settings, value: string | number) { if (settings.value) settings.value = { ...settings.value, [key]: value }; }
async function save() { if (!settings.value) return; try { settings.value = await request<Settings>("/api/admin/settings", { method: "PUT", headers: { "Content-Type": "application/json" }, body: JSON.stringify(settings.value) }); setColorScheme(settings.value.primary_color); message.value = t("settings.saved"); } catch (e) { message.value = e instanceof Error ? e.message : t("settings.saveFailed"); } }
</script>

<template>
  <main v-if="settings" class="page settings-page">
    <section class="page-intro"><p class="mdui-text-color-primary">{{ t('settings.eyebrow') }}</p><h1 class="mdui-typo-display-large">{{ t('settings.title') }}</h1><p class="mdui-text-color-on-surface-variant">{{ t('settings.description') }}</p></section>
    <mdui-list>
      <mdui-list-subheader>{{ t('settings.basics') }}</mdui-list-subheader>
      <mdui-list-item><mdui-text-field :label="t('settings.siteName')" :value="settings.site_name" @input="update('site_name', String(($event.target as HTMLInputElement).value))" /></mdui-list-item>
      <mdui-list-item><mdui-text-field :label="t('settings.copyrightName')" :value="settings.copyright_name" @input="update('copyright_name', String(($event.target as HTMLInputElement).value))" /></mdui-list-item>
      <mdui-list-item><mdui-text-field :label="t('settings.maxUploadKB')" type="number" min="1" max="10240" :value="settings.max_upload_kb" @input="update('max_upload_kb', Number(($event.target as HTMLInputElement).value))" /></mdui-list-item>
      <mdui-list-subheader>{{ t('settings.theme') }}</mdui-list-subheader>
      <mdui-list-item><label>{{ t('settings.primaryColor') }} <input type="color" :value="settings.primary_color" @input="update('primary_color', ($event.target as HTMLInputElement).value)" /></label></mdui-list-item>
      <mdui-list-item><label>{{ t('settings.cardOpacity', { value: settings.card_opacity }) }} <input type="range" min="0" max="100" :value="settings.card_opacity" @input="update('card_opacity', Number(($event.target as HTMLInputElement).value))" /></label></mdui-list-item>
    </mdui-list>
    <mdui-button @click="save">{{ t('settings.save') }}</mdui-button><p v-if="message" class="mdui-text-color-primary">{{ message }}</p>
  </main>
</template>
