import type { Page } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { RolePage } from '../../../pages/RolePage';

/**
 * TC0064 角色表单默认值测试
 *
 * 验证新增角色表单的默认值配置是否正确
 */
test.describe('TC0064 角色表单默认值', () => {
  test.beforeEach(async ({ adminPage }) => {
    await mockMultiTenantPluginState(adminPage, false);
  });

  test('TC0064a: 验证新增角色表单默认值', async ({ adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    const drawer = await rolePage.openCreateDrawer();

    // 1. 验证排序字段默认值为 0
    const sortInput = drawer.getByRole('spinbutton');
    const sortValue = await sortInput.inputValue();
    console.log('Sort input value:', sortValue);
    expect(sortValue).toBe('0');

    // 2. 多租户插件未启用时，默认选中"全部数据"
    expect(await rolePage.selectedDataScopeText(drawer)).toBe('全部数据');
    expect(await rolePage.getDataScopeOptions(drawer)).not.toContain(
      '本租户数据',
    );
  });
});

async function mockMultiTenantPluginState(page: Page, enabled: boolean) {
  await page.unroute('**/api/v1/plugins/dynamic**').catch(() => {});
  await page.route('**/api/v1/plugins/dynamic**', async (route) => {
    await route.fulfill({
      contentType: 'application/json',
      body: JSON.stringify({
        code: 0,
        data: {
          list: [
            {
              enabled: enabled ? 1 : 0,
              generation: 1,
              id: 'multi-tenant',
              installed: 1,
              statusKey: 'sys_plugin.status:multi-tenant',
              version: 'e2e',
            },
          ],
        },
        message: 'success',
      }),
      status: 200,
    });
  });
  await page.evaluate(() => {
    const registryGlobal = globalThis as any;
    registryGlobal.__linaPluginStatePromise = null;
    registryGlobal.__linaPluginStateSignature = null;
  });
}
