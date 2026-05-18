import type { Page, Route } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { PluginPage } from '../../../pages/PluginPage';

const autoPlanPluginID = 'plugin-dependency-auto-plan-e2e';
const blockedPluginID = 'plugin-dependency-blocked-e2e';
const basePluginID = 'plugin-dependency-base-e2e';
const consumerPluginID = 'plugin-dependency-consumer-e2e';
const installNetworkFailurePluginID = 'plugin-dependency-install-network-failure-e2e';
const uninstallNetworkFailurePluginID = 'plugin-dependency-uninstall-network-failure-e2e';

type PluginRow = Record<string, unknown>;
type DependencyCheck = Record<string, unknown>;

function apiEnvelope(data: unknown) {
  return {
    code: 0,
    data,
    message: 'success',
  };
}

function pluginRow(input: {
  description: string;
  id: string;
  installed: number;
  name: string;
}): PluginRow {
  return {
    authorizationRequired: 0,
    authorizationStatus: 'not_required',
    autoEnableForNewTenants: false,
    autoEnableManaged: 0,
    authorizedHostServices: [],
    declaredRoutes: [],
    dependencyCheck: null,
    description: input.description,
    enabled: 0,
    hasMockData: 0,
    id: input.id,
    installMode: 'global',
    installed: input.installed,
    installedAt: '',
    name: input.name,
    requestedHostServices: [],
    scopeNature: 'global',
    statusKey: input.installed === 1 ? 'disabled' : 'not_installed',
    supportsMultiTenant: false,
    type: 'source',
    updatedAt: '',
    version: 'v0.1.0',
  };
}

function emptyDependencyCheck(pluginId: string): DependencyCheck {
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

function autoInstallPlanCheck(): DependencyCheck {
  return {
    ...emptyDependencyCheck(autoPlanPluginID),
    autoInstallPlan: [
      {
        chain: [autoPlanPluginID, basePluginID],
        name: 'Dependency Base',
        pluginId: basePluginID,
        requiredBy: autoPlanPluginID,
        version: 'v0.1.0',
      },
    ],
    dependencies: [
      {
        chain: [autoPlanPluginID, basePluginID],
        currentVersion: 'v0.1.0',
        dependencyId: basePluginID,
        dependencyName: 'Dependency Base',
        discovered: true,
        installMode: 'auto',
        installed: false,
        ownerId: autoPlanPluginID,
        required: true,
        requiredVersion: '>=0.1.0',
        status: 'auto_install_planned',
      },
    ],
  };
}

function installBlockerCheck(): DependencyCheck {
  return {
    ...emptyDependencyCheck(blockedPluginID),
    blockers: [
      {
        chain: [blockedPluginID, basePluginID],
        code: 'dependency_version_unsatisfied',
        currentVersion: 'v0.1.0',
        dependencyId: basePluginID,
        pluginId: blockedPluginID,
        requiredVersion: '>=0.3.0',
      },
    ],
    dependencies: [
      {
        chain: [blockedPluginID, basePluginID],
        currentVersion: 'v0.1.0',
        dependencyId: basePluginID,
        dependencyName: 'Dependency Base',
        discovered: true,
        installMode: 'manual',
        installed: true,
        ownerId: blockedPluginID,
        required: true,
        requiredVersion: '>=0.3.0',
        status: 'version_unsatisfied',
      },
    ],
  };
}

function reverseBlockerCheck(): DependencyCheck {
  return {
    ...emptyDependencyCheck(basePluginID),
    reverseBlockers: [
      {
        chain: [consumerPluginID, basePluginID],
        code: 'reverse_dependency',
        dependencyId: basePluginID,
        pluginId: consumerPluginID,
        requiredVersion: '>=0.1.0',
      },
    ],
    reverseDependents: [
      {
        name: 'Consumer Plugin',
        pluginId: consumerPluginID,
        requiredVersion: '>=0.1.0',
        version: 'v0.1.0',
      },
    ],
  };
}

async function mockPluginDependencyApis(
  page: Page,
  rows: PluginRow[],
  checks: Record<string, DependencyCheck>,
  failingDependencyPluginIds: string[] = [],
) {
  const failingPluginIdSet = new Set(failingDependencyPluginIds);
  await page.route('**/api/v1/plugins**', async (route: Route) => {
    const request = route.request();
    const url = new URL(request.url());
    const path = url.pathname;
    const dependencyMatch = path.match(/\/api\/v1\/plugins\/([^/]+)\/dependencies$/u);

    if (request.method() === 'GET' && dependencyMatch) {
      const pluginId = decodeURIComponent(dependencyMatch[1] ?? '');
      if (failingPluginIdSet.has(pluginId)) {
        await route.abort('failed');
        return;
      }
      await route.fulfill({
        json: apiEnvelope(checks[pluginId] ?? emptyDependencyCheck(pluginId)),
      });
      return;
    }

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

    await route.continue();
  });
}

test.describe('TC-235 插件依赖管理展示', () => {
  test('TC-235a: 安装确认展示服务端返回的自动安装计划', async ({
    adminPage,
  }) => {
    await mockPluginDependencyApis(
      adminPage,
      [
        pluginRow({
          description: 'Used by E2E to verify automatic dependency plan display.',
          id: autoPlanPluginID,
          installed: 0,
          name: 'Dependency Auto Plan Plugin',
        }),
      ],
      { [autoPlanPluginID]: autoInstallPlanCheck() },
    );

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(autoPlanPluginID);
    await pluginPage.openInstallAuthorization(autoPlanPluginID);

    await expect(pluginPage.pluginDependencyAutoInstallPlan()).toBeVisible();
    await expect(pluginPage.pluginDependencyAutoInstallPlan()).toContainText(
      '宿主将先安装以下依赖插件。',
    );
    await expect(pluginPage.pluginDependencyAutoInstallPlan()).toContainText(
      'Dependency Base@v0.1.0',
    );
    await expect(pluginPage.hostServiceAuthConfirmButton()).toBeEnabled();
  });

  test('TC-235b: 安装确认展示依赖阻断并禁用提交', async ({
    adminPage,
  }) => {
    await mockPluginDependencyApis(
      adminPage,
      [
        pluginRow({
          description: 'Used by E2E to verify dependency blockers.',
          id: blockedPluginID,
          installed: 0,
          name: 'Dependency Blocked Plugin',
        }),
      ],
      { [blockedPluginID]: installBlockerCheck() },
    );

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(blockedPluginID);
    await pluginPage.openInstallAuthorization(blockedPluginID);

    await expect(pluginPage.pluginDependencyBlockers()).toBeVisible();
    await expect(pluginPage.pluginDependencyBlockers()).toContainText(
      '请先处理依赖阻断项',
    );
    await expect(pluginPage.pluginDependencyBlockers()).toContainText(
      '依赖版本不满足',
    );
    await expect(pluginPage.pluginDependencyBlockers()).toContainText(
      basePluginID,
    );
    await expect(pluginPage.hostServiceAuthConfirmButton()).toBeDisabled();
    await expect(
      pluginPage.hostServiceAuthInstallAndEnableButton(),
    ).toBeDisabled();
  });

  test('TC-235c: 卸载确认展示反向依赖阻断并禁用提交', async ({
    adminPage,
  }) => {
    await mockPluginDependencyApis(
      adminPage,
      [
        pluginRow({
          description: 'Used by E2E to verify reverse dependency blockers.',
          id: basePluginID,
          installed: 1,
          name: 'Dependency Base',
        }),
      ],
      { [basePluginID]: reverseBlockerCheck() },
    );

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(basePluginID);
    await pluginPage.openUninstallDialog(basePluginID);

    await expect(pluginPage.pluginDependencyReverseBlockers()).toBeVisible();
    await expect(pluginPage.pluginDependencyReverseBlockers()).toContainText(
      '该插件仍被已安装插件依赖。',
    );
    await expect(pluginPage.pluginDependencyReverseBlockers()).toContainText(
      'Consumer Plugin >=0.1.0',
    );
    await expect(pluginPage.uninstallConfirmButton()).toBeDisabled();
  });

  test('TC-235d: 安装弹窗依赖检查网络失败时只显示本地刷新失败提示', async ({
    adminPage,
  }) => {
    await mockPluginDependencyApis(
      adminPage,
      [
        pluginRow({
          description: 'Used by E2E to verify dependency failure toast handling.',
          id: installNetworkFailurePluginID,
          installed: 0,
          name: 'Dependency Install Network Failure Plugin',
        }),
      ],
      {},
      [installNetworkFailurePluginID],
    );

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(installNetworkFailurePluginID);
    await pluginPage.openInstallAuthorization(installNetworkFailurePluginID);

    await expect(
      pluginPage.messageNotice('刷新插件依赖检查结果失败'),
    ).toBeVisible();
    await expect(
      pluginPage.messageNotice('网络异常，请检查您的网络连接后重试。'),
    ).toHaveCount(0);
  });

  test('TC-235e: 卸载弹窗依赖检查网络失败时只显示本地刷新失败提示', async ({
    adminPage,
  }) => {
    await mockPluginDependencyApis(
      adminPage,
      [
        pluginRow({
          description: 'Used by E2E to verify dependency failure toast handling.',
          id: uninstallNetworkFailurePluginID,
          installed: 1,
          name: 'Dependency Uninstall Network Failure Plugin',
        }),
      ],
      {},
      [uninstallNetworkFailurePluginID],
    );

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(uninstallNetworkFailurePluginID);
    await pluginPage.openUninstallDialog(uninstallNetworkFailurePluginID);

    await expect(
      pluginPage.messageNotice('刷新插件依赖检查结果失败'),
    ).toBeVisible();
    await expect(
      pluginPage.messageNotice('网络异常，请检查您的网络连接后重试。'),
    ).toHaveCount(0);
  });
});
