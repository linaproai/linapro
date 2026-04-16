import { test, expect } from '../../fixtures/auth';

test.describe('TC0031 登录日志列表查询', () => {
  test.beforeEach(async ({ adminPage }) => {
    // Navigate to loginlog page and wait for data to load
    const responsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/api/v1/loginlog') &&
        res.request().method() === 'GET' &&
        res.status() === 200,
      { timeout: 15000 },
    );
    await adminPage.goto('/monitor/loginlog');
    await responsePromise;
    await adminPage.waitForTimeout(500);
  });

  test('TC0031a: 登录日志页面加载并展示表格', async ({ adminPage }) => {
    // Table should be visible
    await expect(adminPage.locator('.vxe-table')).toBeVisible();

    // Should have toolbar buttons
    await expect(
      adminPage.getByRole('button', { name: /清\s*空/ }),
    ).toBeVisible();
    await expect(
      adminPage.getByRole('button', { name: /导\s*出/ }),
    ).toBeVisible();
  });

  test('TC0031b: 登录日志包含admin用户记录', async ({ adminPage }) => {
    // Should see login log rows with admin
    const rows = adminPage.locator('.vxe-body--row');
    const count = await rows.count();
    if (count > 0) {
      // At least one row should contain 'admin'
      await expect(
        adminPage.locator('.vxe-body--row').first(),
      ).toContainText('admin');
    }
  });

  test('TC0031c: 按用户账号搜索', async ({ adminPage }) => {
    // Fill user account search field
    await adminPage
      .getByLabel('用户账号', { exact: true })
      .first()
      .fill('admin');

    const requestPromise = adminPage.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/loginlog') &&
        req.method() === 'GET' &&
        req.url().includes('userName='),
      { timeout: 10000 },
    );

    await adminPage.getByRole('button', { name: /搜\s*索/ }).first().click();
    const request = await requestPromise;

    expect(request.url()).toContain('userName=');
    await adminPage.waitForTimeout(500);

    // Results should still contain admin entries
    const rows = adminPage.locator('.vxe-body--row');
    const count = await rows.count();
    expect(count).toBeGreaterThan(0);
  });

  test('TC0031d: 按IP地址搜索', async ({ adminPage }) => {
    // Fill IP address search field
    await adminPage
      .getByLabel('IP地址', { exact: true })
      .first()
      .fill('127.0.0.1');

    const requestPromise = adminPage.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/loginlog') &&
        req.method() === 'GET' &&
        req.url().includes('ip='),
      { timeout: 10000 },
    );

    await adminPage.getByRole('button', { name: /搜\s*索/ }).first().click();
    const request = await requestPromise;

    expect(request.url()).toContain('ip=');
  });

  test('TC0031e: 按登录状态下拉选择搜索', async ({ adminPage }) => {
    // Click the status Select (labeled 登录状态)
    const selectTrigger = adminPage
      .getByLabel('登录状态', { exact: true })
      .first();
    await selectTrigger.click();
    await adminPage.waitForTimeout(300);

    // Should see dropdown options from dict
    const dropdown = adminPage.locator('.ant-select-dropdown:visible');
    await expect(dropdown).toBeVisible();

    // Select the first available option (e.g., "成功")
    const firstOption = dropdown.locator('.ant-select-item-option').first();
    await firstOption.click();
    await adminPage.waitForTimeout(300);

    // Click search and verify request includes status parameter
    const requestPromise = adminPage.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/loginlog') &&
        req.method() === 'GET' &&
        req.url().includes('status='),
      { timeout: 10000 },
    );

    await adminPage.getByRole('button', { name: /搜\s*索/ }).first().click();
    const request = await requestPromise;

    expect(request.url()).toContain('status=');
  });

  test('TC0031f: 搜索后重置恢复全部数据', async ({ adminPage }) => {
    // Get initial row count
    const initialCount = await adminPage.locator('.vxe-body--row').count();

    // Fill a search field with unlikely value
    await adminPage
      .getByLabel('用户账号', { exact: true })
      .first()
      .fill('不存在的用户名称');
    await adminPage.getByRole('button', { name: /搜\s*索/ }).first().click();
    await adminPage.waitForLoadState('networkidle');
    await adminPage.waitForTimeout(500);

    // Reset
    const resetResponsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/api/v1/loginlog') &&
        res.request().method() === 'GET' &&
        res.status() === 200,
      { timeout: 10000 },
    );
    await adminPage.getByRole('button', { name: /重\s*置/ }).first().click();
    await resetResponsePromise;
    await adminPage.waitForTimeout(500);

    // Row count should restore
    const resetCount = await adminPage.locator('.vxe-body--row').count();
    expect(resetCount).toBe(initialCount);
  });

  test('TC0031g: 搜索请求包含登录日期范围参数', async ({ adminPage }) => {
    // Click date range picker
    const rangeInputs = adminPage.locator(
      '.ant-picker-range input[placeholder]',
    );
    await rangeInputs.first().click();
    await adminPage.waitForTimeout(300);

    // Select today in the date picker popup
    const today = adminPage.locator('.ant-picker-cell-today').first();
    await today.click();
    await adminPage.waitForTimeout(200);
    // Select the same day as end date
    await today.click();
    await adminPage.waitForTimeout(300);

    // Click search and verify request includes time range params
    const requestPromise = adminPage.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/loginlog') &&
        req.method() === 'GET' &&
        (req.url().includes('beginTime=') ||
          req.url().includes('endTime=')),
      { timeout: 10000 },
    );

    await adminPage.getByRole('button', { name: /搜\s*索/ }).first().click();
    const request = await requestPromise;

    const url = request.url();
    expect(url).toContain('beginTime=');
    expect(url).toContain('endTime=');
  });
});
