import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../fixtures/auth';
import { MenuPage } from '../../pages/MenuPage';
import { createAdminApiContext, expectSuccess } from '../../support/api/job';

type MenuListItem = {
  id: number;
  name: string;
  path: string;
  perms: string;
  children?: MenuListItem[];
};

type MenuTreeSelectItem = {
  id: number;
  label: string;
  children?: MenuTreeSelectItem[];
};

type MenuDetail = {
  id: number;
  name: string;
  parentName: string;
};

type RoleListItem = {
  id: number;
  key: string;
};

function flattenMenuList(list: MenuListItem[]): MenuListItem[] {
  return list.flatMap((item) => [item, ...flattenMenuList(item.children ?? [])]);
}

function flattenMenuTreeSelect(list: MenuTreeSelectItem[]): MenuTreeSelectItem[] {
  return list.flatMap((item) => [item, ...flattenMenuTreeSelect(item.children ?? [])]);
}

test.describe('TC0114 菜单治理标题国际化专项回归', () => {
  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  test('TC0114a: 英文环境下菜单管理列表显示本地化菜单标题', async ({
    adminPage,
    mainLayout,
  }) => {
    const menuPage = new MenuPage(adminPage);

    await mainLayout.switchLanguage('English');
    await menuPage.goto();
    const searchInput = adminPage.getByRole('textbox', {
      name: /菜单名称|Menu Name/i,
    });
    await searchInput.fill('menus');
    await adminPage.getByRole('button', { name: /搜索|Search/i }).click();

    await expect(
      adminPage.locator('.vxe-body--row', { hasText: 'Menus' }).first(),
    ).toBeVisible();
  });

  test('TC0114b: 英文环境下菜单树与角色菜单树接口返回本地化标题', async () => {
    const localizedList = await expectSuccess<{ list: MenuListItem[] }>(
      await adminApi.get('menu', {
        headers: {
          'Accept-Language': 'en-US',
        },
      }),
    );
    const flatMenus = flattenMenuList(localizedList.list);

    const settingsCatalog = flatMenus.find((item) => item.path === 'setting');
    expect(settingsCatalog?.name).toBe('System Settings');

    const menuManagement = flatMenus.find((item) => item.perms === 'system:menu:list');
    expect(menuManagement?.name).toBe('Menus');

    const treeSelect = await expectSuccess<{ list: MenuTreeSelectItem[] }>(
      await adminApi.get('menu/treeselect', {
        headers: {
          'Accept-Language': 'en-US',
        },
      }),
    );
    const flatTreeSelect = flattenMenuTreeSelect(treeSelect.list);
    const localizedTreeNode = flatTreeSelect.find((item) => item.id === menuManagement?.id);
    expect(localizedTreeNode?.label).toBe('Menus');

    const roles = await expectSuccess<{ list: RoleListItem[]; total: number }>(
      await adminApi.get('role', {
        params: {
          key: 'admin',
          page: 1,
          size: 10,
        },
      }),
    );
    const adminRole = roles.list.find((item) => item.key === 'admin');
    expect(adminRole, 'missing admin role').toBeTruthy();

    const roleMenuTree = await expectSuccess<{
      menus: MenuTreeSelectItem[];
      checkedKeys: number[];
    }>(
      await adminApi.get(`menu/role/${adminRole!.id}`, {
        headers: {
          'Accept-Language': 'en-US',
        },
      }),
    );
    const flatRoleMenus = flattenMenuTreeSelect(roleMenuTree.menus);
    const localizedRoleNode = flatRoleMenus.find((item) => item.id === menuManagement?.id);
    expect(localizedRoleNode?.label).toBe('Menus');
  });

  test('TC0114c: 英文环境下菜单详情保留可编辑原值并本地化父级名称', async () => {
    const localizedList = await expectSuccess<{ list: MenuListItem[] }>(
      await adminApi.get('menu', {
        headers: {
          'Accept-Language': 'en-US',
        },
      }),
    );
    const flatMenus = flattenMenuList(localizedList.list);
    const menuManagement = flatMenus.find((item) => item.perms === 'system:menu:list');

    expect(menuManagement, 'missing system:menu:list menu').toBeTruthy();

    const detail = await expectSuccess<MenuDetail>(
      await adminApi.get(`menu/${menuManagement!.id}`, {
        headers: {
          'Accept-Language': 'en-US',
        },
      }),
    );

    expect(detail.name).toBe('菜单管理');
    expect(detail.parentName).toBe('Access Management');
  });
});
