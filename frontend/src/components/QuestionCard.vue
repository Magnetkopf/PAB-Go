<script setup lang="ts">
import { formatTime, type Question } from "../api";
import { useI18n } from "../i18n";
defineProps<{ question: Question; admin?: boolean }>();
const { t } = useI18n();
</script>

<template>
  <mdui-card variant="elevated" class="question-card">
    <p class="mdui-text-color-primary"><mdui-icon-add-comment /> {{ t('question.askedAt', { name: question.nickname || t('question.anonymous'), time: formatTime(question.created_at) }) }}</p>
    <p>{{ question.content }}</p>
    <img v-if="question.image_filename" class="question-image" :src="`/api/images/${question.image_filename}`" :alt="t('question.attachmentAlt')" />
    <mdui-divider />
    <p class="mdui-text-color-primary"><mdui-icon-question-answer /> {{ t('question.answeredAt', { time: formatTime(question.answered_at) }) }}</p>
    <p>{{ question.answer || t('question.unanswered') }}</p>
  </mdui-card>
</template>
