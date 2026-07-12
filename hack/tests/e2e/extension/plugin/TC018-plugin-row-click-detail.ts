import type { Page, Route } from '@playwright/test';

import { mkdir } from 'node:fs/promises';
import path from 'node:path';

import { test, expect } from '../../../fixtures/auth';
import { PluginPage } from '../../../pages/PluginPage';

const pluginID = 'plugin-row-click-detail-e2e';
const pluginName = 'Plugin Row Click Detail E2E';

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
    description: 'Used by E2E to verify plugin row click opens detail.',
    discoveredVersion: 'v0.1.0',
    effectiveVersion: 'v0.1.0',
    enabled: 1,
    hasMockData: 0,
    id: pluginID,
    installMode: 'global',
    installed: 1,
    installedAt: 1767240000000,
    lastUpgradeFailure: undefined,
    name: pluginName,
    requestedHostServices: [],
    runtimeState: 'normal',
    scopeNature: 'global',
    statusKey: 'enabled',
    supportsMultiTenant: false,
    type: 'source',
    updatedAt: 1767240000000,
    upgradeAvailable: false,
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

async function mockPluginRowClickApis(page: Page) {
  const row = pluginRow();
  let detailRequestCount = 0;

  await page.route('**/api/v1/plugins**', async (route: Route) => {
    const request = route.request();
    const url = new URL(request.url());
    const requestPath = url.pathname;

    if (request.method() === 'GET' && /\/api\/v1\/plugins$/u.test(requestPath)) {
      const id = url.searchParams.get('id')?.trim();
      const list = id && !row.id.includes(id) ? [] : [row];
      await route.fulfill({
        json: apiEnvelope({
          list,
          total: list.length,
        }),
      });
      return;
    }

    if (
      request.method() === 'GET' &&
      requestPath.endsWith(`/plugins/${pluginID}`)
    ) {
      detailRequestCount += 1;
      await route.fulfill({
        json: apiEnvelope({
          ...row,
          description:
            'Used by E2E to verify plugin row click opens detail. Full projection.',
        }),
      });
      return;
    }

    if (request.method() === 'GET' && requestPath.endsWith('/plugins/dynamic')) {
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

    await route.fulfill({ json: apiEnvelope({}) });
  });

  return {
    detailRequestCount: () => detailRequestCount,
  };
}

test.describe('TC-18 插件列表行点击打开详情', () => {
  test('TC-18a: 行悬停为可点击光标，点击非操作列单元格打开详情', async ({
    adminPage,
  }) => {
    const api = await mockPluginRowClickApis(adminPage);
    const pluginPage = new PluginPage(adminPage);

    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);

    await expect(pluginPage.pluginRow(pluginID)).toBeVisible();
    await pluginPage.expectPluginRowClickableCursor(pluginID);
    expect(api.detailRequestCount()).toBe(0);

    await pluginPage.openPluginDetailByRowClick(pluginID);

    await expect(pluginPage.pluginDetailModal()).toBeVisible();
    await expect(pluginPage.pluginDetailModal()).toContainText(pluginName);
    await expect(pluginPage.pluginDetailModal()).toContainText(pluginID);
    await expect(pluginPage.pluginDetailModal()).toContainText('源码插件');
    expect(api.detailRequestCount()).toBe(1);

    await captureEvidence(adminPage, 'plugin-row-click-detail');
  });

  test('TC-18b: 点击状态开关不打开详情弹窗', async ({ adminPage }) => {
    const api = await mockPluginRowClickApis(adminPage);
    const pluginPage = new PluginPage(adminPage);

    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);

    const enabledSwitch = pluginPage.pluginEnabledSwitch(pluginID);
    await expect(enabledSwitch).toBeVisible();
    await enabledSwitch.click();

    await expect(pluginPage.pluginDetailModal()).toHaveCount(0);
    expect(api.detailRequestCount()).toBe(0);
  });
});
