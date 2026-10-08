<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { request, type TelegramSettings } from "../api";
import { showRequestError, showSnackbar } from "../feedback";
import { useI18n } from "../i18n";

const { t } = useI18n();
const telegram = ref<TelegramSettings | null>(null);
const testing = ref(false);
const canTest = computed(() => Boolean(
  telegram.value?.telegram_push_enabled &&
  telegram.value.telegram_bot_token.trim() &&
  telegram.value.telegram_user_id.trim(),
));
onMounted(async () => {
  try {
    telegram.value = await request<TelegramSettings>("/api/admin/telegram/settings");
  } catch (e) {
    showRequestError(e, t("telegram.loadFailed"));
  }
});
function update(key: keyof TelegramSettings, value: string | boolean | number) {
  if (telegram.value) telegram.value = { ...telegram.value, [key]: value };
}
async function save() {
  if (!telegram.value) return;
  try {
    telegram.value = await request<TelegramSettings>("/api/admin/telegram/settings", {
      method: "PUT", headers: { "Content-Type": "application/json" }, body: JSON.stringify(telegram.value),
    });
    showSnackbar(t("telegram.saved"));
  } catch (e) {
    showRequestError(e, t("telegram.saveFailed"));
  }
}
async function test() {
  if (!telegram.value) return;
  testing.value = true;
  try {
    telegram.value = await request<TelegramSettings>("/api/admin/telegram/settings", {
      method: "PUT", headers: { "Content-Type": "application/json" }, body: JSON.stringify(telegram.value),
    });
    await request("/api/admin/telegram/test", { method: "POST" });
    showSnackbar(t("telegram.testSent"));
  } catch (e) {
    showRequestError(e, t("telegram.testFailed"));
  } finally {
    testing.value = false;
  }
}
</script>

<template>
  <main v-if="telegram" class="page settings-page">
    <section class="page-intro">
      <p class="mdui-text-color-primary">Telegram</p>
      <h1 class="mdui-typo-display-large">{{ t('telegram.title') }}</h1>
      <p class="mdui-text-color-on-surface-variant">{{ t('telegram.description') }}</p>
    </section>

    <form class="settings-form" @submit.prevent="save">
      <section class="settings-section">
        <h2 class="mdui-typo-title-large">{{ t('telegram.bot') }}</h2>
        <div class="form-stack">
          <mdui-text-field
            :label="t('telegram.token')"
            type="password"
            :value="telegram.telegram_bot_token"
            @input="update('telegram_bot_token', String(($event.target as HTMLInputElement).value))"
          />
        </div>
      </section>

      <section class="settings-section">
        <h2 class="mdui-typo-title-large">{{ t('telegram.push') }}</h2>
        <div class="form-stack">
          <mdui-list>
            <mdui-list-item
              nonclickable
              :headline="t('telegram.pushEnabled')"
              :description="t('telegram.pushDescription')"
            >
              <mdui-switch
                slot="end-icon"
                :aria-label="t('telegram.pushEnabled')"
                :checked="telegram.telegram_push_enabled"
                @change="update('telegram_push_enabled', ($event.target as HTMLInputElement).checked)"
              />
            </mdui-list-item>
          </mdui-list>
          <mdui-text-field
            :label="t('telegram.userID')"
            :disabled="!telegram.telegram_push_enabled"
            :value="telegram.telegram_user_id"
            @input="update('telegram_user_id', String(($event.target as HTMLInputElement).value))"
          />
          <mdui-button
            class="captcha-try-button"
            type="button"
            variant="outlined"
            :disabled="testing || !canTest"
            :loading="testing"
            @click="test"
          >
            {{ t('telegram.test') }}
          </mdui-button>
        </div>
      </section>

      <section class="settings-section">
        <h2 class="mdui-typo-title-large">{{ t('telegram.ask') }}</h2>
        <div class="form-stack">
          <mdui-list>
            <mdui-list-item
              nonclickable
              :headline="t('telegram.askEnabled')"
              :description="t('telegram.askDescription')"
            >
              <mdui-switch
                slot="end-icon"
                :aria-label="t('telegram.askEnabled')"
                :checked="telegram.telegram_ask_enabled"
                @change="update('telegram_ask_enabled', ($event.target as HTMLInputElement).checked)"
              />
            </mdui-list-item>
          </mdui-list>
          <mdui-text-field
            :label="t('telegram.dailyLimit')"
            type="number"
            min="1"
            max="10000"
            :disabled="!telegram.telegram_ask_enabled"
            :value="telegram.telegram_daily_limit"
            @input="update('telegram_daily_limit', Number(($event.target as HTMLInputElement).value))"
          />
          <p class="mdui-text-color-on-surface-variant">{{ t('telegram.utcNote') }}</p>
          <mdui-list>
            <mdui-list-item
              nonclickable
              :headline="t('telegram.multiplePending')"
              :description="t('telegram.multiplePendingHelp')"
            >
              <mdui-switch
                slot="end-icon"
                :aria-label="t('telegram.multiplePending')"
                :checked="telegram.telegram_allow_multiple_pending"
                :disabled="!telegram.telegram_ask_enabled"
                @change="update('telegram_allow_multiple_pending', ($event.target as HTMLInputElement).checked)"
              />
            </mdui-list-item>
          </mdui-list>
        </div>
      </section>

      <mdui-fab type="submit" extended>
        <mdui-icon-save slot="icon" />
        {{ t('settings.save') }}
      </mdui-fab>
    </form>
  </main>
</template>
