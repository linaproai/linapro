import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';

import {
  buildShellJobPayload,
  clearLogs,
  createAdminApiContext,
  createJob,
  getConfigByKey,
  getDefaultGroup,
  getLog,
  setCronShellEnabled,
  triggerJob,
  updateConfigValue,
} from '../../../support/api/job';

test.describe('TC-95 定时任务超时', () => {
  const jobName = `e2e_job_timeout_${Date.now()}`;
  let api: APIRequestContext;
  let jobId = 0;
  let originalShellSwitch: { id: number; value: string } | null = null;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
    originalShellSwitch = await getConfigByKey(api, 'cron.shell.enabled');
    await setCronShellEnabled(api, true);
  });

  test.afterAll(async () => {
    if (jobId) {
      await clearLogs(api, jobId);
      await api.delete(`job/${jobId}`);
    }
    if (originalShellSwitch) {
      await updateConfigValue(api, originalShellSwitch.id, originalShellSwitch.value);
    }
    await api.dispose();
  });

  test('TC-95a~c: 任务超时后应记录 timeout 状态和超时时长', async () => {
    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(api, buildShellJobPayload({
      groupId: defaultGroup.id,
      name: jobName,
      shellCmd: 'sleep 2',
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
        message: '超时任务应生成 timeout 日志',
      })
      .toBe('timeout');

    const logDetail = await getLog(api, triggered.logId);
    expect(logDetail.errMsg ?? '').toContain('1s');
    expect(logDetail.errMsg ?? '').toContain('timed out');
  });
});
