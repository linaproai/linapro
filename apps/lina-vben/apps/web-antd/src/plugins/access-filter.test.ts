import type { PluginDynamicState } from '#/api/system/plugin/model';
import type { RouteRecordStringComponent } from '@vben/types';

import { getPluginStateMap } from '#/plugins/slot-registry';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { filterDisabledPluginRoutes } from './access-filter';

vi.mock('#/plugins/slot-registry', () => ({
  getPluginStateMap: vi.fn(),
}));

const cachedDisabledPluginState = new Map<string, PluginDynamicState>([
  [
    'linapro-demo-source',
    {
      enabled: 0,
      generation: 2,
      id: 'linapro-demo-source',
      installed: 1,
      runtimeState: 'normal',
      statusKey: 'sys_plugin.status:linapro-demo-source',
      version: 'v0.1.0',
    },
  ],
]);

const freshEnabledPluginState = new Map<string, PluginDynamicState>([
  [
    'linapro-demo-source',
    {
      enabled: 1,
      generation: 3,
      id: 'linapro-demo-source',
      installed: 1,
      runtimeState: 'normal',
      statusKey: 'sys_plugin.status:linapro-demo-source',
      version: 'v0.1.0',
    },
  ],
]);

describe('plugin access route filtering', () => {
  beforeEach(() => {
    vi.mocked(getPluginStateMap).mockResolvedValue(cachedDisabledPluginState);
  });

  it('uses the supplied plugin state snapshot when refreshing dynamic routes', async () => {
    const routes: RouteRecordStringComponent[] = [
      {
        component: 'system/plugin/dynamic-page',
        meta: {
          authority: ['linapro-demo-source'],
          title: 'Source Plugin Demo',
        },
        name: 'LinaproDemoSourceSidebarEntry',
        path: 'linapro-demo-source-sidebar-entry',
      },
    ];

    await expect(
      filterDisabledPluginRoutes(routes),
    ).resolves.toHaveLength(0);
    await expect(
      filterDisabledPluginRoutes(routes, freshEnabledPluginState),
    ).resolves.toHaveLength(1);
    expect(getPluginStateMap).toHaveBeenCalledTimes(1);
  });
});
