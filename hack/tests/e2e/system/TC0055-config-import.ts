import { test, expect } from '../../fixtures/auth';
import { ConfigPage } from '../../pages/ConfigPage';

test.describe('TC0055 参数设置导入', () => {
  test('TC0055a: 点击导入按钮打开导入弹窗', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    await configPage.clickImport();

    const modal = adminPage.getByRole('dialog');
    await expect(modal).toBeVisible({ timeout: 5000 });
    await expect(modal).toContainText('参数设置导入');
  });

  test('TC0055b: 导入弹窗中有下载模板链接', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    await configPage.clickImport();

    const modal = adminPage.getByRole('dialog');
    await expect(modal).toBeVisible({ timeout: 5000 });
    await expect(modal.getByText('下载模板')).toBeVisible();
  });

  test('TC0055c: 导入弹窗中有拖拽上传区域和覆盖模式开关', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    await configPage.clickImport();

    const modal = adminPage.getByRole('dialog');
    await expect(modal).toBeVisible({ timeout: 5000 });
    await expect(modal.getByText('点击或者拖拽到此处上传文件')).toBeVisible();
    await expect(modal.getByText('允许导入xlsx, xls文件')).toBeVisible();
    await expect(modal.getByText(/是否更新\/覆盖已存在的参数设置数据/)).toBeVisible();
  });

  test('TC0055d: 下载模板请求发送到正确的端点', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    await configPage.clickImport();

    const modal = adminPage.getByRole('dialog');
    await expect(modal).toBeVisible({ timeout: 5000 });

    const responsePromise = adminPage.waitForResponse(
      (res) => res.url().includes('/api/v1/config/import-template'),
      { timeout: 10000 },
    );

    await modal.getByText('下载模板').click();
    const response = await responsePromise;

    expect(response.status()).toBe(200);
    expect(response.headers()['content-type']).toContain('spreadsheetml');
  });
});