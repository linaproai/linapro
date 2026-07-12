import type { APIRequestContext } from "@playwright/test";

import { mkdirSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { test, expect } from "../../../fixtures/auth";
import { RolePage } from "../../../pages/RolePage";
import { createAdminApiContext, expectSuccess } from "../../../support/api/job";
import { buildBatchIdsQuery } from "../../../support/api/query-ids";

const repoRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../../../..",
);

type RoleIdentity = {
  code: string;
  name: string;
};

type RoleListItem = {
  id: number;
  key: string;
  name: string;
};

type RoleListResult = {
  list: RoleListItem[];
  total: number;
};

type MenuTreeNode = {
  children?: MenuTreeNode[];
  id: number;
  label: string;
  parentId: number;
  type: string;
};

type ButtonMenuTarget = {
  ancestorIds: number[];
  id: number;
  label: string;
};

function buildRoleIdentity(): RoleIdentity {
  const suffix = `${Date.now().toString(36)}${Math.random().toString(36).slice(2, 6)}`;
  return {
    code: `rk_mode_${suffix}`,
    name: `r_mode_${suffix}`,
  };
}

async function listRolesByName(api: APIRequestContext, name: string) {
  return expectSuccess<RoleListResult>(
    await api.get(`role?page=1&size=100&name=${encodeURIComponent(name)}`),
  );
}

async function cleanupRolesByName(api: APIRequestContext, ...names: string[]) {
  const ids: number[] = [];
  for (const name of names) {
    const result = await listRolesByName(api, name);
    ids.push(
      ...result.list
        .filter((item) => item.name === name)
        .map((item) => item.id),
    );
  }
  if (ids.length === 0) {
    return;
  }
  await expectSuccess(await api.delete(`role?${buildBatchIdsQuery(ids)}`));
}

async function createRoleWithMenuIds(
  api: APIRequestContext,
  role: RoleIdentity,
  menuIds: number[],
) {
  return expectSuccess<{ id: number }>(
    await api.post("role", {
      data: {
        dataScope: 1,
        key: role.code,
        menuIds,
        name: role.name,
        remark: "Issue #82 role menu mode regression",
        sort: 999,
        status: 1,
      },
    }),
  );
}

async function getMenuTree(api: APIRequestContext) {
  const result = await expectSuccess<{ list: MenuTreeNode[] }>(
    await api.get("menu/treeselect"),
  );
  return result.list;
}

function findButtonMenuTarget(
  menus: MenuTreeNode[],
  ancestors: number[] = [],
): ButtonMenuTarget | undefined {
  for (const menu of menus) {
    if (menu.type === "B") {
      return {
        ancestorIds: ancestors,
        id: menu.id,
        label: menu.label,
      };
    }
    const childTarget = findButtonMenuTarget(menu.children ?? [], [
      ...ancestors,
      menu.id,
    ]);
    if (childTarget) {
      return childTarget;
    }
  }
}

function screenshotPath(description: string) {
  const now = new Date();
  const date = now.toISOString().slice(0, 10).replaceAll("-", "");
  const time = now.toTimeString().slice(0, 8).replaceAll(":", "");
  const dir = path.join(repoRoot, "temp", date);
  mkdirSync(dir, { recursive: true });
  return path.join(dir, `${time}-${description}.png`);
}

test.describe("TC-5 角色权限树模式切换", () => {
  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  // https://github.com/linaproai/linapro/issues/82
  test("TC-5a: 编辑按钮-only 权限角色切换模式后不提交额外父级菜单", async ({
    authenticatedPage: adminPage,
  }) => {
    const role = buildRoleIdentity();
    const button = findButtonMenuTarget(await getMenuTree(adminApi));
    expect(button, "需要至少一个按钮权限节点来复现 Issue #82").toBeTruthy();

    await createRoleWithMenuIds(adminApi, role, [button!.id]);

    const rolePage = new RolePage(adminPage);
    try {
      await rolePage.goto();
      await rolePage.searchRole(role.name);
      await expect(
        adminPage.locator(".vxe-body--row", { hasText: role.name }).first(),
      ).toBeVisible();

      const drawer = await rolePage.openEditDrawer(role.name);
      await rolePage.expectMenuSelectionMode(drawer, "independent");
      await expect(drawer.getByText(button!.label).first()).toBeVisible();
      await adminPage.screenshot({
        fullPage: false,
        path: screenshotPath("role-menu-selection-mode-open"),
      });

      await rolePage.switchMenuSelectionMode(drawer, "linked");
      await rolePage.switchMenuSelectionMode(drawer, "independent");
      await rolePage.switchMenuSelectionMode(drawer, "linked");
      await adminPage.screenshot({
        fullPage: false,
        path: screenshotPath("role-menu-selection-mode-switched"),
      });

      const updateRequest = adminPage.waitForRequest((request) => {
        const url = new URL(request.url());
        return (
          request.method() === "PUT" &&
          /^\/api\/v1\/role\/[^/]+$/.test(url.pathname)
        );
      });
      const updateResponse = adminPage.waitForResponse((response) => {
        const request = response.request();
        const url = new URL(response.url());
        return (
          request.method() === "PUT" &&
          /^\/api\/v1\/role\/[^/]+$/.test(url.pathname)
        );
      });

      await rolePage.submitDrawer(drawer);

      const payload = (await updateRequest).postDataJSON() as {
        menuIds?: number[];
      };
      expect(payload.menuIds).toEqual([button!.id]);
      for (const ancestorId of button!.ancestorIds) {
        expect(payload.menuIds).not.toContain(ancestorId);
      }
      await expectSuccess(await updateResponse);
      await rolePage.waitForDrawerHidden(15000);
      await adminPage.screenshot({
        fullPage: false,
        path: screenshotPath("role-menu-selection-mode-saved"),
      });
    } finally {
      await cleanupRolesByName(adminApi, role.name);
    }
  });
});
