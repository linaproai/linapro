<script lang="ts" setup>
import type { EchartsUIType } from '@vben/plugins/echarts';

import { ref, watch } from 'vue';

import { EchartsUI, useEcharts } from '@vben/plugins/echarts';

import type { AnalyticsRadarItem } from './data';

const props = defineProps<{
  indicators: string[];
  label: string;
  series: AnalyticsRadarItem[];
}>();

const chartRef = ref<EchartsUIType>();
const { renderEcharts } = useEcharts(chartRef);

watch(
  () => props,
  (value) => {
    renderEcharts({
      legend: {
        bottom: 0,
        data: value.series.map((item) => item.name),
      },
      radar: {
        indicator: value.indicators.map((item) => ({ name: item })),
        radius: '60%',
        splitNumber: 5,
      },
      series: [
        {
          areaStyle: {
            opacity: 0.22,
            shadowBlur: 0,
            shadowColor: 'rgba(0,0,0,.2)',
            shadowOffsetX: 0,
            shadowOffsetY: 10,
          },
          data: value.series.map((item, index) => ({
            itemStyle: {
              color: index === 0 ? '#1677ff' : '#13c2c2',
            },
            name: item.name,
            value: item.value,
          })),
          itemStyle: {
            borderRadius: 10,
            borderWidth: 2,
          },
          name: value.label,
          symbolSize: 0,
          type: 'radar',
        },
      ],
      tooltip: {},
    });
  },
  { deep: true, immediate: true },
);
</script>

<template>
  <EchartsUI ref="chartRef" />
</template>
