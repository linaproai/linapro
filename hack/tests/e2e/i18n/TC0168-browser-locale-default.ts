import { LoginPage } from "../../pages/LoginPage";
import { config } from "../../fixtures/config";
import { test, expect } from "../../fixtures/auth";

test.describe("TC-168 browser locale default", () => {
  test("TC-168a: Chinese browser language starts the workbench in Simplified Chinese", async ({
    browser,
  }) => {
    const context = await browser.newContext({
      baseURL: config.baseURL,
      locale: "zh-CN",
    });
    const page = await context.newPage();
    const loginPage = new LoginPage(page);

    try {
      await loginPage.goto();

      await expect
        .poll(async () => page.locator("html").getAttribute("lang"))
        .toBe("zh-CN");
      await expect(loginPage.loginSubtitle).toContainText(
        "请输入您的帐户信息以开始管理您的项目",
      );
    } finally {
      await context.close();
    }
  });

  test("TC-168b: Non-Chinese browser language starts the workbench in English", async ({
    browser,
  }) => {
    const context = await browser.newContext({
      baseURL: config.baseURL,
      locale: "en-US",
    });
    const page = await context.newPage();
    const loginPage = new LoginPage(page);

    try {
      await loginPage.goto();

      await expect
        .poll(async () => page.locator("html").getAttribute("lang"))
        .toBe("en-US");
      await expect(loginPage.loginSubtitle).toContainText(
        "Enter your account credentials to start managing your projects",
      );
    } finally {
      await context.close();
    }
  });
});
