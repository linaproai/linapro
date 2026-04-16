import { test, expect } from '../../fixtures/auth';
import { UserPage } from '../../pages/UserPage';

interface DeptTreeNode {
  id: number;
  label: string;
  userCount: number;
  children?: DeptTreeNode[];
}

test.describe('TC0022 用户管理部门树含子部门用户', () => {
  test('TC0022a: 选择父级部门时用户列表包含子部门用户', async ({
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

    // Find a parent dept that has children and userCount > 0
    function findParentWithChildren(
      nodes: DeptTreeNode[],
    ): DeptTreeNode | null {
      for (const node of nodes) {
        if (
          node.id !== 0 &&
          node.children &&
          node.children.length > 0 &&
          node.userCount > 0
        ) {
          return node;
        }
        if (node.children) {
          const found = findParentWithChildren(node.children);
          if (found) return found;
        }
      }
      return null;
    }

    const parentDept = findParentWithChildren(treeNodes);
    if (!parentDept) {
      test.skip();
      return;
    }

    // Extract dept name (without the count suffix)
    const nameMatch = parentDept.label.match(/^(.+)\(\d+\)$/);
    const deptName = nameMatch ? nameMatch[1]! : parentDept.label;

    // Click the parent dept node in the tree
    const requestPromise = adminPage.waitForResponse(
      (resp) =>
        resp.url().includes('/api/v1/user') &&
        !resp.url().includes('/api/v1/user/') &&
        resp.url().includes(`deptId=${parentDept.id}`) &&
        resp.status() === 200,
      { timeout: 15000 },
    );

    // Find and click the tree node by its label text
    const treeNode = adminPage
      .locator('.ant-tree-node-content-wrapper .ant-tree-title')
      .filter({ hasText: deptName });
    await treeNode.first().click();

    const userListResponse = await requestPromise;
    const userListBody = await userListResponse.json();
    const total = userListBody.data?.total ?? userListBody.total ?? 0;

    // The total users returned should match the dept tree's userCount
    // (which includes self + all descendants)
    expect(
      total,
      `Parent dept "${deptName}" (id=${parentDept.id}): API returned ${total} users, tree shows ${parentDept.userCount}`,
    ).toBe(parentDept.userCount);
  });
});
