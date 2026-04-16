import { test, expect } from '../../fixtures/auth';
import { DictPage } from '../../pages/DictPage';

test.describe('TC0053 字典管理导入', () => {
  test('TC0053a: 点击导入按钮打开导入弹窗', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    await dictPage.clickTypeImport();

    const modal = adminPage.getByRole('dialog');
    await expect(modal).toBeVisible({ timeout: 5000 });
    await expect(modal).toContainText('字典管理导入');
  });

  test('TC0053b: 导入弹窗中有下载模板链接', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    await dictPage.clickTypeImport();

    const modal = adminPage.getByRole('dialog');
    await expect(modal).toBeVisible({ timeout: 5000 });
    await expect(modal.getByText('下载模板')).toBeVisible();
  });

  test('TC0053c: 导入弹窗中有拖拽上传区域和覆盖模式开关', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    await dictPage.clickTypeImport();

    const modal = adminPage.getByRole('dialog');
    await expect(modal).toBeVisible({ timeout: 5000 });
    await expect(modal.getByText('点击或者拖拽到此处上传文件')).toBeVisible();
    await expect(modal.getByText('允许导入xlsx, xls文件')).toBeVisible();
    await expect(modal.getByText(/是否更新\/覆盖已存在的字典类型和字典数据/)).toBeVisible();
  });

  test('TC0053d: 下载模板请求发送到正确的端点', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    await dictPage.clickTypeImport();

    const modal = adminPage.getByRole('dialog');
    await expect(modal).toBeVisible({ timeout: 5000 });

    const responsePromise = adminPage.waitForResponse(
      (res) => res.url().includes('/api/v1/dict/import-template'),
      { timeout: 10000 },
    );

    await modal.getByText('下载模板').click();
    const response = await responsePromise;

    expect(response.status()).toBe(200);
    expect(response.headers()['content-type']).toContain('spreadsheetml');
  });
});