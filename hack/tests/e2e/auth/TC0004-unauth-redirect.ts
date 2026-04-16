import { test, expect } from '../../fixtures/auth';

test.describe('TC0004 未登录访问保护页面', () => {
  test('TC0004a: 未登录访问宿主工作区重定向到登录页', async ({ page }) => {
    await page.goto('/dashboard/analytics');
    await page.waitForURL(/auth\/login/, { timeout: 10000 });
    expect(page.url()).toContain('/auth/login');
  });

  test('TC0004b: 未登录访问用户管理重定向到登录页', async ({ page }) => {
    await page.goto('/system/user');
    await page.waitForURL(/auth\/login/, { timeout: 10000 });
    expect(page.url()).toContain('/auth/login');
  });
});
