import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { JobLogPage } from '../../../pages/JobLogPage';

import {
  buildHandlerJobPayload,
  clearLogs,
  createAdminApiContext,
  createJob,
  getDefaultGroup,
  getLog,
  listLogs,
  triggerJob,
} from './helpers';

test.describe('TC-89 执行日志查询与清理', () => {
  const jobName = `e2e_job_log_${Date.now()}`;
  let api: APIRequestContext;
  let jobId = 0;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
  });

  test.afterAll(async () => {
    if (jobId) {
      await clearLogs(api, jobId);
      await api.delete(`job/${jobId}`);
    }
    await api.dispose();
  });

  test('TC-89a~d: 日志支持列表筛选、详情查看，并可按任务维度清空', async ({ adminPage }) => {
    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(api, buildHandlerJobPayload({
      groupId: defaultGroup.id,
      name: jobName,
      status: 'enabled',
    }));
    jobId = created.id;

    const triggered = await triggerJob(api, jobId);
    expect(triggered.logId).toBeGreaterThan(0);

    await expect
      .poll(async () => {
        const detail = await getLog(api, triggered.logId);
        return detail.status;
      }, {
        timeout: 10000,
        message: '日志详情应在触发后进入成功状态',
      })
      .toBe('success');

    const logList = await listLogs(api, jobId);
    expect(logList.total).toBeGreaterThanOrEqual(1);
    expect(logList.list.some((item) => item.id === triggered.logId)).toBeTruthy();

    const logDetail = await getLog(api, triggered.logId);
    expect(logDetail.jobId).toBe(jobId);
    expect(logDetail.jobName).toBe(jobName);
    expect(logDetail.trigger).toBe('manual');
    expect(logDetail.status).toBe('success');

    const jobLogPage = new JobLogPage(adminPage);
    await jobLogPage.goto();
    await jobLogPage.selectJob(jobName);
    await jobLogPage.clickSearch();
    await expect(await jobLogPage.getVisibleRowCount()).toBeGreaterThan(0);

    await jobLogPage.openFirstDetail();
    await expect(await jobLogPage.detailContains(jobName)).toBe(true);
    await expect(await jobLogPage.detailContains('manual')).toBe(true);
    await expect(await jobLogPage.detailContains('success')).toBe(true);

    await jobLogPage.clearLogs();

    await expect
      .poll(async () => {
        const cleared = await listLogs(api, jobId);
        return cleared.total;
      }, {
        timeout: 5000,
        message: '按任务清空日志后应看不到历史记录',
      })
      .toBe(0);
  });
});
