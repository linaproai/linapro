import { execFileSync } from "node:child_process";
import { mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";

import type { APIRequestContext, APIResponse, Page } from "@playwright/test";

import { request as playwrightRequest } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";
import { config } from "../../fixtures/config";
import { LoginPage } from "../../pages/LoginPage";
import { PluginPage } from "../../pages/PluginPage";

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? "http://127.0.0.1:8080/api/v1/";
const mysqlBin = process.env.E2E_MYSQL_BIN ?? "mysql";
const mysqlUser = process.env.E2E_DB_USER ?? "root";
const mysqlPassword = process.env.E2E_DB_PASSWORD ?? "12345678";
const mysqlDatabase = process.env.E2E_DB_NAME ?? "lina";

const pluginID = "plugin-management-permission-e2e";
const pluginVersion = "v0.1.0";
const testRoleName = "插件管理查询角色";
const testRoleKey = "plugin_management_query_role";
const testUsername = "plugin_management_query_user";
const testPassword = "runtime123";
const testNickname = "插件管理查询用户";

type PluginListItem = {
  enabled?: number;
  id: string;
  installed?: number;
};

type UserDetailPayload = {
  deptId?: number;
};

type RoleCreatePayload = {
  id?: number;
};

type UserCreatePayload = {
  id?: number;
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

async function expectApiSuccess<T = any>(
  response: APIResponse,
  message: string,
): Promise<T> {
  assertOk(response, message);

  const payload = (await response.json()) as {
    code?: number;
    data?: T;
    message?: string;
  };
  expect(
    payload?.code,
    `${message}, business code=${payload?.code}, business message=${payload?.message ?? ""}`,
  ).toBe(0);
  return (payload?.data ?? null) as T;
}

async function expectApiFailure(
  response: APIResponse,
  message: string,
  expectedText: string,
) {
  const text = await response.text();
  if (response.ok()) {
    const payload = JSON.parse(text) as {
      code?: number;
      message?: string;
    };
    expect(payload?.code, `${message}, body=${text}`).not.toBe(0);
    expect(text, `${message}, body=${text}`).toContain(expectedText);
    return;
  }
  expect(text, `${message}, status=${response.status()}`).toContain(expectedText);
}

function repoRoot() {
  return path.resolve(process.cwd(), "../..");
}

function tempDir() {
  return path.join(repoRoot(), "temp");
}

function artifactPath() {
  return path.join(tempDir(), `${pluginID}.wasm`);
}

function runtimeStorageArtifactPath() {
  return path.join(tempDir(), "output", `${pluginID}.wasm`);
}

function querySQLRows(sql: string) {
  return execFileSync(
    mysqlBin,
    [
      `-u${mysqlUser}`,
      `-p${mysqlPassword}`,
      mysqlDatabase,
      "-N",
      "-B",
      "-e",
      sql,
    ],
    { encoding: "utf8" },
  )
    .trim()
    .split("\n")
    .filter(Boolean);
}

function execSQL(statements: string[]) {
  execFileSync(
    mysqlBin,
    [
      `-u${mysqlUser}`,
      `-p${mysqlPassword}`,
      mysqlDatabase,
      "-e",
      statements.join(" "),
    ],
    { stdio: "ignore" },
  );
}

function cleanupPluginRows() {
  const escapedID = pluginID.replaceAll("'", "''");
  execSQL([
    `DELETE FROM sys_role_menu WHERE menu_id IN (SELECT menu_ids.id FROM (SELECT id FROM sys_menu WHERE menu_key LIKE 'plugin:${escapedID}:%') AS menu_ids);`,
    `DELETE FROM sys_menu WHERE menu_key LIKE 'plugin:${escapedID}:%';`,
    `DELETE FROM sys_plugin_node_state WHERE plugin_id = '${escapedID}';`,
    `DELETE FROM sys_plugin_resource_ref WHERE plugin_id = '${escapedID}';`,
    `DELETE FROM sys_plugin_migration WHERE plugin_id = '${escapedID}';`,
    `DELETE FROM sys_plugin_release WHERE plugin_id = '${escapedID}';`,
    `DELETE FROM sys_plugin WHERE plugin_id = '${escapedID}';`,
  ]);
}

function cleanupUserAndRoleRows() {
  const escapedUsername = testUsername.replaceAll("'", "''");
  const escapedRoleKey = testRoleKey.replaceAll("'", "''");
  execSQL([
    `DELETE FROM sys_user_role WHERE user_id IN (SELECT user_ids.id FROM (SELECT id FROM sys_user WHERE username = '${escapedUsername}') AS user_ids);`,
    `DELETE FROM sys_user WHERE username = '${escapedUsername}';`,
    `DELETE FROM sys_role_menu WHERE role_id IN (SELECT role_ids.id FROM (SELECT id FROM sys_role WHERE \`key\` = '${escapedRoleKey}') AS role_ids);`,
    `DELETE FROM sys_role WHERE \`key\` = '${escapedRoleKey}';`,
  ]);
}

function cleanupWorkspace() {
  rmSync(artifactPath(), { force: true });
  rmSync(runtimeStorageArtifactPath(), { force: true });
}

function writeULEB128(buffer: number[], value: number) {
  let current = value >>> 0;
  while (true) {
    let byte = current & 0x7f;
    current >>>= 7;
    if (current !== 0) {
      byte |= 0x80;
    }
    buffer.push(byte);
    if (current === 0) {
      return;
    }
  }
}

function appendCustomSection(buffer: number[], name: string, payload: Buffer) {
  const section: number[] = [];
  writeULEB128(section, Buffer.byteLength(name));
  section.push(...Buffer.from(name));
  section.push(...payload);

  buffer.push(0x00);
  writeULEB128(buffer, section.length);
  buffer.push(...section);
}

function writeRuntimeArtifact() {
  mkdirSync(tempDir(), { recursive: true });

  const bytes: number[] = [0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00];
  appendCustomSection(
    bytes,
    "lina.plugin.manifest",
    Buffer.from(
      JSON.stringify({
        description: "Runtime plugin used to verify plugin management action permissions.",
        id: pluginID,
        name: "Plugin Management Permission E2E",
        type: "dynamic",
        version: pluginVersion,
      }),
    ),
  );
  appendCustomSection(
    bytes,
    "lina.plugin.dynamic",
    Buffer.from(
      JSON.stringify({
        abiVersion: "v1",
        frontendAssetCount: 0,
        runtimeKind: "wasm",
        sqlAssetCount: 0,
      }),
    ),
  );

  writeFileSync(artifactPath(), Buffer.from(bytes));
}

async function createAdminApiContext(): Promise<APIRequestContext> {
  const anonymousApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const loginResponse = await anonymousApi.post("auth/login", {
    data: {
      username: config.adminUser,
      password: config.adminPass,
    },
  });
  const loginPayload = await expectApiSuccess<{ accessToken?: string }>(
    loginResponse,
    "管理员登录失败",
  );
  await anonymousApi.dispose();

  expect(loginPayload?.accessToken, "管理员登录后应返回 accessToken").toBeTruthy();
  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${loginPayload.accessToken as string}`,
    },
  });
}

async function loginByPassword(username: string, password: string) {
  const anonymousApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  try {
    const loginResponse = await anonymousApi.post("auth/login", {
      data: {
        username,
        password,
      },
    });
    const loginPayload = await expectApiSuccess<{ accessToken?: string }>(
      loginResponse,
      `登录失败: ${username}`,
    );
    expect(loginPayload?.accessToken, "登录成功后应返回 accessToken").toBeTruthy();
    return loginPayload.accessToken as string;
  } finally {
    await anonymousApi.dispose();
  }
}

async function createApiContext(token: string): Promise<APIRequestContext> {
  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${token}`,
    },
  });
}

async function listPlugins(adminApi: APIRequestContext): Promise<PluginListItem[]> {
  const response = await adminApi.get("plugins");
  const payload = await expectApiSuccess<{ list?: PluginListItem[] }>(
    response,
    "查询插件列表失败",
  );
  return payload?.list ?? [];
}

async function findPlugin(adminApi: APIRequestContext, id = pluginID) {
  const list = await listPlugins(adminApi);
  return list.find((item) => item.id === id) ?? null;
}

async function uploadDynamicPlugin(adminApi: APIRequestContext) {
  const response = await adminApi.post("plugins/dynamic/package", {
    multipart: {
      overwriteSupport: "1",
      file: {
        name: path.basename(artifactPath()),
        mimeType: "application/wasm",
        buffer: readFileSync(artifactPath()),
      },
    },
  });
  await expectApiSuccess(response, "管理员上传动态插件失败");
}

async function installPlugin(adminApi: APIRequestContext) {
  const response = await adminApi.post(`plugins/${pluginID}/install`);
  await expectApiSuccess(response, "管理员安装动态插件失败");
}

async function setPluginEnabled(
  adminApi: APIRequestContext,
  enabled: boolean,
) {
  const response = await adminApi.put(
    enabled ? `plugins/${pluginID}/enable` : `plugins/${pluginID}/disable`,
  );
  await expectApiSuccess(
    response,
    `管理员切换插件状态失败: enabled=${enabled}`,
  );
}

async function getAdminDeptID(adminApi: APIRequestContext) {
  const response = await adminApi.get("user/1");
  const payload = await expectApiSuccess<UserDetailPayload>(
    response,
    "查询管理员详情失败",
  );
  expect(payload?.deptId, "管理员必须绑定部门").toBeTruthy();
  return payload.deptId as number;
}

function lookupMenuID(menuKey: string) {
  const rows = querySQLRows(
    `SELECT id FROM sys_menu WHERE menu_key = '${menuKey.replaceAll("'", "''")}' LIMIT 1;`,
  );
  expect(rows.length, `未找到菜单: ${menuKey}`).toBe(1);
  return Number.parseInt(rows[0]!, 10);
}

async function createQueryOnlyRole(adminApi: APIRequestContext) {
  const menuIDs = [
    lookupMenuID("system"),
    lookupMenuID("system:plugin:list"),
    lookupMenuID("system:plugin:query"),
  ];

  const response = await adminApi.post("role", {
    data: {
      name: testRoleName,
      key: testRoleKey,
      sort: 10,
      dataScope: 3,
      status: 1,
      remark: "Plugin management query-only role",
      menuIds: menuIDs,
    },
  });
  const payload = await expectApiSuccess<RoleCreatePayload>(
    response,
    "创建查询角色失败",
  );
  expect(payload?.id, "角色创建成功后应返回角色ID").toBeTruthy();
  return payload.id as number;
}

async function createQueryOnlyUser(
  adminApi: APIRequestContext,
  deptID: number,
  roleID: number,
) {
  const response = await adminApi.post("user", {
    data: {
      username: testUsername,
      password: testPassword,
      nickname: testNickname,
      deptId: deptID,
      roleIds: [roleID],
      status: 1,
    },
  });
  const payload = await expectApiSuccess<UserCreatePayload>(
    response,
    "创建查询用户失败",
  );
  expect(payload?.id, "用户创建成功后应返回用户ID").toBeTruthy();
  return payload.id as number;
}

async function loginAsQueryOnlyUser(page: Page) {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginPage.loginAndWaitForRedirect(testUsername, testPassword);
}

test.describe("TC-74 插件管理动作权限校验", () => {
  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    await adminApi.dispose();
    cleanupUserAndRoleRows();
    cleanupPluginRows();
    cleanupWorkspace();
  });

  test.beforeEach(async () => {
    cleanupUserAndRoleRows();
    cleanupPluginRows();
    cleanupWorkspace();
    writeRuntimeArtifact();

    const deptID = await getAdminDeptID(adminApi);
    const roleID = await createQueryOnlyRole(adminApi);
    await createQueryOnlyUser(adminApi, deptID, roleID);
  });

  test.afterEach(async () => {
    cleanupUserAndRoleRows();
    cleanupPluginRows();
    cleanupWorkspace();
  });

  test("TC-74a: 仅具备查询权限的用户可查看插件列表但看不到上传/安装动作", async ({
    page,
  }) => {
    await uploadDynamicPlugin(adminApi);
    expect(await findPlugin(adminApi)).toBeTruthy();

    await loginAsQueryOnlyUser(page);

    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);

    await expect(pluginPage.dynamicUploadTrigger).toHaveCount(0);
    await expect(page.getByRole("button", { name: "同步插件" })).toHaveCount(0);
    await expect(page.getByRole("button", { name: /安\s*装/ })).toHaveCount(0);
  });

  test("TC-74b: 仅具备查询权限的用户调用上传与安装接口会被宿主拒绝", async () => {
    const queryUserToken = await loginByPassword(testUsername, testPassword);
    const queryUserApi = await createApiContext(queryUserToken);

    try {
      const uploadResponse = await queryUserApi.post("plugins/dynamic/package", {
        multipart: {
          overwriteSupport: "1",
          file: {
            name: path.basename(artifactPath()),
            mimeType: "application/wasm",
            buffer: readFileSync(artifactPath()),
          },
        },
      });
      await expectApiFailure(
        uploadResponse,
        "查询用户不应允许上传动态插件",
        "plugin:install",
      );

      await uploadDynamicPlugin(adminApi);
      const installResponse = await queryUserApi.post(`plugins/${pluginID}/install`);
      await expectApiFailure(
        installResponse,
        "查询用户不应允许安装动态插件",
        "plugin:install",
      );
    } finally {
      await queryUserApi.dispose();
    }
  });

  test("TC-74c: 已安装插件对无启停/卸载权限的用户隐藏危险动作并拒绝接口调用", async ({
    page,
  }) => {
    await uploadDynamicPlugin(adminApi);
    await installPlugin(adminApi);
    await setPluginEnabled(adminApi, true);
    await expect
      .poll(async () => (await findPlugin(adminApi))?.enabled ?? 0)
      .toBe(1);

    const queryUserToken = await loginByPassword(testUsername, testPassword);
    const queryUserApi = await createApiContext(queryUserToken);

    try {
      await loginAsQueryOnlyUser(page);

      const pluginPage = new PluginPage(page);
      await pluginPage.gotoManage();
      await pluginPage.searchByPluginId(pluginID);

      await expect(page.getByRole("button", { name: /卸\s*载/ })).toHaveCount(0);
      await expect(pluginPage.pluginEnabledSwitch(pluginID)).toHaveClass(
        /ant-switch-disabled/,
      );

      const disableResponse = await queryUserApi.put(`plugins/${pluginID}/disable`);
      await expectApiFailure(
        disableResponse,
        "查询用户不应允许禁用插件",
        "plugin:disable",
      );

      await setPluginEnabled(adminApi, false);
      await expect
        .poll(async () => (await findPlugin(adminApi))?.enabled ?? 1)
        .toBe(0);

      const enableResponse = await queryUserApi.put(`plugins/${pluginID}/enable`);
      await expectApiFailure(
        enableResponse,
        "查询用户不应允许启用插件",
        "plugin:enable",
      );

      await setPluginEnabled(adminApi, true);
      await expect
        .poll(async () => (await findPlugin(adminApi))?.enabled ?? 0)
        .toBe(1);

      const uninstallResponse = await queryUserApi.delete(`plugins/${pluginID}`);
      await expectApiFailure(
        uninstallResponse,
        "查询用户不应允许卸载插件",
        "plugin:uninstall",
      );
    } finally {
      await queryUserApi.dispose();
    }
  });
});
