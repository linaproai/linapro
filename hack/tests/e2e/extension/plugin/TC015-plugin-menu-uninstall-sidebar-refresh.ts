import type { Page, Route } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { workspacePath } from '../../../fixtures/config';
import { PluginPage } from '../../../pages/PluginPage';
import { waitForRouteReady } from '../../../support/ui';

const pluginID = 'plugin-menu-uninstall-sidebar-refresh-e2e';
const pluginMenuName = 'Plugin Menu Refresh E2E';
const pluginManageMenuPattern = /插件管理|Plugin Management/iu;

type PluginRow = ReturnType<typeof pluginRow>;

function apiEnvelope(data: unknown) {
  return {
    code: 0,
    data,
    message: 'success',
  };
}

function publicFrontendSettings() {
  return {
    app: {
      logo: '',
      logoDark: '',
      name: 'LinaPro',
    },
    auth: {
      loginSubtitle: '',
      panelLayout: 'panel-right',
      pageDesc: '',
      pageTitle: '',
    },
    cron: {
      logRetention: {
        mode: 'days',
        value: 30,
      },
      shell: {
        disabledReason: '',
        enabled: false,
        supported: true,
      },
      timezone: {
        current: 'Asia/Shanghai',
      },
    },
    ui: {
      layout: 'sidebar-mixed-nav',
      themeMode: 'light',
      watermarkContent: '',
      watermarkEnabled: false,
    },
    user: {
      defaultAvatar: '',
    },
    workspace: {
      basePath: '/admin',
    },
  };
}

function menuRoutes(includePluginMenu: boolean) {
  const pluginMenus = includePluginMenu
    ? [
        {
          component: 'system/plugin/dynamic-page',
          meta: {
            authority: [pluginID],
            icon: 'lucide:plug',
            order: 50,
            title: pluginMenuName,
          },
          name: 'PluginMenuUninstallSidebarRefreshE2E',
          path: 'plugin-menu-refresh-e2e',
        },
      ]
    : [];

  return [
    {
      children: [
        {
          component: 'dashboard/analytics/index',
          meta: {
            icon: 'lucide:area-chart',
            title: 'page.dashboard.analytics',
          },
          name: 'Analytics',
          path: '/dashboard/analytics',
        },
      ],
      component: 'BasicLayout',
      meta: {
        icon: 'lucide:layout-dashboard',
        order: -1,
        title: 'page.dashboard.title',
      },
      name: 'Dashboard',
      path: '/dashboard',
    },
    {
      children: [
        {
          component: 'system/plugin/index',
          meta: {
            icon: 'lucide:plug',
            title: 'page.routes.system.pluginManagement',
          },
          name: 'PluginManagement',
          path: '/system/plugin',
        },
        ...pluginMenus,
      ],
      component: 'BasicLayout',
      meta: {
        icon: 'lucide:puzzle',
        order: 40,
        title: 'page.routes.system.extensionCenter',
      },
      name: 'Extension',
      path: '/extension',
    },
  ];
}

function pluginRow(installed: 0 | 1, enabled: 0 | 1 = installed): {
  abnormalReason: string;
  authorizationRequired: 0;
  authorizationStatus: 'not_required';
  autoEnableForNewTenants: boolean;
  autoEnableManaged: 0;
  authorizedHostServices: never[];
  declaredRoutes: never[];
  dependencyCheck: ReturnType<typeof emptyDependencyCheck>;
  description: string;
  discoveredVersion: string;
  effectiveVersion: string;
  enabled: 0 | 1;
  hasMockData: 0;
  id: string;
  installMode: string;
  installed: 0 | 1;
  installedAt: string;
  lastUpgradeFailure: undefined;
  name: string;
  requestedHostServices: never[];
  runtimeState: string;
  scopeNature: string;
  statusKey: string;
  supportsMultiTenant: boolean;
  type: string;
  updatedAt: string;
  upgradeAvailable: boolean;
  version: string;
} {
  return {
    abnormalReason: '',
    authorizationRequired: 0,
    authorizationStatus: 'not_required',
    autoEnableForNewTenants: false,
    autoEnableManaged: 0,
    authorizedHostServices: [],
    declaredRoutes: [],
    dependencyCheck: emptyDependencyCheck(),
    description:
      'Installed plugin used to verify sidebar menu refresh after lifecycle changes.',
    discoveredVersion: 'v0.1.0',
    effectiveVersion: 'v0.1.0',
    enabled: installed === 1 ? enabled : 0,
    hasMockData: 0,
    id: pluginID,
    installMode: 'global',
    installed,
    installedAt: installed === 1 ? '2026-06-08T00:00:00Z' : '',
    lastUpgradeFailure: undefined,
    name: 'Plugin Menu Uninstall Sidebar Refresh E2E',
    requestedHostServices: [],
    runtimeState: 'normal',
    scopeNature: 'global',
    statusKey:
      installed !== 1 ? 'not_installed' : enabled === 1 ? 'enabled' : 'disabled',
    supportsMultiTenant: false,
    type: 'source',
    updatedAt: '',
    upgradeAvailable: false,
    version: 'v0.1.0',
  };
}

function dynamicState(row: PluginRow) {
  return {
    enabled: row.enabled,
    generation: 1,
    id: row.id,
    installed: row.installed,
    runtimeState: row.runtimeState,
    statusKey: `sys_plugin.status:${row.statusKey}`,
    version: row.version,
  };
}

function emptyDependencyCheck() {
  return {
    blockers: [],
    cycle: [],
    dependencies: [],
    framework: {
      currentVersion: 'v0.1.0',
      requiredVersion: '',
      status: 'not_declared',
    },
    reverseBlockers: [],
    reverseDependents: [],
    targetId: pluginID,
  };
}

async function mockPluginMenuRefreshApis(page: Page) {
  let installed: 0 | 1 = 1;
  let enabled: 0 | 1 = 1;

  await page.route('**/api/v1/config/public/frontend', async (route) => {
    await route.fulfill({ json: apiEnvelope(publicFrontendSettings()) });
  });

  await page.route('**/api/v1/menus/all', async (route) => {
    await route.fulfill({
      json: apiEnvelope({
        list: menuRoutes(installed === 1 && enabled === 1),
      }),
    });
  });

  await page.route('**/api/v1/plugins**', async (route: Route) => {
    const request = route.request();
    const url = new URL(request.url());
    const path = url.pathname;

    if (request.method() === 'GET' && /\/api\/v1\/plugins$/u.test(path)) {
      const id = url.searchParams.get('id')?.trim();
      const row = pluginRow(installed, enabled);
      const rows = id && !pluginID.includes(id) ? [] : [row];
      await route.fulfill({
        json: apiEnvelope({
          list: rows,
          total: rows.length,
        }),
      });
      return;
    }

    if (request.method() === 'GET' && path.endsWith('/plugins/dynamic')) {
      await route.fulfill({
        json: apiEnvelope({
          list: [dynamicState(pluginRow(installed, enabled))],
        }),
      });
      return;
    }

    if (request.method() === 'GET' && path.endsWith(`/plugins/${pluginID}`)) {
      await route.fulfill({ json: apiEnvelope(pluginRow(installed, enabled)) });
      return;
    }

    if (
      request.method() === 'GET' &&
      path.endsWith(`/plugins/${pluginID}/dependencies`)
    ) {
      await route.fulfill({ json: apiEnvelope(emptyDependencyCheck()) });
      return;
    }

    if (
      request.method() === 'DELETE' &&
      path.endsWith(`/plugins/${pluginID}`)
    ) {
      installed = 0;
      enabled = 0;
      await route.fulfill({ json: apiEnvelope(null) });
      return;
    }

    if (
      request.method() === 'PUT' &&
      path.endsWith(`/plugins/${pluginID}/disable`)
    ) {
      enabled = 0;
      await route.fulfill({ json: apiEnvelope(null) });
      return;
    }

    if (
      request.method() === 'PUT' &&
      path.endsWith(`/plugins/${pluginID}/enable`)
    ) {
      installed = 1;
      enabled = 1;
      await route.fulfill({ json: apiEnvelope(null) });
      return;
    }

    await route.continue();
  });
}

async function expectSidebarMenuWithoutReload(pluginPage: PluginPage) {
  await expect(pluginPage.sidebarMenu).toBeVisible({ timeout: 10_000 });
  await expect(
    pluginPage.sidebarMenu.getByText(pluginManageMenuPattern).first(),
  ).toBeVisible();
}

function tabbarItem(page: Page, name: string) {
  return page.locator('[data-tab-item="true"]').filter({ hasText: name });
}

test.describe('TC-15 插件生命周期后的侧边菜单刷新', () => {
  test('TC-15a: 卸载已访问过菜单的插件后左侧菜单和历史标签无需强刷仍同步刷新', async ({
    adminPage,
  }) => {
    const pageErrors: string[] = [];
    adminPage.on('pageerror', (error) => pageErrors.push(error.message));
    await mockPluginMenuRefreshApis(adminPage);

    const pluginPage = new PluginPage(adminPage);
    await adminPage.goto(workspacePath('/system/plugin'), {
      waitUntil: 'domcontentloaded',
    });
    await waitForRouteReady(adminPage, 15_000);
    await pluginPage.searchByPluginId(pluginID);
    await expectSidebarMenuWithoutReload(pluginPage);

    const pluginMenuItem = pluginPage.sidebarMenu
      .getByRole('menuitem', { name: pluginMenuName })
      .first();
    await expect(pluginMenuItem).toBeVisible();
    await pluginMenuItem.click();
    await waitForRouteReady(adminPage, 15_000);
    await expect(tabbarItem(adminPage, pluginMenuName)).toBeVisible();

    await pluginPage.sidebarMenu
      .getByText(pluginManageMenuPattern)
      .first()
      .click();
    await waitForRouteReady(adminPage, 15_000);
    await pluginPage.searchByPluginId(pluginID);

    await pluginPage.uninstallPlugin(pluginID);

    await expectSidebarMenuWithoutReload(pluginPage);
    await expect(
      pluginPage.sidebarMenu.getByRole('menuitem', { name: pluginMenuName }),
    ).toHaveCount(0);
    await expect(tabbarItem(adminPage, pluginMenuName)).toHaveCount(0);
    expect(pageErrors).toEqual([]);
  });

  test('TC-15b: 禁用已访问过菜单的插件后左侧菜单和历史标签无需强刷仍同步刷新', async ({
    adminPage,
  }) => {
    const pageErrors: string[] = [];
    adminPage.on('pageerror', (error) => pageErrors.push(error.message));
    await mockPluginMenuRefreshApis(adminPage);

    const pluginPage = new PluginPage(adminPage);
    await adminPage.goto(workspacePath('/system/plugin'), {
      waitUntil: 'domcontentloaded',
    });
    await waitForRouteReady(adminPage, 15_000);
    await pluginPage.searchByPluginId(pluginID);
    await expectSidebarMenuWithoutReload(pluginPage);

    const pluginMenuItem = pluginPage.sidebarMenu
      .getByRole('menuitem', { name: pluginMenuName })
      .first();
    await expect(pluginMenuItem).toBeVisible();
    await pluginMenuItem.click();
    await waitForRouteReady(adminPage, 15_000);
    await expect(tabbarItem(adminPage, pluginMenuName)).toBeVisible();

    await pluginPage.sidebarMenu
      .getByText(pluginManageMenuPattern)
      .first()
      .click();
    await waitForRouteReady(adminPage, 15_000);
    await pluginPage.searchByPluginId(pluginID);

    await pluginPage.setPluginEnabled(pluginID, false);

    await expectSidebarMenuWithoutReload(pluginPage);
    await expect(
      pluginPage.sidebarMenu.getByRole('menuitem', { name: pluginMenuName }),
    ).toHaveCount(0);
    await expect(tabbarItem(adminPage, pluginMenuName)).toHaveCount(0);
    expect(pageErrors).toEqual([]);
  });
});
