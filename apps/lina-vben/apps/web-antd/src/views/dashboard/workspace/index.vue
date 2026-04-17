<script lang="ts" setup>
import type {
  WorkbenchProjectItem,
  WorkbenchQuickNavItem,
} from '@vben/common-ui';

import { computed } from 'vue';
import { useRouter } from 'vue-router';

import {
  AnalysisChartCard,
  WorkbenchHeader,
  WorkbenchProject,
  WorkbenchTodo,
  WorkbenchTrends,
} from '@vben/common-ui';
import { IconifyIcon } from '@vben/icons';
import { preferences } from '@vben/preferences';
import { useUserStore } from '@vben/stores';
import { openWindow } from '@vben/utils';
import { Tag } from 'ant-design-vue';

import PluginSlotOutlet from '#/components/plugin/plugin-slot-outlet.vue';
import { pluginSlotKeys } from '#/plugins/plugin-slots';

import AnalyticsVisitsSource from '../analytics/analytics-visits-source.vue';
import {
  workspaceFocusItems,
  workspaceProjectItems,
  workspaceQuickActionItems,
  workspaceTodoItems,
  workspaceTrafficSourceItems,
  workspaceTrendItems,
} from './data';

const userStore = useUserStore();
const router = useRouter();

const displayName = computed(() => userStore.userInfo?.realName || '管理员');
const workspaceDescription = computed(() => {
  return `聚焦核心宿主服务、默认管理工作台、插件扩展能力与 OpenSpec 协作流程。`;
});

function navTo(nav: WorkbenchProjectItem | WorkbenchQuickNavItem) {
  if (nav.url?.startsWith('http')) {
    openWindow(nav.url);
    return;
  }
  if (nav.url?.startsWith('/')) {
    void router.push(nav.url);
  }
}

function quickActionTone(index: number) {
  const toneClasses = [
    'bg-blue-50 text-blue-600',
    'bg-cyan-50 text-cyan-600',
    'bg-emerald-50 text-emerald-600',
    'bg-amber-50 text-amber-600',
    'bg-violet-50 text-violet-600',
    'bg-rose-50 text-rose-600',
  ];

  return toneClasses[index % toneClasses.length];
}
</script>

<template>
  <div class="p-5" data-testid="dashboard-workspace-page">
    <WorkbenchHeader
      :avatar="userStore.userInfo?.avatar || preferences.app.defaultAvatar"
    >
      <template #title>
        欢迎回来，{{ displayName }}。这里是 LinaPro 宿主工作区。
      </template>
      <template #description>
        <span data-testid="dashboard-workspace-description">
          {{ workspaceDescription }}
        </span>
      </template>
    </WorkbenchHeader>

    <PluginSlotOutlet
      :slot-key="pluginSlotKeys.dashboardWorkspaceBefore"
      class="mt-5"
    />

    <div class="mt-5 grid gap-4 xl:grid-cols-4">
      <article
        v-for="item in workspaceFocusItems"
        :key="item.key"
        class="bg-background border-border/60 rounded-3xl border p-5 shadow-sm"
        :data-testid="`dashboard-workspace-focus-${item.key}`"
      >
        <p class="text-foreground/55 text-xs uppercase tracking-[0.18em]">
          {{ item.title }}
        </p>
        <p class="text-foreground mt-3 text-2xl font-semibold">{{ item.value }}</p>
        <p class="text-foreground/70 mt-2 text-sm leading-6">{{ item.description }}</p>
      </article>
    </div>

    <div class="mt-5 flex flex-col gap-5 xl:flex-row">
      <div class="min-w-0 xl:flex-1">
        <div data-testid="dashboard-workspace-projects">
          <WorkbenchProject
            :items="workspaceProjectItems"
            title="重点板块"
            @click="navTo"
          />
        </div>
        <div class="mt-5" data-testid="dashboard-workspace-trends">
          <WorkbenchTrends :items="workspaceTrendItems" title="最新动态" />
        </div>
      </div>

      <div class="xl:w-[440px]">
        <section
          class="bg-background border-border/60 rounded-3xl border p-5 shadow-sm"
          data-testid="dashboard-workspace-quick-actions"
        >
          <div class="flex items-start justify-between gap-4">
            <div>
              <h3 class="text-foreground text-lg font-semibold">重点入口</h3>
              <p class="text-foreground/65 mt-2 text-sm leading-6">
                直接进入当前最常用的管理页，避免模板化占位导航干扰日常巡检。
              </p>
            </div>
            <Tag color="processing">全部可达</Tag>
          </div>

          <div class="mt-4 grid gap-3 sm:grid-cols-2">
            <button
              v-for="(item, index) in workspaceQuickActionItems"
              :key="item.key"
              class="border-border/60 hover:border-primary/50 hover:bg-primary/5 rounded-2xl border p-4 text-left transition-colors"
              :data-testid="`dashboard-workspace-quick-${item.key}`"
              type="button"
              @click="navTo(item)"
            >
              <div class="flex items-start justify-between gap-3">
                <div
                  class="flex h-10 w-10 items-center justify-center rounded-2xl"
                  :class="quickActionTone(index)"
                >
                  <IconifyIcon :icon="item.icon" class="text-lg" />
                </div>
                <span class="bg-foreground/5 text-foreground/60 rounded-full px-2 py-1 text-xs">
                  {{ item.badge }}
                </span>
              </div>
              <div class="text-foreground mt-4 text-sm font-semibold">{{ item.title }}</div>
              <p class="text-foreground/65 mt-2 text-xs leading-5">{{ item.description }}</p>
            </button>
          </div>
        </section>

        <div class="mt-5" data-testid="dashboard-workspace-todos">
          <WorkbenchTodo :items="workspaceTodoItems" title="待办事项" />
        </div>

        <AnalysisChartCard class="mt-5" title="工作台访问来源">
          <div data-testid="dashboard-workspace-traffic-card">
            <AnalyticsVisitsSource :items="workspaceTrafficSourceItems" />
          </div>
        </AnalysisChartCard>
      </div>
    </div>

    <PluginSlotOutlet
      :slot-key="pluginSlotKeys.dashboardWorkspaceAfter"
      class="mt-5"
    />
  </div>
</template>
