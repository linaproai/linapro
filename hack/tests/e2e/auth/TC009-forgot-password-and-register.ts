import type { APIRequestContext } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";
import { workspacePath } from "../../fixtures/config";
import { LoginPage } from "../../pages/LoginPage";
import {
  createAdminApiContext,
  getConfigByKey,
  updateConfigValue,
} from "../../support/api/job";

test.describe("TC-9 忘记密码与创建账号", () => {
  let api: APIRequestContext;
  let originalForgetPassword: { id: number; key: string; value: string };
  let originalRegister: { id: number; key: string; value: string };

  test.beforeAll(async () => {
    api = await createAdminApiContext();
    originalForgetPassword = await getConfigByKey(
      api,
      "sys.auth.forgetPasswordEnabled",
    );
    originalRegister = await getConfigByKey(api, "sys.auth.registerEnabled");
  });

  test.afterAll(async () => {
    await updateConfigValue(
      api,
      originalForgetPassword.id,
      originalForgetPassword.value,
    );
    await updateConfigValue(api, originalRegister.id, originalRegister.value);
    await api.dispose();
  });

  test.afterEach(async () => {
    await updateConfigValue(api, originalForgetPassword.id, "true");
    await updateConfigValue(api, originalRegister.id, "true");
  });

  test("TC-9a: 登录页可进入忘记密码页并返回登录", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();

    await expect(loginPage.forgotPasswordEntry).toBeVisible();
    await loginPage.forgotPasswordEntry.click();
    await page.waitForURL(/\/auth\/forget-password$/, { timeout: 10000 });

    await expect(
      page.getByText("忘记密码", { exact: false }).first(),
    ).toBeVisible();
    await expect(loginPage.forgetPasswordEmailInput).toBeVisible();
    await expect(
      page.getByText("我们将向您发送重置密码", { exact: false }),
    ).toBeVisible();

    await loginPage.backToLoginButton.click();
    await page.waitForURL(/\/auth\/login$/, { timeout: 10000 });
    await expect(loginPage.usernameInput).toBeVisible();
  });

  test("TC-9b: 忘记密码页非法邮箱阻止提交；合法邮箱调用恢复接口", async ({
    page,
  }) => {
    const loginPage = new LoginPage(page);
    await page.goto(workspacePath("/auth/forget-password"));
    await expect(loginPage.forgetPasswordEmailInput).toBeVisible();

    await loginPage.forgetPasswordEmailInput.fill("not-an-email");
    await loginPage.forgetPasswordSubmitButton.click();
    await expect(
      page.getByText("你输入的邮箱格式不正确", { exact: true }),
    ).toBeVisible();

    await loginPage.forgetPasswordEmailInput.fill("user@example.com");
    await loginPage.forgetPasswordSubmitButton.click();
    // Mail channel may or may not be configured in the local stack. Accept
    // either successful acceptance or the explicit unavailable error.
    await expect(
      page
        .getByText("重置邮件已受理", { exact: true })
        .or(page.getByText("密码找回暂时不可用", { exact: false })),
    ).toBeVisible({ timeout: 15000 });
  });

  test("TC-9c: 登录页可进入创建账号页并返回登录", async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();

    await expect(loginPage.createAccountEntry).toBeVisible();
    await loginPage.createAccountEntry.click();
    await page.waitForURL(/\/auth\/register$/, { timeout: 10000 });

    await expect(
      page.getByText("创建一个账号", { exact: false }).first(),
    ).toBeVisible();
    await expect(loginPage.registerUsernameInput).toBeVisible();
    await expect(
      page.getByText("快速开启协作开发", { exact: false }),
    ).toBeVisible();

    await loginPage.goToLoginEntry.click();
    await page.waitForURL(/\/auth\/login$/, { timeout: 10000 });
    await expect(loginPage.usernameInput).toBeVisible();
  });

  test("TC-9c2: 创建账号页可打开隐私政策与服务条款弹窗", async ({ page }) => {
    await page.goto(workspacePath("/auth/register"));
    await expect(
      page.getByTestId("register-privacy-link").or(page.getByText("隐私政策", { exact: true }).first()),
    ).toBeVisible();

    // Prefer testids; fall back to exact link text when vnode attrs are stripped.
    const privacyLink = page.getByTestId("register-privacy-link").or(
      page.locator("a.vben-link", { hasText: "隐私政策" }).first(),
    );
    const termsLink = page.getByTestId("register-terms-link").or(
      page.locator("a.vben-link", { hasText: "条款" }).first(),
    );

    await privacyLink.click();
    await expect(page.getByTestId("register-privacy-modal-body")).toBeVisible();
    await expect(page.getByTestId("register-privacy-modal-body")).toContainText(
      /隐私政策|账号认证|Privacy Policy/i,
    );
    await page.locator(".ant-modal-close").first().click();
    await expect(page.getByTestId("register-privacy-modal-body")).toBeHidden();

    await termsLink.click();
    await expect(page.getByTestId("register-terms-modal-body")).toBeVisible();
    await expect(page.getByTestId("register-terms-modal-body")).toContainText(
      /服务条款|工作区|Terms/i,
    );
  });

  test("TC-9d: 创建账号成功后可登录", async ({ page }) => {
    const loginPage = new LoginPage(page);
    const suffix = Date.now();
    const username = `e2e_reg_${suffix}`;
    const email = `${username}@example.com`;
    const password = "Passw0rd!";

    await page.goto(workspacePath("/auth/register"));
    await expect(loginPage.registerUsernameInput).toBeVisible();

    await loginPage.registerUsernameInput.fill(username);
    await page
      .locator(
        '#email, [name="email"], input[placeholder*="example@"], input[placeholder*="邮箱"]',
      )
      .first()
      .fill(email);
    await loginPage.registerPasswordInput.fill(password);
    await loginPage.registerConfirmPasswordInput.fill(password);
    await loginPage.registerAgreeCheckbox.click();
    await loginPage.registerSubmitButton.click();

    await expect(
      page.getByText("注册成功", { exact: true }),
    ).toBeVisible({ timeout: 15000 });
    await page.waitForURL(/\/auth\/login$/, { timeout: 15000 });

    // Re-enter login page cleanly so the form is fully mounted after registration.
    await loginPage.goto();
    await loginPage.loginAndWaitForRedirect(username, password);
    expect(page.url()).not.toContain("/auth/login");
  });

  test("TC-9e: 创建账号入口位于其他登录方式区域之后", async ({
    loginPage,
    page,
  }) => {
    await loginPage.goto();

    await page.evaluate(() => {
      const region = document.querySelector(
        '[data-testid="login-social-auth-region"]',
      ) as HTMLElement | null;
      if (!region) {
        return;
      }
      region.style.display = "block";
      let host = region.querySelector(
        '[data-testid="login-social-auth-slot"]',
      ) as HTMLElement | null;
      if (!host) {
        host = document.createElement("div");
        host.className =
          "mt-4 flex flex-wrap justify-center plugin-slot-outlet";
        host.setAttribute("data-testid", "login-social-auth-slot");
        region.appendChild(host);
      }
      host.innerHTML =
        '<div class="plugin-slot-outlet__item" style="height:36px;width:36px"></div>';
    });

    await expect(loginPage.createAccountRegion).toBeVisible();
    await expect(loginPage.socialAuthRegionOrDivider).toBeVisible();

    const socialBox = await loginPage.socialAuthRegionOrDivider.boundingBox();
    const createBox = await loginPage.createAccountRegion.boundingBox();
    expect(socialBox).not.toBeNull();
    expect(createBox).not.toBeNull();
    expect(createBox!.y).toBeGreaterThan(socialBox!.y + socialBox!.height - 1);
  });

  test("TC-9f: 关闭系统开关后隐藏入口并回退子路由", async ({
    loginPage,
    page,
  }) => {
    await updateConfigValue(api, originalForgetPassword.id, "false");
    await updateConfigValue(api, originalRegister.id, "false");

    await loginPage.goto();
    await expect(loginPage.forgotPasswordEntry).toBeHidden();
    await expect(loginPage.createAccountEntry).toBeHidden();

    await page.goto(workspacePath("/auth/forget-password"));
    await page.waitForURL(/\/auth\/login$/, { timeout: 10000 });
    await expect(loginPage.usernameInput).toBeVisible();

    await page.goto(workspacePath("/auth/register"));
    await page.waitForURL(/\/auth\/login$/, { timeout: 10000 });
    await expect(loginPage.usernameInput).toBeVisible();
  });
});
