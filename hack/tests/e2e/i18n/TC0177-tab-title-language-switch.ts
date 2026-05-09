import { test, expect } from "../../fixtures/auth";
import { waitForRouteReady } from "../../support/ui";

test.describe("TC-177 tab title language switch", () => {
  test("TC-177a: opened tab titles relocalize immediately after language switch", async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage("简体中文");

    await adminPage.goto("/dashboard/workspace", { waitUntil: "domcontentloaded" });
    await waitForRouteReady(adminPage);
    await adminPage.goto("/system/user", { waitUntil: "domcontentloaded" });
    await waitForRouteReady(adminPage);

    await expect(mainLayout.tabTitle("分析页")).toBeVisible();
    await expect(mainLayout.tabTitle("工作台")).toBeVisible();
    await expect(mainLayout.tabTitle("用户管理")).toBeVisible();

    await mainLayout.switchLanguage("English");

    await expect(mainLayout.tabTitle("Analytics")).toBeVisible();
    await expect(mainLayout.tabTitle("Workspace")).toBeVisible();
    await expect(mainLayout.tabTitle("Users")).toBeVisible();
    await expect(mainLayout.tabTitle("分析页")).toHaveCount(0);
    await expect(mainLayout.tabTitle("工作台")).toHaveCount(0);
    await expect(mainLayout.tabTitle("用户管理")).toHaveCount(0);

    await mainLayout.switchLanguage("简体中文");

    await expect(mainLayout.tabTitle("分析页")).toBeVisible();
    await expect(mainLayout.tabTitle("工作台")).toBeVisible();
    await expect(mainLayout.tabTitle("用户管理")).toBeVisible();
    await expect(mainLayout.tabTitle("Analytics")).toHaveCount(0);
    await expect(mainLayout.tabTitle("Workspace")).toHaveCount(0);
    await expect(mainLayout.tabTitle("Users")).toHaveCount(0);
  });
});
