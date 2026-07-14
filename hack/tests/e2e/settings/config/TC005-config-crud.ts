import { test, expect } from '../../../fixtures/auth';
import { ConfigPage } from '../../../pages/ConfigPage';

test.describe('TC005 参数设置管理', () => {
  const testName = `测试参数_${Date.now()}`;
  const testKey = `test.param.${Date.now()}`;

  test('TC005a: 页面加载并展示数据列表', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    // Verify the page renders a non-empty table before interacting with filters.
    const rowCount = await configPage.getRowCount();
    expect(rowCount).toBeGreaterThanOrEqual(1);

    // Search for a stable seed config instead of assuming it stays on the first page.
    await configPage.fillSearchField('参数键名', 'sys.jwt.expire');
    await configPage.clickSearch();
    const hasJwtExpire = await configPage.hasConfig('sys.jwt.expire');
    expect(hasJwtExpire).toBeTruthy();
  });

  test('TC005b: 按参数名称搜索', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    await configPage.fillSearchField('参数名称', '用户登录');
    await configPage.clickSearch();

    const rowCount = await configPage.getRowCount();
    expect(rowCount).toBeGreaterThanOrEqual(1);

    const hasResult = await configPage.hasConfig('sys.login.blackIPList');
    expect(hasResult).toBeTruthy();
  });

  test('TC005c: 按参数键名搜索', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    await configPage.fillSearchField('参数键名', 'sys.login');
    await configPage.clickSearch();

    const rowCount = await configPage.getRowCount();
    expect(rowCount).toBeGreaterThanOrEqual(1);

    const hasResult = await configPage.hasConfig('sys.login.blackIPList');
    expect(hasResult).toBeTruthy();
  });

  test('TC005d: 重置搜索条件', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    // Search to narrow results
    await configPage.fillSearchField('参数名称', '用户登录');
    await configPage.clickSearch();
    const filteredCount = await configPage.getRowCount();

    // Reset and verify all data shows
    await configPage.clickReset();
    const allCount = await configPage.getRowCount();
    expect(allCount).toBeGreaterThanOrEqual(filteredCount);
  });

  test('TC005e: 创建新参数设置', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    await configPage.create(testName, testKey, 'test_value', '测试备注');

    await expect(adminPage.getByText(/创建成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC005f: 新创建的参数可搜索到', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    await configPage.fillSearchField('参数名称', testName);
    await configPage.clickSearch();

    const hasConfig = await configPage.hasConfig(testName);
    expect(hasConfig).toBeTruthy();
  });

  test('TC005g: 编辑参数设置', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    await configPage.edit(testName, { name: `${testName}_修改`, value: 'updated_value' });

    await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC005h: 删除参数设置', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    await configPage.delete(`${testName}_修改`);

    await expect(adminPage.getByText(/删除成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC005i: 导出按钮功能正常', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    // Click export button
    const exportBtn = adminPage.getByRole('button', { name: /导\s*出/ });
    await expect(exportBtn).toBeVisible();
    await exportBtn.click();

    // Verify modal appears
    const modalContent = adminPage.locator('.ant-modal-content');
    await expect(modalContent).toBeVisible({ timeout: 5000 });
    await expect(modalContent.getByText(/是否导出全部数据/)).toBeVisible();

    // Set up response listener
    const responsePromise = adminPage.waitForResponse(
      (resp) => resp.url().includes('config/export'),
      { timeout: 15000 }
    );

    // Click confirm button
    const confirmBtn = modalContent.getByRole('button', { name: /确\s*认/ });
    await confirmBtn.click();

    // Wait for response and verify
    const response = await responsePromise;
    expect(response.status()).toBe(200);
  });

  test('TC005j: 参数名称与参数键名列左对齐', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    const rowCount = await configPage.getRowCount();
    expect(rowCount).toBeGreaterThanOrEqual(1);

    const nameAlign = await configPage.getColumnAlignment('参数名称');
    expect(nameAlign.headerLeft).toBe(true);
    expect(nameAlign.bodyLeft).toBe(true);

    const keyAlign = await configPage.getColumnAlignment('参数键名');
    expect(keyAlign.headerLeft).toBe(true);
    expect(keyAlign.bodyLeft).toBe(true);
  });
});
