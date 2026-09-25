<script setup lang="ts">
import { ref, watch } from "vue";
import QuestionCard from "../components/QuestionCard.vue";
import { request, type Question } from "../api";
import { useI18n } from "../i18n";

const { t } = useI18n();
const query = ref("");
const questions = ref<Question[]>([]);
let timer: number | undefined;
watch(query, (value) => {
  window.clearTimeout(timer);
  timer = window.setTimeout(async () => {
    if (!value.trim()) { questions.value = []; return; }
    try { questions.value = (await request<{ questions: Question[] }>(`/api/questions/search?q=${encodeURIComponent(value)}`)).questions; } catch { questions.value = []; }
  }, 250);
});
</script>

<template>
  <main class="page">
    <section class="page-intro"><p class="mdui-text-color-primary">{{ t('search.eyebrow') }}</p><h1 class="mdui-typo-display-large">{{ t('search.title') }}</h1></section>
    <mdui-text-field :label="t('search.placeholder')" clearable autofocus :value="query" @input="query = String(($event.target as HTMLInputElement).value)"><mdui-icon-search slot="icon" /></mdui-text-field>
    <section class="card-list">
      <QuestionCard v-for="question in questions" :key="question.id" :question="question" />
      <p v-if="query && !questions.length" class="mdui-text-color-on-surface-variant">{{ t('search.empty') }}</p>
    </section>
  </main>
</template>
