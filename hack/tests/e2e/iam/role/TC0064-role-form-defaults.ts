import { test, expect } from '../../../fixtures/auth';
import { RolePage } from '../../../pages/RolePage';

/**
 * TC0064 角色表单默认值测试
 *
 * 验证新增角色表单的默认值配置是否正确
 */
test.describe('TC0064 角色表单默认值', () => {
  test('TC0064a: 验证新增角色表单默认值', async ({ adminPage }) => {
    const rolePage = new RolePage(adminPage);
    await rolePage.goto();

    const drawer = await rolePage.openCreateDrawer();

    // 1. 验证排序字段默认值为 0
    const sortInput = drawer.getByRole('spinbutton');
    const sortValue = await sortInput.inputValue();
    console.log('Sort input value:', sortValue);
    expect(sortValue).toBe('0');

    // 2. 验证数据权限默认选中"全部数据权限"
    // 使用 hasText 精确匹配"全部数据权限"选项
    const dataScopeRadio = drawer.locator('.ant-radio-button-wrapper-checked').filter({ hasText: '全部数据权限' });
    await expect(dataScopeRadio).toBeVisible({ timeout: 3000 });
  });
});
