<script setup lang="ts">
import { ref } from "vue";
import QuestionCard from "../components/QuestionCard.vue";
import { request, type Question } from "../api";
import { useI18n } from "../i18n";

const { t } = useI18n();
const questions = ref<Question[]>([]);
void request<{ questions: Question[] }>("/api/questions?status=published").then(list => { questions.value = list.questions; }).catch(() => undefined);
</script>

<template>
  <main class="page">
    <section class="page-intro">
      <p class="mdui-text-color-primary">{{ t('explore.eyebrow') }}</p>
      <h1 class="mdui-typo-display-large">{{ t('explore.title') }}</h1>
      <p class="mdui-text-color-on-surface-variant">{{ t('explore.description') }}</p>
    </section>
    <section class="card-list">
      <QuestionCard v-for="question in questions" :key="question.id" :question="question" />
      <p v-if="!questions.length" class="mdui-text-color-on-surface-variant">{{ t('explore.empty') }}</p>
    </section>
  </main>
</template>
