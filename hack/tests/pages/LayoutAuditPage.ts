import type { Locator, Page } from '@playwright/test';

import { waitForRouteReady, waitForTableReady } from '../support/ui';

export class LayoutAuditPage {
  constructor(private page: Page) {}

  async goto(path: string, options?: { tableSelector?: string }) {
    await this.page.goto(path);
    if (options?.tableSelector) {
      await waitForTableReady(this.page, options.tableSelector);
      return;
    }
    await waitForRouteReady(this.page);
  }

  panel(id: string): Locator {
    return this.page.locator(`#${id}`).first();
  }

  formLabel(text: RegExp | string, scope?: Locator): Locator {
    return (scope ?? this.page)
      .locator(
        '.ant-form-item-label, .ant-form-item-label label, label',
        { hasText: text },
      )
      .first();
  }

  tableHeader(text: RegExp | string, scope?: Locator): Locator {
    return (scope ?? this.page)
      .locator('.vxe-header--column, th', { hasText: text })
      .first();
  }
}
