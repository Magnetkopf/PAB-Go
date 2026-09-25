<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { formatTime, request, type Question, type Status } from "../api";
import { useI18n } from "../i18n";

const { t } = useI18n();

const status = ref<Status | "all">("pending");
const questions = ref<Question[]>([]);
const opened = ref("");
const answers = ref<Record<string, string>>({});
const published = ref<Record<string, boolean>>({});
const message = ref("");
async function load() { try { const result = await request<{ questions: Question[] }>(`/api/questions?status=${status.value}`); questions.value = result.questions; answers.value = Object.fromEntries(result.questions.map(q => [q.id, q.answer ?? ""])); published.value = Object.fromEntries(result.questions.map(q => [q.id, q.status === "published"])); } catch (e) { message.value = e instanceof Error ? e.message : t("inbox.loadFailed"); } }
async function save(id: string) { try { await request(`/api/admin/questions/${id}/answer`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ answer: answers.value[id], publish: published.value[id] }) }); message.value = t("inbox.saved"); await load(); } catch (e) { message.value = e instanceof Error ? e.message : t("inbox.saveFailed"); } }
watch(status, load); onMounted(load);
</script>

<template>
  <main class="page">
    <section class="page-intro"><p class="mdui-text-color-primary">{{ t('inbox.eyebrow') }}</p><h1 class="mdui-typo-display-large">{{ t('inbox.title') }}</h1><p class="mdui-text-color-on-surface-variant">{{ t('inbox.description') }}</p></section>
    <mdui-tabs :value="status" full-width>
      <mdui-tab value="pending" @click="status = 'pending'">{{ t('inbox.pending') }}</mdui-tab><mdui-tab value="answered" @click="status = 'answered'">{{ t('inbox.answered') }}</mdui-tab><mdui-tab value="published" @click="status = 'published'">{{ t('inbox.published') }}</mdui-tab><mdui-tab value="all" @click="status = 'all'">{{ t('inbox.all') }}</mdui-tab>
    </mdui-tabs>
    <p v-if="message" class="mdui-text-color-primary">{{ message }}</p>
    <section class="card-list">
      <mdui-card v-for="question in questions" :key="question.id" variant="elevated" class="inbox-card">
        <mdui-list-item :headline="question.nickname || t('question.anonymous')" :description="formatTime(question.created_at)" @click="opened = opened === question.id ? '' : question.id"><mdui-icon-add-comment slot="icon" /><mdui-icon-arrow-forward slot="end-icon" /></mdui-list-item>
        <div v-if="opened === question.id" class="inbox-detail"><p>{{ question.content }}</p><img v-if="question.image_filename" class="question-image" :src="`/api/images/${question.image_filename}`" :alt="t('question.attachmentAlt')" /><form class="form-stack" @submit.prevent="save(question.id)"><mdui-text-field :label="t('inbox.answer')" rows="4" required :value="answers[question.id]" @input="answers[question.id] = String(($event.target as HTMLInputElement).value)" /><mdui-checkbox :checked="published[question.id]" @change="published[question.id] = Boolean(($event.target as HTMLInputElement).checked)">{{ t('inbox.publish') }}</mdui-checkbox><mdui-button type="submit">{{ t('inbox.save') }}</mdui-button></form></div>
      </mdui-card>
      <p v-if="!questions.length" class="mdui-text-color-on-surface-variant">{{ t('inbox.empty') }}</p>
    </section>
  </main>
</template>
