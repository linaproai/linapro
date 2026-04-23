import type { Locator, Page } from '@playwright/test';

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
  await dialog.waitFor({ state: 'visible', timeout });
  await waitForBusyIndicatorsToClear(dialog, timeout);
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
