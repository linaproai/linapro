import type { APIRequestContext, Page } from '@playwright/test';

import { request as playwrightRequest } from '@playwright/test';

import { config } from './config';
import { waitForRouteReady } from '../support/ui';

const apiBaseURL =
  process.env.E2E_API_BASE_URL ?? 'http://127.0.0.1:8080/api/v1/';

type PluginListItem = {
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

function assertOk(response: Awaited<ReturnType<APIRequestContext['get']>>, message: string) {
  if (!response.ok()) {
    throw new Error(`${message}, status=${response.status()}`);
  }
}

export async function createAdminApiContext(): Promise<APIRequestContext> {
  const loginApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const loginResponse = await loginApi.post('auth/login', {
    data: {
      password: config.adminPass,
      username: config.adminUser,
    },
  });
  assertOk(loginResponse, '管理员登录 API 失败');

  const loginResult = unwrapApiData(await loginResponse.json());
  const accessToken = loginResult?.accessToken;
  if (!accessToken) {
    throw new Error('未获取到 accessToken');
  }
  await loginApi.dispose();

  return await playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

export async function syncPlugins(adminApi: APIRequestContext) {
  const response = await adminApi.post('plugins/sync');
  assertOk(response, '同步源码插件失败');
}

export async function listPlugins(
  adminApi: APIRequestContext,
): Promise<PluginListItem[]> {
  const response = await adminApi.get('plugins');
  assertOk(response, '查询插件列表失败');
  const payload = unwrapApiData(await response.json());
  return payload?.list ?? [];
}

export async function findPlugin(
  adminApi: APIRequestContext,
  pluginId: string,
) {
  const items = await listPlugins(adminApi);
  return items.find((item) => item.id === pluginId) ?? null;
}

export async function installPlugin(
  adminApi: APIRequestContext,
  pluginId: string,
) {
  const response = await adminApi.post(`plugins/${pluginId}/install`);
  assertOk(response, `安装插件失败: ${pluginId}`);
}

export async function uninstallPlugin(
  adminApi: APIRequestContext,
  pluginId: string,
  purgeStorageData = false,
) {
  const response = await adminApi.delete(`plugins/${pluginId}`, {
    data: {
      purgeStorageData: purgeStorageData ? 1 : 0,
    },
  });
  assertOk(response, `卸载插件失败: ${pluginId}`);
}

export async function updatePluginStatus(
  adminApi: APIRequestContext,
  pluginId: string,
  enabled: boolean,
) {
  const response = await adminApi.put(
    enabled ? `plugins/${pluginId}/enable` : `plugins/${pluginId}/disable`,
  );
  assertOk(response, `更新插件状态失败: ${pluginId}`);
}

export async function refreshPluginProjection(page: Page) {
  // Always land on a stable host route before reloading so plugin lifecycle
  // changes do not leave the current page stranded on a stale dynamic route.
  await page.goto('/dashboard/analytics', { waitUntil: 'domcontentloaded' });
  await waitForRouteReady(page, 15000);
  await page.reload({ waitUntil: 'domcontentloaded' });
  await waitForRouteReady(page, 15000);
}

export async function ensureSourcePluginInstalled(page: Page, pluginId: string) {
  const adminApi = await createAdminApiContext();
  try {
    await syncPlugins(adminApi);
    const plugin = await findPlugin(adminApi, pluginId);
    if (!plugin) {
      throw new Error(`未找到插件: ${pluginId}`);
    }
    if (plugin.installed !== 1) {
      await installPlugin(adminApi, pluginId);
    }
  } finally {
    await adminApi.dispose();
  }

  await refreshPluginProjection(page);
}

export async function ensureSourcePluginEnabled(page: Page, pluginId: string) {
  const adminApi = await createAdminApiContext();
  try {
    await syncPlugins(adminApi);
    let plugin = await findPlugin(adminApi, pluginId);
    if (!plugin) {
      throw new Error(`未找到插件: ${pluginId}`);
    }
    if (plugin.installed !== 1) {
      await installPlugin(adminApi, pluginId);
      plugin = await findPlugin(adminApi, pluginId);
    }
    if (plugin?.enabled !== 1) {
      await updatePluginStatus(adminApi, pluginId, true);
    }
  } finally {
    await adminApi.dispose();
  }

  await refreshPluginProjection(page);
}

export async function ensureSourcePluginDisabled(page: Page, pluginId: string) {
  const adminApi = await createAdminApiContext();
  try {
    await syncPlugins(adminApi);
    const plugin = await findPlugin(adminApi, pluginId);
    if (plugin?.installed === 1 && plugin.enabled === 1) {
      await updatePluginStatus(adminApi, pluginId, false);
    }
  } finally {
    await adminApi.dispose();
  }

  await refreshPluginProjection(page);
}

export async function ensureSourcePluginUninstalled(page: Page, pluginId: string) {
  const adminApi = await createAdminApiContext();
  try {
    await syncPlugins(adminApi);
    const plugin = await findPlugin(adminApi, pluginId);
    if (plugin?.installed === 1) {
      await uninstallPlugin(adminApi, pluginId);
    }
  } finally {
    await adminApi.dispose();
  }

  await refreshPluginProjection(page);
}
