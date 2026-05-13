import { test, expect } from '../../support/multi-tenant';
import { ensureMultiTenantPluginEnabled } from '../../support/multi-tenant';
import { PluginPage } from '@host-tests/pages/PluginPage';
import {
  createAdminApiContext,
  disablePlugin,
  expectSuccess,
  getPlugin,
  installPlugin,
  syncPlugins,
} from '@host-tests/support/api/job';
import { scenarioTC0207 } from '../../support/multi-tenant-scenarios';
import { pgEscapeLiteral, queryPgScalar } from '@host-tests/support/postgres';

const pluginId = 'monitor-server';

async function resetPluginForUiInstall() {
  const api = await createAdminApiContext();
  try {
    await ensureMultiTenantPluginEnabled(api);
    await syncPlugins(api);
    const plugin = await getPlugin(api, pluginId);
    if (plugin.enabled === 1) {
      await disablePlugin(api, pluginId);
    }
    if (plugin.installed === 1) {
      await expectSuccess(
        await api.delete(`plugins/${pluginId}?purgeStorageData=0`),
      );
    }
  } finally {
    await api.dispose();
  }
}

async function restorePluginInstalled() {
  const api = await createAdminApiContext();
  try {
    await ensureMultiTenantPluginEnabled(api);
    await syncPlugins(api);
    const plugin = await getPlugin(api, pluginId);
    if (plugin.installed !== 1) {
      await installPlugin(api, pluginId);
    }
  } finally {
    await api.dispose();
  }
}

test.describe('TC-207 platform-only 插件强制 global', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-207a: platform-only plugin remains global and hidden from tenant plugin API', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0207();
  });

  test('TC-207b: platform-only install mode is shown and locked inside the install confirmation dialog', async ({
    adminPage,
    multiTenantMode,
  }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await adminPage.setViewportSize({ width: 1366, height: 900 });
    await resetPluginForUiInstall();

    const pluginPage = new PluginPage(adminPage);
    const api = await createAdminApiContext();
    try {
      await pluginPage.gotoManage();
      await pluginPage.searchByPluginId(pluginId);
      await expect(pluginPage.pluginListHelpIcon()).toHaveCount(0);
      await expect(pluginPage.pluginColumnHelpIcon('type')).toBeVisible();
      await expect(pluginPage.pluginColumnHelpIcon('mockData')).toBeVisible();
      await expect(
        pluginPage.pluginColumnHelpIcon('supportsMultiTenant'),
      ).toBeVisible();
      await expect(
        pluginPage.pluginColumnHelpIcon('tenantProvisioning'),
      ).toBeVisible();
      await pluginPage.expectColumnHelpTooltip('type', '源码插件随宿主源码');
      await pluginPage.expectColumnHelpTooltip('mockData', '示例数据');
      await pluginPage.expectColumnHelpTooltip(
        'supportsMultiTenant',
        '支持多租户治理',
      );
      await pluginPage.expectColumnHelpTooltip(
        'tenantProvisioning',
        '新租户创建时是否自动启用',
      );
      const plugin = await getPlugin(api, pluginId);
      expect(plugin.supportsMultiTenant).toBe(false);
      expect(plugin.autoEnableForNewTenants).toBe(false);
      expect(
        queryPgScalar(`
          SELECT COUNT(1)
          FROM sys_plugin_state
          WHERE plugin_id = '${pgEscapeLiteral(pluginId)}'
            AND tenant_id > 0;
        `),
      ).toBe('0');
      await pluginPage.expectTableColumnBetween(
        '支持多租户',
        '示例数据',
        '新租户启用',
      );
      await pluginPage.expectBooleanTableCell(
        pluginPage.pluginSupportsMultiTenantValue(pluginId),
        false,
      );
      await pluginPage.expectTenantProvisioningDisabled(pluginId);
      await pluginPage.openInstallAuthorization(pluginId);

      await expect(pluginPage.installModeStandaloneSelector()).toHaveCount(0);
      await expect(pluginPage.pluginInstallModeSection()).toBeVisible();
      await pluginPage.expectInstallModeSectionDashedBorder();
      await expect(pluginPage.pluginInstallModeSelect()).toContainText('全局');
      await expect(pluginPage.pluginInstallModeSelect()).toHaveClass(
        /ant-select-disabled/,
      );
      await expect(pluginPage.pluginInstallModeDescription()).toContainText(
        '所有租户共享启用状态',
      );
      await pluginPage.expectInstallModeDescriptionWithoutBorder();
      await pluginPage.expectInstallModeDescriptionAfterSelect();
      await expect(pluginPage.pluginInstallModeSelect()).not.toContainText(
        '租户级',
      );
      await expect(pluginPage.pluginInstallModeSection()).toContainText(
        '不支持租户级安装模式的插件只能以全局模式安装',
      );
      await pluginPage.expectInstallModePlatformOnlyHintGap();

      const installResponsePromise = adminPage.waitForResponse(
        (response) =>
          response.url().includes(`/plugins/${pluginId}/install`) &&
          response.request().method() === 'POST',
      );
      await pluginPage.hostServiceAuthConfirmButton().click();

      const installResponse = await installResponsePromise;
      expect(installResponse.status()).toBe(200);
      expect(installResponse.request().postDataJSON()).toMatchObject({
        installMode: 'global',
      });
      await expect(pluginPage.hostServiceAuthDialog()).toHaveCount(0);
      expect(
        queryPgScalar(
          `SELECT install_mode FROM sys_plugin WHERE plugin_id = '${pgEscapeLiteral(pluginId)}';`,
        ),
      ).toBe('global');
    } finally {
      await api.dispose();
      await restorePluginInstalled();
    }
  });
});
