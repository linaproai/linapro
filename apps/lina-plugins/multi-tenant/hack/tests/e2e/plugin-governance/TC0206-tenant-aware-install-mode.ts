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
import { scenarioTC0206 } from '../../support/multi-tenant-scenarios';
import { execPgSQL, pgEscapeLiteral, queryPgScalar } from '@host-tests/support/postgres';

const pluginId = 'monitor-loginlog';

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
    execPgSQL(`
      UPDATE sys_plugin
      SET install_mode = 'tenant_scoped'
      WHERE plugin_id = '${pgEscapeLiteral(pluginId)}';
    `);
  } finally {
    await api.dispose();
  }
}

async function restorePluginDefaultInstallMode() {
  const api = await createAdminApiContext();
  try {
    await ensureMultiTenantPluginEnabled(api);
    await syncPlugins(api);
    const plugin = await getPlugin(api, pluginId);
    if (plugin.installed !== 1) {
      await installPlugin(api, pluginId);
    }
    execPgSQL(`
      UPDATE sys_plugin
      SET install_mode = 'tenant_scoped'
      WHERE plugin_id = '${pgEscapeLiteral(pluginId)}';
    `);
  } finally {
    await api.dispose();
  }
}

test.describe('TC-206 tenant-aware 插件 install_mode', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-206a: tenant-aware plugin registry exposes controllable install mode', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0206();
  });

  test('TC-206b: tenant-aware install mode is selected inside the install confirmation dialog', async ({
    adminPage,
    multiTenantMode,
  }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await adminPage.setViewportSize({ width: 1366, height: 900 });
    await resetPluginForUiInstall();

    const pluginPage = new PluginPage(adminPage);
    try {
      await pluginPage.gotoManage();
      await pluginPage.searchByPluginId(pluginId);
      await pluginPage.expectTableColumnBetween(
        '支持多租户',
        '示例数据',
        '新租户启用',
      );
      await pluginPage.expectBooleanTableCell(
        pluginPage.pluginSupportsMultiTenantValue(pluginId),
        true,
      );
      await pluginPage.openInstallAuthorization(pluginId);

      await expect(pluginPage.installModeStandaloneSelector()).toHaveCount(0);
      await expect(pluginPage.pluginInstallModeSection()).toBeVisible();
      await pluginPage.expectInstallModeSectionDashedBorder();
      await expect(pluginPage.pluginInstallModeSelect()).toContainText('租户级');
      await expect(pluginPage.pluginInstallModeDescription()).toContainText(
        '各租户管理员',
      );
      await pluginPage.expectInstallModeDescriptionWithoutBorder();
      await pluginPage.expectInstallModeDescriptionAfterSelect();

      await pluginPage.selectInstallMode('全局');
      await expect(pluginPage.pluginInstallModeDescription()).toContainText(
        '所有租户共享启用状态',
      );
      await pluginPage.expectInstallModeDescriptionWithoutBorder();
      await pluginPage.expectInstallModeDescriptionAfterSelect();

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
      await restorePluginDefaultInstallMode();
    }
  });
});
