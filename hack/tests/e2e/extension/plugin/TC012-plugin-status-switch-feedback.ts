import type { Page, Route } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { PluginPage } from '../../../pages/PluginPage';

const successPluginID = 'plugin-status-switch-feedback-success-e2e';
const failurePluginID = 'plugin-status-switch-feedback-failure-e2e';

type PluginRow = ReturnType<typeof pluginRow>;

function apiEnvelope(data: unknown) {
  return {
    code: 0,
    data,
    message: 'success',
  };
}

function pluginRow(input: { enabled: 0 | 1; id: string; name: string }) {
  return {
    abnormalReason: '',
    authorizationRequired: 0,
    authorizationStatus: 'not_required',
    autoEnableForNewTenants: false,
    autoEnableManaged: 0,
    authorizedHostServices: [],
    declaredRoutes: [],
    dependencyCheck: null,
    description: 'Used by E2E to verify immediate status switch feedback.',
    discoveredVersion: 'v0.1.0',
    effectiveVersion: 'v0.1.0',
    enabled: input.enabled,
    hasMockData: 0,
    id: input.id,
    installMode: 'global',
    installed: 1,
    installedAt: '',
    lastUpgradeFailure: undefined,
    name: input.name,
    requestedHostServices: [],
    runtimeState: 'normal',
    scopeNature: 'global',
    statusKey: input.enabled === 1 ? 'enabled' : 'disabled',
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

function updateRowEnabled(row: PluginRow, enabled: 0 | 1) {
  row.enabled = enabled;
  row.statusKey = enabled === 1 ? 'enabled' : 'disabled';
}

async function mockPluginStatusApis(page: Page) {
  const rows = [
    pluginRow({
      enabled: 0,
      id: successPluginID,
      name: 'Plugin Status Switch Feedback Success E2E',
    }),
    pluginRow({
      enabled: 0,
      id: failurePluginID,
      name: 'Plugin Status Switch Feedback Failure E2E',
    }),
  ];

  let resolveEnableSuccess: null | (() => void) = null;
  const enableSuccessReceived = new Promise<void>((resolve) => {
    resolveEnableSuccess = resolve;
  });

  await page.route('**/api/v1/plugins**', async (route: Route) => {
    const request = route.request();
    const url = new URL(request.url());
    const path = url.pathname;

    if (request.method() === 'GET' && /\/api\/v1\/plugins$/u.test(path)) {
      const id = url.searchParams.get('id')?.trim();
      const filteredRows = id
        ? rows.filter((row) => String(row.id ?? '').includes(id))
        : rows;
      await route.fulfill({
        json: apiEnvelope({
          list: filteredRows,
          total: filteredRows.length,
        }),
      });
      return;
    }

    if (request.method() === 'GET' && path.endsWith('/plugins/dynamic')) {
      await route.fulfill({
        json: apiEnvelope({
          list: rows.map((row) => dynamicState(row)),
        }),
      });
      return;
    }

    if (
      request.method() === 'PUT' &&
      path.endsWith(`/plugins/${successPluginID}/enable`)
    ) {
      resolveEnableSuccess?.();
      await new Promise((resolve) => setTimeout(resolve, 650));
      updateRowEnabled(rows[0]!, 1);
      await route.fulfill({ json: apiEnvelope(null) });
      return;
    }

    if (
      request.method() === 'PUT' &&
      path.endsWith(`/plugins/${failurePluginID}/enable`)
    ) {
      await route.fulfill({
        json: {
          code: 500,
          data: null,
          errorCode: 'PLUGIN_ENABLE_FAILED',
          message: 'Enable failed for E2E',
          messageKey: 'error.plugin.enable.failed',
          messageParams: {},
        },
        status: 500,
      });
      return;
    }

    await route.continue();
  });

  return {
    waitForEnableSuccessRequest: () => enableSuccessReceived,
  };
}

test.describe('TC-12 插件状态开关即时反馈', () => {
  test('TC-12a: 启用接口较慢时状态开关立即切到目标状态并显示加载态', async ({
    adminPage,
  }) => {
    const mock = await mockPluginStatusApis(adminPage);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(successPluginID);

    const switcher = pluginPage.pluginEnabledSwitch(successPluginID);
    await expect(switcher).toHaveAttribute('aria-checked', 'false');
    await switcher.click();
    await mock.waitForEnableSuccessRequest();

    await expect(switcher).toHaveAttribute('aria-checked', 'true');
    await expect(switcher).toHaveClass(/ant-switch-loading/);
    await expect(switcher).toHaveClass(/ant-switch-disabled/);
    await expect(pluginPage.messageNotices('插件已启用')).toHaveCount(0);

    await expect(switcher).not.toHaveClass(/ant-switch-loading/);
    await expect(pluginPage.messageNotice('插件已启用')).toBeVisible();
  });

  test('TC-12b: 启用失败时状态开关回滚到原状态', async ({
    adminPage,
  }) => {
    await mockPluginStatusApis(adminPage);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(failurePluginID);

    const switcher = pluginPage.pluginEnabledSwitch(failurePluginID);
    await expect(switcher).toHaveAttribute('aria-checked', 'false');
    await switcher.click();

    await expect(switcher).toHaveAttribute('aria-checked', 'false');
    await expect(switcher).not.toHaveClass(/ant-switch-loading/);
    await expect(pluginPage.messageNotice('Enable failed for E2E')).toBeVisible();
  });
});
