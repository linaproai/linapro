import { test, expect } from '../../fixtures/auth';

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

test.describe('TC0026 操作日志列表查询与筛选', () => {
  test.beforeEach(async ({ adminPage }) => {
    // Navigate to operlog page and wait for data to load
    const responsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/api/v1/operlog') &&
        res.request().method() === 'GET' &&
        res.status() === 200,
      { timeout: 15000 },
    );
    await adminPage.goto('/monitor/operlog');
    await responsePromise;
    await adminPage.waitForTimeout(500);
  });

  test('TC0026a: 操作日志页面加载并展示表格', async ({ adminPage }) => {
    // Table should be visible
    await expect(adminPage.locator('.vxe-table')).toBeVisible();

    // Should have toolbar with clean and export buttons
    await expect(
      adminPage.getByRole('button', { name: /清\s*空/ }),
    ).toBeVisible();
    await expect(
      adminPage.getByRole('button', { name: /导\s*出/ }),
    ).toBeVisible();
    await expectPageHeightStable(adminPage, '操作日志页');
  });

  test('TC0026b: 按模块名称搜索', async ({ adminPage }) => {
    // Fill module name search field
    await adminPage.getByLabel('模块名称', { exact: true }).first().fill('认证');

    // Wait for API request containing the search param
    const requestPromise = adminPage.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/operlog') &&
        req.method() === 'GET' &&
        req.url().includes('title='),
      { timeout: 10000 },
    );

    await adminPage.getByRole('button', { name: /搜\s*索/ }).first().click();
    const request = await requestPromise;

    expect(request.url()).toContain('title=');
    await adminPage.waitForTimeout(500);
  });

  test('TC0026c: 按操作人员搜索', async ({ adminPage }) => {
    // Fill operator name search field
    await adminPage
      .getByLabel('操作人员', { exact: true })
      .first()
      .fill('admin');

    const requestPromise = adminPage.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/operlog') &&
        req.method() === 'GET' &&
        req.url().includes('operName='),
      { timeout: 10000 },
    );

    await adminPage.getByRole('button', { name: /搜\s*索/ }).first().click();
    const request = await requestPromise;

    expect(request.url()).toContain('operName=');
    await adminPage.waitForTimeout(500);
  });

  test('TC0026d: 按操作类型下拉选择搜索', async ({ adminPage }) => {
    // The operType Select should have dict options loaded
    const operTypeSelect = adminPage
      .locator('.ant-select')
      .filter({
        has: adminPage.locator('[id*="operType"]'),
      })
      .first();

    // If no id-based selector works, use label association
    const selectTrigger = adminPage
      .getByLabel('操作类型', { exact: true })
      .first();
    await selectTrigger.click();
    await adminPage.waitForTimeout(300);

    // Should see dropdown options from dict
    const dropdown = adminPage.locator('.ant-select-dropdown:visible');
    await expect(dropdown).toBeVisible();

    // Select the first available option
    const firstOption = dropdown.locator('.ant-select-item-option').first();
    const optionText = await firstOption.textContent();
    await firstOption.click();
    await adminPage.waitForTimeout(300);

    // Click search and verify request includes operType parameter
    const requestPromise = adminPage.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/operlog') &&
        req.method() === 'GET' &&
        req.url().includes('operType='),
      { timeout: 10000 },
    );

    await adminPage.getByRole('button', { name: /搜\s*索/ }).first().click();
    const request = await requestPromise;

    expect(request.url()).toContain('operType=');
  });

  test('TC0026e: 按操作结果下拉选择搜索', async ({ adminPage }) => {
    // Click the status Select (labeled 操作结果)
    const selectTrigger = adminPage
      .getByLabel('操作结果', { exact: true })
      .first();
    await selectTrigger.click();
    await adminPage.waitForTimeout(300);

    // Should see dropdown options from dict
    const dropdown = adminPage.locator('.ant-select-dropdown:visible');
    await expect(dropdown).toBeVisible();

    // Select the first available option
    const firstOption = dropdown.locator('.ant-select-item-option').first();
    await firstOption.click();
    await adminPage.waitForTimeout(300);

    // Click search and verify request includes status parameter
    const requestPromise = adminPage.waitForRequest(
      (req) =>
        req.url().includes('/api/v1/operlog') &&
        req.method() === 'GET' &&
        req.url().includes('status='),
      { timeout: 10000 },
    );

    await adminPage.getByRole('button', { name: /搜\s*索/ }).first().click();
    const request = await requestPromise;

    expect(request.url()).toContain('status=');
  });

  test('TC0026f: 搜索后重置恢复全部数据', async ({ adminPage }) => {
    // Get initial row count
    const initialCount = await adminPage.locator('.vxe-body--row').count();

    // Fill a search field
    await adminPage
      .getByLabel('模块名称', { exact: true })
      .first()
      .fill('不存在的模块名称');
    await adminPage.getByRole('button', { name: /搜\s*索/ }).first().click();
    await adminPage.waitForLoadState('networkidle');
    await adminPage.waitForTimeout(500);

    // After filtering, row count should be different (likely 0)
    const filteredCount = await adminPage.locator('.vxe-body--row').count();

    // Reset
    const resetResponsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/api/v1/operlog') &&
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

  test('TC0026g: 搜索请求包含操作时间范围参数', async ({ adminPage }) => {
    // Click date range picker
    const rangePicker = adminPage
      .getByLabel('操作时间', { exact: true })
      .first();

    // Note: RangePicker typically has two inputs; clicking the label area
    // or the first input triggers it
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
        req.url().includes('/api/v1/operlog') &&
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
