import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';

import {
  buildShellJobPayload,
  cancelLog,
  clearLogs,
  createAdminApiContext,
  createJob,
  getConfigByKey,
  restoreCronShellEnabled,
  getDefaultGroup,
  getLog,
  setCronShellEnabled,
  triggerJob,
} from '../../../support/api/job';

test.describe('TC-87 长任务手动终止', () => {
  const jobName = `e2e_cancel_job_${Date.now()}`;
  let api: APIRequestContext;
  let jobId = 0;
  let originalShellSwitch: { id: number; value: string } | null = null;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
    originalShellSwitch = await getConfigByKey(api, 'cron.shell.enabled');
    await setCronShellEnabled(api, true);
    await expect
      .poll(async () => (await getConfigByKey(api, 'cron.shell.enabled')).value)
      .toBe('true');
  });

  test.afterAll(async () => {
    if (jobId) {
      await clearLogs(api, jobId);
      await api.delete(`job/${jobId}`);
    }
    if (originalShellSwitch) {
      await restoreCronShellEnabled(api, originalShellSwitch);
    }
    await api.dispose();
  });

  test('TC-87a~d: 手动终止运行中的 Shell 任务后，日志状态应为 cancelled', async () => {
    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(api, buildShellJobPayload({
      groupId: defaultGroup.id,
      name: jobName,
      shellCmd: 'sleep 30',
      timeoutSeconds: 60,
    }));
    jobId = created.id;

    const triggered = await triggerJob(api, jobId);
    expect(triggered.logId).toBeGreaterThan(0);

    await expect
      .poll(async () => (await getLog(api, triggered.logId)).status, {
        timeout: 5000,
        message: '长任务在取消前应先进入 running 状态',
      })
      .toBe('running');

    await cancelLog(api, triggered.logId);

    await expect
      .poll(async () => (await getLog(api, triggered.logId)).status, {
        timeout: 10000,
        message: '终止后日志状态应变为 cancelled',
      })
      .toBe('cancelled');

    const logDetail = await getLog(api, triggered.logId);
    expect(logDetail.errMsg ?? '').toContain('context canceled');
    expect(logDetail.resultJson ?? '').toContain('"cancelled":true');
  });
});
