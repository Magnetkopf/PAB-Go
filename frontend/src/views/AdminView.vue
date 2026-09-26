<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { request, type Question } from "../api";
import { useI18n } from "../i18n";

const router = useRouter();
const route = useRoute();
const authenticated = ref(false);
const loading = ref(true);
const questions = ref<Question[]>([]);
const username = ref("");
const password = ref("");
const error = ref("");
const notice = ref(route.query.passwordReset === "1" ? "passwordReset" : "");
const { t } = useI18n();
const pendingCount = computed(() => questions.value.filter((question) => question.status === "pending").length);

async function loadDashboard() {
  questions.value = (await request<{ questions: Question[] }>("/api/questions?status=all")).questions;
}

void request<{ authenticated: boolean }>("/api/admin/session").then(async (session) => {
  authenticated.value = session.authenticated;
  if (session.authenticated) await loadDashboard();
}).finally(() => { loading.value = false; });

async function login() {
  error.value = "";
  try {
    await request("/api/admin/login", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ username: username.value, password: password.value }) });
    authenticated.value = true;
    await loadDashboard();
  } catch (e) { error.value = e instanceof Error ? e.message : t("admin.loginFailed"); }
}
</script>

<template>
  <main v-if="!loading && !authenticated" class="page split-page">
    <section><p class="mdui-text-color-primary">{{ t('admin.eyebrow') }}</p><h1 class="mdui-typo-display-large">{{ t('admin.title') }}</h1><p class="mdui-text-color-on-surface-variant">{{ t('admin.loginDescription') }}</p></section>
    <mdui-card variant="elevated" class="panel"><form class="form-stack" @submit.prevent="login"><mdui-text-field :label="t('admin.username')" required :value="username" @input="username = String(($event.target as HTMLInputElement).value)" /><mdui-text-field :label="t('admin.password')" type="password" required :value="password" @input="password = String(($event.target as HTMLInputElement).value)" /><mdui-button type="submit">{{ t('admin.login') }}</mdui-button><p v-if="notice" class="mdui-text-color-primary">{{ t('settings.passwordResetSuccess') }}</p><p v-if="error" class="mdui-text-color-error">{{ error }}</p></form></mdui-card>
  </main>
  <main v-else-if="authenticated" class="page">
    <section class="page-intro"><p class="mdui-text-color-primary">{{ t('admin.eyebrow') }}</p><h1 class="mdui-typo-display-large">{{ t('admin.menuTitle') }}</h1><p class="mdui-text-color-on-surface-variant">{{ t('admin.menuDescription') }}</p></section>
    <section class="admin-dashboard" :aria-label="t('admin.dashboard')">
      <mdui-card variant="filled" class="dashboard-card"><span>{{ t('admin.allQuestions') }}</span><strong>{{ questions.length }}</strong></mdui-card>
      <mdui-card variant="filled" class="dashboard-card"><span>{{ t('admin.pendingQuestions') }}</span><strong>{{ pendingCount }}</strong></mdui-card>
    </section>
    <mdui-list>
    <mdui-list-item :headline="t('admin.questions')" :description="t('admin.questionsDescription')" rounded @click="router.push('/admin/questions')"><mdui-icon-list slot="icon" /><mdui-icon-arrow-forward slot="end-icon" /></mdui-list-item>
    <mdui-list-item :headline="t('admin.settings')" :description="t('admin.settingsDescription')" rounded @click="router.push('/admin/settings')"><mdui-icon-settings slot="icon" /><mdui-icon-arrow-forward slot="end-icon" /></mdui-list-item>
    <mdui-list-item :headline="t('settings.security')" :description="t('settings.securityDescription')" rounded @click="router.push('/admin/security')"><mdui-icon-security slot="icon" /><mdui-icon-arrow-forward slot="end-icon" /></mdui-list-item>

    </mdui-list>
  </main>
</template>
