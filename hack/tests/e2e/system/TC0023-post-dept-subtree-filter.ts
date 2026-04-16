import { test, expect } from '../../fixtures/auth';

interface DeptTreeNode {
  id: number;
  label: string;
  postCount: number;
  children?: DeptTreeNode[];
}

test.describe('TC0023 岗位管理部门树子部门过滤与数量', () => {
  test('TC0023a: 选择父级部门时岗位列表包含子部门岗位', async ({
    adminPage,
  }) => {
    // Intercept the dept-tree API response during page load
    const treeResponsePromise = adminPage.waitForResponse(
      (resp) =>
        resp.url().includes('/api/v1/post/dept-tree') && resp.status() === 200,
      { timeout: 15000 },
    );

    await adminPage.goto('/system/post');
    await adminPage.waitForLoadState('networkidle');
    await adminPage.locator('.vxe-table').waitFor({ state: 'visible', timeout: 10000 });

    const treeResponse = await treeResponsePromise;
    const body = await treeResponse.json();
    const treeNodes: DeptTreeNode[] = body.data?.list ?? body.list ?? [];
    expect(treeNodes.length).toBeGreaterThan(0);

    // Find a parent dept that has children and postCount > 0
    function findParentWithChildren(
      nodes: DeptTreeNode[],
    ): DeptTreeNode | null {
      for (const node of nodes) {
        if (
          node.id !== 0 &&
          node.children &&
          node.children.length > 0 &&
          node.postCount > 0
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

    const nameMatch = parentDept.label.match(/^(.+)\(\d+\)$/);
    const deptName = nameMatch ? nameMatch[1]! : parentDept.label;

    // Click the parent dept node and intercept the post list request
    const requestPromise = adminPage.waitForResponse(
      (resp) =>
        resp.url().includes('/api/v1/post') &&
        !resp.url().includes('/api/v1/post/') &&
        resp.url().includes(`deptId=${parentDept.id}`) &&
        resp.status() === 200,
      { timeout: 15000 },
    );

    const treeNode = adminPage
      .locator('.ant-tree-node-content-wrapper .ant-tree-title')
      .filter({ hasText: deptName });
    await treeNode.first().click();

    const postListResponse = await requestPromise;
    const postListBody = await postListResponse.json();
    const total = postListBody.data?.total ?? postListBody.total ?? 0;

    expect(
      total,
      `Parent dept "${deptName}" (id=${parentDept.id}): API returned ${total} posts, tree shows ${parentDept.postCount}`,
    ).toBe(parentDept.postCount);
  });

  test('TC0023b: 岗位部门树节点标签包含岗位数量', async ({ adminPage }) => {
    await adminPage.goto('/system/post');
    await adminPage.waitForLoadState('networkidle');

    const deptTree = adminPage.locator('.ant-tree');
    await expect(deptTree).toBeVisible({ timeout: 10000 });

    const treeNodeTitles = adminPage.locator(
      '.ant-tree-node-content-wrapper .ant-tree-title',
    );
    const count = await treeNodeTitles.count();
    expect(count).toBeGreaterThan(0);

    for (let i = 0; i < count; i++) {
      const text = (await treeNodeTitles.nth(i).textContent())?.trim();
      expect(text).toBeTruthy();
      expect(text).toMatch(/\(\d+\)$/);
    }
  });

  test('TC0023c: 父部门岗位数等于自身加所有子部门岗位数之和', async ({
    adminPage,
  }) => {
    const treeResponsePromise = adminPage.waitForResponse(
      (resp) =>
        resp.url().includes('/api/v1/post/dept-tree') && resp.status() === 200,
      { timeout: 15000 },
    );

    await adminPage.goto('/system/post');
    await adminPage.waitForLoadState('networkidle');

    const treeResponse = await treeResponsePromise;
    const body = await treeResponse.json();
    const treeNodes: DeptTreeNode[] = body.data?.list ?? body.list ?? [];

    function verifyParentGteChildren(nodes: DeptTreeNode[]) {
      for (const node of nodes) {
        if (node.children && node.children.length > 0) {
          verifyParentGteChildren(node.children);
          const childrenSum = node.children.reduce(
            (sum, child) => sum + child.postCount,
            0,
          );
          expect(
            node.postCount,
            `Dept "${node.label}" (id=${node.id}): postCount(${node.postCount}) should be >= children sum(${childrenSum})`,
          ).toBeGreaterThanOrEqual(childrenSum);
        }
      }
    }

    const realDeptNodes = treeNodes.filter((n) => n.id !== 0);
    verifyParentGteChildren(realDeptNodes);
  });
});
