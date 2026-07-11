import type { Page, Route } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import { PluginPage } from "../../../pages/PluginPage";

const frameworkBlockedPluginID = "plugin-dev-dependency-framework-blocked-e2e";
const blockedPluginID = "plugin-dev-dependency-blocked-e2e";
const basePluginID = "plugin-dev-dependency-base-e2e";
const consumerPluginID = "plugin-dev-dependency-consumer-e2e";
const consumerPluginBID = "plugin-dev-dependency-consumer-b-e2e";
const installNetworkFailurePluginID =
  "plugin-dev-dependency-install-network-failure-e2e";
const uninstallNetworkFailurePluginID =
  "plugin-dev-dependency-uninstall-network-failure-e2e";

type PluginRow = Record<string, unknown>;
type DependencyCheck = Record<string, unknown>;

function apiEnvelope(data: unknown) {
  return {
    code: 0,
    data,
    message: "success",
  };
}

function pluginRow(input: {
  description: string;
  id: string;
  installed: number;
  name: string;
}): PluginRow {
  const requestedHostServices =
    input.installed === 1
      ? []
      : [
          {
            methods: ["get"],
            paths: [`dependency/${input.id}`],
            service: "storage",
          },
        ];
  return {
    authorizationRequired: requestedHostServices.length > 0 ? 1 : 0,
    authorizationStatus:
      requestedHostServices.length > 0 ? "pending" : "not_required",
    autoEnableForNewTenants: false,
    autoEnableManaged: 0,
    authorizedHostServices: [],
    declaredRoutes: [],
    dependencyCheck: null,
    description: input.description,
    enabled: 0,
    hasMockData: 0,
    id: input.id,
    installMode: "global",
    installed: input.installed,
    installedAt: "",
    name: input.name,
    requestedHostServices,
    scopeNature: "platform_only",
    statusKey: input.installed === 1 ? "disabled" : "not_installed",
    supportsMultiTenant: false,
    type: "source",
    updatedAt: "",
    version: "v0.1.0",
  };
}

function emptyDependencyCheck(pluginId: string): DependencyCheck {
  return {
    blockers: [],
    cycle: [],
    dependencies: [],
    framework: {
      currentVersion: "v0.6.0",
      requiredVersion: "",
      status: "not_declared",
    },
    reverseBlockers: [],
    reverseDependents: [],
    targetId: pluginId,
  };
}

function frameworkBlockerCheck(): DependencyCheck {
  return {
    ...emptyDependencyCheck(frameworkBlockedPluginID),
    framework: {
      currentVersion: "v0.6.0",
      requiredVersion: ">=0.7.0",
      status: "unsatisfied",
    },
  };
}

function installBlockerCheck(): DependencyCheck {
  return {
    ...emptyDependencyCheck(blockedPluginID),
    blockers: [
      {
        chain: [blockedPluginID, basePluginID],
        code: "dependency_version_unsatisfied",
        currentVersion: "v0.1.0",
        dependencyId: basePluginID,
        pluginId: blockedPluginID,
        requiredVersion: ">=0.3.0",
      },
    ],
    dependencies: [
      {
        chain: [blockedPluginID, basePluginID],
        currentVersion: "v0.1.0",
        dependencyId: basePluginID,
        dependencyName: "Dependency Base",
        discovered: true,
        installed: true,
        ownerId: blockedPluginID,
        requiredVersion: ">=0.3.0",
        status: "version_unsatisfied",
      },
    ],
  };
}

function reverseBlockerCheck(): DependencyCheck {
  return {
    ...emptyDependencyCheck(basePluginID),
    // reverse_dependency blockers mirror reverseDependents; UI must list each
    // consumer once and must not collapse them to the target plugin ID.
    reverseBlockers: [
      {
        chain: [consumerPluginID, basePluginID],
        code: "reverse_dependency",
        dependencyId: basePluginID,
        pluginId: consumerPluginID,
        requiredVersion: ">=0.1.0",
      },
      {
        chain: [consumerPluginBID, basePluginID],
        code: "reverse_dependency",
        dependencyId: basePluginID,
        pluginId: consumerPluginBID,
        requiredVersion: ">=0.1.0",
      },
    ],
    reverseDependents: [
      {
        name: "Consumer Plugin",
        pluginId: consumerPluginID,
        requiredVersion: ">=0.1.0",
        version: "v0.1.0",
      },
      {
        name: "Consumer Plugin B",
        pluginId: consumerPluginBID,
        requiredVersion: ">=0.1.0",
        version: "v0.1.0",
      },
    ],
  };
}

function reverseBlockerOnlyCheck(): DependencyCheck {
  return {
    ...emptyDependencyCheck(basePluginID),
    reverseBlockers: [
      {
        chain: [consumerPluginID, basePluginID],
        code: "dependency_snapshot_unknown",
        pluginId: consumerPluginID,
      },
    ],
    reverseDependents: [],
  };
}

async function mockPluginDependencyApis(
  page: Page,
  rows: PluginRow[],
  checks: Record<string, DependencyCheck>,
  failingDependencyPluginIds: string[] = [],
) {
  const failingPluginIdSet = new Set(failingDependencyPluginIds);
  await page.route("**/api/v1/plugins**", async (route: Route) => {
    const request = route.request();
    const url = new URL(request.url());
    const path = url.pathname;
    const detailMatch = path.match(/\/api\/v1\/plugins\/([^/]+)\/?$/u);
    const dependencyMatch = path.match(
      /\/api\/v1\/plugins\/([^/]+)\/dependencies$/u,
    );

    if (request.method() === "GET" && detailMatch) {
      const pluginId = decodeURIComponent(detailMatch[1] ?? "");
      const row = rows.find((item) => String(item.id ?? "") === pluginId);
      if (row) {
        await route.fulfill({
          json: apiEnvelope({
            ...row,
            dependencyCheck: checks[pluginId] ?? emptyDependencyCheck(pluginId),
          }),
        });
        return;
      }
    }

    if (request.method() === "GET" && dependencyMatch) {
      const pluginId = decodeURIComponent(dependencyMatch[1] ?? "");
      if (failingPluginIdSet.has(pluginId)) {
        await route.abort("failed");
        return;
      }
      await route.fulfill({
        json: apiEnvelope(checks[pluginId] ?? emptyDependencyCheck(pluginId)),
      });
      return;
    }

    if (request.method() === "GET" && /\/api\/v1\/plugins$/u.test(path)) {
      const id = url.searchParams.get("id")?.trim();
      const filteredRows = id
        ? rows.filter((row) => String(row.id ?? "").includes(id))
        : rows;
      await route.fulfill({
        json: apiEnvelope({
          list: filteredRows,
          total: filteredRows.length,
        }),
      });
      return;
    }

    await route.continue();
  });
}

test.describe("TC-6 插件依赖管理展示", () => {
  test("TC-6a: 安装确认展示框架版本阻断并禁用提交", async ({ adminPage }) => {
    await mockPluginDependencyApis(
      adminPage,
      [
        pluginRow({
          description:
            "Used by E2E to verify framework dependency blocker display.",
          id: frameworkBlockedPluginID,
          installed: 0,
          name: "Dependency Framework Blocked Plugin",
        }),
      ],
      { [frameworkBlockedPluginID]: frameworkBlockerCheck() },
    );

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(frameworkBlockedPluginID);
    await pluginPage.openInstallAuthorization(frameworkBlockedPluginID);

    await expect(pluginPage.pluginDependencyFrameworkBlocker()).toBeVisible();
    await expect(pluginPage.pluginDependencyFrameworkBlocker()).toContainText(
      "当前框架版本不符合该插件要求。",
    );
    await expect(pluginPage.pluginDependencyFrameworkBlocker()).toContainText(
      "要求版本：>=0.7.0；当前版本：v0.6.0。",
    );
    await expect(pluginPage.hostServiceAuthConfirmButton()).toBeDisabled();
    await expect(
      pluginPage.hostServiceAuthInstallAndEnableButton(),
    ).toBeDisabled();
  });

  test("TC-6b: 安装确认展示依赖阻断并禁用提交", async ({ adminPage }) => {
    await mockPluginDependencyApis(
      adminPage,
      [
        pluginRow({
          description: "Used by E2E to verify dependency blockers.",
          id: blockedPluginID,
          installed: 0,
          name: "Dependency Blocked Plugin",
        }),
      ],
      { [blockedPluginID]: installBlockerCheck() },
    );

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(blockedPluginID);
    await pluginPage.openInstallAuthorization(blockedPluginID);

    const dependencyBlockers = pluginPage.pluginDependencyBlockers();
    await expect(dependencyBlockers).toBeVisible();
    await expect(dependencyBlockers).toContainText(
      "依赖条件未满足，请根据下方提示处理后再安装。",
    );
    await expect(dependencyBlockers).toContainText("依赖版本不满足");
    await expect(dependencyBlockers).toContainText(basePluginID);
    // 依赖阻断提示是说明文案，不应使用 Alert 默认标题字号（16px）
    await expect
      .poll(async () => {
        return dependencyBlockers
          .locator(".ant-alert-message")
          .evaluate((el) => getComputedStyle(el).fontSize);
      })
      .toBe("14px");
    await expect(pluginPage.hostServiceAuthConfirmButton()).toBeDisabled();
    await expect(
      pluginPage.hostServiceAuthInstallAndEnableButton(),
    ).toBeDisabled();
  });

  test("TC-6c: 卸载确认展示反向依赖阻断并禁用提交", async ({ adminPage }) => {
    await mockPluginDependencyApis(
      adminPage,
      [
        pluginRow({
          description: "Used by E2E to verify reverse dependency blockers.",
          id: basePluginID,
          installed: 1,
          name: "Dependency Base",
        }),
      ],
      { [basePluginID]: reverseBlockerCheck() },
    );

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(basePluginID);
    await pluginPage.openUninstallDialog(basePluginID);

    const reverseBlockers = pluginPage.pluginDependencyReverseBlockers();
    await expect(reverseBlockers).toBeVisible();
    await expect(reverseBlockers).toContainText(
      "仍有其他已安装插件依赖此插件，请先处理后再卸载。",
    );
    // Distinct downstream consumers must appear once each by display name.
    await expect(reverseBlockers).toContainText("Consumer Plugin >=0.1.0");
    await expect(reverseBlockers).toContainText("Consumer Plugin B >=0.1.0");
    // reverse_dependency blockers share the target ID; must not collapse into
    // duplicate "存在下游依赖 <target> >=0.1.0" tags when dependents are listed.
    await expect(reverseBlockers).not.toContainText(
      `存在下游依赖 ${basePluginID}`,
    );
    const reverseTags = reverseBlockers.locator(".ant-tag");
    await expect(reverseTags).toHaveCount(2);
    await expect(pluginPage.uninstallConfirmButton()).toBeDisabled();
  });

  test("TC-6c2: 仅有 reverseBlockers 时展示下游插件 ID 而非目标插件 ID", async ({
    adminPage,
  }) => {
    await mockPluginDependencyApis(
      adminPage,
      [
        pluginRow({
          description:
            "Used by E2E to verify reverse blocker subject formatting.",
          id: basePluginID,
          installed: 1,
          name: "Dependency Base",
        }),
      ],
      { [basePluginID]: reverseBlockerOnlyCheck() },
    );

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(basePluginID);
    await pluginPage.openUninstallDialog(basePluginID);

    const reverseBlockers = pluginPage.pluginDependencyReverseBlockers();
    await expect(reverseBlockers).toBeVisible();
    await expect(reverseBlockers).toContainText("依赖快照不可用");
    await expect(reverseBlockers).toContainText(consumerPluginID);
    await expect(reverseBlockers).not.toContainText(
      new RegExp(`依赖快照不可用\\s+${basePluginID}`),
    );
    await expect(pluginPage.uninstallConfirmButton()).toBeDisabled();
  });

  test("TC-6d: 安装弹窗依赖检查网络失败时只显示本地刷新失败提示", async ({
    adminPage,
  }) => {
    await mockPluginDependencyApis(
      adminPage,
      [
        pluginRow({
          description:
            "Used by E2E to verify dependency failure toast handling.",
          id: installNetworkFailurePluginID,
          installed: 0,
          name: "Dependency Install Network Failure Plugin",
        }),
      ],
      {},
      [installNetworkFailurePluginID],
    );

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(installNetworkFailurePluginID);
    await pluginPage.openInstallAuthorization(installNetworkFailurePluginID);

    await expect(
      pluginPage.messageNotice("刷新插件依赖检查结果失败"),
    ).toBeVisible();
    await expect(
      pluginPage.messageNotice("网络异常，请检查您的网络连接后重试。"),
    ).toHaveCount(0);
  });

  test("TC-6e: 卸载弹窗依赖检查网络失败时只显示本地刷新失败提示", async ({
    adminPage,
  }) => {
    await mockPluginDependencyApis(
      adminPage,
      [
        pluginRow({
          description:
            "Used by E2E to verify dependency failure toast handling.",
          id: uninstallNetworkFailurePluginID,
          installed: 1,
          name: "Dependency Uninstall Network Failure Plugin",
        }),
      ],
      {},
      [uninstallNetworkFailurePluginID],
    );

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(uninstallNetworkFailurePluginID);
    await pluginPage.openUninstallDialog(uninstallNetworkFailurePluginID);

    await expect(
      pluginPage.messageNotice("刷新插件依赖检查结果失败"),
    ).toBeVisible();
    await expect(
      pluginPage.messageNotice("网络异常，请检查您的网络连接后重试。"),
    ).toHaveCount(0);
  });
});
