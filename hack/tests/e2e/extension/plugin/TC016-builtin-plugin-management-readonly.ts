import type { Page, Route } from '@playwright/test';

import { mkdir } from 'node:fs/promises';
import path from 'node:path';

import { test, expect } from '../../../fixtures/auth';
import { PluginPage } from '../../../pages/PluginPage';

const builtinPluginID = 'linapro-builtin-readonly-e2e';
const managedPluginID = 'linapro-managed-visible-e2e';

function apiEnvelope(data: unknown) {
  return {
    code: 0,
    data,
    message: 'success',
  };
}

function pluginRow(input: {
  distribution: 'builtin' | 'managed';
  id: string;
  installed: 0 | 1;
  name: string;
  upgradeAvailable?: boolean;
}) {
  return {
    abnormalReason: '',
    authorizationRequired: 0,
    authorizationStatus: 'not_required',
    autoEnableForNewTenants: true,
    autoEnableManaged: 0,
    authorizedHostServices: [],
    declaredRoutes: [],
    dependencyCheck: null,
    description: `Used by E2E to verify ${input.distribution} plugin governance.`,
    discoveredVersion: input.upgradeAvailable ? 'v0.2.0' : 'v0.1.0',
    distribution: input.distribution,
    effectiveVersion: 'v0.1.0',
    enabled: input.installed,
    hasMockData: 0,
    id: input.id,
    installMode: 'tenant_scoped',
    installed: input.installed,
    installedAt: 1767240000000,
    lastUpgradeFailure: undefined,
    name: input.name,
    requestedHostServices: [],
    runtimeState: input.upgradeAvailable ? 'pending_upgrade' : 'normal',
    scopeNature: 'tenant_aware',
    statusKey: input.installed === 1 ? 'enabled' : 'disabled',
    supportsMultiTenant: true,
    type: 'source',
    updatedAt: 1767240000000,
    upgradeAvailable: input.upgradeAvailable === true,
    version: 'v0.1.0',
  };
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

async function expectNoRawPluginI18nKeys(page: Page) {
  await expect(page.locator('body')).not.toContainText(/pages\.system\.plugin/u);
}

async function mockPluginApis(page: Page) {
  const managedRow = pluginRow({
    distribution: 'managed',
    id: managedPluginID,
    installed: 0,
    name: 'Managed Visible E2E',
  });
  const builtinRow = pluginRow({
    distribution: 'builtin',
    id: builtinPluginID,
    installed: 1,
    name: 'Builtin Readonly E2E',
    upgradeAvailable: true,
  });
  const listRows = [managedRow, builtinRow];
  const listRequests: string[] = [];

  await page.route('**/api/v1/plugins**', async (route: Route) => {
    const request = route.request();
    const url = new URL(request.url());
    const requestPath = url.pathname;

    if (request.method() === 'GET' && /\/api\/v1\/plugins$/u.test(requestPath)) {
      listRequests.push(url.search);
      const id = url.searchParams.get('id')?.trim();
      const rows = id
        ? listRows.filter((row) => row.id.includes(id))
        : listRows;
      await route.fulfill({
        json: apiEnvelope({
          list: rows,
          total: rows.length,
        }),
      });
      return;
    }

    if (request.method() === 'GET' && requestPath.endsWith('/plugins/dynamic')) {
      await route.fulfill({
        json: apiEnvelope({
          list: [],
        }),
      });
      return;
    }

    if (
      request.method() === 'GET' &&
      requestPath.endsWith(`/plugins/${builtinPluginID}`)
    ) {
      await route.fulfill({ json: apiEnvelope(builtinRow) });
      return;
    }

    if (
      request.method() === 'GET' &&
      requestPath.endsWith(`/plugins/${managedPluginID}`)
    ) {
      await route.fulfill({ json: apiEnvelope(managedRow) });
      return;
    }

    await route.continue();
  });

  return {
    listRequests: () => [...listRequests],
  };
}

test.describe('TC-16 内建插件管理只读治理', () => {
  test('TC-16a: 普通插件管理列表默认展示 builtin 与内置标识', async ({
    adminPage,
  }) => {
    const pageErrors: string[] = [];
    adminPage.on('pageerror', (error) => pageErrors.push(error.message));
    const api = await mockPluginApis(adminPage);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();

    await expect(pluginPage.pluginRow(managedPluginID)).toBeVisible();
    await expect(pluginPage.pluginRow(builtinPluginID)).toBeVisible();
    await expect(pluginPage.pluginBuiltinTag(builtinPluginID)).toBeVisible();
    // Default E2E locale is zh-CN; assert translated copy, not raw i18n keys.
    await expect(pluginPage.pluginBuiltinTag(builtinPluginID)).toHaveText(
      '内置插件',
    );
    await expect(pluginPage.pluginBuiltinTag(managedPluginID)).toHaveCount(0);
    expect(api.listRequests().length).toBeGreaterThan(0);
    await expectNoRawPluginI18nKeys(adminPage);
    await captureEvidence(adminPage, 'builtin-plugin-management-default-list');
    expect(pageErrors).toEqual([]);
  });

  test('TC-16b: builtin 行限制写操作，卸载置灰，并保留详情与管理入口', async ({
    adminPage,
  }) => {
    const pageErrors: string[] = [];
    adminPage.on('pageerror', (error) => pageErrors.push(error.message));
    await mockPluginApis(adminPage);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(builtinPluginID);

    await expect(pluginPage.pluginRow(builtinPluginID)).toBeVisible();
    await expect(pluginPage.pluginBuiltinTag(builtinPluginID)).toBeVisible();
    await expect(pluginPage.pluginEnabledSwitches(builtinPluginID)).toHaveCount(0);
    await expect(
      pluginPage.pluginTenantProvisioningSwitches(builtinPluginID),
    ).toHaveCount(0);
    await pluginPage.expectInstallActionHidden(builtinPluginID);
    await pluginPage.expectUpgradeActionHidden(builtinPluginID);
    // Builtin rows still show uninstall for action-column layout parity, but
    // the control is disabled and must not open the uninstall dialog.
    await pluginPage.expectUninstallActionDisabled(builtinPluginID);
    await pluginPage.pluginUninstallActionWrapper(builtinPluginID).hover();
    await expect(
      pluginPage
        .antTooltip()
        .filter({
          hasText:
            /项目内建能力，不可卸载|project built-in capability and cannot be uninstalled/iu,
        })
        .last(),
    ).toBeVisible();
    await pluginPage.pluginUninstallAction(builtinPluginID).click({
      force: true,
    });
    await expect(pluginPage.uninstallDialog()).toHaveCount(0);
    await expect(pluginPage.pluginDetailAction(builtinPluginID)).toBeVisible();
    await expect(pluginPage.pluginManageAction(builtinPluginID)).toBeVisible();
    await captureEvidence(adminPage, 'builtin-plugin-management-readonly-row');

    await pluginPage.openPluginDetail(builtinPluginID);
    await expect(pluginPage.pluginDetailModal()).toContainText('Builtin Readonly E2E');
    await expect(pluginPage.pluginDetailDistribution()).toBeVisible();
    await expect(pluginPage.pluginDetailDistribution()).toHaveText('项目内建');
    await expect(pluginPage.pluginBuiltinDetailAlert()).toBeVisible();
    await expect(pluginPage.pluginBuiltinDetailAlert()).toContainText(
      '项目内建能力',
    );
    await expectNoRawPluginI18nKeys(adminPage);
    await captureEvidence(adminPage, 'builtin-plugin-management-detail');
    expect(pageErrors).toEqual([]);
  });
});
