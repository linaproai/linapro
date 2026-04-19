import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';

import {
  buildPayloadFromJob,
  createAdminApiContext,
  expectBusinessError,
  getJob,
  listJobs,
  updateJob,
} from './helpers';

test.describe('TC-91 系统内置任务保护', () => {
  let api: APIRequestContext;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
  });

  test.afterAll(async () => {
    await api.dispose();
  });

  test('TC-91a~d: 系统内置任务允许调整 cron，但锁定 handlerRef 并拒绝删除', async () => {
    const jobs = await listJobs(api);
    const builtinJob = jobs.list.find((item) => item.handlerRef === 'host:cleanup-job-logs');
    expect(builtinJob).toBeTruthy();

    const originalDetail = await getJob(api, builtinJob!.id);
    const originalPayload = buildPayloadFromJob(originalDetail);
    const updatedCronExpr = originalDetail.cronExpr === '18 3 * * *' ? '19 3 * * *' : '18 3 * * *';

    try {
      await updateJob(api, builtinJob!.id, {
        ...originalPayload,
        cronExpr: updatedCronExpr,
      });

      const updatedDetail = await getJob(api, builtinJob!.id);
      expect(updatedDetail.cronExpr).toBe(updatedCronExpr);

      await expectBusinessError(
        await api.put(`job/${builtinJob!.id}`, {
          data: {
            ...originalPayload,
            handlerRef: 'host:forbidden-handler-change',
          },
        }),
        '系统内置任务不允许修改处理器引用',
      );

      await expectBusinessError(
        await api.delete(`job/${builtinJob!.id}`),
        '系统内置任务不允许删除',
      );
    } finally {
      await updateJob(api, builtinJob!.id, originalPayload);
    }
  });
});
