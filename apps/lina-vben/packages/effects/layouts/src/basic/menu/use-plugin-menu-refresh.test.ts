import type { MenuRecordRaw } from '@vben/types';

import { createApp, defineComponent, h, nextTick } from 'vue';

import { updatePreferences } from '@vben/preferences';
import { initStores, useAccessStore } from '@vben/stores';

import { mount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { useExtraMenu } from './use-extra-menu';
import { useMixedMenu } from './use-mixed-menu';

const routeMock = vi.hoisted(() => ({
  meta: {} as Record<string, unknown>,
  path: '/system/plugin',
}));

const routerMock = vi.hoisted(() => ({
  afterEach: vi.fn(),
  getRoutes: vi.fn(() => []),
  push: vi.fn(),
  resolve: vi.fn((path: string) => ({ href: path })),
}));

let storeNamespaceIndex = 0;

vi.mock('vue-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-router')>();
  return {
    ...actual,
    useRoute: () => routeMock,
    useRouter: () => routerMock,
  };
});

function extensionMenus(): MenuRecordRaw[] {
  return [
    {
      children: [
        {
          name: 'Plugin Management',
          parent: '/extension',
          parents: ['/extension'],
          path: '/system/plugin',
        },
      ],
      name: 'Extension Center',
      path: '/extension',
    },
  ];
}

function mountMixedMenu() {
  let composable: ReturnType<typeof useMixedMenu> | undefined;
  const wrapper = mount(
    defineComponent({
      setup() {
        composable = useMixedMenu();
        return () => h('div');
      },
    }),
  );
  if (!composable) {
    throw new Error('useMixedMenu was not mounted');
  }
  return { composable, wrapper };
}

function mountExtraMenu() {
  let composable: ReturnType<typeof useExtraMenu> | undefined;
  const wrapper = mount(
    defineComponent({
      setup() {
        composable = useExtraMenu();
        return () => h('div');
      },
    }),
  );
  if (!composable) {
    throw new Error('useExtraMenu was not mounted');
  }
  return { composable, wrapper };
}

describe('plugin menu refresh layout state', () => {
  beforeEach(async () => {
    await initStores(createApp({}), {
      namespace: `plugin-menu-refresh-${storeNamespaceIndex}`,
    });
    storeNamespaceIndex += 1;
    routeMock.path = '/system/plugin';
    routeMock.meta = {};
    routerMock.afterEach.mockClear();
    routerMock.getRoutes.mockClear();
    routerMock.push.mockClear();
    routerMock.resolve.mockClear();
    useAccessStore().setAccessMenus([]);
  });

  afterEach(() => {
    updatePreferences({
      app: { layout: 'sidebar-nav' },
      navigation: { split: true },
    });
  });

  it('recomputes split sidebar menus when plugin lifecycle refresh replaces access menus', async () => {
    updatePreferences({
      app: { layout: 'mixed-nav' },
      navigation: { split: true },
    });

    const { composable, wrapper } = mountMixedMenu();
    expect(composable.sidebarMenus.value).toEqual([]);

    useAccessStore().setAccessMenus(extensionMenus());
    await nextTick();

    expect(composable.sidebarMenus.value.map((item) => item.path)).toEqual([
      '/system/plugin',
    ]);
    expect(composable.headerActive.value).toBe('/extension');

    wrapper.unmount();
  });

  it('recomputes sidebar mixed extra menus when plugin lifecycle refresh replaces access menus', async () => {
    updatePreferences({
      app: { layout: 'sidebar-mixed-nav' },
      navigation: { split: true },
    });

    const { composable, wrapper } = mountExtraMenu();
    expect(composable.extraMenus.value).toEqual([]);

    useAccessStore().setAccessMenus(extensionMenus());
    await nextTick();

    expect(composable.extraMenus.value.map((item) => item.path)).toEqual([
      '/system/plugin',
    ]);
    expect(composable.extraActiveMenu.value).toBe('/extension');

    wrapper.unmount();
  });
});
