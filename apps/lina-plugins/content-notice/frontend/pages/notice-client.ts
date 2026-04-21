import { requestClient } from '#/api/request';

export interface Notice {
  id: number;
  title: string;
  type: number;
  content: string;
  fileIds: string;
  status: number;
  remark: string;
  createdBy: number;
  createdByName: string;
  updatedBy: number;
  createdAt: string;
  updatedAt: string;
}

export interface NoticeListParams {
  pageNum?: number;
  pageSize?: number;
  title?: string;
  type?: number;
  createdBy?: string;
}

export async function noticeList(params?: NoticeListParams) {
  const res = await requestClient.get<{ list: Notice[]; total: number }>('/notice', { params });
  return { items: res.list, total: res.total };
}

export function noticeAdd(data: Partial<Notice>) {
  return requestClient.post('/notice', data);
}

export function noticeUpdate(id: number, data: Partial<Notice>) {
  return requestClient.put(`/notice/${id}`, data);
}

export function noticeDelete(ids: string) {
  return requestClient.delete(`/notice/${ids}`);
}

export function noticeInfo(id: number) {
  return requestClient.get<Notice>(`/notice/${id}`);
}
