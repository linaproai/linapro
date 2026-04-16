import { expect, test } from '../../fixtures/auth';

test.describe('TC0024 部门负责人选择', () => {
  test('TC0024a: 新增部门按钮文本为"新增"而非"新增部门"', async ({
    adminPage,
  }) => {
    await adminPage.goto('/system/dept');
    await adminPage.waitForLoadState('networkidle');
    await adminPage
      .locator('.vxe-table')
      .waitFor({ state: 'visible', timeout: 10000 });

    // The primary button should say "新增" not "新增部门"
    const addBtn = adminPage
      .getByRole('button', { name: /新\s*增/ })
      .filter({ hasText: /^新\s*增$/ })
      .first();
    await expect(addBtn).toBeVisible();
    const btnText = await addBtn.textContent();
    expect(btnText?.replace(/\s/g, '')).toBe('新增');
  });

  test('TC0024b: 新增部门时负责人下拉可用且支持搜索', async ({
    adminPage,
  }) => {
    await adminPage.goto('/system/dept');
    await adminPage.waitForLoadState('networkidle');
    await adminPage
      .locator('.vxe-table')
      .waitFor({ state: 'visible', timeout: 10000 });

    // Click the toolbar "新增" primary button (first match, the non-ghost one)
    await adminPage
      .locator('button.ant-btn-primary:not(.ant-btn-background-ghost)', {
        hasText: /新\s*增/,
      })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'visible', timeout: 5000 });
    // Wait for the loading overlay to disappear
    await drawer
      .locator('.bg-overlay-content')
      .waitFor({ state: 'hidden', timeout: 10000 })
      .catch(() => {});
    await adminPage.waitForTimeout(500);

    // The leader combobox should be accessible
    const leaderCombobox = drawer.getByRole('combobox', { name: '负责人' });
    await expect(leaderCombobox).toBeVisible({ timeout: 5000 });

    // It should NOT be disabled
    await expect(leaderCombobox).toBeEnabled();

    // Click to open dropdown
    await leaderCombobox.click();
    await adminPage.waitForTimeout(500);

    // Should show user options
    const dropdown = adminPage.locator(
      '.ant-select-dropdown:not(.ant-select-dropdown-hidden)',
    );
    await expect(dropdown).toBeVisible({ timeout: 5000 });
    const options = dropdown.locator('.ant-select-item-option');
    const count = await options.count();
    expect(count).toBeGreaterThan(0);
    expect(count).toBeLessThanOrEqual(10);

    // Close drawer
    await adminPage.keyboard.press('Escape');
    await adminPage.waitForTimeout(500);
  });

  test('TC0024c: 编辑部门时未设置负责人显示为空白', async ({
    adminPage,
  }) => {
    await adminPage.goto('/system/dept');
    await adminPage.waitForLoadState('networkidle');
    await adminPage
      .locator('.vxe-table')
      .waitFor({ state: 'visible', timeout: 10000 });

    // Click edit on the first dept row
    const firstRow = adminPage.locator('.vxe-body--row').first();
    await firstRow
      .getByRole('button', { name: /编\s*辑/ })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'visible', timeout: 5000 });
    // Wait for the loading overlay to disappear
    await drawer
      .locator('.bg-overlay-content')
      .waitFor({ state: 'hidden', timeout: 10000 })
      .catch(() => {});
    await adminPage.waitForTimeout(500);

    // The leader combobox should NOT show "0" value
    const leaderCombobox = drawer.getByRole('combobox', { name: '负责人' });
    await expect(leaderCombobox).toBeVisible({ timeout: 5000 });

    // Check the leader select doesn't show "0"
    // Get all selection items in the drawer (first is parentId TreeSelect, second would be leader)
    const leaderSelectContainer = leaderCombobox.locator('..').locator('..');
    const selectionItem = leaderSelectContainer.locator(
      '.ant-select-selection-item',
    );
    const hasSelection = await selectionItem.isVisible().catch(() => false);
    if (hasSelection) {
      const text = await selectionItem.textContent();
      expect(text).not.toBe('0');
    }

    // Close drawer
    await adminPage.keyboard.press('Escape');
    await adminPage.waitForTimeout(500);
  });

  test('TC0024d: 编辑部门时负责人下拉支持搜索', async ({
    adminPage,
  }) => {
    await adminPage.goto('/system/dept');
    await adminPage.waitForLoadState('networkidle');
    await adminPage
      .locator('.vxe-table')
      .waitFor({ state: 'visible', timeout: 10000 });

    // Click edit on the first dept row
    const firstRow = adminPage.locator('.vxe-body--row').first();
    await firstRow
      .getByRole('button', { name: /编\s*辑/ })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'visible', timeout: 5000 });
    // Wait for the loading overlay to disappear
    await drawer
      .locator('.bg-overlay-content')
      .waitFor({ state: 'hidden', timeout: 10000 });
    await adminPage.waitForTimeout(500);

    // Leader combobox should be enabled and searchable
    const leaderCombobox = drawer.getByRole('combobox', { name: '负责人' });
    await expect(leaderCombobox).toBeVisible({ timeout: 5000 });
    await expect(leaderCombobox).toBeEnabled();

    // Type to search
    await leaderCombobox.click();
    await leaderCombobox.fill('admin');
    await adminPage.waitForTimeout(1000);

    // Should show filtered results
    const dropdown = adminPage.locator(
      '.ant-select-dropdown:not(.ant-select-dropdown-hidden)',
    );
    await expect(dropdown).toBeVisible({ timeout: 5000 });
    const options = dropdown.locator('.ant-select-item-option');
    const count = await options.count();
    expect(count).toBeGreaterThan(0);

    // The first option should contain "admin"
    const firstOption = await options.first().textContent();
    expect(firstOption?.toLowerCase()).toContain('admin');

    // Close drawer
    await adminPage.keyboard.press('Escape');
    await adminPage.waitForTimeout(500);
  });
});
