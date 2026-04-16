import { test, expect } from '../../fixtures/auth';

test.describe('TC0029 操作日志导出', () => {
  test('TC0029a: 导出全部数据', async ({ adminPage }) => {
    await adminPage.goto('/monitor/operlog');
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
      (resp) => resp.url().includes('operlog/export'),
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

  test('TC0029b: 导出选中数据', async ({ adminPage }) => {
    await adminPage.goto('/monitor/operlog');
    await adminPage.waitForLoadState('networkidle');
    await adminPage.waitForTimeout(500);

    // Select a row (click checkbox)
    const firstCheckbox = adminPage.locator('.vxe-checkbox--icon').first();
    await firstCheckbox.click();
    await adminPage.waitForTimeout(300);

    // Click export button
    const exportBtn = adminPage.getByRole('button', { name: /导\s*出/ });
    await exportBtn.click();

    // Verify modal appears with selection text
    const modalContent = adminPage.locator('.ant-modal-content');
    await expect(modalContent).toBeVisible({ timeout: 5000 });
    await expect(modalContent.getByText(/是否导出选中的记录/)).toBeVisible();

    // Set up response listener
    const responsePromise = adminPage.waitForResponse(
      (resp) => resp.url().includes('operlog/export'),
      { timeout: 15000 }
    );

    // Click confirm button
    const confirmBtn = modalContent.getByRole('button', { name: /确\s*认/ });
    await confirmBtn.click();

    // Wait for response and verify
    const response = await responsePromise;
    expect(response.status()).toBe(200);
  });
});
