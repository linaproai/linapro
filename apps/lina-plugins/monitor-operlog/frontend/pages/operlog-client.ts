import { requestClient } from '#/api/request';

export interface OperLog {
  id: number;
  title: string;
  operSummary: string;
  operType: string;
  method: string;
  requestMethod: string;
  operName: string;
  operUrl: string;
  operIp: string;
  operParam: string;
  jsonResult: string;
  status: number;
  errorMsg: string;
  costTime: number;
  operTime: string;
}

export interface OperLogListParams {
  pageNum?: number;
  pageSize?: number;
  title?: string;
  operName?: string;
  operType?: string;
  status?: number;
  beginTime?: string;
  endTime?: string;
  orderBy?: string;
  orderDirection?: string;
}

export interface OperLogListResult {
  items: OperLog[];
  total: number;
}

export async function operLogList(params?: OperLogListParams) {
  return await requestClient.get<OperLogListResult>('/operlog', { params });
}

export function operLogDetail(id: number) {
  return requestClient.get<OperLog>(`/operlog/${id}`);
}

export function operLogClean(params?: { beginTime?: string; endTime?: string }) {
  return requestClient.delete('/operlog/clean', { params });
}

export function operLogDelete(ids: number[]) {
  return requestClient.delete(`/operlog/${ids.join(',')}`);
}

export function operLogExport(params?: OperLogListParams) {
  return requestClient.download<Blob>('/operlog/export', { params });
}
