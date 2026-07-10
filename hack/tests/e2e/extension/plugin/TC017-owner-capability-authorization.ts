import { mkdir } from "node:fs/promises";
import path from "node:path";

import type { Locator, Page, Route } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import { PluginPage } from "../../../pages/PluginPage";

const ownerPluginID = "plugin-dev-owner-ai-core-e2e";
const consumerPluginID = "plugin-dev-owner-ai-consumer-e2e";
const missingDependencyPluginID = "plugin-dev-owner-ai-missing-dependency-e2e";
const versionBlockedPluginID = "plugin-dev-owner-ai-version-blocked-e2e";
const ownerCapabilityResourceRef = "purpose:content.summary";

type DependencyCheck = Record<string, unknown>;
type PluginRow = Record<string, unknown>;

function apiEnvelope(data: unknown) {
  return {
    code: 0,
    data,
    message: "success",
  };
}

function ownerAIHostService() {
  return {
    methods: ["text.generate", "text.method_status.get"],
    owner: ownerPluginID,
    resources: [
      {
        attributes: { purpose: "content.summary" },
        ref: ownerCapabilityResourceRef,
      },
    ],
    service: "ai",
    version: "v1",
  };
}

function pluginRow(input: {
  description: string;
  id: string;
  name: string;
}): PluginRow {
  return {
    abnormalReason: "",
    authorizationRequired: 1,
    authorizationStatus: "pending",
    authorizedHostServices: [],
    autoEnableForNewTenants: false,
    autoEnableManaged: 0,
    declaredRoutes: [],
    dependencyCheck: null,
    description: input.description,
    discoveredVersion: "v0.1.0",
    effectiveVersion: "v0.1.0",
    enabled: 0,
    hasMockData: 0,
    id: input.id,
    installMode: "global",
    installed: 0,
    installedAt: null,
    name: input.name,
    requestedHostServices: [ownerAIHostService()],
    runtimeState: "normal",
    scopeNature: "platform_only",
    statusKey: "not_installed",
    supportsMultiTenant: false,
    type: "dynamic",
    updatedAt: null,
    upgradeAvailable: false,
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

function satisfiedOwnerDependencyCheck(pluginId: string): DependencyCheck {
  return {
    ...emptyDependencyCheck(pluginId),
    dependencies: [
      {
        chain: [pluginId, ownerPluginID],
        currentVersion: "v0.1.0",
        dependencyId: ownerPluginID,
        dependencyName: "LinaPro AI Core",
        discovered: true,
        installed: true,
        ownerId: pluginId,
        requiredVersion: ">=0.1.0",
        status: "satisfied",
      },
    ],
  };
}

function missingOwnerDependencyCheck(): DependencyCheck {
  return {
    ...emptyDependencyCheck(missingDependencyPluginID),
    blockers: [
      {
        chain: [missingDependencyPluginID, ownerPluginID],
        code: "dependency_missing",
        dependencyId: ownerPluginID,
        pluginId: missingDependencyPluginID,
        requiredVersion: ">=0.1.0",
      },
    ],
    dependencies: [
      {
        chain: [missingDependencyPluginID, ownerPluginID],
        currentVersion: "",
        dependencyId: ownerPluginID,
        dependencyName: "LinaPro AI Core",
        discovered: false,
        installed: false,
        ownerId: missingDependencyPluginID,
        requiredVersion: ">=0.1.0",
        status: "missing",
      },
    ],
  };
}

function versionUnsatisfiedOwnerDependencyCheck(): DependencyCheck {
  return {
    ...emptyDependencyCheck(versionBlockedPluginID),
    blockers: [
      {
        chain: [versionBlockedPluginID, ownerPluginID],
        code: "dependency_version_unsatisfied",
        currentVersion: "v0.1.0",
        dependencyId: ownerPluginID,
        pluginId: versionBlockedPluginID,
        requiredVersion: ">=0.2.0",
      },
    ],
    dependencies: [
      {
        chain: [versionBlockedPluginID, ownerPluginID],
        currentVersion: "v0.1.0",
        dependencyId: ownerPluginID,
        dependencyName: "LinaPro AI Core",
        discovered: true,
        installed: true,
        ownerId: versionBlockedPluginID,
        requiredVersion: ">=0.2.0",
        status: "version_unsatisfied",
      },
    ],
  };
}

async function mockPluginApis(
  page: Page,
  rows: PluginRow[],
  checks: Record<string, DependencyCheck>,
) {
  const installPayloads = new Map<string, unknown>();

  await page.route("**/api/v1/plugins**", async (route: Route) => {
    const request = route.request();
    const url = new URL(request.url());
    const requestPath = url.pathname;
    const detailMatch = requestPath.match(/\/api\/v1\/plugins\/([^/]+)\/?$/u);
    const dependencyMatch = requestPath.match(
      /\/api\/v1\/plugins\/([^/]+)\/dependencies$/u,
    );
    const installMatch = requestPath.match(
      /\/api\/v1\/plugins\/([^/]+)\/install$/u,
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
      await route.fulfill({
        json: apiEnvelope(checks[pluginId] ?? emptyDependencyCheck(pluginId)),
      });
      return;
    }

    if (request.method() === "POST" && installMatch) {
      const pluginId = decodeURIComponent(installMatch[1] ?? "");
      installPayloads.set(pluginId, parseRequestBody(request.postData()));
      await route.fulfill({
        json: apiEnvelope({
          dependencyCheck: checks[pluginId] ?? emptyDependencyCheck(pluginId),
          enabled: 0,
          id: pluginId,
          installed: 1,
        }),
      });
      return;
    }

    if (
      request.method() === "GET" &&
      /\/api\/v1\/plugins$/u.test(requestPath)
    ) {
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

  return {
    installPayload(pluginId: string) {
      return installPayloads.get(pluginId);
    },
  };
}

function parseRequestBody(body: null | string) {
  if (!body) {
    return {};
  }
  return JSON.parse(body) as unknown;
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
    fullPage: false,
    path: path.join(dir, `${time}-${name}.png`),
  });
}

async function expectNoRawPluginI18nKeys(target: Locator) {
  await expect(target).not.toContainText(/pages\.system\.plugin/u);
}

test.describe("TC-17 owner 能力授权与依赖阻断", () => {
  test("TC-17a: owner 授权来源展示并提交 owner-aware 授权身份", async ({
    adminPage,
  }) => {
    const controls = await mockPluginApis(
      adminPage,
      [
        pluginRow({
          description:
            "Used by E2E to verify owner capability authorization display.",
          id: consumerPluginID,
          name: "Owner AI Consumer",
        }),
      ],
      { [consumerPluginID]: satisfiedOwnerDependencyCheck(consumerPluginID) },
    );

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(consumerPluginID);
    await pluginPage.openInstallAuthorization(consumerPluginID);

    const modal = pluginPage.hostServiceAuthModal();
    await expect(modal).toContainText("宿主服务授权审核");
    await expect(modal).toContainText("申请范围");
    await expect(modal).toContainText("资源");
    await expect(modal).toContainText("ai");
    await expect(modal).toContainText("来源插件");
    await expect(modal).toContainText(ownerPluginID);
    await expect(modal).toContainText("能力版本");
    await expect(modal).toContainText("v1");
    await expect(modal).toContainText("text.generate");
    await expect(modal).toContainText("text.method_status.get");
    await expect(modal).toContainText(ownerCapabilityResourceRef);
    await expect(
      modal.getByTestId(`plugin-host-service-owner-${ownerPluginID}-ai-v1`),
    ).toContainText(ownerPluginID);
    await expect(
      modal.getByTestId(`plugin-host-service-version-${ownerPluginID}-ai-v1`),
    ).toContainText("v1");
    await expectNoRawPluginI18nKeys(modal);
    await captureEvidence(adminPage, "tc017-owner-authorization");

    await pluginPage.confirmHostServiceAuthorization();

    const payload = controls.installPayload(consumerPluginID) as {
      authorization?: {
        services?: Array<Record<string, unknown>>;
      };
    };
    expect(payload?.authorization?.services).toHaveLength(1);
    expect(payload.authorization?.services?.[0]).toMatchObject({
      methods: ["text.generate", "text.method_status.get"],
      owner: ownerPluginID,
      resourceRefs: [ownerCapabilityResourceRef],
      service: "ai",
      version: "v1",
    });
  });

  test("TC-17b: 缺失 owner 插件依赖时阻断安装提交", async ({ adminPage }) => {
    await mockPluginApis(
      adminPage,
      [
        pluginRow({
          description:
            "Used by E2E to verify missing owner dependency blockers.",
          id: missingDependencyPluginID,
          name: "Owner AI Missing Dependency Consumer",
        }),
      ],
      { [missingDependencyPluginID]: missingOwnerDependencyCheck() },
    );

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(missingDependencyPluginID);
    await pluginPage.openInstallAuthorization(missingDependencyPluginID);

    await expect(pluginPage.pluginDependencyBlockers()).toBeVisible();
    await expect(pluginPage.pluginDependencyBlockers()).toContainText(
      "请先处理依赖阻断项",
    );
    await expect(pluginPage.pluginDependencyBlockers()).toContainText(
      "依赖插件缺失",
    );
    await expect(pluginPage.pluginDependencyBlockers()).toContainText(
      ownerPluginID,
    );
    await expect(pluginPage.pluginDependencyBlockers()).toContainText(
      ">=0.1.0",
    );
    await expect(pluginPage.hostServiceAuthConfirmButton()).toBeDisabled();
    await expect(
      pluginPage.hostServiceAuthInstallAndEnableButton(),
    ).toBeDisabled();
    await expectNoRawPluginI18nKeys(pluginPage.hostServiceAuthModal());
    await captureEvidence(adminPage, "tc017-missing-owner-dependency");
  });

  test("TC-17c: owner 插件版本不满足时展示阻断提示", async ({ adminPage }) => {
    await mockPluginApis(
      adminPage,
      [
        pluginRow({
          description:
            "Used by E2E to verify owner dependency version blockers.",
          id: versionBlockedPluginID,
          name: "Owner AI Version Blocked Consumer",
        }),
      ],
      { [versionBlockedPluginID]: versionUnsatisfiedOwnerDependencyCheck() },
    );

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(versionBlockedPluginID);
    await pluginPage.openInstallAuthorization(versionBlockedPluginID);

    await expect(pluginPage.pluginDependencyBlockers()).toBeVisible();
    await expect(pluginPage.pluginDependencyBlockers()).toContainText(
      "请先处理依赖阻断项",
    );
    await expect(pluginPage.pluginDependencyBlockers()).toContainText(
      "依赖版本不满足",
    );
    await expect(pluginPage.pluginDependencyBlockers()).toContainText(
      ownerPluginID,
    );
    await expect(pluginPage.pluginDependencyBlockers()).toContainText(
      ">=0.2.0",
    );
    await expect(pluginPage.hostServiceAuthConfirmButton()).toBeDisabled();
    await expect(
      pluginPage.hostServiceAuthInstallAndEnableButton(),
    ).toBeDisabled();
    await expectNoRawPluginI18nKeys(pluginPage.hostServiceAuthModal());
    await captureEvidence(adminPage, "tc017-owner-version-unsatisfied");
  });
});
