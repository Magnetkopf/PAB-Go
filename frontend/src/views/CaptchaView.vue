<script setup lang="ts">
import { onMounted, ref } from "vue";
import { createChallenge, pbkdf2, sha, type DeriveKeyFunction } from "altcha/lib";
import { request, type CaptchaSettings } from "../api";
import { showRequestError, showSnackbar } from "../feedback";
import { useI18n } from "../i18n";

const { t } = useI18n();
const settings = ref<CaptchaSettings | null>(null);
const previewing = ref(false);
const previewChallenge = ref("");
const previewKey = ref(0);
const previewResult = ref("");

onMounted(async () => { try { settings.value = await request<CaptchaSettings>("/api/admin/captcha/settings"); } catch (e) { showRequestError(e, t("settings.loadFailed")); } });
function update(key: keyof CaptchaSettings, value: string | number | boolean) { if (settings.value) settings.value = { ...settings.value, [key]: value }; }
async function save() { if (!settings.value) return; try { settings.value = await request<CaptchaSettings>("/api/admin/captcha/settings", { method: "PUT", headers: { "Content-Type": "application/json" }, body: JSON.stringify(settings.value) }); showSnackbar(t("settings.saved")); } catch (e) { showRequestError(e, t("settings.saveFailed")); } }

function previewDeriveKey(algorithm: string): DeriveKeyFunction {
  return algorithm.startsWith("PBKDF2/") ? pbkdf2.deriveKey : sha.deriveKey;
}

async function tryCaptcha() {
  if (!settings.value) return;
  const { captcha_algorithm: algorithm, captcha_cost: cost } = settings.value;
  if (!Number.isInteger(cost) || cost < 1000 || cost > 100000) { showSnackbar(t("settings.captchaPreviewInvalid")); return; }
  previewing.value = true;
  previewResult.value = "";
  try {
    const deriveKey = previewDeriveKey(algorithm);
    const challenge = await createChallenge({ algorithm, cost, counter: 1000, deriveKey });
    previewChallenge.value = JSON.stringify(challenge);
    previewKey.value += 1;
    previewResult.value = t("settings.captchaPreviewReady");
  } catch {
    previewResult.value = t("settings.captchaPreviewFailed");
  } finally {
    previewing.value = false;
  }
}

function completePreview(event: Event) {
  try {
    const payload = (event as CustomEvent<{ payload: string }>).detail.payload;
    const solution = JSON.parse(atob(payload)).solution;
    previewResult.value = t("settings.captchaPreviewResult", { attempts: solution.counter + 1, time: Math.round(solution.time) });
  } catch {
    previewResult.value = t("settings.captchaPreviewCompleted");
  }
}
</script>

<template>
  <main v-if="settings" class="page settings-page">
    <section class="page-intro"><p class="mdui-text-color-primary">{{ t('settings.captcha') }}</p><h1 class="mdui-typo-display-large">{{ t('settings.captcha') }}</h1><p class="mdui-text-color-on-surface-variant">{{ t('admin.captchaDescription') }}</p></section>
    <form class="settings-form" @submit.prevent="save">
      <section class="settings-section" :aria-label="t('settings.captcha')">
        <div class="form-stack">
          <mdui-switch :checked="settings.captcha_enabled" @change="update('captcha_enabled', ($event.target as HTMLInputElement).checked)">{{ t('settings.captchaEnabled') }}</mdui-switch>
          <mdui-select :label="t('settings.captchaAlgorithm')" :value="settings.captcha_algorithm" :disabled="!settings.captcha_enabled" @change="update('captcha_algorithm', String(($event.target as HTMLInputElement).value))">
            <mdui-menu-item value="PBKDF2/SHA-256">PBKDF2/SHA-256</mdui-menu-item>
            <mdui-menu-item value="PBKDF2/SHA-384">PBKDF2/SHA-384</mdui-menu-item>
            <mdui-menu-item value="PBKDF2/SHA-512">PBKDF2/SHA-512</mdui-menu-item>
            <mdui-menu-item value="SHA-256">SHA-256</mdui-menu-item>
            <mdui-menu-item value="SHA-384">SHA-384</mdui-menu-item>
            <mdui-menu-item value="SHA-512">SHA-512</mdui-menu-item>
          </mdui-select>
          <mdui-text-field :label="t('settings.captchaCost')" type="number" min="1000" max="100000" step="1000" :disabled="!settings.captcha_enabled" :value="settings.captcha_cost" @input="update('captcha_cost', Number(($event.target as HTMLInputElement).value))" />
          <mdui-button class="captcha-try-button" type="button" variant="outlined" :disabled="previewing" :loading="previewing" @click="tryCaptcha">{{ t('settings.captchaTry') }}</mdui-button>
          <altcha-widget v-if="previewChallenge" :key="previewKey" :challenge="previewChallenge" display="standard" @verified="completePreview" />
          <p v-if="previewResult" class="mdui-text-color-on-surface-variant">{{ previewResult }}</p>
        </div>
      </section>
      <mdui-fab type="submit" extended><mdui-icon-save slot="icon" />{{ t('settings.save') }}</mdui-fab>
    </form>
  </main>
</template>
