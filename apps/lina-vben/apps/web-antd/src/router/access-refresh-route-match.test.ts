import type { RouteRecordRaw } from 'vue-router';

import { defineComponent } from 'vue';
import { describe, expect, it } from 'vitest';

import {
  resolveAccessRefreshNavigation,
  resolveAccessibleRouteRefreshTarget,
} from './access-refresh-route-match';

const routeComponent = defineComponent({
  name: 'RouteStub',
  render: () => null,
});

describe('access refresh route matching', () => {
  it('matches current routes by route name', () => {
    const routes: RouteRecordRaw[] = [
      {
        component: routeComponent,
        name: 'PluginManagement',
        path: '/extension/plugin',
      },
    ];

    expect(
      resolveAccessibleRouteRefreshTarget(routes, {
        matched: [{ name: 'Root', path: '/' }, { name: 'PluginManagement' }],
        path: '/extension/plugin',
      }).accessible,
    ).toBe(true);
  });

  it('keeps a regenerated plugin iframe route when the name changed but the path is still accessible', () => {
    const pluginRoutePath =
      '/link-100-x-assets-plugin-dev-dynamic-hot-upgrade-v0-2-0-index-html';
    const routes: RouteRecordRaw[] = [
      {
        component: routeComponent,
        name: 'Link100XAssetsPluginDevDynamicHotUpgradeV020IndexHtml',
        path: pluginRoutePath,
      },
    ];

    expect(
      resolveAccessibleRouteRefreshTarget(routes, {
        matched: [
          { name: 'Root', path: '/' },
          {
            name: 'Link99XAssetsPluginDevDynamicHotUpgradeV020IndexHtml',
            path: pluginRoutePath,
          },
        ],
        path: pluginRoutePath,
      }).accessible,
    ).toBe(true);
  });

  it('matches nested routes by normalized full path', () => {
    const routes: RouteRecordRaw[] = [
      {
        component: routeComponent,
        name: 'Extension',
        path: '/extension',
        children: [
          {
            component: routeComponent,
            name: 'DynamicPlugin',
            path: 'dynamic-plugin',
          },
        ],
      },
    ];

    expect(
      resolveAccessibleRouteRefreshTarget(routes, {
        matched: [
          { name: 'Root', path: '/' },
          { name: 'DynamicPluginOldName', path: '/extension/dynamic-plugin' },
        ],
        path: '/extension/dynamic-plugin',
      }).accessible,
    ).toBe(true);
  });

  it('returns an equivalent iframe route when a plugin menu route was rebuilt with a new internal path', () => {
    const oldRoutePath =
      '/link-100-x-assets-plugin-dev-dynamic-hot-upgrade-v0-2-0-index-html';
    const newRoutePath =
      '/link-101-x-assets-plugin-dev-dynamic-hot-upgrade-v0-2-0-index-html';
    const iframeSrc =
      '/x-assets/plugin-dev-dynamic-hot-upgrade/v0.2.0/index.html';
    const routes: RouteRecordRaw[] = [
      {
        component: routeComponent,
        meta: {
          iframeSrc,
          title: 'Dynamic Hot Upgrade Plugin',
        },
        name: 'Link101XAssetsPluginDevDynamicHotUpgradeV020IndexHtml',
        path: newRoutePath,
      },
    ];

    expect(
      resolveAccessibleRouteRefreshTarget(
        routes,
        {
          matched: [
            { name: 'Root', path: '/' },
            { name: 'FallbackNotFound', path: '/:path(.*)*' },
          ],
          path: oldRoutePath,
        },
        {
          matched: [
            { name: 'Root', path: '/' },
            {
              meta: {
                iframeSrc,
                title: 'Dynamic Hot Upgrade Plugin',
              },
              name: 'Link100XAssetsPluginDevDynamicHotUpgradeV020IndexHtml',
              path: oldRoutePath,
            },
          ],
          meta: {
            iframeSrc,
            title: 'Dynamic Hot Upgrade Plugin',
          },
          path: oldRoutePath,
        },
      ),
    ).toEqual({
      accessible: true,
      replacementPath: newRoutePath,
    });
  });

  it('rejects a route that only resolves to fallback records', () => {
    const routes: RouteRecordRaw[] = [
      {
        component: routeComponent,
        name: 'Dashboard',
        path: '/dashboard',
      },
    ];

    expect(
      resolveAccessibleRouteRefreshTarget(routes, {
        matched: [
          { name: 'Root', path: '/' },
          { name: 'FallbackNotFound', path: '/:path(.*)*' },
        ],
        path: '/plugin-dev-dynamic-hot-upgrade',
      }).accessible,
    ).toBe(false);
  });

  it('keeps the current plugin iframe route during a failed upgrade while the plugin remains installed and enabled', () => {
    const iframeSrc =
      '/x-assets/plugin-dev-dynamic-hot-upgrade/v0.2.0/index.html';
    const routes: RouteRecordRaw[] = [
      {
        component: routeComponent,
        name: 'Dashboard',
        path: '/dashboard',
      },
    ];
    const pluginStateMap = new Map([
      [
        'plugin-dev-dynamic-hot-upgrade',
        {
          enabled: 1,
          generation: 4,
          id: 'plugin-dev-dynamic-hot-upgrade',
          installed: 1,
          runtimeState: 'upgrade_failed',
          statusKey: 'sys_plugin.status:plugin-dev-dynamic-hot-upgrade',
          version: 'v0.2.0',
        },
      ],
    ]);

    expect(
      resolveAccessibleRouteRefreshTarget(
        routes,
        {
          matched: [
            { name: 'Root', path: '/' },
            { name: 'FallbackNotFound', path: '/:path(.*)*' },
          ],
          path: '/link-102-x-assets-plugin-dev-dynamic-hot-upgrade-v0-2-0-index-html',
        },
        {
          matched: [
            { name: 'Root', path: '/' },
            {
              meta: {
                iframeSrc,
                title: 'Dynamic Hot Upgrade Plugin',
              },
              name: 'Link102XAssetsPluginDevDynamicHotUpgradeV020IndexHtml',
              path: '/link-102-x-assets-plugin-dev-dynamic-hot-upgrade-v0-2-0-index-html',
            },
          ],
          meta: {
            iframeSrc,
            title: 'Dynamic Hot Upgrade Plugin',
          },
          path: '/link-102-x-assets-plugin-dev-dynamic-hot-upgrade-v0-2-0-index-html',
        },
        pluginStateMap,
      ).accessible,
    ).toBe(true);
  });

  it('does not keep an installed normal plugin iframe when no accessible route remains', () => {
    const iframeSrc =
      '/x-assets/plugin-dev-dynamic-hot-upgrade/v0.2.0/index.html';
    const pluginStateMap = new Map([
      [
        'plugin-dev-dynamic-hot-upgrade',
        {
          enabled: 1,
          generation: 4,
          id: 'plugin-dev-dynamic-hot-upgrade',
          installed: 1,
          runtimeState: 'normal',
          statusKey: 'sys_plugin.status:plugin-dev-dynamic-hot-upgrade',
          version: 'v0.2.0',
        },
      ],
    ]);

    expect(
      resolveAccessibleRouteRefreshTarget(
        [],
        {
          matched: [
            { name: 'Root', path: '/' },
            { name: 'FallbackNotFound', path: '/:path(.*)*' },
          ],
          path: '/link-102-x-assets-plugin-dev-dynamic-hot-upgrade-v0-2-0-index-html',
        },
        {
          matched: [
            { name: 'Root', path: '/' },
            {
              meta: {
                iframeSrc,
                title: 'Dynamic Hot Upgrade Plugin',
              },
              name: 'Link102XAssetsPluginDevDynamicHotUpgradeV020IndexHtml',
              path: '/link-102-x-assets-plugin-dev-dynamic-hot-upgrade-v0-2-0-index-html',
            },
          ],
          meta: {
            iframeSrc,
            title: 'Dynamic Hot Upgrade Plugin',
          },
          path: '/link-102-x-assets-plugin-dev-dynamic-hot-upgrade-v0-2-0-index-html',
        },
        pluginStateMap,
      ).accessible,
    ).toBe(false);
  });
});

describe('access refresh navigation decision', () => {
  it('stays silent on an accessible host page such as plugin management', () => {
    expect(
      resolveAccessRefreshNavigation({
        accessibleMatch: { accessible: true },
        currentPath: '/system/plugin',
      }),
    ).toEqual({ kind: 'silent' });
  });

  it('stays silent when the accessible route has no path correction', () => {
    expect(
      resolveAccessRefreshNavigation({
        accessibleMatch: {
          accessible: true,
          replacementPath: '/system/plugin',
        },
        currentPath: '/system/plugin',
      }),
    ).toEqual({ kind: 'silent' });
  });

  it('replaces only when a hosted path correction is required', () => {
    expect(
      resolveAccessRefreshNavigation({
        accessibleMatch: {
          accessible: true,
          replacementPath: '/plugins/demo/settings',
        },
        currentPath: '/old-link-demo-settings',
      }),
    ).toEqual({
      kind: 'replace-path',
      path: '/plugins/demo/settings',
    });
  });

  it('falls back when the current route is no longer accessible', () => {
    expect(
      resolveAccessRefreshNavigation({
        accessibleMatch: { accessible: false },
        currentPath: '/plugins/demo/dashboard',
      }),
    ).toEqual({ kind: 'fallback' });
  });

  it('honors explicit skip and pending plugin page refresh as silent', () => {
    expect(
      resolveAccessRefreshNavigation({
        accessibleMatch: { accessible: false },
        currentPath: '/plugins/demo/dashboard',
        skipRouteNavigation: true,
      }),
    ).toEqual({ kind: 'silent' });

    expect(
      resolveAccessRefreshNavigation({
        accessibleMatch: { accessible: true },
        currentPath: '/plugins/demo/dashboard',
        hasPendingPluginPageRefresh: true,
      }),
    ).toEqual({ kind: 'silent' });
  });

  it('prefers force-default over silent skip', () => {
    expect(
      resolveAccessRefreshNavigation({
        accessibleMatch: { accessible: true },
        currentPath: '/system/plugin',
        forceDefaultRoute: true,
        skipRouteNavigation: true,
      }),
    ).toEqual({ kind: 'force-default' });
  });
});
