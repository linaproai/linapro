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

  get passwordForm(): Locator {
    return this.page.getByTestId('profile-password-form');
  }

  get oldPasswordInput(): Locator {
    return this.passwordForm
      .getByPlaceholder(/请输入当前密码|请输入旧密码|current password|old password/i)
      .first();
  }

  get newPasswordInput(): Locator {
    return this.passwordForm
      .getByPlaceholder(/请输入新密码|new password/i)
      .first();
  }

  get confirmPasswordInput(): Locator {
    return this.passwordForm
      .getByPlaceholder(/请再次输入新密码|请确认新密码|confirm password/i)
      .first();
  }

  get submitPasswordButton(): Locator {
    return this.passwordForm
      .getByRole('button', {
        name: /修\s*改\s*密\s*码|更\s*新\s*密\s*码|Update password|Change password|Submit/i,
      })
      .first();
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
    await this.passwordForm.waitFor({ state: 'visible', timeout: 10000 });
  }

  async submitPasswordChange(oldPassword: string, newPassword: string) {
    await this.oldPasswordInput.fill(oldPassword);
    await this.newPasswordInput.fill(newPassword);
    await this.confirmPasswordInput.fill(newPassword);
    await this.submitPasswordButton.click();
  }
}
