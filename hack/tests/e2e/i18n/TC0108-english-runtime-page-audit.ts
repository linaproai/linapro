import type { APIRequestContext } from "@playwright/test";

import { execFileSync } from "node:child_process";
import { rmSync } from "node:fs";
import path from "node:path";

import { test, expect } from "../../fixtures/auth";
import { ensureSourcePluginEnabled } from "../../fixtures/plugin";
import { NoticePage } from "../../pages/NoticePage";
import { PluginPage } from "../../pages/PluginPage";
import {
  createAdminApiContext,
  disablePlugin,
  enablePlugin,
  getPlugin,
  installPlugin,
  listPlugins,
  syncPlugins,
  uninstallPlugin,
} from "../../support/api/job";
import { waitForRouteReady } from "../../support/ui";

const pluginID = "plugin-demo-dynamic";
const repoRoot = path.resolve(process.cwd(), "../..");
const legacyRuntimeArtifactPath = path.join(
  repoRoot,
  "apps",
  "lina-plugins",
  pluginID,
  "runtime",
  `${pluginID}.wasm`,
);

let adminApi: APIRequestContext;
let originalInstalled = 0;
let originalEnabled = 0;

function ensureRuntimePluginArtifact() {
  execFileSync("make", ["wasm", `p=${pluginID}`, "out=../../temp/output"], {
    cwd: path.join(repoRoot, "apps", "lina-plugins"),
    stdio: "inherit",
  });
  rmSync(legacyRuntimeArtifactPath, { force: true });
}

async function ensurePluginInstalledAndEnabled() {
  const plugin = await getPlugin(adminApi, pluginID);
  if (plugin.installed !== 1) {
    await installPlugin(adminApi, pluginID);
  }

  const refreshedPlugin = await getPlugin(adminApi, pluginID);
  if (refreshedPlugin.enabled !== 1) {
    await enablePlugin(adminApi, pluginID);
  }
}

async function restorePluginState() {
  let plugin = await getPlugin(adminApi, pluginID);

  if (originalInstalled !== 1) {
    if (plugin.enabled === 1) {
      await disablePlugin(adminApi, pluginID);
      plugin = await getPlugin(adminApi, pluginID);
    }
    if (plugin.installed === 1) {
      await uninstallPlugin(adminApi, pluginID);
    }
    return;
  }

  if (originalEnabled !== 1 && plugin.enabled === 1) {
    await disablePlugin(adminApi, pluginID);
  }
}

test.describe("TC0108 英文运行时页面巡检", () => {
  test.beforeAll(async () => {
    ensureRuntimePluginArtifact();
    adminApi = await createAdminApiContext();
    await syncPlugins(adminApi);
    const plugin = await getPlugin(adminApi, pluginID);
    originalInstalled = plugin.installed;
    originalEnabled = plugin.enabled;
  });

  test.afterAll(async () => {
    try {
      await restorePluginState();
    } finally {
      await adminApi.dispose();
    }
  });

  test("TC-108a: 英文环境下服务监控与操作日志运行时文案保持英文", async ({
    adminPage,
    mainLayout,
  }) => {
    test.setTimeout(120_000);
    await ensureSourcePluginEnabled(adminPage, "monitor-server");
    await ensureSourcePluginEnabled(adminPage, "monitor-operlog");
    await ensurePluginInstalledAndEnabled();

    // Toggle the plugin lifecycle to create fresh operation logs for this test.
    await disablePlugin(adminApi, pluginID);
    await enablePlugin(adminApi, pluginID);

    await mainLayout.switchLanguage("English");

    const serverResponsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes("/api/v1/monitor/server") &&
        res.request().method() === "GET" &&
        res.status() === 200,
      { timeout: 20_000 },
    );
    await adminPage.goto("/monitor/server");
    await serverResponsePromise;
    await waitForRouteReady(adminPage);

    const serviceUptime = adminPage
      .locator("dt", { hasText: "Service Uptime" })
      .locator("..")
      .locator("dd")
      .first();
    await expect(serviceUptime).toBeVisible({ timeout: 20_000 });
    const serviceUptimeValue = await serviceUptime.innerText();
    expect(serviceUptimeValue.trim()).not.toBe("");
    expect(serviceUptimeValue).not.toContain("刚启动");
    expect(/[\u3400-\u9fff]/.test(serviceUptimeValue)).toBeFalsy();

    await adminPage.goto("/monitor/operlog");
    await waitForRouteReady(adminPage);

    await expect
      .poll(async () => await adminPage.locator("body").innerText())
      .toMatch(/Enable Plugin/i);

    const operlogBodyText = await adminPage.locator("body").innerText();
    expect(operlogBodyText).toContain("Plugin Management");
    expect(operlogBodyText).toMatch(/Enable Plugin/i);
    expect(operlogBodyText).not.toContain("启用插件");
    expect(operlogBodyText).not.toContain("禁用插件");
  });

  test("TC-108b: 英文环境下未安装动态插件列表元数据保持英文", async ({
    adminPage,
    mainLayout,
  }) => {
    let plugin = await getPlugin(adminApi, pluginID);
    if (plugin.enabled === 1) {
      await disablePlugin(adminApi, pluginID);
      plugin = await getPlugin(adminApi, pluginID);
    }
    if (plugin.installed === 1) {
      await uninstallPlugin(adminApi, pluginID);
    }

    await mainLayout.switchLanguage("English");

    const apiList = await listPlugins(adminApi, pluginID, "en-US");
    const apiPlugin = apiList.list.find((item) => item.id === pluginID);
    expect(apiPlugin).toBeTruthy();
    expect(apiPlugin?.installed).toBe(0);
    expect(apiPlugin?.name).toBe("Dynamic Plugin Demo");
    expect(apiPlugin?.description).toBe(
      "Dynamic wasm sample that demonstrates a host-embedded menu page, plugin-owned SQL CRUD, and a hosted standalone page.",
    );

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);

    await expect(pluginPage.pluginRow(pluginID)).toContainText(
      "Dynamic Plugin Demo",
    );
    await expect(pluginPage.pluginDescriptionCell(pluginID)).toContainText(
      "Dynamic wasm sample that demonstrates a host-embedded menu page, plugin-owned SQL CRUD, and a hosted standalone page.",
    );
    const rowText = await pluginPage.pluginRow(pluginID).innerText();
    expect(rowText).not.toContain("动态插件示例");
    expect(rowText).not.toContain("提供独立的 dynamic wasm 插件样例");
  });

  test("TC-108c: 英文环境下运行时字典标签保持英文而可编辑标题仍可保留原值", async ({
    adminPage,
    mainLayout,
  }) => {
    await ensureSourcePluginEnabled(adminPage, "content-notice");

    const noticePage = new NoticePage(adminPage);

    await mainLayout.switchLanguage("English");

    await noticePage.goto();

    await expect(
      adminPage.getByText("Notice", { exact: true }).first(),
    ).toBeVisible();
    await expect(
      adminPage.getByText("Announcement", { exact: true }).first(),
    ).toBeVisible();
    await expect(
      adminPage.getByText("Draft", { exact: true }).first(),
    ).toBeVisible();
    await expect(
      adminPage.getByText("Published", { exact: true }).first(),
    ).toBeVisible();

    await expect(await noticePage.hasNotice("系统升级通知")).toBe(true);
  });

  test("TC-108d: 英文环境下动态插件页面与独立页种子内容保持英文", async ({
    adminPage,
    mainLayout,
  }) => {
    await ensurePluginInstalledAndEnabled();
    await adminPage.reload({ waitUntil: "domcontentloaded" });
    await waitForRouteReady(adminPage);

    const pluginPage = new PluginPage(adminPage);

    await mainLayout.switchLanguage("English");
    await pluginPage.clickSidebarMenuItem("Dynamic Plugin Demo");
    await waitForRouteReady(adminPage);

    await expect(pluginPage.pluginDemoDynamicTitle()).toHaveText(
      "Dynamic Plugin Demo Is Live",
    );
    await expect(
      pluginPage.pluginDemoDynamicRecordRow("Dynamic Plugin SQL Demo Record"),
    ).toBeVisible();
    await expect(
      adminPage.getByText(
        "This record is seeded by the plugin-demo-dynamic install SQL",
        { exact: false },
      ),
    ).toBeVisible();

    const [standalonePage] = await Promise.all([
      adminPage.context().waitForEvent("page"),
      pluginPage.pluginDemoDynamicOpenStandaloneButton().click(),
    ]);
    await standalonePage.waitForLoadState("domcontentloaded");
    await standalonePage.waitForLoadState("networkidle").catch(() => {});

    await expect.poll(async () => standalonePage.url()).toContain("lang=en-US");
    await expect(
      standalonePage.getByText("Dynamic Plugin Standalone Page", {
        exact: true,
      }),
    ).toBeVisible();
    await expect(
      standalonePage.getByText(
        "This page is served directly by plugin-demo-dynamic as a hosted static asset",
        { exact: false },
      ),
    ).toBeVisible();
    await standalonePage.close();
  });
});
