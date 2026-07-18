import type { Router } from 'vue-router';

import { LOGIN_PATH } from '@vben/constants';
import { preferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';
import { startProgress, stopProgress } from '@vben/utils';

import { accessRoutes, coreRouteNames } from '#/router/routes';
import { useAuthStore, useTenantStore } from '#/store';

import { generateAccess } from './access';
import {
  isPostLoginLandingIntent,
  resolvePostLoginLandingPath,
} from './post-login-landing';
import { canAccessTenantLocation } from './tenant-access';

/**
 * 通用守卫配置
 * @param router
 */
function setupCommonGuard(router: Router) {
  // 记录已经加载的页面
  const loadedPaths = new Map<string, true>();
  const maxLoadedPaths = 50;

  function touchLoadedPath(path: string) {
    if (loadedPaths.has(path)) {
      loadedPaths.delete(path);
    }
    loadedPaths.set(path, true);
    if (loadedPaths.size > maxLoadedPaths) {
      const oldestPath = loadedPaths.keys().next().value;
      if (oldestPath) {
        loadedPaths.delete(oldestPath);
      }
    }
  }

  router.beforeEach((to) => {
    const loaded = loadedPaths.has(to.path);
    to.meta.loaded = loaded;
    if (loaded) {
      touchLoadedPath(to.path);
    }

    // 页面加载进度条
    if (!to.meta.loaded && preferences.transition.progress) {
      startProgress();
    }
    return true;
  });

  router.afterEach((to) => {
    // 记录页面是否加载,如果已经加载，后续的页面切换动画等效果不在重复执行

    touchLoadedPath(to.path);

    // 关闭页面加载进度条
    if (preferences.transition.progress) {
      stopProgress();
    }
  });
}

/**
 * 权限访问守卫配置
 * @param router
 */
function setupAccessGuard(router: Router) {
  router.beforeEach(async (to, from) => {
    const accessStore = useAccessStore();
    const userStore = useUserStore();
    const authStore = useAuthStore();
    const tenantStore = useTenantStore();

    // 基本路由，这些路由不需要进入权限拦截
    if (coreRouteNames.includes(to.name as string)) {
      if (to.path === LOGIN_PATH && accessStore.accessToken) {
        return tenantStore.resolveFallbackPath(
          resolvePostLoginLandingPath({
            preferredPaths: [
              to.query?.redirect as string | undefined,
              userStore.userInfo?.homePath,
            ],
            accessibleMenus: accessStore.accessMenus,
            accessibleRoutes: accessStore.isAccessChecked
              ? accessStore.accessRoutes
              : undefined,
          }),
        );
      }
      return true;
    }

    // accessToken 检查
    if (!accessStore.accessToken) {
      // 明确声明忽略权限访问权限，则可以访问
      if (to.meta.ignoreAccess) {
        return true;
      }

      // 没有访问权限，跳转登录页面
      if (to.fullPath !== LOGIN_PATH) {
        return {
          path: LOGIN_PATH,
          // 如不需要，直接删除 query
          query:
            to.fullPath === preferences.app.defaultHomePath
              ? {}
              : { redirect: encodeURIComponent(to.fullPath) },
          // 携带当前跳转的页面，登录后重新跳转该页面
          replace: true,
        };
      }
      return to;
    }

    // 是否已经生成过动态路由
    if (accessStore.isAccessChecked) {
      if (!canAccessTenantLocation(to)) {
        return {
          path: tenantStore.resolveFallbackPath(
            resolvePostLoginLandingPath({
              preferredPaths: [userStore.userInfo?.homePath],
              accessibleMenus: accessStore.accessMenus,
              accessibleRoutes: accessStore.accessRoutes,
            }),
          ),
          replace: true,
        };
      }
      // Do not rewrite FallbackNotFound here. Missing / disabled plugin routes
      // must still render the 404 page; landing correction is only for
      // post-login / default-home intents (see generateAccess path below).
      return true;
    }

    // 生成路由表
    // 当前登录用户拥有的角色标识列表
    const userInfo = userStore.userInfo || (await authStore.fetchUserInfo());
    const userRoles = userInfo.roles ?? [];

    // 生成菜单和路由
    const { accessibleMenus, accessibleRoutes } = await generateAccess({
      roles: userRoles,
      router,
      // 则会在菜单中显示，但是访问会被重定向到403
      routes: accessRoutes,
    });

    // 保存菜单信息和路由信息
    accessStore.setAccessMenus(accessibleMenus);
    accessStore.setAccessRoutes(accessibleRoutes);
    accessStore.setIsAccessChecked(true);

    if (!canAccessTenantLocation(to)) {
      return {
        path: tenantStore.resolveFallbackPath(
          resolvePostLoginLandingPath({
            preferredPaths: [userInfo.homePath],
            accessibleMenus,
            accessibleRoutes,
          }),
        ),
        replace: true,
      };
    }

    // Landing intents only: login redirect, root, or hardcoded default home.
    // Deep links (incl. dynamic paths and intentionally missing routes) must
    // keep their target so tables load and 404 pages still appear.
    const redirectFromQuery = from.query.redirect as string | undefined;
    if (
      isPostLoginLandingIntent({
        defaultHomePath: preferences.app.defaultHomePath,
        redirectFromQuery,
        toPath: to.path,
      })
    ) {
      const landingPath = tenantStore.resolveFallbackPath(
        resolvePostLoginLandingPath({
          preferredPaths: [redirectFromQuery, userInfo.homePath],
          accessibleMenus,
          accessibleRoutes,
        }),
      );
      return {
        ...router.resolve(landingPath),
        replace: true,
      };
    }

    return {
      ...router.resolve(decodeURIComponent(to.fullPath)),
      replace: true,
    };
  });
}

/**
 * 项目守卫配置
 * @param router
 */
function createRouterGuard(router: Router) {
  /** 通用 */
  setupCommonGuard(router);
  /** 权限访问 */
  setupAccessGuard(router);
}

export { createRouterGuard };
