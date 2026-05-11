import { test, expect } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import {
  waitForConfirmOverlay,
  waitForRouteReady,
  waitForTableReady,
} from '@host-tests/support/ui';

test.describe('TC0036 登录日志批量删除', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'monitor-loginlog');
  });

  test('TC0036a: 删除按钮在未勾选记录时置灰', async ({ adminPage }) => {
    await adminPage.goto('/monitor/loginlog');
    await waitForTableReady(adminPage);

    const deleteBtn = adminPage.getByRole('button', { name: /删\s*除/ });
    await expect(deleteBtn).toBeDisabled();
  });

  test('TC0036b: 勾选记录后删除按钮可点击并执行删除', async ({ adminPage }) => {
    await adminPage.goto('/monitor/loginlog');
    await waitForTableReady(adminPage);

    // Click the first row checkbox
    const firstCheckbox = adminPage.locator('.vxe-table--body .vxe-checkbox--icon').first();
    await firstCheckbox.click();

    // Delete button should now be enabled
    const deleteBtn = adminPage.getByRole('button', { name: /删\s*除/ });
    await expect(deleteBtn).toBeEnabled();

    // Click delete, expect confirmation modal
    await deleteBtn.click();
    const modal = await waitForConfirmOverlay(adminPage);
    await expect(modal).toContainText('确认删除');

    // Confirm delete
    const okBtn = modal.getByRole('button', { name: /确\s*定/ });
    await okBtn.click();
    await waitForRouteReady(adminPage);
  });
});
