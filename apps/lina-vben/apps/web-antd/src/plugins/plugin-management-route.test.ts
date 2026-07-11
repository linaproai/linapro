import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { RegisteredPluginPage } from './page-registry';

import {
  flattenAccessMenuPaths,
  getPluginManagementPages,
  hasPluginManagementPage,
  pickPluginManagementPage,
  resolvePluginManagementPath,
} from './plugin-management-route';

const getPluginPages = vi.hoisted(() =>
  vi.fn<() => RegisteredPluginPage[]>(() => []),
);

vi.mock('./page-registry', () => ({
  getPluginPages: () => getPluginPages(),
}));

function page(
  partial: Partial<RegisteredPluginPage> &
    Pick<RegisteredPluginPage, 'key' | 'pluginId' | 'routePath'>,
): RegisteredPluginPage {
  return {
    capabilities: [],
    component: {} as RegisteredPluginPage['component'],
    filePath:
      partial.filePath ??
      `/lina-plugins/${partial.pluginId}/frontend/pages/x.vue`,
    title: partial.title ?? partial.routePath,
    ...partial,
  };
}

describe('plugin management route resolver', () => {
  beforeEach(() => {
    getPluginPages.mockReset();
    getPluginPages.mockReturnValue([]);
  });

  it('treats settings and management pages as navigable and skips modals', () => {
    getPluginPages.mockReturnValue([
      page({
        key: 'linapro-auth-ldap:settings',
        pluginId: 'linapro-auth-ldap',
        routePath: 'linapro-auth-ldap-settings',
      }),
      page({
        key: 'linapro-auth-ldap:components/settings-modal',
        pluginId: 'linapro-auth-ldap',
        routePath: 'linapro-auth-ldap-components-settings-modal',
      }),
      page({
        key: 'linapro-auth-ldap:settings-drawer',
        pluginId: 'linapro-auth-ldap',
        routePath: 'linapro-auth-ldap-settings-drawer',
      }),
    ]);

    expect(hasPluginManagementPage('linapro-auth-ldap')).toBe(true);
    expect(
      getPluginManagementPages('linapro-auth-ldap').map((item) => item.key),
    ).toEqual(['linapro-auth-ldap:settings']);
    expect(pickPluginManagementPage('linapro-auth-ldap')?.routePath).toBe(
      'linapro-auth-ldap-settings',
    );
  });

  it('picks the first access-menu path when a plugin has multiple pages', () => {
    // Intentionally register pages in reverse alphabetical / last-menu-first order.
    getPluginPages.mockReturnValue([
      page({
        key: 'linapro-ai-core:invocation-logs',
        pluginId: 'linapro-ai-core',
        routePath: '/ai/invocations',
      }),
      page({
        key: 'linapro-ai-core:model-management',
        pluginId: 'linapro-ai-core',
        routePath: '/ai/models',
      }),
      page({
        key: 'linapro-ai-core:provider-management',
        pluginId: 'linapro-ai-core',
        routePath: '/ai/providers',
      }),
      page({
        key: 'linapro-ai-core:tier-management',
        pluginId: 'linapro-ai-core',
        routePath: '/ai/tiers',
      }),
    ]);

    const accessMenus = [
      {
        name: 'AI',
        path: '/ai',
        children: [
          { name: 'Providers', path: '/ai/providers', children: [] },
          { name: 'Models', path: '/ai/models', children: [] },
          { name: 'Tiers', path: '/ai/tiers', children: [] },
          { name: 'Logs', path: '/ai/invocations', children: [] },
        ],
      },
    ] as any;

    expect(flattenAccessMenuPaths(accessMenus)).toEqual([
      '/ai',
      '/ai/providers',
      '/ai/models',
      '/ai/tiers',
      '/ai/invocations',
    ]);

    expect(
      resolvePluginManagementPath(
        'linapro-ai-core',
        {
          getRoutes: () => [
            { path: '/ai/invocations' },
            { path: '/ai/models' },
            { path: '/ai/providers' },
            { path: '/ai/tiers' },
          ],
        } as any,
        accessMenus,
      ),
    ).toBe('/ai/providers');

    expect(
      pickPluginManagementPage('linapro-ai-core', accessMenus)?.routePath,
    ).toBe('/ai/providers');
  });

  it('falls back to the first matching router registration when menus are empty', () => {
    getPluginPages.mockReturnValue([
      page({
        key: 'linapro-ai-core:invocation-logs',
        pluginId: 'linapro-ai-core',
        routePath: '/ai/invocations',
      }),
      page({
        key: 'linapro-ai-core:provider-management',
        pluginId: 'linapro-ai-core',
        routePath: '/ai/providers',
      }),
    ]);

    // Router registration order: providers first (first menu), invocations last.
    expect(
      resolvePluginManagementPath(
        'linapro-ai-core',
        {
          getRoutes: () => [
            { path: '/ai/providers' },
            { path: '/ai/invocations' },
          ],
        } as any,
        [],
      ),
    ).toBe('/ai/providers');
  });

  it('resolves full nested router paths and returns null when route is missing', () => {
    getPluginPages.mockReturnValue([
      page({
        key: 'linapro-auth-ldap:settings',
        pluginId: 'linapro-auth-ldap',
        routePath: 'linapro-auth-ldap-settings',
      }),
    ]);

    const nested = resolvePluginManagementPath('linapro-auth-ldap', {
      getRoutes: () => [
        { path: '/auth-login/linapro-auth-ldap-settings' },
        { path: '/system/plugin' },
      ],
    } as any);
    expect(nested).toBe('/auth-login/linapro-auth-ldap-settings');

    const missing = resolvePluginManagementPath('linapro-auth-ldap', {
      getRoutes: () => [{ path: '/system/plugin' }],
    } as any);
    expect(missing).toBeNull();
  });

  it('returns null when the plugin has no management page', () => {
    getPluginPages.mockReturnValue([
      page({
        key: 'other:settings',
        pluginId: 'other',
        routePath: 'other-settings',
      }),
    ]);

    expect(hasPluginManagementPage('missing-plugin')).toBe(false);
    expect(
      resolvePluginManagementPath(
        'missing-plugin',
        {
          getRoutes: () => [],
        } as any,
      ),
    ).toBeNull();
  });
});
