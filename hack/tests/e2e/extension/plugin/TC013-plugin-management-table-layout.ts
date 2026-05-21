import type { Page, Route } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { PluginPage } from '../../../pages/PluginPage';

const layoutPluginID = 'plugin-management-table-layout-e2e';

function apiEnvelope(data: unknown) {
  return {
    code: 0,
    data,
    message: 'success',
  };
}

function pluginRow() {
  return {
    abnormalReason: '',
    authorizationRequired: 0,
    authorizationStatus: 'not_required',
    autoEnableForNewTenants: false,
    autoEnableManaged: 0,
    authorizedHostServices: [],
    declaredRoutes: [],
    dependencyCheck: null,
    description: 'Used by E2E to verify plugin management table layout.',
    discoveredVersion: 'v0.1.0',
    effectiveVersion: 'v0.1.0',
    enabled: 1,
    hasMockData: 0,
    id: layoutPluginID,
    installMode: 'global',
    installed: 1,
    installedAt: '',
    lastUpgradeFailure: undefined,
    name: 'Plugin Management Table Layout E2E',
    requestedHostServices: [],
    runtimeState: 'normal',
    scopeNature: 'global',
    statusKey: 'enabled',
    supportsMultiTenant: false,
    type: 'source',
    updatedAt: '',
    upgradeAvailable: false,
    version: 'v0.1.0',
  };
}

async function mockPluginListApis(page: Page) {
  const row = pluginRow();

  await page.route('**/api/v1/plugins**', async (route: Route) => {
    const request = route.request();
    const url = new URL(request.url());
    const path = url.pathname;

    if (request.method() === 'GET' && /\/api\/v1\/plugins$/u.test(path)) {
      await route.fulfill({
        json: apiEnvelope({
          list: [row],
          total: 1,
        }),
      });
      return;
    }

    if (request.method() === 'GET' && path.endsWith('/plugins/dynamic')) {
      await route.fulfill({
        json: apiEnvelope({
          list: [
            {
              enabled: row.enabled,
              generation: 1,
              id: row.id,
              installed: row.installed,
              runtimeState: row.runtimeState,
              statusKey: `sys_plugin.status:${row.statusKey}`,
              version: row.version,
            },
          ],
        }),
      });
      return;
    }

    await route.continue();
  });
}

test.describe('TC-13 插件管理列表布局', () => {
  test('TC-13a: 插件管理列表按基础信息顺序展示并补充运行时状态说明', async ({
    adminPage,
  }) => {
    await mockPluginListApis(adminPage);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(layoutPluginID);

    await pluginPage.expectTableColumnOrder([
      '插件标识',
      '插件名称',
      '插件描述',
      '版本号',
      '插件类型',
    ]);
    await pluginPage.expectTableColumnCentered('插件标识');
    await pluginPage.expectTableColumnCentered('插件名称');
    await pluginPage.expectTableColumnCentered('插件描述');
    await pluginPage.expectTableColumnCentered('版本号');
    await pluginPage.expectTableColumnCentered('插件类型');
    await pluginPage.expectTableColumnLeftAligned('插件标识');
    await pluginPage.expectTableColumnLeftAligned('插件名称');
    await pluginPage.expectTableColumnLeftAligned('插件描述');
    await pluginPage.expectTableColumnAfter('运行时状态', '状态');
    await pluginPage.expectTableColumnWiderThan('插件标识', [
      '插件名称',
      '版本号',
      '运行时状态',
    ]);
    await pluginPage.expectTableColumnWiderThan('插件描述', [
      '插件名称',
      '版本号',
    ]);
    await expect(pluginPage.pluginColumnHelpIcon('runtimeState')).toBeVisible();
    await pluginPage.expectColumnHelpTooltip(
      'runtimeState',
      /运行时状态表示插件文件发现版本与数据库有效版本.*状态列表示插件当前是否启用/u,
    );
  });
});
