<script lang="ts" setup>
import type { EchartsUIType } from '@vben/plugins/echarts';

import { preferences } from '@vben/preferences';
import { onMounted, ref, watch } from 'vue';

import { EchartsUI, useEcharts } from '@vben/plugins/echarts';

import { $t } from '#/locales';

const chartRef = ref<EchartsUIType>();
const { renderEcharts } = useEcharts(chartRef);

function renderChart() {
  renderEcharts({
    series: [
      {
        animationDelay() {
          return Math.random() * 400;
        },
        animationEasing: 'exponentialInOut',
        animationType: 'scale',
        center: ['50%', '50%'],
        color: ['#5ab1ef', '#b6a2de', '#67e0e3', '#2ec7c9'],
        data: [
          {
            name: $t('pages.dashboard.analytics.sales.outsourcing'),
            value: 500,
          },
          {
            name: $t('pages.dashboard.analytics.sales.customDevelopment'),
            value: 310,
          },
          {
            name: $t('pages.dashboard.analytics.sales.technicalSupport'),
            value: 274,
          },
          { name: $t('pages.dashboard.analytics.sales.remote'), value: 400 },
        ].sort((left, right) => {
          return left.value - right.value;
        }),
        name: $t('pages.dashboard.analytics.sales.title'),
        radius: '80%',
        roseType: 'radius',
        type: 'pie',
      },
    ],

    tooltip: {
      trigger: 'item',
    },
  });
}

onMounted(renderChart);

watch(
  () => preferences.app.locale,
  () => {
    renderChart();
  },
);
</script>

<template>
  <EchartsUI ref="chartRef" />
</template>
