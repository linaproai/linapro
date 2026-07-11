import type { MenuRecordRaw } from '@vben/types';
import type { Router } from 'vue-router';

import type { RegisteredPluginPage } from './page-registry';

import { getPluginPages } from './page-registry';

function normalizePath(value: string) {
  return value
    .trim()
    .replaceAll('\\', '/')
    .replace(/^\/+/, '')
    .replace(/\/+$/u, '');
}

function toAbsolutePath(path: string) {
  return path.startsWith('/') ? path : `/${path}`;
}

/**
 * Excludes page helpers under frontend/pages/components and modal/drawer
 * companions so only real navigable management entries are considered.
 */
export function isNavigablePluginManagementPage(page: RegisteredPluginPage) {
  const pageKey = page.key.split(':').slice(1).join(':');
  if (!pageKey) {
    return false;
  }
  if (pageKey.startsWith('components/') || pageKey.includes('/components/')) {
    return false;
  }
  if (/(^|\/|-)(modal|drawer)(\.|$)/iu.test(pageKey)) {
    return false;
  }
  return Boolean(normalizePath(page.routePath));
}

export function getPluginManagementPages(pluginId: string) {
  const normalizedPluginId = pluginId.trim();
  if (!normalizedPluginId) {
    return [] as RegisteredPluginPage[];
  }
  // Keep page-registry discovery order. Do not re-sort by routePath alphabetically
  // — that can reverse menu order (e.g. /ai/invocations before /ai/providers).
  return getPluginPages().filter(
    (page) =>
      page.pluginId === normalizedPluginId &&
      isNavigablePluginManagementPage(page),
  );
}

export function hasPluginManagementPage(pluginId: string) {
  return getPluginManagementPages(pluginId).length > 0;
}

function routePathMatches(routePath: string, pageRoutePath: string) {
  const route = normalizePath(routePath);
  const page = normalizePath(pageRoutePath);
  if (!route || !page) {
    return false;
  }
  return route === page || route.endsWith(`/${page}`);
}

function pageMatchesAnyRoutePath(
  pages: RegisteredPluginPage[],
  routePath: string,
) {
  return pages.some((page) => routePathMatches(routePath, page.routePath));
}

/**
 * Depth-first menu paths in the same order as the left navigation tree.
 * This is the source of truth for "first menu" when a plugin has multiple pages.
 */
export function flattenAccessMenuPaths(menus: MenuRecordRaw[]) {
  const paths: string[] = [];
  const walk = (items: MenuRecordRaw[]) => {
    for (const item of items) {
      if (typeof item.path === 'string' && item.path.trim()) {
        paths.push(item.path);
      }
      if (item.children?.length) {
        walk(item.children);
      }
    }
  };
  walk(menus);
  return paths;
}

/**
 * Picks the first navigable management page for a plugin.
 * Prefer the first match in access menu order; fall back to the first page that
 * has a registered router path. Never prefer alphabetical routePath order.
 */
export function pickPluginManagementPage(
  pluginId: string,
  accessMenus: MenuRecordRaw[] = [],
  router?: Pick<Router, 'getRoutes'>,
) {
  const pages = getPluginManagementPages(pluginId);
  if (pages.length === 0) {
    return null;
  }

  for (const menuPath of flattenAccessMenuPaths(accessMenus)) {
    const matched = pages.find((page) =>
      routePathMatches(menuPath, page.routePath),
    );
    if (matched) {
      return matched;
    }
  }

  if (router) {
    for (const route of router.getRoutes()) {
      const matched = pages.find((page) =>
        routePathMatches(route.path, page.routePath),
      );
      if (matched) {
        return matched;
      }
    }
  }

  return pages[0] ?? null;
}

export function resolvePluginManagementPath(
  pluginId: string,
  router: Pick<Router, 'getRoutes'>,
  accessMenus: MenuRecordRaw[] = [],
) {
  const pages = getPluginManagementPages(pluginId);
  if (pages.length === 0) {
    return null;
  }

  // 1) First matching path in sidebar/access menu order (true "first menu").
  for (const menuPath of flattenAccessMenuPaths(accessMenus)) {
    if (pageMatchesAnyRoutePath(pages, menuPath)) {
      return toAbsolutePath(menuPath);
    }
  }

  // 2) Fallback: first matching registered router path in registration order.
  for (const route of router.getRoutes()) {
    if (pageMatchesAnyRoutePath(pages, route.path)) {
      return toAbsolutePath(route.path);
    }
  }

  return null;
}
