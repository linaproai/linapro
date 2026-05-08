import type {
  APIRequestContext,
  APIResponse,
  Page,
} from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import {
  createAdminApiContext,
  ensureSourcePluginEnabled,
  ensureSourcePluginUninstalled,
  findPlugin,
  installPlugin,
  refreshPluginProjection,
  syncPlugins,
  uninstallPlugin,
  updatePluginStatus,
} from '../../../fixtures/plugin';

type AccessibleRouteNode = {
  children?: AccessibleRouteNode[];
  meta?: {
    title?: string;
  };
};

type RuntimePluginState = {
  enabled?: number;
  id: string;
  installed?: number;
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
      await expect
        .poll(
          async () => {
            const hasMetrics =
              (await page.getByText('服务信息').first().isVisible()) &&
              (await page.getByText('服务器信息').first().isVisible());
            const hasEmptyState = await page
              .getByText('当前暂无服务监控数据。')
              .first()
              .isVisible();

            return hasMetrics || hasEmptyState ? 'available' : 'pending';
          },
          {
            message: '服务监控页面应展示指标内容或插件空状态',
            timeout: 10000,
          },
        )
        .toBe('available');
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

async function fetchRuntimePluginStates(
  adminApi: APIRequestContext,
): Promise<RuntimePluginState[]> {
  const response = await adminApi.get('plugins/dynamic');
  assertOk(response, '查询插件运行时状态失败');
  const payload = unwrapApiData(await response.json());
  return payload?.list ?? [];
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

async function expectRuntimePluginState(
  adminApi: APIRequestContext,
  pluginId: string,
  installed: number,
  enabled: number,
) {
  let plugin: RuntimePluginState | undefined;
  for (let attempt = 1; attempt <= 20; attempt += 1) {
    const items = await fetchRuntimePluginStates(adminApi);
    plugin = items.find((item) => item.id === pluginId);
    if (
      (plugin?.installed ?? 0) === installed &&
      (plugin?.enabled ?? 0) === enabled
    ) {
      return;
    }
    await new Promise((resolve) => setTimeout(resolve, 250));
  }

  expect(plugin, `未找到插件运行时状态: ${pluginId}`).toBeTruthy();
  expect(plugin?.installed ?? 0).toBe(installed);
  expect(plugin?.enabled ?? 0).toBe(enabled);
}

async function expectPluginRouteAvailable(
  page: Page,
  item: OfficialPluginCase,
) {
  let lastError: unknown;
  for (let attempt = 1; attempt <= 3; attempt += 1) {
    await page.goto(item.route, { waitUntil: 'domcontentloaded' });
    await page.waitForLoadState('networkidle').catch(() => {});
    await page
      .locator('.ant-message-notice-content', { hasText: /加载菜单中|Loading menu/i })
      .waitFor({ state: 'hidden', timeout: 5000 })
      .catch(() => {});

    try {
      await item.assertAvailable(page);
      return;
    } catch (error) {
      lastError = error;
      await refreshPluginProjection(page);
    }
  }
  throw lastError;
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
        await expectRuntimePluginState(adminApi, item.id, 0, 0);
        await refreshPluginProjection(adminPage);
        await expectMountedTitles(adminApi, item.mountedTitles, false);
        await expectPluginRouteMissing(adminPage, item.route);

        await installPlugin(adminApi, item.id);
        await expectPluginState(adminApi, item.id, 1, 0);
        await expectRuntimePluginState(adminApi, item.id, 1, 0);
        await refreshPluginProjection(adminPage);
        await expectMountedTitles(adminApi, item.mountedTitles, false);
        await expectPluginRouteMissing(adminPage, item.route);

        await updatePluginStatus(adminApi, item.id, true);
        await expectPluginState(adminApi, item.id, 1, 1);
        await expectRuntimePluginState(adminApi, item.id, 1, 1);
        await refreshPluginProjection(adminPage);
        await expectMountedTitles(adminApi, item.mountedTitles, true);
        await expectPluginRouteAvailable(adminPage, item);

        await updatePluginStatus(adminApi, item.id, false);
        await expectPluginState(adminApi, item.id, 1, 0);
        await expectRuntimePluginState(adminApi, item.id, 1, 0);
        await refreshPluginProjection(adminPage);
        await expectMountedTitles(adminApi, item.mountedTitles, false);
        await expectPluginRouteMissing(adminPage, item.route);

        await uninstallPlugin(adminApi, item.id);
        await expectPluginState(adminApi, item.id, 0, 0);
        await expectRuntimePluginState(adminApi, item.id, 0, 0);
        await refreshPluginProjection(adminPage);
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
