import { test, expect } from '../../fixtures/auth';
import { UserPage } from '../../pages/UserPage';

test.describe('TC0011 当前用户禁止自操作', () => {
  test('TC0011a: 当前用户行的操作按钮不可见', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // admin is the current logged-in user, action buttons should be hidden
    const hasActions = await userPage.hasActionButtons('admin');
    expect(hasActions).toBe(false);
  });

  test('TC0011b: 当前用户行的状态开关禁用', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    const isDisabled = await userPage.isStatusSwitchDisabled('admin');
    expect(isDisabled).toBe(true);
  });

  test('TC0011c: 当前用户行的复选框禁用', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    const isDisabled = await userPage.isCheckboxDisabled('admin');
    expect(isDisabled).toBe(true);
  });

  test('TC0011d: 工具栏删除按钮未勾选时置灰禁用', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Toolbar delete button is the non-ghost, non-small primary danger button
    const deleteBtn = adminPage.locator(
      'button.ant-btn-dangerous.ant-btn-primary:not(.ant-btn-sm):not(.ant-btn-background-ghost)',
    );
    await expect(deleteBtn).toBeVisible();
    await expect(deleteBtn).toBeDisabled();
  });
});
