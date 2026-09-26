<script setup lang="ts">
import { onMounted, ref } from "vue";
import { setColorScheme } from "mdui/functions/setColorScheme.js";
import { request, type Settings } from "../api";
import { showRequestError, showSnackbar } from "../feedback";
import { useI18n } from "../i18n";

const { t } = useI18n();
const settings = ref<Settings | null>(null);
onMounted(async () => { try { settings.value = await request<Settings>("/api/settings"); } catch (e) { showRequestError(e, t("settings.loadFailed")); } });
function update(key: keyof Settings, value: string | number) { if (settings.value) settings.value = { ...settings.value, [key]: value }; }
async function save() { if (!settings.value) return; try { settings.value = await request<Settings>("/api/admin/settings", { method: "PUT", headers: { "Content-Type": "application/json" }, body: JSON.stringify(settings.value) }); setColorScheme(settings.value.primary_color); showSnackbar(t("settings.saved")); } catch (e) { showRequestError(e, t("settings.saveFailed")); } }
</script>

<template>
  <main v-if="settings" class="page settings-page">
    <section class="page-intro"><p class="mdui-text-color-primary">{{ t('settings.eyebrow') }}</p><h1 class="mdui-typo-display-large">{{ t('settings.title') }}</h1><p class="mdui-text-color-on-surface-variant">{{ t('settings.description') }}</p></section>
    <form class="settings-form" @submit.prevent="save">
      <section class="settings-section" :aria-label="t('settings.basics')">
        <h2 class="mdui-typo-title-large">{{ t('settings.basics') }}</h2>
        <div class="form-stack">
          <mdui-text-field :label="t('settings.siteName')" :value="settings.site_name" @input="update('site_name', String(($event.target as HTMLInputElement).value))" />
          <mdui-text-field :label="t('settings.copyrightName')" :value="settings.copyright_name" @input="update('copyright_name', String(($event.target as HTMLInputElement).value))" />
          <mdui-text-field :label="t('settings.maxUploadKB')" type="number" min="1" max="10240" :value="settings.max_upload_kb" @input="update('max_upload_kb', Number(($event.target as HTMLInputElement).value))" />
        </div>
      </section>
      <section class="settings-section" :aria-label="t('settings.theme')">
        <h2 class="mdui-typo-title-large">{{ t('settings.theme') }}</h2>
        <div class="form-stack">
          <label class="settings-native-field"><span>{{ t('settings.primaryColor') }}</span>
          <input type="color" :value="settings.primary_color" @input="update('primary_color', ($event.target as HTMLInputElement).value)" />
          </label>
          <label class="settings-native-field"><span>{{ t('settings.cardOpacity', { value: settings.card_opacity }) }}</span>
          <mdui-slider min="0" max="100" :value="settings.card_opacity" @input="update('card_opacity', Number(($event.target as HTMLInputElement).value))"></mdui-slider>
          </label>
        </div>
      </section>
      <mdui-fab type="submit" extended><mdui-icon-save slot="icon" />{{ t('settings.save') }}</mdui-fab>
    </form>
  </main>
</template>
