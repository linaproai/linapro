import { test, expect } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import {
  waitForBusyIndicatorsToClear,
  waitForDialogReady,
  waitForTableReady,
} from '@host-tests/support/ui';

test.describe('TC0029 操作日志导出', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'monitor-operlog');
  });

  test('TC0029a: 导出全部数据', async ({ adminPage }) => {
    await adminPage.goto('/monitor/operlog');
    await waitForTableReady(adminPage);

    // Click export button
    const exportBtn = adminPage.getByRole('button', { name: /导\s*出/ });
    await expect(exportBtn).toBeVisible({ timeout: 10000 });
    await exportBtn.click();

    // Verify modal appears
    const modalContent = await waitForDialogReady(adminPage.locator('.ant-modal-wrap'));
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
    await waitForBusyIndicatorsToClear(adminPage);
    const contentType = response.headers()['content-type'];
    expect(contentType).toContain('application/vnd.openxmlformats-officedocument.spreadsheetml.sheet');
  });

  test('TC0029b: 导出选中数据', async ({ adminPage }) => {
    await adminPage.goto('/monitor/operlog');
    await waitForTableReady(adminPage);

    // Select a row (click checkbox)
    const firstCheckbox = adminPage.locator('.vxe-checkbox--icon').first();
    await firstCheckbox.click();
    await waitForBusyIndicatorsToClear(adminPage);

    // Click export button
    const exportBtn = adminPage.getByRole('button', { name: /导\s*出/ });
    await exportBtn.click();

    // Verify modal appears with selection text
    const modalContent = await waitForDialogReady(adminPage.locator('.ant-modal-wrap'));
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
    await waitForBusyIndicatorsToClear(adminPage);
  });
});
