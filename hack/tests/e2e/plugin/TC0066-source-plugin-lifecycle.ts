import { execFileSync } from "node:child_process";
import type { APIRequestContext, APIResponse, Page } from "@playwright/test";

import { request as playwrightRequest } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";
import { config } from "../../fixtures/config";
import { LoginPage } from "../../pages/LoginPage";
import { PluginPage } from "../../pages/PluginPage";

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? "http://127.0.0.1:8080/api/v1/";
const pluginID = "plugin-demo-source";
const pluginMenuName = "源码插件示例";
const pluginSummaryMessage =
  "这是一条来自 plugin-demo-source 接口的简要介绍，用于验证源码插件菜单页可读取插件后端数据。";
const mysqlBin = process.env.E2E_MYSQL_BIN ?? "mysql";
const mysqlUser = process.env.E2E_DB_USER ?? "root";
const mysqlPassword = process.env.E2E_DB_PASSWORD ?? "12345678";
const mysqlDatabase = process.env.E2E_DB_NAME ?? "lina";

type PluginListItem = {
  id: string;
  enabled?: number;
  installed?: number;
  installedAt?: string;
  status?: number;
};

type UserMenuNode = {
  name: string;
  type: string;
  children?: UserMenuNode[];
};

type UserRouteNode = {
  children?: UserRouteNode[];
  meta?: {
    title?: string;
  };
};

function unwrapApiData(payload: any) {
  if (payload && typeof payload === "object" && "data" in payload) {
    return payload.data;
  }
  return payload;
}

function assertOk(response: APIResponse, message: string) {
  expect(response.ok(), `${message}, status=${response.status()}`).toBeTruthy();
}

async function createAdminApiContext(): Promise<APIRequestContext> {
  const loginApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const loginResponse = await loginApi.post("auth/login", {
    data: {
      username: config.adminUser,
      password: config.adminPass,
    },
  });
  assertOk(loginResponse, "管理员登录 API 失败");

  const loginResult = unwrapApiData(await loginResponse.json());
  const accessToken = loginResult?.accessToken;
  expect(accessToken, "未获取到 accessToken").toBeTruthy();
  await loginApi.dispose();

  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

async function syncPlugins(adminApi: APIRequestContext) {
  const response = await adminApi.post("plugins/sync");
  assertOk(response, "同步源码插件失败");
}

async function listPlugins(
  adminApi: APIRequestContext,
): Promise<PluginListItem[]> {
  const response = await adminApi.get("plugins");
  assertOk(response, "查询插件列表失败");
  const payload = unwrapApiData(await response.json());
  return payload?.list ?? [];
}

async function fetchCurrentUserMenus(
  adminApi: APIRequestContext,
): Promise<UserMenuNode[]> {
  const response = await adminApi.get("user/info");
  assertOk(response, "查询当前用户信息失败");
  const payload = unwrapApiData(await response.json());
  return payload?.menus ?? [];
}

async function fetchCurrentUserRoutes(
  adminApi: APIRequestContext,
): Promise<UserRouteNode[]> {
  const response = await adminApi.get("menus/all");
  assertOk(response, "查询当前用户动态路由失败");
  const payload = unwrapApiData(await response.json());
  return payload?.list ?? [];
}

async function fetchPluginSummary(adminApi: APIRequestContext) {
  return await adminApi.get(`plugins/${pluginID}/summary`);
}

async function fetchPluginPing(apiContext: APIRequestContext) {
  return await apiContext.get(`plugins/${pluginID}/ping`);
}

function hasMenuName(list: UserMenuNode[], name: string): boolean {
  return list.some((item) => {
    if (item.name === name) {
      return true;
    }
    return hasMenuName(item.children ?? [], name);
  });
}

function hasButtonMenuNode(list: UserMenuNode[]): boolean {
  return list.some((item) => {
    if (item.type === "B") {
      return true;
    }
    return hasButtonMenuNode(item.children ?? []);
  });
}

function hasRouteTitle(list: UserRouteNode[], title: string): boolean {
  return list.some((item) => {
    if (item?.meta?.title === title) {
      return true;
    }
    return hasRouteTitle(item?.children ?? [], title);
  });
}

async function findPlugin(adminApi: APIRequestContext, id = pluginID) {
  const list = await listPlugins(adminApi);
  return list.find((item) => item.id === id) ?? null;
}

async function updatePluginStatus(
  adminApi: APIRequestContext,
  id: string,
  enabled: boolean,
) {
  const url = enabled ? `plugins/${id}/enable` : `plugins/${id}/disable`;
  const response = await adminApi.put(url);
  assertOk(response, `更新插件状态失败: enabled=${enabled}`);
}

function resetPluginRegistryRow(id: string) {
  execFileSync(
    mysqlBin,
    [
      `-u${mysqlUser}`,
      `-p${mysqlPassword}`,
      mysqlDatabase,
      "-e",
      `DELETE FROM sys_plugin WHERE plugin_id = '${id.replaceAll("'", "''")}';`,
    ],
    {
      stdio: "ignore",
    },
  );
}

async function loginAsAdmin(page: Page) {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);
}

test.describe("TC-66 源码插件生命周期", () => {
  let adminApi: APIRequestContext | null = null;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    if (adminApi) {
      await adminApi.dispose();
    }
  });

  test("TC-66a: 同步 source 插件后自动处于已集成且默认启用态", async ({
    page,
  }) => {
    resetPluginRegistryRow(pluginID);
    await syncPlugins(adminApi!);

    const pluginAfterSync = await findPlugin(adminApi!);
    expect(pluginAfterSync, `同步后应发现 ${pluginID}`).toBeTruthy();
    expect(pluginAfterSync?.installed, "源码插件同步后应直接处于已集成态").toBe(
      1,
    );
    expect(
      "runtime" in ((pluginAfterSync ?? {}) as Record<string, unknown>),
      "插件列表接口不应再返回重复的 runtime 字段",
    ).toBeFalsy();
    expect(
      pluginAfterSync?.installedAt,
      "源码插件同步后应记录接入时间",
    ).toBeTruthy();
    expect(
      pluginAfterSync?.enabled ?? pluginAfterSync?.status,
      "源码插件首次同步后应默认启用",
    ).toBe(1);

    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await expect(loginPage.pluginLoginSlot).toHaveCount(0);
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await expect(pluginPage.pluginRow(pluginID)).toBeVisible();
    await expect(pluginPage.pluginEnabledSwitch(pluginID)).toHaveAttribute(
      "aria-checked",
      "true",
    );
    await pluginPage.expectTableColumnVisible("插件类型");
    await pluginPage.expectTableColumnVisible("安装时间");
    await pluginPage.expectTableColumnHidden("交付方式");
    await pluginPage.expectTableColumnHidden("接入态");
    await pluginPage.expectTableColumnHidden("入口");
    await pluginPage.expectTableColumnHidden("生命周期");
    await pluginPage.expectTableColumnHidden("治理摘要");
    await pluginPage.expectTableColumnBetween("描述", "版本", "状态");
    await pluginPage.expectDescriptionUsesNativeTooltip(pluginID);
    await expect(pluginPage.pluginInstallButton(pluginID)).toHaveCount(0);
    await pluginPage.expectSourcePluginDisabledUninstall(pluginID);
    await pluginPage.expectCrudSlotsHidden();
    await pluginPage.expectHeaderSlotsHidden();
  });

  test("TC-66b: 启用后仅左侧菜单页可正常展示，且不渲染额外 slots", async ({
    page,
  }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, true);

    const pluginAfterEnable = await findPlugin(adminApi!);
    expect(pluginAfterEnable?.enabled ?? pluginAfterEnable?.status).toBe(1);

    const pluginPage = new PluginPage(page);
    await loginAsAdmin(page);

    await pluginPage.gotoWorkspace();
    await pluginPage.expectWorkspaceSlotHidden();
    await pluginPage.gotoManage();
    await pluginPage.expectHeaderSlotsHidden();
    await pluginPage.expectCrudSlotsHidden();
    await pluginPage.openSidebarExampleFromMenu();
  });

  test("TC-66c: 启用后可验证插件路由与鉴权访问控制", async ({ page }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, true);

    const anonymousApi = await playwrightRequest.newContext({
      baseURL: apiBaseURL,
    });
    const pingResponse = await fetchPluginPing(anonymousApi);
    assertOk(pingResponse, "查询插件公开 ping 路由失败");
    const pingPayload = unwrapApiData(await pingResponse.json());
    expect(pingPayload?.message, "插件公开路由应允许匿名访问").toBe("pong");

    const anonymousSummaryResponse = await fetchPluginSummary(anonymousApi);
    expect(
      anonymousSummaryResponse.status(),
      "插件受保护摘要路由在未鉴权时应返回 401",
    ).toBe(401);
    await anonymousApi.dispose();

    const summaryResponse = await fetchPluginSummary(adminApi!);
    assertOk(summaryResponse, "查询插件摘要路由失败");
    const summaryPayload = unwrapApiData(await summaryResponse.json());
    expect(
      summaryPayload?.message,
      "插件摘要应仅返回页面实际使用的简介文案",
    ).toBe(pluginSummaryMessage);

    await loginAsAdmin(page);
  });

  test("TC-66d: 禁用后不渲染源码样例额外内容且隐藏菜单", async ({ page }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, false);

    const summaryResponse = await fetchPluginSummary(adminApi!);
    expect(summaryResponse.status(), "插件禁用后插件自有路由应返回 404").toBe(
      404,
    );

    const pluginAfterDisable = await findPlugin(adminApi!);
    expect(pluginAfterDisable?.enabled ?? pluginAfterDisable?.status ?? 0).toBe(
      0,
    );

    const pluginPage = new PluginPage(page);
    await loginAsAdmin(page);

    await pluginPage.gotoWorkspace();
    await pluginPage.expectWorkspaceSlotHidden();
    await pluginPage.gotoManage();
    await pluginPage.expectHeaderSlotsHidden();
    await pluginPage.expectCrudSlotsHidden();
    await pluginPage.expectSidebarMenuHidden(pluginMenuName);
  });

  test("TC-66e: 禁用后源码插件仍保留已集成态且无需重新安装", async ({
    page,
  }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, false);

    const pluginAfterDisable = await findPlugin(adminApi!);
    expect(
      pluginAfterDisable,
      "禁用后仍应可在清单中发现 source 插件",
    ).toBeTruthy();
    expect(
      pluginAfterDisable?.installed ?? 0,
      "源码插件禁用后仍应保持已集成态",
    ).toBe(1);

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await expect(pluginPage.pluginRow(pluginID)).toBeVisible();
    await expect(pluginPage.pluginEnabledSwitch(pluginID)).toHaveAttribute(
      "aria-checked",
      "false",
    );
    await expect(pluginPage.pluginInstallButton(pluginID)).toHaveCount(0);
    await pluginPage.expectSourcePluginDisabledUninstall(pluginID);
  });

  test("TC-66f: 登录态在线启用后立即刷新左侧菜单且不渲染额外 slots", async ({
    page,
  }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, false);

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await pluginPage.expectHeaderSlotsHidden();
    await pluginPage.expectCrudSlotsHidden();
    await pluginPage.expectSidebarMenuHidden(pluginMenuName);

    await pluginPage.setPluginEnabled(pluginID, true);

    await pluginPage.expectSidebarMenuVisible(pluginMenuName);
    await pluginPage.expectHeaderSlotsHidden();
    await pluginPage.expectCrudSlotsHidden();
    await pluginPage.gotoWorkspace();
    await pluginPage.expectWorkspaceSlotHidden();
  });

  test("TC-66g: 登录态在线禁用后立即隐藏左侧菜单且保持无额外 slots", async ({
    page,
  }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, true);

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoWorkspace();
    await pluginPage.expectWorkspaceSlotHidden();
    await pluginPage.gotoManage();
    await pluginPage.expectHeaderSlotsHidden();
    await pluginPage.expectCrudSlotsHidden();
    await pluginPage.expectSidebarMenuVisible(pluginMenuName);

    await pluginPage.setPluginEnabled(pluginID, false);

    await pluginPage.expectHeaderSlotsHidden();
    await pluginPage.expectCrudSlotsHidden();
    await pluginPage.expectSidebarMenuHidden(pluginMenuName);
    await pluginPage.gotoWorkspace();
    await pluginPage.expectWorkspaceSlotHidden();
  });

  test("TC-66h: 当前会话重新获得焦点后自动同步外部插件状态变更", async ({
    page,
  }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, false);

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await pluginPage.expectHeaderSlotsHidden();
    await pluginPage.expectCrudSlotsHidden();
    await pluginPage.expectSidebarMenuHidden(pluginMenuName);

    await updatePluginStatus(adminApi!, pluginID, true);
    await page.evaluate(() => {
      window.dispatchEvent(new Event("focus"));
      document.dispatchEvent(new Event("visibilitychange"));
    });

    await pluginPage.expectSidebarMenuVisible(pluginMenuName);
    await pluginPage.expectHeaderSlotsHidden();
    await pluginPage.expectCrudSlotsHidden();
    await pluginPage.gotoWorkspace();
    await pluginPage.expectWorkspaceSlotHidden();
  });

  test("TC-66i: 按钮权限不会被返回为左侧导航菜单或动态路由", async ({
    page,
  }) => {
    const currentUserMenus = await fetchCurrentUserMenus(adminApi!);
    expect(
      hasButtonMenuNode(currentUserMenus),
      "user/info 不应再返回按钮类型菜单",
    ).toBeFalsy();
    expect(
      hasMenuName(currentUserMenus, "插件查询"),
      "user/info 不应包含插件查询按钮菜单",
    ).toBeFalsy();

    const currentUserRoutes = await fetchCurrentUserRoutes(adminApi!);
    expect(
      hasRouteTitle(currentUserRoutes, "插件查询"),
      "menus/all 不应包含插件查询按钮路由",
    ).toBeFalsy();
    expect(
      hasRouteTitle(currentUserRoutes, "用户查询"),
      "menus/all 不应包含用户查询按钮路由",
    ).toBeFalsy();

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await expect(page).toHaveURL(/\/system\/plugin$/);
    await expect(page).toHaveTitle("插件管理 - Lina");
    await pluginPage.expectSidebarMenuHidden("插件查询");
    await pluginPage.expectSidebarMenuHidden("用户查询");
  });

  test("TC-66j: 当前会话重新获得焦点但插件状态未变化时不重复刷新菜单", async ({
    page,
  }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, true);

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await pluginPage.expectSidebarMenuVisible(pluginMenuName);
    await page.waitForTimeout(1200);

    const menuResponses: string[] = [];
    page.on("response", (response) => {
      if (
        response.request().method() === "GET" &&
        response.url().includes("/api/v1/menus/all")
      ) {
        menuResponses.push(response.url());
      }
    });

    await page.evaluate(() => {
      window.dispatchEvent(new Event("focus"));
      document.dispatchEvent(new Event("visibilitychange"));
    });

    await page.waitForTimeout(1200);
    await pluginPage.expectSidebarMenuVisible(pluginMenuName);
    expect(
      menuResponses,
      "插件状态未变化时，焦点恢复不应重复拉取菜单",
    ).toHaveLength(0);
  });

  test("TC-66k: 登录后打开插件管理页时公共插件状态接口不重复重查", async ({
    page,
  }) => {
    await syncPlugins(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, true);

    const runtimeStateResponses: string[] = [];
    page.on("response", (response) => {
      if (
        response.request().method() === "GET" &&
        response.url().includes("/api/v1/plugins/dynamic")
      ) {
        runtimeStateResponses.push(response.url());
      }
    });

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await pluginPage.expectSidebarMenuVisible(pluginMenuName);
    await page.waitForTimeout(1500);

    expect(
      runtimeStateResponses.length,
      "登录并打开插件管理页时，公共插件状态接口不应重复触发多次",
    ).toBeLessThanOrEqual(2);
  });
});
