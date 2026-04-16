import { test, expect } from '../../fixtures/auth';
import { UserPage } from '../../pages/UserPage';

test.describe('TC0007 用户列表搜索', () => {
  test('TC0007a: 按用户名搜索', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Search for a known test user
    await userPage.fillSearchField('用户账号', 'user001');
    await userPage.clickSearch();

    // Verify filtered results
    const hasUser = await userPage.hasUser('user001');
    expect(hasUser).toBeTruthy();

    const rowCount = await userPage.getVisibleRowCount();
    expect(rowCount).toBeGreaterThan(0);
  });

  test('TC0007b: 按用户昵称搜索', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    await userPage.fillSearchField('用户昵称', '张伟');
    await userPage.clickSearch();

    const rowCount = await userPage.getVisibleRowCount();
    expect(rowCount).toBeGreaterThan(0);
  });

  test('TC0007c: 按手机号搜索', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    await userPage.fillSearchField('手机号码', '138');
    await userPage.clickSearch();

    const rowCount = await userPage.getVisibleRowCount();
    expect(rowCount).toBeGreaterThan(0);
  });

  test('TC0007d: 搜索后重置恢复全部数据', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Get initial row count
    const initialCount = await userPage.getVisibleRowCount();
    expect(initialCount).toBeGreaterThan(0);

    // Search with a specific term
    await userPage.fillSearchField('用户账号', 'user001');
    await userPage.clickSearch();

    const filteredCount = await userPage.getVisibleRowCount();
    expect(filteredCount).toBeLessThanOrEqual(initialCount);
    expect(filteredCount).toBeGreaterThan(0);

    // Reset
    await userPage.clickReset();

    const resetCount = await userPage.getVisibleRowCount();
    expect(resetCount).toBe(initialCount);
  });

  test('TC0007e: 搜索请求包含正确的筛选参数', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    await userPage.fillSearchField('用户昵称', '测试');

    const requestPromise = adminPage.waitForRequest(
      (req) => req.url().includes('/api/v1/user') && req.method() === 'GET' && req.url().includes('nickname'),
      { timeout: 10000 },
    );

    await userPage.clickSearch();
    const request = await requestPromise;

    const url = request.url();
    expect(url).toContain('nickname=');
  });
});
