import type { BasicUserInfo, UserInfo } from '@vben/types';

import { requestClient } from '#/api/request';

/**
 * 当前项目 `/user/info` 接口返回的完整用户信息。
 */
export interface AppUserInfo extends UserInfo, BasicUserInfo {
  roles?: string[];
}

/**
 * 获取用户信息
 */
export async function getUserInfoApi() {
  return requestClient.get<AppUserInfo>('/user/info');
}
