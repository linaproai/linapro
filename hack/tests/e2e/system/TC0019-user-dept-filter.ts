import { test, expect } from '../../fixtures/auth';
import { UserPage } from '../../pages/UserPage';

test.describe('TC0019 用户管理部门过滤', () => {
  test('TC0019a: 用户管理页面左侧部门树可见', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Verify the DeptTree component is rendered on the left side
    const deptTree = adminPage.locator('.ant-tree');
    await expect(deptTree).toBeVisible({ timeout: 10000 });

    // Verify at least one tree node is rendered
    const treeNodes = adminPage.locator('.ant-tree-node-content-wrapper');
    const nodeCount = await treeNodes.count();
    expect(nodeCount).toBeGreaterThan(0);
  });

  test('TC0019b: 选择部门后用户列表按部门过滤', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Set up request interception to verify deptId is included
    const requestPromise = adminPage.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/user') &&
        !req.url().includes('/api/v1/user/') &&
        req.method() === 'GET',
      { timeout: 15000 },
    );

    // Click a dept node in the left tree
    const deptNode = adminPage
      .locator('.ant-tree-node-content-wrapper')
      .first();
    await deptNode.click();
    await adminPage.waitForLoadState('networkidle');

    const request = await requestPromise;

    // Verify the request URL includes deptId parameter
    expect(request.url()).toContain('deptId');
  });
});
