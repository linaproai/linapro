import type { Page } from '@playwright/test';

export class FilePage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/system/file');
    await this.page.waitForLoadState('networkidle');
    await this.page
      .locator('.vxe-table')
      .waitFor({ state: 'visible', timeout: 10000 });
  }

  /** Get the count of rows in the file list table */
  async getRowCount(): Promise<number> {
    const rows = this.page.locator(
      '.vxe-table--body-wrapper .vxe-body--row',
    );
    return await rows.count();
  }

  /** Check if a file with the given name exists in the table */
  async hasFile(originalName: string): Promise<boolean> {
    const cell = this.page.locator('.vxe-body--row').filter({
      hasText: originalName,
    });
    return (await cell.count()) > 0;
  }

  /** Click the file upload button to open upload modal */
  async openFileUploadModal() {
    await this.page
      .getByRole('button', { name: '文件上传' })
      .click();
    await this.page
      .locator('[role="dialog"]')
      .waitFor({ state: 'visible', timeout: 5000 });
  }

  /** Click the image upload button to open upload modal */
  async openImageUploadModal() {
    await this.page
      .getByRole('button', { name: '图片上传' })
      .click();
    await this.page
      .locator('[role="dialog"]')
      .waitFor({ state: 'visible', timeout: 5000 });
  }

  /** Delete a file row by original name */
  async deleteFile(originalName: string) {
    const row = this.page.locator('.vxe-body--row').filter({
      hasText: originalName,
    });
    await row.getByRole('button', { name: /删\s*除/ }).click();
    // Confirm popconfirm
    await this.page
      .getByRole('button', { name: /确\s*认|OK|Yes/i })
      .click();
  }
}
