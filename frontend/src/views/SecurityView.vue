<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { formatTime, request, type Session } from "../api";
import { showRequestError } from "../feedback";
import { useI18n } from "../i18n";

const router = useRouter();
const { t } = useI18n();
const sessions = ref<Session[]>([]);
const loading = ref(true);
const currentPassword = ref("");
const newPassword = ref("");
const resettingPassword = ref(false);

async function loadSessions() {
  try {
    sessions.value = (await request<{ sessions: Session[] }>("/api/admin/sessions")).sessions;
  } catch (e) { showRequestError(e, t("settings.sessionsLoadFailed")); }
  finally { loading.value = false; }
}

async function revokeSession(session: Session) {
  try {
    await request<void>(`/api/admin/sessions/${session.id}`, { method: "DELETE" });
    if (session.current) {
      await router.push("/admin");
      return;
    }
    sessions.value = sessions.value.filter((item) => item.id !== session.id);
  } catch (e) { showRequestError(e, t("settings.sessionRevokeFailed")); }
}

async function resetPassword() {
  resettingPassword.value = true;
  try {
    await request<void>("/api/admin/password", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ current_password: currentPassword.value, new_password: newPassword.value }),
    });
    currentPassword.value = "";
    newPassword.value = "";
    await router.push({ path: "/admin", query: { passwordReset: "1" } });
  } catch (e) { showRequestError(e, t("settings.passwordResetFailed")); }
  finally { resettingPassword.value = false; }
}

onMounted(loadSessions);
</script>

<template>
  <main class="page security-page">
    <section class="page-intro"><p class="mdui-text-color-primary">{{ t('settings.security') }}!?</p><h1 class="mdui-typo-display-large">{{ t('settings.security') }}</h1><p class="mdui-text-color-on-surface-variant">{{ t('settings.securityDescription') }}</p></section>
    <section class="security-grid">
      <mdui-card variant="outlined" class="password-card">
        <form class="form-stack" @submit.prevent="resetPassword">
          <div><h2 class="mdui-typo-title-large">{{ t('settings.resetPassword') }}</h2><p class="mdui-text-color-on-surface-variant">{{ t('settings.resetPasswordDescription') }}</p></div>
          <mdui-text-field :label="t('settings.currentPassword')" type="password" required :value="currentPassword" @input="currentPassword = String(($event.target as HTMLInputElement).value)" />
          <mdui-text-field :label="t('settings.newPassword')" type="password" required :value="newPassword" @input="newPassword = String(($event.target as HTMLInputElement).value)" />
          <mdui-button type="submit" :disabled="resettingPassword">{{ resettingPassword ? t('settings.resettingPassword') : t('settings.resetPassword') }}</mdui-button>
        </form>
      </mdui-card>
      <section class="sessions-panel">
        <div><h2 class="mdui-typo-title-large">{{ t('settings.sessions') }}</h2><p class="mdui-text-color-on-surface-variant">{{ t('settings.sessionsDescription') }}</p></div>
        <mdui-card v-if="!loading" variant="outlined" class="session-card">
          <div v-for="session in sessions" :key="session.id" class="session-row">
            <mdui-icon-devices class="session-icon" />
            <div class="session-details"><strong>{{ session.current ? t('settings.currentSession') : t('settings.session') }}</strong><span>{{ t('settings.sessionCreated', { time: formatTime(session.created_at) }) }}</span><span>{{ t('settings.sessionExpires', { time: formatTime(session.expires_at) }) }}</span></div>
            <mdui-button variant="outlined" @click="revokeSession(session)">{{ t('settings.revoke') }}</mdui-button>
          </div>
          <p v-if="!sessions.length">{{ t('settings.noSessions') }}</p>
        </mdui-card>
        <p v-else>{{ t('settings.loadingSessions') }}</p>
      </section>
    </section>
  </main>
</template>
