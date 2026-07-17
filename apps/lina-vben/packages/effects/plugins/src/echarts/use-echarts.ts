import type { EChartsOption } from 'echarts';

import type { Ref } from 'vue';

import type { Nullable } from '@vben/types';

import type EchartsUI from './echarts-ui.vue';

import {
  computed,
  nextTick,
  onActivated,
  onBeforeUnmount,
  onDeactivated,
  onMounted,
  ref,
  unref,
  watch,
} from 'vue';

import { usePreferences } from '@vben/preferences';
import { cloneDeep } from '@vben/utils';

import {
  tryOnUnmounted,
  useDebounceFn,
  useResizeObserver,
  useTimeoutFn,
  useWindowSize,
} from '@vueuse/core';

import echarts from './echarts';

type EchartsUIType = typeof EchartsUI | undefined;

interface EchartsSeriesDataPatch {
  series: Array<{
    data: unknown;
    id: number | string;
  }>;
}

interface UpdateData {
  (
    option: EchartsSeriesDataPatch,
    notMerge?: false,
    lazyUpdate?: boolean,
  ): Promise<echarts.ECharts | null>;
  (
    option: EChartsOption,
    notMerge: true,
    lazyUpdate?: boolean,
  ): Promise<echarts.ECharts | null>;
}

function useEcharts(chartRef: Ref<EchartsUIType>) {
  let chartInstance: echarts.ECharts | null = null;
  let appliedDark = false;
  let canonicalOptions: EChartsOption | null = null;
  // echarts是否处于激活状态
  const isActiveRef = ref(false);

  const { isDark } = usePreferences();
  const { height, width } = useWindowSize();
  const resizeHandler: () => void = useDebounceFn(resize, 200);

  const getChartEl = (): HTMLElement | null => {
    const refValue = chartRef?.value as unknown;
    if (!refValue) return null;
    if (refValue instanceof HTMLElement) {
      return refValue;
    }
    const maybeComponent = refValue as { $el?: HTMLElement };
    return maybeComponent.$el ?? null;
  };

  onMounted(() => (isActiveRef.value = true));
  onActivated(() => (isActiveRef.value = true));
  onDeactivated(() => (isActiveRef.value = false));
  onBeforeUnmount(() => (isActiveRef.value = false));

  const isElHidden = (el: HTMLElement | null): boolean => {
    if (!el) return true;
    return el.offsetHeight === 0 || el.offsetWidth === 0;
  };

  const getOptions = computed((): EChartsOption => {
    if (!isDark.value) {
      return {};
    }

    return {
      backgroundColor: 'transparent',
    };
  });

  const initCharts = () => {
    const el = getChartEl();
    if (!el) {
      return;
    }
    chartInstance = echarts.init(el, isDark.value ? 'dark' : 'default');
    appliedDark = isDark.value;

    return chartInstance;
  };

  const renderEcharts = (
    options: EChartsOption,
  ): Promise<Nullable<echarts.ECharts>> => {
    if (!unref(isActiveRef)) {
      return Promise.resolve(null);
    }
    return new Promise((resolve, reject) => {
      if (chartRef.value?.offsetHeight === 0) {
        useTimeoutFn(async () => {
          try {
            resolve(await renderEcharts(options));
          } catch (error) {
            reject(error);
          }
        }, 30);
        return;
      }
      nextTick(() => {
        const el = getChartEl();
        if (isElHidden(el)) {
          useTimeoutFn(async () => {
            try {
              resolve(await renderEcharts(options));
            } catch (error) {
              reject(error);
            }
          }, 30);
          return;
        }
        useTimeoutFn(() => {
          try {
            if (!chartInstance || chartInstance?.getDom() !== el) {
              chartInstance?.dispose();
              const instance = initCharts();
              if (!instance) return;
            }
            const commitCanonicalOptions = prepareCanonicalOptions(
              options,
              true,
            );
            chartInstance?.setOption(
              {
                ...options,
                ...getOptions.value,
              },
              { notMerge: true },
            );
            commitCanonicalOptions();
            resolve(chartInstance);
          } catch (error) {
            reject(error);
          }
        }, 30);
      });
    });
  };

  const updateData: UpdateData = (
    option: EChartsOption | EchartsSeriesDataPatch,
    notMerge = false, // false = 合并（保留动画），true = 完全替换
    lazyUpdate = false, // true 时不立即重绘，适合短时间内多次调用
  ): Promise<echarts.ECharts | null> => {
    const chartOption = option as EChartsOption;
    return new Promise((resolve, reject) => {
      nextTick(() => {
        try {
          if (!chartInstance) {
            if (!notMerge) {
              throw new Error(
                'Incremental chart data requires an existing complete option baseline.',
              );
            }
            renderEcharts(chartOption).then(resolve, reject);
            return;
          }

          // 合并你原有的全局配置（比如 backgroundColor）
          const finalOption = {
            ...chartOption,
            ...getOptions.value,
          };
          const commitCanonicalOptions = prepareCanonicalOptions(
            chartOption,
            notMerge,
          );

          chartInstance.setOption(finalOption, {
            notMerge,
            lazyUpdate,
            // silent: true,     // 如果追求极致性能可开启（关闭所有事件）
          });
          commitCanonicalOptions();

          resolve(chartInstance);
        } catch (error) {
          reject(error);
        }
      });
    });
  };

  function resize() {
    const el = getChartEl();
    if (isElHidden(el)) {
      return;
    }
    chartInstance?.resize({
      animation: {
        duration: 300,
        easing: 'quadraticIn',
      },
    });
  }

  watch([width, height], () => {
    resizeHandler?.();
  });

  useResizeObserver(chartRef as never, resizeHandler);

  watch([isDark, isActiveRef], ([dark, isActive]) => {
    if (!chartInstance || !isActive || dark === appliedDark) {
      return;
    }

    if (!canonicalOptions) {
      return;
    }

    const existingAnimators = collectChartAnimators(chartInstance);
    // 先把完整业务配置提交为新的主题恢复基线，再由 setTheme 统一绘制目标主题。
    chartInstance.setOption(
      {
        ...canonicalOptions,
        ...getOptions.value,
      },
      {
        lazyUpdate: true,
        notMerge: true,
        silent: true,
      },
    );
    chartInstance.setTheme(dark ? 'dark' : 'default', { silent: true });
    // 只将本次换肤新增的 update 动画推进到终态，再统一提交快照像素。
    // 切换前已运行的业务动画与循环动画保持运行。
    finishThemeAnimations(chartInstance, existingAnimators);
    appliedDark = dark;
  });

  function prepareCanonicalOptions(options: EChartsOption, notMerge: boolean) {
    if (notMerge) {
      const nextCanonicalOptions = cloneDeep(options);
      return () => {
        canonicalOptions = nextCanonicalOptions;
      };
    }

    if (!canonicalOptions) {
      throw new Error(
        'Incremental chart data requires an existing complete option baseline.',
      );
    }

    const updates = resolveSeriesDataUpdates(canonicalOptions, options);
    return () => {
      updates.forEach(({ data, series }) => {
        series.data = data;
      });
    };
  }

  tryOnUnmounted(() => {
    // 销毁实例，释放资源
    chartInstance?.dispose();
  });
  return {
    isActive: isActiveRef,
    renderEcharts,
    resize,
    updateData,
    getChartInstance: () => chartInstance,
  };
}

type ChartRenderer = ReturnType<echarts.ECharts['getZr']>;
type RendererElement = ReturnType<ChartRenderer['storage']['getRoots']>[number];
type RendererAnimator = RendererElement['animators'][number];

function forEachChartElement(
  renderer: ChartRenderer,
  callback: (element: RendererElement) => void,
) {
  const visited = new Set<RendererElement>();
  const visit = (element: null | RendererElement | undefined) => {
    if (!element || visited.has(element)) {
      return;
    }
    visited.add(element);
    callback(element);
    visit(element.getTextContent());
    visit(element.getTextGuideLine());
    visit(element.getClipPath());
  };

  renderer.storage.getRoots().forEach((root) => {
    visit(root);
    root.traverse(visit);
  });
}

function collectChartAnimators(chartInstance: echarts.ECharts) {
  const renderer = chartInstance.getZr();
  const animators = new Set<RendererAnimator>();
  forEachChartElement(renderer, (element) => {
    element.animators.forEach((animator) => animators.add(animator));
  });
  return animators;
}

function finishThemeAnimations(
  chartInstance: echarts.ECharts,
  existingAnimators: Set<RendererAnimator>,
) {
  const renderer = chartInstance.getZr();
  forEachChartElement(renderer, (element) => {
    [...element.animators].forEach((animator) => {
      if (!existingAnimators.has(animator) && animator.scope === 'update') {
        animator.stop(true);
      }
    });
  });
  renderer.flush();
}

type MutableSeriesOption = Record<string, unknown> & {
  data?: unknown;
  id?: unknown;
};

interface SeriesDataUpdate {
  data: unknown;
  series: MutableSeriesOption;
}

const INCREMENTAL_DATA_ERROR =
  'Incremental chart updates must contain only series data with unique, stable ids; pass notMerge=true for a complete option replacement.';

function resolveSeriesDataUpdates(
  canonicalOptions: EChartsOption,
  patch: EChartsOption,
): SeriesDataUpdate[] {
  const patchRecord = patch as Record<string, unknown>;
  const patchKeys = Object.keys(patchRecord);
  if (patchKeys.length !== 1 || patchKeys[0] !== 'series') {
    throw new Error(INCREMENTAL_DATA_ERROR);
  }

  const canonicalSeries = normalizeSeriesOptions(
    (canonicalOptions as Record<string, unknown>).series,
  );
  const patchSeries = normalizeSeriesOptions(patchRecord.series);
  if (patchSeries.length === 0) {
    throw new Error(INCREMENTAL_DATA_ERROR);
  }

  const canonicalById = new Map<string, MutableSeriesOption | null>();
  canonicalSeries.forEach((series) => {
    if (!isSeriesOption(series) || !isStableId(series.id)) {
      return;
    }
    const id = String(series.id);
    canonicalById.set(id, canonicalById.has(id) ? null : series);
  });

  const patchIds = new Set<string>();
  return patchSeries.map((series) => {
    if (
      !isSeriesOption(series) ||
      !isStableId(series.id) ||
      !Object.prototype.hasOwnProperty.call(series, 'data')
    ) {
      throw new Error(INCREMENTAL_DATA_ERROR);
    }
    const seriesKeys = Object.keys(series);
    if (
      seriesKeys.length !== 2 ||
      !seriesKeys.includes('id') ||
      !seriesKeys.includes('data')
    ) {
      throw new Error(INCREMENTAL_DATA_ERROR);
    }

    const id = String(series.id);
    const canonicalSeriesOption = canonicalById.get(id);
    if (!canonicalSeriesOption || patchIds.has(id)) {
      throw new Error(INCREMENTAL_DATA_ERROR);
    }
    patchIds.add(id);

    return {
      data: cloneDeep(series.data),
      series: canonicalSeriesOption,
    };
  });
}

function isSeriesOption(value: unknown): value is MutableSeriesOption {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function isStableId(value: unknown) {
  return (
    (typeof value === 'string' || typeof value === 'number') &&
    String(value).trim() !== ''
  );
}

function normalizeSeriesOptions(value: unknown): unknown[] {
  if (Array.isArray(value)) {
    return value;
  }
  return value === undefined || value === null ? [] : [value];
}

export { useEcharts };

export type { EchartsSeriesDataPatch, EchartsUIType };
