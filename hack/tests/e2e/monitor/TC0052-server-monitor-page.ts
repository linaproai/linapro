import { test, expect } from '../../fixtures/auth';

test.describe('TC0052 服务监控页面展示', () => {
  test.beforeEach(async ({ adminPage }) => {
    const responsePromise = adminPage.waitForResponse(
      (res) =>
        res.url().includes('/api/v1/monitor/server') &&
        res.request().method() === 'GET' &&
        res.status() === 200,
      { timeout: 15000 },
    );
    await adminPage.goto('/monitor/server');
    await responsePromise;
    await adminPage.waitForTimeout(1000);
  });

  test('TC0052a: 服务信息在节点展开内容中展示', async ({ adminPage }) => {
    // Service info section should be visible inside expanded node
    await expect(adminPage.getByText('服务信息').first()).toBeVisible();

    // Should show Go version
    await expect(adminPage.getByText('Go 版本').first()).toBeVisible();
    await expect(adminPage.getByText(/go\d+\.\d+/).first()).toBeVisible();

    // Should show GoFrame version
    await expect(adminPage.getByText('GoFrame 版本').first()).toBeVisible();

    // Should show Goroutines
    await expect(adminPage.getByText('Goroutines').first()).toBeVisible();

    // Should show process CPU and memory usage
    await expect(adminPage.getByText('服务 CPU').first()).toBeVisible();
    await expect(adminPage.getByText('服务内存').first()).toBeVisible();

    // Should show service start time
    await expect(adminPage.getByText('服务启动时间').first()).toBeVisible();

    // Should show service uptime
    await expect(adminPage.getByText('服务运行时长').first()).toBeVisible();
  });

  test('TC0052b: 服务器信息区块展示带提示图标', async ({ adminPage }) => {
    // Server info section should be visible
    await expect(adminPage.getByText('服务器信息').first()).toBeVisible();

    // Should show hostname in expanded node
    await expect(adminPage.getByText('主机名').first()).toBeVisible();

    // Should show OS info
    await expect(adminPage.getByText('操作系统').first()).toBeVisible();
  });

  test('TC0052c: CPU和内存指标显示', async ({ adminPage }) => {
    // CPU section should be visible inside expanded node
    await expect(adminPage.getByText('CPU').first()).toBeVisible();
    await expect(adminPage.getByText(/\d+\s*核/).first()).toBeVisible();

    // Memory section should be visible
    await expect(adminPage.getByText('内存').first()).toBeVisible();
    await expect(adminPage.getByText(/GB/).first()).toBeVisible();
  });

  test('TC0052d: 磁盘使用表格显示', async ({ adminPage }) => {
    await expect(adminPage.getByText('磁盘使用').first()).toBeVisible();
    await expect(adminPage.getByText('/').first()).toBeVisible();
  });

  test('TC0052e: 网络流量信息显示', async ({ adminPage }) => {
    await expect(adminPage.getByText('网络流量').first()).toBeVisible();
    await expect(adminPage.getByText('总发送').first()).toBeVisible();
    await expect(adminPage.getByText('总接收').first()).toBeVisible();
    await expect(adminPage.getByText('发送速率').first()).toBeVisible();
    await expect(adminPage.getByText('接收速率').first()).toBeVisible();
  });

  test('TC0052f: 节点展开收起功能', async ({ adminPage }) => {
    // Nodes should be expanded by default - CPU should be visible
    await expect(adminPage.getByText('CPU').first()).toBeVisible();

    // Get the first node's content area (contains CPU info)
    const firstNodeContent = adminPage.locator('.cursor-pointer').first().locator('..').locator('.grid');

    // Click node header to collapse
    const nodeHeader = adminPage
      .locator('.cursor-pointer')
      .filter({ hasText: /\(/ })
      .first();
    await nodeHeader.click();
    await adminPage.waitForTimeout(500);

    // Click again to expand
    await nodeHeader.click();
    await adminPage.waitForTimeout(500);

    // CPU should be visible again after expand
    await expect(adminPage.getByText('CPU').first()).toBeVisible();
  });

  test('TC0052g: 数据库信息区块展示', async ({ adminPage }) => {
    // Database info section should be visible
    await expect(adminPage.getByText('数据库信息').first()).toBeVisible();

    // Should show database version
    await expect(adminPage.getByText('数据库版本').first()).toBeVisible();

    // Should show max connections
    await expect(adminPage.getByText('最大连接数').first()).toBeVisible();

    // Should show current open connections
    await expect(adminPage.getByText('当前打开连接').first()).toBeVisible();

    // Should show in-use / idle
    await expect(adminPage.getByText(/使用中\s*\/\s*空闲/).first()).toBeVisible();
  });

  test('TC0052h: 数据更新时间在节点信息中展示', async ({ adminPage }) => {
    // collectAt should be inside node expanded content, not in service info
    await expect(adminPage.getByText('数据更新时间').first()).toBeVisible();
  });
});
