import { LoginPage } from '../../pages/LoginPage';
import { config } from '../../fixtures/config';
import { test, expect } from '../../fixtures/auth';

test.describe('TC0002 登录失败', () => {
  test('TC0002a: 错误密码显示错误提示', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login(config.adminUser, 'wrongpassword');

    await expect(loginPage.errorMessage).toBeVisible({ timeout: 5000 });
  });

  test('TC0002b: 错误密码后仍在登录页', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.login(config.adminUser, 'wrongpassword');

    // Wait a bit for any redirect attempt
    await page.waitForTimeout(2000);
    expect(page.url()).toContain('/auth/login');
  });
});
