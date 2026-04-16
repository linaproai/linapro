import type { RouteRecordRaw, Router } from 'vue-router';

import { preferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';
import { resetStaticRoutes } from '@vben/utils';

import { useAuthStore } from '#/store';
import { getPendingPluginPageRefresh } from '#/plugins/plugin-page-refresh';

import { generateAccess } from './access';
import { routes } from './routes';
import { accessRoutes } from './routes';

let accessRefreshTask: null | Promise<void> = null;
let accessRefreshQueued = false;
let accessRefreshShowLoadingToast = false;

function collectAccessibleRouteNames(
  routeList: RouteRecordRaw[],
  names: Set<string> = new Set(),
) {
  for (const route of routeList) {
    if (typeof route.name === 'string' && route.name) {
      names.add(route.name);
    }
    if (route.children?.length) {
      collectAccessibleRouteNames(route.children, names);
    }
  }
  return names;
}

/**
 * Refreshes menus and dynamic routes for the current logged-in user once.
 */
async function performAccessibleStateRefresh(
  router: Router,
  {
    showLoadingToast = false,
    skipRouteNavigation = false,
  }: {
    showLoadingToast?: boolean;
    skipRouteNavigation?: boolean;
  } = {},
) {
  const accessStore = useAccessStore();
  const authStore = useAuthStore();
  const userStore = useUserStore();

  if (!accessStore.accessToken) {
    return;
  }

  const currentFullPath = router.currentRoute.value.fullPath;
  const userInfo = await authStore.fetchUserInfo();
  const userRoles = userStore.userInfo?.roles ?? [];

  resetStaticRoutes(router, routes);
  accessStore.setIsAccessChecked(false);

  const { accessibleMenus, accessibleRoutes } = await generateAccess(
    {
      roles: userRoles,
      router,
      routes: accessRoutes,
    },
    {
      showLoadingToast,
    },
  );

  accessStore.setAccessMenus(accessibleMenus);
  accessStore.setAccessRoutes(accessibleRoutes);
  accessStore.setIsAccessChecked(true);

  if (skipRouteNavigation) {
    return;
  }

  const accessibleNames = collectAccessibleRouteNames(accessibleRoutes);
  const resolved = router.resolve(currentFullPath);
  const hasAccessibleMatch = resolved.matched.some((route) => {
    return typeof route.name === 'string' && accessibleNames.has(route.name);
  });
  const pendingPluginPageRefresh = getPendingPluginPageRefresh(
    router.currentRoute.value,
  );

  if (pendingPluginPageRefresh) {
    return;
  }

  if (hasAccessibleMatch) {
    const refreshedLocation = router.resolve(currentFullPath);
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

  const fallbackPath =
    userInfo.homePath || preferences.app.defaultHomePath || '/';
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
    showLoadingToast = false,
    skipRouteNavigation = false,
  }: {
    showLoadingToast?: boolean;
    skipRouteNavigation?: boolean;
  } = {},
) {
  accessRefreshQueued = true;
  accessRefreshShowLoadingToast ||= showLoadingToast;

  if (!accessRefreshTask) {
    accessRefreshTask = (async () => {
      while (accessRefreshQueued) {
        const shouldShowLoadingToast = accessRefreshShowLoadingToast;
        accessRefreshQueued = false;
        accessRefreshShowLoadingToast = false;

        await performAccessibleStateRefresh(router, {
          showLoadingToast: shouldShowLoadingToast,
          skipRouteNavigation,
        });
      }
    })().finally(() => {
      accessRefreshTask = null;
      accessRefreshQueued = false;
      accessRefreshShowLoadingToast = false;
    });
  }

  return accessRefreshTask;
}

export { refreshAccessibleState };
