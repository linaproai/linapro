import type { Ref } from 'vue';

import type { EchartsSeriesDataPatch, EchartsUIType } from './use-echarts';

import { mount } from '@vue/test-utils';
import { defineComponent, h, nextTick, ref } from 'vue';

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useEcharts } from './use-echarts';

const preferencesMocks = vi.hoisted(() => ({
  usePreferences: vi.fn(),
}));

const echartsMocks = vi.hoisted(() => ({
  init: vi.fn(),
}));

const vueUseMocks = vi.hoisted(() => ({
  useResizeObserver: vi.fn(),
}));

vi.mock('@vben/preferences', () => ({
  usePreferences: preferencesMocks.usePreferences,
}));

vi.mock('@vueuse/core', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@vueuse/core')>();
  return {
    ...actual,
    useResizeObserver: vueUseMocks.useResizeObserver,
  };
});

vi.mock('./echarts', () => ({
  default: {
    init: echartsMocks.init,
  },
}));

function createVisibleElement() {
  const element = document.createElement('div');
  Object.defineProperties(element, {
    offsetHeight: { configurable: true, value: 400 },
    offsetWidth: { configurable: true, value: 800 },
  });
  return element;
}

interface RendererAnimatorMock {
  scope?: string;
  stop: ReturnType<typeof vi.fn>;
}

function mountUseEcharts(initialDark: boolean) {
  const isDark = ref(initialDark);
  const element = createVisibleElement();
  const flush = vi.fn();
  const existingUpdateAnimator: RendererAnimatorMock = {
    scope: 'update',
    stop: vi.fn(),
  };
  const businessLoopAnimator: RendererAnimatorMock = {
    scope: 'business-loop',
    stop: vi.fn(),
  };
  const themeAnimators: RendererAnimatorMock[] = [];
  const rendererElement = {
    animators: [existingUpdateAnimator, businessLoopAnimator],
    getClipPath: vi.fn(() => null),
    getTextContent: vi.fn(() => null),
    getTextGuideLine: vi.fn(() => null),
    traverse: vi.fn(),
  };
  const getRoots = vi.fn(() => [rendererElement]);
  const setTheme = vi.fn(() => {
    const themeAnimator: RendererAnimatorMock = {
      scope: 'update',
      stop: vi.fn(),
    };
    themeAnimator.stop.mockImplementation(() => {
      const index = rendererElement.animators.indexOf(themeAnimator);
      if (index !== -1) {
        rendererElement.animators.splice(index, 1);
      }
    });
    themeAnimators.push(themeAnimator);
    rendererElement.animators.push(themeAnimator);
  });
  const chartInstance = {
    clear: vi.fn(),
    dispose: vi.fn(),
    getDom: vi.fn(() => element),
    getZr: vi.fn(() => ({
      flush,
      storage: { getRoots },
    })),
    resize: vi.fn(),
    setOption: vi.fn(),
    setTheme,
  };

  preferencesMocks.usePreferences.mockReturnValue({ isDark });
  echartsMocks.init.mockReturnValue(chartInstance);

  let composable: ReturnType<typeof useEcharts> | undefined;
  const chartRef = ref({ $el: element }) as unknown as Ref<EchartsUIType>;
  const wrapper = mount(
    defineComponent({
      setup() {
        composable = useEcharts(chartRef);
        return () => h('div');
      },
    }),
  );

  if (!composable) {
    throw new Error('useEcharts was not mounted');
  }

  return {
    chartInstance,
    businessLoopAnimator,
    composable,
    element,
    existingUpdateAnimator,
    flush,
    getRoots,
    isDark,
    themeAnimators,
    wrapper,
  };
}

async function renderChart(composable: ReturnType<typeof useEcharts>) {
  const renderPromise = composable.renderEcharts({
    series: [{ data: [1, 2, 3], id: 'series-1', type: 'line' }],
  });
  await nextTick();
  await vi.runAllTimersAsync();
  return await renderPromise;
}

describe('useEcharts theme switching', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('initializes the chart with the current dark theme', async () => {
    const fixture = mountUseEcharts(true);

    try {
      const renderedInstance = await renderChart(fixture.composable);

      expect(renderedInstance).toBe(fixture.chartInstance);
      expect(echartsMocks.init).toHaveBeenCalledTimes(1);
      expect(echartsMocks.init).toHaveBeenCalledWith(fixture.element, 'dark');
      expect(fixture.chartInstance.clear).not.toHaveBeenCalled();
      expect(fixture.chartInstance.setOption).toHaveBeenCalledWith(
        expect.objectContaining({
          series: [{ data: [1, 2, 3], id: 'series-1', type: 'line' }],
        }),
        { notMerge: true },
      );
    } finally {
      fixture.wrapper.unmount();
    }
  });

  // 回归：https://github.com/vbenjs/vue-vben-admin/issues/6919
  it('reuses the instance and applies the dynamic theme', async () => {
    const fixture = mountUseEcharts(false);

    try {
      await renderChart(fixture.composable);
      const initialInstance = fixture.composable.getChartInstance();
      const initialSetOptionCalls =
        fixture.chartInstance.setOption.mock.calls.length;
      const initialResizeCalls = fixture.chartInstance.resize.mock.calls.length;

      expect(echartsMocks.init).toHaveBeenCalledWith(
        fixture.element,
        'default',
      );

      fixture.isDark.value = true;
      await nextTick();

      expect(fixture.composable.getChartInstance()).toBe(initialInstance);
      expect(fixture.chartInstance.setTheme).toHaveBeenLastCalledWith('dark', {
        silent: true,
      });
      expect(fixture.chartInstance.setOption).toHaveBeenNthCalledWith(
        initialSetOptionCalls + 1,
        {
          backgroundColor: 'transparent',
          series: [{ data: [1, 2, 3], id: 'series-1', type: 'line' }],
        },
        {
          lazyUpdate: true,
          notMerge: true,
          silent: true,
        },
      );
      expect(fixture.flush).toHaveBeenCalledOnce();
      expect(fixture.getRoots).toHaveBeenCalledTimes(2);
      expect(fixture.themeAnimators).toHaveLength(1);
      expect(fixture.themeAnimators[0]?.stop).toHaveBeenCalledOnce();
      expect(fixture.themeAnimators[0]?.stop).toHaveBeenCalledWith(true);
      expect(fixture.existingUpdateAnimator.stop).not.toHaveBeenCalled();
      expect(fixture.businessLoopAnimator.stop).not.toHaveBeenCalled();
      expect(
        fixture.chartInstance.setTheme.mock.invocationCallOrder[0],
      ).toBeLessThan(fixture.getRoots.mock.invocationCallOrder[1] ?? 0);
      expect(fixture.getRoots.mock.invocationCallOrder[1]).toBeLessThan(
        fixture.flush.mock.invocationCallOrder[0] ?? 0,
      );

      fixture.isDark.value = false;
      await nextTick();

      expect(fixture.composable.getChartInstance()).toBe(initialInstance);
      expect(fixture.chartInstance.setTheme.mock.calls).toEqual([
        ['dark', { silent: true }],
        ['default', { silent: true }],
      ]);
      expect(fixture.chartInstance.setOption).toHaveBeenNthCalledWith(
        initialSetOptionCalls + 2,
        {
          series: [{ data: [1, 2, 3], id: 'series-1', type: 'line' }],
        },
        {
          lazyUpdate: true,
          notMerge: true,
          silent: true,
        },
      );
      expect(echartsMocks.init).toHaveBeenCalledTimes(1);
      expect(fixture.chartInstance.dispose).not.toHaveBeenCalled();
      expect(fixture.chartInstance.setOption).toHaveBeenCalledTimes(
        initialSetOptionCalls + 2,
      );
      expect(fixture.flush).toHaveBeenCalledTimes(2);
      expect(fixture.getRoots).toHaveBeenCalledTimes(4);
      expect(fixture.themeAnimators).toHaveLength(2);
      fixture.themeAnimators.forEach(({ stop }) => {
        expect(stop).toHaveBeenCalledOnce();
        expect(stop).toHaveBeenCalledWith(true);
      });
      expect(fixture.existingUpdateAnimator.stop).not.toHaveBeenCalled();
      expect(fixture.businessLoopAnimator.stop).not.toHaveBeenCalled();
      expect(fixture.chartInstance.resize).toHaveBeenCalledTimes(
        initialResizeCalls,
      );
    } finally {
      fixture.wrapper.unmount();
    }
  });

  it('compacts incremental data into the complete theme baseline', async () => {
    const fixture = mountUseEcharts(false);

    try {
      await renderChart(fixture.composable);
      await fixture.composable.updateData({
        series: [{ data: [4, 5, 6], id: 'series-1' }],
      });
      await fixture.composable.updateData({
        series: [{ data: [7, 8, 9], id: 'series-1' }],
      });
      const callsBeforeTheme =
        fixture.chartInstance.setOption.mock.calls.length;

      fixture.isDark.value = true;
      await nextTick();

      expect(fixture.chartInstance.setTheme).toHaveBeenCalledWith('dark', {
        silent: true,
      });
      expect(fixture.chartInstance.setOption).toHaveBeenNthCalledWith(
        callsBeforeTheme + 1,
        {
          backgroundColor: 'transparent',
          series: [{ data: [7, 8, 9], id: 'series-1', type: 'line' }],
        },
        {
          lazyUpdate: true,
          notMerge: true,
          silent: true,
        },
      );
      expect(fixture.chartInstance.setOption).toHaveBeenCalledTimes(
        callsBeforeTheme + 1,
      );
    } finally {
      fixture.wrapper.unmount();
    }
  });

  it('rejects ambiguous incremental option patches', async () => {
    const fixture = mountUseEcharts(false);

    try {
      await renderChart(fixture.composable);

      await expect(
        fixture.composable.updateData({
          series: [{ data: [4, 5, 6] }],
        } as unknown as EchartsSeriesDataPatch),
      ).rejects.toThrow(/unique, stable ids/u);
    } finally {
      fixture.wrapper.unmount();
    }
  });

  it('requires a complete baseline before incremental data', async () => {
    const fixture = mountUseEcharts(false);

    try {
      await expect(
        fixture.composable.updateData({
          series: [{ data: [4, 5, 6], id: 'series-1' }],
        }),
      ).rejects.toThrow(/existing complete option baseline/u);
    } finally {
      fixture.wrapper.unmount();
    }
  });

  it('defers theme work for deactivated charts until activation', async () => {
    const fixture = mountUseEcharts(false);

    try {
      await renderChart(fixture.composable);
      fixture.composable.isActive.value = false;
      fixture.isDark.value = true;
      await nextTick();

      expect(fixture.chartInstance.setTheme).not.toHaveBeenCalled();

      fixture.composable.isActive.value = true;
      await nextTick();

      expect(fixture.chartInstance.setTheme).toHaveBeenCalledOnce();
      expect(fixture.chartInstance.setTheme).toHaveBeenCalledWith('dark', {
        silent: true,
      });
    } finally {
      fixture.wrapper.unmount();
    }
  });
});
