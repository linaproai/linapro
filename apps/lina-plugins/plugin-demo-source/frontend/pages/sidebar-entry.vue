<script lang="ts">
export const pluginPageMeta = {
  routePath: 'plugin-demo-source-sidebar-entry',
  title: '源码插件示例',
};
</script>

<script setup lang="ts">
import { onMounted, ref } from 'vue';

import {
  Card as ACard,
  TypographyParagraph as ATypographyParagraph,
  TypographyTitle as ATypographyTitle,
} from 'ant-design-vue';

import { requestClient } from '#/api/request';
import { Page } from '#/plugins/dynamic';

interface PluginSummary {
  message: string;
}

const intro = ref('');

async function loadSummary() {
  const summary = await requestClient.get<PluginSummary>('/plugins/plugin-demo-source/summary');
  intro.value = summary.message;
}

onMounted(() => {
  void loadSummary();
});
</script>

<template>
  <Page :auto-content-height="true">
    <a-card :bordered="false" class="border border-slate-200">
      <a-typography-title :level="3" class="!mb-3">
        源码插件示例已生效
      </a-typography-title>
      <a-typography-paragraph class="!mb-0 text-slate-600" v-if="intro">
        {{ intro }}
      </a-typography-paragraph>
    </a-card>
  </Page>
</template>
