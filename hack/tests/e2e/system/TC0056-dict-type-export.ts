import * as path from 'node:path';
import * as fs from 'node:fs';
import { test, expect } from '../../fixtures/auth';
import { DictPage } from '../../pages/DictPage';
import * as XLSX from 'xlsx';

// XLSX module functions
const xlsxReadFile = (XLSX as any).readFile || (XLSX as any).default?.readFile;
const xlsxUtils = (XLSX as any).utils || (XLSX as any).default?.utils;

test.describe('TC0056 字典管理导出', () => {
  const tempDir = '/tmp/lina-e2e-dict-export';

  test.beforeAll(() => {
    if (!fs.existsSync(tempDir)) {
      fs.mkdirSync(tempDir, { recursive: true });
    }
  });

  test.afterAll(() => {
    if (fs.existsSync(tempDir)) {
      fs.rmSync(tempDir, { recursive: true, force: true });
    }
  });

  test('TC0056a: 导出全部字典类型和数据', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Click export button in type panel
    const exportBtn = adminPage.getByRole('button', { name: /导\s*出/ }).first();
    await expect(exportBtn).toBeVisible({ timeout: 10000 });
    await exportBtn.click();

    // Verify modal appears with combined export message
    const modalContent = adminPage.locator('.ant-modal-content');
    await expect(modalContent).toBeVisible({ timeout: 5000 });
    await expect(modalContent.getByText(/字典类型.*字典数据/)).toBeVisible();

    // Set up response listener for combined export endpoint
    const responsePromise = adminPage.waitForResponse(
      (resp) => resp.url().includes('dict/export'),
      { timeout: 15000 },
    );

    // Click confirm button
    const confirmBtn = modalContent.getByRole('button', { name: /确\s*认/ });
    await confirmBtn.click();

    // Wait for response and verify
    const response = await responsePromise;
    expect(response.status()).toBe(200);
  });

  test('TC0056b: 导出文件格式验证（双Sheet结构）', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Click export button
    const exportBtn = adminPage.getByRole('button', { name: /导\s*出/ }).first();
    await expect(exportBtn).toBeVisible({ timeout: 10000 });

    // Set up download listener
    const downloadPromise = adminPage.waitForEvent('download', { timeout: 30000 });

    await exportBtn.click();

    // Confirm in modal
    const modalContent = adminPage.locator('.ant-modal-content');
    await expect(modalContent).toBeVisible({ timeout: 5000 });
    await modalContent.getByRole('button', { name: /确\s*认/ }).click();

    // Wait for download
    const download = await downloadPromise;
    const exportPath = path.join(tempDir, 'dict-export-all.xlsx');
    await download.saveAs(exportPath);

    // Verify file exists
    expect(fs.existsSync(exportPath)).toBeTruthy();

    // Read and verify Excel structure
    const workbook = xlsxReadFile(exportPath);
    const sheetNames = workbook.SheetNames;

    // Should have two sheets: 字典类型 and 字典数据
    expect(sheetNames).toContain('字典类型');
    expect(sheetNames).toContain('字典数据');

    // Verify dict type sheet headers
    const typeSheet = workbook.Sheets['字典类型'];
    const typeData = xlsxUtils.sheet_to_json(typeSheet, { header: 1 }) as string[][];
    expect(typeData.length).toBeGreaterThan(1); // Has header + at least one data row
    const typeHeaders = typeData[0];
    expect(typeHeaders).toContain('字典名称');
    expect(typeHeaders).toContain('字典类型');

    // Verify dict data sheet headers
    const dataSheet = workbook.Sheets['字典数据'];
    const dataData = xlsxUtils.sheet_to_json(dataSheet, { header: 1 }) as string[][];
    expect(dataData.length).toBeGreaterThan(1); // Has header + at least one data row
    const dataHeaders = dataData[0];
    expect(dataHeaders).toContain('所属类型');
    expect(dataHeaders).toContain('字典标签');
    expect(dataHeaders).toContain('字典值');
  });

  test('TC0056c: 按字典名称筛选导出', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Search for specific dict type
    await dictPage.fillTypeSearchField('字典名称', '性别');
    await dictPage.clickTypeSearch();

    // Wait for results to filter
    await adminPage.waitForTimeout(500);

    // Click export button
    const exportBtn = adminPage.getByRole('button', { name: /导\s*出/ }).first();
    await expect(exportBtn).toBeVisible({ timeout: 10000 });

    // Set up download listener
    const downloadPromise = adminPage.waitForEvent('download', { timeout: 30000 });

    await exportBtn.click();

    // Confirm in modal
    const modalContent = adminPage.locator('.ant-modal-content');
    await expect(modalContent).toBeVisible({ timeout: 5000 });
    await modalContent.getByRole('button', { name: /确\s*认/ }).click();

    // Wait for download
    const download = await downloadPromise;
    const exportPath = path.join(tempDir, 'dict-export-filtered.xlsx');
    await download.saveAs(exportPath);

    // Verify file exists
    expect(fs.existsSync(exportPath)).toBeTruthy();

    // Read and verify filtered content
    const workbook = xlsxReadFile(exportPath);
    const typeSheet = workbook.Sheets['字典类型'];
    const typeData = xlsxUtils.sheet_to_json(typeSheet) as Array<{ 字典名称: string }>;

    // All exported types should contain '性别' in name
    expect(typeData.length).toBeGreaterThan(0);
    for (const row of typeData) {
      expect(row['字典名称']).toContain('性别');
    }

    // Reset search
    await dictPage.clickTypeReset();
  });
});