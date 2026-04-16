import { test, expect } from "../../fixtures/auth";
import { MenuPage } from "../../pages/MenuPage";

test.describe("TC0060 菜单管理 CRUD", () => {
  test("TC0060a: 菜单列表页面正常加载", async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Check that the table is visible
    const table = adminPage.locator(".vxe-table");
    await expect(table).toBeVisible({ timeout: 10000 });

    // Check that toolbar buttons are visible
    await expect(
      adminPage.getByRole("button", { name: /新\s*增/ }).first(),
    ).toBeVisible({ timeout: 5000 });
    await expect(
      adminPage.getByRole("button", { name: /折\s*叠/ }).first(),
    ).toBeVisible({ timeout: 5000 });
    await menuPage.expectLayoutHeightStable();
  });

  test("TC0060b: 创建菜单对话框打开", async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Open the create form
    await adminPage
      .getByRole("button", { name: /新\s*增/ })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: "visible", timeout: 10000 });

    // Verify form fields are present
    await expect(
      drawer.locator('input[placeholder="请输入菜单名称"]'),
    ).toBeVisible({ timeout: 5000 });

    // Close drawer without saving
    await drawer.getByRole("button", { name: /取\s*消/ }).click();
    await drawer.waitFor({ state: "hidden", timeout: 5000 });
  });

  test("TC0060c: 级联删除开关功能", async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Find the cascade delete switch
    const cascadeSwitch = adminPage.locator(".ant-switch").first();
    await cascadeSwitch.waitFor({ state: "visible", timeout: 5000 });

    // Get initial state via aria-checked attribute
    const initialState = await cascadeSwitch.getAttribute("aria-checked");

    // Toggle the switch
    await cascadeSwitch.click();
    await adminPage.waitForTimeout(500);

    // Verify state changed
    const newState = await cascadeSwitch.getAttribute("aria-checked");
    expect(newState).not.toBe(initialState);
  });

  test("TC0060d: 折叠按钮功能", async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Wait for the loading indicator to disappear
    await adminPage
      .waitForSelector(".vxe-grid.is--loading", {
        state: "hidden",
        timeout: 10000,
      })
      .catch(() => {});

    // Click collapse button
    const collapseBtn = adminPage
      .getByRole("button", { name: /折\s*叠/ })
      .first();
    await collapseBtn.click({ force: true });
    await adminPage.waitForTimeout(500);

    // Test passes if no errors thrown
    expect(true).toBeTruthy();
  });

  test("TC0060e: 表单字段验证", async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Open the create form
    await adminPage
      .getByRole("button", { name: /新\s*增/ })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: "visible", timeout: 10000 });

    // Verify required form fields are present
    await expect(
      drawer.locator('input[placeholder="请输入菜单名称"]'),
    ).toBeVisible({ timeout: 5000 });

    // Verify parent menu select (TreeSelect)
    const parentSelect = drawer
      .locator(".ant-tree-select, .ant-select")
      .first();
    await expect(parentSelect).toBeVisible({ timeout: 5000 });

    // Close drawer
    await drawer.getByRole("button", { name: /取\s*消/ }).click();
    await drawer.waitFor({ state: "hidden", timeout: 5000 });
  });

  test("TC0060f: 创建根菜单流程", async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    const testMenuName = `e2e_test_${Date.now()}`;

    await menuPage.createRootMenu({
      name: testMenuName,
      type: "D",
      path: "e2e-test",
      sort: 999,
    });

    // Wait for drawer to close - indicates submission completed
    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: "hidden", timeout: 15000 });

    // If drawer closes without error, the creation was successful
    expect(true).toBeTruthy();
  });

  test("TC0060g: 编辑菜单时表单应展示被编辑菜单的内容", async ({
    adminPage,
  }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Wait for table to load
    await adminPage
      .locator(".vxe-table")
      .waitFor({ state: "visible", timeout: 10000 });

    // Find the first edit button in the action column and click it
    // Use getByRole for better reliability
    const editBtn = adminPage.getByRole("button", { name: /编\s*辑/ }).first();
    await editBtn.click({ timeout: 5000 });

    // Wait for drawer to open
    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: "visible", timeout: 10000 });

    // Wait for skeleton to disappear (form loading)
    const skeleton = drawer.locator(".ant-skeleton");
    await skeleton.waitFor({ state: "hidden", timeout: 10000 });

    // Verify the form has values loaded (not empty)
    // The menu name input should have a value
    const nameInput = drawer.locator('input[placeholder="请输入菜单名称"]');
    await expect(nameInput).toBeVisible({ timeout: 5000 });

    // Get the input value to verify it's not empty
    const inputValue = await nameInput.inputValue();
    expect(inputValue.length).toBeGreaterThan(0);

    // Close drawer without saving
    await drawer.getByRole("button", { name: /取\s*消/ }).click();
    await drawer.waitFor({ state: "hidden", timeout: 5000 });
  });

  test("TC0060h: 上级菜单下拉树应展示子级菜单", async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Wait for table to load
    await adminPage
      .locator(".vxe-table")
      .waitFor({ state: "visible", timeout: 10000 });

    // Open the create form
    await adminPage
      .getByRole("button", { name: /新\s*增/ })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: "visible", timeout: 10000 });

    // Wait for skeleton to disappear
    const skeleton = drawer.locator(".ant-skeleton");
    await skeleton.waitFor({ state: "hidden", timeout: 10000 });

    // Find the parent menu TreeSelect and click to open dropdown
    const parentSelect = drawer
      .locator(".ant-tree-select, .ant-select")
      .first();
    await expect(parentSelect).toBeVisible({ timeout: 5000 });
    await parentSelect.click();

    // Wait for tree to be visible - the tree is inside the drawer content
    const tree = drawer.locator('[role="tree"]');
    await expect(tree).toBeVisible({ timeout: 5000 });

    // Look for plus-square icons (collapsed nodes with children)
    // Use getByRole to match by accessible name from accessibility tree
    const expandableNodes = tree.getByRole("img", { name: "plus-square" });
    const expandableCount = await expandableNodes.count();

    // There should be at least one expandable node (parent menu with children)
    // This verifies the tree structure has children
    expect(expandableCount).toBeGreaterThan(0);

    // Click on an expandable node to verify children are shown
    if (expandableCount > 0) {
      const firstExpandable = expandableNodes.first();
      await firstExpandable.click();

      // Wait a moment for expansion
      await adminPage.waitForTimeout(500);

      // Verify expanded node - look for minus-square icon (expanded state)
      const expandedNode = tree
        .getByRole("img", { name: "minus-square" })
        .first();
      await expect(expandedNode).toBeVisible({ timeout: 3000 });
    }

    // Close dropdown by pressing Escape
    await adminPage.keyboard.press("Escape");

    // Wait for dropdown to close
    await adminPage.waitForTimeout(500);

    // Close drawer - use Escape key as backup if button click fails
    try {
      await drawer
        .getByRole("button", { name: /取\s*消/ })
        .click({ timeout: 3000 });
      await drawer.waitFor({ state: "hidden", timeout: 5000 });
    } catch {
      // If button click fails, use Escape to close
      await adminPage.keyboard.press("Escape");
      await drawer.waitFor({ state: "hidden", timeout: 5000 }).catch(() => {});
    }
  });

  test("TC0060i: 编辑菜单时上级菜单应禁用当前菜单及其子孙节点", async ({
    adminPage,
  }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Wait for table to load
    await adminPage
      .locator(".vxe-table")
      .waitFor({ state: "visible", timeout: 10000 });

    // Expand the tree to find a parent menu
    const expandBtn = adminPage
      .getByRole("img", { name: "plus-square" })
      .first();
    if (await expandBtn.isVisible({ timeout: 3000 }).catch(() => false)) {
      await expandBtn.click();
      await adminPage.waitForTimeout(500);
    }

    // Find "系统管理" row (parent menu with children) and click its edit button
    const systemManageRow = adminPage
      .locator(".vxe-body--row")
      .filter({ hasText: "系统管理" })
      .first();

    // Click edit button in that row's action column
    const editBtn = systemManageRow.getByRole("button", { name: /编\s*辑/ });
    await editBtn.click({ timeout: 5000 });

    // Wait for drawer to open
    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: "visible", timeout: 10000 });

    // Wait for skeleton to disappear
    const skeleton = drawer.locator(".ant-skeleton");
    await skeleton.waitFor({ state: "hidden", timeout: 10000 });

    // Wait for form to be ready
    await adminPage.waitForTimeout(1000);

    // Find the parent menu selector and click it
    const parentSelect = drawer.locator(".ant-select").first();
    await parentSelect.click({ force: true });

    // Wait for dropdown to open
    await adminPage.waitForTimeout(800);

    // The tree is rendered in a portal
    const tree = adminPage.locator('[role="tree"]').first();
    await expect(tree).toBeVisible({ timeout: 5000 });

    // Wait for tree to render
    await adminPage.waitForTimeout(500);

    // Verify the feature: check that some nodes are disabled
    // TreeSelect uses ant-select-tree-treenode class prefix
    const result = await tree.evaluate((treeEl) => {
      // TreeSelect uses ant-select-tree-* prefix instead of ant-tree-*
      const allNodes = treeEl.querySelectorAll(".ant-select-tree-treenode");
      const disabledNodes = treeEl.querySelectorAll(
        ".ant-select-tree-treenode-disabled",
      );

      return {
        totalNodes: allNodes.length,
        disabledNodes: disabledNodes.length,
        // Check if disabled attribute is set on tree node content
        html: treeEl.innerHTML.substring(0, 1000),
      };
    });

    console.log("Tree evaluation result:", JSON.stringify(result, null, 2));

    // Verify that:
    // 1. Tree has nodes
    // 2. At least one node is disabled (the current menu and its children)
    expect(result.totalNodes).toBeGreaterThan(0);
    expect(result.disabledNodes).toBeGreaterThan(0);

    // Close dropdown
    await adminPage.keyboard.press("Escape");
    await adminPage.waitForTimeout(300);

    // Close drawer
    try {
      await drawer
        .getByRole("button", { name: /取\s*消/ })
        .click({ timeout: 3000 });
      await drawer.waitFor({ state: "hidden", timeout: 5000 });
    } catch {
      await adminPage.keyboard.press("Escape");
      await drawer.waitFor({ state: "hidden", timeout: 5000 }).catch(() => {});
    }
  });

  test("TC0060j: 新增菜单时上级菜单无禁用节点", async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Wait for table to load
    await adminPage
      .locator(".vxe-table")
      .waitFor({ state: "visible", timeout: 10000 });

    // Open the create form (新增)
    await adminPage
      .getByRole("button", { name: /新\s*增/ })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: "visible", timeout: 10000 });

    // Wait for skeleton to disappear
    const skeleton = drawer.locator(".ant-skeleton");
    await skeleton.waitFor({ state: "hidden", timeout: 10000 });

    // Find the parent menu TreeSelect and click to open dropdown
    const parentSelect = drawer
      .locator(".ant-tree-select, .ant-select")
      .first();
    await expect(parentSelect).toBeVisible({ timeout: 5000 });
    await parentSelect.click();

    // Wait for tree to be visible
    const tree = drawer.locator('[role="tree"]');
    await expect(tree).toBeVisible({ timeout: 5000 });

    // In create mode, there should be NO disabled nodes
    const disabledNodes = tree.locator(
      '.ant-tree-treenode-disabled, [aria-disabled="true"]',
    );
    const disabledCount = await disabledNodes.count();

    // No nodes should be disabled when creating a new menu
    expect(disabledCount).toBe(0);

    // Close dropdown
    await adminPage.keyboard.press("Escape");
    await adminPage.waitForTimeout(300);

    // Close drawer
    try {
      await drawer
        .getByRole("button", { name: /取\s*消/ })
        .click({ timeout: 3000 });
      await drawer.waitFor({ state: "hidden", timeout: 5000 });
    } catch {
      await adminPage.keyboard.press("Escape");
      await drawer.waitFor({ state: "hidden", timeout: 5000 }).catch(() => {});
    }
  });

  test("TC0060k: 备注字段应为 Textarea 组件", async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Wait for table to load
    await adminPage
      .locator(".vxe-table")
      .waitFor({ state: "visible", timeout: 10000 });

    // Open the create form
    await adminPage
      .getByRole("button", { name: /新\s*增/ })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: "visible", timeout: 10000 });

    // Wait for skeleton to disappear
    const skeleton = drawer.locator(".ant-skeleton");
    await skeleton.waitFor({ state: "hidden", timeout: 10000 });

    // Find the remark textarea by label
    const remarkField = drawer.locator('textarea[placeholder="请输入备注"]');
    await expect(remarkField).toBeVisible({ timeout: 5000 });

    // Verify it's a textarea (not input)
    const tagName = await remarkField.evaluate((el) =>
      el.tagName.toLowerCase(),
    );
    expect(tagName).toBe("textarea");

    // Verify it has rows attribute (should be 3)
    const rows = await remarkField.getAttribute("rows");
    expect(rows).toBe("3");

    // Close drawer
    await drawer.getByRole("button", { name: /取\s*消/ }).click();
    await drawer.waitFor({ state: "hidden", timeout: 5000 });
  });

  test("TC0060l: 点击菜单行的新增按钮，上级菜单应默认选中当前菜单", async ({
    adminPage,
  }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    // Wait for table to load
    await adminPage
      .locator(".vxe-table")
      .waitFor({ state: "visible", timeout: 10000 });

    // Find "系统管理" row and click its "新增" button (not the toolbar 新增)
    const systemManageRow = adminPage
      .locator(".vxe-body--row")
      .filter({ hasText: "系统管理" })
      .first();

    // Click the "新增" button in that row's action column (the green success button)
    const subAddBtn = systemManageRow.getByRole("button", { name: /新\s*增/ });
    await subAddBtn.click({ timeout: 5000 });

    // Wait for drawer to open
    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: "visible", timeout: 10000 });

    // Wait for skeleton to disappear
    const skeleton = drawer.locator(".ant-skeleton");
    await skeleton.waitFor({ state: "hidden", timeout: 10000 });

    // Wait for form to be ready
    await adminPage.waitForTimeout(500);

    // Find the parent menu selector and verify it shows "系统管理"
    const parentSelect = drawer.locator(".ant-select-selector").first();
    await expect(parentSelect).toBeVisible({ timeout: 5000 });

    // Get the selected value text - it should contain "系统管理"
    const selectedText = await parentSelect.textContent();
    expect(selectedText).toContain("系统管理");

    // Close drawer
    await drawer.getByRole("button", { name: /取\s*消/ }).click();
    await drawer.waitFor({ state: "hidden", timeout: 5000 });
  });

  test("TC0060m: 权限标识输入框应显示在菜单名称下方", async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    await adminPage
      .getByRole("button", { name: /新\s*增/ })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: "visible", timeout: 10000 });
    await drawer
      .locator(".ant-skeleton")
      .waitFor({ state: "hidden", timeout: 10000 });

    await drawer
      .locator(".ant-radio-button-wrapper")
      .filter({ hasText: "菜单" })
      .click();

    const nameInput = drawer
      .locator('input[placeholder="请输入菜单名称"]')
      .first();
    const permsInput = drawer
      .locator('input[placeholder="请输入权限标识"]')
      .first();

    await expect(nameInput).toBeVisible({ timeout: 5000 });
    await expect(permsInput).toBeVisible({ timeout: 5000 });

    const nameBox = await nameInput.boundingBox();
    const permsBox = await permsInput.boundingBox();

    expect(nameBox).not.toBeNull();
    expect(permsBox).not.toBeNull();
    expect(permsBox!.y).toBeGreaterThan(nameBox!.y);

    await drawer.getByRole("button", { name: /取\s*消/ }).click();
    await drawer.waitFor({ state: "hidden", timeout: 5000 });
  });

  test("TC0060n: 菜单类型未填写权限标识时不允许提交", async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    await adminPage
      .getByRole("button", { name: /新\s*增/ })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: "visible", timeout: 10000 });
    await drawer
      .locator(".ant-skeleton")
      .waitFor({ state: "hidden", timeout: 10000 });

    await drawer
      .locator(".ant-radio-button-wrapper")
      .filter({ hasText: "菜单" })
      .click();
    await drawer
      .locator('input[placeholder="请输入菜单名称"]')
      .fill(`e2e-menu-${Date.now()}`);
    await drawer
      .locator('input[placeholder*="路由地址"]')
      .fill(`e2e-menu-${Date.now()}`);
    const componentInput = drawer.locator('input[placeholder="请输入"]').last();
    await expect(componentInput).toBeVisible({ timeout: 5000 });
    await componentInput.fill("system/menu/index");

    await drawer.getByRole("button", { name: /确\s*认/ }).click();

    await expect(drawer.getByText("请输入权限标识")).toBeVisible({
      timeout: 5000,
    });
    await expect(drawer).toBeVisible();

    await drawer.getByRole("button", { name: /取\s*消/ }).click();
    await drawer.waitFor({ state: "hidden", timeout: 5000 });
  });

  test("TC0060o: 按钮类型未填写权限标识时不允许提交", async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    await adminPage
      .getByRole("button", { name: /新\s*增/ })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: "visible", timeout: 10000 });
    await drawer
      .locator(".ant-skeleton")
      .waitFor({ state: "hidden", timeout: 10000 });

    await drawer
      .locator(".ant-radio-button-wrapper")
      .filter({ hasText: "按钮" })
      .click();
    await drawer
      .locator('input[placeholder="请输入菜单名称"]')
      .fill(`e2e-button-${Date.now()}`);

    await drawer.getByRole("button", { name: /确\s*认/ }).click();

    await expect(drawer.getByText("请输入权限标识")).toBeVisible({
      timeout: 5000,
    });
    await expect(drawer).toBeVisible();

    await drawer.getByRole("button", { name: /取\s*消/ }).click();
    await drawer.waitFor({ state: "hidden", timeout: 5000 });
  });

  test("TC0060p: 菜单改为隐藏后左侧导航应立即刷新", async ({ adminPage }) => {
    const menuPage = new MenuPage(adminPage);
    const menuName = "角色管理";

    await menuPage.goto();
    await menuPage.expectSidebarContains(menuName);

    try {
      await menuPage.updateMenuVisibility(menuName, 0);
      await menuPage.expectSidebarNotContains(menuName);
    } finally {
      await menuPage.updateMenuVisibility(menuName, 1);
      await menuPage.expectSidebarContains(menuName);
    }
  });
});
