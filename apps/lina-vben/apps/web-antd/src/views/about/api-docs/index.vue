<script setup lang="ts">
import { computed } from 'vue';

import { useAccessStore } from '@vben/stores';

defineOptions({ name: 'ApiDocs' });

const accessStore = useAccessStore();
const iframeSrc = computed(() => {
  const token = accessStore.accessToken || '';
  return `/stoplight/apidocs.html?token=${encodeURIComponent(token)}`;
});
</script>

<template>
  <iframe class="api-docs-iframe" :src="iframeSrc" />
</template>

<style scoped>
.api-docs-iframe {
  width: 100%;
  height: calc(100vh - 100px);
  border: none;
  display: block;
}
</style>
