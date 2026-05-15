import { test, expect } from '../../../fixtures/auth';
import { DictPage } from '../../../pages/DictPage';

test.describe('TC0012 字典类型管理 CRUD', () => {
  function makeDictRecord(label: string) {
    const suffix = `${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
    return {
      name: `测试字典_${label}_${suffix}`,
      type: `test_dict_${label}_${suffix}`,
    };
  }

  test('TC0012a: 创建新字典类型', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    const record = makeDictRecord('create');
    await dictPage.goto();

    try {
      await dictPage.createType(record.name, record.type);

      await expect(adminPage.getByText(/创建成功|success/i)).toBeVisible({
        timeout: 5000,
      });
    } finally {
      if (await dictPage.hasType(record.name)) {
        await dictPage.deleteType(record.name);
      }
    }
  });

  test('TC0012b: 字典类型列表中可见新创建的类型', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    const record = makeDictRecord('list');
    await dictPage.goto();

    try {
      await dictPage.createType(record.name, record.type);

      // Search for the created type
      await dictPage.fillTypeSearchField('字典名称', record.name);
      await dictPage.clickTypeSearch();

      const hasType = await dictPage.hasType(record.name);
      expect(hasType).toBeTruthy();
    } finally {
      if (await dictPage.hasType(record.name)) {
        await dictPage.deleteType(record.name);
      }
    }
  });

  test('TC0012c: 编辑字典类型', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    const record = makeDictRecord('edit');
    const updatedName = `${record.name}_修改`;
    await dictPage.goto();

    try {
      await dictPage.createType(record.name, record.type);
      await dictPage.editType(record.name, { name: updatedName });

      await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible({
        timeout: 5000,
      });
    } finally {
      if (await dictPage.hasType(updatedName)) {
        await dictPage.deleteType(updatedName);
      } else if (await dictPage.hasType(record.name)) {
        await dictPage.deleteType(record.name);
      }
    }
  });

  test('TC0012d: 删除字典类型', async ({ adminPage }) => {
    const dictPage = new DictPage(adminPage);
    const record = makeDictRecord('delete');
    await dictPage.goto();
    await dictPage.createType(record.name, record.type);
    const response = await dictPage.deleteType(record.name);

    expect(response.status()).toBe(200);
    expect(await dictPage.hasType(record.name)).toBeFalsy();
  });
});
