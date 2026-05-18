import { test, expect } from '../../fixtures/auth';
import { config } from '../../fixtures/config';
import { LoginPage } from '../../pages/LoginPage';
import { MainLayout } from '../../pages/MainLayout';

test.describe('TC0003 登出流程', () => {
  test('TC0003a: 登出后跳转到登录页', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);

    const mainLayout = new MainLayout(page);
    await mainLayout.logout();
    expect(page.url()).toContain('/auth/login');
  });
});
