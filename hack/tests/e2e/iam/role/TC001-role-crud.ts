import type { APIRequestContext } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import { RolePage } from "../../../pages/RolePage";
import { createAdminApiContext, expectSuccess } from "../../../support/api/job";

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

function buildRoleIdentity(scope: string): RoleIdentity {
  const suffix = `${Date.now().toString(36)}${Math.random().toString(36).slice(2, 6)}`;
  return {
    code: `rk_${scope}_${suffix}`,
    name: `r_${scope}_${suffix}`,
  };
}

function buildRepeatedIdsQuery(ids: number[]) {
  const params = new URLSearchParams();
  for (const id of ids) {
    params.append("ids", String(id));
  }
  return params.toString();
}

async function listRolesByName(api: APIRequestContext, name: string) {
  return expectSuccess<RoleListResult>(
    await api.get(`role?page=1&size=100&name=${encodeURIComponent(name)}`),
  );
}

async function createRoleFixture(
  api: APIRequestContext,
  role: RoleIdentity,
) {
  return expectSuccess<{ id: number }>(
    await api.post("role", {
      data: {
        dataScope: 1,
        key: role.code,
        name: role.name,
        remark: "E2E测试角色",
        sort: 999,
        status: 1,
      },
    }),
  );
}

async function cleanupRolesByName(
  api: APIRequestContext,
  ...names: string[]
) {
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
  await expectSuccess(await api.delete(`role?${buildRepeatedIdsQuery(ids)}`));
}

async function expectPageHeightStable(page: any, pageName: string) {
  const samples = await page.evaluate(async () => {
    const values: number[] = [];
    for (let index = 0; index < 4; index += 1) {
      values.push(document.documentElement.scrollHeight);
      if (index < 3) {
        await new Promise<void>((resolve) => {
          requestAnimationFrame(() => requestAnimationFrame(() => resolve()));
        });
      }
    }
    return values;
  });

  expect(
    Math.max(...samples) - Math.min(...samples),
    `${pageName}高度未稳定，采样结果: ${samples.join(", ")}`,
  ).toBeLessThanOrEqual(16);
}

/**
 * TC001 角色管理 E2E 测试
 *
 * 测试覆盖：
 * - 角色列表页面加载
 * - 角色创建功能
 * - 角色编辑功能
 * - 角色删除功能
 * - 角色状态切换
 * - 角色菜单分配
 */
test.describe("TC001 角色管理 CRUD", () => {
  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  test("TC001a: 角色列表页面正常加载", async ({ authenticatedPage: adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    // 验证表格可见
    const table = adminPage.locator(".vxe-table");
    await expect(table).toBeVisible({ timeout: 10000 });

    // 验证工具栏按钮可见
    await expect(
      adminPage.getByRole("button", { name: /新\s*增/ }).first(),
    ).toBeVisible({ timeout: 5000 });
    await expectPageHeightStable(adminPage, "角色管理页");
  });

  test("TC001b: 创建角色对话框打开", async ({ authenticatedPage: adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    // 点击新增按钮
    const drawer = await rolePage.openCreateDrawer();

    // 验证表单字段存在
    await expect(
      drawer.locator('input[placeholder="请输入角色名称"]'),
    ).toBeVisible({ timeout: 5000 });
    await expect(
      drawer.locator('input[placeholder="如: admin, user等"]'),
    ).toBeVisible({ timeout: 5000 });

    // Drawer will be closed by the test cleanup
  });

  test("TC001c: 创建新角色", async ({ authenticatedPage: adminPage }) => {
    const role = buildRoleIdentity("create");
    const rolePage = new RolePage(adminPage);
    try {
      await rolePage.goto();

      await rolePage.createRole({
        code: role.code,
        name: role.name,
        remark: "E2E测试角色",
        sort: 999,
      });

      // 等待抽屉关闭表示提交完成
      await rolePage.waitForDrawerHidden(15000);

      // 验证角色已创建
      await rolePage.searchRole(role.name);
      const hasRole = await rolePage.hasRole(role.name);
      expect(hasRole).toBeTruthy();
    } finally {
      await cleanupRolesByName(adminApi, role.name);
    }
  });

  test("TC001d: 编辑角色", async ({ authenticatedPage: adminPage }) => {
    const role = buildRoleIdentity("edit");
    const newName = `${role.name}_edited`;
    await createRoleFixture(adminApi, role);

    const rolePage = new RolePage(adminPage);
    try {
      await rolePage.goto();
      await rolePage.searchRole(role.name);
      await expect(
        adminPage.locator(".vxe-body--row", { hasText: role.name }).first(),
      ).toBeVisible();

      // 编辑角色
      await rolePage.editRole(role.name, newName);

      // 等待抽屉关闭
      await rolePage.waitForDrawerHidden();

      // 验证编辑成功
      await rolePage.goto();
      await rolePage.searchRole(newName);
      const hasRole = await rolePage.hasRole(newName);
      expect(hasRole).toBeTruthy();
    } finally {
      await cleanupRolesByName(adminApi, role.name, newName);
    }
  });

  test("TC001e: 角色状态切换", async ({ authenticatedPage: adminPage }) => {
    const role = buildRoleIdentity("status");
    await createRoleFixture(adminApi, role);

    const rolePage = new RolePage(adminPage);
    try {
      await rolePage.goto();
      await rolePage.searchRole(role.name);
      await expect(
        adminPage.locator(".vxe-body--row", { hasText: role.name }).first(),
      ).toBeVisible();

      // 获取当前状态
      const switchEl = adminPage.locator(".vxe-body--row .ant-switch").first();
      const initialState = await switchEl.getAttribute("aria-checked");

      // 切换状态
      await rolePage.toggleStatus(role.name);

      // 验证状态已改变
      const newState = await switchEl.getAttribute("aria-checked");
      expect(newState).not.toBe(initialState);

      // 恢复原状态
      await rolePage.toggleStatus(role.name);
    } finally {
      await cleanupRolesByName(adminApi, role.name);
    }
  });

  test("TC001f: 角色菜单分配", async ({ authenticatedPage: adminPage }) => {
    const role = buildRoleIdentity("menus");
    await createRoleFixture(adminApi, role);

    const rolePage = new RolePage(adminPage);
    try {
      await rolePage.goto();
      await rolePage.searchRole(role.name);
      await expect(
        adminPage.locator(".vxe-body--row", { hasText: role.name }).first(),
      ).toBeVisible();

      // 编辑角色并分配菜单
      await rolePage.assignMenusToRole(role.name, ["权限管理"]);

      // 等待抽屉关闭
      await rolePage.waitForDrawerHidden();
    } finally {
      await cleanupRolesByName(adminApi, role.name);
    }
  });

  test("TC001g: 删除角色", async ({ authenticatedPage: adminPage }) => {
    const role = buildRoleIdentity("delete");
    await createRoleFixture(adminApi, role);

    const rolePage = new RolePage(adminPage);
    try {
      await rolePage.goto();
      await rolePage.searchRole(role.name);
      await expect(
        adminPage.locator(".vxe-body--row", { hasText: role.name }).first(),
      ).toBeVisible();

      // 删除角色
      await rolePage.deleteRole(role.name);

      // 验证角色已删除
      await rolePage.goto();
      await rolePage.searchRole(role.name);
      const hasRole = await rolePage.hasRole(role.name);
      expect(hasRole).toBeFalsy();
    } finally {
      await cleanupRolesByName(adminApi, role.name);
    }
  });

  test("TC001h: 角色搜索功能", async ({ authenticatedPage: adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    // 搜索管理员角色
    await rolePage.searchRole("管理员");

    // 验证搜索结果
    const hasAdmin = await rolePage.hasRole("管理员");
    expect(hasAdmin).toBeTruthy();

    // 重置搜索
    await rolePage.resetSearch();

    // 验证重置后能看到更多角色
    const rowsBefore = await adminPage.locator(".vxe-body--row").count();
    await rolePage.searchRole("管理员");
    const rowsAfter = await adminPage.locator(".vxe-body--row").count();
    expect(rowsBefore).toBeGreaterThanOrEqual(rowsAfter);
  });

  test("TC001i: 超级管理员角色不可编辑删除", async ({ authenticatedPage: adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    // 搜索超级管理员角色 (id=1)
    await rolePage.searchRole("超级管理员");

    // 验证超级管理员存在
    const hasAdmin = await rolePage.hasRole("超级管理员");
    expect(hasAdmin).toBeTruthy();

    // 验证状态开关被禁用
    const isDisabled = await rolePage.isStatusSwitchDisabled("超级管理员");
    expect(isDisabled).toBeTruthy();

    // 验证复选框被禁用
    const isCheckboxDisabled = await rolePage.isCheckboxDisabled("超级管理员");
    expect(isCheckboxDisabled).toBeTruthy();
  });
});
