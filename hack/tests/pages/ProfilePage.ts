import type { Locator, Page } from '@playwright/test';

import { waitForRouteReady } from '../support/ui';

export class ProfilePage {
  constructor(private page: Page) {}

  get profilePanel(): Locator {
    return this.page.locator('.ant-card').first();
  }

  get nicknameInput(): Locator {
    return this.page
      .getByPlaceholder(/请输入昵称|Please enter a nickname/i)
      .first();
  }

  get passwordTab(): Locator {
    return this.page.getByRole('tab', { name: /密码|Password/i }).first();
  }

  panelDisplayName(name: string): Locator {
    return this.profilePanel.getByText(name, { exact: true }).first();
  }

  async goto() {
    await this.page.goto('/profile');
    await waitForRouteReady(this.page);
    await this.nicknameInput.waitFor({ state: 'visible', timeout: 10000 });
  }

  async openPasswordTab() {
    await this.passwordTab.click();
    await waitForRouteReady(this.page);
  }
}
