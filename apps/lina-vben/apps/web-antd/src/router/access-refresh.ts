import type { Router } from 'vue-router';

import { preferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';
import { resetStaticRoutes } from '@vben/utils';

import { useAuthStore } from '#/store';
import { getPendingPluginPageRefresh } from '#/plugins/plugin-page-refresh';
import { getPluginStateMap } from '#/plugins/slot-registry';
import { useTenantStore } from '#/store/tenant';

import { generateAccess } from './access';
import { resolveAccessibleRouteRefreshTarget } from './access-refresh-route-match';
import { routes } from './routes';
import { accessRoutes } from './routes';

let accessRefreshTask: null | Promise<void> = null;
let accessRefreshQueued = false;
let accessRefreshRefreshUserInfo = false;
let accessRefreshShowLoadingToast = false;
let accessRefreshSkipRouteNavigation = false;
let accessRefreshForceDefaultRoute = false;

async function forceReplacePath(router: Router, path: string) {
  const location = router.resolve(path);
  await router.replace({
    force: true,
    hash: location.hash,
    path: location.path,
    query: location.query,
  });
}

/**
 * Refreshes menus and dynamic routes for the current logged-in user once.
 */
async function performAccessibleStateRefresh(
  router: Router,
  {
    refreshUserInfo = true,
    forceDefaultRoute = false,
    showLoadingToast = false,
    skipRouteNavigation = false,
  }: {
    forceDefaultRoute?: boolean;
    refreshUserInfo?: boolean;
    showLoadingToast?: boolean;
    skipRouteNavigation?: boolean;
  } = {},
) {
  const accessStore = useAccessStore();
  const authStore = useAuthStore();
  const tenantStore = useTenantStore();
  const userStore = useUserStore();

  if (!accessStore.accessToken) {
    return;
  }

  const currentRoute = router.currentRoute.value;
  const currentFullPath = currentRoute.fullPath;
  const userInfo = refreshUserInfo
    ? await authStore.fetchUserInfo()
    : userStore.userInfo;
  const userRoles = userInfo?.roles ?? userStore.userRoles ?? [];
  const pluginStateMap = await getPluginStateMap(true);

  resetStaticRoutes(router, routes);
  accessStore.setIsAccessChecked(false);

  const { accessibleMenus, accessibleRoutes } = await generateAccess(
    {
      roles: userRoles,
      router,
      routes: accessRoutes,
    },
    {
      pluginStateMap,
      showLoadingToast,
    },
  );

  accessStore.setAccessMenus(accessibleMenus);
  accessStore.setAccessRoutes(accessibleRoutes);
  accessStore.setIsAccessChecked(true);

  if (skipRouteNavigation) {
    return;
  }

  const fallbackPath = tenantStore.resolveFallbackPath(
    userInfo?.homePath || preferences.app.defaultHomePath || '/',
  );

  if (forceDefaultRoute) {
    await forceReplacePath(router, fallbackPath);
    return;
  }

  const resolved = router.resolve(currentFullPath);
  const accessibleMatch = resolveAccessibleRouteRefreshTarget(
    accessibleRoutes,
    resolved,
    currentRoute,
    pluginStateMap,
  );
  const pendingPluginPageRefresh = getPendingPluginPageRefresh(
    router.currentRoute.value,
  );

  if (pendingPluginPageRefresh) {
    return;
  }

  if (accessibleMatch.accessible) {
    const refreshTarget = accessibleMatch.replacementPath
      ? {
          hash: currentRoute.hash,
          path: accessibleMatch.replacementPath,
          query: currentRoute.query,
        }
      : currentFullPath;
    const refreshedLocation = router.resolve(refreshTarget);
    // Force a rematch even when the URL is unchanged so regenerated route meta
    // (for example refreshed iframe asset URLs) becomes visible immediately.
    await router.replace({
      force: true,
      hash: refreshedLocation.hash,
      path: refreshedLocation.path,
      query: refreshedLocation.query,
    });
    return;
  }

  if (router.currentRoute.value.fullPath !== fallbackPath) {
    await router.replace(fallbackPath);
  }
}

/**
 * Refreshes menus and dynamic routes for the current logged-in user.
 *
 * Multiple triggers can happen back-to-back: page CRUD, plugin state sync, and
 * focus/visibility events. The refresh must therefore run as a small queue
 * instead of a plain "single-flight" promise. Otherwise, a second refresh that
 * arrives while the first one is still fetching menus will incorrectly reuse
 * stale results and the sidebar will not reflect the latest server state.
 */
async function refreshAccessibleState(
  router: Router,
  {
    forceDefaultRoute = false,
    refreshUserInfo = true,
    showLoadingToast = false,
    skipRouteNavigation = false,
  }: {
    forceDefaultRoute?: boolean;
    refreshUserInfo?: boolean;
    showLoadingToast?: boolean;
    skipRouteNavigation?: boolean;
  } = {},
) {
  accessRefreshQueued = true;
  accessRefreshRefreshUserInfo ||= refreshUserInfo;
  accessRefreshShowLoadingToast ||= showLoadingToast;
  accessRefreshSkipRouteNavigation ||= skipRouteNavigation;
  accessRefreshForceDefaultRoute ||= forceDefaultRoute;

  if (!accessRefreshTask) {
    accessRefreshTask = (async () => {
      while (accessRefreshQueued) {
        const shouldRefreshUserInfo = accessRefreshRefreshUserInfo;
        const shouldShowLoadingToast = accessRefreshShowLoadingToast;
        const shouldForceDefaultRoute = accessRefreshForceDefaultRoute;
        const shouldSkipRouteNavigation =
          accessRefreshSkipRouteNavigation && !shouldForceDefaultRoute;
        accessRefreshQueued = false;
        accessRefreshRefreshUserInfo = false;
        accessRefreshShowLoadingToast = false;
        accessRefreshSkipRouteNavigation = false;
        accessRefreshForceDefaultRoute = false;

        await performAccessibleStateRefresh(router, {
          forceDefaultRoute: shouldForceDefaultRoute,
          refreshUserInfo: shouldRefreshUserInfo,
          showLoadingToast: shouldShowLoadingToast,
          skipRouteNavigation: shouldSkipRouteNavigation,
        });
      }
    })().finally(() => {
      accessRefreshTask = null;
      accessRefreshQueued = false;
      accessRefreshRefreshUserInfo = false;
      accessRefreshShowLoadingToast = false;
      accessRefreshSkipRouteNavigation = false;
      accessRefreshForceDefaultRoute = false;
    });
  }

  return accessRefreshTask;
}

export { refreshAccessibleState };
