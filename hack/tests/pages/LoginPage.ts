import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";

import { waitForRouteReady } from "../support/ui";

export class LoginPage {
  constructor(private page: Page) {}

  private async waitForLocalePersistence(locale: string) {
    await expect
      .poll(async () => {
        try {
          return await this.page.evaluate(() => {
            const key = Object.keys(localStorage).find((item) =>
              item.endsWith('preferences-locale'),
            );
            if (!key) {
              return '';
            }
            try {
              return JSON.parse(localStorage.getItem(key) || '{}')?.value || '';
            } catch {
              return '';
            }
          });
        } catch {
          return '';
        }
      })
      .toBe(locale);
  }

  get appName() {
    return this.page
      .getByText(/LinaPro|Console|品牌/, { exact: false })
      .first();
  }

  get pageTitle() {
    return this.page.locator("h1").first();
  }

  get loadingTitle() {
    return this.page.locator("#__app-loading__ .title").first();
  }

  get pageDescription() {
    return this.page
      .locator("p")
      .filter({ hasText: /宿主|工作台|品牌|平台|能力/ })
      .first();
  }

  get loginSubtitle() {
    return this.page.locator("span.text-muted-foreground").first();
  }

  get forgotPasswordEntry() {
    return this.page.getByText("忘记密码?", { exact: true }).first();
  }

  get createAccountEntry() {
    return this.page.getByText("创建账号", { exact: true }).first();
  }

  get mobileLoginButton() {
    return this.page.getByRole("button", { name: "手机号登录" });
  }

  get qrCodeLoginButton() {
    return this.page.getByRole("button", { name: "扫码登录" });
  }

  get thirdPartyLoginTitle() {
    return this.page.getByText("其他登录方式", { exact: true }).first();
  }

  get leftAuthPanel() {
    return this.page.locator('.side-content[data-side="left"]').first();
  }

  get centerAuthPanel() {
    return this.page.locator('.side-content[data-side="bottom"]').first();
  }

  get rightAuthPanel() {
    return this.page.locator('.side-content[data-side="right"]').first();
  }

  get usernameInput() {
    return this.page
      .locator(
        '#username, [name="username"], input[placeholder*="用户名"], input[placeholder*="username"], input[placeholder*="account"]',
      )
      .first();
  }

  get passwordInput() {
    return this.page
      .locator(
        '#password, [name="password"], input[placeholder*="密码"], input[placeholder*="password"]',
      )
      .first();
  }

  get loginButton() {
    // The main login button has aria-label="login", distinguishing it from
    // "手机号登录" and "扫码登录" buttons
    return this.page.locator('button[aria-label="login"]');
  }

  get errorMessage() {
    return this.page.getByText(
      /用户名或密码错误|incorrect|invalid|error|失败/i,
    );
  }

  get pluginLoginSlot() {
    return this.page.getByText(
      "plugin-demo-source 已向登录页公开区注册扩展内容，用于验证 `auth.login.after` 插槽。",
    );
  }

  getText(text: string) {
    return this.page.getByText(text, { exact: true }).first();
  }

  get languageToggleTrigger() {
    return this.page.getByTestId("language-toggle-trigger").first();
  }

  async goto() {
    await this.gotoPath("/auth/login");
  }

  async gotoPath(path: string) {
    await this.page.goto(path);
    await this.usernameInput.waitFor({ state: "visible" });
    await this.loginButton.waitFor({ state: "visible" });
  }

  async getDocumentTitle() {
    return this.page.evaluate(() => document.title);
  }

  async getLoadingTitleFontFamily() {
    return this.loadingTitle.evaluate(
      (node) => getComputedStyle(node).fontFamily,
    );
  }

  async getRootFontFamily() {
    return this.page.evaluate(() => {
      return getComputedStyle(document.documentElement).fontFamily;
    });
  }

  async switchLanguage(label: "English" | "简体中文") {
    const locale = label === "English" ? "en-US" : "zh-CN";
    await this.languageToggleTrigger.click();
    await this.page.getByText(label, { exact: true }).last().click();
    await this.waitForLocalePersistence(locale);
    await expect
      .poll(async () => await this.page.locator("html").getAttribute("lang"))
      .toBe(locale);
    await this.page.waitForLoadState("networkidle");
    await waitForRouteReady(this.page);
  }

  async login(username: string, password: string) {
    await this.usernameInput.fill(username);
    await this.passwordInput.fill(password);
    await this.loginButton.click();
  }

  async loginAndWaitForRedirect(username: string, password: string) {
    await this.login(username, password);
    await this.page.waitForURL((url) => !url.pathname.includes("/auth/login"), {
      timeout: 15000,
    });
  }
}
