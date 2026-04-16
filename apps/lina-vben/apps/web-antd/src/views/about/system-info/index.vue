<script setup lang="ts">
import { h, onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import { type ComponentInfo, getSystemInfo } from '#/api/about';

import { PROJECT_INFO } from '../config';

defineOptions({ name: 'SystemInfo' });

interface DescriptionItem {
  content: any;
  title: string;
}

const renderLink = (href: string, text: string) =>
  h(
    'a',
    { href, target: '_blank', class: 'vben-link' },
    { default: () => text },
  );

// 关于项目 - 第一行：名称 + 主页 + 版本号 + 许可
const projectRow1: DescriptionItem[] = [
  { title: '项目名称', content: PROJECT_INFO.name },
  {
    title: '项目主页',
    content: renderLink(PROJECT_INFO.homepage, '点击查看'),
  },
  { title: '版本号', content: PROJECT_INFO.version },
  { title: '开源许可', content: PROJECT_INFO.license },
];

// 关于项目 - 第二行：项目介绍
const projectRow2: DescriptionItem[] = [
  { title: '项目介绍', content: PROJECT_INFO.description },
];

const backendItems = ref<DescriptionItem[]>([]);
const frontendItems = ref<DescriptionItem[]>([]);
const loading = ref(true);

const mapComponents = (components: ComponentInfo[]): DescriptionItem[] =>
  (components || []).map((item) => ({
    title: item.name,
    content: h('div', [
      h('span', { class: 'text-foreground/80' }, item.version),
      h('span', { class: 'mx-2 text-foreground/30' }, '|'),
      renderLink(item.url, item.description),
    ]),
  }));

onMounted(async () => {
  try {
    const info = await getSystemInfo();
    backendItems.value = mapComponents(info.backendComponents);
    frontendItems.value = mapComponents(info.frontendComponents);
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <Page>
    <!-- 关于项目 -->
    <div class="card-box p-5">
      <h5 class="text-lg text-foreground">关于项目</h5>
      <div class="mt-4">
        <dl class="grid grid-cols-2 md:grid-cols-4">
          <template v-for="item in projectRow1" :key="item.title">
            <div
              class="border-t border-border px-4 py-3 sm:col-span-1 sm:px-0"
            >
              <dt class="text-sm/6 font-medium text-foreground">
                {{ item.title }}
              </dt>
              <dd class="mt-1 text-sm/6 text-foreground">
                <component
                  :is="item.content"
                  v-if="typeof item.content === 'object'"
                />
                <span v-else>{{ item.content }}</span>
              </dd>
            </div>
          </template>
        </dl>
        <dl class="grid">
          <template v-for="item in projectRow2" :key="item.title">
            <div
              class="border-t border-border px-4 py-3 sm:px-0"
            >
              <dt class="text-sm/6 font-medium text-foreground">
                {{ item.title }}
              </dt>
              <dd class="mt-1 text-sm/6 text-foreground">
                <span>{{ item.content }}</span>
              </dd>
            </div>
          </template>
        </dl>
      </div>
    </div>

    <!-- 后端组件 -->
    <div class="card-box mt-6 p-5">
      <h5 class="text-lg text-foreground">后端组件</h5>
      <div class="mt-4">
        <dl
          v-if="!loading"
          class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4"
        >
          <template v-for="item in backendItems" :key="item.title">
            <div
              class="border-t border-border px-4 py-3 sm:col-span-1 sm:px-0"
            >
              <dt class="text-sm text-foreground">
                {{ item.title }}
              </dt>
              <dd class="mt-1 text-sm text-foreground/80 sm:mt-2">
                <component
                  :is="item.content"
                  v-if="typeof item.content === 'object'"
                />
                <span v-else>{{ item.content }}</span>
              </dd>
            </div>
          </template>
        </dl>
        <div v-else class="py-8 text-center text-foreground/60">加载中...</div>
      </div>
    </div>

    <!-- 前端组件 -->
    <div class="card-box mt-6 p-5">
      <h5 class="text-lg text-foreground">前端组件</h5>
      <div class="mt-4">
        <dl
          v-if="!loading"
          class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4"
        >
          <template v-for="item in frontendItems" :key="item.title">
            <div
              class="border-t border-border px-4 py-3 sm:col-span-1 sm:px-0"
            >
              <dt class="text-sm text-foreground">
                {{ item.title }}
              </dt>
              <dd class="mt-1 text-sm text-foreground/80 sm:mt-2">
                <component
                  :is="item.content"
                  v-if="typeof item.content === 'object'"
                />
                <span v-else>{{ item.content }}</span>
              </dd>
            </div>
          </template>
        </dl>
        <div v-else class="py-8 text-center text-foreground/60">加载中...</div>
      </div>
    </div>
  </Page>
</template>
