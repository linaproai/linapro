import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';

import {
  buildHandlerJobPayload,
  createAdminApiContext,
  createJob,
  getDefaultGroup,
  getJob,
  getLog,
  listLogs,
  triggerJob,
} from './helpers';

test.describe('TC-86 手动触发任务', () => {
  const jobName = `e2e_manual_job_${Date.now()}`;
  let api: APIRequestContext;
  let jobId = 0;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
  });

  test.afterAll(async () => {
    if (jobId) {
      await api.delete(`job/${jobId}`);
    }
    await api.dispose();
  });

  test('TC-86a~d: 手动触发生成 manual 日志，且不计入 executedCount', async () => {
    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(api, buildHandlerJobPayload({
      groupId: defaultGroup.id,
      name: jobName,
      status: 'enabled',
      cronExpr: '0 0 1 1 *',
    }));
    jobId = created.id;

    const beforeDetail = await getJob(api, jobId);
    const triggered = await triggerJob(api, jobId);
    expect(triggered.logId).toBeGreaterThan(0);

    await expect
      .poll(async () => {
        const logDetail = await getLog(api, triggered.logId);
        return logDetail.status;
      }, {
        timeout: 10000,
        message: '手动触发后应生成成功日志',
      })
      .toBe('success');

    const logDetail = await getLog(api, triggered.logId);
    expect(logDetail.trigger).toBe('manual');

    const afterDetail = await getJob(api, jobId);
    expect(afterDetail.executedCount).toBe(beforeDetail.executedCount);

    const logList = await listLogs(api, jobId);
    expect(logList.list.some((item) => item.id === triggered.logId)).toBeTruthy();
  });
});
