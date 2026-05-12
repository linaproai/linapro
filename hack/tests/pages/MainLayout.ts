import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";

import { waitForRouteReady } from "../support/ui";

export class MainLayout {
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

  get sidebar() {
    return this.page
      .locator('[class*="sidebar"], [class*="menu"], nav')
      .first();
  }

  get languageToggleTrigger() {
    return this.page.getByTestId("language-toggle-trigger").first();
  }

  get brandLogoImage() {
    return this.page.locator('img[alt^="LinaPro"]:visible').first();
  }

  get brandLogoMark() {
    return this.page.locator(".vben-logo__mark:visible").first();
  }

  sidebarMenuItem(label: string) {
    return this.sidebar.getByText(label, { exact: true }).first();
  }

  tabTitle(label: string) {
    return this.page
      .locator('[data-tab-item="true"] span[title]')
      .filter({ hasText: label })
      .first();
  }

  activeTabTitle() {
    return this.page
      .locator('[data-tab-item="true"].is-active span[title]')
      .first();
  }

  get userDropdownTrigger() {
    return this.page.getByTestId("layout-user-dropdown-trigger").first();
  }

  get userDropdownMenu() {
    return this.page.getByTestId("layout-user-dropdown-menu");
  }

  get userDropdownProfile() {
    return this.page.getByTestId("layout-user-dropdown-profile");
  }

  get userDropdownName() {
    return this.page.getByTestId("layout-user-dropdown-name");
  }

  get preferencesTrigger() {
    return this.page.getByTestId("preferences-trigger").first();
  }

  get preferencesDrawerTitle() {
    return this.page.getByTestId("preferences-drawer-title").first();
  }

  get preferencesDrawerSubtitle() {
    return this.page.getByTestId("preferences-drawer-subtitle").first();
  }

  get preferencesDrawer() {
    return this.page
      .locator('[role="dialog"], [data-slot="sheet-content"]')
      .filter({
        has: this.page.locator('[data-testid="preferences-drawer-title"]'),
      })
      .first();
  }

  get tenantSwitcher() {
    return this.page.getByTestId("tenant-switcher");
  }

  get workspaceFooterCopyright() {
    return this.page
      .locator("footer")
      .filter({ hasText: "Copyright ©" })
      .first()
      .getByText(/Copyright ©/);
  }

  async navigateTo(menuGroup: string, menuItem: string) {
    await this.page.getByText(menuGroup).click();
    await this.page.getByText(menuItem).click();
    await this.page.waitForLoadState("networkidle");
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

  async switchLanguageFromPreferences(
    label: "English" | "简体中文",
  ) {
    const localeMap = {
      English: "en-US",
      简体中文: "zh-CN",
    } as const;
    const locale = localeMap[label];
    if (!(await this.preferencesDrawer.isVisible().catch(() => false))) {
      await this.openPreferences();
    }
    await this.preferencesDrawer.getByRole("tab", { name: /General|通用/ }).click();
    await this.preferencesDrawer.getByRole("combobox").first().click();
    await this.page.getByRole("option", { name: label }).click();
    await this.waitForLocalePersistence(locale);
    await expect
      .poll(async () => await this.page.locator("html").getAttribute("lang"))
      .toBe(locale);
    await waitForRouteReady(this.page);
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

  async getBrandLogoGlowInfo() {
    await expect(this.brandLogoMark).toBeVisible();

    return this.brandLogoMark.evaluate((mark) => {
      const root = mark.closest(".vben-logo");
      const link = mark.closest("a");
      const image = mark.querySelector("img") as HTMLElement;
      const beforeStyle = window.getComputedStyle(mark, "::before");
      const imageStyle = image ? window.getComputedStyle(image) : null;
      const markRect = mark.getBoundingClientRect();
      const linkRect = link?.getBoundingClientRect();

      return {
        beforeBackgroundImage: beforeStyle.backgroundImage,
        beforeFilter: beforeStyle.filter,
        beforeOpacity: beforeStyle.opacity,
        imageClientHeight: image?.clientHeight ?? 0,
        imageClientWidth: image?.clientWidth ?? 0,
        imageFilter: imageStyle?.filter ?? "",
        isDarkRoot: root?.classList.contains("dark") ?? false,
        linkHeight: linkRect?.height ?? 0,
        linkOverflow: link ? window.getComputedStyle(link).overflow : "",
        markHeight: markRect.height,
        markWidth: markRect.width,
      };
    });
  }

  async ensureThemeMode(mode: "dark" | "light") {
    const shouldBeDark = mode === "dark";
    const isDark = await this.page.evaluate(() =>
      document.documentElement.classList.contains("dark"),
    );

    if (isDark !== shouldBeDark) {
      await this.page
        .locator(`button[aria-label="${mode}"]:visible`)
        .first()
        .click();
    }

    await expect
      .poll(async () =>
        this.page.evaluate(() =>
          document.documentElement.classList.contains("dark"),
        ),
      )
      .toBe(shouldBeDark);
    await waitForRouteReady(this.page);
  }

  async openUserDropdown() {
    await this.userDropdownTrigger.click();
    await expect(this.userDropdownMenu).toBeVisible();
  }

  async openPreferences() {
    await expect(this.preferencesTrigger).toBeVisible();
    await this.preferencesTrigger.click();
    await expect(this.preferencesDrawerTitle).toBeVisible();
    await expect(this.preferencesDrawer).toBeVisible();
  }

  async openPreferencesTab(label: string | RegExp) {
    await this.openPreferences();
    await this.preferencesDrawer.getByRole("tab", { name: label }).click();
  }

  async logout() {
    // Use keyboard shortcut Alt+Q to trigger the logout modal
    // This avoids the complex DOM interaction with the user dropdown
    await this.page.keyboard.press("Alt+KeyQ");

    // Wait for the confirmation modal to appear
    // The modal asks "是否退出登录？" with 确认/取消 buttons
    const confirmBtn = this.page.getByRole("button", {
      name: /确\s*认|confirm/i,
    });
    await confirmBtn.waitFor({ state: "visible", timeout: 1500 });
    await confirmBtn.click();

    // Wait for redirect to login page
    await this.page.waitForURL(/auth\/login/, { timeout: 10000 });
  }
}
