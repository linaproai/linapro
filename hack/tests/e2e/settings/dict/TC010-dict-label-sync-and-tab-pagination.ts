import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { DictPage } from '../../../pages/DictPage';
import { MainLayout } from '../../../pages/MainLayout';
import { UserPage } from '../../../pages/UserPage';
import { createAdminApiContext, expectSuccess } from '../../../support/api/job';
import { waitForRouteReady } from '../../../support/ui';

type ListResult<T> = {
  list: T[];
  total: number;
};

type DictDataItem = {
  cssClass: string;
  dictType: string;
  id: number;
  isBuiltin: number;
  label: string;
  remark: string;
  sort: number;
  status: number;
  tagStyle: string;
  value: string;
};

type MenuItem = {
  children?: MenuItem[];
  isCache: number;
  menuKey?: string;
  perms?: string;
};

const routeCacheMenuKeys = [
  'system:menu:list',
  'system:dict:list',
  'system:config:list',
];

function flattenMenus(list: MenuItem[]): MenuItem[] {
  return list.flatMap((item) => [item, ...flattenMenus(item.children ?? [])]);
}

async function createDictType(
  api: APIRequestContext,
  name: string,
  type: string,
) {
  const result = await expectSuccess<{ id: number }>(
    await api.post('dict/type', {
      data: {
        name,
        remark: '',
        status: 1,
        type,
      },
    }),
  );
  return result.id;
}

async function createDictData(
  api: APIRequestContext,
  dictType: string,
  label: string,
  value: string,
  sort: number,
) {
  const result = await expectSuccess<{ id: number }>(
    await api.post('dict/data', {
      data: {
        cssClass: '',
        dictType,
        label,
        remark: '',
        sort,
        status: 1,
        tagStyle: 'primary',
        value,
      },
    }),
  );
  return result.id;
}

async function deleteDictType(api: APIRequestContext, id: number) {
  await expectSuccess(await api.delete(`dict/type/${id}`));
}

async function listDictData(api: APIRequestContext, dictType: string) {
  return expectSuccess<ListResult<DictDataItem>>(
    await api.get(
      `dict/data?pageNum=1&pageSize=100&dictType=${encodeURIComponent(
        dictType,
      )}`,
    ),
  );
}

async function updateDictData(
  api: APIRequestContext,
  item: DictDataItem,
  patch: Partial<DictDataItem>,
) {
  await expectSuccess(
    await api.put(`dict/data/${item.id}`, {
      data: {
        cssClass: patch.cssClass ?? item.cssClass,
        dictType: patch.dictType ?? item.dictType,
        label: patch.label ?? item.label,
        remark: patch.remark ?? item.remark,
        sort: patch.sort ?? item.sort,
        status: patch.status ?? item.status,
        tagStyle: patch.tagStyle ?? item.tagStyle,
        value: patch.value ?? item.value,
      },
    }),
  );
}

async function expectBuiltInRouteCacheEnabled(api: APIRequestContext) {
  const menus = await expectSuccess<ListResult<MenuItem>>(await api.get('menu'));
  const menuByKey = new Map(
    flattenMenus(menus.list).map((item) => [item.menuKey ?? item.perms, item]),
  );

  for (const key of routeCacheMenuKeys) {
    const menu = menuByKey.get(key);
    expect(menu, `${key} should exist`).toBeTruthy();
    expect(menu!.isCache, `${key} should enable route cache`).toBe(1);
  }
}

test.describe('TC-10 字典标签同步与 Tab 分页保持', () => {
  let api: APIRequestContext;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
  });

  test.afterAll(async () => {
    await api.dispose();
  });

  // Menu status switches use static i18n labels after the cascade-status UI
  // change. User management still binds status switch text to
  // sys_normal_disable, so it is the correct surface for dict label sync.
  test('TC-10a: 修改字典标签后已打开用户列表同步显示最新标签', async ({
    adminPage,
  }) => {
    const dictList = await listDictData(api, 'sys_normal_disable');
    const normalData = dictList.list.find((item) => item.value === '1');
    expect(normalData, 'sys_normal_disable=1 should exist').toBeTruthy();

    const originalLabel = normalData!.label;
    const updatedLabel = `同步后_${Date.now()}`;
    const userPage = new UserPage(adminPage);
    const dictPage = new DictPage(adminPage);
    const userStatusSwitch = (label: string) =>
      adminPage.locator('.ant-switch').filter({ hasText: label });

    try {
      await userPage.goto();
      await expect(userStatusSwitch(originalLabel).first()).toBeVisible();

      await dictPage.goto();
      await dictPage.clickTypeRow('sys_normal_disable');
      await dictPage.editData(originalLabel, { label: updatedLabel });
      await expect(adminPage.getByText(/更新成功|success/i)).toBeVisible();

      await new MainLayout(adminPage).tabTitle(/用户管理|Users/i).click();
      await waitForRouteReady(adminPage);

      await expect(userStatusSwitch(updatedLabel).first()).toBeVisible();
      await expect(userStatusSwitch(originalLabel)).toHaveCount(0);
    } finally {
      const latest = await listDictData(api, 'sys_normal_disable');
      const current = latest.list.find((item) => item.value === '1');
      if (current && current.label !== originalLabel) {
        await updateDictData(api, current, { label: originalLabel });
      }
    }
  });

  test('TC-10b: 切换工作台 Tab 后字典数据分页保持当前页', async ({
    adminPage,
  }) => {
    await expectBuiltInRouteCacheEnabled(api);

    const stamp = Date.now();
    const dictType = `tab_keep_${stamp}`;
    const dictName = `Tab分页保持_${stamp}`;
    const dictTypeId = await createDictType(api, dictName, dictType);

    try {
      for (let index = 1; index <= 25; index += 1) {
        await createDictData(
          api,
          dictType,
          `分页选项_${stamp}_${String(index).padStart(2, '0')}`,
          `value_${index}`,
          index,
        );
      }

      const dictPage = new DictPage(adminPage);
      await dictPage.goto();
      await dictPage.clickTypeRow(dictType);
      await dictPage.gotoDataPage(2);
      await expect.poll(() => dictPage.getDataActivePage()).toBe(2);

      const mainLayout = new MainLayout(adminPage);
      await mainLayout.expandSidebarGroup(/系统设置|Settings/i);
      await mainLayout.sidebarMenuItem(/参数设置|Parameters/i).click();
      await waitForRouteReady(adminPage);
      await mainLayout.tabTitle('字典管理').click();
      await waitForRouteReady(adminPage);

      await expect.poll(() => dictPage.getDataActivePage()).toBe(2);
    } finally {
      await deleteDictType(api, dictTypeId).catch(() => {});
    }
  });
});
