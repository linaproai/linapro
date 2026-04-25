import type { APIRequestContext } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";
import { NoticePage } from "../../pages/NoticePage";
import { PluginPage } from "../../pages/PluginPage";
import {
  createAdminApiContext,
  disablePlugin,
  enablePlugin,
  getPlugin,
  installPlugin,
  syncPlugins,
  uninstallPlugin,
} from "../../support/api/job";
import { waitForRouteReady } from "../../support/ui";

const pluginID = "plugin-demo-dynamic";

let adminApi: APIRequestContext;
let originalInstalled = 0;
let originalEnabled = 0;

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
    await ensurePluginInstalledAndEnabled();

    // Toggle the plugin lifecycle to create fresh operation logs for this test.
    await disablePlugin(adminApi, pluginID);
    await enablePlugin(adminApi, pluginID);

    await mainLayout.switchLanguage("English");

    await adminPage.goto("/monitor/server");
    await waitForRouteReady(adminPage);

    const serviceUptimeValue = await adminPage
      .locator("dt", { hasText: "Service Uptime" })
      .locator("..")
      .locator("dd")
      .first()
      .innerText();
    expect(serviceUptimeValue.trim()).not.toBe("");
    expect(serviceUptimeValue).not.toContain("刚启动");
    expect(/[\u3400-\u9fff]/.test(serviceUptimeValue)).toBeFalsy();

    await adminPage.goto("/monitor/operlog");
    await waitForRouteReady(adminPage);

    await expect
      .poll(async () => await adminPage.locator("body").innerText())
      .toContain("Enable Plugin");

    const operlogBodyText = await adminPage.locator("body").innerText();
    expect(operlogBodyText).toContain("Plugin Management");
    expect(operlogBodyText).toContain("Enable Plugin");
    expect(operlogBodyText).not.toContain("启用插件");
    expect(operlogBodyText).not.toContain("禁用插件");
  });

  test("TC-108b: 英文环境下运行时字典标签保持英文而可编辑标题仍可保留原值", async ({
    adminPage,
    mainLayout,
  }) => {
    const noticePage = new NoticePage(adminPage);

    await mainLayout.switchLanguage("English");

    await noticePage.goto();

    await expect(adminPage.getByText("Notice", { exact: true }).first()).toBeVisible();
    await expect(adminPage.getByText("Announcement", { exact: true }).first()).toBeVisible();
    await expect(adminPage.getByText("Draft", { exact: true }).first()).toBeVisible();
    await expect(adminPage.getByText("Published", { exact: true }).first()).toBeVisible();

    await expect(await noticePage.hasNotice("系统升级通知")).toBe(true);
  });

  test("TC-108c: 英文环境下动态插件页面与独立页种子内容保持英文", async ({
    adminPage,
    mainLayout,
  }) => {
    await ensurePluginInstalledAndEnabled();

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

    await expect
      .poll(async () => standalonePage.url())
      .toContain("lang=en-US");
    await expect(
      standalonePage.getByText("Standalone Page Opened Successfully", {
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
