import { requestClient } from '#/api/request';

export interface LoginLog {
  id: number;
  userName: string;
  status: number;
  ip: string;
  browser: string;
  os: string;
  msg: string;
  loginTime: string;
}

export interface LoginLogListParams {
  pageNum?: number;
  pageSize?: number;
  userName?: string;
  ip?: string;
  status?: number;
  beginTime?: string;
  endTime?: string;
  orderBy?: string;
  orderDirection?: string;
}

export interface LoginLogListResult {
  items: LoginLog[];
  total: number;
}

export async function loginLogList(params?: LoginLogListParams) {
  return await requestClient.get<LoginLogListResult>('/loginlog', { params });
}

export function loginLogDetail(id: number) {
  return requestClient.get<LoginLog>(`/loginlog/${id}`);
}

export function loginLogClean(params?: { beginTime?: string; endTime?: string }) {
  return requestClient.delete('/loginlog/clean', { params });
}

export function loginLogDelete(ids: number[]) {
  return requestClient.delete(`/loginlog/${ids.join(',')}`);
}

export function loginLogExport(params?: LoginLogListParams) {
  return requestClient.download<Blob>('/loginlog/export', { params });
}
