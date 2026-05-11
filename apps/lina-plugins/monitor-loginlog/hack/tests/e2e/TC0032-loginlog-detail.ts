import { test, expect } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import { waitForDialogReady, waitForTableReady } from '@host-tests/support/ui';

test.describe('TC0032 登录日志详情查看', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'monitor-loginlog');
  });

  test('TC0032a: 点击详情按钮打开详情弹窗', async ({ adminPage }) => {
    await adminPage.goto('/monitor/loginlog');
    await waitForTableReady(adminPage);

    const rows = adminPage.locator('.vxe-body--row');
    const rowCount = await rows.count();
    if (rowCount === 0) {
      test.skip(true, 'No login logs to test');
      return;
    }

    // Click the detail button on the first row
    const detailBtn = adminPage.getByRole('button', { name: /详\s*情/ }).first();
    await detailBtn.click();

    // Modal should be visible with detail content
    await expect(adminPage.locator('text=登录日志详情')).toBeVisible();
    const modal = await waitForDialogReady(adminPage.getByLabel('登录日志详情'));
    await expect(modal.locator('text=用户账号')).toBeVisible();
    await expect(modal.locator('text=登录状态')).toBeVisible();
  });
});
