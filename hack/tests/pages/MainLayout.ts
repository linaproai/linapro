import type { Page } from '@playwright/test';

export class MainLayout {
  constructor(private page: Page) {}

  get sidebar() {
    return this.page.locator('[class*="sidebar"], [class*="menu"], nav').first();
  }

  async navigateTo(menuGroup: string, menuItem: string) {
    await this.page.getByText(menuGroup).click();
    await this.page.getByText(menuItem).click();
    await this.page.waitForLoadState('networkidle');
  }

  async logout() {
    // Use keyboard shortcut Alt+Q to trigger the logout modal
    // This avoids the complex DOM interaction with the user dropdown
    await this.page.keyboard.press('Alt+KeyQ');

    // Wait for the confirmation modal to appear
    // The modal asks "是否退出登录？" with 确认/取消 buttons
    const confirmBtn = this.page.getByRole('button', {
      name: /确\s*认|confirm/i,
    });
    await confirmBtn.waitFor({ state: 'visible', timeout: 5000 });
    await confirmBtn.click();

    // Wait for redirect to login page
    await this.page.waitForURL(/auth\/login/, { timeout: 10000 });
  }
}
