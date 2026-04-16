import { test, expect } from '../../fixtures/auth';
import { DictPage } from '../../pages/DictPage';

test.describe('TC0012 字典类型管理 CRUD', () => {
  const testTypeName = `测试字典_${Date.now()}`;
  const testTypeCode = `test_dict_${Date.now()}`;

  test('TC0012a: 创建新字典类型', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();
    await dictPage.createType(testTypeName, testTypeCode);

    await expect(adminPage.getByText(/创建成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC0012b: 字典类型列表中可见新创建的类型', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Search for the created type
    await dictPage.fillTypeSearchField('字典名称', testTypeName);
    await dictPage.clickTypeSearch();

    const hasType = await dictPage.hasType(testTypeName);
    expect(hasType).toBeTruthy();
  });

  test('TC0012c: 编辑字典类型', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();
    await dictPage.editType(testTypeName, { name: `${testTypeName}修改` });

    await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC0012d: 删除字典类型', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();
    await dictPage.deleteType(`${testTypeName}修改`);

    await expect(adminPage.getByText(/删除成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });
});
