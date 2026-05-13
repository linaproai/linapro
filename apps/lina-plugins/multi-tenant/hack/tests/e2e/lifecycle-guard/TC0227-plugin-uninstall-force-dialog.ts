import { test, expect } from '../../support/multi-tenant';
import {
  createAdminApiContext,
  enablePlugin,
  getPlugin,
  installPlugin,
  syncPlugins,
} from '@host-tests/support/api/job';
import { createTenant, deleteTenant } from '../../support/multi-tenant';
import { PluginPage } from '@host-tests/pages/PluginPage';

test.describe('TC-227 多租户插件卸载保护弹窗', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-227a: lifecycle guard veto opens localized force confirmation dialog and resubmits with force', async ({
    adminPage,
    multiTenantMode,
  }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');

    const api = await createAdminApiContext();
    const pluginId = 'multi-tenant';
    const tenant = await createTenant(api, {
      code: `tc227-${Date.now()}`.slice(0, 32),
      name: 'TC227 Tenant',
    });
    try {
      await syncPlugins(api);
      const plugin = await getPlugin(api, pluginId);
      if (plugin.installed !== 1) {
        await installPlugin(api, pluginId);
      }
      if (plugin.enabled !== 1) {
        await enablePlugin(api, pluginId);
      }

      const pluginPage = new PluginPage(adminPage);
      await pluginPage.gotoManage();
      await pluginPage.searchByPluginId(pluginId);

      const vetoResponsePromise = adminPage.waitForResponse(
        (response) =>
          response.url().includes(`/plugins/${pluginId}`) &&
          response.request().method() === 'DELETE',
      );
      await pluginPage.openUninstallDialogAndConfirm(pluginId);

      const vetoResponse = await vetoResponsePromise;
      expect(vetoResponse.url()).not.toContain('force=true');
      expect((await vetoResponse.json()).errorCode).toBe(
        'PLUGIN_LIFECYCLE_GUARD_VETOED',
      );

      await expect(pluginPage.lifecycleGuardDialog()).toBeVisible();
      await expect(pluginPage.uninstallDialog()).toHaveCount(0);
      await expect(pluginPage.lifecycleGuardDialog()).toHaveCSS('gap', '10px');
      await expect(pluginPage.lifecycleGuardReasonAlert()).not.toContainText(
        '插件返回了阻止当前操作的原因。',
      );
      await expect(pluginPage.lifecycleGuardReasonText()).toContainText(
        '当前插件阻止操作，原因：',
      );
      await expect(pluginPage.lifecycleGuardReasonText()).toContainText(
        '仍存在租户，请先删除租户，再卸载插件。',
      );
      await expect(pluginPage.lifecycleGuardForceAlert()).toContainText(
        '强制卸载会绕过上述保护并清理插件数据，请确认你理解该风险。',
      );
      await expect(pluginPage.lifecycleGuardForceAlert()).toContainText(
        '输入插件 ID "multi-tenant" 以启用强制卸载。',
      );
      await expect(pluginPage.lifecycleGuardDialog()).toContainText(pluginId);
      await expect(pluginPage.lifecycleGuardConfirmButton()).toBeDisabled();

      await pluginPage.lifecycleGuardForcePluginIdInput().fill(pluginId);
      await expect(pluginPage.lifecycleGuardConfirmButton()).toBeEnabled();

      const forceResponsePromise = adminPage.waitForResponse(
        (response) =>
          response.url().includes(`/plugins/${pluginId}`) &&
          response.url().includes('force=true') &&
          response.request().method() === 'DELETE',
      );
      await pluginPage.lifecycleGuardConfirmButton().click();
      const forceResponse = await forceResponsePromise;
      const forcePayload = await forceResponse.json();
      expect(forcePayload.code).toBe(0);
      await expect(pluginPage.lifecycleGuardDialog()).toHaveCount(0);

      const pluginAfterForce = await getPlugin(api, pluginId);
      expect(pluginAfterForce.installed).toBe(0);
    } finally {
      await deleteTenant(api, tenant.id).catch(() => {});
      const plugin = await getPlugin(api, pluginId).catch(() => null);
      if (plugin?.installed !== 1) {
        await installPlugin(api, pluginId).catch(() => {});
      }
      const refreshed = await getPlugin(api, pluginId).catch(() => null);
      if (refreshed?.enabled !== 1) {
        await enablePlugin(api, pluginId).catch(() => {});
      }
      await api.dispose();
    }
  });
});
