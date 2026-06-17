import type { TabDefinition } from '@vben/types';

import { useTabbarStore } from '@vben/stores';

import { getPluginPages } from './page-registry';

function normalizePath(value: unknown) {
  if (typeof value !== 'string') {
    return '';
  }
  const normalized = value.trim().replaceAll('\\', '/').split(/[?#]/u)[0] ?? '';
  return normalized.replace(/^\/+/, '').replace(/\/+$/u, '');
}

function extractAssetPluginId(value: unknown) {
  if (typeof value !== 'string') {
    return '';
  }
  return value.match(/\/x-assets\/([^/]+)\//)?.[1] ?? '';
}

function authorityMatchesPlugin(tab: TabDefinition, pluginId: string) {
  const authority = tab.meta?.authority;
  const authorityItems =
    typeof authority === 'string'
      ? [authority]
      : Array.isArray(authority)
        ? authority
        : [];
  return authorityItems.some(
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

function pluginRoutePathOwners() {
  const routePathOwners = new Map<string, string>();
  for (const page of getPluginPages()) {
    const routePath = normalizePath(page.routePath);
    if (routePath) {
      routePathOwners.set(routePath, page.pluginId);
    }
  }
  return routePathOwners;
}

function tabMatchesPlugin(
  tab: TabDefinition,
  pluginId: string,
  registeredRoutePaths: Set<string>,
  routePathOwners: Map<string, string>,
) {
  const tabPaths = [
    tab.path,
    tab.fullPath,
    tab.meta?.activePath,
    tab.meta?.link,
    ...(tab.matched?.map((item) => item.path) ?? []),
  ].map((item) => normalizePath(item));

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

  const registeredOwners = tabPaths
    .map((path) => routePathOwners.get(path))
    .filter((owner): owner is string => typeof owner === 'string' && !!owner);
  if (registeredOwners.length > 0) {
    return registeredOwners.includes(pluginId);
  }

  if (
    tabPaths.some(
      (path) =>
        path &&
        (registeredRoutePaths.has(path) || pathMatchesPluginRoute(path, pluginId)),
    )
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
  const routePathOwners = pluginRoutePathOwners();
  const registeredRoutePaths = new Set(
    [...routePathOwners.entries()]
      .filter(([, pluginId]) => pluginId === normalizedPluginId)
      .map(([routePath]) => routePath),
  );
  const staleKeys = tabbarStore.getTabs
    .filter((tab) => !tab.meta?.affixTab)
    .filter((tab) =>
      tabMatchesPlugin(
        tab,
        normalizedPluginId,
        registeredRoutePaths,
        routePathOwners,
      ),
    )
    .map((tab) => tab.key)
    .filter((key): key is string => typeof key === 'string' && key !== '');

  if (staleKeys.length > 0) {
    await tabbarStore._bulkCloseByKeys(staleKeys);
  }
}
