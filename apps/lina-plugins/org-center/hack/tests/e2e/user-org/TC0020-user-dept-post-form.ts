import { test, expect } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import { UserPage } from '@host-tests/pages/UserPage';
import {
  waitForBusyIndicatorsToClear,
  waitForDialogReady,
  waitForDropdown,
} from '@host-tests/support/ui';

test.describe('TC0020 用户表单部门岗位字段', () => {
  test.beforeEach(async ({ adminPage }) => {
    await ensureSourcePluginEnabled(adminPage, 'org-center');
  });

  test('TC0020a: 用户编辑表单包含部门和岗位字段', async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    // Click the "新增" button to open the user drawer
    await adminPage
      .getByRole('button', { name: /新\s*增/ })
      .click();

    // Wait for drawer to open
    const drawer = await waitForDialogReady(adminPage.locator('[role="dialog"]'));

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
    const drawer = await waitForDialogReady(adminPage.locator('[role="dialog"]'));

    // Set up request interception for post list when dept changes
    const requestPromise = adminPage.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/user/post-options') &&
        req.method() === 'GET',
      { timeout: 15000 },
    );

    // Click on the dept TreeSelect to open it
    const deptField = drawer.getByLabel('部门', { exact: false }).first();
    await deptField.click();
    const deptDropdown = await waitForDropdown(adminPage);

    // Select the first available dept node in the tree dropdown
    const deptOption = deptDropdown
      .locator('.ant-select-tree-node-content-wrapper')
      .first();
    const request = await Promise.all([
      requestPromise,
      deptOption.click(),
    ]).then(([capturedRequest]) => capturedRequest);
    await waitForBusyIndicatorsToClear(adminPage);

    // Verify that a post-related API request was triggered after dept selection
    expect(request.url()).toContain('/api/v1/user/post-options');
  });
});
