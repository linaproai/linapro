import { execFileSync } from "node:child_process";
import { readFileSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";

import type { APIRequestContext, APIResponse } from "@playwright/test";

import { request as playwrightRequest, expect } from "@playwright/test";

import { test } from "../../fixtures/auth";
import { config } from "../../fixtures/config";

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? "http://127.0.0.1:8080/api/v1/";
const mysqlBin = process.env.E2E_MYSQL_BIN ?? "mysql";
const mysqlUser = process.env.E2E_DB_USER ?? "root";
const mysqlPassword = process.env.E2E_DB_PASSWORD ?? "12345678";
const mysqlDatabase = process.env.E2E_DB_NAME ?? "lina";

const pluginID = "plugin-dynamic-governance";
const pluginName = "Runtime Governance Plugin";
const pluginVersion = "v0.2.0";
const pluginMenuKey = "plugin:plugin-dynamic-governance:main-entry";
const pluginButtonMenuKey = "plugin:plugin-dynamic-governance:records:list";
const pluginMenuName = "运行时治理示例";
const pluginPermission = "plugin-dynamic-governance:records:list";
const pluginRecordTable = "plugin_runtime_governance_record";
const testRoleName = "运行时治理角色";
const testRoleKey = "runtime_governance_role";
const testUsername = "runtime_governance_user";
const testPassword = "runtime123";
const testNickname = "运行时治理用户";

type RuntimeFrontendAsset = {
  path: string;
  content: string;
  contentType: string;
};

type RuntimeSQLAsset = {
  key: string;
  content: string;
};

type RuntimeMenuSpec = {
  key: string;
  parent_key?: string;
  name: string;
  path?: string;
  component?: string;
  perms?: string;
  icon?: string;
  type?: string;
  sort?: number;
  visible?: number;
  status?: number;
  is_frame?: number;
  is_cache?: number;
  query?: Record<string, unknown>;
  query_param?: string;
  remark?: string;
};

type RuntimeResourceSpec = {
  key: string;
  type: string;
  table: string;
  fields: Array<{ name: string; column: string }>;
  filters?: Array<{ param: string; column: string; operator: string }>;
  orderBy: { column: string; direction: string };
  dataScope?: {
    userColumn?: string;
    deptColumn?: string;
  };
};

type UserInfoPayload = {
  menus?: Array<MenuTreeNode>;
  permissions?: string[];
};

type MenuTreeNode = {
  name?: string;
  children?: MenuTreeNode[];
};

type UserDetailPayload = {
  id?: number;
  deptId?: number;
};

type RoleCreatePayload = {
  id?: number;
};

type UserCreatePayload = {
  id?: number;
};

type PluginResourceResponse = {
  list?: Array<Record<string, unknown>>;
  total?: number;
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

function repoRoot() {
  return path.resolve(process.cwd(), "../..");
}

function tempDir() {
  return path.join(repoRoot(), "temp");
}

function runtimeStorageDir() {
  return path.join(tempDir(), "output");
}

function tempArtifactPath() {
  return path.join(tempDir(), `${pluginID}.wasm`);
}

function runtimeStorageArtifactPath() {
  return path.join(runtimeStorageDir(), `${pluginID}.wasm`);
}

function runtimePluginDir() {
  return path.join(repoRoot(), "apps", "lina-plugins", pluginID);
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

function buildRuntimeWasmArtifact(options: {
  id: string;
  name: string;
  version: string;
  description: string;
  menus?: RuntimeMenuSpec[];
  frontendAssets?: RuntimeFrontendAsset[];
  installSQLAssets?: RuntimeSQLAsset[];
  uninstallSQLAssets?: RuntimeSQLAsset[];
  resourceSpecs?: RuntimeResourceSpec[];
}) {
  const menus = options.menus ?? [];
  const frontendAssets = options.frontendAssets ?? [];
  const installSQLAssets = options.installSQLAssets ?? [];
  const uninstallSQLAssets = options.uninstallSQLAssets ?? [];
  const resourceSpecs = options.resourceSpecs ?? [];

  const manifestPayload = Buffer.from(
    JSON.stringify({
      id: options.id,
      name: options.name,
      version: options.version,
      type: "dynamic",
      description: options.description,
      menus,
    }),
  );
  const runtimePayload = Buffer.from(
    JSON.stringify({
      runtimeKind: "wasm",
      abiVersion: "v1",
      frontendAssetCount: frontendAssets.length,
      sqlAssetCount: installSQLAssets.length + uninstallSQLAssets.length,
    }),
  );

  const bytes: number[] = [0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00];
  appendCustomSection(bytes, "lina.plugin.manifest", manifestPayload);
  appendCustomSection(bytes, "lina.plugin.dynamic", runtimePayload);

  if (frontendAssets.length > 0) {
    appendCustomSection(
      bytes,
      "lina.plugin.frontend.assets",
      Buffer.from(
        JSON.stringify(
          frontendAssets.map((asset) => ({
            path: asset.path,
            contentBase64: Buffer.from(asset.content).toString("base64"),
            contentType: asset.contentType,
          })),
        ),
      ),
    );
  }
  if (installSQLAssets.length > 0) {
    appendCustomSection(
      bytes,
      "lina.plugin.install.sql",
      Buffer.from(JSON.stringify(installSQLAssets)),
    );
  }
  if (uninstallSQLAssets.length > 0) {
    appendCustomSection(
      bytes,
      "lina.plugin.uninstall.sql",
      Buffer.from(JSON.stringify(uninstallSQLAssets)),
    );
  }
  if (resourceSpecs.length > 0) {
    appendCustomSection(
      bytes,
      "lina.plugin.backend.resources",
      Buffer.from(JSON.stringify(resourceSpecs)),
    );
  }
  return Buffer.from(bytes);
}

function writeRuntimeArtifact(content: Buffer) {
  mkdirSync(tempDir(), { recursive: true });
  writeFileSync(tempArtifactPath(), content);
  return tempArtifactPath();
}

function cleanupRuntimeWorkspace() {
  rmSync(runtimePluginDir(), { force: true, recursive: true });
  rmSync(tempArtifactPath(), { force: true });
  rmSync(runtimeStorageArtifactPath(), { force: true });
}

function cleanupGovernanceRows() {
  const escapedPluginID = pluginID.replaceAll("'", "''");
  const escapedRoleKey = testRoleKey.replaceAll("'", "''");
  const escapedUsername = testUsername.replaceAll("'", "''");
  execSQL([
    `DROP TABLE IF EXISTS ${pluginRecordTable};`,
    `DELETE FROM sys_role_menu WHERE role_id IN (SELECT role_ids.id FROM (SELECT id FROM sys_role WHERE \`key\`='${escapedRoleKey}') AS role_ids);`,
    `DELETE FROM sys_user_role WHERE role_id IN (SELECT role_ids.id FROM (SELECT id FROM sys_role WHERE \`key\`='${escapedRoleKey}') AS role_ids);`,
    `DELETE FROM sys_user_dept WHERE user_id IN (SELECT user_ids.id FROM (SELECT id FROM sys_user WHERE username='${escapedUsername}') AS user_ids);`,
    `DELETE FROM sys_user_post WHERE user_id IN (SELECT user_ids.id FROM (SELECT id FROM sys_user WHERE username='${escapedUsername}') AS user_ids);`,
    `DELETE FROM sys_user_role WHERE user_id IN (SELECT user_ids.id FROM (SELECT id FROM sys_user WHERE username='${escapedUsername}') AS user_ids);`,
    `DELETE FROM sys_user WHERE username='${escapedUsername}';`,
    `DELETE FROM sys_role WHERE \`key\`='${escapedRoleKey}';`,
    `DELETE FROM sys_role_menu WHERE menu_id IN (SELECT menu_ids.id FROM (SELECT id FROM sys_menu WHERE menu_key IN ('${pluginMenuKey}', '${pluginButtonMenuKey}')) AS menu_ids);`,
    `DELETE FROM sys_menu WHERE menu_key IN ('${pluginMenuKey}', '${pluginButtonMenuKey}');`,
    `DELETE FROM sys_plugin_node_state WHERE plugin_id='${escapedPluginID}';`,
    `DELETE FROM sys_plugin_resource_ref WHERE plugin_id='${escapedPluginID}';`,
    `DELETE FROM sys_plugin_migration WHERE plugin_id='${escapedPluginID}';`,
    `DELETE FROM sys_plugin_release WHERE plugin_id='${escapedPluginID}';`,
    `DELETE FROM sys_plugin WHERE plugin_id='${escapedPluginID}';`,
  ]);
}

async function loginByPassword(
  username: string,
  password: string,
): Promise<string> {
  const anonymousApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const loginResponse = await anonymousApi.post("auth/login", {
    data: { username, password },
  });
  assertOk(loginResponse, `登录失败: ${username}`);
  const loginPayload = unwrapApiData(await loginResponse.json());
  await anonymousApi.dispose();
  expect(loginPayload?.accessToken, "登录后应返回 accessToken").toBeTruthy();
  return loginPayload.accessToken as string;
}

async function createApiContext(accessToken: string): Promise<APIRequestContext> {
  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

async function createAdminApiContext() {
  return createApiContext(await loginByPassword(config.adminUser, config.adminPass));
}

async function uploadDynamicPlugin(adminApi: APIRequestContext, artifactPath: string) {
  const response = await adminApi.post("plugins/dynamic/package", {
    multipart: {
      overwriteSupport: "0",
      file: {
        name: path.basename(artifactPath),
        mimeType: "application/wasm",
        buffer: readFileSync(artifactPath),
      },
    },
  });
  assertOk(response, "上传动态插件失败");
}

async function installPlugin(adminApi: APIRequestContext) {
  const response = await adminApi.post(`plugins/${pluginID}/install`);
  assertOk(response, "安装动态插件失败");
}

async function setPluginEnabled(adminApi: APIRequestContext, enabled: boolean) {
  const response = await adminApi.put(
    enabled ? `plugins/${pluginID}/enable` : `plugins/${pluginID}/disable`,
  );
  assertOk(response, `切换插件状态失败: ${enabled}`);
}

async function getAdminDeptID(adminApi: APIRequestContext) {
  const response = await adminApi.get("user/1");
  assertOk(response, "查询管理员详情失败");
  const payload = unwrapApiData(await response.json()) as UserDetailPayload;
  expect(payload?.deptId, "管理员必须绑定部门，测试才能验证数据权限").toBeTruthy();
  return payload.deptId as number;
}

async function createRole(adminApi: APIRequestContext, menuIDs: number[]) {
  const response = await adminApi.post("role", {
    data: {
      name: testRoleName,
      key: testRoleKey,
      sort: 10,
      dataScope: 3,
      status: 1,
      remark: "Runtime governance verification role",
      menuIds: menuIDs,
    },
  });
  assertOk(response, "创建 runtime 治理角色失败");
  const payload = unwrapApiData(await response.json()) as RoleCreatePayload;
  expect(payload?.id, "角色创建成功后应返回角色ID").toBeTruthy();
  return payload.id as number;
}

async function createUser(
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
  assertOk(response, "创建 runtime 治理用户失败");
  const payload = unwrapApiData(await response.json()) as UserCreatePayload;
  expect(payload?.id, "用户创建成功后应返回用户ID").toBeTruthy();
  return payload.id as number;
}

async function fetchUserInfo(apiContext: APIRequestContext) {
  const response = await apiContext.get("user/info");
  assertOk(response, "查询用户信息失败");
  return (unwrapApiData(await response.json()) ?? {}) as UserInfoPayload;
}

async function queryPluginResource(apiContext: APIRequestContext) {
  const response = await apiContext.get(
    `plugins/${pluginID}/resources/records?pageNum=1&pageSize=20`,
  );
  assertOk(response, "查询动态插件资源失败");
  return (unwrapApiData(await response.json()) ?? {}) as PluginResourceResponse;
}

function hasMenuName(list: MenuTreeNode[], name: string): boolean {
  return list.some((item) => {
    if (item.name === name) {
      return true;
    }
    return hasMenuName(item.children ?? [], name);
  });
}

function getPluginMenuIDs() {
  const rows = querySQLRows(
    `SELECT id FROM sys_menu WHERE menu_key IN ('${pluginMenuKey}', '${pluginButtonMenuKey}') ORDER BY menu_key ASC;`,
  );
  return rows.map((item) => Number.parseInt(item, 10)).filter(Number.isFinite);
}

function seedUserOwnedPluginRecord(userID: number, deptID: number) {
  execSQL([
    `INSERT INTO ${pluginRecordTable} (title, owner_user_id, owner_dept_id) VALUES ('admin-owned', 1, ${deptID});`,
    `INSERT INTO ${pluginRecordTable} (title, owner_user_id, owner_dept_id) VALUES ('user-owned', ${userID}, ${deptID});`,
  ]);
}

function buildRuntimeGovernanceArtifact() {
  const frontendAssetPath = `/plugin-assets/${pluginID}/${pluginVersion}/index.html`;

  return buildRuntimeWasmArtifact({
    id: pluginID,
    name: pluginName,
    version: pluginVersion,
    description: "Runtime plugin used to verify role-menu governance and data scope filtering.",
    menus: [
      {
        key: pluginMenuKey,
        name: pluginMenuName,
        path: frontendAssetPath,
        perms: "",
        icon: "ant-design:safety-outlined",
        type: "M",
        sort: -1,
        remark: "Runtime governance menu.",
      },
      {
        key: pluginButtonMenuKey,
        parent_key: pluginMenuKey,
        name: "运行时治理按钮",
        perms: pluginPermission,
        icon: "ant-design:api-outlined",
        type: "B",
        sort: 0,
        remark: "Runtime governance button permission.",
      },
    ],
    frontendAssets: [
      {
        path: "index.html",
        content: `<html><body><h1>${pluginName}</h1><p>runtime governance asset</p></body></html>`,
        contentType: "text/html; charset=utf-8",
      },
    ],
    installSQLAssets: [
      {
        key: "001-plugin-dynamic-governance.sql",
        content: [
          `CREATE TABLE IF NOT EXISTS ${pluginRecordTable} (`,
          "  id INT PRIMARY KEY AUTO_INCREMENT,",
          "  title VARCHAR(64) NOT NULL,",
          "  owner_user_id INT NOT NULL,",
          "  owner_dept_id INT NOT NULL",
          ");",
        ].join("\n"),
      },
    ],
    uninstallSQLAssets: [
      {
        key: "001-plugin-dynamic-governance.sql",
        content: [
          `DROP TABLE IF EXISTS ${pluginRecordTable};`,
        ].join("\n"),
      },
    ],
    resourceSpecs: [
      {
        key: "records",
        type: "table-list",
        table: pluginRecordTable,
        fields: [
          { name: "id", column: "id" },
          { name: "title", column: "title" },
          { name: "ownerUserId", column: "owner_user_id" },
          { name: "ownerDeptId", column: "owner_dept_id" },
        ],
        orderBy: { column: "id", direction: "asc" },
        dataScope: {
          userColumn: "owner_user_id",
          deptColumn: "owner_dept_id",
        },
      },
    ],
  });
}

test.describe("TC-69 动态插件权限治理", () => {
  let adminApi: APIRequestContext | null = null;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    cleanupRuntimeWorkspace();
    cleanupGovernanceRows();
    if (adminApi) {
      await adminApi.dispose();
    }
  });

  test.beforeEach(async () => {
    cleanupRuntimeWorkspace();
    cleanupGovernanceRows();
  });

  test.afterEach(async () => {
    cleanupRuntimeWorkspace();
    cleanupGovernanceRows();
  });

  test("TC-69a: 动态插件菜单和按钮权限会跟随角色授权、禁用隐藏与重新启用恢复", async () => {
    const artifactPath = writeRuntimeArtifact(buildRuntimeGovernanceArtifact());
    await uploadDynamicPlugin(adminApi!, artifactPath);
    await installPlugin(adminApi!);
    await setPluginEnabled(adminApi!, true);

    const adminDeptID = await getAdminDeptID(adminApi!);
    const menuIDs = getPluginMenuIDs();
    expect(menuIDs.length, "动态插件菜单和按钮权限都应写入 sys_menu").toBe(2);

    const roleID = await createRole(adminApi!, menuIDs);
    const userID = await createUser(adminApi!, adminDeptID, roleID);
    seedUserOwnedPluginRecord(userID, adminDeptID);

    const userApi = await createApiContext(await loginByPassword(testUsername, testPassword));
    const userInfo = await fetchUserInfo(userApi);
    expect(
      hasMenuName(userInfo.menus ?? [], pluginMenuName),
      "角色授权后，用户应看到动态插件菜单",
    ).toBeTruthy();
    expect(
      userInfo.permissions ?? [],
      "角色授权后，用户应拿到动态插件按钮权限",
    ).toContain(pluginPermission);

    const selfScopedRecords = await queryPluginResource(userApi);
    expect(selfScopedRecords.total, "仅本人数据权限应只返回用户自己的插件记录").toBe(1);
    expect(selfScopedRecords.list?.[0]?.title).toBe("user-owned");

    await setPluginEnabled(adminApi!, false);
    const disabledInfo = await fetchUserInfo(userApi);
    expect(
      hasMenuName(disabledInfo.menus ?? [], pluginMenuName),
      "禁用动态插件后，菜单应从用户视图中隐藏",
    ).toBeFalsy();
    expect(disabledInfo.permissions ?? []).not.toContain(pluginPermission);

    await setPluginEnabled(adminApi!, true);
    const restoredInfo = await fetchUserInfo(userApi);
    expect(
      hasMenuName(restoredInfo.menus ?? [], pluginMenuName),
      "重新启用后，既有角色授权关系应自动恢复",
    ).toBeTruthy();
    expect(restoredInfo.permissions ?? []).toContain(pluginPermission);

    const restoredRecords = await queryPluginResource(userApi);
    expect(restoredRecords.total).toBe(1);
    expect(restoredRecords.list?.[0]?.title).toBe("user-owned");

    await userApi.dispose();
  });
});
