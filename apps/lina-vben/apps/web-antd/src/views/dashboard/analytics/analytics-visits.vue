<script lang="ts" setup>
import type { EchartsUIType } from '@vben/plugins/echarts';

import { EchartsUI, useEcharts } from '@vben/plugins/echarts';
import { preferences } from '@vben/preferences';
import { onMounted, ref, watch } from 'vue';

const chartRef = ref<EchartsUIType>();
const { renderEcharts } = useEcharts(chartRef);

function buildMonthLabels() {
  const formatter = new Intl.DateTimeFormat(preferences.app.locale, {
    month: 'short',
  });
  return Array.from({ length: 12 }).map((_item, index) =>
    formatter.format(new Date(2026, index, 1)),
  );
}

function renderChart() {
  renderEcharts({
    grid: {
      bottom: 0,
      containLabel: true,
      left: '1%',
      right: '1%',
      top: '2 %',
    },
    series: [
      {
        barMaxWidth: 80,
        data: [
          3000, 2000, 3333, 5000, 3200, 4200, 3200, 2100, 3000, 5100, 6000,
          3200, 4800,
        ],
        type: 'bar',
      },
    ],
    tooltip: {
      axisPointer: {
        lineStyle: {
          width: 1,
        },
      },
      trigger: 'axis',
    },
    xAxis: {
      data: buildMonthLabels(),
      type: 'category',
    },
    yAxis: {
      max: 8000,
      splitNumber: 4,
      type: 'value',
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
