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
  listLogs,
  setCronShellEnabled,
  updateConfigValue,
} from './helpers';

test.describe('TC-88 最大执行次数自动停用', () => {
  const unlimitedJobName = `e2e_unlimited_exec_job_${Date.now()}`;
  const limitedJobName = `e2e_max_exec_job_${Date.now()}`;
  let api: APIRequestContext;
  const jobIds: number[] = [];
  let originalShellSwitch: { id: number; value: string } | null = null;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
    originalShellSwitch = await getConfigByKey(api, 'cron.shell.enabled');
    await setCronShellEnabled(api, true);
  });

  test.afterAll(async () => {
    for (const jobId of jobIds) {
      await deleteJob(api, jobId).catch(() => undefined);
    }
    if (originalShellSwitch) {
      await updateConfigValue(api, originalShellSwitch.id, originalShellSwitch.value);
    }
    await api.dispose();
  });

  test('TC-88a~b: maxExecutions=0 时仍应累计执行次数', async () => {
    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(api, buildShellJobPayload({
      groupId: defaultGroup.id,
      name: unlimitedJobName,
      status: 'enabled',
      cronExpr: '*/1 * * * * *',
      maxExecutions: 0,
    }));
    jobIds.push(created.id);

    await expect
      .poll(async () => {
        const detail = await getJob(api, created.id);
        const list = await listJobs(api, unlimitedJobName);
        const listItem = list.list.find((item) => item.id === created.id);
        return (
          detail.status === 'enabled'
          && detail.executedCount >= 1
          && (listItem?.executedCount ?? 0) >= 1
        );
      }, {
        timeout: 15000,
        message: '无限执行任务发生定时触发后仍应累计 executedCount',
      })
      .toBeTruthy();
  });

  test('TC-88c~e: 达到 maxExecutions 后任务自动停用并写入 stopReason', async () => {
    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(api, buildShellJobPayload({
      groupId: defaultGroup.id,
      name: limitedJobName,
      status: 'enabled',
      cronExpr: '*/1 * * * * *',
      maxExecutions: 1,
    }));
    jobIds.push(created.id);

    await expect
      .poll(async () => {
        const detail = await getJob(api, created.id);
        return {
          status: detail.status,
          stopReason: detail.stopReason,
          executedCount: detail.executedCount,
        };
      }, {
        timeout: 15000,
        message: '任务达到最大执行次数后应自动停用',
      })
      .toEqual({
        status: 'disabled',
        stopReason: 'max_executions_reached',
        executedCount: 1,
      });

    const logList = await listLogs(api, created.id);
    expect(logList.list.some((item) => item.status === 'success')).toBeTruthy();
  });
});
