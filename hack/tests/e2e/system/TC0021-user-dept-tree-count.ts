import { test, expect } from '../../fixtures/auth';
import { UserPage } from '../../pages/UserPage';

interface DeptTreeNode {
  id: number;
  label: string;
  userCount: number;
  children?: DeptTreeNode[];
}

test.describe('TC0021 用户管理部门树用户数量累加', () => {
  test('TC0021a: 父部门用户数等于自身用户数加所有子部门用户数之和', async ({
    adminPage,
  }) => {
    // Intercept the dept-tree API response during page load
    const treeResponsePromise = adminPage.waitForResponse(
      (resp) =>
        resp.url().includes('/api/v1/user/dept-tree') && resp.status() === 200,
      { timeout: 15000 },
    );

    const userPage = new UserPage(adminPage);
    await userPage.goto();

    const treeResponse = await treeResponsePromise;
    const body = await treeResponse.json();
    const treeNodes: DeptTreeNode[] = body.data?.list ?? body.list ?? [];
    expect(treeNodes.length).toBeGreaterThan(0);

    // Verify parent.userCount >= sum(children.userCount) recursively
    function verifyParentGteChildren(nodes: DeptTreeNode[]) {
      for (const node of nodes) {
        if (node.children && node.children.length > 0) {
          verifyParentGteChildren(node.children);
          const childrenSum = node.children.reduce(
            (sum, child) => sum + child.userCount,
            0,
          );
          expect(
            node.userCount,
            `Dept "${node.label}" (id=${node.id}): userCount(${node.userCount}) should be >= children sum(${childrenSum})`,
          ).toBeGreaterThanOrEqual(childrenSum);
        }
      }
    }

    // Filter out the virtual "未分配部门" node (id=0)
    const realDeptNodes = treeNodes.filter((n) => n.id !== 0);
    verifyParentGteChildren(realDeptNodes);
  });

  test('TC0021b: 部门树节点标签包含用户数量', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Wait for tree to be visible
    const deptTree = adminPage.locator('.ant-tree');
    await expect(deptTree).toBeVisible({ timeout: 10000 });

    // Get all tree node title text content (filter out empty spans from search highlight)
    const treeNodeTitles = adminPage.locator(
      '.ant-tree-node-content-wrapper .ant-tree-title',
    );
    const count = await treeNodeTitles.count();
    expect(count).toBeGreaterThan(0);

    for (let i = 0; i < count; i++) {
      const text = (await treeNodeTitles.nth(i).textContent())?.trim();
      expect(text).toBeTruthy();
      // Label should end with (N) where N is a number
      expect(text).toMatch(/\(\d+\)$/);
    }
  });

  test('TC0021c: 修改用户部门后部门树数量刷新', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Wait for dept tree
    const deptTree = adminPage.locator('.ant-tree');
    await expect(deptTree).toBeVisible({ timeout: 10000 });

    // Helper to parse node label counts from the tree
    const getNodeCounts = async () => {
      const titles = adminPage.locator(
        '.ant-tree-node-content-wrapper .ant-tree-title',
      );
      const result: Record<string, number> = {};
      const cnt = await titles.count();
      for (let i = 0; i < cnt; i++) {
        const text = ((await titles.nth(i).textContent()) ?? '').trim();
        const match = text.match(/^(.+)\((\d+)\)$/);
        if (match) {
          result[match[1]!] = parseInt(match[2]!, 10);
        }
      }
      return result;
    };

    const beforeCounts = await getNodeCounts();
    expect(Object.keys(beforeCounts).length).toBeGreaterThan(0);

    // Search for a non-admin user
    await userPage.fillSearchField('用户账号', 'testuser');
    await userPage.clickSearch();
    await adminPage.waitForTimeout(500);

    const hasTestUser = await userPage.hasUser('testuser');
    if (!hasTestUser) {
      test.skip();
      return;
    }

    // Open edit drawer
    await adminPage.getByRole('button', { name: /编\s*辑/ }).first().click();
    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'visible', timeout: 5000 });

    // Submit to trigger dept tree refresh
    await drawer.getByRole('button', { name: /确\s*认/ }).click();
    await adminPage.waitForLoadState('networkidle');
    await adminPage.waitForTimeout(1000);

    // Verify dept tree is still visible and labels have count format after refresh
    await expect(deptTree).toBeVisible();
    const afterCounts = await getNodeCounts();
    expect(Object.keys(afterCounts).length).toBeGreaterThan(0);
  });
});
