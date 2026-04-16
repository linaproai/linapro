import { test, expect } from '../../fixtures/auth';
import { UserPage } from '../../pages/UserPage';

test.describe('TC0010 重置密码', () => {
  test('TC0010a: 点击更多菜单中的重置密码打开弹窗', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Click "更多" dropdown on first row
    await adminPage.getByRole('button', { name: '更多' }).first().click();
    await adminPage.waitForTimeout(300);

    // Click "重置密码"
    await adminPage.getByText('重置密码').click();

    // Verify the reset password dialog opens
    const dialog = adminPage.getByRole('dialog').filter({ hasText: '重置密码' });
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Verify password input exists
    await expect(dialog.getByPlaceholder(/请输入新的密码/)).toBeVisible();
  });

  test('TC0010b: 重置密码API调用成功', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Set up response listener BEFORE triggering the action
    const responsePromise = adminPage.waitForResponse(
      (res) => res.url().includes('/reset-password') && res.request().method() === 'PUT',
      { timeout: 15000 },
    );

    // Click "更多" dropdown on first row
    await adminPage.getByRole('button', { name: '更多' }).first().click();
    await adminPage.waitForTimeout(300);

    // Click "重置密码"
    await adminPage.getByText('重置密码').click();

    const dialog = adminPage.getByRole('dialog').filter({ hasText: '重置密码' });
    await expect(dialog).toBeVisible({ timeout: 5000 });

    // Fill new password
    await dialog.getByPlaceholder(/请输入新的密码/).fill('NewPass12345');

    // Click confirm
    await dialog.getByRole('button', { name: /确\s*认/ }).click();

    const response = await responsePromise;
    expect(response.status()).toBe(200);
  });
});
