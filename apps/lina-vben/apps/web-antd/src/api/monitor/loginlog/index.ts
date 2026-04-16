import type { LoginLog, LoginLogListParams, LoginLogListResult } from './model';

import { requestClient } from '#/api/request';

/** 登录日志列表 */
export async function loginLogList(params?: LoginLogListParams) {
  return await requestClient.get<LoginLogListResult>('/loginlog', { params });
}

/** 登录日志详情 */
export function loginLogDetail(id: number) {
  return requestClient.get<LoginLog>(`/loginlog/${id}`);
}

/** 清除登录日志 */
export function loginLogClean(params?: {
  beginTime?: string;
  endTime?: string;
}) {
  return requestClient.delete('/loginlog/clean', { params });
}

/** 批量删除登录日志 */
export function loginLogDelete(ids: number[]) {
  return requestClient.delete(`/loginlog/${ids.join(',')}`);
}

/** 导出登录日志 */
export function loginLogExport(params?: LoginLogListParams) {
  return requestClient.download<Blob>('/loginlog/export', { params });
}
