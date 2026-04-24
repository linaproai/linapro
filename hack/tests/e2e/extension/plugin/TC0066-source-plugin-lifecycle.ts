import { execFileSync } from "node:child_process";
import {
  existsSync,
  mkdirSync,
  readFileSync,
  readdirSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";
import type { APIRequestContext, APIResponse, Page } from "@playwright/test";

import { request as playwrightRequest } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import { config } from "../../../fixtures/config";
import { LoginPage } from "../../../pages/LoginPage";
import { PluginPage } from "../../../pages/PluginPage";

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? "http://127.0.0.1:8080/api/v1/";
const demoControlPluginID = "demo-control";
const pluginID = "plugin-demo-source";
const pluginMenuName = "源码插件示例";
const pluginSummaryMessage =
  "这是一条来自 plugin-demo-source 接口的简要介绍，用于验证源码插件菜单页可读取插件后端数据。";
const mysqlBin = process.env.E2E_MYSQL_BIN ?? "mysql";
const mysqlUser = process.env.E2E_DB_USER ?? "root";
const mysqlPassword = process.env.E2E_DB_PASSWORD ?? "12345678";
const mysqlDatabase = process.env.E2E_DB_NAME ?? "lina";
const pluginRecordSeedTitle = "源码插件 SQL 示例记录";
const pluginRecordTableName = "plugin_demo_source_record";
const repoRoot = path.resolve(process.cwd(), "../..");
const pluginDemoSourceStorageRoot = path.resolve(
  repoRoot,
  "apps/lina-core/temp/upload/plugin-demo-source",
);
const pluginDemoSourceFixturePath = path.resolve(
  repoRoot,
  "temp/TC0066-plugin-demo-source-note.txt",
);
const pluginDemoSourceFixtureContent = "TC0066 plugin demo source attachment";
const pluginDemoSourceDownloadPath = path.resolve(
  repoRoot,
  "temp/TC0066-plugin-demo-source-download.txt",
);

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

async function installPlugin(adminApi: APIRequestContext, id = pluginID) {
  const response = await adminApi.post(`plugins/${id}/install`);
  assertOk(response, `安装插件失败: ${id}`);
}

async function installAndEnablePlugin(
  adminApi: APIRequestContext,
  id = pluginID,
) {
  await installPlugin(adminApi, id);
  await updatePluginStatus(adminApi, id, true);
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

async function listDemoRecords(adminApi: APIRequestContext) {
  const response = await adminApi.get(`plugins/${pluginID}/records`);
  assertOk(response, "查询源码插件示例记录失败");
  const payload = unwrapApiData(await response.json());
  return payload?.list ?? [];
}

async function createDemoRecord(
  adminApi: APIRequestContext,
  title: string,
  content: string,
  filePath?: string,
) {
  const multipart: Record<string, any> = {
    title,
    content,
  };
  if (filePath) {
    multipart.file = {
      buffer: readFileSync(filePath),
      mimeType: "text/plain",
      name: path.basename(filePath),
    };
  }
  const response = await adminApi.post(`plugins/${pluginID}/records`, {
    multipart,
  });
  assertOk(response, `创建源码插件示例记录失败: ${title}`);
  const payload = unwrapApiData(await response.json());
  expect(payload?.id, "创建源码插件示例记录成功后应返回记录ID").toBeTruthy();
  return payload.id as number;
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
  const escapedID = id.replaceAll("'", "''");
  const menuKeyPattern = `plugin:${escapedID}:%`;

  execFileSync(
    mysqlBin,
    [
      `-u${mysqlUser}`,
      `-p${mysqlPassword}`,
      mysqlDatabase,
      "-e",
      [
        `DELETE FROM sys_role_menu WHERE menu_id IN (SELECT id FROM (SELECT id FROM sys_menu WHERE menu_key LIKE '${menuKeyPattern}') AS plugin_menu_ids);`,
        `DELETE FROM sys_menu WHERE menu_key LIKE '${menuKeyPattern}';`,
        `DELETE FROM sys_plugin_state WHERE plugin_id = '${escapedID}';`,
        `DELETE FROM sys_plugin_node_state WHERE plugin_id = '${escapedID}';`,
        `DELETE FROM sys_plugin_resource_ref WHERE plugin_id = '${escapedID}';`,
        `DELETE FROM sys_plugin_migration WHERE plugin_id = '${escapedID}';`,
        `DELETE FROM sys_plugin_release WHERE plugin_id = '${escapedID}';`,
        `DELETE FROM sys_plugin WHERE plugin_id = '${escapedID}';`,
      ].join(" "),
    ],
    {
      stdio: "ignore",
    },
  );
}

function resetPluginDemoSourceData() {
  if (pluginDemoSourceTableExists()) {
    execFileSync(
      mysqlBin,
      [
        `-u${mysqlUser}`,
        `-p${mysqlPassword}`,
        mysqlDatabase,
        "-e",
        `DROP TABLE IF EXISTS \`${pluginRecordTableName}\`;`,
      ],
      {
        stdio: "ignore",
      },
    );
  }
  rmSync(pluginDemoSourceStorageRoot, { force: true, recursive: true });
  rmSync(pluginDemoSourceDownloadPath, { force: true });
}

function queryMysqlLines(sql: string) {
  const output = execFileSync(
    mysqlBin,
    [
      `-u${mysqlUser}`,
      `-p${mysqlPassword}`,
      "-N",
      "-B",
      mysqlDatabase,
      "-e",
      sql,
    ],
    {
      encoding: "utf8",
    },
  );
  return output
    .split(/\r?\n/u)
    .map((line) => line.trim())
    .filter(Boolean);
}

function pluginDemoSourceTableExists() {
  const escapedDatabase = mysqlDatabase.replaceAll("'", "''");
  const lines = queryMysqlLines(
    `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '${escapedDatabase}' AND table_name = '${pluginRecordTableName}';`,
  );
  return lines[0] === "1";
}

function listPluginDemoSourceRecordTitles() {
  if (!pluginDemoSourceTableExists()) {
    return [];
  }
  return queryMysqlLines(
    `SELECT title FROM \`${pluginRecordTableName}\` ORDER BY id ASC;`,
  );
}

function hasPluginDemoSourceStoredFiles() {
  return directoryContainsFiles(pluginDemoSourceStorageRoot);
}

function directoryContainsFiles(dirPath: string): boolean {
  if (!existsSync(dirPath)) {
    return false;
  }
  for (const entry of readdirSync(dirPath, { withFileTypes: true })) {
    const fullPath = path.join(dirPath, entry.name);
    if (entry.isFile()) {
      return true;
    }
    if (entry.isDirectory() && directoryContainsFiles(fullPath)) {
      return true;
    }
  }
  return false;
}

async function loginAsAdmin(page: Page) {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);
}

async function waitForCounterStable(
  readCounter: () => number,
  stableMs: number,
) {
  let lastValue = readCounter();
  let lastChangedAt = Date.now();

  await expect
    .poll(
      () => {
        const currentValue = readCounter();
        if (currentValue !== lastValue) {
          lastValue = currentValue;
          lastChangedAt = Date.now();
        }
        return Date.now() - lastChangedAt;
      },
      {
        timeout: stableMs + 2000,
        intervals: [100, 100, 200, 200, 500],
      },
    )
    .toBeGreaterThanOrEqual(stableMs);

  return lastValue;
}

test.describe("TC-66 源码插件生命周期", () => {
  let adminApi: APIRequestContext | null = null;

  test.beforeAll(async () => {
    mkdirSync(path.dirname(pluginDemoSourceFixturePath), { recursive: true });
    writeFileSync(pluginDemoSourceFixturePath, pluginDemoSourceFixtureContent);
    adminApi = await createAdminApiContext();
  });

  test.beforeEach(async () => {
    resetPluginDemoSourceData();
    resetPluginRegistryRow(demoControlPluginID);
    // `GET /plugins` re-syncs source manifests and refreshes the host-side
    // enabled snapshot. This lets the suite clear a previously enabled
    // demo-control guard before the write-heavy lifecycle scenarios begin.
    await listPlugins(adminApi!);
    resetPluginRegistryRow(pluginID);
    await syncPlugins(adminApi!);
  });

  test.afterAll(async () => {
    resetPluginDemoSourceData();
    resetPluginRegistryRow(pluginID);
    rmSync(pluginDemoSourceFixturePath, { force: true });
    if (adminApi) {
      await adminApi.dispose();
    }
  });

  test("TC-66a: 同步 source 插件后保持未安装且默认禁用态", async ({ page }) => {
    const pluginAfterSync = await findPlugin(adminApi!);
    expect(pluginAfterSync, `同步后应发现 ${pluginID}`).toBeTruthy();
    expect(pluginAfterSync?.installed, "源码插件同步后应保持未安装").toBe(0);
    expect(
      "runtime" in ((pluginAfterSync ?? {}) as Record<string, unknown>),
      "插件列表接口不应再返回重复的 runtime 字段",
    ).toBeFalsy();
    expect(
      pluginAfterSync?.installedAt ?? "",
      "源码插件同步后不应记录安装时间",
    ).toBeFalsy();
    expect(
      pluginAfterSync?.enabled ?? pluginAfterSync?.status,
      "源码插件首次同步后应默认禁用",
    ).toBe(0);

    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await expect(loginPage.pluginLoginSlot).toHaveCount(0);
    await loginPage.loginAndWaitForRedirect(config.adminUser, config.adminPass);

    const pluginPage = new PluginPage(page);
    await pluginPage.gotoWorkspace();
    await pluginPage.expectWorkspaceSlotHidden();
    await pluginPage.gotoManage();
    await expect(pluginPage.pluginRow(pluginID)).toBeVisible();
    await expect(pluginPage.pluginEnabledSwitch(pluginID)).toHaveAttribute(
      "aria-checked",
      "false",
    );
    await pluginPage.expectPluginSwitchDisabled(pluginID);
    await pluginPage.expectTableColumnVisible("插件类型");
    await pluginPage.expectTableColumnVisible("安装时间");
    await pluginPage.expectTableColumnHidden("交付方式");
    await pluginPage.expectTableColumnHidden("接入态");
    await pluginPage.expectTableColumnHidden("入口");
    await pluginPage.expectTableColumnHidden("生命周期");
    await pluginPage.expectTableColumnHidden("治理摘要");
    await pluginPage.expectTableColumnBetween("描述", "版本", "状态");
    await pluginPage.expectDescriptionUsesNativeTooltip(pluginID);
    await pluginPage.expectInstallActionVisible(pluginID);
    await pluginPage.expectUninstallActionHidden(pluginID);
    await pluginPage.expectCrudSlotsHidden();
    await pluginPage.expectHeaderSlotsHidden();
    await pluginPage.expectSidebarMenuHidden(pluginMenuName);
  });

  test("TC-66b: 页面安装前先展示源码插件详情，安装后进入已安装未启用态且仍不渲染额外 slots", async ({
    page,
  }) => {
    const summaryBeforeInstall = await fetchPluginSummary(adminApi!);
    expect(
      summaryBeforeInstall.status(),
      "未安装时源码插件受保护路由应返回 404",
    ).toBe(404);

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await pluginPage.openInstallAuthorization(pluginID);
    await expect(pluginPage.hostServiceAuthModal()).toContainText(pluginMenuName);
    await expect(pluginPage.hostServiceAuthModal()).toContainText(pluginID);
    await expect(pluginPage.hostServiceAuthModal()).toContainText("源码插件");
    await expect(pluginPage.hostServiceAuthModal()).toContainText(
      "提供左侧菜单页面与公开/受保护路由示例的源码插件",
    );
    await pluginPage.confirmHostServiceAuthorization();

    const pluginAfterInstall = await findPlugin(adminApi!);
    expect(pluginAfterInstall, `安装后应发现 ${pluginID}`).toBeTruthy();
    expect(pluginAfterInstall?.installed, "源码插件安装后应处于已安装态").toBe(
      1,
    );
    expect(
      pluginAfterInstall?.installedAt,
      "源码插件安装后应记录安装时间",
    ).toBeTruthy();
    expect(
      pluginAfterInstall?.enabled ?? pluginAfterInstall?.status,
      "源码插件安装后应保持禁用，等待显式启用",
    ).toBe(0);

    await expect(pluginPage.pluginEnabledSwitch(pluginID)).toHaveAttribute(
      "aria-checked",
      "false",
    );
    await pluginPage.expectInstallActionHidden(pluginID);
    await pluginPage.expectUninstallActionVisible(pluginID);
    await pluginPage.expectSidebarMenuHidden(pluginMenuName);
    await pluginPage.expectHeaderSlotsHidden();
    await pluginPage.expectCrudSlotsHidden();
    await pluginPage.gotoWorkspace();
    await pluginPage.expectWorkspaceSlotHidden();

    const summaryAfterInstall = await fetchPluginSummary(adminApi!);
    expect(
      summaryAfterInstall.status(),
      "已安装但未启用时源码插件受保护路由仍应返回 404",
    ).toBe(404);
  });

  test("TC-66c: 启用后仅左侧菜单页可正常展示，且不渲染额外 slots", async ({
    page,
  }) => {
    await installAndEnablePlugin(adminApi!);

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

  test("TC-66d: 启用后可验证插件路由与鉴权访问控制", async () => {
    await installAndEnablePlugin(adminApi!);

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
  });

  test("TC-66e: 禁用后不渲染源码样例额外内容且隐藏菜单", async ({ page }) => {
    await installAndEnablePlugin(adminApi!);
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

  test("TC-66f: 禁用后源码插件仍保留已安装态并支持卸载", async ({ page }) => {
    await installAndEnablePlugin(adminApi!);
    await updatePluginStatus(adminApi!, pluginID, false);

    const pluginAfterDisable = await findPlugin(adminApi!);
    expect(
      pluginAfterDisable,
      "禁用后仍应可在清单中发现 source 插件",
    ).toBeTruthy();
    expect(
      pluginAfterDisable?.installed ?? 0,
      "源码插件禁用后仍应保持已安装态",
    ).toBe(1);

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await expect(pluginPage.pluginRow(pluginID)).toBeVisible();
    await expect(pluginPage.pluginEnabledSwitch(pluginID)).toHaveAttribute(
      "aria-checked",
      "false",
    );
    await pluginPage.expectInstallActionHidden(pluginID);
    await pluginPage.expectUninstallActionVisible(pluginID);
  });

  test("TC-66g: 登录态在线启用后立即刷新左侧菜单且不渲染额外 slots", async ({
    page,
  }) => {
    await installPlugin(adminApi!);

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

  test("TC-66h: 登录态在线禁用后立即隐藏左侧菜单且保持无额外 slots", async ({
    page,
  }) => {
    await installAndEnablePlugin(adminApi!);

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

  test("TC-66i: 当前会话重新获得焦点后自动同步外部插件状态变更", async ({
    page,
  }) => {
    await installPlugin(adminApi!);

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

  test("TC-66j: 登录态在线卸载后立即回到未安装态并隐藏菜单", async ({
    page,
  }) => {
    await installAndEnablePlugin(adminApi!);

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await pluginPage.expectSidebarMenuVisible(pluginMenuName);

    await pluginPage.uninstallPlugin(pluginID);

    const pluginAfterUninstall = await findPlugin(adminApi!);
    expect(
      pluginAfterUninstall,
      `卸载后仍应能看到 ${pluginID} 条目`,
    ).toBeTruthy();
    expect(
      pluginAfterUninstall?.installed ?? 1,
      "源码插件卸载后应回到未安装态",
    ).toBe(0);
    expect(
      pluginAfterUninstall?.enabled ?? pluginAfterUninstall?.status ?? 1,
      "源码插件卸载后应保持禁用",
    ).toBe(0);

    await expect(pluginPage.pluginEnabledSwitch(pluginID)).toHaveAttribute(
      "aria-checked",
      "false",
    );
    await pluginPage.expectPluginSwitchDisabled(pluginID);
    await pluginPage.expectInstallActionVisible(pluginID);
    await pluginPage.expectUninstallActionHidden(pluginID);
    await pluginPage.expectSidebarMenuHidden(pluginMenuName);
    await pluginPage.expectHeaderSlotsHidden();
    await pluginPage.expectCrudSlotsHidden();
    await pluginPage.gotoWorkspace();
    await pluginPage.expectWorkspaceSlotHidden();

    const summaryResponse = await fetchPluginSummary(adminApi!);
    expect(summaryResponse.status(), "插件卸载后插件自有路由应返回 404").toBe(
      404,
    );
  });

  test("TC-66k: 按钮权限不会被返回为左侧导航菜单或动态路由", async ({
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
    await expect(page).toHaveTitle("插件管理 - LinaPro");
    await pluginPage.expectSidebarMenuHidden("插件查询");
    await pluginPage.expectSidebarMenuHidden("用户查询");
  });

  test("TC-66l: 当前会话重新获得焦点但插件状态未变化时不重复刷新菜单", async ({
    page,
  }) => {
    await installAndEnablePlugin(adminApi!);

    const menuResponses: string[] = [];
    page.on("response", (response) => {
      if (
        response.request().method() === "GET" &&
        response.url().includes("/api/v1/menus/all")
      ) {
        menuResponses.push(response.url());
      }
    });

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await pluginPage.expectSidebarMenuVisible(pluginMenuName);
    const baselineMenuResponseCount = await waitForCounterStable(
      () => menuResponses.length,
      1200,
    );

    await page.evaluate(() => {
      window.dispatchEvent(new Event("focus"));
      document.dispatchEvent(new Event("visibilitychange"));
    });

    const focusMenuResponseCount = await waitForCounterStable(
      () => menuResponses.length,
      1200,
    );
    await pluginPage.expectSidebarMenuVisible(pluginMenuName);
    expect(
      focusMenuResponseCount,
      "插件状态未变化时，焦点恢复不应重复拉取菜单",
    ).toBe(baselineMenuResponseCount);
  });

  test("TC-66m: 登录后打开插件管理页时公共插件状态接口不重复重查", async ({
    page,
  }) => {
    await installAndEnablePlugin(adminApi!);

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
    const runtimeStateResponseCount = await waitForCounterStable(
      () => runtimeStateResponses.length,
      1500,
    );

    expect(
      runtimeStateResponseCount,
      "登录并打开插件管理页时，公共插件状态接口不应重复触发多次",
    ).toBeLessThanOrEqual(2);
  });

  test("TC-66n: 示例页面可对插件安装 SQL 创建的数据表执行增删改查并下载附件", async ({
    page,
  }) => {
    const createdTitle = "TC0066-UI-新增记录";
    const updatedTitle = "TC0066-UI-编辑后记录";

    await installAndEnablePlugin(adminApi!);

    const pluginPage = new PluginPage(page);
    await loginAsAdmin(page);
    await pluginPage.gotoManage();
    await pluginPage.openSidebarExampleFromMenu();
    await expect(
      pluginPage.pluginSourceRecordRow(pluginRecordSeedTitle),
    ).toBeVisible();

    await pluginPage.createSourceDemoRecord(
      createdTitle,
      "TC0066 通过页面新增一条示例记录",
      pluginDemoSourceFixturePath,
    );

    const recordsAfterCreate = await listDemoRecords(adminApi!);
    expect(
      recordsAfterCreate.map((item: { title: string }) => item.title),
      "插件示例页面新增后应能从接口查询到对应记录",
    ).toContain(createdTitle);

    const download = await pluginPage.downloadSourceDemoAttachment(
      path.basename(pluginDemoSourceFixturePath),
    );
    await download.saveAs(pluginDemoSourceDownloadPath);
    expect(download.suggestedFilename(), "附件下载应保留原始文件名").toBe(
      path.basename(pluginDemoSourceFixturePath),
    );
    expect(
      readFileSync(pluginDemoSourceDownloadPath, "utf8"),
      "下载后的附件内容应与上传文件一致",
    ).toContain(pluginDemoSourceFixtureContent);

    await pluginPage.editSourceDemoRecord(
      createdTitle,
      updatedTitle,
      "TC0066 更新后的记录内容",
    );
    await expect(pluginPage.pluginSourceRecordRow(createdTitle)).toHaveCount(0);

    await pluginPage.deleteSourceDemoRecord(updatedTitle);

    const remainingTitles = (await listDemoRecords(adminApi!)).map(
      (item: { title: string }) => item.title,
    );
    expect(remainingTitles).toContain(pluginRecordSeedTitle);
    expect(remainingTitles).not.toContain(updatedTitle);
    expect(
      hasPluginDemoSourceStoredFiles(),
      "删除带附件记录后应同步清理插件自有存储文件",
    ).toBeFalsy();
  });

  test("TC-66o: 禁用源码插件后插件示例数据表记录和存储文件仍然保留", async ({
    page,
  }) => {
    const customTitle = "TC0066-禁用后保留记录";

    await installAndEnablePlugin(adminApi!);
    await createDemoRecord(
      adminApi!,
      customTitle,
      "禁用插件后该记录仍应保留",
      pluginDemoSourceFixturePath,
    );

    expect(
      pluginDemoSourceTableExists(),
      "安装并写入后插件示例表应存在",
    ).toBeTruthy();
    expect(
      hasPluginDemoSourceStoredFiles(),
      "安装并上传附件后应存在插件自有文件",
    ).toBeTruthy();

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await pluginPage.setPluginEnabled(pluginID, false);

    const pluginAfterDisable = await findPlugin(adminApi!);
    expect(pluginAfterDisable?.installed ?? 0).toBe(1);
    expect(pluginAfterDisable?.enabled ?? pluginAfterDisable?.status ?? 1).toBe(
      0,
    );
    expect(
      listPluginDemoSourceRecordTitles(),
      "源码插件禁用后插件示例表数据应被保留",
    ).toEqual(expect.arrayContaining([pluginRecordSeedTitle, customTitle]));
    expect(
      hasPluginDemoSourceStoredFiles(),
      "源码插件禁用后插件自有存储文件应被保留",
    ).toBeTruthy();
  });

  test("TC-66p: 卸载时不勾选清理存储数据会保留表数据和文件，并在重装后恢复展示", async ({
    page,
  }) => {
    const customTitle = "TC0066-卸载保留记录";

    await installAndEnablePlugin(adminApi!);
    await createDemoRecord(
      adminApi!,
      customTitle,
      "卸载时取消清理选项后应保留该记录",
      pluginDemoSourceFixturePath,
    );

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await pluginPage.uninstallPluginWithOptions(pluginID, false);

    expect(
      pluginDemoSourceTableExists(),
      "卸载时不清理存储数据应保留插件示例表",
    ).toBeTruthy();
    expect(
      listPluginDemoSourceRecordTitles(),
      "卸载时不清理存储数据应保留插件示例表中的业务记录",
    ).toEqual(expect.arrayContaining([pluginRecordSeedTitle, customTitle]));
    expect(
      hasPluginDemoSourceStoredFiles(),
      "卸载时不清理存储数据应保留插件自有存储文件",
    ).toBeTruthy();

    await pluginPage.installPlugin(pluginID);
    await pluginPage.setPluginEnabled(pluginID, true);
    await pluginPage.openSidebarExampleFromMenu();
    await expect(
      pluginPage.pluginSourceRecordRow(pluginRecordSeedTitle),
    ).toBeVisible();
    await expect(pluginPage.pluginSourceRecordRow(customTitle)).toBeVisible();
  });

  test("TC-66q: 卸载时勾选清理存储数据会删除表和文件，重装后仅恢复安装 SQL 初始数据", async ({
    page,
  }) => {
    const customTitle = "TC0066-卸载清理记录";

    await installAndEnablePlugin(adminApi!);
    await createDemoRecord(
      adminApi!,
      customTitle,
      "卸载时勾选清理选项后应删除该记录",
      pluginDemoSourceFixturePath,
    );

    await loginAsAdmin(page);
    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await pluginPage.uninstallPluginWithOptions(pluginID, true);

    expect(
      pluginDemoSourceTableExists(),
      "卸载时勾选清理存储数据后应删除插件示例表",
    ).toBeFalsy();
    expect(
      hasPluginDemoSourceStoredFiles(),
      "卸载时勾选清理存储数据后应删除插件自有存储文件",
    ).toBeFalsy();

    await pluginPage.installPlugin(pluginID);
    await pluginPage.setPluginEnabled(pluginID, true);
    await pluginPage.openSidebarExampleFromMenu();

    expect(
      listPluginDemoSourceRecordTitles(),
      "重新安装后只应恢复安装 SQL 初始化的种子记录",
    ).toEqual([pluginRecordSeedTitle]);
    expect(
      hasPluginDemoSourceStoredFiles(),
      "重新安装后在未重新上传附件前不应残留旧插件文件",
    ).toBeFalsy();
    await expect(
      pluginPage.pluginSourceRecordRow(pluginRecordSeedTitle),
    ).toBeVisible();
    await expect(pluginPage.pluginSourceRecordRow(customTitle)).toHaveCount(0);
  });
});
