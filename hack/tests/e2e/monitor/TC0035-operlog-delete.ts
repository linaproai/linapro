import { test, expect } from '../../fixtures/auth';

test.describe('TC0035 操作日志批量删除', () => {
  test('TC0035a: 删除按钮在未勾选记录时置灰', async ({ adminPage }) => {
    await adminPage.goto('/monitor/operlog');
    await adminPage.waitForLoadState('networkidle');
    await adminPage.waitForTimeout(500);

    const deleteBtn = adminPage.getByRole('button', { name: /删\s*除/ });
    await expect(deleteBtn).toBeDisabled();
  });

  test('TC0035b: 勾选记录后删除按钮可点击并执行删除', async ({ adminPage }) => {
    await adminPage.goto('/monitor/operlog');
    await adminPage.waitForLoadState('networkidle');
    await adminPage.waitForTimeout(500);

    // Click the first row checkbox
    const firstCheckbox = adminPage.locator('.vxe-table--body .vxe-checkbox--icon').first();
    await firstCheckbox.click();

    // Delete button should now be enabled
    const deleteBtn = adminPage.getByRole('button', { name: /删\s*除/ });
    await expect(deleteBtn).toBeEnabled();

    // Click delete, expect confirmation modal
    await deleteBtn.click();
    const modal = adminPage.locator('.ant-modal-confirm');
    await expect(modal).toBeVisible();
    await expect(modal).toContainText('确认删除');

    // Confirm delete
    const okBtn = modal.getByRole('button', { name: /确\s*定/ });
    await okBtn.click();
    await adminPage.waitForTimeout(500);
  });
});
