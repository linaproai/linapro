import { execFileSync } from "node:child_process";
import { mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";

import type { APIRequestContext, APIResponse } from "@playwright/test";

import { request as playwrightRequest, expect } from "@playwright/test";

import { test } from "../../fixtures/auth";
import { config } from "../../fixtures/config";
import { PluginPage } from "../../pages/PluginPage";

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? "http://127.0.0.1:8080/api/v1/";
const mysqlBin = process.env.E2E_MYSQL_BIN ?? "mysql";
const mysqlUser = process.env.E2E_DB_USER ?? "root";
const mysqlPassword = process.env.E2E_DB_PASSWORD ?? "12345678";
const mysqlDatabase = process.env.E2E_DB_NAME ?? "lina";

const pluginID = "plugin-dynamic-host-auth-ui";
const pluginVersion = "v0.1.0";
const networkURLPattern = "https://*.example.com/api";
const dataTableName = "sys_plugin_node_state";
const dataTableComment = "插件节点状态表";

type PluginListItem = {
  authorizedHostServices?: Array<{
    resources?: Array<{ ref: string }>;
    service: string;
  }>;
  authorizationStatus?: string;
  enabled?: number;
  id: string;
  installed?: number;
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
      password: config.adminPass,
      username: config.adminUser,
    },
  });
  assertOk(loginResponse, "管理员登录失败");
  const loginResult = unwrapApiData(await loginResponse.json());
  const accessToken = loginResult?.accessToken;
  expect(accessToken, "未获取到管理员 accessToken").toBeTruthy();
  await loginApi.dispose();

  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

async function listPlugins(adminApi: APIRequestContext): Promise<PluginListItem[]> {
  const response = await adminApi.get("plugins");
  assertOk(response, "查询插件列表失败");
  const payload = unwrapApiData(await response.json());
  return payload?.list ?? [];
}

async function findPlugin(adminApi: APIRequestContext, pluginId = pluginID) {
  const list = await listPlugins(adminApi);
  return list.find((item) => item.id === pluginId) ?? null;
}

async function uploadDynamicPlugin(
  adminApi: APIRequestContext,
  artifactPath: string,
) {
  const response = await adminApi.post("plugins/dynamic/package", {
    multipart: {
      file: {
        buffer: readFileSync(artifactPath),
        mimeType: "application/wasm",
        name: path.basename(artifactPath),
      },
      overwriteSupport: "1",
    },
  });
  assertOk(response, "上传动态插件失败");
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

function cleanupPluginRows() {
  const escapedId = pluginID.replaceAll("'", "''");
  execFileSync(
    mysqlBin,
    [
      `-u${mysqlUser}`,
      `-p${mysqlPassword}`,
      mysqlDatabase,
      "-e",
      [
        `DELETE FROM sys_plugin_node_state WHERE plugin_id = '${escapedId}';`,
        `DELETE FROM sys_plugin_resource_ref WHERE plugin_id = '${escapedId}';`,
        `DELETE FROM sys_plugin_migration WHERE plugin_id = '${escapedId}';`,
        `DELETE FROM sys_plugin_release WHERE plugin_id = '${escapedId}';`,
        `DELETE FROM sys_plugin WHERE plugin_id = '${escapedId}';`,
      ].join(" "),
    ],
    {
      stdio: "ignore",
    },
  );
}

function cleanupPluginWorkspace() {
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

function writeAuthorizationReviewArtifact() {
  mkdirSync(tempDir(), { recursive: true });
  const bytes: number[] = [0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00];

  appendCustomSection(
    bytes,
    "lina.plugin.manifest",
    Buffer.from(
      JSON.stringify({
        description: "Host service authorization review plugin.",
        id: pluginID,
        name: "Host Service Authorization Review Plugin",
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
  appendCustomSection(
    bytes,
    "lina.plugin.backend.host-services",
    Buffer.from(
      JSON.stringify([
        {
          methods: ["info.now"],
          service: "runtime",
        },
        {
          methods: ["request"],
          resources: [
            {
              url: networkURLPattern,
            },
          ],
          service: "network",
        },
        {
          methods: ["list", "get"],
          resources: {
            tables: [dataTableName],
          },
          service: "data",
        },
      ]),
    ),
  );

  writeFileSync(artifactPath(), Buffer.from(bytes));
}

test.describe("TC-73 插件安装/启用时审查 hostServices 授权", () => {
  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    cleanupPluginWorkspace();
    cleanupPluginRows();
    writeAuthorizationReviewArtifact();
    await uploadDynamicPlugin(adminApi, artifactPath());
  });

  test.afterAll(async () => {
    await adminApi.dispose();
    cleanupPluginRows();
    cleanupPluginWorkspace();
  });

  test("TC-73a~c: 安装与启用弹窗展示申请权限并持久化最终授权结果", async ({
    adminPage,
  }) => {
    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);

    await pluginPage.openInstallAuthorization(pluginID);
    await expect(pluginPage.hostServiceAuthModal()).toContainText("网络服务");
    await expect(pluginPage.hostServiceAuthModal()).toContainText(networkURLPattern);
    await expect(pluginPage.hostServiceAuthModal()).toContainText("数据服务");
    await expect(pluginPage.hostServiceAuthModal()).toContainText(dataTableName);
    await expect(pluginPage.hostServiceAuthModal()).toContainText(dataTableComment);
    await pluginPage.setHostServiceAuthorization(
      pluginID,
      "network",
      networkURLPattern,
      false,
    );
    await pluginPage.confirmHostServiceAuthorization();
    await expect
      .poll(async () => (await findPlugin(adminApi, pluginID))?.installed ?? 0)
      .toBe(1);

    const installedPlugin = await findPlugin(adminApi, pluginID);
    expect(installedPlugin?.installed).toBe(1);
    expect(installedPlugin?.authorizationStatus).toBe("confirmed");
    expect(
      installedPlugin?.authorizedHostServices?.some(
        (service) => service.service === "network",
      ) ?? false,
    ).toBeFalsy();

    await pluginPage.searchByPluginId(pluginID);
    await pluginPage.openEnableAuthorization(pluginID);
    await expect(pluginPage.hostServiceAuthModal()).toContainText("已确认");
    await expect(pluginPage.hostServiceAuthModal()).toContainText(dataTableComment);
    await expect(
      pluginPage.hostServiceAuthCheckbox(pluginID, "network", networkURLPattern),
    ).not.toBeChecked();
    await expect(
      pluginPage.hostServiceAuthCheckbox(pluginID, "data", dataTableName),
    ).toBeChecked();
    await pluginPage.confirmHostServiceAuthorization();
    await expect(pluginPage.pluginEnabledSwitch(pluginID)).toHaveAttribute(
      "aria-checked",
      "true",
    );

    const enabledPlugin = await findPlugin(adminApi, pluginID);
    expect(enabledPlugin?.enabled).toBe(1);
    expect(
      enabledPlugin?.authorizedHostServices?.some(
        (service) => service.service === "network",
      ) ?? false,
    ).toBeFalsy();
  });
});
