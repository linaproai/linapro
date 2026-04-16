import * as path from 'node:path';
import * as fs from 'node:fs';
import { test, expect } from '../../fixtures/auth';
import { ConfigPage } from '../../pages/ConfigPage';
import * as XLSX from 'xlsx';

// XLSX module functions
const xlsxReadFile = (XLSX as any).readFile || (XLSX as any).default?.readFile;
const xlsxUtils = (XLSX as any).utils || (XLSX as any).default?.utils;

test.describe('TC0058 参数设置导入完整流程', () => {
  const tempDir = '/tmp/lina-e2e-config-import';

  // Test config keys with unique suffix
  const testKeyPrefix = `e2e.test.config.${Date.now()}`;

  /**
   * Create an Excel file with config data
   */
  function createConfigExcel(filePath: string, configs: Array<{ name: string; key: string; value: string; remark?: string }>) {
    const workbook = xlsxUtils.book_new();

    // Create worksheet with headers
    const headers = ['参数名称', '参数键名', '参数键值', '备注'];
    const rows = [headers, ...configs.map(c => [c.name, c.key, c.value, c.remark || ''])];

    const worksheet = xlsxUtils.aoa_to_sheet(rows);
    xlsxUtils.book_append_sheet(workbook, worksheet, 'Sheet1');

    // Write to file
    (XLSX as any).writeFile(workbook, filePath);
  }

  test.beforeAll(() => {
    // Create temp directory
    if (!fs.existsSync(tempDir)) {
      fs.mkdirSync(tempDir, { recursive: true });
    }
  });

  test.afterAll(() => {
    // Cleanup temp files
    if (fs.existsSync(tempDir)) {
      fs.rmSync(tempDir, { recursive: true, force: true });
    }
  });

  test('TC0058a: 下载导入模板并验证格式', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    // Open import modal
    await configPage.clickImport();
    const modal = adminPage.getByRole('dialog');
    await expect(modal).toBeVisible({ timeout: 5000 });

    // Download template
    const templatePath = path.join(tempDir, 'template.xlsx');
    const downloadPromise = adminPage.waitForEvent('download', { timeout: 10000 });
    await modal.getByText('下载模板').click();
    const download = await downloadPromise;
    await download.saveAs(templatePath);

    // Close modal
    await adminPage.keyboard.press('Escape');
    await adminPage.waitForTimeout(500);

    // Verify template file exists and is a valid Excel
    expect(fs.existsSync(templatePath)).toBeTruthy();
    const stats = fs.statSync(templatePath);
    expect(stats.size).toBeGreaterThan(1000);

    // Verify template structure by reading it
    const workbook = xlsxReadFile(templatePath);
    const sheetNames = workbook.SheetNames;
    expect(sheetNames.length).toBeGreaterThan(0);

    const sheet = workbook.Sheets[sheetNames[0]];
    expect(sheet).toBeDefined();

    // Verify headers in row 1
    const headerRow = xlsxUtils.sheet_to_json(sheet, { header: 1 })[0] as string[];
    expect(headerRow).toContain('参数名称');
    expect(headerRow).toContain('参数键名');
    expect(headerRow).toContain('参数键值');
  });

  test('TC0058b: 导入弹窗UI组件完整性', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    await configPage.clickImport();
    const modal = adminPage.getByRole('dialog').filter({ hasText: '参数设置导入' });
    await expect(modal).toBeVisible({ timeout: 5000 });

    // Verify all UI elements
    await expect(modal.getByText(/点击或者拖拽.*上传/)).toBeVisible();
    await expect(modal.getByText(/允许导入.*xlsx.*xls/)).toBeVisible();
    await expect(modal.getByText('下载模板')).toBeVisible();
    await expect(modal.getByText(/是否更新.*覆盖.*已存在/)).toBeVisible();

    // Close modal
    await adminPage.keyboard.press('Escape');
  });

  test('TC0058c: 导入新数据成功并在列表中可见', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    // Create test data file with 2 new configs
    const testKey1 = `${testKeyPrefix}.new1`;
    const testKey2 = `${testKeyPrefix}.new2`;
    const importFilePath = path.join(tempDir, 'import-new-data.xlsx');

    createConfigExcel(importFilePath, [
      { name: 'E2E测试配置1', key: testKey1, value: '测试值1', remark: '自动化测试创建' },
      { name: 'E2E测试配置2', key: testKey2, value: '测试值2', remark: '自动化测试创建' },
    ]);

    // Open import modal
    await configPage.clickImport();
    const modal = adminPage.getByRole('dialog').filter({ hasText: '参数设置导入' });
    await expect(modal).toBeVisible({ timeout: 5000 });

    // Upload file using file chooser
    const fileChooserPromise = adminPage.waitForEvent('filechooser', { timeout: 5000 });
    await modal.locator('.ant-upload-drag').click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles(importFilePath);

    // Wait for file to appear in upload list
    const uploadItem = modal.locator('.ant-upload-list-item');
    await expect(uploadItem).toBeVisible({ timeout: 10000 });

    // Wait for Vue reactivity to process the file
    await adminPage.waitForTimeout(1500);

    // Click confirm button
    const confirmBtn = modal.getByRole('button', { name: /确\s*认/ });

    // Set up response listener before clicking
    const importResponsePromise = adminPage.waitForResponse(
      (res) => res.url().includes('/config/import') && res.request().method() === 'POST',
      { timeout: 30000 },
    );

    await confirmBtn.click();

    // Wait for response
    const importResponse = await importResponsePromise;
    expect(importResponse.status()).toBe(200);

    // Parse response - handle wrapped response format { code, message, data }
    const responseJson = await importResponse.json();
    const responseBody = responseJson.data || responseJson;
    expect(responseBody.success).toBe(2);
    expect(responseBody.fail).toBe(0);

    // Wait for import modal to close
    await expect(modal).not.toBeVisible({ timeout: 5000 });

    // Close the success modal if visible
    await adminPage.waitForTimeout(500);
    const successModal = adminPage.locator('.ant-modal-wrap').filter({ hasText: '成功导入' });
    if (await successModal.isVisible({ timeout: 3000 }).catch(() => false)) {
      await successModal.getByRole('button', { name: /确\s*定|OK|知道了/i }).click();
      await adminPage.waitForTimeout(300);
    }

    // Verify data appears in list - search for first config
    await configPage.fillSearchField('参数键名', testKey1);
    await configPage.clickSearch();
    expect(await configPage.hasConfig('E2E测试配置1')).toBeTruthy();

    // Verify second config
    await configPage.fillSearchField('参数键名', testKey2);
    await configPage.clickSearch();
    expect(await configPage.hasConfig('E2E测试配置2')).toBeTruthy();

    // Cleanup: delete the test configs
    await configPage.clickReset();
    await configPage.delete('E2E测试配置1');
    await configPage.delete('E2E测试配置2');
  });

  test('TC0058d: 不开启覆盖模式时重复key导入失败并显示错误', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    // Create an existing config
    const testKey = `${testKeyPrefix}.duplicate`;
    await configPage.create('已存在配置', testKey, '原值');

    // Verify it was created
    await configPage.fillSearchField('参数键名', testKey);
    await configPage.clickSearch();
    expect(await configPage.hasConfig('已存在配置')).toBeTruthy();

    // Create import file with duplicate key
    const importFilePath = path.join(tempDir, 'import-duplicate.xlsx');
    createConfigExcel(importFilePath, [
      { name: '重复配置', key: testKey, value: '新值', remark: '尝试覆盖' },
    ]);

    // Open import modal
    await configPage.clickReset();
    await configPage.clickImport();
    const modal = adminPage.getByRole('dialog').filter({ hasText: '参数设置导入' });
    await expect(modal).toBeVisible({ timeout: 5000 });

    // Verify overwrite switch is OFF
    const switchEl = modal.locator('.ant-switch');
    expect(await switchEl.isChecked()).toBe(false);

    // Upload file using file chooser
    const fileChooserPromise = adminPage.waitForEvent('filechooser', { timeout: 5000 });
    await modal.locator('.ant-upload-drag').click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles(importFilePath);

    // Wait for file to appear
    const uploadItem = modal.locator('.ant-upload-list-item');
    await expect(uploadItem).toBeVisible({ timeout: 10000 });
    await adminPage.waitForTimeout(1500);

    // Submit import
    const importResponsePromise = adminPage.waitForResponse(
      (res) => res.url().includes('/config/import') && res.request().method() === 'POST',
      { timeout: 30000 },
    );
    await modal.getByRole('button', { name: /确\s*认/ }).click();

    // Wait for response
    const importResponse = await importResponsePromise;
    const responseJson = await importResponse.json();
    const responseBody = responseJson.data || responseJson;

    // Should have 1 failure (duplicate key)
    expect(responseBody.success).toBe(0);
    expect(responseBody.fail).toBe(1);
    expect(responseBody.failList[0].reason).toContain('已存在');

    // Wait for import modal to close
    await expect(modal).not.toBeVisible({ timeout: 5000 });

    // Wait for error/success modal to appear
    await adminPage.waitForTimeout(500);
    const resultModal = adminPage.locator('.ant-modal-wrap').filter({ hasText: '失败' });

    // Close result modal if visible
    if (await resultModal.isVisible({ timeout: 3000 }).catch(() => false)) {
      await resultModal.getByRole('button', { name: /确\s*定|OK|知道了/i }).click();
      await adminPage.waitForTimeout(300);
    }

    // Verify original value unchanged
    await configPage.fillSearchField('参数键名', testKey);
    await configPage.clickSearch();
    expect(await configPage.hasConfig('已存在配置')).toBeTruthy();

    // Cleanup
    await configPage.clickReset();
    await configPage.delete('已存在配置');
  });

  test('TC0058e: 开启覆盖模式更新现有配置', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    // Create an existing config
    const testKey = `${testKeyPrefix}.override`;
    await configPage.create('待覆盖配置', testKey, '原值');

    // Verify it was created with original value
    await configPage.fillSearchField('参数键名', testKey);
    await configPage.clickSearch();
    expect(await configPage.hasConfig('待覆盖配置')).toBeTruthy();

    // Create import file with new value
    const importFilePath = path.join(tempDir, 'import-override.xlsx');
    createConfigExcel(importFilePath, [
      { name: '已更新配置', key: testKey, value: '更新后的值', remark: '覆盖更新' },
    ]);

    // Open import modal
    await configPage.clickReset();
    await configPage.clickImport();
    const modal = adminPage.getByRole('dialog').filter({ hasText: '参数设置导入' });
    await expect(modal).toBeVisible({ timeout: 5000 });

    // Enable overwrite mode
    const switchEl = modal.locator('.ant-switch');
    await switchEl.click();
    await adminPage.waitForTimeout(300);
    expect(await switchEl.isChecked()).toBe(true);

    // Upload file using file chooser
    const fileChooserPromise = adminPage.waitForEvent('filechooser', { timeout: 5000 });
    await modal.locator('.ant-upload-drag').click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles(importFilePath);

    // Wait for file to appear
    const uploadItem = modal.locator('.ant-upload-list-item');
    await expect(uploadItem).toBeVisible({ timeout: 10000 });
    await adminPage.waitForTimeout(1500);

    // Submit import
    const importResponsePromise = adminPage.waitForResponse(
      (res) => res.url().includes('/config/import') && res.request().method() === 'POST',
      { timeout: 30000 },
    );
    await modal.getByRole('button', { name: /确\s*认/ }).click();

    // Wait for response
    const importResponse = await importResponsePromise;
    const responseJson = await importResponse.json();
    const responseBody = responseJson.data || responseJson;

    // Should succeed with 1 update
    expect(responseBody.success).toBe(1);
    expect(responseBody.fail).toBe(0);

    // Wait for modals to close
    await expect(modal).not.toBeVisible({ timeout: 5000 });

    // Close success modal if visible
    await adminPage.waitForTimeout(500);
    const successModal = adminPage.locator('.ant-modal-wrap').filter({ hasText: '成功导入' });
    if (await successModal.isVisible({ timeout: 3000 }).catch(() => false)) {
      await successModal.getByRole('button', { name: /确\s*定|OK|知道了/i }).click();
      await adminPage.waitForTimeout(300);
    }

    // Verify the config was updated - search by key
    await configPage.fillSearchField('参数键名', testKey);
    await configPage.clickSearch();

    // Should show updated name
    expect(await configPage.hasConfig('已更新配置')).toBeTruthy();

    // Cleanup
    await configPage.clickReset();
    await configPage.delete('已更新配置');
  });

  test('TC0058f: 导入弹窗覆盖开关可切换', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    await configPage.clickImport();
    const modal = adminPage.getByRole('dialog').filter({ hasText: '参数设置导入' });
    await expect(modal).toBeVisible({ timeout: 5000 });

    // Find the switch
    const switchEl = modal.locator('.ant-switch');
    await expect(switchEl).toBeVisible();

    // Initial state should be OFF
    const initialState = await switchEl.isChecked();
    expect(initialState).toBe(false);

    // Toggle ON
    await switchEl.click();
    await adminPage.waitForTimeout(300);
    expect(await switchEl.isChecked()).toBe(true);

    // Toggle back OFF
    await switchEl.click();
    await adminPage.waitForTimeout(300);
    expect(await switchEl.isChecked()).toBe(false);

    // Close modal
    await adminPage.keyboard.press('Escape');
  });

  test('TC0058g: 导入确认按钮存在且可点击', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    await configPage.clickImport();
    const modal = adminPage.getByRole('dialog').filter({ hasText: '参数设置导入' });
    await expect(modal).toBeVisible({ timeout: 5000 });

    // Verify confirm button exists
    const confirmBtn = modal.getByRole('button', { name: /确\s*认/ });
    await expect(confirmBtn).toBeVisible();

    // Close modal
    await adminPage.keyboard.press('Escape');
  });

  test('TC0058h: 导入包含必填字段为空的记录显示失败', async ({ adminPage }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();

    // Create import file with missing required fields
    const importFilePath = path.join(tempDir, 'import-invalid.xlsx');
    const workbook = xlsxUtils.book_new();
    const worksheet = xlsxUtils.aoa_to_sheet([
      ['参数名称', '参数键名', '参数键值', '备注'],
      ['', 'test.key.empty.name', '值1', ''], // Empty name
      ['测试配置2', '', '值2', ''], // Empty key
      ['测试配置3', 'test.key.empty.value', '', ''], // Empty value
    ]);
    xlsxUtils.book_append_sheet(workbook, worksheet, 'Sheet1');
    (XLSX as any).writeFile(workbook, importFilePath);

    // Open import modal
    await configPage.clickImport();
    const modal = adminPage.getByRole('dialog').filter({ hasText: '参数设置导入' });
    await expect(modal).toBeVisible({ timeout: 5000 });

    // Upload file using file chooser
    const fileChooserPromise = adminPage.waitForEvent('filechooser', { timeout: 5000 });
    await modal.locator('.ant-upload-drag').click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles(importFilePath);

    // Wait for file to appear
    const uploadItem = modal.locator('.ant-upload-list-item');
    await expect(uploadItem).toBeVisible({ timeout: 10000 });
    await adminPage.waitForTimeout(1500);

    // Submit import
    const importResponsePromise = adminPage.waitForResponse(
      (res) => res.url().includes('/config/import') && res.request().method() === 'POST',
      { timeout: 30000 },
    );
    await modal.getByRole('button', { name: /确\s*认/ }).click();

    // Wait for response
    const importResponse = await importResponsePromise;
    const responseJson = await importResponse.json();
    const responseBody = responseJson.data || responseJson;

    // All 3 records should fail due to missing required fields
    expect(responseBody.success).toBe(0);
    expect(responseBody.fail).toBe(3);

    // Wait for import modal to close
    await expect(modal).not.toBeVisible({ timeout: 5000 });

    // Verify error modal shows failure details
    await adminPage.waitForTimeout(500);
    const resultModal = adminPage.locator('.ant-modal-wrap').filter({ hasText: '失败' });
    if (await resultModal.isVisible({ timeout: 3000 }).catch(() => false)) {
      await resultModal.getByRole('button', { name: /确\s*定|OK|知道了/i }).click();
      await adminPage.waitForTimeout(300);
    }
  });
});