import type {
  APIRequestContext,
  APIResponse,
  Page,
} from '@playwright/test';

import { expect } from '@playwright/test';

import {
  createAdminApiContext,
  ensureSourcePluginEnabledViaAPI,
  findPlugin,
  refreshPluginProjection,
  syncPlugins,
} from '../fixtures/plugin';

export type SourcePluginLifecycleCase = {
  assertAvailable: (page: Page) => Promise<void>;
  id: string;
  mountedTitles: string[];
  route: string;
};

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

export async function expectMountedTitles(
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

export async function expectPluginState(
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

export async function expectRuntimePluginState(
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

export async function expectPluginRouteAvailable(
  page: Page,
  item: SourcePluginLifecycleCase,
) {
  let lastError: unknown;
  for (let attempt = 1; attempt <= 3; attempt += 1) {
    await page.goto(item.route, { waitUntil: 'domcontentloaded' });
    await page.waitForLoadState('networkidle').catch(() => {});
    await page
      .locator('.ant-message-notice-content', {
        hasText: /加载菜单中|Loading menu/i,
      })
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

export async function expectPluginRouteMissing(page: Page, route: string) {
  await page.goto(route);
  await page.waitForLoadState('networkidle');
  await expect(page.getByText('未找到页面')).toBeVisible({ timeout: 10000 });
}

export async function smokeSourcePluginLifecycle(
  page: Page,
  item: SourcePluginLifecycleCase,
) {
  const adminApi = await createAdminApiContext();
  try {
    await syncPlugins(adminApi);
    await ensureSourcePluginEnabledViaAPI(adminApi, item.id);
    await expectPluginState(adminApi, item.id, 1, 1);
    await expectRuntimePluginState(adminApi, item.id, 1, 1);
    await expectMountedTitles(adminApi, item.mountedTitles, true);
    await refreshPluginProjection(page);
    await expectPluginRouteAvailable(page, item);
  } finally {
    await adminApi.dispose();
  }
}
