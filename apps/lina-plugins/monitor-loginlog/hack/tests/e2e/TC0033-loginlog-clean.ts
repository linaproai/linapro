import { test, expect } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import {
  waitForConfirmOverlay,
  waitForRouteReady,
  waitForTableReady,
} from '@host-tests/support/ui';

test.describe('TC0033 登录日志清理', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'monitor-loginlog');
  });

  test('TC0033a: 点击清空按钮弹出确认对话框', async ({ adminPage }) => {
    await adminPage.goto('/monitor/loginlog');
    await waitForTableReady(adminPage);

    const cleanBtn = adminPage.getByRole('button', { name: /清\s*空/ });
    await cleanBtn.click();

    const modal = await waitForConfirmOverlay(adminPage);
    await expect(modal.getByText(/确认清空所有登录日志吗[?？]?/)).toBeVisible();

    // Cancel to close
    const cancelBtn = modal.getByRole('button', { name: /取\s*消/ });
    await cancelBtn.click();
    await modal.waitFor({ state: 'hidden', timeout: 5000 }).catch(() => {});
  });

  test('TC0033b: 确认清空操作成功执行', async ({ adminPage }) => {
    await adminPage.goto('/monitor/loginlog');
    await waitForTableReady(adminPage);

    const cleanBtn = adminPage.getByRole('button', { name: /清\s*空/ });
    await cleanBtn.click();

    const modal = await waitForConfirmOverlay(adminPage);

    // Set up response interception before clicking OK
    const responsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/api/v1/loginlog/clean') && res.request().method() === 'DELETE',
      { timeout: 10000 },
    );

    // Click OK to confirm (Ant Design uses "确定")
    const okBtn = modal.getByRole('button', { name: /确\s*定|OK/ });
    await okBtn.click();

    const response = await responsePromise;
    expect(response.status()).toBe(200);
    await waitForRouteReady(adminPage);
  });
});
