import type { Locator, Page } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import { RolePage } from "../../../pages/RolePage";
import { waitForConfirmOverlay } from "../../../support/ui";

test.describe("TC-167 角色权限抽屉关闭交互", () => {
  test("TC-167a: 修改权限后取消关闭会提示并可确认关闭", async ({
    adminPage,
  }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();
    await adminPage.evaluate(() => {
      localStorage.removeItem("menu_select_fullscreen_read");
    });

    const drawer = await rolePage.openCreateDrawer({ keepTourOpen: true });
    await expectPermissionGuideOpen(adminPage);

    const selectedCount = drawer.getByTestId("menu-permission-selected-count");
    await expect(selectedCount).toBeVisible({ timeout: 5000 });

    const spacing = await drawer
      .getByTestId("menu-permission-toolbar")
      .evaluate((toolbar) => {
        const mode = toolbar.querySelector(
          '[data-testid="menu-permission-association-mode"]',
        );
        const count = toolbar.querySelector(
          '[data-testid="menu-permission-selected-count"]',
        );
        const modeRect = mode?.getBoundingClientRect();
        const countRect = count?.getBoundingClientRect();
        const columnGap = Number.parseFloat(
          getComputedStyle(toolbar).columnGap,
        );
        return {
          columnGap,
          gap: modeRect && countRect ? countRect.left - modeRect.right : 0,
          hasClass: count?.classList.contains("permission-selection-count"),
          toolbarWidth: toolbar.getBoundingClientRect().width,
        };
      });
    expect(spacing.hasClass).toBeTruthy();
    expect(spacing.columnGap).toBeGreaterThanOrEqual(16);
    expect(spacing.gap).toBeGreaterThanOrEqual(12);
    expect(spacing.toolbarWidth).toBeGreaterThan(0);

    await checkPermissionAndClose(adminPage, drawer, async () => {
      await drawer.getByRole("button", { name: /取\s*消|Cancel/i }).click();
    });

    await expect(drawer).toBeHidden({ timeout: 10000 });
  });

  test("TC-167b: 修改权限后关闭图标会提示并可确认关闭", async ({
    adminPage,
  }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();
    await adminPage.evaluate(() => {
      localStorage.removeItem("menu_select_fullscreen_read");
    });

    const drawer = await rolePage.openCreateDrawer({ keepTourOpen: true });
    await expectPermissionGuideOpen(adminPage);
    await checkPermissionAndClose(adminPage, drawer, async () => {
      await rolePage.clickDrawerCloseIcon(drawer);
    });

    await expect(drawer).toBeHidden({ timeout: 10000 });
  });

  test("TC-167c: 修改权限后点击遮罩会提示并可确认关闭", async ({
    adminPage,
  }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();
    await adminPage.evaluate(() => {
      localStorage.removeItem("menu_select_fullscreen_read");
    });

    const drawer = await rolePage.openCreateDrawer({ keepTourOpen: true });
    await expectPermissionGuideOpen(adminPage);
    await checkPermissionAndClose(adminPage, drawer, async () => {
      await rolePage.clickDrawerOverlay();
    });

    await expect(drawer).toBeHidden({ timeout: 10000 });
  });

  test("TC-167d: 编辑角色修改权限后关闭图标会提示并可确认关闭", async ({
    adminPage,
  }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();
    await adminPage.evaluate(() => {
      localStorage.removeItem("menu_select_fullscreen_read");
    });

    const drawer = await rolePage.openEditDrawer("admin", {
      keepTourOpen: true,
    });
    await expectPermissionGuideOpen(adminPage);
    await checkPermissionAndClose(adminPage, drawer, async () => {
      await rolePage.clickDrawerCloseIcon(drawer);
    });

    await expect(drawer).toBeHidden({ timeout: 10000 });
  });
});

async function expectPermissionGuideOpen(page: Page) {
  await expect(page.locator(".ant-tour:visible")).toBeVisible({
    timeout: 5000,
  });
  await expect(page.locator(".ant-tour-mask:visible svg")).toHaveCount(0);
}

async function checkPermissionAndClose(
  page: Page,
  drawer: Locator,
  closeAction: () => Promise<void>,
) {
  const permissionCheckbox = drawer
    .locator(".ant-checkbox-wrapper", {
      hasText: /查询|Search|新增|Add|编辑|Edit/,
    })
    .first();
  await permissionCheckbox.waitFor({ state: "visible", timeout: 10000 });
  await permissionCheckbox.click();

  await closeAction();
  const confirmOverlay = await waitForConfirmOverlay(page);
  await expect(confirmOverlay).toContainText(/未保存的修改|unsaved changes/i);
  await confirmOverlay
    .getByRole("button", { name: /确\s*认|OK|Confirm/i })
    .last()
    .click();
}
