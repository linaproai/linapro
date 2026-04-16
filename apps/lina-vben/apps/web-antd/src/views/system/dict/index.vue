<script setup lang="ts">
import { onUnmounted } from 'vue';

import { Page } from '@vben/common-ui';

import DictDataPanel from './data/index.vue';
import { emitter } from './mitt';
import DictTypePanel from './type/index.vue';

onUnmounted(() => emitter.off('rowClick'));
</script>

<template>
  <Page
    :auto-content-height="true"
    content-class="flex flex-col lg:flex-row gap-4"
  >
    <DictTypePanel id="dict-type" class="flex-1 overflow-hidden" />
    <DictDataPanel id="dict-data" class="flex-1 overflow-hidden" />
  </Page>
</template>
