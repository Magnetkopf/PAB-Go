<script setup lang="ts">
import { ref } from "vue";
import { request, type Settings } from "../api";
import { showRequestError, showSnackbar } from "../feedback";
import { useI18n } from "../i18n";

const { t } = useI18n();
const nickname = ref("");
const content = ref("");
const image = ref<File | null>(null);
const maxUploadKB = ref(1024);
void request<Settings>("/api/settings").then(settings => { maxUploadKB.value = settings.max_upload_kb; }).catch(() => undefined);

function selectImage(event: Event) { image.value = (event.target as HTMLInputElement).files?.[0] ?? null; }

async function submit() {
  try {
    if (image.value && image.value.size > maxUploadKB.value * 1024) throw new Error(t("ask.imageTooLarge", { size: maxUploadKB.value }));
    const form = new FormData();
    form.set("nickname", nickname.value); form.set("content", content.value);
    if (image.value) form.set("image", image.value);
    await request("/api/questions", { method: "POST", body: form });
    nickname.value = ""; content.value = ""; image.value = null; showSnackbar(t("ask.submitted"));
  } catch (error) { showRequestError(error, t("ask.failed")); }
}
</script>

<template>
  <main class="page split-page">
    <section>
      <p class="mdui-text-color-primary">{{ t('ask.eyebrow') }}</p>
      <h1 class="mdui-typo-display-large">{{ t('ask.defaultTitle') }}</h1>
      <p class="mdui-text-color-on-surface-variant">{{ t('ask.description') }}</p>
    </section>
    <mdui-card variant="elevated" class="panel">
      <form class="form-stack" @submit.prevent="submit">
        <mdui-text-field :label="t('ask.nickname')" :value="nickname" @input="nickname = String(($event.target as HTMLInputElement).value)" />
        <mdui-text-field :label="t('ask.content')" rows="7" required maxlength="1000" counter :value="content" @input="content = String(($event.target as HTMLInputElement).value)" />
        <label class="attachment-picker"><input type="file" accept="image/png,image/jpeg,image/gif,image/webp" @change="selectImage" /><mdui-button variant="outlined" type="button"><mdui-icon-attachment slot="icon" />{{ image ? image.name : t('ask.attachment') }}</mdui-button></label>
        <mdui-button type="submit"><mdui-icon-arrow-forward slot="end-icon" />{{ t('ask.submit') }}</mdui-button>
      </form>
    </mdui-card>
  </main>
</template>
