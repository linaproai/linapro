import type { Locator, Page } from '@playwright/test';

import {
  waitForBusyIndicatorsToClear,
  waitForRouteReady,
  waitForTableReady,
} from '../support/ui';

export class RoleAuthUserPage {
  constructor(private page: Page) {}

  private get usernameSearchInput(): Locator {
    return this.page.getByLabel(/用户账号|User Account/i).first();
  }

  userRow(username: string): Locator {
    return this.page.locator('.vxe-body--row:visible', { hasText: username }).first();
  }

  async goto(roleId: number) {
    await this.page.goto(`/system/role-auth/user/${roleId}`);
    await waitForTableReady(this.page);
    await this.usernameSearchInput.waitFor({ state: 'visible', timeout: 10000 });
  }

  async searchByUsername(username: string) {
    await this.page.getByRole('button', { name: /重\s*置|Reset/i }).first().click();
    await waitForRouteReady(this.page);
    await waitForBusyIndicatorsToClear(this.page);
    await this.usernameSearchInput.clear();
    await this.usernameSearchInput.fill(username);
    await this.page.getByRole('button', { name: /搜\s*索|Search/i }).first().click();
    await waitForRouteReady(this.page);
    await waitForBusyIndicatorsToClear(this.page);
  }
}
