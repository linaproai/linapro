import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';

import {
  buildShellJobPayload,
  createAdminApiContext,
  createJob,
  deleteJob,
  getConfigByKey,
  getDefaultGroup,
  getJob,
  listJobs,
  setCronShellEnabled,
  updateConfigValue,
  updateJob,
} from '../../../support/api/job';

test.describe('TC-83 Shell 类型任务 CRUD', () => {
  const jobName = `e2e_shell_job_${Date.now()}`;
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
      await deleteJob(api, jobId);
    }
    if (originalShellSwitch) {
      await updateConfigValue(api, originalShellSwitch.id, originalShellSwitch.value);
    }
    await api.dispose();
  });

  test('TC-83a~e: Shell 任务支持新增、详情、编辑与删除', async () => {
    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(api, buildShellJobPayload({
      groupId: defaultGroup.id,
      name: jobName,
      shellCmd: "printf 'shell-create'",
      workDir: '/tmp',
      env: {
        E2E_MODE: 'create',
      },
      status: 'disabled',
    }));
    jobId = created.id;

    const detail = await getJob(api, jobId);
    expect(detail.taskType).toBe('shell');
    expect(detail.handlerRef).toBe('');
    expect(detail.shellCmd).toBe("printf 'shell-create'");
    expect(detail.workDir).toBe('/tmp');
    expect(JSON.parse(detail.env ?? '{}')).toEqual({ E2E_MODE: 'create' });

    await updateJob(api, jobId, buildShellJobPayload({
      groupId: defaultGroup.id,
      name: `${jobName}_updated`,
      shellCmd: "printf 'shell-update'",
      workDir: '/tmp',
      env: {
        E2E_MODE: 'updated',
        TRACE: '1',
      },
      status: 'disabled',
    }));

    const updatedDetail = await getJob(api, jobId);
    expect(updatedDetail.name).toBe(`${jobName}_updated`);
    expect(updatedDetail.shellCmd).toBe("printf 'shell-update'");
    expect(JSON.parse(updatedDetail.env ?? '{}')).toEqual({
      E2E_MODE: 'updated',
      TRACE: '1',
    });

    await deleteJob(api, jobId);
    jobId = 0;

    const jobs = await listJobs(api, jobName);
    expect(jobs.list.some((item) => item.name.includes(jobName))).toBeFalsy();
  });
});
