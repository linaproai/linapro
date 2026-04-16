import type { APIRequestContext } from "@playwright/test";
import { test, expect, request as playwrightRequest } from "@playwright/test";
import { LoginPage } from "../../pages/LoginPage";
import { MainLayout } from "../../pages/MainLayout";
import { config } from "../../fixtures/config";

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? "http://localhost:8080/api/v1/";

type MenuNode = {
  id: number;
  name: string;
  children?: MenuNode[];
};

type RouteNode = {
  children?: RouteNode[];
  meta?: {
    hideInMenu?: boolean;
    title?: string;
  };
};

function flattenMenus(list: MenuNode[]): MenuNode[] {
  return list.flatMap((item) => [item, ...flattenMenus(item.children ?? [])]);
}

function findMenuNodeByName(
  list: MenuNode[],
  menuName: string,
): MenuNode | null {
  for (const item of list) {
    if (item.name === menuName) {
      return item;
    }
    const match = findMenuNodeByName(item.children ?? [], menuName);
    if (match) {
      return match;
    }
  }
  return null;
}

function findRouteNodeByTitle(
  list: RouteNode[],
  title: string,
): RouteNode | null {
  for (const item of list) {
    if (item.meta?.title === title) {
      return item;
    }
    const match = findRouteNodeByTitle(item.children ?? [], title);
    if (match) {
      return match;
    }
  }
  return null;
}

function getVisibleChildTitles(node: RouteNode | null): string[] {
  return (node?.children ?? [])
    .filter((item) => !item.meta?.hideInMenu)
    .map((item) => item.meta?.title ?? "")
    .filter(Boolean);
}

function getVisibleRootTitles(list: RouteNode[]): string[] {
  return list
    .filter((item) => !item.meta?.hideInMenu)
    .map((item) => item.meta?.title ?? "")
    .filter(Boolean);
}

async function createAdminApiContext(): Promise<APIRequestContext> {
  const loginApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const loginResponse = await loginApi.post("auth/login", {
    data: {
      username: config.adminUser,
      password: config.adminPass,
    },
  });
  expect(loginResponse.ok()).toBeTruthy();

  const loginResult = await loginResponse.json();
  const accessToken = loginResult.data?.accessToken;
  expect(accessToken).toBeTruthy();
  await loginApi.dispose();

  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

async function getMenuIdsByNames(api: APIRequestContext, menuNames: string[]) {
  const response = await api.get("menu");
  expect(response.ok()).toBeTruthy();

  const result = await response.json();
  const flatMenus = flattenMenus(result.data?.list ?? []);

  return menuNames.map((menuName) => {
    const menu = flatMenus.find((item) => item.name === menuName);
    expect(menu, `missing menu: ${menuName}`).toBeTruthy();
    return menu!.id;
  });
}

async function getCurrentUserRouteTree(
  api: APIRequestContext,
): Promise<RouteNode[]> {
  const response = await api.get("menus/all");
  expect(response.ok()).toBeTruthy();

  const result = await response.json();
  return result.data?.list ?? [];
}

test.describe("TC0063 登录后菜单显示", () => {
  const uniqueSuffix = Date.now().toString();
  const testRoleName = `e2e_menu_role_${Date.now()}`;
  const testRoleCode = `emr_${uniqueSuffix}`;
  const testUserUsername = `e2e_menu_user_${Date.now()}`;
  const testUserPassword = "test123456";
  const noRoleUsername = `e2e_no_role_${Date.now()}`;
  let adminApi: APIRequestContext | null = null;
  let testRoleId = 0;
  let testUserId = 0;
  let noRoleUserId = 0;
  let roleMenuIds: number[] = [];
  let expandedRoleMenuIds: number[] = [];

  test.beforeAll(async () => {
    const api = await createAdminApiContext();
    adminApi = api;
    roleMenuIds = await getMenuIdsByNames(api, ["系统管理", "用户管理"]);
    expandedRoleMenuIds = await getMenuIdsByNames(api, [
      "系统管理",
      "用户管理",
      "角色管理",
    ]);

    const createRoleResponse = await api.post("role", {
      data: {
        name: testRoleName,
        key: testRoleCode,
        sort: 900,
        dataScope: 1,
        status: 1,
        remark: "E2E测试角色-用于菜单显示测试",
        menuIds: roleMenuIds,
      },
    });
    expect(createRoleResponse.ok()).toBeTruthy();
    const createRoleResult = await createRoleResponse.json();
    expect(createRoleResult.code, createRoleResult.message).toBe(0);
    testRoleId = createRoleResult.data?.id ?? 0;
    expect(testRoleId).toBeGreaterThan(0);

    const createUserResponse = await api.post("user", {
      data: {
        username: testUserUsername,
        password: testUserPassword,
        nickname: "E2E菜单测试用户",
        roleIds: [testRoleId],
      },
    });
    expect(createUserResponse.ok()).toBeTruthy();
    const createUserResult = await createUserResponse.json();
    expect(createUserResult.code, createUserResult.message).toBe(0);
    testUserId = createUserResult.data?.id ?? 0;
    expect(testUserId).toBeGreaterThan(0);

    const createNoRoleUserResponse = await api.post("user", {
      data: {
        username: noRoleUsername,
        password: testUserPassword,
        nickname: "E2E无角色用户",
      },
    });
    expect(createNoRoleUserResponse.ok()).toBeTruthy();
    const createNoRoleUserResult = await createNoRoleUserResponse.json();
    expect(createNoRoleUserResult.code, createNoRoleUserResult.message).toBe(0);
    noRoleUserId = createNoRoleUserResult.data?.id ?? 0;
    expect(noRoleUserId).toBeGreaterThan(0);
  });

  test("TC0063a: 超级管理员登录后显示完整菜单", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);
    await page.goto("/system/menu");
    await page.waitForLoadState("networkidle");

    // Wait for sidebar/menu to render
    await page.waitForTimeout(2000);
    const sidebarMenu = page.getByRole("menu").first();

    // Admin should see system management menu
    const systemMenu = sidebarMenu.getByText("系统管理").first();
    await expect(systemMenu).toBeVisible({ timeout: 5000 });

    // Admin should see menu management
    const menuManagement = sidebarMenu.getByText("菜单管理").first();
    await expect(menuManagement).toBeVisible({ timeout: 5000 });

    // Admin should see role management
    const roleManagement = sidebarMenu.getByText("角色管理").first();
    await expect(roleManagement).toBeVisible({ timeout: 5000 });

    const currentUserRoutes = await getCurrentUserRouteTree(adminApi!);
    const visibleRootTitles = getVisibleRootTitles(currentUserRoutes);
    expect(visibleRootTitles.indexOf("仪表盘")).toBeGreaterThanOrEqual(0);
    expect(visibleRootTitles.indexOf("系统管理")).toBeGreaterThanOrEqual(0);
    expect(visibleRootTitles.indexOf("仪表盘")).toBeLessThan(
      visibleRootTitles.indexOf("系统管理"),
    );

    const systemRoute = findRouteNodeByTitle(currentUserRoutes, "系统管理");
    const visibleSystemChildren = getVisibleChildTitles(systemRoute);
    expect(visibleSystemChildren).toEqual([
      "用户管理",
      "角色管理",
      "菜单管理",
      "部门管理",
      "岗位管理",
      "字典管理",
      "通知公告",
      "参数设置",
      "文件管理",
      "插件管理",
    ]);

    const currentMenusResponse = await adminApi!.get("menu");
    expect(currentMenusResponse.ok()).toBeTruthy();
    const currentMenusResult = await currentMenusResponse.json();
    const systemMenuNode = findMenuNodeByName(
      currentMenusResult.data?.list ?? [],
      "系统管理",
    );
    expect(systemMenuNode, "系统管理菜单应存在").toBeTruthy();
    const systemChildNames = (systemMenuNode?.children ?? []).map(
      (item) => item.name,
    );
    expect(systemChildNames).toContain("插件管理");

    const rootPluginMenu = (currentMenusResult.data?.list ?? []).find(
      (item: MenuNode) => item.name === "插件管理",
    );
    expect(rootPluginMenu, "插件管理不应再作为顶级菜单存在").toBeFalsy();
  });

  test("TC0063b: 普通用户登录后仅显示授权菜单", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(testUserUsername, testUserPassword);
    await page.goto("/system/user");
    await page.waitForLoadState("networkidle");

    // Wait for sidebar/menu to render
    await page.waitForTimeout(2000);
    const sidebarMenu = page.getByRole("menu").first();

    const systemMenu = sidebarMenu.getByText("系统管理").first();
    await expect(systemMenu).toBeVisible({ timeout: 5000 });

    const userManagement = sidebarMenu.getByText("用户管理").first();
    await expect(userManagement).toBeVisible({ timeout: 5000 });

    // Should NOT see system management (unless role has that menu)
    const menuManagement = sidebarMenu.getByText("菜单管理").first();
    const isMenuMgmtVisible = await menuManagement
      .isVisible({ timeout: 2000 })
      .catch(() => false);
    expect(isMenuMgmtVisible).toBeFalsy();
  });

  test("TC0063c: 无角色用户登录后无菜单", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(noRoleUsername, testUserPassword);

    // Wait for page to load
    await page.waitForTimeout(2000);
    await expect(page.getByText("未找到页面")).toBeVisible({ timeout: 5000 });
  });

  test("TC0063d: 不同用户菜单权限差异", async ({ page }) => {
    const loginPage = new LoginPage(page);
    const systemMenuEntries = [
      "用户管理",
      "角色管理",
      "菜单管理",
      "部门管理",
      "岗位管理",
      "字典管理",
      "通知公告",
      "参数设置",
      "文件管理",
    ];

    // First login as admin and check available menus
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);
    await page.goto("/system/menu");
    await page.waitForLoadState("networkidle");

    await page.waitForTimeout(2000);
    const adminSidebar = page.getByRole("menu").first();

    const adminMenuCount = (
      await Promise.all(
        systemMenuEntries.map((menuName) =>
          adminSidebar
            .getByText(menuName, { exact: true })
            .first()
            .isVisible({ timeout: 1000 })
            .catch(() => false),
        ),
      )
    ).filter(Boolean).length;

    // Logout
    const mainLayout = new MainLayout(page);
    await mainLayout.logout();

    // Login as test user
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(testUserUsername, testUserPassword);
    await page.goto("/system/user");
    await page.waitForLoadState("networkidle");

    await page.waitForTimeout(2000);
    const testSidebar = page.getByRole("menu").first();

    const testMenuCount = (
      await Promise.all(
        systemMenuEntries.map((menuName) =>
          testSidebar
            .getByText(menuName, { exact: true })
            .first()
            .isVisible({ timeout: 1000 })
            .catch(() => false),
        ),
      )
    ).filter(Boolean).length;

    // Admin should have more menus than test user
    expect(adminMenuCount).toBeGreaterThan(testMenuCount);
  });

  test("TC0063e: 菜单变更后需重新登录生效", async ({ browser }) => {
    const updateRoleResponse = await adminApi!.put(`role/${testRoleId}`, {
      data: {
        id: testRoleId,
        name: testRoleName,
        key: testRoleCode,
        sort: 900,
        dataScope: 1,
        status: 1,
        remark: "E2E测试角色-用于菜单显示测试",
        menuIds: expandedRoleMenuIds,
      },
    });
    expect(updateRoleResponse.ok()).toBeTruthy();
    const updateRoleResult = await updateRoleResponse.json();
    expect(updateRoleResult.code, updateRoleResult.message).toBe(0);

    // Now login as test user in a new context
    const testContext = await browser.newContext();
    const testPage = await testContext.newPage();

    const testLogin = new LoginPage(testPage);
    await testLogin.goto();
    await testLogin.loginAndWaitForRedirect(testUserUsername, testUserPassword);
    await testPage.goto("/system/user");
    await testPage.waitForLoadState("networkidle");

    await testPage.waitForTimeout(2000);
    const sidebarMenu = testPage.getByRole("menu").first();

    const roleManagement = sidebarMenu.getByText("角色管理").first();
    await expect(roleManagement).toBeVisible({ timeout: 5000 });

    await testContext.close();
  });

  test("TC0063f: 刷新页面时菜单仅装载一次", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);
    await page.goto("/system/plugin");
    await page.waitForLoadState("networkidle");

    const menuResponses: string[] = [];
    page.on("response", (response) => {
      if (
        response.request().method() === "GET" &&
        response.url().includes("/api/v1/menus/all")
      ) {
        menuResponses.push(response.url());
      }
    });

    await page.reload();
    await page.waitForLoadState("networkidle");
    await page.waitForTimeout(500);

    expect(menuResponses, "刷新页面时不应重复拉取菜单").toHaveLength(1);
  });

  test.afterAll(async () => {
    if (testUserId > 0) {
      await adminApi?.delete(`user/${testUserId}`);
    }
    if (noRoleUserId > 0) {
      await adminApi?.delete(`user/${noRoleUserId}`);
    }
    if (testRoleId > 0) {
      await adminApi?.delete(`role/${testRoleId}`);
    }
    await adminApi?.dispose();
  });
});
