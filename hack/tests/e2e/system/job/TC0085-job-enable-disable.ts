import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';

import {
  buildHandlerJobPayload,
  createAdminApiContext,
  createJob,
  getDefaultGroup,
  getJob,
  updateJobStatus,
} from './helpers';

test.describe('TC-85 定时任务启停', () => {
  const jobName = `e2e_toggle_job_${Date.now()}`;
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

  test('TC-85a~c: 任务支持从停用切换到启用，再切回停用', async () => {
    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(api, buildHandlerJobPayload({
      groupId: defaultGroup.id,
      name: jobName,
      status: 'disabled',
    }));
    jobId = created.id;

    let detail = await getJob(api, jobId);
    expect(detail.status).toBe('disabled');

    await updateJobStatus(api, jobId, 'enabled');
    detail = await getJob(api, jobId);
    expect(detail.status).toBe('enabled');

    await updateJobStatus(api, jobId, 'disabled');
    detail = await getJob(api, jobId);
    expect(detail.status).toBe('disabled');
  });
});
