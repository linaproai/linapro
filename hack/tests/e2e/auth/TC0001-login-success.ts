import { LoginPage } from '../../pages/LoginPage';
import { config } from '../../fixtures/config';
import { test, expect } from '../../fixtures/auth';

test.describe('TC0001 登录验证', () => {
  test('TC0001a: 登录后跳转到宿主工作区', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);
    expect(page.url()).not.toContain('/auth/login');
  });

  test('TC0001b: 登录后页面正常加载', async ({ adminPage }) => {
    // Should see the main layout (sidebar or dashboard content)
    await expect(
      adminPage.locator('body'),
    ).not.toHaveText(/404|error|not found/i);
  });
});
