import type { RouteRecordStringComponent } from '@vben/types';

import { requestClient } from '#/api/request';

/**
 * 获取用户所有菜单
 */
export async function getAllMenusApi() {
  const response = await requestClient.get<{ list: RouteRecordStringComponent[] }>('/menus/all');
  return response.list ?? [];
}
