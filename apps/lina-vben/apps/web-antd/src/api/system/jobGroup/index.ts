import type {
  JobGroupListParams,
  JobGroupPayload,
  JobGroupRecord,
} from './model';

import { requestClient } from '#/api/request';

/** 获取任务分组列表 */
export async function jobGroupList(params?: JobGroupListParams) {
  const res = await requestClient.get<{
    list: JobGroupRecord[];
    total: number;
  }>('/job-group', { params });
  return { items: res.list, total: res.total };
}

/** 新增任务分组 */
export function jobGroupCreate(data: JobGroupPayload) {
  return requestClient.post<{ id: number }>('/job-group', data);
}

/** 更新任务分组 */
export function jobGroupUpdate(id: number, data: JobGroupPayload) {
  return requestClient.put(`/job-group/${id}`, data);
}

/** 删除任务分组 */
export function jobGroupDelete(ids: Array<number> | number | string) {
  const target = Array.isArray(ids) ? ids.join(',') : ids;
  return requestClient.delete(`/job-group/${target}`);
}
