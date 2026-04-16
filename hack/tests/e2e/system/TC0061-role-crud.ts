import { test, expect } from '../../fixtures/auth';
import { RolePage } from '../../pages/RolePage';

async function expectPageHeightStable(page: any, pageName: string) {
  const samples: number[] = [];

  for (let index = 0; index < 4; index += 1) {
    samples.push(
      await page.evaluate(() => document.documentElement.scrollHeight),
    );
    if (index < 3) {
      await page.waitForTimeout(400);
    }
  }

  expect(
    Math.max(...samples) - Math.min(...samples),
    `${pageName}高度未稳定，采样结果: ${samples.join(', ')}`,
  ).toBeLessThanOrEqual(16);
}

/**
 * TC0061 角色管理 E2E 测试
 *
 * 测试覆盖：
 * - 角色列表页面加载
 * - 角色创建功能
 * - 角色编辑功能
 * - 角色删除功能
 * - 角色状态切换
 * - 角色菜单分配
 */
test.describe('TC0061 角色管理 CRUD', () => {
  const testRoleName = `e2e_role_${Date.now()}`;
  const testRoleCode = `e2e_role_code_${Date.now()}`;

  test('TC0061a: 角色列表页面正常加载', async ({ adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    // 验证表格可见
    const table = adminPage.locator('.vxe-table');
    await expect(table).toBeVisible({ timeout: 10000 });

    // 验证工具栏按钮可见
    await expect(
      adminPage.getByRole('button', { name: /新\s*增/ }).first(),
    ).toBeVisible({ timeout: 5000 });
    await expectPageHeightStable(adminPage, '角色管理页');
  });

  test('TC0061b: 创建角色对话框打开', async ({ adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    // 点击新增按钮
    await adminPage
      .getByRole('button', { name: /新\s*增/ })
      .first()
      .click();

    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'visible', timeout: 10000 });

    // Dismiss any tour overlay if present
    const endTourBtn = adminPage.getByRole('button', { name: '结束导览' });
    if (await endTourBtn.isVisible({ timeout: 1000 }).catch(() => false)) {
      await endTourBtn.click();
      await adminPage.waitForTimeout(500);
    }

    // 验证表单字段存在
    await expect(
      drawer.locator('input[placeholder="请输入角色名称"]'),
    ).toBeVisible({ timeout: 5000 });
    await expect(
      drawer.locator('input[placeholder="如: admin, user等"]'),
    ).toBeVisible({ timeout: 5000 });

    // Drawer will be closed by the test cleanup
  });

  test('TC0061c: 创建新角色', async ({ adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    await rolePage.createRole({
      name: testRoleName,
      code: testRoleCode,
      sort: 999,
      remark: 'E2E测试角色',
    });

    // 等待抽屉关闭表示提交完成
    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'hidden', timeout: 15000 });

    // 验证角色已创建
    await rolePage.searchRole(testRoleName);
    const hasRole = await rolePage.hasRole(testRoleName);
    expect(hasRole).toBeTruthy();
  });

  test('TC0061d: 编辑角色', async ({ adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    // 先确保测试角色存在
    await rolePage.searchRole(testRoleName);
    let hasRole = await rolePage.hasRole(testRoleName);
    if (!hasRole) {
      await rolePage.resetSearch();
      await rolePage.createRole({
        name: testRoleName,
        code: testRoleCode,
        sort: 999,
      });
      await adminPage.waitForTimeout(500);
    }

    // 编辑角色
    const newName = `${testRoleName}_edited`;
    await rolePage.editRole(testRoleName, newName);

    // 等待抽屉关闭
    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'hidden', timeout: 10000 });

    // 验证编辑成功
    await rolePage.goto();
    await rolePage.searchRole(newName);
    hasRole = await rolePage.hasRole(newName);
    expect(hasRole).toBeTruthy();

    // 更新测试变量以便后续测试使用
    (test as any).testRoleName = newName;
  });

  test('TC0061e: 角色状态切换', async ({ adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    // 使用编辑后的角色名
    const currentRoleName = (test as any).testRoleName || testRoleName;

    await rolePage.searchRole(currentRoleName);
    const hasRole = await rolePage.hasRole(currentRoleName);
    if (!hasRole) {
      // 角色不存在，跳过测试
      test.skip();
      return;
    }

    // 获取当前状态
    const switchEl = adminPage.locator('.vxe-body--row .ant-switch').first();
    const initialState = await switchEl.getAttribute('aria-checked');

    // 切换状态
    await rolePage.toggleStatus(currentRoleName);

    // 验证状态已改变
    const newState = await switchEl.getAttribute('aria-checked');
    expect(newState).not.toBe(initialState);

    // 恢复原状态
    await rolePage.toggleStatus(currentRoleName);
  });

  test('TC0061f: 角色菜单分配', async ({ adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    const currentRoleName = (test as any).testRoleName || testRoleName;

    await rolePage.searchRole(currentRoleName);
    let hasRole = await rolePage.hasRole(currentRoleName);
    if (!hasRole) {
      await rolePage.resetSearch();
      await rolePage.searchRole(currentRoleName);
      hasRole = await rolePage.hasRole(currentRoleName);
      if (!hasRole) {
        test.skip();
        return;
      }
    }

    // 编辑角色并分配菜单
    await rolePage.assignMenusToRole(currentRoleName, ['系统管理']);

    // 等待抽屉关闭
    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'hidden', timeout: 10000 });
  });

  test('TC0061g: 删除角色', async ({ adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    const currentRoleName = (test as any).testRoleName || testRoleName;

    // 先搜索确保角色存在
    await rolePage.searchRole(currentRoleName);
    let hasRole = await rolePage.hasRole(currentRoleName);
    if (!hasRole) {
      await rolePage.resetSearch();
      await rolePage.searchRole(currentRoleName);
      hasRole = await rolePage.hasRole(currentRoleName);
      if (!hasRole) {
        // 角色不存在，跳过
        test.skip();
        return;
      }
    }

    // 删除角色
    await rolePage.deleteRole(currentRoleName);

    // 验证角色已删除
    await rolePage.goto();
    await rolePage.searchRole(currentRoleName);
    hasRole = await rolePage.hasRole(currentRoleName);
    expect(hasRole).toBeFalsy();
  });

  test('TC0061h: 角色搜索功能', async ({ adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    // 搜索管理员角色
    await rolePage.searchRole('管理员');

    // 验证搜索结果
    const hasAdmin = await rolePage.hasRole('管理员');
    expect(hasAdmin).toBeTruthy();

    // 重置搜索
    await rolePage.resetSearch();

    // 验证重置后能看到更多角色
    const rowsBefore = await adminPage.locator('.vxe-body--row').count();
    await rolePage.searchRole('管理员');
    const rowsAfter = await adminPage.locator('.vxe-body--row').count();
    expect(rowsBefore).toBeGreaterThanOrEqual(rowsAfter);
  });

  test('TC0061i: 超级管理员角色不可编辑删除', async ({ adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    // 搜索超级管理员角色 (id=1)
    await rolePage.searchRole('超级管理员');

    // 验证超级管理员存在
    const hasAdmin = await rolePage.hasRole('超级管理员');
    expect(hasAdmin).toBeTruthy();

    // 验证状态开关被禁用
    const isDisabled = await rolePage.isStatusSwitchDisabled('超级管理员');
    expect(isDisabled).toBeTruthy();

    // 验证复选框被禁用
    const isCheckboxDisabled = await rolePage.isCheckboxDisabled('超级管理员');
    expect(isCheckboxDisabled).toBeTruthy();
  });
});
