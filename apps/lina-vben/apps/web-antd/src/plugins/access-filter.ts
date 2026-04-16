import type { PluginDynamicState } from '#/api/system/plugin/model';
import type { RouteRecordStringComponent } from '@vben/types';

import { getPluginPages } from '#/plugins/page-registry';
import { getPluginStateMap } from '#/plugins/slot-registry';

type RouteLike = Pick<
  RouteRecordStringComponent,
  'children' | 'component' | 'meta' | 'path'
>;

function normalizePath(path?: string) {
  return typeof path === 'string' ? path.replace(/^\//, '') : '';
}

function buildDisabledPluginSet(items: PluginDynamicState[]) {
  const disabledPluginIds = new Set<string>();
  for (const item of items) {
    if (item.installed !== 1 || item.enabled !== 1) {
      disabledPluginIds.add(item.id);
    }
  }
  return disabledPluginIds;
}

function buildExactRoutePluginMap() {
  const routePluginMap = new Map<string, string>();
  for (const page of getPluginPages()) {
    routePluginMap.set(page.routePath, page.pluginId);
  }
  return routePluginMap;
}

function resolveAuthority(item: RouteLike) {
  const authority = item.meta?.authority;
  if (typeof authority === 'string') {
    return authority;
  }
  if (Array.isArray(authority)) {
    return authority.join(',');
  }
  return '';
}

function resolveComponent(item: RouteLike) {
  return typeof item.component === 'string' ? item.component : '';
}

function resolvePluginId(
  item: RouteLike,
  exactRoutePluginMap: Map<string, string>,
  pluginIds: string[],
) {
  const normalizedPath = normalizePath(item.path);
  const exactPluginId = normalizedPath
    ? exactRoutePluginMap.get(normalizedPath)
    : undefined;
  if (exactPluginId) {
    return exactPluginId;
  }

  const authority = resolveAuthority(item);
  const component = resolveComponent(item);

  for (const pluginId of pluginIds) {
    if (
      normalizedPath === pluginId ||
      normalizedPath.startsWith(`${pluginId}-`) ||
      normalizedPath.startsWith(`plugins/${pluginId}/`)
    ) {
      return pluginId;
    }
    if (authority === pluginId || authority.startsWith(`${pluginId}:`)) {
      return pluginId;
    }
    if (
      component.includes(`/plugins/${pluginId}/`) ||
      component.includes(`lina-plugins/${pluginId}/`)
    ) {
      return pluginId;
    }
  }

  return '';
}

function filterRouteTree<T extends RouteLike>(
  routes: T[],
  exactRoutePluginMap: Map<string, string>,
  disabledPluginIds: Set<string>,
  pluginIds: string[],
): T[] {
  return routes.flatMap((route) => {
    const nextChildren = Array.isArray(route.children)
      ? filterRouteTree(
          route.children as T[],
          exactRoutePluginMap,
          disabledPluginIds,
          pluginIds,
        )
      : route.children;

    const pluginId = resolvePluginId(route, exactRoutePluginMap, pluginIds);
    if (pluginId && disabledPluginIds.has(pluginId)) {
      return [];
    }

    if (
      Array.isArray(route.children) &&
      route.children.length > 0 &&
      (!nextChildren || nextChildren.length === 0) &&
      !route.component
    ) {
      return [];
    }

    if (nextChildren !== route.children) {
      return [{ ...route, children: nextChildren }];
    }

    return [route];
  });
}

export async function filterDisabledPluginRoutes<T extends RouteLike>(
  routes: T[],
) {
  if (routes.length === 0) {
    return routes;
  }

  try {
    const pluginStateMap = await getPluginStateMap();
    const runtimeStates: PluginDynamicState[] = [];
    const seenPluginIDs = new Set<string>();
    for (const item of pluginStateMap.values()) {
      if (!item?.id || seenPluginIDs.has(item.id)) {
        continue;
      }
      seenPluginIDs.add(item.id);
      runtimeStates.push(item);
    }
    if (runtimeStates.length === 0) {
      return routes;
    }

    const disabledPluginIds = buildDisabledPluginSet(runtimeStates);
    if (disabledPluginIds.size === 0) {
      return routes;
    }

    const pluginIds = runtimeStates.map((item) => item.id).filter(Boolean);
    const exactRoutePluginMap = buildExactRoutePluginMap();
    return filterRouteTree(
      routes,
      exactRoutePluginMap,
      disabledPluginIds,
      pluginIds,
    );
  } catch (error) {
    console.error('[plugin-access] failed to filter plugin routes', error);
    return routes;
  }
}
