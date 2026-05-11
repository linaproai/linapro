import type { Page } from '@host-tests/support/playwright';

import { test, expect } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import { UserPage } from '@host-tests/pages/UserPage';

interface DeptTreeNode {
  id: number;
  label: string;
  userCount: number;
  children?: DeptTreeNode[];
}

function getDeptTreeNodes(payload: any): DeptTreeNode[] {
  return payload.data?.list ?? payload.list ?? [];
}

function findDeptNodeByID(nodes: DeptTreeNode[], id: number): DeptTreeNode | null {
  for (const node of nodes) {
    if (node.id === id) {
      return node;
    }
    const childNode = findDeptNodeByID(node.children ?? [], id);
    if (childNode) {
      return childNode;
    }
  }
  return null;
}

function getRequiredUnassignedUserCount(nodes: DeptTreeNode[]) {
  const unassignedNode = findDeptNodeByID(nodes, 0);
  expect(unassignedNode, 'Dept tree response should include the virtual Unassigned node').toBeTruthy();
  return unassignedNode!.userCount;
}

async function reloadUserPageAndReadDeptTree(page: Page, userPage: UserPage) {
  const treeResponsePromise = page.waitForResponse(
    (resp) =>
      resp.url().includes('/api/v1/user/dept-tree') && resp.status() === 200,
    { timeout: 15000 },
  );
  await userPage.goto();
  const treeResponse = await treeResponsePromise;
  return getDeptTreeNodes(await treeResponse.json());
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
    const treeNodes = getDeptTreeNodes(body);
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
      const beforeTreeNodes = await reloadUserPageAndReadDeptTree(
        adminPage,
        userPage,
      );
      expect(beforeTreeNodes.length).toBeGreaterThan(0);
      const beforeUnassignedCount =
        getRequiredUnassignedUserCount(beforeTreeNodes);
      expect(beforeUnassignedCount).toBeGreaterThan(0);

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

      const afterTreeNodes = await reloadUserPageAndReadDeptTree(
        adminPage,
        userPage,
      );
      expect(afterTreeNodes.length).toBeGreaterThan(0);
      expect(getRequiredUnassignedUserCount(afterTreeNodes)).toBe(
        beforeUnassignedCount - 1,
      );
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
