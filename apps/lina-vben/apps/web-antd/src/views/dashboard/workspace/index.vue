<script lang="ts" setup>
import type {
  WorkbenchProjectItem,
  WorkbenchQuickNavItem,
  WorkbenchTodoItem,
  WorkbenchTrendItem,
} from '@vben/common-ui';

import { computed } from 'vue';
import { useRouter } from 'vue-router';

import {
  AnalysisChartCard,
  WorkbenchHeader,
  WorkbenchProject,
  WorkbenchQuickNav,
  WorkbenchTodo,
  WorkbenchTrends,
} from '@vben/common-ui';
import { preferences } from '@vben/preferences';
import { useUserStore } from '@vben/stores';
import { openWindow } from '@vben/utils';

import PluginSlotOutlet from '#/components/plugin/plugin-slot-outlet.vue';
import { $t } from '#/locales';
import { pluginSlotKeys } from '#/plugins/plugin-slots';

import AnalyticsVisitsSource from '../analytics/analytics-visits-source.vue';

const userStore = useUserStore();

const projectItems = computed<WorkbenchProjectItem[]>(() => [
  {
    content: $t('pages.dashboard.workspace.projects.linapro.content'),
    date: '2026-05-01',
    group: $t('pages.dashboard.workspace.projects.linapro.group'),
    logo: '/logo.webp',
    title: 'LinaPro',
    url: 'https://linapro.ai',
  },
  {
    content: $t('pages.dashboard.workspace.projects.goframe.content'),
    date: '2026-05-01',
    group: $t('pages.dashboard.workspace.projects.goframe.group'),
    logo: '/goframe-logo.webp',
    title: 'GoFrame',
    url: 'https://goframe.org',
  },
  {
    color: '#42b883',
    content: $t('pages.dashboard.workspace.projects.vue.content'),
    date: '2026-05-01',
    group: $t('pages.dashboard.workspace.projects.vue.group'),
    icon: 'ion:logo-vue',
    title: 'Vue',
    url: 'https://vuejs.org',
  },
  {
    content: $t('pages.dashboard.workspace.projects.vben.content'),
    date: '2026-05-01',
    group: $t('pages.dashboard.workspace.projects.vben.group'),
    logo: '/vben-logo.webp',
    title: 'Vben',
    url: 'https://www.vben.pro',
  },
  {
    color: '',
    content: $t('pages.dashboard.workspace.projects.antDesign.content'),
    date: '2026-05-01',
    group: $t('pages.dashboard.workspace.projects.antDesign.group'),
    icon: 'svg:antdv-next-logo',
    title: 'Ant Design',
    url: 'https://antdv.com',
  },
  {
    color: '#3178c6',
    content: $t('pages.dashboard.workspace.projects.typescript.content'),
    date: '2026-05-01',
    group: $t('pages.dashboard.workspace.projects.typescript.group'),
    icon: 'simple-icons:typescript',
    title: 'TypeScript',
    url: 'https://www.typescriptlang.org',
  },
]);

const quickNavItems = computed<WorkbenchQuickNavItem[]>(() => [
  {
    color: '#2563eb',
    icon: 'lucide:users',
    title: $t('pages.dashboard.workspace.quickNav.userManagement'),
    url: '/system/user',
  },
  {
    color: '#f59e0b',
    icon: 'lucide:menu',
    title: $t('pages.dashboard.workspace.quickNav.menuManagement'),
    url: '/system/menu',
  },
  {
    color: '#14b8a6',
    icon: 'lucide:sliders-horizontal',
    title: $t('pages.dashboard.workspace.quickNav.systemParameters'),
    url: '/system/config',
  },
  {
    color: '#8b5cf6',
    icon: 'lucide:puzzle',
    title: $t('pages.dashboard.workspace.quickNav.extensionCenter'),
    url: '/system/plugin',
  },
  {
    color: '#0ea5e9',
    icon: 'lucide:file-code',
    title: $t('pages.dashboard.workspace.quickNav.apiDocs'),
    url: '/about/api-docs',
  },
  {
    color: '#ef4444',
    icon: 'lucide:clock-3',
    title: $t('pages.dashboard.workspace.quickNav.scheduledJobs'),
    url: '/system/job',
  },
]);

const todoItems = computed<WorkbenchTodoItem[]>(() => [
  {
    completed: false,
    content: $t('pages.dashboard.workspace.todos.reviewShortcuts.content'),
    date: '2026-04-29 10:00:00',
    title: $t('pages.dashboard.workspace.todos.reviewShortcuts.title'),
  },
  {
    completed: true,
    content: $t('pages.dashboard.workspace.todos.verifyI18n.content'),
    date: '2026-04-29 11:00:00',
    title: $t('pages.dashboard.workspace.todos.verifyI18n.title'),
  },
  {
    completed: false,
    content: $t('pages.dashboard.workspace.todos.preparePluginDemo.content'),
    date: '2026-04-29 14:00:00',
    title: $t('pages.dashboard.workspace.todos.preparePluginDemo.title'),
  },
  {
    completed: false,
    content: $t('pages.dashboard.workspace.todos.inspectJobLogs.content'),
    date: '2026-04-29 16:00:00',
    title: $t('pages.dashboard.workspace.todos.inspectJobLogs.title'),
  },
  {
    completed: false,
    content: $t('pages.dashboard.workspace.todos.syncApiDocs.content'),
    date: '2026-04-30 09:30:00',
    title: $t('pages.dashboard.workspace.todos.syncApiDocs.title'),
  },
]);

const trendItems = computed<WorkbenchTrendItem[]>(() => [
  {
    avatar: 'svg:avatar-1',
    content: $t('pages.dashboard.workspace.trends.items.releasedWorkspaceNav'),
    date: $t('pages.dashboard.workspace.trends.justNow'),
    title: $t('pages.dashboard.workspace.trends.people.platformTeam'),
  },
  {
    avatar: 'svg:avatar-2',
    content: $t('pages.dashboard.workspace.trends.items.alignedAccessMenus'),
    date: $t('pages.dashboard.workspace.trends.oneHourAgo'),
    title: $t('pages.dashboard.workspace.trends.people.accessTeam'),
  },
  {
    avatar: 'svg:avatar-3',
    content: $t('pages.dashboard.workspace.trends.items.publishedPluginDemo'),
    date: $t('pages.dashboard.workspace.trends.oneDayAgo'),
    title: $t('pages.dashboard.workspace.trends.people.extensionCenter'),
  },
  {
    avatar: 'svg:avatar-4',
    content: $t('pages.dashboard.workspace.trends.items.refreshedApiDocs'),
    date: $t('pages.dashboard.workspace.trends.twoDaysAgo'),
    title: $t('pages.dashboard.workspace.trends.people.apiDocs'),
  },
  {
    avatar: 'svg:avatar-1',
    content: $t('pages.dashboard.workspace.trends.items.checkedSchedulerLogs'),
    date: $t('pages.dashboard.workspace.trends.threeDaysAgo'),
    title: $t('pages.dashboard.workspace.trends.people.scheduler'),
  },
  {
    avatar: 'svg:avatar-2',
    content: $t('pages.dashboard.workspace.trends.items.updatedPublicConfig'),
    date: $t('pages.dashboard.workspace.trends.oneWeekAgo'),
    title: $t('pages.dashboard.workspace.trends.people.configCenter'),
  },
  {
    avatar: 'svg:avatar-3',
    content: $t('pages.dashboard.workspace.trends.items.draftedOpenSpec'),
    date: $t('pages.dashboard.workspace.trends.oneWeekAgo'),
    title: $t('pages.dashboard.workspace.trends.people.aiWorkflow'),
  },
  {
    avatar: 'svg:avatar-4',
    content: $t('pages.dashboard.workspace.trends.items.passedPlaywright'),
    date: '2026-04-21 09:30',
    title: $t('pages.dashboard.workspace.trends.people.testSuite'),
  },
  {
    avatar: 'svg:avatar-4',
    content: $t('pages.dashboard.workspace.trends.items.archivedDemoIteration'),
    date: '2026-04-18 18:00',
    title: $t('pages.dashboard.workspace.trends.people.releasePipeline'),
  },
]);

const router = useRouter();
const displayUserName = computed(
  () =>
    userStore.userInfo?.realName || $t('pages.dashboard.workspace.defaultName'),
);
const welcomeTitle = computed(() =>
  $t('pages.dashboard.workspace.greeting', { name: displayUserName.value }),
);

function navTo(nav: WorkbenchProjectItem | WorkbenchQuickNavItem) {
  if (nav.url?.startsWith('http')) {
    openWindow(nav.url);
    return;
  }
  if (nav.url?.startsWith('/')) {
    void router.push(nav.url);
  }
}
</script>

<template>
  <div class="p-5" data-testid="dashboard-workspace-page">
    <WorkbenchHeader
      :avatar="userStore.userInfo?.avatar || preferences.app.defaultAvatar"
      :project-label="$t('pages.dashboard.workspace.stats.projects')"
      :team-label="$t('pages.dashboard.workspace.stats.team')"
      :todo-label="$t('pages.dashboard.workspace.stats.todos')"
    >
      <template #title>
        {{ welcomeTitle }}
      </template>
      <template #description>
        <span data-testid="dashboard-workspace-description">
          {{ $t('pages.dashboard.workspace.weather') }}
        </span>
      </template>
    </WorkbenchHeader>

    <PluginSlotOutlet
      :slot-key="pluginSlotKeys.dashboardWorkspaceBefore"
      class="mt-5"
    />

    <div class="mt-5 flex flex-col lg:flex-row">
      <div class="mr-4 w-full lg:w-3/5">
        <div data-testid="dashboard-workspace-projects">
          <WorkbenchProject
            :items="projectItems"
            :title="$t('pages.dashboard.workspace.sections.projects')"
            @click="navTo"
          />
        </div>
        <div class="mt-5" data-testid="dashboard-workspace-trends">
          <WorkbenchTrends
            :items="trendItems"
            :title="$t('pages.dashboard.workspace.sections.trends')"
          />
        </div>
      </div>
      <div class="w-full lg:w-2/5">
        <div data-testid="dashboard-workspace-quick-nav">
          <WorkbenchQuickNav
            :items="quickNavItems"
            class="mt-5 lg:mt-0"
            :title="$t('pages.dashboard.workspace.sections.quickNav')"
            @click="navTo"
          />
        </div>
        <div class="mt-5" data-testid="dashboard-workspace-todos">
          <WorkbenchTodo
            :items="todoItems"
            :title="$t('pages.dashboard.workspace.sections.todos')"
          />
        </div>
        <AnalysisChartCard
          class="mt-5"
          :title="$t('pages.dashboard.workspace.sections.trafficSources')"
        >
          <div data-testid="dashboard-workspace-traffic-card">
            <AnalyticsVisitsSource />
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
