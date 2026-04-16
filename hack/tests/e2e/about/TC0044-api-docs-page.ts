import { test, expect } from '../../fixtures/auth';

test.describe('TC0044 系统接口页面', () => {
  test('TC0044a: 系统接口页面通过 iframe 加载 Stoplight Elements', async ({
    adminPage,
  }) => {
    await adminPage.goto('/about/api-docs');
    // Verify the iframe is visible
    const iframe = adminPage.locator('iframe.api-docs-iframe');
    await expect(iframe).toBeVisible({ timeout: 10_000 });
    // Wait for Stoplight Elements to render inside the iframe
    const frame = adminPage.frameLocator('iframe.api-docs-iframe');
    const apiElement = frame.locator('elements-api');
    await expect(apiElement).toBeAttached({ timeout: 15_000 });
    // Verify Overview is visible in sidebar
    await expect(frame.getByText('Overview')).toBeVisible({ timeout: 15_000 });
    // Verify ENDPOINTS section is visible
    await expect(frame.getByText('ENDPOINTS')).toBeVisible();
  });

  test('TC0044b: 系统接口页面不污染主页面样式', async ({ adminPage }) => {
    await adminPage.goto('/about/api-docs');
    const iframe = adminPage.locator('iframe.api-docs-iframe');
    await expect(iframe).toBeVisible({ timeout: 10_000 });
    // Main page should not have any Stoplight stylesheets injected
    const stoplightLinks = await adminPage
      .locator('link[href*="stoplight/styles"]')
      .count();
    expect(stoplightLinks).toBe(0);
  });

  test('TC0044c: Overview 页面显示 API 标题和描述', async ({ adminPage }) => {
    await adminPage.goto('/about/api-docs');
    const frame = adminPage.frameLocator('iframe.api-docs-iframe');
    // Wait for content to load
    await expect(frame.getByText('Overview')).toBeVisible({ timeout: 15_000 });
    // Verify the right panel shows API title and description
    await expect(
      frame.locator('h1', { hasText: 'Lina Framework API' }),
    ).toBeVisible({ timeout: 10_000 });
    await expect(frame.getByText('v0.5.0')).toBeVisible();
  });

  test('TC0044d: 隐藏 powered by Stoplight 标识', async ({ adminPage }) => {
    await adminPage.goto('/about/api-docs');
    const frame = adminPage.frameLocator('iframe.api-docs-iframe');
    await expect(frame.getByText('Overview')).toBeVisible({ timeout: 15_000 });
    // "powered by Stoplight" link should be hidden
    const poweredBy = frame.locator('a', { hasText: 'Stoplight' });
    await expect(poweredBy).toBeHidden();
  });

  test('TC0044e: 模块名称粗体、接口名称非粗体', async ({
    adminPage,
  }) => {
    await adminPage.goto('/about/api-docs');
    const frame = adminPage.frameLocator('iframe.api-docs-iframe');
    await expect(frame.getByText('Overview')).toBeVisible({ timeout: 15_000 });
    // Click on a module to expand it
    const moduleItem = frame.locator('[title="认证管理"]');
    await moduleItem.click();
    // Module name should be bold (font-weight 700)
    const moduleText = frame
      .locator('[title="认证管理"] .sl-flex-1')
      .first();
    await expect(moduleText).toHaveCSS('font-weight', '700');
    // Endpoint name should not be bold (font-weight 400)
    // Use "用户登录" endpoint which exists in auth module
    const endpointText = frame
      .locator('[title="用户登录"] .sl-flex-1')
      .first();
    await expect(endpointText).toBeVisible();
    const fontWeight = await endpointText.evaluate(
      (el) => window.getComputedStyle(el).fontWeight,
    );
    expect(fontWeight).toBe('400');
  });

  test('TC0044f: 接口地址背景块全宽展示，GET 在左、路径在右', async ({
    adminPage,
  }) => {
    await adminPage.goto('/about/api-docs');
    const frame = adminPage.frameLocator('iframe.api-docs-iframe');
    await expect(frame.getByText('Overview')).toBeVisible({ timeout: 15_000 });
    // Expand module and click endpoint
    await frame.locator('[title="用户管理"]').click();
    await frame.locator('[title="获取用户列表"]').click();
    // Find the method/path block
    const pathBlock = frame.locator(
      'div[title*="/api/v1/user"]',
    );
    await expect(pathBlock).toBeVisible({ timeout: 10_000 });
    // Block should be full width (display: flex, not inline-flex)
    await expect(pathBlock).toHaveCSS('display', 'flex');
    await expect(pathBlock).toHaveCSS('width', /.+/);
    await expect(pathBlock).toHaveCSS('justify-content', 'space-between');
  });

  test('TC0044g: SCHEMAS 区域默认折叠且可展开', async ({ adminPage }) => {
    await adminPage.goto('/about/api-docs');
    const frame = adminPage.frameLocator('iframe.api-docs-iframe');
    await expect(frame.getByText('Overview')).toBeVisible({ timeout: 15_000 });
    // SCHEMAS header should be visible
    const schemasHeader = frame.locator('.schemas-section-header');
    await expect(schemasHeader).toBeVisible();
    // Schema items should be hidden by default (collapsed)
    const firstSchema = frame
      .locator('.ElementsTableOfContentsItem[href*="/schemas/"]')
      .first();
    await expect(firstSchema).toBeHidden();
    // Click to expand
    await schemasHeader.click();
    await expect(firstSchema).toBeVisible();
    // Click again to collapse
    await schemasHeader.click();
    await expect(firstSchema).toBeHidden();
  });
});
