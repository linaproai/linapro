import { test, expect } from '../../fixtures/auth';

test.describe('TC0028 操作日志清理', () => {
  test('TC0028a: 点击清空按钮弹出确认对话框', async ({ adminPage }) => {
    await adminPage.goto('/monitor/operlog');
    await adminPage.waitForLoadState('networkidle');
    await adminPage.waitForTimeout(500);

    const cleanBtn = adminPage.getByRole('button', { name: /清\s*空/ });
    await cleanBtn.click();
    await adminPage.waitForTimeout(300);

    // Confirm modal should appear
    const modal = adminPage.locator('.ant-modal-confirm');
    await expect(modal).toBeVisible();
    await expect(modal.locator('text=确认要清空所有操作日志数据吗')).toBeVisible();

    // Cancel to close
    const cancelBtn = modal.getByRole('button', { name: /取\s*消/ });
    await cancelBtn.click();
  });

  test('TC0028b: 确认清空操作成功执行', async ({ adminPage }) => {
    await adminPage.goto('/monitor/operlog');
    await adminPage.waitForLoadState('networkidle');
    await adminPage.waitForTimeout(500);

    const cleanBtn = adminPage.getByRole('button', { name: /清\s*空/ });
    await cleanBtn.click();
    await adminPage.waitForTimeout(300);

    const modal = adminPage.locator('.ant-modal-confirm');
    await expect(modal).toBeVisible();

    // Set up response interception before clicking OK
    const responsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/api/v1/operlog/clean') && res.request().method() === 'DELETE',
      { timeout: 10000 },
    );

    // Click OK to confirm (Ant Design uses "确定")
    const okBtn = modal.getByRole('button', { name: /确\s*定|OK/ });
    await okBtn.click();

    const response = await responsePromise;
    expect(response.status()).toBe(200);
  });
});
