import { mkdir } from 'node:fs/promises';
import path from 'node:path';

import type { Page, Route } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { PluginPage } from '../../../pages/PluginPage';

// Uses real source-plugin ids that ship navigable frontend pages so the
// build-time page-registry can enable the Manage button when installed.
// Host-only (`plugins=0`) excludes this file via hostOnlyExcludedEntries —
// page-registry is empty without official plugin discovery. Covered by the
// plugin-full extension:plugin suite instead.
const managedPluginID = 'linapro-monitor-loginlog';
const managedPluginName = 'Login Log Management';
const notInstalledManagedPluginID = 'linapro-content-notice';
const notInstalledManagedPluginName = 'Notice Management Not Installed';
const unmanagedPluginID = 'plugin-no-management-page-e2e';
const unmanagedPluginName = 'No Management Page E2E';

function apiEnvelope(data: unknown) {
  return {
    code: 0,
    data,
    message: 'success',
  };
}

function pluginRow(partial: {
  description: string;
  enabled?: 0 | 1;
  id: string;
  installed?: 0 | 1;
  name: string;
  statusKey?: string;
}) {
  const installed = partial.installed ?? 1;
  const enabled = partial.enabled ?? (installed === 1 ? 1 : 0);
  return {
    abnormalReason: '',
    authorizationRequired: 0,
    authorizationStatus: 'not_required',
    autoEnableForNewTenants: false,
    autoEnableManaged: 0,
    authorizedHostServices: [],
    declaredRoutes: [],
    dependencyCheck: null,
    description: partial.description,
    discoveredVersion: 'v0.1.0',
    effectiveVersion: 'v0.1.0',
    enabled,
    hasMockData: 0,
    id: partial.id,
    installMode: 'global',
    installed,
    installedAt: installed === 1 ? 1767240000000 : null,
    lastUpgradeFailure: undefined,
    name: partial.name,
    requestedHostServices: [],
    runtimeState: 'normal',
    scopeNature: 'global',
    statusKey:
      partial.statusKey ??
      (installed !== 1 ? 'not_installed' : enabled === 1 ? 'enabled' : 'disabled'),
    supportsMultiTenant: false,
    type: 'source',
    updatedAt: 1767240000000,
    upgradeAvailable: false,
    version: 'v0.1.0',
  };
}

function rows() {
  return [
    pluginRow({
      description: 'Installed and has a registered management page.',
      id: managedPluginID,
      name: managedPluginName,
    }),
    pluginRow({
      description: 'Has a management page but is not installed yet.',
      enabled: 0,
      id: notInstalledManagedPluginID,
      installed: 0,
      name: notInstalledManagedPluginName,
    }),
    pluginRow({
      description: 'Has no frontend management page.',
      id: unmanagedPluginID,
      name: unmanagedPluginName,
    }),
  ];
}

async function mockPluginManageEntryApis(page: Page) {
  await page.route('**/api/v1/plugins**', async (route: Route) => {
    const request = route.request();
    const url = new URL(request.url());
    const requestPath = url.pathname;

    if (request.method() === 'GET' && /\/api\/v1\/plugins$/u.test(requestPath)) {
      const idFilter = url.searchParams.get('id')?.trim() ?? '';
      const list = rows().filter((item) => !idFilter || item.id.includes(idFilter));
      await route.fulfill({
        json: apiEnvelope({
          list,
          total: list.length,
        }),
      });
      return;
    }

    if (request.method() === 'GET' && requestPath.endsWith('/plugins/dynamic')) {
      await route.fulfill({
        json: apiEnvelope({
          list: rows().map((item) => ({
            enabled: item.enabled,
            generation: 1,
            id: item.id,
            installed: item.installed,
            runtimeState: item.runtimeState,
            statusKey: `sys_plugin.status:${item.statusKey}`,
            version: item.version,
          })),
        }),
      });
      return;
    }

    if (request.method() === 'GET' && /\/api\/v1\/plugins\/[^/]+$/u.test(requestPath)) {
      const id = requestPath.split('/').pop() ?? '';
      const row = rows().find((item) => item.id === id);
      if (row) {
        await route.fulfill({ json: apiEnvelope(row) });
        return;
      }
    }

    await route.fulfill({ json: apiEnvelope({}) });
  });
}

async function captureEvidence(page: Page, name: string) {
  await page.waitForTimeout(300);
  const now = new Date();
  const day = new Intl.DateTimeFormat('en-CA', {
    day: '2-digit',
    month: '2-digit',
    timeZone: 'Asia/Shanghai',
    year: 'numeric',
  })
    .format(now)
    .replaceAll('-', '');
  const time = new Intl.DateTimeFormat('en-GB', {
    hour: '2-digit',
    hour12: false,
    minute: '2-digit',
    second: '2-digit',
    timeZone: 'Asia/Shanghai',
  })
    .format(now)
    .replaceAll(':', '');
  const dir = path.resolve(process.cwd(), '..', '..', 'temp', day);
  await mkdir(dir, { recursive: true });
  await page.screenshot({
    fullPage: true,
    path: path.join(dir, `${time}-${name}.png`),
  });
}

test.describe('TC-19 插件管理列表管理入口', () => {
  test('TC-19a: 已安装有管理页可点，无管理页或未安装时置灰', async ({
    adminPage,
  }) => {
    await mockPluginManageEntryApis(adminPage);
    const pluginPage = new PluginPage(adminPage);

    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId('plugin-no-management-page-e2e');
    await expect(pluginPage.pluginRow(unmanagedPluginID)).toBeVisible();

    const unmanagedManage = pluginPage.pluginManageAction(unmanagedPluginID);
    await expect(unmanagedManage).toBeVisible();
    await expect(unmanagedManage).toBeDisabled();
    await expect(unmanagedManage).toHaveText(/管\s*理|Manage/iu);

    await pluginPage.searchByPluginId(notInstalledManagedPluginID);
    await expect(pluginPage.pluginRow(notInstalledManagedPluginID)).toBeVisible();

    const notInstalledManage = pluginPage.pluginManageAction(
      notInstalledManagedPluginID,
    );
    await expect(notInstalledManage).toBeVisible();
    await expect(notInstalledManage).toBeDisabled();
    await expect(notInstalledManage).toHaveText(/管\s*理|Manage/iu);

    await pluginPage.searchByPluginId(managedPluginID);
    await expect(pluginPage.pluginRow(managedPluginID)).toBeVisible();

    const managedManage = pluginPage.pluginManageAction(managedPluginID);
    await expect(managedManage).toBeVisible();
    await expect(managedManage).toBeEnabled();
    await expect(managedManage).toHaveText(/管\s*理|Manage/iu);

    await captureEvidence(adminPage, 'plugin-management-manage-entry-states');
  });

  test('TC-19b: 点击管理进入登录日志页或提示当前路由不可用，且不打开详情', async ({
    adminPage,
  }) => {
    await mockPluginManageEntryApis(adminPage);
    const pluginPage = new PluginPage(adminPage);

    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(managedPluginID);
    await expect(pluginPage.pluginManageAction(managedPluginID)).toBeEnabled();

    // Race navigation and the short-lived warning toast. Waiting for URL first
    // can miss ant-design message toasts that auto-dismiss in ~3s.
    // Swallow the losing waiter so its later timeout does not surface as an
    // unhandled rejection after Promise.race settles.
    const unavailableMessage = adminPage.getByText(
      /当前会话无法访问该插件管理页面|not available in the current session/iu,
    );
    const outcomePromise = Promise.race([
      adminPage
        .waitForURL(/loginlog/iu, { timeout: 8_000 })
        .then(() => 'navigated' as const)
        .catch(() => null),
      unavailableMessage
        .waitFor({ state: 'visible', timeout: 8_000 })
        .then(() => 'unavailable' as const)
        .catch(() => null),
    ]);

    await pluginPage.openPluginManagement(managedPluginID);
    await expect(pluginPage.pluginDetailModal()).toHaveCount(0);

    const outcome = await outcomePromise;
    expect(
      outcome,
      'expected Manage to navigate to loginlog or show the unavailable toast',
    ).not.toBeNull();
    if (outcome === 'navigated') {
      expect(adminPage.url()).toMatch(/loginlog/iu);
      await expect(adminPage).not.toHaveURL(/\/system\/plugin(?:\?|$)/u);
    } else {
      await expect(unavailableMessage).toBeVisible();
      await expect(adminPage).toHaveURL(/\/system\/plugin/u);
    }

    await captureEvidence(adminPage, 'plugin-management-manage-entry-click');
  });
});
