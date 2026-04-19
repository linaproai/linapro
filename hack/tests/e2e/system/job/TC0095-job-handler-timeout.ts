import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';

import {
  buildHandlerJobPayload,
  clearLogs,
  createAdminApiContext,
  createJob,
  getDefaultGroup,
  getLog,
  triggerJob,
} from './helpers';

test.describe('TC-95 Handler 任务超时', () => {
  const jobName = `e2e_handler_timeout_${Date.now()}`;
  let api: APIRequestContext;
  let jobId = 0;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
  });

  test.afterAll(async () => {
    if (jobId) {
      await clearLogs(api, jobId);
      await api.delete(`job/${jobId}`);
    }
    await api.dispose();
  });

  test('TC-95a~c: Handler 任务超时后应记录 timeout 状态和超时时长', async () => {
    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(api, buildHandlerJobPayload({
      groupId: defaultGroup.id,
      name: jobName,
      handlerRef: 'host:wait',
      params: {
        seconds: 2,
      },
      timeoutSeconds: 1,
      status: 'enabled',
      cronExpr: '0 0 1 1 *',
    }));
    jobId = created.id;

    const triggered = await triggerJob(api, jobId);
    expect(triggered.logId).toBeGreaterThan(0);

    await expect
      .poll(async () => (await getLog(api, triggered.logId)).status, {
        timeout: 10000,
        message: '超时 Handler 任务应生成 timeout 日志',
      })
      .toBe('timeout');

    const logDetail = await getLog(api, triggered.logId);
    expect(logDetail.errMsg ?? '').toContain('1s');
    expect(logDetail.errMsg ?? '').toContain('超时');
  });
});
