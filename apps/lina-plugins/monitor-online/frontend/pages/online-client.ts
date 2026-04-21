import { requestClient } from '#/api/request';

export interface OnlineUser {
  tokenId: string;
  username: string;
  deptName: string;
  ip: string;
  browser: string;
  os: string;
  loginTime: string;
}

export interface OnlineListResult {
  items: OnlineUser[];
  total: number;
}

export interface OnlineListParams {
  pageNum?: number;
  pageSize?: number;
  username?: string;
  ip?: string;
}

export function onlineList(params?: OnlineListParams) {
  return requestClient.get<OnlineListResult>('/monitor/online/list', { params });
}

export function forceLogout(tokenId: string) {
  return requestClient.delete(`/monitor/online/${tokenId}`);
}
