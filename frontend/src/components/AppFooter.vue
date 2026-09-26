<script setup lang="ts">
defineProps<{ copyrightName: string }>();

const year = new Date().getFullYear();
</script>

<template>
  <footer v-if="copyrightName" class="app-footer">
    <mdui-divider />
    <p>© {{ year }} {{ copyrightName }}</p>
  </footer>
</template>
