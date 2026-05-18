import type {
  JobCronPreviewResult,
  JobHandlerDetail,
  JobHandlerOption,
  JobListParams,
  JobListResult,
  JobLogListParams,
  JobLogListResult,
  JobLogRecord,
  JobPayload,
  JobRecord,
  JobTriggerResult,
} from './model';

import { requestClient } from '#/api/request';

/** 获取任务列表 */
export async function jobList(params?: JobListParams) {
  const res = await requestClient.get<{ list: JobRecord[]; total: number }>(
    '/job',
    { params },
  );
  return { items: res.list, total: res.total } satisfies JobListResult;
}

/** 获取任务详情 */
export function jobDetail(id: number) {
  return requestClient.get<JobRecord>(`/job/${id}`);
}

/** 新增任务 */
export function jobCreate(data: JobPayload) {
  return requestClient.post<{ id: number }>('/job', data);
}

/** 更新任务 */
export function jobUpdate(id: number, data: JobPayload) {
  return requestClient.put(`/job/${id}`, data);
}

/** 删除任务 */
export function jobDelete(ids: Array<number> | number | string) {
  const target = Array.isArray(ids) ? ids.join(',') : ids;
  return requestClient.delete(`/job/${target}`);
}

/** 更新任务状态 */
export function jobUpdateStatus(id: number, status: 'disabled' | 'enabled') {
  return requestClient.put(`/job/${id}/status`, { status });
}

/** 手动触发任务 */
export function jobTrigger(id: number) {
  return requestClient.post<JobTriggerResult>(`/job/${id}/trigger`);
}

/** 重置任务执行计数 */
export function jobReset(id: number) {
  return requestClient.post(`/job/${id}/reset`);
}

/** 预览 Cron 表达式 */
export function jobCronPreview(expr: string, timezone: string) {
  return requestClient.get<JobCronPreviewResult>('/job/cron-preview', {
    params: { expr, timezone },
  });
}

/** 获取任务处理器列表 */
export function jobHandlerList(params?: { keyword?: string; source?: string }) {
  return requestClient.get<{ list: JobHandlerOption[] }>('/job/handler', {
    params,
  });
}

/** 获取任务处理器详情 */
export function jobHandlerDetail(ref: string) {
  return requestClient.get<JobHandlerDetail>(
    `/job/handler/${encodeURIComponent(ref)}`,
  );
}

/** 获取执行日志列表 */
export async function jobLogList(params?: JobLogListParams) {
  const res = await requestClient.get<{ list: JobLogRecord[]; total: number }>(
    '/job/log',
    { params },
  );
  return { items: res.list, total: res.total } satisfies JobLogListResult;
}

/** 获取执行日志详情 */
export function jobLogDetail(id: number) {
  return requestClient.get<JobLogRecord>(`/job/log/${id}`);
}

/** 清空执行日志 */
export function jobLogClear(jobId?: number) {
  return requestClient.delete('/job/log', {
    params: typeof jobId === 'number' ? { jobId } : undefined,
  });
}

/** 批量删除执行日志 */
export function jobLogDelete(ids: Array<number> | number | string) {
  const target = Array.isArray(ids) ? ids.join(',') : ids;
  return requestClient.delete('/job/log', {
    params: { logIds: target },
  });
}

/** 终止运行中的任务实例 */
export function jobLogCancel(id: number) {
  return requestClient.post(`/job/log/${id}/cancel`);
}
