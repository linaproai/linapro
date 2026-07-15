import type { PluginDynamicState } from '#/api/system/plugin/model';
import type { RouteRecordRaw } from 'vue-router';

interface AccessibleRouteTargets {
  hostedPaths: Map<string, string>;
  names: Set<string>;
  paths: Set<string>;
}

interface ResolvedRouteLike {
  matched: Array<{
    meta?: Record<string, unknown>;
    name?: unknown;
    path?: string;
  }>;
  meta?: Record<string, unknown>;
  path?: string;
  query?: Record<string, unknown>;
}

type AccessibleRouteLike = Pick<
  RouteRecordRaw,
  'children' | 'meta' | 'name' | 'path'
>;

interface AccessibleRouteRefreshTarget {
  accessible: boolean;
  replacementPath?: string;
}

function normalizeRoutePath(path: string) {
  const normalized = path.replaceAll('\\', '/').replace(/^\/+/, '');
  return normalized === '/' ? '' : normalized.replace(/\/+$/, '');
}

function normalizeHostedTarget(value: unknown) {
  if (typeof value !== 'string' || !value) {
    return '';
  }
  const match = value.match(/\/x-assets\/[^?#]+/);
  return match?.[0] ? match[0].replace(/\/+$/, '') : '';
}

function runtimeStateBlocksNewEntry(runtimeState: unknown) {
  return typeof runtimeState === 'string' && runtimeState !== 'normal';
}

function joinRoutePaths(parentPath: string, childPath: string) {
  if (!childPath) {
    return parentPath;
  }
  if (childPath.startsWith('/')) {
    return childPath;
  }
  if (!parentPath || parentPath === '/') {
    return `/${childPath}`;
  }
  return `${parentPath.replace(/\/+$/, '')}/${childPath.replace(/^\/+/, '')}`;
}

function collectAccessibleRouteTargets(
  routeList: AccessibleRouteLike[],
  targets: AccessibleRouteTargets = {
    hostedPaths: new Map<string, string>(),
    names: new Set<string>(),
    paths: new Set<string>(),
  },
  parentPath = '',
) {
  for (const route of routeList) {
    if (typeof route.name === 'string' && route.name) {
      targets.names.add(route.name);
    }

    const routePath =
      typeof route.path === 'string'
        ? joinRoutePaths(parentPath, route.path)
        : parentPath;
    if (routePath) {
      targets.paths.add(normalizeRoutePath(routePath));
    }

    const hostedTarget = normalizeHostedTarget(route.meta?.iframeSrc);
    if (hostedTarget && routePath && !targets.hostedPaths.has(hostedTarget)) {
      targets.hostedPaths.set(hostedTarget, routePath);
    }

    if (route.children?.length) {
      collectAccessibleRouteTargets(
        route.children as AccessibleRouteLike[],
        targets,
        routePath,
      );
    }
  }
  return targets;
}

function resolveHostedTargetFromRoute(route?: ResolvedRouteLike) {
  if (!route) {
    return '';
  }
  const candidates: unknown[] = [
    route.meta?.iframeSrc,
    route.meta?.link,
    route.query?.embeddedSrc,
    route.path,
  ];
  for (const matched of route.matched) {
    candidates.push(matched.meta?.iframeSrc, matched.meta?.link, matched.path);
  }
  for (const candidate of candidates) {
    const hostedTarget = normalizeHostedTarget(candidate);
    if (hostedTarget) {
      return hostedTarget;
    }
  }
  return '';
}

function routeRecordMatchesTargets(
  route: ResolvedRouteLike['matched'][number],
  targets: AccessibleRouteTargets,
) {
  if (typeof route.name === 'string' && targets.names.has(route.name)) {
    return true;
  }
  if (typeof route.path === 'string') {
    return targets.paths.has(normalizeRoutePath(route.path));
  }
  return false;
}

function resolveAccessibleRouteRefreshTarget(
  accessibleRoutes: RouteRecordRaw[],
  resolved: ResolvedRouteLike,
  currentRoute?: ResolvedRouteLike,
  pluginStateMap?: Map<string, PluginDynamicState>,
): AccessibleRouteRefreshTarget {
  const targets = collectAccessibleRouteTargets(accessibleRoutes);
  if (
    resolved.matched.some((route) => routeRecordMatchesTargets(route, targets))
  ) {
    return { accessible: true };
  }
  if (typeof resolved.path === 'string') {
    if (targets.paths.has(normalizeRoutePath(resolved.path))) {
      return { accessible: true };
    }
  }

  const hostedTarget =
    resolveHostedTargetFromRoute(currentRoute) ||
    resolveHostedTargetFromRoute(resolved);
  const replacementPath = hostedTarget
    ? targets.hostedPaths.get(hostedTarget)
    : undefined;
  if (replacementPath) {
    return {
      accessible: true,
      replacementPath,
    };
  }

  const pluginId = hostedTarget.match(/\/x-assets\/([^/]+)\//)?.[1] ?? '';
  const pluginState = pluginId ? pluginStateMap?.get(pluginId) : null;
  if (
    pluginState?.installed === 1 &&
    pluginState.enabled === 1 &&
    runtimeStateBlocksNewEntry(pluginState.runtimeState)
  ) {
    return { accessible: true };
  }

  return { accessible: false };
}

/**
 * Decides whether an access refresh should navigate after menus/routes are rebuilt.
 *
 * Silent mode keeps the current page component mounted when the route remains
 * reachable. Force rematch is reserved for path corrections and explicit
 * default-route resets; pending plugin page refreshes keep their own remount path.
 */
type AccessRefreshNavigationDecision =
  | { kind: 'fallback' }
  | { kind: 'force-default' }
  | { kind: 'replace-path'; path: string }
  | { kind: 'silent' };

function resolveAccessRefreshNavigation(options: {
  accessibleMatch: AccessibleRouteRefreshTarget;
  currentPath: string;
  forceDefaultRoute?: boolean;
  hasPendingPluginPageRefresh?: boolean;
  skipRouteNavigation?: boolean;
}): AccessRefreshNavigationDecision {
  const {
    accessibleMatch,
    currentPath,
    forceDefaultRoute = false,
    hasPendingPluginPageRefresh = false,
    skipRouteNavigation = false,
  } = options;

  if (forceDefaultRoute) {
    return { kind: 'force-default' };
  }
  if (skipRouteNavigation || hasPendingPluginPageRefresh) {
    return { kind: 'silent' };
  }
  if (!accessibleMatch.accessible) {
    return { kind: 'fallback' };
  }

  const replacementPath = accessibleMatch.replacementPath;
  if (
    replacementPath &&
    normalizeRoutePath(replacementPath) !== normalizeRoutePath(currentPath)
  ) {
    return { kind: 'replace-path', path: replacementPath };
  }

  // Current route is still reachable. Rebuild menus/routes silently so host
  // static pages (e.g. plugin management) are not remounted by force rematch.
  return { kind: 'silent' };
}

export {
  resolveAccessRefreshNavigation,
  resolveAccessibleRouteRefreshTarget,
};
export type { AccessRefreshNavigationDecision };
