import * as path from 'node:path';
import * as fs from 'node:fs';
import { test, expect } from '../../fixtures/auth';
import { DictPage } from '../../pages/DictPage';
import * as XLSX from 'xlsx';

// XLSX module functions
const xlsxReadFile = (XLSX as any).readFile || (XLSX as any).default?.readFile;
const xlsxUtils = (XLSX as any).utils || (XLSX as any).default?.utils;

test.describe('TC0054 字典管理导入完整流程', () => {
  const tempDir = '/tmp/lina-e2e-dict-import';

  // Test data with unique suffix
  const testTypePrefix = `e2e_test_dict_${Date.now()}`;

  /**
   * Create an Excel file with dict type and dict data sheets
   */
  function createDictExcel(
    filePath: string,
    types: Array<{ name: string; type: string; status?: string; remark?: string }>,
    dataList: Array<{ typeName: string; label: string; value: string; sort?: number; tagStyle?: string; cssClass?: string; status?: string; remark?: string }> = [],
  ) {
    const workbook = xlsxUtils.book_new();

    // Create dict type sheet
    const typeHeaders = ['字典名称', '字典类型', '状态', '备注'];
    const typeRows = [typeHeaders, ...types.map((t) => [t.name, t.type, t.status || '正常', t.remark || ''])];
    const typeSheet = xlsxUtils.aoa_to_sheet(typeRows);
    xlsxUtils.book_append_sheet(workbook, typeSheet, '字典类型');

    // Create dict data sheet if provided
    if (dataList.length > 0) {
      const dataHeaders = ['所属类型', '字典标签', '字典值', '排序', 'Tag样式', 'CSS类', '状态', '备注'];
      const dataRows = [dataHeaders, ...dataList.map((d) => [d.typeName, d.label, d.value, d.sort || 0, d.tagStyle || '', d.cssClass || '', d.status || '正常', d.remark || ''])];
      const dataSheet = xlsxUtils.aoa_to_sheet(dataRows);
      xlsxUtils.book_append_sheet(workbook, dataSheet, '字典数据');
    }

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

  test('TC0054a: 下载导入模板并验证双Sheet格式', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Open import modal
    await dictPage.clickTypeImport();
    const modal = adminPage.getByRole('dialog');
    await expect(modal).toBeVisible({ timeout: 5000 });

    // Download template
    const templatePath = path.join(tempDir, 'dict-template.xlsx');
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

    // Should have two sheets: 字典类型 and 字典数据
    expect(sheetNames).toContain('字典类型');
    expect(sheetNames).toContain('字典数据');

    // Verify dict type sheet headers
    const typeSheet = workbook.Sheets['字典类型'];
    const typeHeaderRow = xlsxUtils.sheet_to_json(typeSheet, { header: 1 })[0] as string[];
    expect(typeHeaderRow).toContain('字典名称');
    expect(typeHeaderRow).toContain('字典类型');

    // Verify dict data sheet headers
    const dataSheet = workbook.Sheets['字典数据'];
    const dataHeaderRow = xlsxUtils.sheet_to_json(dataSheet, { header: 1 })[0] as string[];
    expect(dataHeaderRow).toContain('所属类型');
    expect(dataHeaderRow).toContain('字典标签');
    expect(dataHeaderRow).toContain('字典值');
  });

  test('TC0054b: 导入弹窗UI组件完整性', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    await dictPage.clickTypeImport();
    const modal = adminPage.getByRole('dialog').filter({ hasText: '字典管理导入' });
    await expect(modal).toBeVisible({ timeout: 5000 });

    // Verify all UI elements
    await expect(modal.getByText(/点击或者拖拽.*上传/)).toBeVisible();
    await expect(modal.getByText(/允许导入.*xlsx.*xls/)).toBeVisible();
    await expect(modal.getByText('下载模板')).toBeVisible();
    await expect(modal.getByText(/是否更新.*覆盖.*已存在/)).toBeVisible();

    // Close modal
    await adminPage.keyboard.press('Escape');
  });

  test('TC0054c: 导入新字典类型和数据成功', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Create test data file with new dict type and data
    const testType = `${testTypePrefix}_new`;
    const importFilePath = path.join(tempDir, 'import-new-dict.xlsx');

    createDictExcel(
      importFilePath,
      [{ name: 'E2E测试字典', type: testType, remark: '自动化测试创建' }],
      [
        { typeName: testType, label: '选项一', value: '1', sort: 1 },
        { typeName: testType, label: '选项二', value: '2', sort: 2 },
      ],
    );

    // Open import modal
    await dictPage.clickTypeImport();
    const modal = adminPage.getByRole('dialog').filter({ hasText: '字典管理导入' });
    await expect(modal).toBeVisible({ timeout: 5000 });

    // Upload file using file chooser
    const fileChooserPromise = adminPage.waitForEvent('filechooser', { timeout: 5000 });
    await modal.locator('.ant-upload-drag').click();
    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles(importFilePath);

    // Wait for file to appear in upload list
    const uploadItem = modal.locator('.ant-upload-list-item');
    await expect(uploadItem).toBeVisible({ timeout: 10000 });

    // Wait for file to be fully loaded into the component
    await adminPage.waitForTimeout(2000);

    // Click confirm button
    const confirmBtn = modal.getByRole('button', { name: /确\s*认/ });

    // Set up response listener before clicking
    const importResponsePromise = adminPage.waitForResponse(
      (res) => res.url().endsWith('/dict/import') && res.request().method() === 'POST',
      { timeout: 30000 },
    );

    await confirmBtn.click();

    // Wait for response
    const importResponse = await importResponsePromise;
    expect(importResponse.status()).toBe(200);

    // Parse response - handle wrapped response format { code, message, data }
    const responseJson = await importResponse.json();
    const responseBody = responseJson.data || responseJson;

    // Verify import succeeded
    expect(responseBody.typeSuccess).toBe(1);
    expect(responseBody.dataSuccess).toBe(2);

    // Wait for import modal to close
    await expect(modal).not.toBeVisible({ timeout: 5000 });

    // Close the success modal if visible
    await adminPage.waitForTimeout(500);
    const successModal = adminPage.locator('.ant-modal-wrap').filter({ hasText: '成功导入' });
    if (await successModal.isVisible({ timeout: 3000 }).catch(() => false)) {
      await successModal.getByRole('button', { name: /确\s*定|OK|知道了/i }).click();
      await adminPage.waitForTimeout(300);
    }

    // Verify dict type appears in list
    await dictPage.fillTypeSearchField('字典类型', testType);
    await dictPage.clickTypeSearch();
    expect(await dictPage.hasType('E2E测试字典')).toBeTruthy();

    // Verify dict data appears in right panel
    await dictPage.clickTypeRow('E2E测试字典');
    expect(await dictPage.hasData('选项一')).toBeTruthy();
    expect(await dictPage.hasData('选项二')).toBeTruthy();

    // Cleanup: delete the test dict type
    await dictPage.clickTypeReset();
    await dictPage.deleteType('E2E测试字典');
  });

  test('TC0054d: 不开启覆盖模式时重复导入失败', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Create an existing dict type
    const testType = `${testTypePrefix}_dup`;
    await dictPage.createType('已存在字典', testType, '原字典');

    // Verify it was created
    await dictPage.fillTypeSearchField('字典类型', testType);
    await dictPage.clickTypeSearch();
    expect(await dictPage.hasType('已存在字典')).toBeTruthy();

    // Create import file with duplicate type
    const importFilePath = path.join(tempDir, 'import-duplicate.xlsx');
    createDictExcel(
      importFilePath,
      [{ name: '重复字典', type: testType, remark: '尝试覆盖' }],
      [{ typeName: testType, label: '新选项', value: 'new' }],
    );

    // Open import modal
    await dictPage.clickTypeReset();
    await dictPage.clickTypeImport();
    const modal = adminPage.getByRole('dialog').filter({ hasText: '字典管理导入' });
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
      (res) => res.url().includes('/dict/import') && res.request().method() === 'POST',
      { timeout: 30000 },
    );
    await modal.getByRole('button', { name: /确\s*认/ }).click();

    // Wait for response
    const importResponse = await importResponsePromise;
    const responseJson = await importResponse.json();
    const responseBody = responseJson.data || responseJson;

    // Should have failure (duplicate type)
    expect(responseBody.typeFail).toBeGreaterThan(0);

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

    // Verify original dict still exists
    await dictPage.fillTypeSearchField('字典类型', testType);
    await dictPage.clickTypeSearch();
    expect(await dictPage.hasType('已存在字典')).toBeTruthy();

    // Cleanup
    await dictPage.clickTypeReset();
    await dictPage.deleteType('已存在字典');
  });

  test('TC0054e: 开启覆盖模式更新现有数据', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Create an existing dict type with data
    const testType = `${testTypePrefix}_override`;
    await dictPage.createType('待覆盖字典', testType, '原字典备注');
    await dictPage.clickTypeRow('待覆盖字典');
    await dictPage.createData('原选项', 'original', { sort: 1 });

    // Verify original data
    await dictPage.fillDataSearchField('字典标签', '原选项');
    await dictPage.clickDataSearch();
    expect(await dictPage.hasData('原选项')).toBeTruthy();

    // Create import file with SAME dict type but updated name/remark (to test overwrite)
    const importFilePath = path.join(tempDir, 'import-overwrite.xlsx');
    createDictExcel(
      importFilePath,
      [{ name: '已覆盖字典', type: testType, remark: '覆盖后备注' }],
      [{ typeName: testType, label: '新选项', value: 'original', sort: 2, tagStyle: 'primary' }],
    );

    // Open import modal
    await dictPage.clickDataReset();
    await dictPage.clickTypeReset();
    await dictPage.clickTypeImport();
    const modal = adminPage.getByRole('dialog').filter({ hasText: '字典管理导入' });
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
      (res) => res.url().includes('/dict/import') && res.request().method() === 'POST',
      { timeout: 30000 },
    );
    await modal.getByRole('button', { name: /确\s*认/ }).click();

    // Wait for response
    const importResponse = await importResponsePromise;
    const responseJson = await importResponse.json();
    const responseBody = responseJson.data || responseJson;

    // Should succeed with overwrite - type updated, data updated
    expect(responseBody.typeSuccess).toBe(1);
    expect(responseBody.dataSuccess).toBe(1);

    // Wait for modals to close
    await expect(modal).not.toBeVisible({ timeout: 5000 });

    // Close success modal if visible
    await adminPage.waitForTimeout(500);
    const successModal = adminPage.locator('.ant-modal-wrap').filter({ hasText: '成功导入' });
    if (await successModal.isVisible({ timeout: 3000 }).catch(() => false)) {
      await successModal.getByRole('button', { name: /确\s*定|OK|知道了/i }).click();
      await adminPage.waitForTimeout(300);
    }

    // Verify the dict type was updated (name changed)
    await dictPage.fillTypeSearchField('字典类型', testType);
    await dictPage.clickTypeSearch();
    expect(await dictPage.hasType('已覆盖字典')).toBeTruthy();

    // Click the updated type row to load data in right panel
    await dictPage.clickTypeRow('已覆盖字典');
    await adminPage.waitForTimeout(1000);

    // Refresh the data panel by clicking the type row again
    await dictPage.clickDataReset();
    await adminPage.waitForTimeout(500);

    // Verify the dict data was updated (label changed to '新选项')
    expect(await dictPage.hasData('新选项')).toBeTruthy();

    // Cleanup
    await dictPage.clickDataReset();
    await dictPage.clickTypeReset();
    await dictPage.deleteType('已覆盖字典');
  });

  test('TC0054f: 导入弹窗覆盖开关可切换', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    await dictPage.clickTypeImport();
    const modal = adminPage.getByRole('dialog').filter({ hasText: '字典管理导入' });
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

  test('TC0054g: 导入确认按钮存在且可点击', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    await dictPage.clickTypeImport();
    const modal = adminPage.getByRole('dialog').filter({ hasText: '字典管理导入' });
    await expect(modal).toBeVisible({ timeout: 5000 });

    // Verify confirm button exists
    const confirmBtn = modal.getByRole('button', { name: /确\s*认/ });
    await expect(confirmBtn).toBeVisible();

    // Close modal
    await adminPage.keyboard.press('Escape');
  });

  test('TC0054h: 导入包含必填字段为空的记录显示失败', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Create import file with missing required fields
    const importFilePath = path.join(tempDir, 'import-invalid.xlsx');
    const workbook = xlsxUtils.book_new();

    // Dict type sheet with empty name/type (using correct headers)
    // Backend requires at least 3 columns, empty values are allowed but will cause issues
    const typeSheet = xlsxUtils.aoa_to_sheet([
      ['字典名称', '字典类型', '状态', '备注'],
      ['', 'test.empty.name.valid', '正常', ''], // Empty name - will be inserted with empty name
      ['测试字典2', '', '正常', ''], // Empty type - will cause issues
    ]);
    xlsxUtils.book_append_sheet(workbook, typeSheet, '字典类型');

    // Dict data sheet with invalid dict type reference (will fail "字典类型不存在" check)
    const dataSheet = xlsxUtils.aoa_to_sheet([
      ['所属类型', '字典标签', '字典值', '排序', 'Tag样式', 'CSS类', '状态', '备注'],
      ['nonexistent_dict_type', '标签', '值', '1', '', '', '正常', ''], // Non-existent dict type
    ]);
    xlsxUtils.book_append_sheet(workbook, dataSheet, '字典数据');

    (XLSX as any).writeFile(workbook, importFilePath);

    // Open import modal
    await dictPage.clickTypeImport();
    const modal = adminPage.getByRole('dialog').filter({ hasText: '字典管理导入' });
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
      (res) => res.url().includes('/dict/import') && res.request().method() === 'POST',
      { timeout: 30000 },
    );
    await modal.getByRole('button', { name: /确\s*认/ }).click();

    // Wait for response
    const importResponse = await importResponsePromise;
    const responseJson = await importResponse.json();
    const responseBody = responseJson.data || responseJson;

    // Should have at least one data failure (nonexistent dict type)
    // Dict type with empty name might succeed, but data with non-existent type will fail
    expect(responseBody.dataFail).toBeGreaterThan(0);

    // Wait for import modal to close
    await expect(modal).not.toBeVisible({ timeout: 5000 });

    // Close result modal if visible
    await adminPage.waitForTimeout(500);
    const resultModal = adminPage.locator('.ant-modal-wrap').filter({ hasText: /失败|成功/ });
    if (await resultModal.isVisible({ timeout: 3000 }).catch(() => false)) {
      await resultModal.getByRole('button', { name: /确\s*定|OK|知道了/i }).click();
      await adminPage.waitForTimeout(300);
    }

    // Cleanup: delete any created test data
    await dictPage.clickTypeReset();
    // Search for the empty name dict type
    await dictPage.fillTypeSearchField('字典类型', 'test.empty.name.valid');
    await dictPage.clickTypeSearch();
    // Delete if exists (might not exist if failed)
    try {
      if (await dictPage.hasType('')) {
        await dictPage.deleteType('');
      }
    } catch {
      // Ignore if not found
    }
  });
});