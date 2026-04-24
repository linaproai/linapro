<script setup lang="ts">
import { h, onMounted, ref, watch } from 'vue';

import { Page } from '@vben/common-ui';
import { preferences } from '@vben/preferences';

import {
  type ComponentInfo,
  type FrameworkInfo,
  getSystemInfo,
} from '#/api/about';
import { $t } from '#/locales';

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

const frameworkItems = ref<DescriptionItem[]>([]);
const frameworkDescription = ref('');
const backendItems = ref<DescriptionItem[]>([]);
const frontendItems = ref<DescriptionItem[]>([]);
const loading = ref(true);

const mapFramework = (framework: FrameworkInfo): DescriptionItem[] => [
  {
    title: $t('page.about.systemInfo.items.frameworkName'),
    content: framework.name,
  },
  {
    title: $t('page.about.systemInfo.items.homepage'),
    content: renderLink(framework.homepage, $t('page.about.systemInfo.viewLink')),
  },
  {
    title: $t('page.about.systemInfo.items.repositoryUrl'),
    content: renderLink(
      framework.repositoryUrl,
      $t('page.about.systemInfo.viewLink'),
    ),
  },
  { title: $t('page.about.systemInfo.items.version'), content: framework.version },
  { title: $t('page.about.systemInfo.items.license'), content: framework.license },
];

const mapComponents = (components: ComponentInfo[]): DescriptionItem[] =>
  (components || []).map((item) => ({
    title: item.name,
    content: h('div', [
      h('span', { class: 'text-foreground/80' }, item.version),
      h('span', { class: 'mx-2 text-foreground/30' }, '|'),
      renderLink(item.url, item.description),
    ]),
  }));

async function loadSystemInfo() {
  loading.value = true;
  try {
    const info = await getSystemInfo();
    frameworkItems.value = mapFramework(info.framework);
    frameworkDescription.value = info.framework.description;
    backendItems.value = mapComponents(info.backendComponents);
    frontendItems.value = mapComponents(info.frontendComponents);
  } finally {
    loading.value = false;
  }
}

onMounted(async () => {
  await loadSystemInfo();
});

watch(
  () => preferences.app.locale,
  async () => {
    await loadSystemInfo();
  },
);
</script>

<template>
  <Page>
    <!-- 关于项目 -->
    <div class="card-box p-5">
      <h5 class="text-lg text-foreground">
        {{ $t('page.about.systemInfo.sections.framework') }}
      </h5>
      <div class="mt-4">
        <dl v-if="!loading" class="grid grid-cols-2 md:grid-cols-4">
          <template v-for="item in frameworkItems" :key="item.title">
            <div class="border-t border-border px-4 py-3 sm:col-span-1 sm:px-0">
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
        <dl v-if="!loading" class="grid">
          <div class="border-t border-border px-4 py-3 sm:px-0">
            <dt class="text-sm/6 font-medium text-foreground">
              {{ $t('page.about.systemInfo.items.description') }}
            </dt>
            <dd class="mt-1 text-sm/6 text-foreground">
              <span>{{ frameworkDescription }}</span>
            </dd>
          </div>
        </dl>
        <div v-else class="py-8 text-center text-foreground/60">
          {{ $t('page.about.systemInfo.loading') }}
        </div>
      </div>
    </div>

    <!-- 后端组件 -->
    <div class="card-box mt-6 p-5">
      <h5 class="text-lg text-foreground">
        {{ $t('page.about.systemInfo.sections.backend') }}
      </h5>
      <div class="mt-4">
        <dl
          v-if="!loading"
          class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4"
        >
          <template v-for="item in backendItems" :key="item.title">
            <div class="border-t border-border px-4 py-3 sm:col-span-1 sm:px-0">
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
        <div v-else class="py-8 text-center text-foreground/60">
          {{ $t('page.about.systemInfo.loading') }}
        </div>
      </div>
    </div>

    <!-- 前端组件 -->
    <div class="card-box mt-6 p-5">
      <h5 class="text-lg text-foreground">
        {{ $t('page.about.systemInfo.sections.frontend') }}
      </h5>
      <div class="mt-4">
        <dl
          v-if="!loading"
          class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4"
        >
          <template v-for="item in frontendItems" :key="item.title">
            <div class="border-t border-border px-4 py-3 sm:col-span-1 sm:px-0">
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
        <div v-else class="py-8 text-center text-foreground/60">
          {{ $t('page.about.systemInfo.loading') }}
        </div>
      </div>
    </div>
  </Page>
</template>
