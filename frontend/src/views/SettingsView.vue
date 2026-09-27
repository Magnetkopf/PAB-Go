<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { setColorScheme } from "mdui/functions/setColorScheme.js";
import { request, type Settings, type TelegramSettings } from "../api";
import { showRequestError, showSnackbar } from "../feedback";
import { useI18n } from "../i18n";

const { t } = useI18n();
const settings = ref<Settings | null>(null);
const telegram = ref<TelegramSettings>({ telegram_enabled: false, telegram_bot_token: "", telegram_user_id: "" });
const testingTelegram = ref(false);
const telegramReady = computed(() => Boolean(telegram.value.telegram_bot_token.trim() && telegram.value.telegram_user_id.trim()));
onMounted(async () => {
  try {
    [settings.value, telegram.value] = await Promise.all([
      request<Settings>("/api/settings"),
      request<TelegramSettings>("/api/admin/telegram/settings"),
    ]);
  } catch (e) { showRequestError(e, t("settings.loadFailed")); }
});
function update(key: keyof Settings, value: string | number | boolean) { if (settings.value) settings.value = { ...settings.value, [key]: value }; }
function updateTelegram(key: keyof TelegramSettings, value: string | boolean) {
  telegram.value = { ...telegram.value, [key]: value };
}
async function saveTelegram() {
  telegram.value = await request<TelegramSettings>("/api/admin/telegram/settings", {
    method: "PUT", headers: { "Content-Type": "application/json" }, body: JSON.stringify(telegram.value),
  });
}
async function save() {
  if (!settings.value) return;
  try {
    const savedSettings = await request<Settings>("/api/admin/settings", { method: "PUT", headers: { "Content-Type": "application/json" }, body: JSON.stringify(settings.value) });
    await saveTelegram();
    settings.value = savedSettings;
    setColorScheme(settings.value.primary_color);
    showSnackbar(t("settings.saved"));
  } catch (e) { showRequestError(e, t("settings.saveFailed")); }
}
async function testTelegram() {
  testingTelegram.value = true;
  try {
    await saveTelegram();
    await request("/api/admin/telegram/test", { method: "POST" });
    showSnackbar(t("settings.telegramTestSent"));
  } catch (e) { showRequestError(e, t("settings.telegramTestFailed")); }
  finally { testingTelegram.value = false; }
}
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
      <section class="settings-section" :aria-label="t('settings.telegram')">
        <h2 class="mdui-typo-title-large">{{ t('settings.telegram') }}</h2>
        <p class="mdui-text-color-on-surface-variant">{{ t('settings.telegramDescription') }}</p>
        <div class="form-stack">
          <mdui-switch :checked="telegram.telegram_enabled" @change="updateTelegram('telegram_enabled', ($event.target as HTMLInputElement).checked)">{{ t('settings.telegramEnabled') }}</mdui-switch>
          <mdui-text-field :label="t('settings.telegramBotToken')" type="password" :disabled="!telegram.telegram_enabled" :value="telegram.telegram_bot_token" @input="updateTelegram('telegram_bot_token', String(($event.target as HTMLInputElement).value))" />
          <mdui-text-field :label="t('settings.telegramUserID')" :disabled="!telegram.telegram_enabled" :value="telegram.telegram_user_id" @input="updateTelegram('telegram_user_id', String(($event.target as HTMLInputElement).value))" />
          <mdui-button class="captcha-try-button" type="button" variant="outlined" :disabled="testingTelegram || !telegram.telegram_enabled || !telegramReady" :loading="testingTelegram" @click="testTelegram">{{ t('settings.telegramTest') }}</mdui-button>
        </div>
      </section>
      <mdui-fab type="submit" extended><mdui-icon-save slot="icon" />{{ t('settings.save') }}</mdui-fab>
    </form>
  </main>
</template>
