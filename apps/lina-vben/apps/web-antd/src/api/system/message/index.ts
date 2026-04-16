import type { UserMessage } from './model';

import { requestClient } from '#/api/request';

/** 获取未读消息数量 */
export async function messageUnreadCount() {
  const res = await requestClient.get<{ count: number }>(
    '/user/message/count',
  );
  return res.count;
}

/** 获取消息列表 */
export async function messageList(params?: {
  pageNum?: number;
  pageSize?: number;
}) {
  const res = await requestClient.get<{
    list: UserMessage[];
    total: number;
  }>('/user/message', { params });
  return { items: res.list ?? [], total: res.total };
}

/** 标记消息已读 */
export function messageRead(id: number) {
  return requestClient.put(`/user/message/${id}/read`);
}

/** 标记全部消息已读 */
export function messageReadAll() {
  return requestClient.put('/user/message/read-all');
}

/** 删除单条消息 */
export function messageDelete(id: number) {
  return requestClient.delete(`/user/message/${id}`);
}

/** 清空全部消息 */
export function messageClear() {
  return requestClient.delete('/user/message/clear');
}
