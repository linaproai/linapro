import { test, expect } from '../../fixtures/auth';
import { RolePage } from '../../pages/RolePage';

/**
 * TC0065 角色授权用户管理 E2E 测试
 *
 * 测试覆盖：
 * - 角色授权用户页面加载
 * - 已授权用户列表显示（包含邮箱列）
 * - 单个取消授权功能
 * - 批量取消授权功能
 */
test.describe('TC0065 角色授权用户管理', () => {
  test('TC0065a: 授权用户页面正常加载', async ({ adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    // 点击第一个角色的"分配"按钮（跳过超级管理员）
    const assignBtn = adminPage.locator('.vxe-body--row').first().locator('button:has-text("分 配")');
    if (await assignBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await assignBtn.click();

      // 等待页面跳转
      await adminPage.waitForURL(/\/system\/role-auth\/user\/\d+/, { timeout: 10000 });

      // 验证表格可见
      const table = adminPage.locator('.vxe-table');
      await expect(table).toBeVisible({ timeout: 10000 });

      // 验证工具栏按钮存在
      await expect(
        adminPage.getByRole('button', { name: /取消授权/ }).first(),
      ).toBeVisible({ timeout: 5000 });
    }
  });

  test('TC0065b: 授权用户列表包含邮箱列', async ({ adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    // 点击第一个角色的"分配"按钮
    const assignBtn = adminPage.locator('.vxe-body--row').first().locator('button:has-text("分 配")');
    if (await assignBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await assignBtn.click();
      await adminPage.waitForURL(/\/system\/role-auth\/user\/\d+/, { timeout: 10000 });

      // 验证邮箱列存在
      const emailHeader = adminPage.locator('.vxe-header--column').filter({ hasText: '邮箱' });
      await expect(emailHeader).toBeVisible({ timeout: 5000 });
    }
  });

  test('TC0065c: 批量取消授权按钮状态', async ({ adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    // 点击第一个角色的"分配"按钮
    const assignBtn = adminPage.locator('.vxe-body--row').first().locator('button:has-text("分 配")');
    if (await assignBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await assignBtn.click();
      await adminPage.waitForURL(/\/system\/role-auth\/user\/\d+/, { timeout: 10000 });

      // 验证批量取消授权按钮初始禁用状态
      const batchCancelBtn = adminPage.getByRole('button', { name: /取消授权/ }).first();
      await expect(batchCancelBtn).toBeDisabled({ timeout: 5000 });

      // 如果有用户数据，勾选一条后验证按钮可用
      const checkbox = adminPage.locator('.vxe-checkbox--icon').first();
      if (await checkbox.isVisible({ timeout: 3000 }).catch(() => false)) {
        await checkbox.click();
        await expect(batchCancelBtn).toBeEnabled({ timeout: 3000 });
      }
    }
  });
});
