import type { Page } from "@playwright/test";

export class LoginPage {
  constructor(private page: Page) {}

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

  async goto() {
    await this.page.goto("/auth/login");
    await this.usernameInput.waitFor({ state: "visible" });
    await this.loginButton.waitFor({ state: "visible" });
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
