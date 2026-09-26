<script setup lang="ts">
import { computed, ref } from "vue";
import type { AltchaWidgetElement } from "altcha";
import { useRoute, useRouter } from "vue-router";
import { request, type Question, type Settings } from "../api";
import { showRequestError, showSnackbar } from "../feedback";
import { useI18n } from "../i18n";

const router = useRouter();
const route = useRoute();
const authenticated = ref(false);
const loading = ref(true);
const questions = ref<Question[]>([]);
const username = ref("");
const password = ref("");
const captchaEnabled = ref(false);
const captcha = ref<AltchaWidgetElement | null>(null);
const captchaPayload = ref("");
const { t } = useI18n();
const pendingCount = computed(() => questions.value.filter((question) => question.status === "pending").length);

async function loadDashboard() {
  questions.value = (await request<{ questions: Question[] }>("/api/questions?status=all")).questions;
}

void request<{ authenticated: boolean }>("/api/admin/session").then(async (session) => {
  authenticated.value = session.authenticated;
  if (session.authenticated) await loadDashboard();
}).finally(() => { loading.value = false; });

if (route.query.passwordReset === "1") showSnackbar(t("settings.passwordResetSuccess"));
void request<Settings>("/api/settings").then(settings => { captchaEnabled.value = settings.captcha_enabled; }).catch(() => undefined);
function rememberCaptcha(event: Event) { captchaPayload.value = (event as CustomEvent<{ payload?: string }>).detail.payload ?? ""; }

async function login() {
  let payload: string | undefined;
  if (captchaEnabled.value) {
    if (captcha.value?.getState() !== "verified" || !captchaPayload.value) { showSnackbar(t("admin.captchaRequired")); return; }
    payload = captchaPayload.value;
  }
  try {
    await request("/api/admin/login", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ username: username.value, password: password.value, altcha: payload }) });
    authenticated.value = true;
    await loadDashboard();
  } catch (e) { captchaPayload.value = ""; captcha.value?.reset(); showRequestError(e, t("admin.loginFailed")); }
}
</script>

<template>
  <main v-if="!loading && !authenticated" class="page split-page">
    <section><p class="mdui-text-color-primary">{{ t('admin.eyebrow') }}</p><h1 class="mdui-typo-display-large">{{ t('admin.title') }}</h1><p class="mdui-text-color-on-surface-variant">{{ t('admin.loginDescription') }}</p></section>
    <mdui-card variant="elevated" class="panel"><form class="form-stack" @submit.prevent="login"><mdui-text-field :label="t('admin.username')" required :value="username" @input="username = String(($event.target as HTMLInputElement).value)" /><mdui-text-field :label="t('admin.password')" type="password" required :value="password" @input="password = String(($event.target as HTMLInputElement).value)" /><altcha-widget v-if="captchaEnabled" ref="captcha" challenge="/api/captcha/challenge/login" display="standard" @verified="rememberCaptcha" /><mdui-button type="submit">{{ t('admin.login') }}</mdui-button></form></mdui-card>
  </main>
  <main v-else-if="authenticated" class="page">
    <section class="page-intro"><p class="mdui-text-color-primary">{{ t('admin.eyebrow') }}</p><h1 class="mdui-typo-display-large">{{ t('admin.menuTitle') }}</h1><p class="mdui-text-color-on-surface-variant">{{ t('admin.menuDescription') }}</p></section>
    <section class="admin-dashboard" :aria-label="t('admin.dashboard')">
      <mdui-card variant="filled" class="dashboard-card"><span>{{ t('admin.allQuestions') }}</span><strong>{{ questions.length }}</strong></mdui-card>
      <mdui-card variant="filled" class="dashboard-card"><span>{{ t('admin.pendingQuestions') }}</span><strong>{{ pendingCount }}</strong></mdui-card>
    </section>
    <mdui-list>
    <mdui-list-item :headline="t('admin.questions')" :description="t('admin.questionsDescription')" @click="router.push('/admin/questions')"><mdui-icon-list slot="icon" /><mdui-icon-arrow-forward slot="end-icon" /></mdui-list-item>
    <mdui-list-item :headline="t('admin.settings')" :description="t('admin.settingsDescription')" @click="router.push('/admin/settings')"><mdui-icon-settings slot="icon" /><mdui-icon-arrow-forward slot="end-icon" /></mdui-list-item>
    <mdui-list-item :headline="t('settings.security')" :description="t('settings.securityDescription')" @click="router.push('/admin/security')"><mdui-icon-security slot="icon" /><mdui-icon-arrow-forward slot="end-icon" /></mdui-list-item>

    </mdui-list>
  </main>
</template>
