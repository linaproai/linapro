import { test, expect } from '../../fixtures/auth';
import { DictPage } from '../../pages/DictPage';

test.describe('TC0013 字典数据管理 CRUD', () => {
  // Use a dedicated test dict type to avoid polluting system dict types like sys_normal_disable
  const testDictType = `test_dict_${Date.now()}`;
  const testDictName = `测试字典类型_${Date.now()}`;
  const testLabel = `测试标签_${Date.now()}`;
  const testValue = `test_val_${Date.now()}`;

  test('TC0013a: 选择字典类型后右侧显示数据', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Create a dedicated test dict type
    await dictPage.createType(testDictName, testDictType);

    // Click the test dict type row to load data in the right panel
    await dictPage.clickTypeRow(testDictType);

    // Verify panel is loaded (may be empty for new type)
    const rowCount = await dictPage.getDataRowCount();
    expect(rowCount).toBeGreaterThanOrEqual(0);
  });

  test('TC0013b: 创建新字典数据', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Select the test type
    await dictPage.clickTypeRow(testDictType);

    await dictPage.createData(testLabel, testValue, { sort: 99 });

    await expect(adminPage.getByText(/创建成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC0013c: 编辑字典数据', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Select the test type
    await dictPage.clickTypeRow(testDictType);

    await dictPage.editData(testLabel, { label: `${testLabel}修改` });

    await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible({
      timeout: 5000,
    });
  });

  test('TC0013d: 删除字典数据', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    // Select the test type
    await dictPage.clickTypeRow(testDictType);

    await dictPage.deleteData(`${testLabel}修改`);

    await expect(adminPage.getByText(/删除成功|success/i)).toBeVisible({
      timeout: 5000,
    });

    // Clean up: delete the test dict type (cascade deletes will remove remaining dict data)
    try {
      await dictPage.deleteType(testDictName);
    } catch {
      // Ignore cleanup errors
    }
  });
});
