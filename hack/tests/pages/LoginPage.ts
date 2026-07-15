import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";

import { workspacePath } from "../fixtures/config";
import { waitForRouteReady } from "../support/ui";

export class LoginPage {
  constructor(private page: Page) {}

  private async waitForLocalePersistence(locale: string) {
    await expect
      .poll(async () => {
        try {
          return await this.page.evaluate(() => {
            const key = Object.keys(localStorage).find((item) =>
              item.endsWith("preferences-locale"),
            );
            if (!key) {
              return "";
            }
            try {
              return JSON.parse(localStorage.getItem(key) || "{}")?.value || "";
            } catch {
              return "";
            }
          });
        } catch {
          return "";
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

  get brandLogoImage() {
    return this.page.locator('img[alt^="LinaPro"]:visible').first();
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
    return this.page
      .getByTestId("login-create-account-link")
      .or(this.page.getByText("创建账号", { exact: true }).first());
  }

  get createAccountRegion() {
    return this.page.getByTestId("login-create-account");
  }

  get socialAuthRegionOrDivider() {
    return this.page.getByTestId("login-social-auth-region");
  }

  get forgetPasswordEmailInput() {
    return this.page
      .locator(
        '#email, [name="email"], input[placeholder*="example@"], input[placeholder*="邮箱"], input[placeholder*="email"]',
      )
      .first();
  }

  get forgetPasswordSubmitButton() {
    return this.page.locator('button[aria-label="submit"]');
  }

  get registerSubmitButton() {
    return this.page.locator('button[aria-label="register"]');
  }

  get registerUsernameInput() {
    return this.page
      .locator(
        '#username, [name="username"], input[placeholder*="用户名"], input[placeholder*="username"], input[placeholder*="account"]',
      )
      .first();
  }

  get registerPasswordInput() {
    return this.page
      .locator(
        '#password, [name="password"], input[placeholder*="密码"], input[placeholder*="password"]',
      )
      .first();
  }

  get registerConfirmPasswordInput() {
    return this.page
      .locator(
        '#confirmPassword, [name="confirmPassword"], input[placeholder*="确认密码"], input[placeholder*="Confirm"]',
      )
      .first();
  }

  get registerAgreeCheckbox() {
    return this.page.locator('button[role="checkbox"], input[type="checkbox"]').first();
  }

  get goToLoginEntry() {
    return this.page
      .getByText(/去登录|Login instead/, { exact: true })
      .first();
  }

  get backToLoginButton() {
    return this.page.getByRole("button", { name: /返回|Back/i }).first();
  }

  get mobileLoginButton() {
    return this.page.getByRole("button", { name: "手机号登录" });
  }

  get qrCodeLoginButton() {
    return this.page.getByRole("button", { name: "扫码登录" });
  }

  /**
   * Host region for full-width protocol / directory login buttons
   * (`auth.login.after`: generic OIDC, LDAP, …).
   * Hidden when no plugin injects into that slot.
   */
  get externalAuthRegion() {
    return this.page.getByTestId("login-external-auth-region");
  }

  /**
   * Host region for platform social icon logins
   * (`auth.login.social`: Google, Discord, …) under “其他登录方式”.
   * Hidden when no plugin injects into that slot.
   */
  get socialAuthRegion() {
    return this.page.getByTestId("login-social-auth-region");
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

  /**
   * Slogan illustration image on the left/right login layout side panel.
   * Absent when `sys.auth.sloganImage` is empty.
   */
  get sloganImage() {
    return this.page.getByTestId("login-slogan-image");
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

  /**
   * Host-owned outlet for protocol / directory login extensions
   * (`auth.login.after`). Plugin-specific entries must not be hard-coded here.
   */
  get externalAuthSlot() {
    return this.page.getByTestId("login-external-auth-slot");
  }

  get externalAuthSlotItems() {
    return this.externalAuthSlot.locator(".plugin-slot-outlet__item");
  }

  /**
   * Host-owned outlet for platform social icon extensions
   * (`auth.login.social`). Plugin-specific entries must not be hard-coded here.
   */
  get socialAuthSlot() {
    return this.page.getByTestId("login-social-auth-slot");
  }

  get socialAuthSlotItems() {
    return this.socialAuthSlot.locator(".plugin-slot-outlet__item");
  }

  get errorMessage() {
    return this.page.getByText(
      /用户名或密码错误|incorrect|invalid|error|失败/i,
    );
  }

  getText(text: string) {
    return this.page.getByText(text, { exact: true }).first();
  }

  get languageToggleTrigger() {
    return this.page.getByTestId("language-toggle-trigger").first();
  }

  async goto() {
    await this.gotoPath(workspacePath("/auth/login"));
  }

  async gotoPath(path: string) {
    await this.page.goto(path);
    await this.usernameInput.waitFor({ state: "visible" });
    await this.loginButton.waitFor({ state: "visible" });
  }

  async getDocumentTitle() {
    return this.page.evaluate(() => document.title);
  }

  async getBodyText() {
    return this.page.locator("body").innerText();
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

  async getBrandLogoInfo() {
    await expect(this.brandLogoImage).toBeVisible();
    await expect
      .poll(async () =>
        this.brandLogoImage.evaluate(
          (img) => (img as HTMLImageElement).naturalWidth,
        ),
      )
      .toBeGreaterThan(0);

    return this.brandLogoImage.evaluate((node) => {
      const img = node as HTMLImageElement;
      return {
        currentSrc: img.currentSrc,
        height: img.clientHeight,
        naturalHeight: img.naturalHeight,
        naturalWidth: img.naturalWidth,
        parentText:
          (img.closest("a") ?? img.parentElement)?.textContent?.trim() ?? "",
        src: img.getAttribute("src") ?? "",
        width: img.clientWidth,
      };
    });
  }

  async switchLanguage(label: "English" | "简体中文") {
    const localeMap = {
      English: "en-US",
      简体中文: "zh-CN",
    } as const;
    const locale = localeMap[label];
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
