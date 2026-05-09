import { mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";

import type { APIRequestContext, APIResponse } from "@playwright/test";

import { request as playwrightRequest, expect } from "@playwright/test";

import { test } from "../../../fixtures/auth";
import { config } from "../../../fixtures/config";
import { PluginPage } from "../../../pages/PluginPage";
import { RolePage } from "../../../pages/RolePage";
import { execPgSQLStatements, pgEscapeLiteral } from "../../../support/postgres";

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? "http://127.0.0.1:8080/api/v1/";

const pluginID = "plugin-dynamic-host-auth-ui";
const pluginVersion = "v0.1.0";
const pluginName = "Host Service Authorization Review Plugin";
const pluginDescription =
  "用于演示动态插件在授权审查与详情弹窗中同时展示宿主服务范围、注册路由清单，以及较长描述信息时的治理体验优化。";
const networkURLPattern = "https://*.example.com/api";
const storagePath = "plugin-demo/records";
const dataTableName = "sys_plugin_node_state";
const dataTableComment = "Plugin node state table";
const routeSummaryPath = `/api/v1/extensions/${pluginID}/review-summary`;
const routeHealthPath = `/api/v1/extensions/${pluginID}/healthz`;
const routeAuditPath = `/api/v1/extensions/${pluginID}/audit-log`;
const routePermission = `${pluginID}:review:query`;
const routeAuditPermission = `${pluginID}:audit:query`;

type PluginListItem = {
  authorizedHostServices?: Array<{
    resources?: Array<{ ref: string }>;
    service: string;
  }>;
  declaredRoutes?: Array<{
    access?: string;
    method?: string;
    permission?: string;
    publicPath?: string;
    summary?: string;
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
  const escapedId = pgEscapeLiteral(pluginID);
  execPgSQLStatements([
    `DELETE FROM sys_plugin_node_state WHERE plugin_id = '${escapedId}';`,
    `DELETE FROM sys_plugin_resource_ref WHERE plugin_id = '${escapedId}';`,
    `DELETE FROM sys_plugin_migration WHERE plugin_id = '${escapedId}';`,
    `DELETE FROM sys_plugin_release WHERE plugin_id = '${escapedId}';`,
    `DELETE FROM sys_plugin WHERE plugin_id = '${escapedId}';`,
  ]);
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
        description: pluginDescription,
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
            paths: [storagePath],
          },
          service: "storage",
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
  appendCustomSection(
    bytes,
    "lina.plugin.backend.routes",
    Buffer.from(
      JSON.stringify([
        {
          access: "login",
          description: "返回当前插件版本的评审摘要。",
          method: "GET",
          path: "/review-summary",
          permission: routePermission,
          summary: "查询评审摘要",
        },
        {
          access: "public",
          description: "返回动态插件公开探活结果。",
          method: "GET",
          path: "/healthz",
          summary: "公开健康检查",
        },
        {
          access: "login",
          description: "返回动态插件审计日志回放结果。",
          method: "GET",
          path: "/audit-log",
          permission: routeAuditPermission,
          summary: "审计日志回放",
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

  test("TC-73a~d: 安装弹窗展示插件详情与授权排序，安装时持久化授权结果，后续启用不再重复确认，并校验角色授权中的动态路由权限文案", async ({
    adminPage,
  }) => {
    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);

    await pluginPage.openInstallAuthorization(pluginID);
    const hostServiceAuthModal = pluginPage.hostServiceAuthModal();
    await expect(hostServiceAuthModal).toContainText(pluginName);
    await expect(hostServiceAuthModal).toContainText(pluginID);
    await expect(hostServiceAuthModal).toContainText(pluginVersion);
    await expect(hostServiceAuthModal).toContainText("动态插件");
    await expect(hostServiceAuthModal).toContainText(pluginDescription);
    await expect(hostServiceAuthModal).toContainText("宿主服务授权审核");
    await expect(hostServiceAuthModal).toContainText("声明的路由");
    await expect(hostServiceAuthModal).toContainText("查询评审摘要");
    await expect(hostServiceAuthModal).toContainText("公开健康检查");
    await expect(hostServiceAuthModal).toContainText(routeSummaryPath);
    await expect(hostServiceAuthModal).toContainText(routeHealthPath);
    await expect(hostServiceAuthModal).toContainText(routePermission);
    await expect(hostServiceAuthModal).toContainText("登录访问");
    await expect(hostServiceAuthModal).toContainText("公开访问");
    await expect(hostServiceAuthModal).toContainText("数据");
    await expect(hostServiceAuthModal).toContainText("存储");
    await expect(hostServiceAuthModal).toContainText("网络");
    await expect(hostServiceAuthModal).toContainText("运行时");
    await expect(hostServiceAuthModal).toContainText("申请范围");
    await expect(hostServiceAuthModal).toContainText("存储路径");
    await expect(hostServiceAuthModal).toContainText("数据表");
    await expect(hostServiceAuthModal).toContainText("路径");
    await expect(hostServiceAuthModal).not.toContainText("申请存储路径");
    await expect(hostServiceAuthModal).not.toContainText("申请数据表名");
    await expect(hostServiceAuthModal).not.toContainText("申请访问地址");
    await expect(hostServiceAuthModal).toContainText(storagePath);
    await expect(hostServiceAuthModal).toContainText(
      `${dataTableName} (${dataTableComment})`,
    );
    await expect(hostServiceAuthModal).toContainText(networkURLPattern);
    await expect(
      hostServiceAuthModal.getByTestId(
        `plugin-host-service-auth-list-${pluginID}-storage`,
      ),
    ).toBeVisible();
    expect(
      await hostServiceAuthModal
        .getByTestId(`plugin-host-service-auth-list-${pluginID}-storage`)
        .evaluate((node) => node.tagName),
    ).toBe("DIV");
    await expect(
      hostServiceAuthModal.getByTestId(
        `plugin-host-service-auth-list-${pluginID}-data`,
      ),
    ).toBeVisible();
    expect(
      await hostServiceAuthModal
        .getByTestId(`plugin-host-service-auth-list-${pluginID}-data`)
        .evaluate((node) => node.tagName),
    ).toBe("DIV");
    await expect(
      hostServiceAuthModal.getByTestId(
        `plugin-host-service-auth-list-${pluginID}-network`,
      ),
    ).toBeVisible();
    expect(
      await hostServiceAuthModal
        .getByTestId(`plugin-host-service-auth-list-${pluginID}-network`)
        .evaluate((node) => node.tagName),
    ).toBe("DIV");
    await expect(
      hostServiceAuthModal.getByTestId(
        `plugin-host-service-auth-item-${pluginID}-storage-${storagePath}`,
      ),
    ).toHaveText(storagePath);
    await expect(
      hostServiceAuthModal.getByTestId(
        `plugin-host-service-auth-item-${pluginID}-data-${dataTableName}`,
      ),
    ).toHaveText(`${dataTableName} (${dataTableComment})`);
    await expect(
      hostServiceAuthModal.getByTestId(
        `plugin-host-service-auth-item-${pluginID}-network-${networkURLPattern}`,
      ),
    ).toHaveText(networkURLPattern);
    await expect(hostServiceAuthModal.getByRole("checkbox")).toHaveCount(0);
    await expect(hostServiceAuthModal).not.toContainText(
      "当前授权状态",
    );
    await expect(hostServiceAuthModal).not.toContainText(
      "未勾选的资源将不会被授权",
    );
    await expect(hostServiceAuthModal).not.toContainText(
      "请选择允许该插件访问",
    );
    await expect(hostServiceAuthModal).not.toContainText(
      "该 URL 模式一旦授权，插件即可直接访问命中的 HTTP 地址。",
    );
    await expect(hostServiceAuthModal).not.toContainText("数据表列表");
    await expect(hostServiceAuthModal).not.toContainText("存储目录前缀列表");
    await expect(hostServiceAuthModal).not.toContainText("URL 模式列表");
    await expect(hostServiceAuthModal).not.toContainText("治理目标:");
    await expect(hostServiceAuthModal).not.toContainText(
      "允许方法:",
    );
    await expect(hostServiceAuthModal).not.toContainText("无需额外确认");
    await expect(hostServiceAuthModal).not.toContainText(
      "当前服务未声明需要单独勾选的资源，宿主将按服务级方法摘要治理。",
    );
    await expect(
      hostServiceAuthModal.getByTestId("plugin-route-review-item-0"),
    ).toBeVisible();
    await expect(
      hostServiceAuthModal.getByTestId("plugin-route-review-item-1"),
    ).toBeVisible();
    await expect(
      hostServiceAuthModal.getByTestId("plugin-route-review-item-2"),
    ).toHaveCount(0);
    await expect(hostServiceAuthModal).not.toContainText("审计日志回放");
    await expect(hostServiceAuthModal).not.toContainText(routeAuditPath);
    const authEffectiveScopeBackground = await hostServiceAuthModal
      .getByTestId("plugin-host-service-scope-label-storage-storage-review")
      .evaluate((node) => getComputedStyle(node).backgroundColor);
    const authSummaryScopeBackground = await hostServiceAuthModal
      .getByTestId("plugin-host-service-summary-label-storage-storage-review")
      .evaluate((node) => getComputedStyle(node).backgroundColor);
    const installModalText = await hostServiceAuthModal.innerText();
    expect(installModalText.indexOf("数据")).toBeGreaterThanOrEqual(0);
    expect(installModalText.indexOf("数据")).toBeLessThan(
      installModalText.indexOf("存储"),
    );
    expect(installModalText.indexOf("存储")).toBeLessThan(
      installModalText.indexOf("网络"),
    );
    expect(installModalText.indexOf("网络")).toBeLessThan(
      installModalText.indexOf("运行时"),
    );
    const authHostServiceTitleTop = await hostServiceAuthModal
      .getByTestId("plugin-host-service-section-title")
      .evaluate((node) => node.getBoundingClientRect().top);
    const authRouteTitleTop = await hostServiceAuthModal
      .getByTestId("plugin-route-section-title")
      .evaluate((node) => node.getBoundingClientRect().top);
    const authHostServiceTitleFontWeight = await hostServiceAuthModal
      .getByTestId("plugin-host-service-section-title")
      .evaluate((node) => Number.parseInt(getComputedStyle(node).fontWeight, 10));
    const authRouteTitleFontWeight = await hostServiceAuthModal
      .getByTestId("plugin-route-section-title")
      .evaluate((node) => Number.parseInt(getComputedStyle(node).fontWeight, 10));
    const authHostServiceTitleFontSize = await hostServiceAuthModal
      .getByTestId("plugin-host-service-section-title")
      .evaluate((node) => getComputedStyle(node).fontSize);
    const authRouteTitleFontSize = await hostServiceAuthModal
      .getByTestId("plugin-route-section-title")
      .evaluate((node) => getComputedStyle(node).fontSize);
    expect(authHostServiceTitleTop).toBeLessThan(authRouteTitleTop);
    expect(authHostServiceTitleFontWeight).toBeGreaterThanOrEqual(600);
    expect(authRouteTitleFontWeight).toBeGreaterThanOrEqual(600);
    expect(authHostServiceTitleFontSize).toBe("15px");
    expect(authRouteTitleFontSize).toBe("15px");
    await expect(pluginPage.pluginRouteReviewToggle()).toHaveText("展开");
    await pluginPage.pluginRouteReviewToggle().click();
    await expect(pluginPage.pluginRouteReviewToggle()).toHaveText("收起");
    await expect(hostServiceAuthModal).toContainText("审计日志回放");
    await expect(hostServiceAuthModal).toContainText(routeAuditPath);
    await expect(hostServiceAuthModal).toContainText(routeAuditPermission);
    await expect(
      hostServiceAuthModal.getByTestId("plugin-route-review-item-2"),
    ).toBeVisible();
    await pluginPage.confirmHostServiceAuthorization();
    await expect
      .poll(async () => (await findPlugin(adminApi, pluginID))?.installed ?? 0)
      .toBe(1);

    const installedPlugin = await findPlugin(adminApi, pluginID);
    expect(installedPlugin?.installed).toBe(1);
    expect(installedPlugin?.authorizationStatus).toBe("confirmed");
    expect(installedPlugin?.declaredRoutes?.length ?? 0).toBe(3);
    expect(
      installedPlugin?.declaredRoutes?.some(
        (route) =>
          route.publicPath === routeSummaryPath &&
          route.access === "login" &&
          route.permission === routePermission &&
          route.summary === "查询评审摘要",
      ) ?? false,
    ).toBeTruthy();
    expect(
      installedPlugin?.authorizedHostServices?.some(
        (service) => service.service === "network",
      ) ?? false,
    ).toBeTruthy();

    await pluginPage.searchByPluginId(pluginID);
    const pluginSwitch = pluginPage.pluginEnabledSwitch(pluginID);
    await expect(pluginSwitch).toHaveAttribute("aria-checked", "false");
    await pluginSwitch.click();
    await expect(pluginPage.hostServiceAuthDialog()).toHaveCount(0);
    await expect(pluginSwitch).toHaveAttribute(
      "aria-checked",
      "true",
    );
    await expect(adminPage.getByText("插件已启用").last()).toBeVisible();

    const enabledPlugin = await findPlugin(adminApi, pluginID);
    expect(enabledPlugin?.enabled).toBe(1);
    expect(
      enabledPlugin?.authorizedHostServices?.some(
        (service) => service.service === "network",
      ) ?? false,
    ).toBeTruthy();

    await pluginPage.openPluginDetail(pluginID);
    const detailModal = pluginPage.pluginDetailModal();
    await expect(detailModal).toContainText("宿主服务范围");
    await expect(detailModal).toContainText("声明的路由");
    await expect(detailModal).toContainText("生效范围");
    await expect(detailModal).not.toContainText("当前生效范围");
    await expect(detailModal).toContainText("存储路径");
    await expect(detailModal).toContainText("数据表");
    await expect(detailModal).toContainText("路径");
    await expect(detailModal).not.toContainText("授权存储路径");
    await expect(detailModal).not.toContainText("授权数据表名");
    await expect(detailModal).not.toContainText("授权访问地址");
    await expect(detailModal).toContainText(
      "以下范围展示的是当前插件元数据与授权快照。",
    );
    await expect(detailModal).toContainText(storagePath);
    await expect(detailModal).toContainText(
      `${dataTableName} (${dataTableComment})`,
    );
    await expect(detailModal).toContainText(networkURLPattern);
    await expect(detailModal).not.toContainText("宿主服务申请清单");
    await expect(detailModal).not.toContainText("宿主服务授权快照");
    await expect(detailModal).not.toContainText("数据表边界");
    await expect(detailModal).not.toContainText("与申请清单一致");
    await expect(detailModal).not.toContainText("授权要求");
    await expect(pluginPage.pluginDetailDescriptionRow()).toBeVisible();
    await expect(pluginPage.pluginDetailDescriptionRow()).toContainText(
      pluginDescription,
    );
    await expect(detailModal).toContainText("查询评审摘要");
    await expect(detailModal).toContainText("公开健康检查");
    await expect(detailModal).toContainText(routeSummaryPath);
    await expect(detailModal).toContainText(routeHealthPath);
    await expect(detailModal).not.toContainText("审计日志回放");
    await expect(detailModal).not.toContainText(routeAuditPath);
    const detailHostServiceTitleTop = await detailModal
      .getByTestId("plugin-host-service-section-title")
      .evaluate((node) => node.getBoundingClientRect().top);
    const detailRouteTitleTop = await detailModal
      .getByTestId("plugin-route-section-title")
      .evaluate((node) => node.getBoundingClientRect().top);
    const detailHostServiceTitleFontWeight = await detailModal
      .getByTestId("plugin-host-service-section-title")
      .evaluate((node) => Number.parseInt(getComputedStyle(node).fontWeight, 10));
    const detailRouteTitleFontWeight = await detailModal
      .getByTestId("plugin-route-section-title")
      .evaluate((node) => Number.parseInt(getComputedStyle(node).fontWeight, 10));
    const detailHostServiceTitleFontSize = await detailModal
      .getByTestId("plugin-host-service-section-title")
      .evaluate((node) => getComputedStyle(node).fontSize);
    const detailRouteTitleFontSize = await detailModal
      .getByTestId("plugin-route-section-title")
      .evaluate((node) => getComputedStyle(node).fontSize);
    expect(detailHostServiceTitleTop).toBeLessThan(detailRouteTitleTop);
    expect(detailHostServiceTitleFontWeight).toBeGreaterThanOrEqual(600);
    expect(detailRouteTitleFontWeight).toBeGreaterThanOrEqual(600);
    expect(detailHostServiceTitleFontSize).toBe("15px");
    expect(detailRouteTitleFontSize).toBe("15px");
    await expect(pluginPage.pluginRouteReviewToggle()).toHaveText("展开");
    await pluginPage.pluginRouteReviewToggle().click();
    await expect(detailModal).toContainText("审计日志回放");
    await expect(detailModal).toContainText(routeAuditPath);
    await expect(detailModal).toContainText(routeAuditPermission);
    const effectiveScopeBackground = await detailModal
      .getByTestId("plugin-host-service-scope-label-storage-storage-effective")
      .evaluate((node) => getComputedStyle(node).backgroundColor);
    const summaryScopeBackground = await detailModal
      .getByTestId("plugin-host-service-summary-label-storage-storage-effective")
      .evaluate((node) => getComputedStyle(node).backgroundColor);
    expect(authEffectiveScopeBackground).toBe(effectiveScopeBackground);
    expect(authSummaryScopeBackground).toBe(summaryScopeBackground);
    expect(summaryScopeBackground).not.toBe(effectiveScopeBackground);
    const summaryTop = await detailModal
      .getByTestId("plugin-host-service-summary-label-storage-storage-effective")
      .evaluate((node) => node.getBoundingClientRect().top);
    const firstStorageItemTop = await detailModal
      .locator('.ant-tag', { hasText: storagePath })
      .first()
      .evaluate((node) => node.getBoundingClientRect().top);
    expect(Math.abs(summaryTop - firstStorageItemTop)).toBeLessThan(24);

    const rolePage = new RolePage(adminPage);
    await rolePage.goto();
    const roleDrawer = await rolePage.openCreateDrawer();
    await expect(roleDrawer).toContainText("动态路由权限（资源：审核，动作：查询）");
    await expect(roleDrawer).toContainText("动态路由权限（资源：审计，动作：查询）");
    await expect(roleDrawer).not.toContainText(routePermission);
    await expect(roleDrawer).not.toContainText(routeAuditPermission);
  });
});
