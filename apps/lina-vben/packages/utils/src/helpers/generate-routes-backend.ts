import type { RouteRecordRaw } from 'vue-router';

import type {
  ComponentRecordType,
  GenerateMenuAndRoutesOptions,
  RouteRecordStringComponent,
} from '@vben-core/typings';

import { mapTree } from '@vben-core/shared/utils';

/**
 * 判断路由是否在菜单中显示但访问时展示 403（让用户知悉功能并申请权限）
 */
function menuHasVisibleWithForbidden(route: RouteRecordRaw): boolean {
  return !!route.meta?.menuVisibleWithForbidden;
}

/**
 * 动态生成路由 - 后端方式
 * 对 meta.menuVisibleWithForbidden 为 true 的项直接替换为 403 组件，让用户知悉功能并申请权限。
 */
async function generateRoutesByBackend(
  options: GenerateMenuAndRoutesOptions,
): Promise<RouteRecordRaw[]> {
  const {
    fetchMenuListAsync,
    layoutMap = {},
    pageMap = {},
    forbiddenComponent,
  } = options;

  try {
    const menuRoutes = await fetchMenuListAsync?.();
    if (!menuRoutes) {
      return [];
    }

    const normalizePageMap: ComponentRecordType = {};

    for (const [key, value] of Object.entries(pageMap)) {
      normalizePageMap[normalizeViewPath(key)] = value;
    }

    let routes = convertRoutes(menuRoutes, layoutMap, normalizePageMap);

    if (forbiddenComponent) {
      routes = mapTree(routes, (route) => {
        if (menuHasVisibleWithForbidden(route)) {
          route.component = forbiddenComponent;
        }
        return route;
      });
    }

    return routes;
  } catch (error) {
    console.error(error);
    throw error;
  }
}

function convertRoutes(
  routes: RouteRecordStringComponent[],
  layoutMap: ComponentRecordType,
  pageMap: ComponentRecordType,
): RouteRecordRaw[] {
  return mapTree(routes, (node) => {
    const route = node as unknown as RouteRecordRaw;
    const { component, name } = node;

    if (!name) {
      console.error('route name is required', route);
    }

    // layout转换
    if (component && layoutMap[component]) {
      route.component = layoutMap[component];
      // 页面组件转换
    } else if (component) {
      const pageKey = normalizeViewPath(component);
      if (pageMap[pageKey]) {
        route.component = pageMap[pageKey];
      } else {
        console.error(`route component is invalid: ${pageKey}`, route);
        route.component = pageMap['/_core/fallback/not-found.vue'];
      }
    }

    return route;
  });
}

function normalizeViewPath(path: string): string {
  // Remove # prefix if present (used by some backend implementations)
  let normalizedPath = path;
  if (normalizedPath.startsWith('#')) {
    normalizedPath = normalizedPath.substring(1);
  }
  normalizedPath = normalizedPath.replaceAll('\\', '/');

  // 去除相对路径前缀
  normalizedPath = normalizedPath.replace(/^(\.\/|\.\.\/)+/, '');

  const pluginPageMatch = normalizedPath.match(
    /(?:^|\/)lina-plugins\/([^/]+)\/frontend\/src\/pages\/(.+)$/,
  );
  if (pluginPageMatch?.[1] && pluginPageMatch[2]) {
    const pluginId = pluginPageMatch[1];
    let pluginPagePath = pluginPageMatch[2];
    if (!pluginPagePath.endsWith('.vue')) {
      pluginPagePath = `${pluginPagePath}.vue`;
    }
    return `/plugins/${pluginId}/pages/${pluginPagePath}`;
  }

  // 确保路径以 '/' 开头
  const viewPath = normalizedPath.startsWith('/')
    ? normalizedPath
    : `/${normalizedPath}`;

  // Remove /views prefix and handle the path
  // Input: /views/system/menu/index or /system/menu/index
  // Output: /system/menu/index
  let result = viewPath.replace(/^\/views/, '');

  // Add .vue extension if not present
  if (!result.endsWith('.vue')) {
    result = `${result}.vue`;
  }

  return result;
}
export { generateRoutesByBackend };
