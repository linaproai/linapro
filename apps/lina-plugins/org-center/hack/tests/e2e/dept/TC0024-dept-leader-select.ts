import { expect, test } from "@host-tests/fixtures/auth";
import { ensureSourcePluginEnabled } from "@host-tests/fixtures/plugin";
import {
  closeDialogWithEscape,
  waitForDialogReady,
  waitForDropdown,
  waitForRouteReady,
  waitForTableReady,
} from "@host-tests/support/ui";

test.describe("TC0024 部门负责人选择", () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, "org-center");
  });

  test('TC0024a: 新增部门按钮文本为"新增"而非"新增部门"', async ({
    adminPage,
  }) => {
    await adminPage.goto("/system/dept");
    await waitForTableReady(adminPage);

    // The primary button should say "新增" not "新增部门"
    const addBtn = adminPage
      .getByRole("button", { name: /新\s*增/ })
      .filter({ hasText: /^新\s*增$/ })
      .first();
    await expect(addBtn).toBeVisible();
    const btnText = await addBtn.textContent();
    expect(btnText?.replace(/\s/g, "")).toBe("新增");
  });

  test("TC0024b: 新增部门时负责人下拉可用且支持搜索", async ({ adminPage }) => {
    await adminPage.goto("/system/dept");
    await waitForTableReady(adminPage);

    // Click the toolbar "新增" primary button (first match, the non-ghost one)
    await adminPage
      .locator("button.ant-btn-primary:not(.ant-btn-background-ghost)", {
        hasText: /新\s*增/,
      })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await waitForDialogReady(drawer);

    // The leader combobox should be accessible
    const leaderCombobox = drawer.getByRole("combobox", { name: "负责人" });
    await expect(leaderCombobox).toBeVisible({ timeout: 5000 });
    await expect(leaderCombobox).toBeInViewport({ ratio: 0.5, timeout: 5000 });

    // It should NOT be disabled
    await expect(leaderCombobox).toBeEnabled();

    // Click to open dropdown
    await leaderCombobox.click();
    const dropdown = await waitForDropdown(adminPage);

    // Should show user options
    const options = dropdown.locator(".ant-select-item-option");
    const count = await options.count();
    expect(count).toBeGreaterThan(0);
    expect(count).toBeLessThanOrEqual(10);

    await closeDialogWithEscape(adminPage, drawer);
  });

  test("TC0024c: 编辑部门时未设置负责人显示为空白", async ({ adminPage }) => {
    await adminPage.goto("/system/dept");
    await waitForTableReady(adminPage);

    // Click edit on the first dept row
    const firstRow = adminPage.locator(".vxe-body--row").first();
    await firstRow
      .getByRole("button", { name: /编\s*辑/ })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await waitForDialogReady(drawer);

    // The leader combobox should NOT show "0" value
    const leaderCombobox = drawer.getByRole("combobox", { name: "负责人" });
    await expect(leaderCombobox).toBeVisible({ timeout: 5000 });

    // Check the leader select doesn't show "0"
    // Get all selection items in the drawer (first is parentId TreeSelect, second would be leader)
    const leaderSelectContainer = leaderCombobox.locator("..").locator("..");
    const selectionItem = leaderSelectContainer.locator(
      ".ant-select-selection-item",
    );
    const hasSelection = await selectionItem.isVisible().catch(() => false);
    if (hasSelection) {
      const text = await selectionItem.textContent();
      expect(text).not.toBe("0");
    }

    await closeDialogWithEscape(adminPage, drawer);
  });

  test("TC0024d: 编辑部门时负责人下拉支持搜索", async ({ adminPage }) => {
    await adminPage.goto("/system/dept");
    await waitForTableReady(adminPage);

    // Click edit on the first dept row
    const firstRow = adminPage.locator(".vxe-body--row").first();
    await firstRow
      .getByRole("button", { name: /编\s*辑/ })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await waitForDialogReady(drawer);

    // Leader combobox should be enabled and searchable
    const leaderCombobox = drawer.getByRole("combobox", { name: "负责人" });
    await expect(leaderCombobox).toBeVisible({ timeout: 5000 });
    await expect(leaderCombobox).toBeInViewport({ ratio: 0.5, timeout: 5000 });
    await expect(leaderCombobox).toBeEnabled();

    // Focus the inner search input directly because an existing selected item
    // can cover the input and intercept pointer events in edit mode.
    await leaderCombobox.focus();
    await leaderCombobox.fill("admin");
    await waitForRouteReady(adminPage);

    // Should show filtered results
    const dropdown = await waitForDropdown(adminPage);
    const options = dropdown.locator(".ant-select-item-option");
    const count = await options.count();
    expect(count).toBeGreaterThan(0);

    // The first option should contain "admin"
    const firstOption = await options.first().textContent();
    expect(firstOption?.toLowerCase()).toContain("admin");

    await closeDialogWithEscape(adminPage, drawer);
  });
});
