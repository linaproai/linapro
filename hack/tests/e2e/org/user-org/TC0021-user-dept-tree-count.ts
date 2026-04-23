import { test, expect } from '../../../fixtures/auth';
import { ensureSourcePluginEnabled } from '../../../fixtures/plugin';
import { UserPage } from '../../../pages/UserPage';

interface DeptTreeNode {
  id: number;
  label: string;
  userCount: number;
  children?: DeptTreeNode[];
}

test.describe('TC0021 用户管理部门树用户数量累加', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'org-center');
  });

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
    test.setTimeout(120_000);

    const userPage = new UserPage(adminPage);
    const username = `tc0021-${Date.now()}`;
    const password = 'Admin123!';

    await userPage.goto();
    await userPage.createUser(username, password, 'TC0021');

    try {
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
      expect(beforeCounts['未分配部门']).toBeGreaterThan(0);

      await adminPage.evaluate(async ({ username }) => {
        const accessKey = Object.keys(localStorage).find((key) =>
          key.endsWith('core-access'),
        );
        const accessStateText = accessKey ? localStorage.getItem(accessKey) : null;
        const accessState = accessStateText ? JSON.parse(accessStateText) : null;
        const accessToken = accessState?.accessToken;
        if (!accessToken) {
          throw new Error('Missing access token in localStorage.');
        }

        const headers = {
          Authorization: `Bearer ${accessToken}`,
          'Content-Type': 'application/json',
        };
        const parsePayload = async (response: Response) => {
          const payload = await response.json();
          return payload?.data ?? payload;
        };
        const listResponse = await fetch(
          `/api/v1/user?pageNum=1&pageSize=20&username=${encodeURIComponent(username)}`,
          { headers },
        );
        if (!listResponse.ok) {
          throw new Error(`List user failed: ${listResponse.status}`);
        }
        const listPayload = await parsePayload(listResponse);
        const user = (listPayload?.list ?? [])[0];
        if (!user?.id) {
          throw new Error('Created user was not found.');
        }

        const deptTreeResponse = await fetch('/api/v1/user/dept-tree', { headers });
        if (!deptTreeResponse.ok) {
          throw new Error(`Load dept tree failed: ${deptTreeResponse.status}`);
        }
        const deptTreePayload = await parsePayload(deptTreeResponse);

        const findFirstDept = (nodes: any[]): any => {
          for (const node of nodes ?? []) {
            if (node?.id && node.id !== 0) {
              return node;
            }
            const childNode = findFirstDept(node?.children ?? []);
            if (childNode) {
              return childNode;
            }
          }
          return null;
        };

        const targetDept = findFirstDept(deptTreePayload?.list ?? []);
        if (!targetDept?.id) {
          throw new Error('No real department node available.');
        }

        const detailResponse = await fetch(`/api/v1/user/${user.id}`, { headers });
        if (!detailResponse.ok) {
          throw new Error(`Load user detail failed: ${detailResponse.status}`);
        }
        const detail = await parsePayload(detailResponse);

        const updatePayload = {
          id: user.id,
          deptId: targetDept.id,
          email: detail?.email ?? '',
          nickname: detail?.nickname ?? username,
          password: '',
          phone: detail?.phone ?? '',
          postIds: [],
          remark: detail?.remark ?? '',
          roleIds: detail?.roleIds ?? [],
          sex: Number(detail?.sex ?? 0),
          status: Number(detail?.status ?? 1),
        };

        const updateResponse = await fetch(`/api/v1/user/${user.id}`, {
          body: JSON.stringify(updatePayload),
          headers,
          method: 'PUT',
        });
        if (!updateResponse.ok) {
          throw new Error(`Update user failed: ${updateResponse.status}`);
        }
      }, { username });

      await userPage.goto();

      // Verify dept tree is still visible and the unassigned count is refreshed.
      await expect(deptTree).toBeVisible();
      const afterCounts = await getNodeCounts();
      expect(Object.keys(afterCounts).length).toBeGreaterThan(0);
      expect(afterCounts['未分配部门']).toBe(beforeCounts['未分配部门'] - 1);
    } finally {
      if (adminPage.isClosed()) {
        return;
      }
      await adminPage.evaluate(async ({ username }) => {
        const accessKey = Object.keys(localStorage).find((key) =>
          key.endsWith('core-access'),
        );
        const accessStateText = accessKey ? localStorage.getItem(accessKey) : null;
        const accessState = accessStateText ? JSON.parse(accessStateText) : null;
        const accessToken = accessState?.accessToken;
        if (!accessToken) {
          return;
        }

        const headers = {
          Authorization: `Bearer ${accessToken}`,
        };
        const listResponse = await fetch(
          `/api/v1/user?pageNum=1&pageSize=20&username=${encodeURIComponent(username)}`,
          { headers },
        );
        if (!listResponse.ok) {
          return;
        }
        const listPayload = await listResponse.json();
        const user = (listPayload?.data?.list ?? listPayload?.list ?? [])[0];
        if (!user?.id) {
          return;
        }

        await fetch(`/api/v1/user/${user.id}`, {
          headers,
          method: 'DELETE',
        });
      }, { username });
    }
  });
});
