import { expect, type Locator, type Page } from '@playwright/test';

const busySelector = [
  '.ant-spin-spinning',
  '.ant-skeleton',
  '.vxe-loading',
  '.vxe-grid.is--loading',
  '.vxe-table.is--loading',
  '[aria-busy="true"]',
].join(', ');

export async function waitForBusyIndicatorsToClear(
  scope: Locator | Page,
  timeout = 10000,
) {
  await scope.locator(busySelector).first().waitFor({
    state: 'hidden',
    timeout,
  }).catch(() => {});
}

export async function waitForRouteReady(page: Page, timeout = 10000) {
  await page.waitForLoadState('domcontentloaded');
  await page.waitForLoadState('networkidle', { timeout }).catch(() => {});
  await waitForBusyIndicatorsToClear(page, timeout);
}

export async function waitForTableReady(
  page: Page,
  selector = '.vxe-table',
  timeout = 10000,
) {
  const table = page.locator(selector).first();
  await table.waitFor({ state: 'visible', timeout });
  await waitForRouteReady(page, timeout);
}

export async function waitForDialogReady(dialog: Locator, timeout = 10000) {
  const deadline = Date.now() + timeout;
  while (Date.now() < deadline) {
    const count = await dialog.count();
    for (let index = 0; index < count; index += 1) {
      const target = dialog.nth(index);
      if (await target.isVisible().catch(() => false)) {
        await waitForBusyIndicatorsToClear(target, timeout);
        return target;
      }
    }
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
  throw new Error(`Timed out waiting for dialog to become visible after ${timeout}ms`);
}

export async function waitForConfirmOverlay(page: Page, timeout = 5000) {
  const overlay = page
    .locator('.ant-popconfirm:visible, .ant-popover:visible, .ant-modal-confirm:visible')
    .first();
  await overlay.waitFor({ state: 'visible', timeout });
  return overlay;
}

export async function waitForDropdown(page: Page, timeout = 5000) {
  const dropdown = page.locator('.ant-select-dropdown:visible').last();
  await dropdown.waitFor({ state: 'visible', timeout });
  return dropdown;
}

export async function closeDialogWithEscape(
  page: Page,
  dialog: Locator,
  timeout = 5000,
) {
  await page.keyboard.press('Escape');
  await dialog.waitFor({ state: 'hidden', timeout }).catch(() => {});
  await waitForBusyIndicatorsToClear(page, timeout);
}

export async function waitForUploadReady(
  scope: Locator | Page,
  timeout = 10000,
) {
  const uploadItem = scope.locator('.ant-upload-list-item').last();
  await uploadItem.waitFor({ state: 'visible', timeout });
  await scope
    .locator('.ant-upload-list-item.ant-upload-list-item-uploading')
    .first()
    .waitFor({ state: 'hidden', timeout })
    .catch(() => {});
  await waitForBusyIndicatorsToClear(scope, timeout);
  return uploadItem;
}

export async function setSwitchChecked(
  switchEl: Locator,
  checked: boolean,
  timeout = 5000,
) {
  const expected = checked ? 'true' : 'false';
  if ((await switchEl.getAttribute('aria-checked')) !== expected) {
    await switchEl.click();
  }
  await expect(switchEl).toHaveAttribute('aria-checked', expected, { timeout });
}

export async function dismissResultDialog(
  page: Page,
  title: string | RegExp,
  timeout = 2000,
) {
  const dialog = page.locator('.ant-modal-wrap:visible').filter({ hasText: title }).last();
  const appeared = await dialog
    .waitFor({ state: 'visible', timeout })
    .then(() => true)
    .catch(() => false);
  if (!appeared) {
    return false;
  }
  await dialog
    .getByRole('button', { name: /确\s*定|OK|知道了/i })
    .last()
    .click();
  await dialog.waitFor({ state: 'hidden', timeout: 5000 }).catch(() => {});
  await waitForBusyIndicatorsToClear(page, 5000);
  return true;
}
