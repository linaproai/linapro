<script lang="ts" setup>
import type { AnalysisOverviewItem } from '@vben/common-ui';
import type { TabOption } from '@vben/types';

import { computed } from 'vue';

import {
  AnalysisChartCard,
  AnalysisChartsTabs,
  AnalysisOverview,
} from '@vben/common-ui';
import {
  SvgBellIcon,
  SvgCakeIcon,
  SvgCardIcon,
  SvgDownloadIcon,
} from '@vben/icons';

import { $t } from '#/locales';

import AnalyticsTrends from './analytics-trends.vue';
import AnalyticsVisitsData from './analytics-visits-data.vue';
import AnalyticsVisitsSales from './analytics-visits-sales.vue';
import AnalyticsVisitsSource from './analytics-visits-source.vue';
import AnalyticsVisits from './analytics-visits.vue';

const overviewItems = computed<AnalysisOverviewItem[]>(() => [
  {
    icon: SvgCardIcon,
    title: $t('pages.dashboard.analytics.overview.users.title'),
    totalTitle: $t('pages.dashboard.analytics.overview.users.total'),
    totalValue: 120_000,
    value: 2000,
  },
  {
    icon: SvgCakeIcon,
    title: $t('pages.dashboard.analytics.overview.visits.title'),
    totalTitle: $t('pages.dashboard.analytics.overview.visits.total'),
    totalValue: 500_000,
    value: 20_000,
  },
  {
    icon: SvgDownloadIcon,
    title: $t('pages.dashboard.analytics.overview.downloads.title'),
    totalTitle: $t('pages.dashboard.analytics.overview.downloads.total'),
    totalValue: 120_000,
    value: 8000,
  },
  {
    icon: SvgBellIcon,
    title: $t('pages.dashboard.analytics.overview.usage.title'),
    totalTitle: $t('pages.dashboard.analytics.overview.usage.total'),
    totalValue: 50_000,
    value: 5000,
  },
]);

const chartTabs = computed<TabOption[]>(() => [
  {
    label: $t('pages.dashboard.analytics.tabs.trends'),
    value: 'trends',
  },
  {
    label: $t('pages.dashboard.analytics.tabs.visits'),
    value: 'visits',
  },
]);
</script>

<template>
  <div class="p-5" data-testid="dashboard-analytics-page">
    <div data-testid="dashboard-analytics-overview">
      <AnalysisOverview :items="overviewItems" />
    </div>
    <AnalysisChartsTabs
      :tabs="chartTabs"
      class="mt-5"
      data-testid="dashboard-analytics-tabs"
    >
      <template #trends>
        <AnalyticsTrends />
      </template>
      <template #visits>
        <AnalyticsVisits />
      </template>
    </AnalysisChartsTabs>

    <div class="mt-5 w-full md:flex">
      <AnalysisChartCard
        class="mt-5 md:mt-0 md:mr-4 md:w-1/3"
        data-testid="dashboard-analytics-visit-card"
        :title="$t('pages.dashboard.analytics.cards.channels')"
      >
        <AnalyticsVisitsData />
      </AnalysisChartCard>
      <AnalysisChartCard
        class="mt-5 md:mt-0 md:mr-4 md:w-1/3"
        data-testid="dashboard-analytics-source-card"
        :title="$t('pages.dashboard.analytics.cards.sources')"
      >
        <AnalyticsVisitsSource />
      </AnalysisChartCard>
      <AnalysisChartCard
        class="mt-5 md:mt-0 md:w-1/3"
        data-testid="dashboard-analytics-sales-card"
        :title="$t('pages.dashboard.analytics.sales.title')"
      >
        <AnalyticsVisitsSales />
      </AnalysisChartCard>
    </div>
  </div>
</template>
