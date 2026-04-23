import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';

import {
  buildPayloadFromJob,
  createAdminApiContext,
  expectBusinessError,
  getJob,
  getLog,
  listJobs,
  triggerJob,
} from '../../../support/api/job';

test.describe('TC-91 源码注册任务只读保护', () => {
  let api: APIRequestContext;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
  });

  test.afterAll(async () => {
    await api.dispose();
  });

  test('TC-91a~e: 源码注册任务拒绝更新、删除、状态切换与重置，但允许立即执行', async () => {
    const jobs = await listJobs(api);
    const builtinJob = jobs.list.find((item) => item.handlerRef === 'host:cleanup-job-logs');
    expect(builtinJob).toBeTruthy();

    const originalDetail = await getJob(api, builtinJob!.id);
    const originalPayload = buildPayloadFromJob(originalDetail);

    await expectBusinessError(
      await api.put(`job/${builtinJob!.id}`, {
        data: {
          ...originalPayload,
          cronExpr: '18 3 * * *',
          env: {},
          handlerRef: '',
          params: {},
          shellCmd: "printf 'readonly update'",
          taskType: 'shell',
          workDir: '',
        },
      }),
      '源码注册任务不允许修改',
    );

    await expectBusinessError(
      await api.delete(`job/${builtinJob!.id}`),
      '源码注册任务不允许删除',
    );

    await expectBusinessError(
      await api.put(`job/${builtinJob!.id}/status`, {
        data: { status: 'disabled' },
      }),
      '源码注册任务不允许修改状态',
    );

    await expectBusinessError(
      await api.post(`job/${builtinJob!.id}/reset`),
      '源码注册任务不允许重置执行次数',
    );

    const triggered = await triggerJob(api, builtinJob!.id);
    expect(triggered.logId).toBeGreaterThan(0);

    await expect
      .poll(
        async () => {
          const detail = await getLog(api, triggered.logId);
          return detail.status;
        },
        {
          timeout: 10000,
          message: '源码注册任务应仍支持手动执行',
        },
      )
      .toBe('success');
  });
});
