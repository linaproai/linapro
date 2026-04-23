import { test, expect } from "../../../fixtures/auth";
import { PluginPage } from "../../../pages/PluginPage";

const pluginID = "plugin-demo-source";
const pluginName = "源码插件示例";
const pluginVersion = "v0.1.0";
const pluginDescription = "提供左侧菜单页面与公开/受保护路由示例的源码插件";

test.describe("TC-78 插件详情弹窗", () => {
  let pluginPage: PluginPage;

  test.beforeEach(async ({ adminPage }) => {
    pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(pluginID);
  });

  test("TC-78a: 点击详情按钮展示插件基础治理信息", async () => {
    await pluginPage.openPluginDetail(pluginID);

    const modal = pluginPage.pluginDetailModal();
    await expect(modal).toContainText(pluginName);
    await expect(modal).toContainText(pluginID);
    await expect(modal).toContainText("源码插件");
    await expect(modal).toContainText(pluginVersion);
    await expect(modal).toContainText(pluginDescription);
    await expect(modal).toContainText("接入状态");
    await expect(modal).toContainText("当前状态");
    await expect(modal).toContainText("授权要求");
    await expect(modal).toContainText("授权状态");
    await expect(modal).toContainText("安装时间");
    await expect(modal).toContainText("更新时间");
  });

  test("TC-78b: 源码插件详情页不展示多余的宿主服务空状态提示", async () => {
    await pluginPage.openPluginDetail(pluginID);
    await expect(pluginPage.pluginDetailEmptyHostServices()).toHaveCount(0);
    await expect(pluginPage.pluginDetailModal()).not.toContainText(
      "当前动态插件未声明额外宿主服务。",
    );
  });
});
