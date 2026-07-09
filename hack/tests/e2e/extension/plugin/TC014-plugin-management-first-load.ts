import { mkdir } from "node:fs/promises";
import path from "node:path";

import type { Page, Route } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import { PluginPage } from "../../../pages/PluginPage";

const pluginID = "plugin-management-first-load-e2e";

function apiEnvelope(data: unknown) {
  return {
    code: 0,
    data,
    message: "success",
  };
}

function summaryPluginListRow() {
  return {
    abnormalReason: "",
    authorizationRequired: 1,
    authorizationStatus: "pending",
    autoEnableForNewTenants: false,
    autoEnableManaged: 0,
    description: "Summary projection excludes governance payload.",
    discoveredVersion: "v0.1.0",
    effectiveVersion: "v0.1.0",
    enabled: 0,
    hasMockData: 0,
    id: pluginID,
    installMode: "tenant_scoped",
    installed: 0,
    installedAt: 1767240000000,
    lastUpgradeFailure: undefined,
    name: "First Load E2E",
    runtimeState: "normal",
    scopeNature: "tenant_aware",
    statusKey: "disabled",
    supportsMultiTenant: true,
    type: "dynamic",
    updatedAt: 1767240000000,
    upgradeAvailable: false,
    version: "v0.1.0",
  };
}

function completePluginDetail() {
  return {
    ...summaryPluginListRow(),
    description: "Complete detail projection includes governance payload.",
    declaredRoutes: [
      {
        access: "authenticated",
        description: "Route detail returned by the detail projection.",
        method: "GET",
        permission: `${pluginID}:report:query`,
        publicPath: "/governed-report",
        summary: "Governed report",
      },
    ],
    dependencyCheck: emptyDependencyCheck(),
    requestedHostServices: [
      {
        methods: ["get"],
        paths: ["reports/"],
        service: "storage",
      },
    ],
    authorizedHostServices: [],
  };
}

function emptyDependencyCheck() {
  return {
    blockers: [],
    cycle: [],
    dependencies: [],
    framework: {
      currentVersion: "v0.1.0",
      requiredVersion: "",
      status: "not_declared",
    },
    reverseBlockers: [],
    reverseDependents: [],
    targetId: pluginID,
  };
}

async function captureEvidence(page: Page, name: string) {
  await page.waitForTimeout(300);
  const now = new Date();
  const day = new Intl.DateTimeFormat("en-CA", {
    day: "2-digit",
    month: "2-digit",
    timeZone: "Asia/Shanghai",
    year: "numeric",
  })
    .format(now)
    .replaceAll("-", "");
  const time = new Intl.DateTimeFormat("en-GB", {
    hour: "2-digit",
    hour12: false,
    minute: "2-digit",
    second: "2-digit",
    timeZone: "Asia/Shanghai",
  })
    .format(now)
    .replaceAll(":", "");
  const dir = path.resolve(process.cwd(), "..", "..", "temp", day);
  await mkdir(dir, { recursive: true });
  await page.screenshot({
    fullPage: true,
    path: path.join(dir, `${time}-${name}.png`),
  });
}

async function expectNoRawPluginI18nKeys(page: Page) {
  await expect(page.locator("body")).not.toContainText(/pages\.system\.plugin/u);
}

async function mockPluginManagementApis(page: Page) {
  let dependencyRequestCount = 0;
  let detailRequestCount = 0;
  let listRequestCount = 0;

  await page.route("**/api/v1/plugins**", async (route: Route) => {
    const request = route.request();
    const url = new URL(request.url());
    const path = url.pathname;

    if (request.method() === "GET" && /\/api\/v1\/plugins$/u.test(path)) {
      listRequestCount += 1;
      const id = url.searchParams.get("id")?.trim();
      const rows = id && !pluginID.includes(id) ? [] : [summaryPluginListRow()];
      await route.fulfill({
        json: apiEnvelope({
          list: rows,
          total: rows.length,
        }),
      });
      return;
    }

    if (request.method() === "GET" && path.endsWith("/plugins/dynamic")) {
      await route.fulfill({
        json: apiEnvelope({
          list: [],
        }),
      });
      return;
    }

    if (
      request.method() === "GET" &&
      path.endsWith(`/plugins/${pluginID}`)
    ) {
      detailRequestCount += 1;
      await route.fulfill({
        json: apiEnvelope(completePluginDetail()),
      });
      return;
    }

    if (
      request.method() === "GET" &&
      path.endsWith(`/plugins/${pluginID}/dependencies`)
    ) {
      dependencyRequestCount += 1;
      await route.fulfill({
        json: apiEnvelope(emptyDependencyCheck()),
      });
      return;
    }

    await route.continue();
  });

  return {
    dependencyRequestCount: () => dependencyRequestCount,
    detailRequestCount: () => detailRequestCount,
    listRequestCount: () => listRequestCount,
  };
}

test.describe("TC-14 插件管理首次加载优化", () => {
  test("TC-14a: 首屏只请求分页摘要列表，详情弹窗按需加载治理字段", async ({
    adminPage,
  }) => {
    const pageErrors: string[] = [];
    adminPage.on("pageerror", (error) => pageErrors.push(error.message));
    const api = await mockPluginManagementApis(adminPage);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);

    await expect(pluginPage.pluginRow(pluginID)).toBeVisible();
    await expect(pluginPage.pluginNameCell(pluginID)).toContainText(
      "First Load E2E",
    );
    await expect(
      pluginPage.pluginRuntimeState(pluginID),
    ).toContainText(/正常|Normal/iu);
    await expect(pluginPage.tableHeaderCell("插件类型")).toBeVisible();
    await expect(pluginPage.tableHeaderCell("安装时间")).toBeVisible();
    await expect(
      adminPage.getByTestId(`plugin-detail-button-${pluginID}`).last(),
    ).toBeVisible();
    expect(api.listRequestCount()).toBeGreaterThan(0);
    expect(api.detailRequestCount()).toBe(0);
    expect(api.dependencyRequestCount()).toBe(0);
    await expectNoRawPluginI18nKeys(adminPage);
    await captureEvidence(adminPage, "plugin-management-first-load");
    expect(pageErrors).toEqual([]);

    await pluginPage.openPluginDetail(pluginID);

    await expect(pluginPage.pluginDetailModal()).toContainText(
      "Complete detail projection includes governance payload.",
    );
    await expect(pluginPage.pluginDetailModal()).toContainText("reports/");
    await expect(
      adminPage.getByTestId("plugin-route-review-list").last(),
    ).toContainText("/governed-report");
    expect(api.detailRequestCount()).toBe(1);
    expect(api.dependencyRequestCount()).toBe(0);
    await expectNoRawPluginI18nKeys(adminPage);
    await captureEvidence(adminPage, "plugin-management-detail-modal");
    expect(pageErrors).toEqual([]);
  });

  test("TC-14b: 安装授权弹窗打开后才请求详情和依赖检查", async ({
    adminPage,
  }) => {
    const pageErrors: string[] = [];
    adminPage.on("pageerror", (error) => pageErrors.push(error.message));
    const api = await mockPluginManagementApis(adminPage);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);
    expect(api.detailRequestCount()).toBe(0);
    expect(api.dependencyRequestCount()).toBe(0);

    await pluginPage.openInstallAuthorization(pluginID);

    await expect(pluginPage.hostServiceAuthModal()).toContainText("reports/");
    await expect(
      adminPage
        .getByTestId(
          `plugin-host-service-auth-list-${pluginID}-storage`,
        )
        .last(),
    ).toContainText("reports/");
    await expect(
      adminPage.getByTestId("plugin-route-review-list").last(),
    ).toContainText("/governed-report");
    expect(api.detailRequestCount()).toBe(1);
    expect(api.dependencyRequestCount()).toBe(1);
    await expectNoRawPluginI18nKeys(adminPage);
    await captureEvidence(adminPage, "plugin-management-install-auth-modal");
    expect(pageErrors).toEqual([]);
  });
});
