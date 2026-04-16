import { test, expect } from '../../fixtures/auth';

test.describe('TC0050 在线用户搜索过滤', () => {
  test.beforeEach(async ({ adminPage }) => {
    const responsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/api/v1/monitor/online/list') &&
        res.request().method() === 'GET' &&
        res.status() === 200,
      { timeout: 15000 },
    );
    await adminPage.goto('/monitor/online');
    await responsePromise;
    await adminPage.waitForTimeout(500);
  });

  test('TC0050a: 按用户名搜索能过滤结果', async ({ adminPage }) => {
    // Fill username search
    await adminPage
      .getByLabel('用户账号', { exact: true })
      .first()
      .fill('admin');

    // Click search and wait for response
    const requestPromise = adminPage.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/monitor/online/list') &&
        req.url().includes('username=admin'),
    );
    await adminPage.getByRole('button', { name: /搜\s*索/ }).click();
    await requestPromise;
    await adminPage.waitForTimeout(500);

    // Results should still show admin
    const rows = adminPage.locator('.vxe-body--row');
    const count = await rows.count();
    expect(count).toBeGreaterThan(0);
  });

  test('TC0050b: 搜索不存在的用户返回空结果', async ({ adminPage }) => {
    // Fill a non-existent username
    await adminPage
      .getByLabel('用户账号', { exact: true })
      .first()
      .fill('nonexistent_user_xyz');

    // Click search and wait for response
    const responsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/api/v1/monitor/online/list') &&
        res.status() === 200,
    );
    await adminPage.getByRole('button', { name: /搜\s*索/ }).click();
    await responsePromise;
    await adminPage.waitForTimeout(500);

    // Should show empty or no rows
    const rows = adminPage.locator('.vxe-body--row');
    const count = await rows.count();
    expect(count).toBe(0);
  });

  test('TC0050c: 重置搜索恢复完整列表', async ({ adminPage }) => {
    // Fill and search
    await adminPage
      .getByLabel('用户账号', { exact: true })
      .first()
      .fill('nonexistent_user_xyz');
    await adminPage.getByRole('button', { name: /搜\s*索/ }).click();
    await adminPage.waitForTimeout(500);

    // Reset
    const responsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/api/v1/monitor/online/list') &&
        res.status() === 200,
    );
    await adminPage.getByRole('button', { name: /重\s*置/ }).click();
    await responsePromise;
    await adminPage.waitForTimeout(500);

    // Should have results again
    const rows = adminPage.locator('.vxe-body--row');
    const count = await rows.count();
    expect(count).toBeGreaterThan(0);
  });
});
