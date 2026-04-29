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
    color: '',
    content: $t('pages.dashboard.workspace.projects.github.content'),
    date: '2021-04-01',
    group: $t('pages.dashboard.workspace.projects.github.group'),
    icon: 'carbon:logo-github',
    title: 'Github',
    url: 'https://github.com',
  },
  {
    color: '#3fb27f',
    content: $t('pages.dashboard.workspace.projects.vue.content'),
    date: '2021-04-01',
    group: $t('pages.dashboard.workspace.projects.vue.group'),
    icon: 'ion:logo-vue',
    title: 'Vue',
    url: 'https://vuejs.org',
  },
  {
    color: '#e18525',
    content: $t('pages.dashboard.workspace.projects.html5.content'),
    date: '2021-04-01',
    group: $t('pages.dashboard.workspace.projects.html5.group'),
    icon: 'ion:logo-html5',
    title: 'Html5',
    url: 'https://developer.mozilla.org/zh-CN/docs/Web/HTML',
  },
  {
    color: '#bf0c2c',
    content: $t('pages.dashboard.workspace.projects.angular.content'),
    date: '2021-04-01',
    group: $t('pages.dashboard.workspace.projects.angular.group'),
    icon: 'ion:logo-angular',
    title: 'Angular',
    url: 'https://angular.io',
  },
  {
    color: '#00d8ff',
    content: $t('pages.dashboard.workspace.projects.react.content'),
    date: '2021-04-01',
    group: $t('pages.dashboard.workspace.projects.react.group'),
    icon: 'bx:bxl-react',
    title: 'React',
    url: 'https://reactjs.org',
  },
  {
    color: '#EBD94E',
    content: $t('pages.dashboard.workspace.projects.js.content'),
    date: '2021-04-01',
    group: $t('pages.dashboard.workspace.projects.js.group'),
    icon: 'ion:logo-javascript',
    title: 'Js',
    url: 'https://developer.mozilla.org/zh-CN/docs/Web/JavaScript',
  },
]);

// The reference project points some quick-nav cards to demo routes that do not
// exist here, so these items map to the closest reachable pages in LinaPro.
const quickNavItems = computed<WorkbenchQuickNavItem[]>(() => [
  {
    color: '#1fdaca',
    icon: 'ion:home-outline',
    title: $t('pages.dashboard.workspace.quickNav.home'),
    url: '/',
  },
  {
    color: '#bf0c2c',
    icon: 'ion:grid-outline',
    title: $t('pages.dashboard.workspace.quickNav.dashboard'),
    url: '/dashboard',
  },
  {
    color: '#e18525',
    icon: 'ion:layers-outline',
    title: $t('pages.dashboard.workspace.quickNav.components'),
    url: '/about/api-docs',
  },
  {
    color: '#3fb27f',
    icon: 'ion:settings-outline',
    title: $t('pages.dashboard.workspace.quickNav.system'),
    url: '/system/user',
  },
  {
    color: '#4daf1bc9',
    icon: 'ion:key-outline',
    title: $t('pages.dashboard.workspace.quickNav.access'),
    url: '/system/role',
  },
  {
    color: '#00d8ff',
    icon: 'ion:bar-chart-outline',
    title: $t('pages.dashboard.workspace.quickNav.charts'),
    url: '/dashboard/analytics',
  },
]);

const todoItems = computed<WorkbenchTodoItem[]>(() => [
  {
    completed: false,
    content: $t('pages.dashboard.workspace.todos.reviewFrontend.content'),
    date: '2024-07-30 11:00:00',
    title: $t('pages.dashboard.workspace.todos.reviewFrontend.title'),
  },
  {
    completed: true,
    content: $t('pages.dashboard.workspace.todos.optimizePerformance.content'),
    date: '2024-07-30 11:00:00',
    title: $t('pages.dashboard.workspace.todos.optimizePerformance.title'),
  },
  {
    completed: false,
    content: $t('pages.dashboard.workspace.todos.securityReview.content'),
    date: '2024-07-30 11:00:00',
    title: $t('pages.dashboard.workspace.todos.securityReview.title'),
  },
  {
    completed: false,
    content: $t('pages.dashboard.workspace.todos.updateDependencies.content'),
    date: '2024-07-30 11:00:00',
    title: $t('pages.dashboard.workspace.todos.updateDependencies.title'),
  },
  {
    completed: false,
    content: $t('pages.dashboard.workspace.todos.fixUiIssues.content'),
    date: '2024-07-30 11:00:00',
    title: $t('pages.dashboard.workspace.todos.fixUiIssues.title'),
  },
]);

const trendItems = computed<WorkbenchTrendItem[]>(() => [
  {
    avatar: 'svg:avatar-1',
    content: $t('pages.dashboard.workspace.trends.items.createdVue'),
    date: $t('pages.dashboard.workspace.trends.justNow'),
    title: $t('pages.dashboard.workspace.trends.people.william'),
  },
  {
    avatar: 'svg:avatar-2',
    content: $t('pages.dashboard.workspace.trends.items.followedWilliam'),
    date: $t('pages.dashboard.workspace.trends.oneHourAgo'),
    title: $t('pages.dashboard.workspace.trends.people.evan'),
  },
  {
    avatar: 'svg:avatar-3',
    content: $t('pages.dashboard.workspace.trends.items.postedUpdate'),
    date: $t('pages.dashboard.workspace.trends.oneDayAgo'),
    title: $t('pages.dashboard.workspace.trends.people.chris'),
  },
  {
    avatar: 'svg:avatar-4',
    content: $t('pages.dashboard.workspace.trends.items.viteArticle'),
    date: $t('pages.dashboard.workspace.trends.twoDaysAgo'),
    title: 'Vben',
  },
  {
    avatar: 'svg:avatar-1',
    content: $t('pages.dashboard.workspace.trends.items.answeredOptimization'),
    date: $t('pages.dashboard.workspace.trends.threeDaysAgo'),
    title: $t('pages.dashboard.workspace.trends.people.peter'),
  },
  {
    avatar: 'svg:avatar-2',
    content: $t('pages.dashboard.workspace.trends.items.closedRunProject'),
    date: $t('pages.dashboard.workspace.trends.oneWeekAgo'),
    title: $t('pages.dashboard.workspace.trends.people.jack'),
  },
  {
    avatar: 'svg:avatar-3',
    content: $t('pages.dashboard.workspace.trends.items.postedUpdate'),
    date: $t('pages.dashboard.workspace.trends.oneWeekAgo'),
    title: $t('pages.dashboard.workspace.trends.people.william'),
  },
  {
    avatar: 'svg:avatar-4',
    content: $t('pages.dashboard.workspace.trends.items.pushedGithub'),
    date: '2021-04-01 20:00',
    title: $t('pages.dashboard.workspace.trends.people.william'),
  },
  {
    avatar: 'svg:avatar-4',
    content: $t('pages.dashboard.workspace.trends.items.adminVbenArticle'),
    date: '2021-03-01 20:00',
    title: 'Vben',
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
