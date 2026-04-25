import type { APIRequestContext, Page } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import {
  createAdminApiContext,
  findPlugin,
  installPlugin,
  syncPlugins,
  uninstallPlugin,
  updatePluginStatus,
} from "../../../fixtures/plugin";
import { PluginPage } from "../../../pages/PluginPage";

const pluginID = "plugin-demo-source";

function unwrapApiData(payload: any) {
  if (payload && typeof payload === "object" && "data" in payload) {
    return payload.data;
  }
  return payload;
}

async function mockAutoEnableManagedPlugin(page: Page, targetPluginID: string) {
  await page.route("**/api/v1/plugins**", async (route) => {
    if (route.request().method() !== "GET") {
      await route.continue();
      return;
    }

    const response = await route.fetch();
    const payload = await response.json();
    const data = unwrapApiData(payload);
    if (Array.isArray(data?.list)) {
      data.list = data.list.map((item: Record<string, unknown>) => ({
        ...item,
        autoEnableManaged: item.id === targetPluginID ? 1 : 0,
      }));
    }

    await route.fulfill({
      json: payload,
      response,
    });
  });
}

test.describe("TC-104 plugin.autoEnable 管理提示", () => {
  let adminApi: APIRequestContext;
  let originalEnabled = 0;
  let originalInstalled = 0;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    await syncPlugins(adminApi);
    const plugin = await findPlugin(adminApi, pluginID);
    originalInstalled = plugin?.installed ?? 0;
    originalEnabled = plugin?.enabled ?? 0;
  });

  test.afterAll(async () => {
    try {
      await syncPlugins(adminApi);
      const plugin = await findPlugin(adminApi, pluginID);
      if (!plugin) {
        return;
      }
      if (originalInstalled !== 1 && plugin.installed === 1) {
        await uninstallPlugin(adminApi, pluginID);
        return;
      }
      if (
        originalInstalled === 1 &&
        originalEnabled !== 1 &&
        plugin.enabled === 1
      ) {
        await updatePluginStatus(adminApi, pluginID, false);
      }
      if (
        originalInstalled === 1 &&
        originalEnabled === 1 &&
        plugin.enabled !== 1
      ) {
        await updatePluginStatus(adminApi, pluginID, true);
      }
    } finally {
      await adminApi.dispose();
    }
  });

  test.beforeEach(async ({ adminPage }) => {
    await syncPlugins(adminApi);
    const plugin = await findPlugin(adminApi, pluginID);
    if (!plugin) {
      throw new Error(`未找到插件: ${pluginID}`);
    }
    if (plugin.installed !== 1) {
      await installPlugin(adminApi, pluginID);
    }
    if (plugin.enabled !== 1) {
      await updatePluginStatus(adminApi, pluginID, true);
    }
    await mockAutoEnableManagedPlugin(adminPage, pluginID);
  });

  test("TC-104a: 列表和详情展示 plugin.autoEnable 管理标识", async ({
    adminPage,
  }) => {
    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);

    await expect(pluginPage.pluginAutoEnableTag(pluginID)).toBeVisible();
    await pluginPage.pluginAutoEnableTag(pluginID).hover();
    await expect(pluginPage.antTooltip()).toContainText(
      "宿主自动启用策略管理",
    );

    await pluginPage.openPluginDetail(pluginID);
    await expect(pluginPage.pluginDetailModal()).toContainText("启动管理");
    await expect(pluginPage.pluginDetailModal()).toContainText(
      "宿主管理自动启用",
    );
    await expect(pluginPage.pluginAutoEnableDetailAlert()).toContainText(
      "手动变更可能在下一次同步时被运行时恢复",
    );
  });

  test("TC-104b: 禁用与卸载前提示仅为运行期临时治理", async ({ adminPage }) => {
    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);

    const pluginSwitch = pluginPage.pluginEnabledSwitch(pluginID);
    await expect(pluginSwitch).toHaveAttribute("aria-checked", "true");

    await pluginSwitch.click();
    await expect(pluginPage.pluginManagedActionDialog()).toContainText(
      "运行时仍可能重新应用宿主管理的自动启用策略",
    );
    await pluginPage.cancelManagedActionWarning();
    await expect(pluginSwitch).toHaveAttribute("aria-checked", "true");

    await pluginSwitch.click();
    await pluginPage.confirmManagedActionWarning();
    await expect(pluginSwitch).toHaveAttribute("aria-checked", "false");
    await expect(pluginPage.messageNotice("插件已停用")).toBeVisible();

    await pluginPage.openUninstallDialog(pluginID);
    await expect(pluginPage.pluginAutoEnableUninstallAlert()).toContainText(
      "运行时仍可能按宿主策略重新安装或启用",
    );
    await pluginPage.cancelUninstallDialog();
  });
});
