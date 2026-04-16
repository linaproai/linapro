import { LoginPage } from '../../pages/LoginPage';
import { MainLayout } from '../../pages/MainLayout';
import { config } from '../../fixtures/config';
import { test, expect } from '../../fixtures/auth';

test.describe('TC0003 登出流程', () => {
  test('TC0003a: 登出后跳转到登录页', async ({ adminPage }) => {
    const mainLayout = new MainLayout(adminPage);
    await mainLayout.logout();
    expect(adminPage.url()).toContain('/auth/login');
  });
});
