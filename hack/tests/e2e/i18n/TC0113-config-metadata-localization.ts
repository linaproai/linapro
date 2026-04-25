import * as fs from 'node:fs';
import * as path from 'node:path';
import type { APIRequestContext, Locator } from '@playwright/test';

import * as XLSX from 'xlsx';

import { test, expect } from '../../fixtures/auth';
import { ConfigPage } from '../../pages/ConfigPage';
import { createAdminApiContext, expectSuccess } from '../../support/api/job';
import { closeDialogWithEscape, waitForDialogReady } from '../../support/ui';

const xlsxReadFile = (XLSX as any).readFile || (XLSX as any).default?.readFile;
const xlsxUtils = (XLSX as any).utils || (XLSX as any).default?.utils;

test.describe('TC0113 参数设置元数据国际化', () => {
  const tempDir = '/tmp/lina-e2e-config-i18n';
  const seedConfigKey = 'sys.auth.pageTitle';

  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    if (!fs.existsSync(tempDir)) {
      fs.mkdirSync(tempDir, { recursive: true });
    }
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    await adminApi.dispose();
    if (fs.existsSync(tempDir)) {
      fs.rmSync(tempDir, { recursive: true, force: true });
    }
  });

  test('TC0113a: 英文环境下配置列表显示本地化名称但保留原始参数值', async ({
    adminPage,
    mainLayout,
  }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();
    await mainLayout.switchLanguage('English');

    await configPage.fillSearchField('参数键名', seedConfigKey);
    await configPage.clickSearch();

    const row = configPage.findRowByExactKey(seedConfigKey);
    await expect(row).toBeVisible();
    await expect(row.getByText('Login - Page Title', { exact: true })).toBeVisible();
    await expect(row.getByText('AI驱动的全栈开发框架', { exact: true })).toBeVisible();
  });

  test('TC0113b: 英文环境下配置编辑回填继续使用数据库原始值', async ({
    adminPage,
    mainLayout,
  }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();
    await mainLayout.switchLanguage('English');

    await configPage.fillSearchField('参数键名', seedConfigKey);
    await configPage.clickSearch();

    const row = configPage.findRowByExactKey(seedConfigKey);
    await expect(row).toBeVisible();
    await clickRowAction(row, /编\s*辑|Edit/i);

    const dialog = adminPage.getByRole('dialog');
    await waitForDialogReady(dialog);

    await expect(dialog.getByLabel(/参数名称|Parameter Name/i)).toHaveValue('登录展示-页面标题');
    await expect(dialog.getByLabel(/参数键名|Parameter Key/i)).toHaveValue(seedConfigKey);
    await expect(dialog.getByLabel(/参数键值|Parameter Value/i)).toHaveValue('AI驱动的全栈开发框架');
    await expect(dialog.getByLabel(/备注|Remark/i)).toHaveValue('控制登录页顶部主标题文案。');

    await closeDialogWithEscape(adminPage, dialog);
  });

  test('TC0113c: 英文环境下配置导入模板返回本地化表头与示例元数据', async ({
    adminPage,
    mainLayout,
  }) => {
    const configPage = new ConfigPage(adminPage);
    await configPage.goto();
    await mainLayout.switchLanguage('English');

    await configPage.clickImport();
    const dialog = adminPage.getByRole('dialog');
    await waitForDialogReady(dialog);

    const templatePath = path.join(tempDir, 'config-template-en.xlsx');
    const downloadPromise = adminPage.waitForEvent('download', { timeout: 15000 });
    await dialog.getByText(/下载模板|Download Template/i).click();
    const download = await downloadPromise;
    await download.saveAs(templatePath);

    await closeDialogWithEscape(adminPage, dialog);

    const workbook = xlsxReadFile(templatePath);
    const sheet = workbook.Sheets[workbook.SheetNames[0]];
    const rows = xlsxUtils.sheet_to_json(sheet, { header: 1 }) as string[][];

    expect(rows[0]).toEqual(['Parameter Name', 'Parameter Key', 'Parameter Value', 'Remark']);
    expect(rows[1]?.[0]).toBe('Authentication - JWT Expiration');
    expect(rows[1]?.[1]).toBe('sys.jwt.expire');
    expect(rows[1]?.[2]).toBe('24h');
  });

  test('TC0113d: 公共前端配置接口在英文环境下返回本地化品牌文案', async () => {
    const response = await adminApi.get('config/public/frontend', {
      headers: {
        'Accept-Language': 'en-US',
      },
    });
    const data = await expectSuccess<{
      app: { name: string };
      auth: { pageDesc: string; pageTitle: string; loginSubtitle: string };
      ui: { watermarkContent: string };
    }>(response);

    expect(data.app.name).toBe('LinaPro');
    expect(data.auth.pageTitle).toBe('AI-driven full-stack development framework');
    expect(data.auth.pageDesc).toBe(
      'Built for evolving business needs, with an out-of-the-box admin entry point and a flexible pluggable extension model',
    );
    expect(data.auth.loginSubtitle).toBe(
      'Enter your account credentials to start managing your projects',
    );
    expect(data.ui.watermarkContent).toBe('LinaPro');
  });
});

async function clickRowAction(row: Locator, label: RegExp) {
  await row.hover();
  await row.getByRole('button', { name: label }).first().click();
}
