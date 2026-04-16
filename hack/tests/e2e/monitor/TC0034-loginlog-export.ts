import { test, expect } from '../../fixtures/auth';

test.describe('TC0034 登录日志导出', () => {
  test('TC0034a: 导出全部数据', async ({ adminPage }) => {
    await adminPage.goto('/monitor/loginlog');
    await adminPage.waitForLoadState('networkidle');
    await adminPage.waitForTimeout(500);

    // Click export button
    const exportBtn = adminPage.getByRole('button', { name: /导\s*出/ });
    await expect(exportBtn).toBeVisible({ timeout: 10000 });
    await exportBtn.click();

    // Verify modal appears
    const modalContent = adminPage.locator('.ant-modal-content');
    await expect(modalContent).toBeVisible({ timeout: 5000 });
    await expect(modalContent.getByText(/是否导出全部数据/)).toBeVisible();

    // Set up response listener
    const responsePromise = adminPage.waitForResponse(
      (resp) => resp.url().includes('loginlog/export'),
      { timeout: 15000 }
    );

    // Click confirm button
    const confirmBtn = modalContent.getByRole('button', { name: /确\s*认/ });
    await confirmBtn.click();

    // Wait for response and verify
    const response = await responsePromise;
    expect(response.status()).toBe(200);
    const contentType = response.headers()['content-type'];
    expect(contentType).toContain('application/vnd.openxmlformats-officedocument.spreadsheetml.sheet');
  });
});
