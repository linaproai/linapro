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
      bottom: 0,
      data: [
        $t('pages.dashboard.analytics.channels.visits'),
        $t('pages.dashboard.analytics.channels.trend'),
      ],
    },
    radar: {
      indicator: [
        {
          name: $t('pages.dashboard.analytics.channels.web'),
        },
        {
          name: $t('pages.dashboard.analytics.channels.mobile'),
        },
        {
          name: 'Ipad',
        },
        {
          name: $t('pages.dashboard.analytics.channels.client'),
        },
        {
          name: $t('pages.dashboard.analytics.channels.thirdParty'),
        },
        {
          name: $t('pages.dashboard.analytics.channels.other'),
        },
      ],
      radius: '60%',
      splitNumber: 8,
    },
    series: [
      {
        areaStyle: {
          opacity: 1,
          shadowBlur: 0,
          shadowColor: 'rgba(0,0,0,.2)',
          shadowOffsetX: 0,
          shadowOffsetY: 10,
        },
        data: [
          {
            itemStyle: {
              color: '#b6a2de',
            },
            name: $t('pages.dashboard.analytics.channels.visits'),
            value: [90, 50, 86, 40, 50, 20],
          },
          {
            itemStyle: {
              color: '#5ab1ef',
            },
            name: $t('pages.dashboard.analytics.channels.trend'),
            value: [70, 75, 70, 76, 20, 85],
          },
        ],
        itemStyle: {
          borderRadius: 10,
          borderWidth: 2,
        },
        symbolSize: 0,
        type: 'radar',
      },
    ],
    tooltip: {},
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
