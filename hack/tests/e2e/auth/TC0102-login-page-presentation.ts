import type { APIRequestContext } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";
import {
  createAdminApiContext,
  getConfigByKey,
  updateConfigValue,
} from "../../support/api/job";

test.describe("TC-102 登录页展示收口与布局", () => {
  let api: APIRequestContext;
  let originalLayout: { id: number; key: string; value: string };

  test.beforeAll(async () => {
    api = await createAdminApiContext();
    originalLayout = await getConfigByKey(api, "sys.auth.loginPanelLayout");
  });

  test.afterAll(async () => {
    await updateConfigValue(api, originalLayout.id, originalLayout.value);
    await api.dispose();
  });

  test("TC-102a: 登录页隐藏未实现入口并回退未实现认证子路由", async ({
    loginPage,
    page,
  }) => {
    await loginPage.goto();

    await expect(loginPage.forgotPasswordEntry).toBeHidden();
    await expect(loginPage.createAccountEntry).toBeHidden();
    await expect(loginPage.mobileLoginButton).toBeHidden();
    await expect(loginPage.qrCodeLoginButton).toBeHidden();
    await expect(loginPage.thirdPartyLoginTitle).toBeHidden();

    for (const path of [
      "/auth/code-login",
      "/auth/qrcode-login",
      "/auth/forget-password",
      "/auth/register",
    ]) {
      await page.goto(path);
      await page.waitForURL(/\/auth\/login$/, { timeout: 10000 });
      await expect(loginPage.usernameInput).toBeVisible();
    }
  });

  test("TC-102b: 登录页默认使用居右登录框布局", async ({ loginPage }) => {
    await updateConfigValue(api, originalLayout.id, "panel-right");

    await loginPage.goto();

    await expect(loginPage.rightAuthPanel).toBeVisible();
    await expect(loginPage.leftAuthPanel).toBeHidden();
    await expect(loginPage.centerAuthPanel).toBeHidden();
  });

  test("TC-102c: 修改系统参数后登录页按配置切换布局", async ({ loginPage }) => {
    await updateConfigValue(api, originalLayout.id, "panel-left");
    await loginPage.goto();
    await expect(loginPage.leftAuthPanel).toBeVisible();
    await expect(loginPage.rightAuthPanel).toBeHidden();

    await updateConfigValue(api, originalLayout.id, "panel-center");
    await loginPage.goto();
    await expect(loginPage.centerAuthPanel).toBeVisible();
    await expect(loginPage.leftAuthPanel).toBeHidden();
  });

  test("TC-102d: 登录页密码占位符跟随默认双语切换", async ({ loginPage }) => {
    await loginPage.goto();

    await expect(loginPage.passwordInput).toHaveAttribute(
      "placeholder",
      "请输入密码",
    );

    await loginPage.switchLanguage("English");
    await expect(loginPage.passwordInput).toHaveAttribute(
      "placeholder",
      "Please enter password",
    );
  });
});
