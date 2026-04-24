<script setup lang="ts">
import type { PluginRouteReviewItem } from '#/api/system/plugin/model';

import { computed, ref, watch } from 'vue';

import { Button, Tag } from 'ant-design-vue';

interface Props {
  collapsedCount?: number;
  routes: PluginRouteReviewItem[];
}

const props = withDefaults(defineProps<Props>(), {
  collapsedCount: 2,
});

const expanded = ref(false);

const hasExpandableRoutes = computed(() => {
  return props.routes.length > props.collapsedCount;
});

const visibleRoutes = computed(() => {
  if (expanded.value || !hasExpandableRoutes.value) {
    return props.routes;
  }
  return props.routes.slice(0, props.collapsedCount);
});

watch(
  () => props.routes,
  () => {
    expanded.value = false;
  },
  { deep: true },
);

function getAccessLabel(access: string) {
  return access === 'public' ? '公开访问' : '登录访问';
}

function getAccessColor(access: string) {
  return access === 'public' ? 'gold' : 'green';
}

function buildRouteKey(route: PluginRouteReviewItem, index: number) {
  return `${route.method}-${route.publicPath}-${index}`;
}

function toggleExpanded() {
  expanded.value = !expanded.value;
}
</script>

<template>
  <div data-testid="plugin-route-review-list" class="flex flex-col gap-3">
    <div
      v-for="(route, index) in visibleRoutes"
      :key="buildRouteKey(route, index)"
      :data-testid="`plugin-route-review-item-${index}`"
      class="rounded-md border border-[var(--ant-color-border)] p-4"
    >
      <div class="flex flex-wrap items-center gap-2">
        <Tag color="blue">{{ route.method }}</Tag>
        <Tag :color="getAccessColor(route.access)">
          {{ getAccessLabel(route.access) }}
        </Tag>
        <Tag v-if="route.permission">
          {{ route.permission }}
        </Tag>
      </div>

      <div
        class="mt-2 break-all rounded bg-[var(--ant-color-fill-quaternary)] px-3 py-2 font-mono text-[13px] text-[var(--ant-color-text)]"
      >
        {{ route.publicPath }}
      </div>

      <div
        v-if="route.summary"
        class="mt-2 text-[13px] font-medium text-[var(--ant-color-text)]"
      >
        {{ route.summary }}
      </div>

      <div
        v-if="route.description"
        class="mt-1 text-[12px] leading-6 text-[var(--ant-color-text-secondary)]"
      >
        {{ route.description }}
      </div>
    </div>

    <div
      v-if="hasExpandableRoutes"
      class="flex items-center justify-center"
    >
      <Button
        data-testid="plugin-route-review-toggle"
        type="link"
        @click="toggleExpanded"
      >
        {{ expanded ? '收起' : '展开' }}
      </Button>
    </div>
  </div>
</template>
