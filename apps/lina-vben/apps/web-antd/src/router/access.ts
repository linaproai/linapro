import type {
  ComponentRecordType,
  GenerateMenuAndRoutesOptions,
  RouteRecordRaw,
} from '@vben/types';

import { generateAccessible } from '@vben/access';
import { preferences } from '@vben/preferences';

import { message } from 'ant-design-vue';

import { getAllMenusApi } from '#/api';
import { BasicLayout, IFrameView } from '#/layouts';
import { $t } from '#/locales';
import { filterDisabledPluginRoutes } from '#/plugins/access-filter';

import { filterTenantAccessRoutes } from './tenant-access';

const forbiddenComponent = () => import('#/views/_core/fallback/forbidden.vue');

async function generateAccess(
  options: GenerateMenuAndRoutesOptions,
  { showLoadingToast = true }: { showLoadingToast?: boolean } = {},
) {
  const accessRoutes = filterTenantAccessRoutes(options.routes);
  const hiddenFrontendRoutes = collectHiddenFrontendRoutes(accessRoutes);
  const hostPageMap: ComponentRecordType = import.meta.glob(
    '../views/**/*.vue',
  );

  const layoutMap: ComponentRecordType = {
    BasicLayout,
    IFrameView,
  };

  const result = await generateAccessible(preferences.app.accessMode, {
    ...options,
    routes: accessRoutes,
    fetchMenuListAsync: async () => {
      if (showLoadingToast) {
        message.loading({
          content: `${$t('common.loadingMenu')}...`,
          duration: 1.5,
        });
      }
      const routes = await getAllMenusApi();
      return await filterDisabledPluginRoutes(
        filterTenantAccessRoutes(routes),
      );
    },
    // 可以指定没有权限跳转403页面
    forbiddenComponent,
    // 如果 route.meta.menuVisibleWithForbidden = true
    layoutMap,
    pageMap: hostPageMap,
  });

  const registeredHiddenRoutes = registerHiddenRoutes(
    options.router,
    hiddenFrontendRoutes,
  );

  return {
    ...result,
    accessibleRoutes: [...result.accessibleRoutes, ...registeredHiddenRoutes],
  };
}

function collectHiddenFrontendRoutes(routes: RouteRecordRaw[]): RouteRecordRaw[] {
  return routes.flatMap((route) => {
    const hiddenCurrent =
      route.meta?.hideInMenu && route.component
        ? [
            {
              ...route,
              meta: route.meta ? { ...route.meta } : undefined,
            },
          ]
        : [];

    return [
      ...hiddenCurrent,
      ...collectHiddenFrontendRoutes(route.children ?? []),
    ];
  });
}

function registerHiddenRoutes(
  router: GenerateMenuAndRoutesOptions['router'],
  routes: RouteRecordRaw[],
): RouteRecordRaw[] {
  const registered: RouteRecordRaw[] = [];

  for (const route of routes) {
    const routeName = route.name;
    if (typeof routeName !== 'string' || router.hasRoute(routeName)) {
      continue;
    }

    const hiddenRoute: RouteRecordRaw = {
      ...route,
      meta: route.meta ? { ...route.meta } : undefined,
    };
    router.addRoute('Root', hiddenRoute);
    registered.push(hiddenRoute);
  }

  return registered;
}

export { generateAccess };
