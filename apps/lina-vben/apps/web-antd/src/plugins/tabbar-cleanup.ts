import type { TabDefinition } from '@vben/types';

import { useTabbarStore } from '@vben/stores';

import { getPluginPages } from './page-registry';

function normalizePath(value: unknown) {
  if (typeof value !== 'string') {
    return '';
  }
  return value.trim().replace(/^\//, '');
}

function extractAssetPluginId(value: unknown) {
  if (typeof value !== 'string') {
    return '';
  }
  return value.match(/\/x-assets\/([^/]+)\//)?.[1] ?? '';
}

function authorityMatchesPlugin(tab: TabDefinition, pluginId: string) {
  const authority = tab.meta?.authority;
  if (!Array.isArray(authority)) {
    return false;
  }
  return authority.some(
    (item) =>
      typeof item === 'string' &&
      (item === pluginId || item.startsWith(`${pluginId}:`)),
  );
}

function pathMatchesPluginRoute(path: string, pluginId: string) {
  return (
    path === pluginId ||
    path.startsWith(`${pluginId}-`) ||
    path.startsWith(`plugins/${pluginId}/`)
  );
}

function pluginRoutePaths(pluginId: string) {
  return new Set(
    getPluginPages()
      .filter((page) => page.pluginId === pluginId)
      .map((page) => normalizePath(page.routePath))
      .filter(Boolean),
  );
}

function tabMatchesPlugin(
  tab: TabDefinition,
  pluginId: string,
  registeredRoutePaths: Set<string>,
) {
  const tabPaths = [
    tab.path,
    tab.fullPath,
    tab.meta?.activePath,
    tab.meta?.link,
    ...(tab.matched?.map((item) => item.path) ?? []),
  ].map((item) => normalizePath(item));

  if (
    tabPaths.some(
      (path) =>
        path &&
        (registeredRoutePaths.has(path) || pathMatchesPluginRoute(path, pluginId)),
    )
  ) {
    return true;
  }

  if (
    [
      tab.fullPath,
      tab.meta?.iframeSrc,
      tab.meta?.link,
      (tab.meta?.query as Record<string, unknown> | undefined)?.embeddedSrc,
    ].some((item) => extractAssetPluginId(item) === pluginId)
  ) {
    return true;
  }

  return authorityMatchesPlugin(tab, pluginId);
}

export async function closePluginTabs(pluginId: string) {
  const normalizedPluginId = pluginId.trim();
  if (!normalizedPluginId) {
    return;
  }

  const tabbarStore = useTabbarStore();
  const registeredRoutePaths = pluginRoutePaths(normalizedPluginId);
  const staleKeys = tabbarStore.getTabs
    .filter((tab) => !tab.meta?.affixTab)
    .filter((tab) =>
      tabMatchesPlugin(tab, normalizedPluginId, registeredRoutePaths),
    )
    .map((tab) => tab.key)
    .filter((key): key is string => typeof key === 'string' && key !== '');

  if (staleKeys.length > 0) {
    await tabbarStore._bulkCloseByKeys(staleKeys);
  }
}
