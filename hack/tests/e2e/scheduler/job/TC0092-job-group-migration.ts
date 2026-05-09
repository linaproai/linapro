import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';

import {
  buildShellJobPayload,
  createAdminApiContext,
  createGroup,
  createJob,
  deleteGroup,
  getConfigByKey,
  restoreCronShellEnabled,
  getDefaultGroup,
  getJob,
  setCronShellEnabled,
} from '../../../support/api/job';

test.describe('TC-92 删除分组自动迁移任务', () => {
  const groupCode = `e2e_migrate_group_${Date.now()}`;
  const groupName = `E2E迁移分组_${Date.now()}`;
  const jobName = `e2e_migrate_job_${Date.now()}`;
  let api: APIRequestContext;
  let groupId = 0;
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
    if (groupId) {
      await api.delete(`job-group/${groupId}`);
    }
    if (originalShellSwitch) {
      await restoreCronShellEnabled(api, originalShellSwitch);
    }
    await api.dispose();
  });

  test('TC-92a~c: 删除非默认分组后，分组下任务迁移到默认分组', async () => {
    const defaultGroup = await getDefaultGroup(api);
    const createdGroup = await createGroup(api, {
      code: groupCode,
      name: groupName,
      remark: 'migration target',
      sortOrder: 30,
    });
    groupId = createdGroup.id;

    const createdJob = await createJob(api, buildShellJobPayload({
      groupId,
      name: jobName,
      status: 'disabled',
    }));
    jobId = createdJob.id;

    await deleteGroup(api, groupId);
    groupId = 0;

    const migratedJob = await getJob(api, jobId);
    expect(migratedJob.groupId).toBe(defaultGroup.id);
    expect(migratedJob.groupCode).toBe(defaultGroup.code);
    expect(migratedJob.groupName).toBe(defaultGroup.name);
  });
});
