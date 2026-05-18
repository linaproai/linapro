<script setup lang="ts">
import { computed } from 'vue';

import { preferences } from '@vben/preferences';
import { useAccessStore } from '@vben/stores';

defineOptions({ name: 'ApiDocs' });

const accessStore = useAccessStore();
const iframeSrc = computed(() => {
  const params = new URLSearchParams();
  params.set('lang', preferences.app.locale || 'zh-CN');
  params.set('token', accessStore.accessToken || '');
  return `/stoplight/apidocs.html?${params.toString()}`;
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
