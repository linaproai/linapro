import { test, expect } from '../../fixtures/auth';

test.describe('TC0051 在线用户强制下线', () => {
  test.beforeEach(async ({ adminPage }) => {
    const responsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/api/v1/monitor/online/list') &&
        res.request().method() === 'GET' &&
        res.status() === 200,
      { timeout: 15000 },
    );
    await adminPage.goto('/monitor/online');
    await responsePromise;
    await adminPage.waitForTimeout(500);
  });

  test('TC0051a: 强制下线按钮显示确认弹窗', async ({ adminPage }) => {
    // Click the force logout button
    const forceLogoutBtn = adminPage
      .getByRole('button', { name: /强制下线/ })
      .first();
    await expect(forceLogoutBtn).toBeVisible();
    await forceLogoutBtn.click();

    // Popconfirm should appear
    await expect(
      adminPage.getByText(/确认强制下线/),
    ).toBeVisible();
  });

  test('TC0051b: 取消强制下线不执行操作', async ({ adminPage }) => {
    const rowsBefore = await adminPage.locator('.vxe-body--row').count();

    // Click force logout
    await adminPage
      .getByRole('button', { name: /强制下线/ })
      .first()
      .click();

    // Cancel the popconfirm
    const cancelBtn = adminPage.getByRole('button', { name: /取\s*消/ }).or(
      adminPage.locator('.ant-popconfirm .ant-btn:not(.ant-btn-primary)'),
    );
    await cancelBtn.first().click();
    await adminPage.waitForTimeout(500);

    // Row count should remain the same
    const rowsAfter = await adminPage.locator('.vxe-body--row').count();
    expect(rowsAfter).toBe(rowsBefore);
  });
});
