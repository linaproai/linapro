import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';

import {
  buildShellJobPayload,
  createAdminApiContext,
  createJob,
  getConfigByKey,
  getDefaultGroup,
  getJob,
  previewCron,
  setCronShellEnabled,
  updateConfigValue,
} from '../../../support/api/job';

test.describe('TC-93 时区持久化与 Cron 预览', () => {
  const jobName = `e2e_timezone_job_${Date.now()}`;
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
      await api.delete(`job/${jobId}`);
    }
    if (originalShellSwitch) {
      await updateConfigValue(api, originalShellSwitch.id, originalShellSwitch.value);
    }
    await api.dispose();
  });

  test('TC-93a~d: 时区字段可持久化，Cron 预览支持 5 段表达式并返回对应时区结果', async () => {
    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(api, buildShellJobPayload({
      groupId: defaultGroup.id,
      name: jobName,
      cronExpr: '17 3 * * *',
      timezone: 'UTC',
      status: 'disabled',
    }));
    jobId = created.id;

    const detail = await getJob(api, jobId);
    expect(detail.timezone).toBe('UTC');
    expect(detail.cronExpr).toBe('17 3 * * *');

    const utcPreview = await previewCron(api, '17 3 * * *', 'UTC');
    expect(utcPreview.times).toHaveLength(5);
    expect(utcPreview.times.every((item) => item.endsWith('Z') || item.endsWith('+00:00'))).toBeTruthy();

    const shanghaiPreview = await previewCron(api, '17 3 * * *', 'Asia/Shanghai');
    expect(shanghaiPreview.times).toHaveLength(5);
    expect(shanghaiPreview.times.every((item) => item.endsWith('+08:00'))).toBeTruthy();
    expect(shanghaiPreview.times[0]).not.toBe(utcPreview.times[0]);
  });
});
