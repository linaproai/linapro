import { test, expect } from '../../fixtures/auth';
import { DictPage } from '../../pages/DictPage';

test.describe('TC0057 字典数据面板无独立导出导入功能', () => {
  test('TC0057a: 字典数据面板没有导出按钮', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Select a dict type row to load dict data in right panel
    await dictPage.clickTypeRow('sys_oper_type');

    // Wait for data to load
    await adminPage.waitForTimeout(500);

    // Data panel should NOT have export button
    const dataPanel = adminPage.locator('#dict-data');
    await expect(dataPanel.getByRole('button', { name: /导\s*出/ })).not.toBeVisible();
  });

  test('TC0057b: 字典数据面板没有导入按钮', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Select a dict type row to load dict data in right panel
    await dictPage.clickTypeRow('sys_oper_type');

    // Wait for data to load
    await adminPage.waitForTimeout(500);

    // Data panel should NOT have import button
    const dataPanel = adminPage.locator('#dict-data');
    await expect(dataPanel.getByRole('button', { name: /导\s*入/ })).not.toBeVisible();
  });

  test('TC0057c: 字典数据面板有新增和删除按钮', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Select a dict type row to load dict data in right panel
    await dictPage.clickTypeRow('sys_oper_type');

    // Wait for data to load
    await adminPage.waitForTimeout(500);

    // Assert against the toolbar actions only. The data table also renders
    // row-level delete buttons, so an unscoped role query becomes ambiguous.
    const dataPanel = adminPage.locator('#dict-data');
    const dataToolbar = dataPanel.locator('.vxe-grid--toolbar, .vxe-toolbar').first();
    await expect(dataToolbar.getByRole('button', { name: /新\s*增/ }).first()).toBeVisible();
    await expect(dataToolbar.getByRole('button', { name: /删\s*除/ }).first()).toBeVisible();
  });
});
