import { execFileSync } from "node:child_process";
import { mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";

import type { APIRequestContext, APIResponse, Page } from "@playwright/test";

import { request as playwrightRequest, expect } from "@playwright/test";

import { test } from "../../fixtures/auth";
import { config } from "../../fixtures/config";
import { PluginPage } from "../../pages/PluginPage";

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? "http://127.0.0.1:8080/api/v1/";
const publicBaseURL =
  process.env.E2E_PUBLIC_BASE_URL ?? apiBaseURL.replace(/\/api\/v1\/?$/, "");
const mysqlBin = process.env.E2E_MYSQL_BIN ?? "mysql";
const mysqlUser = process.env.E2E_DB_USER ?? "root";
const mysqlPassword = process.env.E2E_DB_PASSWORD ?? "12345678";
const mysqlDatabase = process.env.E2E_DB_NAME ?? "lina";

const pluginID = "plugin-dynamic-hot-upgrade";
const pluginName = "Dynamic Hot Upgrade Plugin";
const pluginMenuKey = "plugin:plugin-dynamic-hot-upgrade:iframe-entry";
const pluginMenuName = "动态插件热升级示例";
const versionOne = "v0.1.0";
const versionTwo = "v0.2.0";
const versionThree = "v0.3.0";
const markerOne = "hot-upgrade-version-one";
const markerTwo = "hot-upgrade-version-two";
const markerThree = "hot-upgrade-version-three";

type PluginRegistryRow = {
  currentState: string;
  enabled: number;
  generation: number;
  installed: number;
  version: string;
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

function safeParseJSON(value: string) {
  try {
    return JSON.parse(value);
  } catch {
    return null;
  }
}

async function assertApiFailure(response: APIResponse, message: string) {
  const payloadText = await response.text();
  if (!response.ok()) {
    expect(
      payloadText.length > 0,
      `${message} should include an error payload when the HTTP status fails`,
    ).toBeTruthy();
    return safeParseJSON(payloadText);
  }

  const payload = safeParseJSON(payloadText) as null | { code?: number };
  expect(
    payload && typeof payload.code === "number",
    `${message} should expose a business error code when HTTP status stays 2xx`,
  ).toBeTruthy();
  expect(payload?.code, `${message} should return a non-zero business error code`).not.toBe(
    0,
  );
  return payload;
}

function repoRoot() {
  return path.resolve(process.cwd(), "../..");
}

function tempDir() {
  return path.join(repoRoot(), "temp");
}

function tempFixtureDir() {
  return path.join(tempDir(), "plugin-hot-upgrade");
}

function tempArtifactPath(version: string) {
  return path.join(tempFixtureDir(), `${pluginID}-${version}.wasm`);
}

function runtimeStorageDir() {
  return path.join(tempDir(), "output");
}

function runtimeStorageArtifactPath() {
  return path.join(runtimeStorageDir(), `${pluginID}.wasm`);
}

function runtimeReleaseArchiveDir() {
  return path.join(runtimeStorageDir(), "releases", pluginID);
}

function hostedAssetPath(version: string, relativePath = "index.html") {
  return `/plugin-assets/${pluginID}/${version}/${relativePath}`;
}

function hostedAssetURL(version: string, relativePath = "index.html") {
  return `${publicBaseURL}${hostedAssetPath(version, relativePath)}`;
}

function cleanupRuntimeWorkspace() {
  rmSync(tempFixtureDir(), { force: true, recursive: true });
  rmSync(runtimeStorageArtifactPath(), { force: true });
  rmSync(runtimeReleaseArchiveDir(), { force: true, recursive: true });
}

function cleanupRuntimeRows() {
  const escapedID = pluginID.replaceAll("'", "''");
  const escapedMenuKey = pluginMenuKey.replaceAll("'", "''");

  execFileSync(
    mysqlBin,
    [
      `-u${mysqlUser}`,
      `-p${mysqlPassword}`,
      mysqlDatabase,
      "-e",
      [
        `DELETE FROM sys_role_menu WHERE menu_id IN (SELECT menu_ids.id FROM (SELECT id FROM sys_menu WHERE menu_key = '${escapedMenuKey}') AS menu_ids);`,
        `DELETE FROM sys_menu WHERE menu_key = '${escapedMenuKey}';`,
        `DELETE FROM sys_plugin_node_state WHERE plugin_id = '${escapedID}';`,
        `DELETE FROM sys_plugin_resource_ref WHERE plugin_id = '${escapedID}';`,
        `DELETE FROM sys_plugin_migration WHERE plugin_id = '${escapedID}';`,
        `DELETE FROM sys_plugin_release WHERE plugin_id = '${escapedID}';`,
        `DELETE FROM sys_plugin WHERE plugin_id = '${escapedID}';`,
      ].join(" "),
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

function buildManifestMenu(version: string) {
  return [
    {
      key: pluginMenuKey,
      name: pluginMenuName,
      path: hostedAssetPath(version),
      perms: "plugin-dynamic-hot-upgrade:view",
      icon: "ant-design:history-outlined",
      type: "M",
      sort: -4,
      remark: "Hot-upgrade runtime iframe entry used by Playwright verification.",
    },
  ];
}

function buildRuntimeFrontendHTML(version: string, marker: string) {
  return [
    "<html>",
    "<body>",
    `  <main data-testid="plugin-hot-upgrade-page">`,
    `    <h1>${pluginName}</h1>`,
    `    <p data-testid="plugin-hot-upgrade-version">${version}</p>`,
    `    <p data-testid="plugin-hot-upgrade-marker">${marker}</p>`,
    "  </main>",
    "</body>",
    "</html>",
  ].join("\n");
}

function buildRuntimeWasmArtifact(options: {
  brokenUpgradeSQL?: boolean;
  marker: string;
  version: string;
}) {
  const installSQLAssets = options.brokenUpgradeSQL
    ? [
        {
          key: `001-${pluginID}.sql`,
          content: "INSERT INTO missing_hot_upgrade_table(id) VALUES (1);",
        },
      ]
    : [];

  const manifestPayload = Buffer.from(
    JSON.stringify({
      id: pluginID,
      name: pluginName,
      version: options.version,
      type: "dynamic",
      description: "Runtime plugin used by Playwright hot-upgrade verification.",
      menus: buildManifestMenu(options.version),
    }),
  );
  const runtimePayload = Buffer.from(
    JSON.stringify({
      runtimeKind: "wasm",
      abiVersion: "v1",
      frontendAssetCount: 1,
      sqlAssetCount: installSQLAssets.length,
    }),
  );
  const frontendAssetsPayload = Buffer.from(
    JSON.stringify([
      {
        path: "index.html",
        contentBase64: Buffer.from(
          buildRuntimeFrontendHTML(options.version, options.marker),
        ).toString("base64"),
        contentType: "text/html; charset=utf-8",
      },
    ]),
  );

  const bytes: number[] = [0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00];
  appendCustomSection(bytes, "lina.plugin.manifest", manifestPayload);
  appendCustomSection(bytes, "lina.plugin.dynamic", runtimePayload);
  appendCustomSection(
    bytes,
    "lina.plugin.frontend.assets",
    frontendAssetsPayload,
  );
  if (installSQLAssets.length > 0) {
    appendCustomSection(
      bytes,
      "lina.plugin.install.sql",
      Buffer.from(JSON.stringify(installSQLAssets)),
    );
  }
  return Buffer.from(bytes);
}

function writeRuntimeArtifact(version: string, marker: string, broken = false) {
  mkdirSync(tempFixtureDir(), { recursive: true });
  const artifactPath = tempArtifactPath(version);
  writeFileSync(
    artifactPath,
    buildRuntimeWasmArtifact({
      version,
      marker,
      brokenUpgradeSQL: broken,
    }),
  );
  return artifactPath;
}

async function createAdminApiContext(): Promise<APIRequestContext> {
  const anonymousApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const loginResponse = await anonymousApi.post("auth/login", {
    data: {
      username: config.adminUser,
      password: config.adminPass,
    },
  });
  assertOk(loginResponse, "管理员登录 API 失败");
  const loginPayload = unwrapApiData(await loginResponse.json());
  await anonymousApi.dispose();

  expect(loginPayload?.accessToken, "管理员登录后应返回 accessToken").toBeTruthy();
  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${loginPayload.accessToken as string}`,
    },
  });
}

async function uploadDynamicPlugin(
  adminApi: APIRequestContext,
  artifactPath: string,
) {
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
  assertOk(response, `上传动态插件失败: ${artifactPath}`);
}

async function installPlugin(adminApi: APIRequestContext) {
  const response = await adminApi.post(`plugins/${pluginID}/install`);
  assertOk(response, "安装动态插件失败");
}

async function installPluginExpectFailure(adminApi: APIRequestContext) {
  return await adminApi.post(`plugins/${pluginID}/install`);
}

async function setPluginEnabled(
  adminApi: APIRequestContext,
  enabled: boolean,
) {
  const response = await adminApi.put(
    enabled ? `plugins/${pluginID}/enable` : `plugins/${pluginID}/disable`,
  );
  assertOk(response, `切换动态插件状态失败: enabled=${enabled}`);
}

function getPluginRegistryRow() {
  const escapedID = pluginID.replaceAll("'", "''");
  const rows = querySQLRows(
    [
      "SELECT",
      "  IFNULL(version, ''),",
      "  installed,",
      "  status,",
      "  generation,",
      "  IFNULL(current_state, '')",
      "FROM sys_plugin",
      `WHERE plugin_id = '${escapedID}'`,
      "LIMIT 1;",
    ].join(" "),
  );
  if (rows.length === 0) {
    return null;
  }

  const [version = "", installed = "0", enabled = "0", generation = "0", currentState = ""] =
    rows[0]!.split("\t");
  return {
    version,
    installed: Number(installed),
    enabled: Number(enabled),
    generation: Number(generation),
    currentState,
  } satisfies PluginRegistryRow;
}

async function waitForPluginRegistryState(
  expected: {
    enabled: number;
    installed: number;
    version?: string;
  },
) {
  await expect
    .poll(async () => {
      const state = getPluginRegistryRow();
      const versionMatches =
        expected.version === undefined
          ? true
          : (state?.version ?? "") === expected.version;
      return (
        (state?.enabled ?? -1) === expected.enabled &&
        (state?.installed ?? -1) === expected.installed &&
        versionMatches
      );
    })
    .toBe(true);
}

async function expectHostedAsset(
  page: Page,
  version: string,
  expectedStatus: number,
  marker?: string,
) {
  const response = await page.request.get(hostedAssetURL(version));
  expect(response.status()).toBe(expectedStatus);
  if (marker) {
    expect(await response.text()).toContain(marker);
  }
  return response;
}

async function triggerPluginRegistryFocusCheck(page: Page) {
  // The shell compares plugin registry snapshots when the current tab regains
  // focus. Triggering a focus event simulates an operator changing plugin state
  // elsewhere while this page remains open.
  await page.evaluate(() => {
    window.dispatchEvent(new Event("focus"));
  });
}

async function bootstrapEnabledRuntimePlugin(
  page: Page,
  adminApi: APIRequestContext,
  version: string,
  marker: string,
) {
  const pluginPage = new PluginPage(page);
  await pluginPage.gotoWorkspace();

  const artifactPath = writeRuntimeArtifact(version, marker);
  await uploadDynamicPlugin(adminApi, artifactPath);
  await installPlugin(adminApi);
  await waitForPluginRegistryState({
    installed: 1,
    enabled: 0,
  });

  await setPluginEnabled(adminApi, true);
  await waitForPluginRegistryState({
    installed: 1,
    enabled: 1,
  });

  await triggerPluginRegistryFocusCheck(page);
  await expect(pluginPage.sidebarMenuItem(pluginMenuName)).toBeVisible();
  return pluginPage;
}

async function openPluginPageAtVersion(
  page: Page,
  adminApi: APIRequestContext,
  version: string,
  marker: string,
) {
  const pluginPage = await bootstrapEnabledRuntimePlugin(
    page,
    adminApi,
    version,
    marker,
  );
  const pluginMenuItem = pluginPage.sidebarMenuItem(pluginMenuName);
  await expect(pluginMenuItem).toBeVisible();
  await pluginMenuItem.click();
  await expect(
    pluginPage.pluginIframeFrame().getByRole("heading", { name: pluginName }),
  ).toBeVisible();
  await expect(
    pluginPage.pluginIframeFrame().getByText(marker, { exact: true }),
  ).toBeVisible();
  await expect(pluginPage.pluginIframe()).toHaveAttribute(
    "src",
    new RegExp(`${pluginID}/${version}/index\\.html`),
  );
  return pluginPage;
}

test.describe("TC-70 动态插件热升级与回滚", () => {
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

  test("TC-70a: 当前插件页热升级时旧版本资源继续可访问并出现刷新提示", async ({
    adminPage,
  }) => {
    const pluginPage = await openPluginPageAtVersion(
      adminPage,
      adminApi!,
      versionOne,
      markerOne,
    );

    const upgradeArtifactPath = writeRuntimeArtifact(versionTwo, markerTwo);
    await uploadDynamicPlugin(adminApi!, upgradeArtifactPath);
    await installPlugin(adminApi!);

    await waitForPluginRegistryState({
      installed: 1,
      enabled: 1,
    });

    // The archived release must keep the old asset reachable even after the
    // mutable staging file has been replaced by the new upload.
    await expectHostedAsset(adminPage, versionOne, 200, markerOne);
    await expectHostedAsset(adminPage, versionTwo, 200, markerTwo);

    await triggerPluginRegistryFocusCheck(adminPage);
    await expect(pluginPage.pluginPageRefreshNotice()).toBeVisible();
    await expect(pluginPage.pluginPageRefreshButton()).toBeVisible();

    // The page must stay on the old generation until the user explicitly
    // accepts the refresh prompt.
    await expect(
      pluginPage.pluginIframeFrame().getByText(markerOne, { exact: true }),
    ).toBeVisible();
    await expect(pluginPage.pluginIframe()).toHaveAttribute(
      "src",
      new RegExp(`${pluginID}/${versionOne}/index\\.html`),
    );
  });

  test("TC-70b: 点击刷新当前页面后切换到新代际资源", async ({
    adminPage,
  }) => {
    const pluginPage = await openPluginPageAtVersion(
      adminPage,
      adminApi!,
      versionOne,
      markerOne,
    );

    const upgradeArtifactPath = writeRuntimeArtifact(versionTwo, markerTwo);
    await uploadDynamicPlugin(adminApi!, upgradeArtifactPath);
    await installPlugin(adminApi!);
    await waitForPluginRegistryState({
      installed: 1,
      enabled: 1,
    });

    await triggerPluginRegistryFocusCheck(adminPage);
    await expect(pluginPage.pluginPageRefreshButton()).toBeVisible();

    await pluginPage.pluginPageRefreshButton().click();
    await expect(
      pluginPage.pluginIframeFrame().getByText(markerTwo, { exact: true }),
    ).toBeVisible();
    await expect(pluginPage.pluginIframe()).toHaveAttribute(
      "src",
      new RegExp(`${pluginID}/${versionTwo}/index\\.html`),
    );
    await expect(pluginPage.pluginPageRefreshNotice()).toHaveCount(0);
  });

  test("TC-70c: 非插件页面用户在热升级后保持无感", async ({ adminPage }) => {
    const pluginPage = await bootstrapEnabledRuntimePlugin(
      adminPage,
      adminApi!,
      versionOne,
      markerOne,
    );

    await pluginPage.gotoWorkspace();
    const upgradeArtifactPath = writeRuntimeArtifact(versionTwo, markerTwo);
    await uploadDynamicPlugin(adminApi!, upgradeArtifactPath);
    await installPlugin(adminApi!);
    await waitForPluginRegistryState({
      installed: 1,
      enabled: 1,
    });

    await triggerPluginRegistryFocusCheck(adminPage);
    await expect(adminPage).toHaveURL(/\/dashboard\/workspace(?:\/)?$/);
    await expect(pluginPage.pluginPageRefreshNotice()).toHaveCount(0);
    await expect(
      adminPage.getByText("开始您一天的工作吧！").first(),
    ).toBeVisible();
  });

  test("TC-70d: 升级失败时宿主回滚到稳定版本并保护当前页面", async ({
    adminPage,
  }) => {
    const pluginPage = await openPluginPageAtVersion(
      adminPage,
      adminApi!,
      versionTwo,
      markerTwo,
    );

    const failedArtifactPath = writeRuntimeArtifact(
      versionThree,
      markerThree,
      true,
    );
    await uploadDynamicPlugin(adminApi!, failedArtifactPath);
    const failedInstallResponse = await installPluginExpectFailure(adminApi!);
    await assertApiFailure(
      failedInstallResponse,
      "升级失败时应返回错误响应",
    );

    await waitForPluginRegistryState({
      installed: 1,
      enabled: 1,
    });

    await expectHostedAsset(adminPage, versionTwo, 200, markerTwo);
    await expectHostedAsset(adminPage, versionThree, 404);

    await triggerPluginRegistryFocusCheck(adminPage);
    await expect(pluginPage.pluginPageRefreshNotice()).toHaveCount(0);
    await expect(
      pluginPage.pluginIframeFrame().getByText(markerTwo, { exact: true }),
    ).toBeVisible();
    await expect(pluginPage.pluginIframe()).toHaveAttribute(
      "src",
      new RegExp(`${pluginID}/${versionTwo}/index\\.html`),
    );
  });
});
