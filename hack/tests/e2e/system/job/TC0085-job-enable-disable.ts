import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { JobPage } from '../../../pages/JobPage';

import {
  buildShellJobPayload,
  createAdminApiContext,
  createJob,
  getConfigByKey,
  getDefaultGroup,
  getJob,
  setCronShellEnabled,
  updateConfigValue,
} from './helpers';

test.describe('TC-85 定时任务启停', () => {
  const jobName = `e2e_toggle_job_${Date.now()}`;
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

  test('TC-85a~d: 任务支持通过编辑弹窗从停用切换到启用，再切回停用', async ({ adminPage }) => {
    const jobPage = new JobPage(adminPage);
    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(api, buildShellJobPayload({
      groupId: defaultGroup.id,
      name: jobName,
      status: 'disabled',
    }));
    jobId = created.id;

    let detail = await getJob(api, jobId);
    expect(detail.status).toBe('disabled');

    await jobPage.goto();
    await jobPage.fillSearchKeyword(jobName);
    await jobPage.clickSearch();
    await expect(await jobPage.hasAction('job-enable-')).toBe(false);

    await jobPage.openEditSearchedJob();
    await jobPage.setTaskStatus('启用');
    await jobPage.save();
    await expect
      .poll(async () => (await getJob(api, jobId)).status, {
        timeout: 5000,
        message: '编辑任务并将状态设为启用后应成功生效',
      })
      .toBe('enabled');

    await jobPage.openEditSearchedJob();
    await jobPage.setTaskStatus('停用');
    await jobPage.save();
    await expect
      .poll(async () => (await getJob(api, jobId)).status, {
        timeout: 5000,
        message: '编辑任务并将状态设为停用后应成功生效',
      })
      .toBe('disabled');
  });
});
