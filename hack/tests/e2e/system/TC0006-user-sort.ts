import { test, expect } from '../../fixtures/auth';
import { UserPage } from '../../pages/UserPage';

test.describe('TC0006 用户列表排序', () => {
  test('TC0006a: 点击用户名列头可触发排序', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Verify table has data (seed data should be loaded)
    const rowCount = await userPage.getVisibleRowCount();
    expect(rowCount).toBeGreaterThan(0);

    // Click username column header to sort
    await userPage.clickColumnSort('名称');

    // Verify sort was applied by checking the sort class on visible header
    const header = adminPage.locator('.vxe-header--column.fixed--visible', { hasText: '名称' }).first();
    await expect(header.locator('.vxe-cell--sort')).toBeVisible({ timeout: 3000 });
  });

  test('TC0006b: 点击创建时间列头可触发排序', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    await userPage.clickColumnSort('创建时间');

    const header = adminPage.locator('.vxe-header--column.fixed--visible', { hasText: '创建时间' }).first();
    await expect(header.locator('.vxe-cell--sort')).toBeVisible({ timeout: 3000 });
  });

  test('TC0006c: 排序请求包含正确的排序参数', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();
    await adminPage.waitForLoadState('networkidle');

    // Set up request and response intercept BEFORE clicking
    const responsePromise = adminPage.waitForResponse(
      (res) => res.url().includes('/api/v1/user') && res.request().method() === 'GET' && res.url().includes('orderBy'),
      { timeout: 15000 },
    );

    // Click the column header to trigger sort
    await userPage.clickColumnSort('名称');

    const response = await responsePromise;
    const url = response.url();
    expect(url).toContain('orderBy=username');
    expect(url).toMatch(/orderDirection=(asc|desc)/);
  });
});
