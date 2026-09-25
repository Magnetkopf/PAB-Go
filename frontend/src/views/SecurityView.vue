<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { formatTime, request, type Session } from "../api";
import { useI18n } from "../i18n";

const router = useRouter();
const { t } = useI18n();
const sessions = ref<Session[]>([]);
const loading = ref(true);
const message = ref("");

async function loadSessions() {
  message.value = "";
  try {
    sessions.value = (await request<{ sessions: Session[] }>("/api/admin/sessions")).sessions;
  } catch (e) { message.value = e instanceof Error ? e.message : t("settings.sessionsLoadFailed"); }
  finally { loading.value = false; }
}

async function revokeSession(session: Session) {
  message.value = "";
  try {
    await request<void>(`/api/admin/sessions/${session.id}`, { method: "DELETE" });
    if (session.current) {
      await router.push("/admin");
      return;
    }
    sessions.value = sessions.value.filter((item) => item.id !== session.id);
  } catch (e) { message.value = e instanceof Error ? e.message : t("settings.sessionRevokeFailed"); }
}

onMounted(loadSessions);
</script>

<template>
  <main class="page security-page">
    <section class="page-intro"><p class="mdui-text-color-primary">{{ t('settings.security') }}</p><h1 class="mdui-typo-display-large">{{ t('settings.sessions') }}</h1><p class="mdui-text-color-on-surface-variant">{{ t('settings.sessionsDescription') }}</p></section>
    <mdui-card v-if="!loading" variant="outlined" class="session-card">
      <div v-for="session in sessions" :key="session.id" class="session-row">
        <mdui-icon-devices class="session-icon" />
        <div class="session-details"><strong>{{ session.current ? t('settings.currentSession') : t('settings.session') }}</strong><span>{{ t('settings.sessionCreated', { time: formatTime(session.created_at) }) }}</span><span>{{ t('settings.sessionExpires', { time: formatTime(session.expires_at) }) }}</span></div>
        <mdui-button variant="outlined" @click="revokeSession(session)">{{ t('settings.revoke') }}</mdui-button>
      </div>
      <p v-if="!sessions.length">{{ t('settings.noSessions') }}</p>
    </mdui-card>
    <p v-else>{{ t('settings.loadingSessions') }}</p>
    <p v-if="message" class="mdui-text-color-error">{{ message }}</p>
  </main>
</template>
