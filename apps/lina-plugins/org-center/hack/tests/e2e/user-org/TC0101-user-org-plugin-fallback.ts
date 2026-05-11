import { test, expect } from '@host-tests/fixtures/auth';
import {
  ensureSourcePluginDisabled,
  ensureSourcePluginEnabled,
} from '@host-tests/fixtures/plugin';
import { UserPage } from '@host-tests/pages/UserPage';

test.describe('TC-101 用户页面组织插件降级', () => {
  test('TC-101a: 组织插件停用时用户页面隐藏部门树与部门岗位字段', async ({
    adminPage,
  }) => {
    await ensureSourcePluginDisabled(adminPage, 'org-center');

    const userPage = new UserPage(adminPage);
    await userPage.goto();

    await expect(adminPage.locator('.ant-tree')).toHaveCount(0);

    await adminPage.getByRole('button', { name: /新\s*增/ }).click();
    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'visible', timeout: 5000 });

    await expect(drawer.getByText('部门', { exact: true })).toHaveCount(0);
    await expect(drawer.getByText('岗位', { exact: true })).toHaveCount(0);
  });

  test('TC-101b: 组织插件启用后用户页面恢复部门树与部门岗位字段', async ({
    adminPage,
  }) => {
    await ensureSourcePluginEnabled(adminPage, 'org-center');

    const userPage = new UserPage(adminPage);
    await userPage.goto();

    await expect(adminPage.locator('.ant-tree')).toBeVisible({ timeout: 10000 });

    await adminPage.getByRole('button', { name: /新\s*增/ }).click();
    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'visible', timeout: 5000 });

    await expect(drawer.getByText('部门', { exact: true })).toBeVisible();
    await expect(drawer.getByText('岗位', { exact: true })).toBeVisible();
  });
});
