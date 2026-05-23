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

function completePluginListRow() {
  return {
    abnormalReason: "",
    authorizationRequired: 1,
    authorizationStatus: "pending",
    autoEnableForNewTenants: false,
    autoEnableManaged: 0,
    description: "Complete list projection includes governance payload.",
    discoveredVersion: "v0.1.0",
    effectiveVersion: "v0.1.0",
    enabled: 0,
    hasMockData: 0,
    id: pluginID,
    installMode: "tenant_scoped",
    installed: 0,
    installedAt: null,
    lastUpgradeFailure: undefined,
    name: "Plugin Management First Load E2E",
    runtimeState: "normal",
    scopeNature: "tenant_aware",
    statusKey: "disabled",
    supportsMultiTenant: true,
    type: "dynamic",
    updatedAt: null,
    upgradeAvailable: false,
    version: "v0.1.0",
    declaredRoutes: [
      {
        access: "authenticated",
        description: "Route detail returned by the full list projection.",
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
    autoInstallPlan: [],
    autoInstalled: [],
    blockers: [],
    cycle: [],
    dependencies: [],
    framework: {
      currentVersion: "v0.1.0",
      requiredVersion: "",
      status: "not_declared",
    },
    manualInstallRequired: [],
    reverseBlockers: [],
    reverseDependents: [],
    softUnsatisfied: [],
    targetId: pluginID,
  };
}

async function mockPluginManagementApis(page: Page) {
  let detailRequestCount = 0;

  await page.route("**/api/v1/plugins**", async (route: Route) => {
    const request = route.request();
    const url = new URL(request.url());
    const path = url.pathname;

    if (request.method() === "GET" && /\/api\/v1\/plugins$/u.test(path)) {
      const id = url.searchParams.get("id")?.trim();
      const rows = id && !pluginID.includes(id) ? [] : [completePluginListRow()];
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
        json: apiEnvelope(completePluginListRow()),
      });
      return;
    }

    if (
      request.method() === "GET" &&
      path.endsWith(`/plugins/${pluginID}/dependencies`)
    ) {
      await route.fulfill({
        json: apiEnvelope(emptyDependencyCheck()),
      });
      return;
    }

    await route.continue();
  });

  return {
    detailRequestCount: () => detailRequestCount,
  };
}

test.describe("TC-14 插件管理首次加载优化", () => {
  test("TC-14a: 完整列表投影已包含治理字段并可直接渲染详情", async ({
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
      "Plugin Management First Load E2E",
    );
    await expect(
      pluginPage.pluginRuntimeState(pluginID),
    ).toContainText(/正常|Normal/iu);
    expect(api.detailRequestCount()).toBe(0);
    expect(pageErrors).toEqual([]);

    await pluginPage.openPluginDetail(pluginID);

    await expect(pluginPage.pluginDetailModal()).toContainText(
      "Complete list projection includes governance payload.",
    );
    await expect(pluginPage.pluginDetailModal()).toContainText("reports/");
    await expect(
      adminPage.getByTestId("plugin-route-review-list").last(),
    ).toContainText("/governed-report");
    expect(api.detailRequestCount()).toBe(0);
    expect(pageErrors).toEqual([]);
  });

  test("TC-14b: 安装授权弹窗直接复用列表治理字段展示授权范围", async ({
    adminPage,
  }) => {
    const api = await mockPluginManagementApis(adminPage);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);
    expect(api.detailRequestCount()).toBe(0);

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
    expect(api.detailRequestCount()).toBe(0);
  });
});
