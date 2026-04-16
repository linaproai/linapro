import { test, expect } from '../../fixtures/auth';
import { UserPage } from '../../pages/UserPage';

test.describe('TC0020 用户表单部门岗位字段', () => {
  test('TC0020a: 用户编辑表单包含部门和岗位字段', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Click the "新增" button to open the user drawer
    await adminPage
      .getByRole('button', { name: /新\s*增/ })
      .click();

    // Wait for drawer to open
    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'visible', timeout: 5000 });

    // Verify dept TreeSelect field exists
    const deptField = drawer.getByLabel('部门', { exact: false }).first();
    await expect(deptField).toBeVisible({ timeout: 5000 });

    // Verify post Select field exists
    const postField = drawer.getByLabel('岗位', { exact: false }).first();
    await expect(postField).toBeVisible({ timeout: 5000 });
  });

  test('TC0020b: 选择部门后岗位选项自动加载', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Click the "新增" button to open the user drawer
    await adminPage
      .getByRole('button', { name: /新\s*增/ })
      .click();

    // Wait for drawer to open
    const drawer = adminPage.locator('[role="dialog"]');
    await drawer.waitFor({ state: 'visible', timeout: 5000 });

    // Set up request interception for post list when dept changes
    const requestPromise = adminPage.waitForRequest(
      (req) => req.url().includes('/api/v1/post') && req.method() === 'GET',
      { timeout: 15000 },
    );

    // Click on the dept TreeSelect to open it
    const deptField = drawer.getByLabel('部门', { exact: false }).first();
    await deptField.click();
    await adminPage.waitForTimeout(300);

    // Select the first available dept node in the tree dropdown
    const deptOption = adminPage
      .locator('.ant-select-tree-node-content-wrapper')
      .first();
    await deptOption.click();
    await adminPage.waitForTimeout(500);

    // Verify that a post-related API request was triggered after dept selection
    const request = await requestPromise;
    expect(request.url()).toContain('/api/v1/post');
  });
});
