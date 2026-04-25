import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";

import { waitForRouteReady } from "../support/ui";

export class MainLayout {
  constructor(private page: Page) {}

  private async waitForLocalePersistence(locale: string) {
    await expect
      .poll(async () => {
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
    return this.page.locator('[data-tab-item="true"].is-active span[title]').first();
  }

  get userDropdownTrigger() {
    return this.page.getByTestId('layout-user-dropdown-trigger').first();
  }

  get userDropdownMenu() {
    return this.page.getByTestId('layout-user-dropdown-menu');
  }

  get userDropdownProfile() {
    return this.page.getByTestId('layout-user-dropdown-profile');
  }

  get userDropdownName() {
    return this.page.getByTestId('layout-user-dropdown-name');
  }

  async navigateTo(menuGroup: string, menuItem: string) {
    await this.page.getByText(menuGroup).click();
    await this.page.getByText(menuItem).click();
    await this.page.waitForLoadState("networkidle");
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

  async openUserDropdown() {
    await this.userDropdownTrigger.click();
    await expect(this.userDropdownMenu).toBeVisible();
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
    await confirmBtn.waitFor({ state: "visible", timeout: 5000 });
    await confirmBtn.click();

    // Wait for redirect to login page
    await this.page.waitForURL(/auth\/login/, { timeout: 10000 });
  }
}
