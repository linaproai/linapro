import type { Notice, NoticeListParams } from './model';

import { requestClient } from '#/api/request';

/** 通知公告列表 */
export async function noticeList(params?: NoticeListParams) {
  const res = await requestClient.get<{ list: Notice[]; total: number }>(
    '/notice',
    { params },
  );
  return { items: res.list, total: res.total };
}

/** 新增通知公告 */
export function noticeAdd(data: Partial<Notice>) {
  return requestClient.post('/notice', data);
}

/** 更新通知公告 */
export function noticeUpdate(id: number, data: Partial<Notice>) {
  return requestClient.put(`/notice/${id}`, data);
}

/** 删除通知公告 */
export function noticeDelete(ids: string) {
  return requestClient.delete(`/notice/${ids}`);
}

/** 获取通知公告详情 */
export function noticeInfo(id: number) {
  return requestClient.get<Notice>(`/notice/${id}`);
}
