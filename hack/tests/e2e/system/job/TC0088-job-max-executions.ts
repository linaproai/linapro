import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';

import {
  buildHandlerJobPayload,
  createAdminApiContext,
  createJob,
  getDefaultGroup,
  getJob,
  listLogs,
} from './helpers';

test.describe('TC-88 最大执行次数自动停用', () => {
  const jobName = `e2e_max_exec_job_${Date.now()}`;
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

  test('TC-88a~c: 达到 maxExecutions 后任务自动停用并写入 stopReason', async () => {
    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(api, buildHandlerJobPayload({
      groupId: defaultGroup.id,
      name: jobName,
      status: 'enabled',
      cronExpr: '*/1 * * * * *',
      maxExecutions: 1,
    }));
    jobId = created.id;

    await expect
      .poll(async () => {
        const detail = await getJob(api, jobId);
        return {
          status: detail.status,
          stopReason: detail.stopReason,
          executedCount: detail.executedCount,
        };
      }, {
        timeout: 15000,
        message: '任务达到最大执行次数后应自动停用',
      })
      .toEqual({
        status: 'disabled',
        stopReason: 'max_executions_reached',
        executedCount: 1,
      });

    const logList = await listLogs(api, jobId);
    expect(logList.list.some((item) => item.status === 'success')).toBeTruthy();
  });
});
