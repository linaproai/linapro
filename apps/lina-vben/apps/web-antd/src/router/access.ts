import type {
  ComponentRecordType,
  GenerateMenuAndRoutesOptions,
} from '@vben/types';

import { generateAccessible } from '@vben/access';
import { preferences } from '@vben/preferences';

import { message } from 'ant-design-vue';

import { getAllMenusApi } from '#/api';
import { BasicLayout, IFrameView } from '#/layouts';
import { $t } from '#/locales';
import { filterDisabledPluginRoutes } from '#/plugins/access-filter';

const forbiddenComponent = () => import('#/views/_core/fallback/forbidden.vue');

async function generateAccess(
  options: GenerateMenuAndRoutesOptions,
  { showLoadingToast = true }: { showLoadingToast?: boolean } = {},
) {
  const hostPageMap: ComponentRecordType = import.meta.glob(
    '../views/**/*.vue',
  );

  const layoutMap: ComponentRecordType = {
    BasicLayout,
    IFrameView,
  };

  return await generateAccessible(preferences.app.accessMode, {
    ...options,
    fetchMenuListAsync: async () => {
      if (showLoadingToast) {
        message.loading({
          content: `${$t('common.loadingMenu')}...`,
          duration: 1.5,
        });
      }
      const routes = await getAllMenusApi();
      return await filterDisabledPluginRoutes(routes);
    },
    // 可以指定没有权限跳转403页面
    forbiddenComponent,
    // 如果 route.meta.menuVisibleWithForbidden = true
    layoutMap,
    pageMap: hostPageMap,
  });
}

export { generateAccess };
