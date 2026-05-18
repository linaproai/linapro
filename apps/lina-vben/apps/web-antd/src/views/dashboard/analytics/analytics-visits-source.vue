<script lang="ts" setup>
import type { EchartsUIType } from '@vben/plugins/echarts';

import { preferences } from '@vben/preferences';
import { EchartsUI, useEcharts } from '@vben/plugins/echarts';
import { onMounted, ref, watch } from 'vue';

import { $t } from '#/locales';

const chartRef = ref<EchartsUIType>();
const { renderEcharts } = useEcharts(chartRef);

function renderChart() {
  renderEcharts({
    legend: {
      bottom: '2%',
      left: 'center',
    },
    series: [
      {
        animationDelay() {
          return Math.random() * 100;
        },
        animationEasing: 'exponentialInOut',
        animationType: 'scale',
        avoidLabelOverlap: false,
        color: ['#5ab1ef', '#b6a2de', '#67e0e3', '#2ec7c9'],
        data: [
          { name: $t('pages.dashboard.analytics.sources.search'), value: 1048 },
          { name: $t('pages.dashboard.analytics.sources.direct'), value: 735 },
          { name: $t('pages.dashboard.analytics.sources.email'), value: 580 },
          { name: $t('pages.dashboard.analytics.sources.ads'), value: 484 },
        ],
        emphasis: {
          label: {
            fontSize: 12,
            fontWeight: 'bold',
            show: true,
          },
        },
        itemStyle: {
          borderRadius: 10,
          borderWidth: 2,
        },
        label: {
          position: 'center',
          show: false,
        },
        labelLine: {
          show: false,
        },
        name: $t('pages.dashboard.analytics.cards.sources'),
        radius: ['40%', '65%'],
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
