import type { Locator, Page } from '@playwright/test';

import {
  waitForConfirmOverlay,
  waitForDialogReady,
  waitForRouteReady,
  waitForTableReady,
} from '../support/ui';

export class JobGroupPage {
  constructor(private page: Page) {}

  private get dialog(): Locator {
    return this.page.locator('[role="dialog"]').last();
  }

  async goto() {
    await this.page.goto('/system/job-group');
    await waitForTableReady(this.page, '[data-testid="job-group-page"]');
  }

  async fillSearchField(label: string, value: string) {
    const input = this.page.getByLabel(label, { exact: true }).first();
    await input.clear();
    await input.fill(value);
  }

  async clickSearch() {
    await this.page.getByRole('button', { name: /搜\s*索/ }).first().click();
    await waitForRouteReady(this.page);
  }

  async clickReset() {
    await this.page.getByRole('button', { name: /重\s*置/ }).first().click();
    await waitForRouteReady(this.page);
  }

  async createGroup(params: {
    code: string;
    name: string;
    remark?: string;
    sortOrder?: number;
  }) {
    await this.page.getByTestId('job-group-add').click();
    await waitForDialogReady(this.dialog);

    await this.dialog.getByLabel('分组编码').fill(params.code);
    await this.dialog.getByLabel('分组名称').fill(params.name);
    if (typeof params.sortOrder === 'number') {
      await this.dialog.getByRole('spinbutton', { name: '显示排序' }).fill(String(params.sortOrder));
    }
    if (params.remark) {
      await this.dialog.getByLabel('备注').fill(params.remark);
    }

    await this.dialog.getByRole('button', { name: /确\s*认/ }).click();
    await waitForRouteReady(this.page);
    await this.dialog.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {});
  }

  async editSearchedGroup(fields: {
    name?: string;
    remark?: string;
    sortOrder?: number;
  }) {
    await this.page.locator('[data-testid^="job-group-edit-"]').first().click();
    await waitForDialogReady(this.dialog);

    if (fields.name) {
      const input = this.dialog.getByLabel('分组名称');
      await input.clear();
      await input.fill(fields.name);
    }
    if (typeof fields.sortOrder === 'number') {
      const input = this.dialog.getByRole('spinbutton', { name: '显示排序' });
      await input.fill(String(fields.sortOrder));
    }
    if (fields.remark) {
      const input = this.dialog.getByLabel('备注');
      await input.clear();
      await input.fill(fields.remark);
    }

    await this.dialog.getByRole('button', { name: /确\s*认/ }).click();
    await waitForRouteReady(this.page);
    await this.dialog.waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {});
  }

  async deleteSearchedGroup() {
    await this.page.locator('[data-testid^="job-group-delete-"]').first().click();
    await this.confirmPopconfirm();
    await waitForRouteReady(this.page);
  }

  async isDefaultDeleteDisabled(code = 'default') {
    await this.fillSearchField('分组编码', code);
    await this.clickSearch();
    return this.page
      .locator('[data-testid^="job-group-delete-"]')
      .first()
      .isDisabled();
  }

  async hasGroup(text: string) {
    return this.page
      .locator('.vxe-body--row', { hasText: text })
      .first()
      .isVisible({ timeout: 3000 })
      .catch(() => false);
  }

  private async confirmPopconfirm() {
    const popconfirm = await waitForConfirmOverlay(this.page);
    await popconfirm.getByRole('button', { name: /确\s*定|OK|是/i }).click();
  }
}
