import type { Page, Route } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { PluginPage } from '../../../pages/PluginPage';

const upgradePluginID = 'plugin-runtime-upgrade-e2e';
const abnormalPluginID = 'plugin-runtime-abnormal-e2e';

type PluginRow = Record<string, unknown>;

function apiEnvelope(data: unknown) {
  return {
    code: 0,
    data,
    message: 'success',
  };
}

function pluginRow(input: {
  abnormalReason?: string;
  discoveredVersion: string;
  effectiveVersion: string;
  id: string;
  name: string;
  runtimeState: string;
  upgradeAvailable: boolean;
}): PluginRow {
  return {
    abnormalReason: input.abnormalReason,
    authorizationRequired: 0,
    authorizationStatus: 'not_required',
    autoEnableForNewTenants: false,
    autoEnableManaged: 0,
    authorizedHostServices: [],
    declaredRoutes: [],
    dependencyCheck: null,
    description: 'Used by E2E to verify explicit plugin runtime upgrade.',
    discoveredVersion: input.discoveredVersion,
    effectiveVersion: input.effectiveVersion,
    enabled: 0,
    hasMockData: 0,
    id: input.id,
    installMode: 'global',
    installed: 1,
    installedAt: '',
    lastUpgradeFailure:
      input.runtimeState === 'upgrade_failed'
        ? {
            detail: 'Previous upgrade SQL failed.',
            errorCode: 'plugin_upgrade_migration_failed',
            messageKey: 'plugin.runtimeUpgrade.failure.migrationFailed',
            phase: 'sql',
            releaseId: 23,
            releaseVersion: input.discoveredVersion,
          }
        : undefined,
    name: input.name,
    requestedHostServices: [],
    runtimeState: input.runtimeState,
    scopeNature: 'global',
    statusKey: 'disabled',
    supportsMultiTenant: false,
    type: 'dynamic',
    updatedAt: '',
    upgradeAvailable: input.upgradeAvailable,
    version: input.effectiveVersion,
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

function manifestSnapshot(input: {
  description: string;
  id: string;
  name: string;
  version: string;
}) {
  return {
    authorizedHostServices: [],
    backendHookCount: 0,
    defaultInstallMode: 'global',
    description: input.description,
    frontendPageCount: 0,
    frontendSlotCount: 0,
    hostServiceAuthConfirmed: true,
    hostServiceAuthRequired: false,
    id: input.id,
    installSqlCount: 1,
    manifestDeclared: true,
    menuCount: 1,
    mockSqlCount: 1,
    name: input.name,
    requestedHostServices: [],
    resourceSpecCount: 0,
    routeCount: 1,
    routeExecutionEnabled: true,
    runtimeFrontendAssetCount: 0,
    runtimeKind: 'wasm',
    runtimeSqlAssetCount: 1,
    scopeNature: 'global',
    supportsMultiTenant: false,
    type: 'dynamic',
    uninstallSqlCount: 1,
    version: input.version,
  };
}

function upgradePreview() {
  return {
    dependencyCheck: emptyDependencyCheck(upgradePluginID),
    discoveredVersion: 'v0.2.0',
    effectiveVersion: 'v0.1.0',
    fromManifest: manifestSnapshot({
      description: 'Effective manifest before upgrade.',
      id: upgradePluginID,
      name: 'Runtime Upgrade E2E',
      version: 'v0.1.0',
    }),
    hostServicesDiff: {
      added: [],
      authorizationChanged: false,
      authorizationRequired: false,
      changed: [],
      removed: [],
    },
    pluginId: upgradePluginID,
    riskHints: [
      'plugin.runtimeUpgrade.risk.upgradeSqlRequiresReview',
      'plugin.runtimeUpgrade.risk.mockSqlExcluded',
    ],
    runtimeState: 'pending_upgrade',
    sqlSummary: {
      installSqlCount: 1,
      mockSqlCount: 1,
      runtimeSqlAssetCount: 1,
      uninstallSqlCount: 1,
    },
    toManifest: manifestSnapshot({
      description: 'Target manifest after upgrade.',
      id: upgradePluginID,
      name: 'Runtime Upgrade E2E',
      version: 'v0.2.0',
    }),
  };
}

async function mockPluginRuntimeUpgradeApis(page: Page) {
  let upgraded = false;

  await page.route('**/api/v1/plugins**', async (route: Route) => {
    const request = route.request();
    const url = new URL(request.url());
    const path = url.pathname;

    if (
      request.method() === 'GET' &&
      path.endsWith(`/plugins/${upgradePluginID}/upgrade/preview`)
    ) {
      await route.fulfill({ json: apiEnvelope(upgradePreview()) });
      return;
    }

    if (
      request.method() === 'POST' &&
      path.endsWith(`/plugins/${upgradePluginID}/upgrade`)
    ) {
      upgraded = true;
      await route.fulfill({
        json: apiEnvelope({
          discoveredVersion: 'v0.2.0',
          effectiveVersion: 'v0.2.0',
          executed: true,
          fromVersion: 'v0.1.0',
          pluginId: upgradePluginID,
          runtimeState: 'normal',
          toVersion: 'v0.2.0',
        }),
      });
      return;
    }

    if (request.method() === 'GET' && /\/api\/v1\/plugins$/u.test(path)) {
      const id = url.searchParams.get('id')?.trim();
      const rows = [
        upgraded
          ? pluginRow({
              discoveredVersion: 'v0.2.0',
              effectiveVersion: 'v0.2.0',
              id: upgradePluginID,
              name: 'Runtime Upgrade E2E',
              runtimeState: 'normal',
              upgradeAvailable: false,
            })
          : pluginRow({
              discoveredVersion: 'v0.2.0',
              effectiveVersion: 'v0.1.0',
              id: upgradePluginID,
              name: 'Runtime Upgrade E2E',
              runtimeState: 'pending_upgrade',
              upgradeAvailable: true,
            }),
        pluginRow({
          abnormalReason: 'discovered_version_lower_than_effective',
          discoveredVersion: 'v0.1.0',
          effectiveVersion: 'v0.2.0',
          id: abnormalPluginID,
          name: 'Runtime Abnormal E2E',
          runtimeState: 'abnormal',
          upgradeAvailable: false,
        }),
      ];
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

test.describe('TC-236 插件运行时升级', () => {
  test('TC-236a: 待升级插件展示升级动作并在确认后刷新为正常状态', async ({
    adminPage,
  }) => {
    await mockPluginRuntimeUpgradeApis(adminPage);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(upgradePluginID);

    await expect(pluginPage.pluginRuntimeState(upgradePluginID)).toContainText(
      /待升级|Pending Upgrade/iu,
    );
    await expect(pluginPage.pluginVersionValue(upgradePluginID)).toContainText(
      'v0.1.0 -> v0.2.0',
    );

    await pluginPage.openRuntimeUpgradeDialog(upgradePluginID);
    await expect(pluginPage.pluginUpgradeFromManifest()).toContainText('v0.1.0');
    await expect(pluginPage.pluginUpgradeToManifest()).toContainText('v0.2.0');
    await expect(pluginPage.pluginUpgradeSqlSummary()).toContainText(
      /安装\/升级 SQL：1|Install\/upgrade SQL: 1/iu,
    );
    await expect(pluginPage.pluginUpgradeRiskSectionTitle()).toBeVisible();

    await pluginPage.confirmRuntimeUpgrade();
    await expect(pluginPage.pluginRuntimeState(upgradePluginID)).toContainText(
      /正常|Normal/iu,
    );
    await expect(pluginPage.pluginVersionValue(upgradePluginID)).toContainText(
      'v0.2.0',
    );
  });

  test('TC-236b: 异常插件展示人工修复提示且不显示升级确认动作', async ({
    adminPage,
  }) => {
    await mockPluginRuntimeUpgradeApis(adminPage);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(abnormalPluginID);

    await expect(pluginPage.pluginRuntimeState(abnormalPluginID)).toContainText(
      /异常|Abnormal/iu,
    );
    await expect(pluginPage.pluginManualRepairTag(abnormalPluginID)).toContainText(
      /人工修复|Manual Repair/iu,
    );
    await expect(pluginPage.pluginVersionValue(abnormalPluginID)).toContainText(
      'v0.2.0 -> v0.1.0',
    );
  });
});
