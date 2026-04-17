<script lang="ts" setup>
import type { AnalysisOverviewItem } from '@vben/common-ui';
import type { TabOption } from '@vben/types';

import { computed, ref } from 'vue';

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

import AnalyticsTrends from './analytics-trends.vue';
import AnalyticsVisitsData from './analytics-visits-data.vue';
import AnalyticsVisitsSales from './analytics-visits-sales.vue';
import AnalyticsVisitsSource from './analytics-visits-source.vue';
import AnalyticsVisits from './analytics-visits.vue';
import {
  analyticsRangeData,
  analyticsRangeOptions,
  type AnalyticsOverviewMetric,
  type AnalyticsOverviewMetricKey,
  type AnalyticsRangeKey,
} from './data';

const activeRange = ref<AnalyticsRangeKey>('week');

const iconMap: Record<AnalyticsOverviewMetricKey, AnalysisOverviewItem['icon']> = {
  hostCalls: SvgCardIcon,
  pluginActivity: SvgCakeIcon,
  regressionRuns: SvgBellIcon,
  workspaceVisits: SvgDownloadIcon,
};

const currentRange = computed(() => analyticsRangeData[activeRange.value]);

const overviewItems = computed<AnalysisOverviewItem[]>(() => {
  return currentRange.value.overview.map((item: AnalyticsOverviewMetric) => ({
    icon: iconMap[item.key],
    title: item.title,
    totalTitle: item.totalTitle,
    totalValue: item.totalValue,
    value: item.value,
  }));
});

const chartTabs: TabOption[] = [
  {
    label: '流量趋势',
    value: 'trends',
  },
  {
    label: '发布节奏',
    value: 'visits',
  },
];
</script>

<template>
  <div class="p-5" data-testid="dashboard-analytics-page">
    <section
      class="from-background to-background/80 border-border/60 rounded-3xl border bg-gradient-to-br p-6 shadow-sm"
      data-testid="dashboard-analytics-hero"
    >
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div class="max-w-3xl">
          <p class="text-foreground/50 text-sm">默认管理工作台 / 分析页</p>
          <h2 class="text-foreground mt-2 text-2xl font-semibold">
            宿主工作区运行概览
          </h2>
          <p class="text-foreground/70 mt-3 text-sm leading-6" data-testid="dashboard-analytics-summary">
            {{ currentRange.summary }}
          </p>
        </div>
        <div class="flex flex-col items-start gap-3 lg:items-end">
          <span class="text-foreground/50 text-sm">{{ currentRange.updatedAt }}</span>
          <div class="flex flex-wrap gap-2" data-testid="dashboard-analytics-range-group">
            <button
              v-for="item in analyticsRangeOptions"
              :key="item.value"
              :class="[
                'rounded-full border px-4 py-2 text-sm transition-colors',
                item.value === activeRange
                  ? 'border-primary bg-primary text-white shadow-sm'
                  : 'border-border/70 bg-background text-foreground/75 hover:border-primary/40 hover:text-foreground',
              ]"
              :data-testid="`dashboard-range-${item.value}`"
              type="button"
              @click="activeRange = item.value"
            >
              {{ item.label }}
            </button>
          </div>
        </div>
      </div>

      <div class="mt-5 grid gap-4 xl:grid-cols-3">
        <article
          v-for="insight in currentRange.insights"
          :key="insight.title"
          :class="[
            'rounded-2xl border px-4 py-4',
            insight.tone === 'emerald' && 'border-emerald-200 bg-emerald-50/80',
            insight.tone === 'amber' && 'border-amber-200 bg-amber-50/80',
            insight.tone === 'cyan' && 'border-cyan-200 bg-cyan-50/80',
          ]"
          data-testid="dashboard-analytics-insight"
        >
          <p class="text-foreground/55 text-xs uppercase tracking-[0.18em]">
            {{ insight.title }}
          </p>
          <p class="text-foreground mt-3 text-2xl font-semibold">{{ insight.value }}</p>
          <p class="text-foreground/70 mt-2 text-sm leading-6">{{ insight.description }}</p>
        </article>
      </div>
    </section>

    <div class="mt-5" data-testid="dashboard-analytics-overview">
      <AnalysisOverview :items="overviewItems" />
    </div>

    <AnalysisChartsTabs :tabs="chartTabs" class="mt-5" data-testid="dashboard-analytics-tabs">
      <template #trends>
        <AnalyticsTrends
          :axis="currentRange.trendAxis"
          :series="currentRange.trendSeries"
        />
      </template>
      <template #visits>
        <AnalyticsVisits
          :axis="currentRange.cadenceAxis"
          :label="currentRange.cadenceLabel"
          :series="currentRange.cadenceSeries"
        />
      </template>
    </AnalysisChartsTabs>

    <div class="mt-5 grid gap-5 xl:grid-cols-3">
      <AnalysisChartCard :title="currentRange.touchpointLabel" data-testid="dashboard-analytics-touchpoint-card">
        <AnalyticsVisitsData
          :indicators="currentRange.radarIndicators"
          :label="currentRange.touchpointLabel"
          :series="currentRange.radarSeries"
        />
      </AnalysisChartCard>
      <AnalysisChartCard title="来源结构" data-testid="dashboard-analytics-source-card">
        <AnalyticsVisitsSource :items="currentRange.sourceItems" />
      </AnalysisChartCard>
      <AnalysisChartCard title="交付构成" data-testid="dashboard-analytics-sales-card">
        <AnalyticsVisitsSales :items="currentRange.salesItems" />
      </AnalysisChartCard>
    </div>
  </div>
</template>
