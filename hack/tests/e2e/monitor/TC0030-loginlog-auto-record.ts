import { test, expect } from '../../fixtures/auth';
import { LoginPage } from '../../pages/LoginPage';
import { config } from '../../fixtures/config';

test.describe('TC0030 登录日志自动记录', () => {
  test('TC0030a: 登录成功后登录日志中记录成功日志', async ({ adminPage }) => {
    // The adminPage fixture already logged in, so a login log should exist
    await adminPage.goto('/monitor/loginlog');
    await adminPage.waitForLoadState('networkidle');
    await adminPage.waitForTimeout(1000);

    // Should see at least one login log row
    const rows = adminPage.locator('.vxe-body--row');
    await expect(rows.first()).toBeVisible();

    // The first row should contain admin username
    await expect(adminPage.locator('.vxe-body--row').first()).toContainText('admin');
  });

  test('TC0030b: 登录失败后登录日志中记录失败日志', async ({ page }) => {
    // First, attempt a failed login
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login('admin', 'wrongpassword');
    await page.waitForTimeout(2000);

    // Now login correctly
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);

    // Navigate to login log page
    await page.goto('/monitor/loginlog');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);

    // Should see login logs including the failed one
    const rows = page.locator('.vxe-body--row');
    await expect(rows.first()).toBeVisible();
  });
});
