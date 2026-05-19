import type { Page, Route } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { PluginPage } from '../../../pages/PluginPage';

const installPluginID = 'plugin-dev-linapro-ops-demo-guard-install-toast-e2e';
const uninstallPluginID = 'plugin-dev-linapro-ops-demo-guard-uninstall-toast-e2e';
const demoControlMessage = '演示模式已开启，禁止执行写操作';
const demoControlMessageKey = 'error.demo.control.write.denied';

type PluginRowInput = {
  id: string;
  installed: 0 | 1;
  name: string;
};

function apiEnvelope(data: unknown) {
  return {
    code: 0,
    data,
    message: 'success',
  };
}

function demoControlRejection() {
  return {
    code: 403,
    data: null,
    errorCode: 'DEMO_CONTROL_WRITE_DENIED',
    message: demoControlMessage,
    messageKey: demoControlMessageKey,
    messageParams: {},
  };
}

function pluginRow(input: PluginRowInput) {
  return {
    abnormalReason: '',
    authorizationRequired: 0,
    authorizationStatus: 'not_required',
    autoEnableForNewTenants: false,
    autoEnableManaged: 0,
    authorizedHostServices: [],
    declaredRoutes: [],
    dependencyCheck: emptyDependencyCheck(input.id),
    description: 'Used by E2E to verify linapro-ops-demo-guard toast ownership.',
    discoveredVersion: 'v0.1.0',
    effectiveVersion: 'v0.1.0',
    enabled: 0,
    hasMockData: 0,
    id: input.id,
    installMode: 'global',
    installed: input.installed,
    installedAt: '',
    lastUpgradeFailure: undefined,
    name: input.name,
    requestedHostServices: [],
    runtimeState: 'normal',
    scopeNature: 'global',
    statusKey: input.installed === 1 ? 'disabled' : 'not_installed',
    supportsMultiTenant: false,
    type: 'source',
    updatedAt: '',
    upgradeAvailable: false,
    version: 'v0.1.0',
  };
}

function emptyDependencyCheck(pluginId: string) {
  return {
    autoInstallPlan: [],
    autoInstalled: [],
    blockers: [],
    cycle: [],
    dependencies: [],
    framework: {
      currentVersion: 'v0.6.0',
      requiredVersion: '',
      status: 'not_declared',
    },
    manualInstallRequired: [],
    reverseBlockers: [],
    reverseDependents: [],
    softUnsatisfied: [],
    targetId: pluginId,
  };
}

async function mockPluginLifecycleApis(page: Page) {
  const rows = [
    pluginRow({
      id: installPluginID,
      installed: 0,
      name: 'Demo Control Install Toast E2E',
    }),
    pluginRow({
      id: uninstallPluginID,
      installed: 1,
      name: 'Demo Control Uninstall Toast E2E',
    }),
  ];

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

    if (
      request.method() === 'GET' &&
      path.endsWith(`/plugins/${installPluginID}/dependencies`)
    ) {
      await route.fulfill({
        json: apiEnvelope(emptyDependencyCheck(installPluginID)),
      });
      return;
    }

    if (
      request.method() === 'GET' &&
      path.endsWith(`/plugins/${uninstallPluginID}/dependencies`)
    ) {
      await route.fulfill({
        json: apiEnvelope(emptyDependencyCheck(uninstallPluginID)),
      });
      return;
    }

    if (
      request.method() === 'POST' &&
      path.endsWith(`/plugins/${installPluginID}/install`)
    ) {
      await route.fulfill({
        json: demoControlRejection(),
        status: 403,
      });
      return;
    }

    if (
      request.method() === 'DELETE' &&
      path.endsWith(`/plugins/${uninstallPluginID}`)
    ) {
      await route.fulfill({
        json: demoControlRejection(),
        status: 403,
      });
      return;
    }

    await route.continue();
  });
}

async function expectSingleDemoControlToast(pluginPage: PluginPage) {
  await expect(pluginPage.messageNotices(demoControlMessage)).toHaveCount(1);
  await expect(pluginPage.messageNotices('网络异常，请检查您的网络连接后重试。')).toHaveCount(0);
  await expect(pluginPage.messageNotices('服务器内部错误')).toHaveCount(0);
}

test.describe('TC-238 linapro-ops-demo-guard 插件生命周期错误提示', () => {
  test('TC-238a: 安装弹窗提交被 linapro-ops-demo-guard 拦截时只显示一条只读错误', async ({
    adminPage,
  }) => {
    await mockPluginLifecycleApis(adminPage);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(installPluginID);
    await pluginPage.openInstallAuthorization(installPluginID);
    await expect(pluginPage.messageNotices(demoControlMessage)).toHaveCount(0);
    await pluginPage.hostServiceAuthConfirmButton().click();

    await expectSingleDemoControlToast(pluginPage);
    await expect(pluginPage.hostServiceAuthDialog()).toBeVisible();
  });

  test('TC-238b: 卸载弹窗提交被 linapro-ops-demo-guard 拦截时只显示一条只读错误', async ({
    adminPage,
  }) => {
    await mockPluginLifecycleApis(adminPage);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(uninstallPluginID);
    await expect(pluginPage.messageNotices(demoControlMessage)).toHaveCount(0);
    await pluginPage.openUninstallDialogAndConfirm(uninstallPluginID);

    await expectSingleDemoControlToast(pluginPage);
    await expect(pluginPage.uninstallDialog()).toBeVisible();
  });
});
