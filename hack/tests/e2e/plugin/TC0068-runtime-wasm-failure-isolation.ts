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

const goodPluginID = "plugin-dynamic-hook-good";
const badPluginID = "plugin-dynamic-hook-bad";
const goodPluginVersion = "v0.2.0";
const badPluginVersion = "v0.2.0";
const goodPluginLogTable = "plugin_runtime_hook_good_log";

type PluginListItem = {
  id: string;
  enabled?: number;
  installed?: number;
};

type PluginResourceResponse = {
  list?: Array<Record<string, unknown>>;
  total?: number;
};

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

type RuntimeHookSpec = {
  event: string;
  action: string;
  mode?: string;
  table?: string;
  fields?: Record<string, string>;
  timeoutMs?: number;
  sleepMs?: number;
  errorMessage?: string;
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

function repoRoot() {
  return path.resolve(process.cwd(), "../..");
}

function tempDir() {
  return path.join(repoRoot(), "temp");
}

function tempArtifactPath(pluginID: string) {
  return path.join(tempDir(), `${pluginID}.wasm`);
}

function runtimeStorageDir() {
  return path.join(tempDir(), "output");
}

function runtimeStorageArtifactPath(pluginID: string) {
  return path.join(runtimeStorageDir(), `${pluginID}.wasm`);
}

function runtimeReleaseArchiveDir(pluginID: string) {
  return path.join(runtimeStorageDir(), "releases", pluginID);
}

function runtimePluginDir(pluginID: string) {
  return path.join(repoRoot(), "apps", "lina-plugins", pluginID);
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
  hookSpecs?: RuntimeHookSpec[];
  resourceSpecs?: RuntimeResourceSpec[];
}) {
  const menus = options.menus ?? [];
  const frontendAssets = options.frontendAssets ?? [];
  const installSQLAssets = options.installSQLAssets ?? [];
  const uninstallSQLAssets = options.uninstallSQLAssets ?? [];
  const hookSpecs = options.hookSpecs ?? [];
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
  if (hookSpecs.length > 0) {
    appendCustomSection(
      bytes,
      "lina.plugin.backend.hooks",
      Buffer.from(JSON.stringify(hookSpecs)),
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

function writeRuntimeArtifact(pluginID: string, content: Buffer) {
  mkdirSync(tempDir(), { recursive: true });
  const artifactPath = tempArtifactPath(pluginID);
  writeFileSync(artifactPath, content);
  return artifactPath;
}

function cleanupRuntimeWorkspace() {
  for (const pluginID of [goodPluginID, badPluginID]) {
    rmSync(runtimePluginDir(pluginID), { force: true, recursive: true });
    rmSync(tempArtifactPath(pluginID), { force: true });
    // Dynamic uploads persist to plugin.dynamic.storagePath instead of the
    // plugin workspace, so the E2E fixture must remove both the mutable staged
    // artifact and the archived release directory to stay self-contained.
    rmSync(runtimeStorageArtifactPath(pluginID), { force: true });
    rmSync(runtimeReleaseArchiveDir(pluginID), { force: true, recursive: true });
  }
}

function cleanupRuntimeRows() {
  const goodID = goodPluginID.replaceAll("'", "''");
  const badID = badPluginID.replaceAll("'", "''");
  execSQL([
    `DROP TABLE IF EXISTS ${goodPluginLogTable};`,
    `DELETE FROM sys_plugin_node_state WHERE plugin_id IN ('${goodID}', '${badID}');`,
    `DELETE FROM sys_plugin_resource_ref WHERE plugin_id IN ('${goodID}', '${badID}');`,
    `DELETE FROM sys_plugin_migration WHERE plugin_id IN ('${goodID}', '${badID}');`,
    `DELETE FROM sys_plugin_release WHERE plugin_id IN ('${goodID}', '${badID}');`,
    `DELETE FROM sys_plugin WHERE plugin_id IN ('${goodID}', '${badID}');`,
  ]);
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

function queryGoodHookRowCount() {
  const output = execFileSync(
    mysqlBin,
    [
      `-u${mysqlUser}`,
      `-p${mysqlPassword}`,
      mysqlDatabase,
      "-N",
      "-B",
      "-e",
      `SELECT COUNT(1) FROM ${goodPluginLogTable};`,
    ],
    { encoding: "utf8" },
  ).trim();
  return Number.parseInt(output || "0", 10);
}

async function loginByPassword(
  username: string,
  password: string,
): Promise<string> {
  const anonymousApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const loginResponse = await anonymousApi.post("auth/login", {
    data: { username, password },
  });
  const loginPayload = await expectApiSuccess<{ accessToken?: string }>(
    loginResponse,
    `登录失败: ${username}`,
  );
  await anonymousApi.dispose();
  expect(loginPayload?.accessToken, "登录后应返回 accessToken").toBeTruthy();
  return loginPayload.accessToken as string;
}

async function createAdminApiContext(): Promise<APIRequestContext> {
  const accessToken = await loginByPassword(config.adminUser, config.adminPass);
  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

async function listPlugins(
  adminApi: APIRequestContext,
): Promise<PluginListItem[]> {
  const response = await adminApi.get("plugins");
  const payload = await expectApiSuccess<{ list?: PluginListItem[] }>(
    response,
    "查询插件列表失败",
  );
  return payload?.list ?? [];
}

async function findPlugin(adminApi: APIRequestContext, pluginID: string) {
  const list = await listPlugins(adminApi);
  return list.find((item) => item.id === pluginID) ?? null;
}

async function uploadDynamicPlugin(
  adminApi: APIRequestContext,
  artifactPath: string,
  overwrite = false,
) {
  const response = await adminApi.post("plugins/dynamic/package", {
    multipart: {
      overwriteSupport: overwrite ? "1" : "0",
      file: {
        name: path.basename(artifactPath),
        mimeType: "application/wasm",
        buffer: readFileSync(artifactPath),
      },
    },
  });
  await expectApiSuccess(
    response,
    `上传 runtime wasm 失败: ${artifactPath}`,
  );
}

async function installPlugin(adminApi: APIRequestContext, pluginID: string) {
  const response = await adminApi.post(`plugins/${pluginID}/install`);
  await expectApiSuccess(response, `安装动态插件失败: ${pluginID}`);
}

async function setPluginEnabled(
  adminApi: APIRequestContext,
  pluginID: string,
  enabled: boolean,
) {
  const response = await adminApi.put(
    enabled ? `plugins/${pluginID}/enable` : `plugins/${pluginID}/disable`,
  );
  await expectApiSuccess(response, `切换插件启停失败: ${pluginID}`);
}

async function queryPluginResource(
  adminApi: APIRequestContext,
  pluginID: string,
  resourceID: string,
): Promise<PluginResourceResponse> {
  const response = await adminApi.get(
    `plugins/${pluginID}/resources/${resourceID}?pageNum=1&pageSize=20`,
  );
  return (
    (await expectApiSuccess<PluginResourceResponse>(
      response,
      `查询插件资源失败: ${pluginID}/${resourceID}`,
    )) ?? {}
  );
}

function buildGoodRuntimeArtifact() {
  return buildRuntimeWasmArtifact({
    id: goodPluginID,
    name: "Runtime Hook Good",
    version: goodPluginVersion,
    description: "Runtime plugin that records successful login hooks.",
    installSQLAssets: [
      {
        key: "001-plugin-dynamic-hook-good.sql",
        content: [
          `CREATE TABLE IF NOT EXISTS ${goodPluginLogTable} (`,
          "  id INT PRIMARY KEY AUTO_INCREMENT,",
          "  user_name VARCHAR(64) NOT NULL,",
          "  event_name VARCHAR(128) NOT NULL,",
          "  created_at DATETIME NULL",
          ");",
        ].join("\n"),
      },
    ],
    uninstallSQLAssets: [
      {
        key: "001-plugin-dynamic-hook-good.sql",
        content: `DROP TABLE IF EXISTS ${goodPluginLogTable};`,
      },
    ],
    hookSpecs: [
      {
        event: "auth.login.succeeded",
        action: "insert",
        mode: "blocking",
        timeoutMs: 1000,
        table: goodPluginLogTable,
        fields: {
          user_name: "event.userName",
          event_name: "event.message",
          created_at: "now",
        },
      },
    ],
    resourceSpecs: [
      {
        key: "login-audit",
        type: "table-list",
        table: goodPluginLogTable,
        fields: [
          { name: "userName", column: "user_name" },
          { name: "eventName", column: "event_name" },
        ],
        orderBy: { column: "id", direction: "desc" },
      },
    ],
  });
}

function buildBadRuntimeArtifact() {
  return buildRuntimeWasmArtifact({
    id: badPluginID,
    name: "Runtime Hook Bad",
    version: badPluginVersion,
    description:
      "Runtime plugin that intentionally times out and returns errors to verify host isolation.",
    hookSpecs: [
      {
        event: "auth.login.succeeded",
        action: "sleep",
        mode: "blocking",
        timeoutMs: 50,
        sleepMs: 800,
      },
      {
        event: "auth.login.succeeded",
        action: "error",
        mode: "blocking",
        timeoutMs: 300,
        errorMessage: "runtime hook failed on purpose",
      },
    ],
  });
}

async function prepareEnabledRuntimePlugins(adminApi: APIRequestContext) {
  const goodArtifactPath = writeRuntimeArtifact(goodPluginID, buildGoodRuntimeArtifact());
  const badArtifactPath = writeRuntimeArtifact(badPluginID, buildBadRuntimeArtifact());

  await uploadDynamicPlugin(adminApi, goodArtifactPath);
  await uploadDynamicPlugin(adminApi, badArtifactPath);

  await installPlugin(adminApi, goodPluginID);
  await installPlugin(adminApi, badPluginID);

  await setPluginEnabled(adminApi, goodPluginID, true);
  await setPluginEnabled(adminApi, badPluginID, true);
}

test.describe("TC-68 运行时 wasm 失败隔离与回收", () => {
  let adminApi: APIRequestContext | null = null;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    cleanupRuntimeWorkspace();
    cleanupRuntimeRows();
    if (adminApi) {
      await adminApi.dispose();
    }
  });

  test.beforeEach(async () => {
    cleanupRuntimeWorkspace();
    cleanupRuntimeRows();
  });

  test.afterEach(async () => {
    cleanupRuntimeWorkspace();
    cleanupRuntimeRows();
  });

  test("TC-68a: runtime hook 超时和报错不会阻断宿主登录，也不会影响其他 runtime hook 执行", async () => {
    await prepareEnabledRuntimePlugins(adminApi!);

    const accessToken = await loginByPassword(config.adminUser, config.adminPass);
    expect(accessToken, "即使存在失败的 runtime hook，宿主登录仍应成功").toBeTruthy();

    const resourceData = await queryPluginResource(
      adminApi!,
      goodPluginID,
      "login-audit",
    );
    expect(resourceData.total, "成功的 runtime hook 应写入登录审计记录").toBe(1);
    expect(resourceData.list?.[0]?.userName).toBe(config.adminUser);
    expect(resourceData.list?.[0]?.eventName).toBe("登录成功");

    const goodPlugin = await findPlugin(adminApi!, goodPluginID);
    const badPlugin = await findPlugin(adminApi!, badPluginID);
    expect(goodPlugin?.enabled).toBe(1);
    expect(badPlugin?.enabled).toBe(1);
  });

  test("TC-68b: 禁用动态插件后会停止其 Hook 执行，重新启用后会恢复", async () => {
    await prepareEnabledRuntimePlugins(adminApi!);

    await loginByPassword(config.adminUser, config.adminPass);
    expect(queryGoodHookRowCount()).toBe(1);

    await setPluginEnabled(adminApi!, goodPluginID, false);
    await loginByPassword(config.adminUser, config.adminPass);
    expect(queryGoodHookRowCount(), "禁用后 runtime hook 不应继续参与宿主登录链路").toBe(1);

    await setPluginEnabled(adminApi!, goodPluginID, true);
    await loginByPassword(config.adminUser, config.adminPass);
    expect(queryGoodHookRowCount(), "重新启用后 runtime hook 应恢复执行").toBe(2);
  });
});
