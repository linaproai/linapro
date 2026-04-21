import type {
  APIRequestContext,
  APIResponse,
  Page,
} from '@playwright/test';

import { test, expect } from '../../fixtures/auth';
import {
  createAdminApiContext,
  ensureSourcePluginEnabled,
  ensureSourcePluginUninstalled,
  findPlugin,
  installPlugin,
  syncPlugins,
  uninstallPlugin,
  updatePluginStatus,
} from '../../fixtures/plugin';

type AccessibleRouteNode = {
  children?: AccessibleRouteNode[];
  meta?: {
    title?: string;
  };
};

type OfficialPluginCase = {
  assertAvailable: (page: Page) => Promise<void>;
  id: string;
  mountedTitles: string[];
  route: string;
};

const officialPluginCases: OfficialPluginCase[] = [
  {
    id: 'monitor-operlog',
    mountedTitles: ['操作日志'],
    route: '/monitor/operlog',
    assertAvailable: async (page) => {
      await expect(page.locator('.vxe-table')).toBeVisible({ timeout: 10000 });
      await expect(page.getByRole('button', { name: /清\s*空/ })).toBeVisible();
    },
  },
  {
    id: 'monitor-loginlog',
    mountedTitles: ['登录日志'],
    route: '/monitor/loginlog',
    assertAvailable: async (page) => {
      await expect(page.locator('.vxe-table')).toBeVisible({ timeout: 10000 });
      await expect(page.getByRole('button', { name: /导\s*出/ })).toBeVisible();
    },
  },
  {
    id: 'monitor-server',
    mountedTitles: ['服务监控'],
    route: '/monitor/server',
    assertAvailable: async (page) => {
      await expect(page.getByText('服务信息').first()).toBeVisible({
        timeout: 10000,
      });
      await expect(page.getByText('服务器信息').first()).toBeVisible();
    },
  },
  {
    id: 'monitor-online',
    mountedTitles: ['在线用户'],
    route: '/monitor/online',
    assertAvailable: async (page) => {
      await expect(page.locator('.vxe-table')).toBeVisible({ timeout: 10000 });
      await expect(page.getByText(/在线用户列表/)).toBeVisible();
    },
  },
  {
    id: 'org-center',
    mountedTitles: ['部门管理', '岗位管理'],
    route: '/system/dept',
    assertAvailable: async (page) => {
      await expect(page.locator('.vxe-table')).toBeVisible({ timeout: 10000 });
      await expect(page.getByRole('button', { name: /新\s*增/ }).first()).toBeVisible();
    },
  },
  {
    id: 'content-notice',
    mountedTitles: ['通知公告'],
    route: '/system/notice',
    assertAvailable: async (page) => {
      await expect(page.locator('.vxe-table')).toBeVisible({ timeout: 10000 });
      await expect(page.getByRole('button', { name: /新\s*增/ }).first()).toBeVisible();
    },
  },
];

function unwrapApiData(payload: any) {
  if (payload && typeof payload === 'object' && 'data' in payload) {
    return payload.data;
  }
  return payload;
}

function assertOk(response: APIResponse, message: string) {
  expect(response.ok(), `${message}, status=${response.status()}`).toBeTruthy();
}

async function fetchCurrentUserRoutes(
  adminApi: APIRequestContext,
): Promise<AccessibleRouteNode[]> {
  const response = await adminApi.get('menus/all');
  assertOk(response, '查询当前用户动态路由失败');
  const payload = unwrapApiData(await response.json());
  return payload?.list ?? [];
}

function hasRouteTitle(list: AccessibleRouteNode[], title: string): boolean {
  return list.some((item) => {
    if (item?.meta?.title === title) {
      return true;
    }
    return hasRouteTitle(item?.children ?? [], title);
  });
}

async function expectMountedTitles(
  adminApi: APIRequestContext,
  titles: string[],
  mounted: boolean,
) {
  const routes = await fetchCurrentUserRoutes(adminApi);
  for (const title of titles) {
    expect(
      hasRouteTitle(routes, title),
      `${mounted ? '缺少' : '不应存在'}菜单标题: ${title}`,
    ).toBe(mounted);
  }
}

async function expectPluginState(
  adminApi: APIRequestContext,
  pluginId: string,
  installed: number,
  enabled: number,
) {
  const plugin = await findPlugin(adminApi, pluginId);
  expect(plugin, `未找到插件: ${pluginId}`).toBeTruthy();
  expect(plugin?.installed ?? 0).toBe(installed);
  expect(plugin?.enabled ?? 0).toBe(enabled);
}

async function expectPluginRouteAvailable(
  page: Page,
  item: OfficialPluginCase,
) {
  await page.goto(item.route);
  await page.waitForLoadState('networkidle');
  await item.assertAvailable(page);
}

async function expectPluginRouteMissing(page: Page, route: string) {
  await page.goto(route);
  await page.waitForLoadState('networkidle');
  await expect(page.getByText('未找到页面')).toBeVisible({ timeout: 10000 });
}

test.describe('TC-98 官方源码插件生命周期', () => {
  for (const [index, item] of officialPluginCases.entries()) {
    const suffix = String.fromCharCode('a'.charCodeAt(0) + index);

    test(`TC0098${suffix}: ${item.id} 支持安装、启用、停用、卸载与菜单挂载切换`, async ({
      adminPage,
    }) => {
      const adminApi = await createAdminApiContext();

      try {
        await syncPlugins(adminApi);

        await ensureSourcePluginUninstalled(adminPage, item.id);
        await expectPluginState(adminApi, item.id, 0, 0);
        await expectMountedTitles(adminApi, item.mountedTitles, false);
        await expectPluginRouteMissing(adminPage, item.route);

        await installPlugin(adminApi, item.id);
        await adminPage.reload({ waitUntil: 'networkidle' });
        await adminPage.waitForTimeout(300);
        await expectPluginState(adminApi, item.id, 1, 0);
        await expectMountedTitles(adminApi, item.mountedTitles, false);
        await expectPluginRouteMissing(adminPage, item.route);

        await updatePluginStatus(adminApi, item.id, true);
        await adminPage.reload({ waitUntil: 'networkidle' });
        await adminPage.waitForTimeout(300);
        await expectPluginState(adminApi, item.id, 1, 1);
        await expectMountedTitles(adminApi, item.mountedTitles, true);
        await expectPluginRouteAvailable(adminPage, item);

        await updatePluginStatus(adminApi, item.id, false);
        await adminPage.reload({ waitUntil: 'networkidle' });
        await adminPage.waitForTimeout(300);
        await expectPluginState(adminApi, item.id, 1, 0);
        await expectMountedTitles(adminApi, item.mountedTitles, false);
        await expectPluginRouteMissing(adminPage, item.route);

        await uninstallPlugin(adminApi, item.id);
        await adminPage.reload({ waitUntil: 'networkidle' });
        await adminPage.waitForTimeout(300);
        await expectPluginState(adminApi, item.id, 0, 0);
        await expectMountedTitles(adminApi, item.mountedTitles, false);
        await expectPluginRouteMissing(adminPage, item.route);
      } finally {
        await adminApi.dispose();
        await adminPage.goto('/dashboard/analysis');
        await ensureSourcePluginEnabled(adminPage, item.id);
      }
    });
  }
});
