import type { OperLog, OperLogListParams, OperLogListResult } from './model';

import { requestClient } from '#/api/request';

/** 操作日志列表 */
export async function operLogList(params?: OperLogListParams) {
  return await requestClient.get<OperLogListResult>('/operlog', { params });
}

/** 操作日志详情 */
export function operLogDetail(id: number) {
  return requestClient.get<OperLog>(`/operlog/${id}`);
}

/** 清除操作日志 */
export function operLogClean(params?: {
  beginTime?: string;
  endTime?: string;
}) {
  return requestClient.delete('/operlog/clean', { params });
}

/** 批量删除操作日志 */
export function operLogDelete(ids: number[]) {
  return requestClient.delete(`/operlog/${ids.join(',')}`);
}

/** 导出操作日志 */
export function operLogExport(params?: OperLogListParams) {
  return requestClient.download<Blob>('/operlog/export', { params });
}
