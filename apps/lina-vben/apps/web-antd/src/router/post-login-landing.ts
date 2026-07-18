/**
 * Post-login landing path resolution.
 *
 * Prefer an accessible target over the hardcoded preferences default home
 * (`/dashboard/analytics`). When the workbench menu is disabled, login must
 * land on the first navigable sidebar menu instead of a 404 page.
 */

type MenuLike = {
  children?: MenuLike[];
  disabled?: boolean;
  path?: string;
  show?: boolean;
};

type RouteLike = {
  children?: RouteLike[];
  component?: unknown;
  meta?: {
    hideInMenu?: boolean;
    link?: unknown;
  };
  path?: string;
};

/** Aligns with backend resolveHomePath empty-tree fallback. */
const SAFE_FALLBACK_HOME_PATH = '/profile';

function normalizeLandingPath(path?: null | string): string {
  if (typeof path !== 'string') {
    return '';
  }
  const trimmed = path.trim();
  if (!trimmed || trimmed === '#') {
    return '';
  }
  try {
    const decoded = decodeURIComponent(trimmed);
    if (!decoded.startsWith('/') || decoded.startsWith('//')) {
      return '';
    }
    // Drop query/hash for accessibility checks; preserve path only.
    const pathOnly = decoded.split(/[?#]/u)[0] ?? '';
    if (!pathOnly || pathOnly === '/') {
      return pathOnly === '/' ? '/' : '';
    }
    return pathOnly.replace(/\/+$/u, '') || '/';
  } catch {
    return '';
  }
}

function joinRoutePaths(parentPath: string, childPath: string): string {
  if (!childPath) {
    return parentPath;
  }
  if (childPath.startsWith('/')) {
    return childPath;
  }
  if (!parentPath || parentPath === '/') {
    return `/${childPath}`;
  }
  return `${parentPath.replace(/\/+$/u, '')}/${childPath.replace(/^\/+/u, '')}`;
}

function pathKey(path: string): string {
  const normalized = normalizeLandingPath(path);
  if (!normalized || normalized === '/') {
    return normalized;
  }
  return normalized.replace(/^\/+/u, '').replace(/\/+$/u, '');
}

function collectAccessiblePathKeys(
  routes: RouteLike[] | undefined,
  keys: Set<string> = new Set<string>(),
  parentPath = '',
): Set<string> {
  if (!routes?.length) {
    return keys;
  }
  for (const route of routes) {
    const routePath =
      typeof route.path === 'string'
        ? joinRoutePaths(parentPath, route.path)
        : parentPath;
    const key = pathKey(routePath);
    if (key) {
      keys.add(key);
    }
    if (route.children?.length) {
      collectAccessiblePathKeys(route.children, keys, routePath);
    }
  }
  return keys;
}

function isPathAccessible(
  path: string,
  accessiblePathKeys: Set<string> | null,
): boolean {
  const normalized = normalizeLandingPath(path);
  if (!normalized) {
    return false;
  }
  if (!accessiblePathKeys) {
    // Routes not assembled yet: trust non-empty preferred paths (e.g. homePath).
    return true;
  }
  if (accessiblePathKeys.size === 0) {
    return false;
  }
  const key = pathKey(normalized);
  if (accessiblePathKeys.has(key)) {
    return true;
  }
  // Allow prefix match for nested redirects like /system → /system/user.
  for (const candidate of accessiblePathKeys) {
    if (candidate.startsWith(`${key}/`) || key.startsWith(`${candidate}/`)) {
      return true;
    }
  }
  return false;
}

function isExternalOrHashLink(path: string): boolean {
  return (
    path.startsWith('http://') ||
    path.startsWith('https://') ||
    path.startsWith('//') ||
    path.startsWith('mailto:') ||
    path.startsWith('javascript:')
  );
}

/**
 * Depth-first first navigable menu path (sidebar order).
 */
function findFirstNavigableMenuPath(menus?: MenuLike[]): string {
  if (!menus?.length) {
    return '';
  }
  for (const menu of menus) {
    if (menu.disabled || menu.show === false) {
      continue;
    }
    const children = menu.children ?? [];
    if (children.length > 0) {
      const childPath = findFirstNavigableMenuPath(children);
      if (childPath) {
        return childPath;
      }
      continue;
    }
    const path = normalizeLandingPath(menu.path);
    if (!path || path === '/' || isExternalOrHashLink(path)) {
      continue;
    }
    return path;
  }
  return '';
}

/**
 * Depth-first first leaf route path from accessible dynamic routes.
 */
function findFirstNavigableRoutePath(routes?: RouteLike[]): string {
  if (!routes?.length) {
    return '';
  }
  for (const route of routes) {
    if (route.meta?.hideInMenu) {
      continue;
    }
    const children = route.children ?? [];
    const routePath =
      typeof route.path === 'string' ? normalizeLandingPath(route.path) : '';

    if (children.length > 0) {
      const nested = findFirstNavigableRoutePath(
        children.map((child) => ({
          ...child,
          path:
            typeof child.path === 'string'
              ? joinRoutePaths(routePath || '', child.path)
              : routePath,
        })),
      );
      if (nested) {
        return nested;
      }
      continue;
    }

    if (!routePath || routePath === '/' || isExternalOrHashLink(routePath)) {
      continue;
    }
    if (typeof route.meta?.link === 'string' && route.meta.link) {
      continue;
    }
    return routePath;
  }
  return '';
}

interface ResolvePostLoginLandingOptions {
  /**
   * When provided, preferred paths must exist in this route tree.
   * Omit while dynamic routes are not ready yet.
   */
  accessibleRoutes?: RouteLike[];
  /** Sidebar menus (preferred source for “first menu”). */
  accessibleMenus?: MenuLike[];
  /** Ordered candidates: query redirect, homePath, current fullPath, etc. */
  preferredPaths?: Array<null | string | undefined>;
  safeFallback?: string;
}

/**
 * Resolve the post-login / default landing path.
 *
 * Priority:
 * 1. First preferred path that is accessible (or any non-empty preferred path
 *    when accessible routes are not ready)
 * 2. First navigable sidebar menu
 * 3. First navigable accessible route
 * 4. Safe fallback (`/profile`)
 */
function resolvePostLoginLandingPath(
  options: ResolvePostLoginLandingOptions = {},
): string {
  const accessiblePathKeys =
    options.accessibleRoutes === undefined
      ? null
      : collectAccessiblePathKeys(options.accessibleRoutes);

  for (const preferred of options.preferredPaths ?? []) {
    const path = normalizeLandingPath(preferred);
    if (!path || path === '/') {
      continue;
    }
    if (isPathAccessible(path, accessiblePathKeys)) {
      return path;
    }
  }

  const firstMenu = findFirstNavigableMenuPath(options.accessibleMenus);
  if (firstMenu && isPathAccessible(firstMenu, accessiblePathKeys)) {
    return firstMenu;
  }

  const firstRoute = findFirstNavigableRoutePath(options.accessibleRoutes);
  if (firstRoute) {
    return firstRoute;
  }

  return normalizeLandingPath(options.safeFallback) || SAFE_FALLBACK_HOME_PATH;
}

export {
  SAFE_FALLBACK_HOME_PATH,
  findFirstNavigableMenuPath,
  findFirstNavigableRoutePath,
  isPathAccessible,
  normalizeLandingPath,
  resolvePostLoginLandingPath,
};
export type { MenuLike, ResolvePostLoginLandingOptions, RouteLike };
