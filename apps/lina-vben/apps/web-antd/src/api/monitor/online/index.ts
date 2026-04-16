import type { OnlineListParams, OnlineListResult } from './model';

import { requestClient } from '#/api/request';

/** 在线用户列表 */
export function onlineList(params?: OnlineListParams) {
  return requestClient.get<OnlineListResult>('/monitor/online/list', {
    params,
  });
}

/** 强制下线 */
export function forceLogout(tokenId: string) {
  return requestClient.delete(`/monitor/online/${tokenId}`);
}
