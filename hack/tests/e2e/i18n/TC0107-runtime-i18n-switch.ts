import type { APIRequestContext } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";
import { PluginPage } from "../../pages/PluginPage";
import {
  createAdminApiContext,
  disablePlugin,
  expectSuccess,
  getPlugin,
  syncPlugins,
  uninstallPlugin,
} from "../../support/api/job";
import { waitForRouteReady } from "../../support/ui";

const pluginID = "plugin-demo-dynamic";
const pluginMenuNameEnglish = "Dynamic Plugin Demo";
const pluginMenuNameChinese = "动态插件示例";

type DictDataItem = {
  label: string;
  value: string;
};

let adminApi: APIRequestContext;
let originalInstalled = 0;
let originalEnabled = 0;

test.describe("TC0107 运行时国际化切换", () => {
  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    await syncPlugins(adminApi);
    const plugin = await getPlugin(adminApi, pluginID);
    originalInstalled = plugin.installed;
    originalEnabled = plugin.enabled;
  });

  test.afterAll(async () => {
    try {
      let plugin = await getPlugin(adminApi, pluginID);
      if (originalEnabled !== 1 && plugin.enabled === 1) {
        await disablePlugin(adminApi, pluginID);
        plugin = await getPlugin(adminApi, pluginID);
      }
      if (originalInstalled !== 1 && plugin.installed === 1) {
        await uninstallPlugin(adminApi, pluginID);
      }
    } finally {
      await adminApi.dispose();
    }
  });

  test("TC-107a: 登录页语言切换会同步刷新公共前端文案", async ({
    loginPage,
  }) => {
    await loginPage.goto();

    await expect(loginPage.loginSubtitle).toContainText(
      "请输入您的帐户信息以开始管理您的项目",
    );

    await loginPage.switchLanguage("English");
    await expect(loginPage.loginSubtitle).toContainText(
      "Enter your account credentials to start managing your projects",
    );

    await loginPage.switchLanguage("简体中文");
    await expect(loginPage.loginSubtitle).toContainText(
      "请输入您的帐户信息以开始管理您的项目",
    );
  });

  test("TC-107b: 已登录会话切换语言后菜单、系统信息与字典动态元数据会同步刷新", async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage("English");

    await expect(
      adminPage.getByText("Extensions", { exact: true }).first(),
    ).toBeVisible();
    await expect(
      adminPage.getByText("Settings", { exact: true }).first(),
    ).toBeVisible();

    await adminPage.goto("/about/system-info");
    await waitForRouteReady(adminPage);
    const systemInfoContent = adminPage.locator('[id="__vben_main_content"]');
    await expect(
      systemInfoContent.getByText("About LinaPro", { exact: true }),
    ).toBeVisible();
    await expect(
      systemInfoContent.getByText("Framework Name", { exact: true }),
    ).toBeVisible();
    await expect(
      systemInfoContent.getByText(
        "AI-driven full-stack development framework",
        { exact: false },
      ),
    ).toBeVisible();

    const dictData = await expectSuccess<{ list: DictDataItem[] }>(
      await adminApi.get("dict/data/type/sys_user_sex", {
        headers: {
          "Accept-Language": "en-US",
        },
      }),
    );
    expect(dictData.list.map((item) => item.label)).toContain("Male");
  });

  test("TC-107c: 动态插件页面在语言切换后刷新运行时翻译与宿主上下文", async ({
    adminPage,
    mainLayout,
  }) => {
    const pluginPage = new PluginPage(adminPage);

    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);
    const plugin = await getPlugin(adminApi, pluginID);
    if (plugin.installed !== 1) {
      await pluginPage.installAndEnablePlugin(pluginID);
    } else {
      await pluginPage.setPluginEnabled(pluginID, true);
    }

    await mainLayout.switchLanguage("English");
    await expect(
      pluginPage.sidebarMenuItem(pluginMenuNameEnglish),
    ).toBeVisible();
    await pluginPage.clickSidebarMenuItem(pluginMenuNameEnglish);
    await expect(pluginPage.pluginDemoDynamicTitle()).toHaveText(
      "Dynamic Plugin Demo Is Live",
    );
    await expect(pluginPage.pluginDemoDynamicDescription()).toContainText(
      "This page is mounted from the plugin-demo-dynamic embedded entry",
    );

    await mainLayout.switchLanguage("简体中文");
    await expect(
      pluginPage.sidebarMenuItem(pluginMenuNameChinese),
    ).toBeVisible();
    await expect(pluginPage.pluginDemoDynamicTitle()).toHaveText(
      "动态插件示例已生效",
    );
    await expect(pluginPage.pluginDemoDynamicDescription()).toContainText(
      "该页面来自 plugin-demo-dynamic 的动态挂载入口",
    );
  });
});
