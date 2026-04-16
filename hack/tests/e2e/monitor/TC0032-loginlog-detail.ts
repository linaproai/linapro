import { test, expect } from '../../fixtures/auth';

test.describe('TC0032 登录日志详情查看', () => {
  test('TC0032a: 点击详情按钮打开详情弹窗', async ({ adminPage }) => {
    await adminPage.goto('/monitor/loginlog');
    await adminPage.waitForLoadState('networkidle');
    await adminPage.waitForTimeout(1000);

    const rows = adminPage.locator('.vxe-body--row');
    const rowCount = await rows.count();
    if (rowCount === 0) {
      test.skip(true, 'No login logs to test');
      return;
    }

    // Click the detail button on the first row
    const detailBtn = adminPage.getByRole('button', { name: /详\s*情/ }).first();
    await detailBtn.click();
    await adminPage.waitForTimeout(1000);

    // Modal should be visible with detail content
    await expect(adminPage.locator('text=登录日志详情')).toBeVisible();
    const modal = adminPage.getByLabel('登录日志详情');
    await expect(modal.locator('text=用户账号')).toBeVisible();
    await expect(modal.locator('text=登录状态')).toBeVisible();
  });
});
