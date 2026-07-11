import type { Page, Route } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { PluginPage } from '../../../pages/PluginPage';

const layoutPluginID = 'plugin-management-table-layout-e2e';
const installLayoutPluginID = 'plugin-management-install-layout-e2e';

function apiEnvelope(data: unknown) {
  return {
    code: 0,
    data,
    message: 'success',
  };
}

function pluginRow(overrides: Record<string, unknown> = {}) {
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
    discoveredVersion: 'v0.2.0',
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
    runtimeState: 'pending_upgrade',
    scopeNature: 'global',
    statusKey: 'enabled',
    supportsMultiTenant: false,
    type: 'source',
    updatedAt: '',
    upgradeAvailable: true,
    version: 'v0.1.0',
    ...overrides,
  };
}

function installLayoutPluginRow() {
  return pluginRow({
    description:
      'Used by E2E to verify plugin install modal Descriptions label layout.',
    discoveredVersion: 'v0.1.0',
    effectiveVersion: '',
    enabled: 0,
    id: installLayoutPluginID,
    installed: 0,
    name: 'Plugin Install Layout E2E',
    runtimeState: 'not_installed',
    statusKey: 'not_installed',
    upgradeAvailable: false,
  });
}

async function mockPluginListApis(
  page: Page,
  row: ReturnType<typeof pluginRow> = pluginRow(),
) {
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

    // Install modal refreshes dependency projection after open.
    if (
      request.method() === 'GET' &&
      path.endsWith(`/plugins/${row.id}/dependencies`)
    ) {
      await route.fulfill({
        json: apiEnvelope({
          blockers: [],
          cycle: [],
          framework: { status: 'satisfied' },
          missing: [],
          optionalMissing: [],
          status: 'ok',
        }),
      });
      return;
    }

    // Detail / install modal loads full projection before open; without this
    // mock the click fails closed against a non-existent backend plugin id.
    if (
      request.method() === 'GET' &&
      path.endsWith(`/plugins/${row.id}`)
    ) {
      await route.fulfill({
        json: apiEnvelope(row),
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
    await pluginPage.expectTableColumnWiderThan('版本号', [
      '插件类型',
      '状态',
      '运行时状态',
      '示例数据',
      '支持多租户',
      '新租户启用',
    ]);
    await pluginPage.expectTableColumnWidthAtMost('插件类型', 112);
    await pluginPage.expectTableColumnWidthAtMost('状态', 100);
    await pluginPage.expectTableColumnWidthAtMost('运行时状态', 116);
    await pluginPage.expectTableColumnWidthAtMost('示例数据', 108);
    await pluginPage.expectTableColumnWidthAtMost('支持多租户', 126);
    await pluginPage.expectTableColumnWidthAtMost('新租户启用', 130);
    // Action column: detail + manage + one lifecycle button, single non-wrapping row.
    // Fixed-right header cells are not visible in the main header table.
    await pluginPage.expectPluginActionColumnWidthAtMost(layoutPluginID, 210);
    await pluginPage.expectPluginActionButtonsSingleLine(layoutPluginID);
    await expect(pluginPage.pluginVersionValue(layoutPluginID)).toContainText(
      /v0\.1\.0\s*->\s*v0\.2\.0/u,
    );
    await pluginPage.expectPluginVersionNotClipped(layoutPluginID);
    await expect(pluginPage.pluginColumnHelpIcon('runtimeState')).toBeVisible();
    await pluginPage.expectColumnHelpTooltip(
      'runtimeState',
      /运行时状态表示插件文件发现版本与数据库有效版本.*状态列表示插件当前是否启用/u,
    );
  });

  test('TC-13b: 插件详情页最左标签列保持单行不换行', async ({ adminPage }) => {
    await mockPluginListApis(adminPage);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(layoutPluginID);
    await pluginPage.openPluginDetail(layoutPluginID);

    await expect(pluginPage.pluginDetailModal()).toBeVisible();
    await expect(pluginPage.pluginDetailDescriptions()).toBeVisible();
    // Multi-character / multi-word field names are the ones that used to wrap.
    await expect(pluginPage.pluginDetailModal()).toContainText('授权状态');
    await expect(pluginPage.pluginDetailModal()).toContainText('有效版本');
    await expect(pluginPage.pluginDetailModal()).toContainText('发现版本');
    await pluginPage.expectPluginDetailLabelsNoWrap();
  });

  test('TC-13c: 插件安装页最左标题列保持单行不换行', async ({ adminPage }) => {
    await mockPluginListApis(adminPage, installLayoutPluginRow());

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(installLayoutPluginID);
    await pluginPage.openInstallAuthorization(installLayoutPluginID);

    await expect(pluginPage.hostServiceAuthModal()).toBeVisible();
    await expect(pluginPage.pluginInstallDescriptions()).toBeVisible();
    await expect(pluginPage.hostServiceAuthModal()).toContainText('插件名称');
    await expect(pluginPage.hostServiceAuthModal()).toContainText('插件标识');
    await expect(pluginPage.hostServiceAuthModal()).toContainText('插件类型');
    await pluginPage.expectPluginInstallLabelsNoWrap();
  });
});
