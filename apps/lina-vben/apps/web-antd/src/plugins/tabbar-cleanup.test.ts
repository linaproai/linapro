import type { TabDefinition } from '@vben/types';

import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useTabbarStore } from '@vben/stores';

import { closePluginTabs } from './tabbar-cleanup';

vi.mock('./page-registry', () => ({
  getPluginPages: () => [
    {
      pluginId: 'media',
      routePath: 'media',
    },
    {
      pluginId: 'media-library',
      routePath: 'media-library',
    },
  ],
}));

function tab(
  path: string,
  title: string,
  meta: Record<string, unknown> = {},
) {
  const routeMeta = {
    ...meta,
    title,
  } as TabDefinition['meta'];
  return {
    fullPath: path,
    key: path,
    matched: [{ meta: routeMeta, path }],
    meta: routeMeta,
    name: title,
    path,
  } as TabDefinition;
}

describe('plugin tabbar cleanup', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('closes tabs that belong to the uninstalled plugin only', async () => {
    const tabbarStore = useTabbarStore();
    tabbarStore.tabs = [
      tab('/dashboard/analytics', 'Analytics', { affixTab: true }),
      tab('/system/plugin', 'Plugin Management'),
      tab('/media', 'Media'),
      tab('/media-library', 'Media Library'),
      tab('/watermark-service', 'Watermark', {
        authority: ['watermark-service:page'],
      }),
    ];

    await closePluginTabs('media');

    expect(tabbarStore.getTabs.map((item) => item.path)).toEqual([
      '/dashboard/analytics',
      '/system/plugin',
      '/media-library',
      '/watermark-service',
    ]);
  });

  it('closes dynamic asset tabs by embedded source plugin id', async () => {
    const tabbarStore = useTabbarStore();
    tabbarStore.tabs = [
      tab('/system/plugin', 'Plugin Management'),
      tab('/dynamic-page', 'Dynamic Media', {
        query: {
          embeddedSrc: '/x-assets/media/v0.1.0/pages/index.js',
        },
      }),
    ];

    await closePluginTabs('media');

    expect(tabbarStore.getTabs.map((item) => item.path)).toEqual([
      '/system/plugin',
    ]);
  });

  it('closes plugin tabs by string authority metadata', async () => {
    const tabbarStore = useTabbarStore();
    tabbarStore.tabs = [
      tab('/system/plugin', 'Plugin Management'),
      tab('/dynamic-media', 'Dynamic Media', {
        authority: 'media:page',
      }),
    ];

    await closePluginTabs('media');

    expect(tabbarStore.getTabs.map((item) => item.path)).toEqual([
      '/system/plugin',
    ]);
  });
});
