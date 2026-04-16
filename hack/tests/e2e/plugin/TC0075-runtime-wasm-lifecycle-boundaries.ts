import { execFileSync } from "node:child_process";
import { mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";

import type { APIRequestContext, APIResponse } from "@playwright/test";

import { request as playwrightRequest } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";
import { config } from "../../fixtures/config";

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? "http://127.0.0.1:8080/api/v1/";
const publicBaseURL =
  process.env.E2E_PUBLIC_BASE_URL ?? apiBaseURL.replace(/\/api\/v1\/?$/, "");
const mysqlBin = process.env.E2E_MYSQL_BIN ?? "mysql";
const mysqlUser = process.env.E2E_DB_USER ?? "root";
const mysqlPassword = process.env.E2E_DB_PASSWORD ?? "12345678";
const mysqlDatabase = process.env.E2E_DB_NAME ?? "lina";

const primaryPluginID = "plugin-lifecycle-boundary-e2e";
const badABIPluginID = "plugin-lifecycle-bad-abi-e2e";
const baseVersion = "v1.0.0";
const higherVersion = "v1.1.0";
const lowerVersion = "v0.9.0";
const networkURLPattern = "https://*.example.com/api";
const dataTableName = "sys_plugin_node_state";
const pageMenuKey = `plugin:${primaryPluginID}:page-entry`;
const buttonMenuKey = `plugin:${primaryPluginID}:inspect`;
const assetPathForVersion = (version: string) =>
  `/plugin-assets/${primaryPluginID}/${version}/index.html`;

type HostServiceSpec = {
  methods: string[];
  paths?: string[];
  resources?: Array<{ url?: string }>;
  service: string;
  tables?: string[];
};

type HostServiceAuthorizationReq = {
  services: Array<{
    methods?: string[];
    paths?: string[];
    resourceRefs?: string[];
    service: string;
    tables?: string[];
  }>;
};

type PluginListItem = {
  authorizedHostServices?: Array<Record<string, unknown>>;
  enabled?: number;
  id: string;
  installed?: number;
  version?: string;
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

function artifactPath(pluginID: string, version: string) {
  return path.join(tempDir(), `${pluginID}-${version}.wasm`);
}

function runtimeStorageArtifactPath(pluginID: string) {
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

function querySQLInt(sql: string) {
  const rows = querySQLRows(sql);
  if (rows.length === 0) {
    return 0;
  }
  return Number.parseInt(rows[0]!, 10) || 0;
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

function cleanupPluginRows(pluginIDs: string[]) {
  const statements: string[] = [];

  for (const pluginID of pluginIDs) {
    const escapedID = pluginID.replaceAll("'", "''");
    statements.push(
      `DELETE FROM sys_role_menu WHERE menu_id IN (SELECT menu_ids.id FROM (SELECT id FROM sys_menu WHERE menu_key LIKE 'plugin:${escapedID}:%') AS menu_ids);`,
      `DELETE FROM sys_menu WHERE menu_key LIKE 'plugin:${escapedID}:%';`,
      `DELETE FROM sys_plugin_node_state WHERE plugin_id = '${escapedID}';`,
      `DELETE FROM sys_plugin_resource_ref WHERE plugin_id = '${escapedID}';`,
      `DELETE FROM sys_plugin_migration WHERE plugin_id = '${escapedID}';`,
      `DELETE FROM sys_plugin_release WHERE plugin_id = '${escapedID}';`,
      `DELETE FROM sys_plugin WHERE plugin_id = '${escapedID}';`,
    );
  }

  execSQL(statements);
}

function cleanupWorkspace() {
  rmSync(artifactPath(primaryPluginID, baseVersion), { force: true });
  rmSync(artifactPath(primaryPluginID, higherVersion), { force: true });
  rmSync(artifactPath(primaryPluginID, lowerVersion), { force: true });
  rmSync(artifactPath(badABIPluginID, baseVersion), { force: true });
  rmSync(runtimeStorageArtifactPath(primaryPluginID), { force: true });
  rmSync(runtimeStorageArtifactPath(badABIPluginID), { force: true });
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

function buildRuntimeArtifact(options: {
  abiVersion?: string;
  hostServices?: HostServiceSpec[];
  id: string;
  includeMenus?: boolean;
  includePageAsset?: boolean;
  version: string;
}) {
  const includeMenus = options.includeMenus ?? false;
  const includePageAsset = options.includePageAsset ?? false;
  const frontendAssets = includePageAsset
    ? [
        {
          path: "index.html",
          contentBase64: Buffer.from(
            `<html><body><h1>${options.id}-${options.version}</h1></body></html>`,
          ).toString("base64"),
          contentType: "text/html",
        },
      ]
    : [];
  const menus = includeMenus
    ? [
        {
          key: pageMenuKey,
          name: "生命周期边界示例",
          path: assetPathForVersion(options.version),
          perms: `${primaryPluginID}:page:view`,
          icon: "ant-design:appstore-outlined",
          type: "M",
          sort: 1,
          remark: "Runtime lifecycle boundary menu entry.",
        },
        {
          key: buttonMenuKey,
          parent_key: pageMenuKey,
          name: "边界按钮权限",
          perms: `${primaryPluginID}:inspect`,
          type: "B",
          sort: 1,
          remark: "Runtime lifecycle boundary button permission.",
        },
      ]
    : [];

  const bytes: number[] = [0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00];
  appendCustomSection(
    bytes,
    "lina.plugin.manifest",
    Buffer.from(
      JSON.stringify({
        description: "Runtime plugin used to verify lifecycle boundary scenarios.",
        id: options.id,
        menus,
        name: `${options.id}-${options.version}`,
        type: "dynamic",
        version: options.version,
      }),
    ),
  );
  appendCustomSection(
    bytes,
    "lina.plugin.dynamic",
    Buffer.from(
      JSON.stringify({
        abiVersion: options.abiVersion ?? "v1",
        frontendAssetCount: frontendAssets.length,
        runtimeKind: "wasm",
        sqlAssetCount: 0,
      }),
    ),
  );

  if (frontendAssets.length > 0) {
    appendCustomSection(
      bytes,
      "lina.plugin.frontend.assets",
      Buffer.from(JSON.stringify(frontendAssets)),
    );
  }
  if ((options.hostServices ?? []).length > 0) {
    appendCustomSection(
      bytes,
      "lina.plugin.backend.host-services",
      Buffer.from(
        JSON.stringify(
          (options.hostServices ?? []).map((item) => ({
            methods: item.methods,
            resources:
              item.service === "data"
                ? { tables: item.tables ?? [] }
                : item.resources ?? [],
            service: item.service,
          })),
        ),
      ),
    );
  }

  return Buffer.from(bytes);
}

function writeArtifact(pluginID: string, version: string, content: Buffer) {
  mkdirSync(tempDir(), { recursive: true });
  const filePath = artifactPath(pluginID, version);
  writeFileSync(filePath, content);
  return filePath;
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

async function createPublicContext(): Promise<APIRequestContext> {
  return playwrightRequest.newContext({ baseURL: publicBaseURL });
}

async function listPlugins(adminApi: APIRequestContext): Promise<PluginListItem[]> {
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
  filePath: string,
  overwrite = true,
) {
  const response = await adminApi.post("plugins/dynamic/package", {
    multipart: {
      overwriteSupport: overwrite ? "1" : "0",
      file: {
        name: path.basename(filePath),
        mimeType: "application/wasm",
        buffer: readFileSync(filePath),
      },
    },
  });
  await expectApiSuccess(response, `上传动态插件失败: ${filePath}`);
}

async function installPlugin(adminApi: APIRequestContext, pluginID: string) {
  const response = await adminApi.post(`plugins/${pluginID}/install`);
  await expectApiSuccess(response, `安装动态插件失败: ${pluginID}`);
}

async function installPluginWithAuthorization(
  adminApi: APIRequestContext,
  pluginID: string,
  authorization: HostServiceAuthorizationReq,
) {
  const response = await adminApi.post(`plugins/${pluginID}/install`, {
    data: { authorization },
  });
  await expectApiSuccess(response, `安装动态插件失败: ${pluginID}`);
}

async function uninstallPlugin(adminApi: APIRequestContext, pluginID: string) {
  const response = await adminApi.delete(`plugins/${pluginID}`);
  await expectApiSuccess(response, `卸载动态插件失败: ${pluginID}`);
}

async function setPluginEnabled(
  adminApi: APIRequestContext,
  pluginID: string,
  enabled: boolean,
) {
  const response = await adminApi.put(
    enabled ? `plugins/${pluginID}/enable` : `plugins/${pluginID}/disable`,
  );
  await expectApiSuccess(
    response,
    `切换插件状态失败: ${pluginID} enabled=${enabled}`,
  );
}

async function setPluginEnabledWithAuthorization(
  adminApi: APIRequestContext,
  pluginID: string,
  enabled: boolean,
  authorization: HostServiceAuthorizationReq,
) {
  const response = await adminApi.put(
    enabled ? `plugins/${pluginID}/enable` : `plugins/${pluginID}/disable`,
    enabled ? { data: { authorization } } : undefined,
  );
  await expectApiSuccess(
    response,
    `切换插件状态失败: ${pluginID} enabled=${enabled}`,
  );
}

function buildAuthorizationPayload(
  hostServices: HostServiceSpec[],
): HostServiceAuthorizationReq {
  return {
    services: hostServices.map((hostService) => ({
      methods: [...hostService.methods],
      paths: [...(hostService.paths ?? [])],
      resourceRefs: (hostService.resources ?? [])
        .map((resource) => resource.url ?? "")
        .filter(Boolean),
      service: hostService.service,
      tables: [...(hostService.tables ?? [])],
    })),
  };
}

async function expectHostedAssetStatus(
  publicApi: APIRequestContext,
  pathName: string,
  expectedStatus: number,
) {
  const response = await publicApi.get(pathName);
  expect(response.status()).toBe(expectedStatus);
  return response;
}

test.describe("TC-75 Runtime Wasm Lifecycle Boundaries", () => {
  let adminApi: APIRequestContext;
  let publicApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    publicApi = await createPublicContext();
  });

  test.afterAll(async () => {
    await adminApi.dispose();
    await publicApi.dispose();
    cleanupPluginRows([primaryPluginID, badABIPluginID]);
    cleanupWorkspace();
  });

  test.beforeEach(async () => {
    cleanupPluginRows([primaryPluginID, badABIPluginID]);
    cleanupWorkspace();
  });

  test.afterEach(async () => {
    cleanupPluginRows([primaryPluginID, badABIPluginID]);
    cleanupWorkspace();
  });

  test("TC-75a: 卸载动态插件后会清理菜单绑定与治理资源，并停止公开托管资源访问", async () => {
    const requestedHostServices: HostServiceSpec[] = [
      {
        methods: ["request"],
        resources: [{ url: networkURLPattern }],
        service: "network",
      },
      {
        methods: ["list", "get"],
        service: "data",
        tables: [dataTableName],
      },
    ];
    const baseArtifactPath = writeArtifact(
      primaryPluginID,
      baseVersion,
      buildRuntimeArtifact({
        hostServices: requestedHostServices,
        id: primaryPluginID,
        includeMenus: true,
        includePageAsset: true,
        version: baseVersion,
      }),
    );
    const authorization = buildAuthorizationPayload(requestedHostServices);

    await uploadDynamicPlugin(adminApi, baseArtifactPath);
    await installPluginWithAuthorization(adminApi, primaryPluginID, authorization);
    await setPluginEnabledWithAuthorization(
      adminApi,
      primaryPluginID,
      true,
      authorization,
    );

    await expect
      .poll(async () => (await findPlugin(adminApi, primaryPluginID))?.enabled ?? 0)
      .toBe(1);

    const menuKeys = [pageMenuKey, buttonMenuKey];
    const menuIDs = querySQLRows(
      `SELECT id FROM sys_menu WHERE menu_key IN ('${pageMenuKey}', '${buttonMenuKey}') ORDER BY menu_key ASC;`,
    );
    expect(menuIDs.length, "安装后应生成插件菜单和按钮权限").toBe(2);

    const roleMenuCount = querySQLInt(
      `SELECT COUNT(*) FROM sys_role_menu WHERE menu_id IN (${menuIDs.join(",")});`,
    );
    expect(roleMenuCount, "安装后应为管理员角色写入菜单绑定").toBeGreaterThan(0);

    const resourceRefCount = querySQLInt(
      `SELECT COUNT(*) FROM sys_plugin_resource_ref WHERE plugin_id = '${primaryPluginID}';`,
    );
    expect(resourceRefCount, "安装后应写入插件治理资源索引").toBeGreaterThan(0);

    const installedPlugin = await findPlugin(adminApi, primaryPluginID);
    expect(
      installedPlugin?.authorizedHostServices?.length ?? 0,
      "安装后应持久化授权快照",
    ).toBeGreaterThan(0);

    await expectHostedAssetStatus(
      publicApi,
      assetPathForVersion(baseVersion),
      200,
    );

    await uninstallPlugin(adminApi, primaryPluginID);
    await expect
      .poll(async () => (await findPlugin(adminApi, primaryPluginID))?.installed ?? 1)
      .toBe(0);
    await expect
      .poll(async () => (await findPlugin(adminApi, primaryPluginID))?.enabled ?? 1)
      .toBe(0);

    const menuCountAfterUninstall = querySQLInt(
      `SELECT COUNT(*) FROM sys_menu WHERE menu_key IN ('${menuKeys.join("','")}');`,
    );
    expect(menuCountAfterUninstall).toBe(0);

    const roleMenuCountAfterUninstall = querySQLInt(
      `SELECT COUNT(*) FROM sys_role_menu WHERE menu_id IN (${menuIDs.join(",")});`,
    );
    expect(roleMenuCountAfterUninstall).toBe(0);

    const resourceRefCountAfterUninstall = querySQLInt(
      `SELECT COUNT(*) FROM sys_plugin_resource_ref WHERE plugin_id = '${primaryPluginID}';`,
    );
    expect(resourceRefCountAfterUninstall).toBe(0);

    await expectHostedAssetStatus(
      publicApi,
      assetPathForVersion(baseVersion),
      404,
    );
  });

  test("TC-75b: 宿主会拒绝不兼容 ABI 与已安装插件的同版本上传", async () => {
    const baseArtifactPath = writeArtifact(
      primaryPluginID,
      baseVersion,
      buildRuntimeArtifact({
        id: primaryPluginID,
        includeMenus: false,
        includePageAsset: false,
        version: baseVersion,
      }),
    );
    const badABIArtifactPath = writeArtifact(
      badABIPluginID,
      baseVersion,
      buildRuntimeArtifact({
        abiVersion: "v2",
        id: badABIPluginID,
        includeMenus: false,
        includePageAsset: false,
        version: baseVersion,
      }),
    );

    await uploadDynamicPlugin(adminApi, baseArtifactPath);
    await installPlugin(adminApi, primaryPluginID);
    await expect
      .poll(async () => (await findPlugin(adminApi, primaryPluginID))?.installed ?? 0)
      .toBe(1);

    const sameVersionResponse = await adminApi.post("plugins/dynamic/package", {
      multipart: {
        overwriteSupport: "1",
        file: {
          name: path.basename(baseArtifactPath),
          mimeType: "application/wasm",
          buffer: readFileSync(baseArtifactPath),
        },
      },
    });
    await expectApiFailure(
      sameVersionResponse,
      "已安装插件不应允许上传同版本产物",
      "只允许上传更高版本",
    );

    const badABIResponse = await adminApi.post("plugins/dynamic/package", {
      multipart: {
        overwriteSupport: "1",
        file: {
          name: path.basename(badABIArtifactPath),
          mimeType: "application/wasm",
          buffer: readFileSync(badABIArtifactPath),
        },
      },
    });
    await expectApiFailure(
      badABIResponse,
      "不兼容 ABI 的动态插件必须被宿主拒绝",
      "ABI 版本不受支持",
    );

    expect(await findPlugin(adminApi, badABIPluginID)).toBeNull();
  });

  test("TC-75c: 宿主会拒绝低版本产物回退覆盖已安装版本", async () => {
    const higherArtifactPath = writeArtifact(
      primaryPluginID,
      higherVersion,
      buildRuntimeArtifact({
        id: primaryPluginID,
        includeMenus: false,
        includePageAsset: false,
        version: higherVersion,
      }),
    );
    const lowerArtifactPath = writeArtifact(
      primaryPluginID,
      lowerVersion,
      buildRuntimeArtifact({
        id: primaryPluginID,
        includeMenus: false,
        includePageAsset: false,
        version: lowerVersion,
      }),
    );

    await uploadDynamicPlugin(adminApi, higherArtifactPath);
    await installPlugin(adminApi, primaryPluginID);
    await expect
      .poll(async () => (await findPlugin(adminApi, primaryPluginID))?.version ?? "")
      .toBe(higherVersion);

    writeFileSync(
      runtimeStorageArtifactPath(primaryPluginID),
      readFileSync(lowerArtifactPath),
    );

    const rollbackInstallResponse = await adminApi.post(
      `plugins/${primaryPluginID}/install`,
    );
    await expectApiFailure(
      rollbackInstallResponse,
      "宿主不应允许回退到更低版本的动态插件产物",
      "不支持回退到更低版本",
    );

    await expect
      .poll(async () => (await findPlugin(adminApi, primaryPluginID))?.version ?? "")
      .toBe(higherVersion);
  });
});
