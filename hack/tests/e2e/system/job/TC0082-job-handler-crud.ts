import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { JobPage } from '../../../pages/JobPage';

import {
  createAdminApiContext,
  deleteJob,
  getDefaultGroup,
  getJob,
  listJobs,
} from './helpers';

test.describe('TC-82 Handler 类型任务 CRUD', () => {
  const jobName = `e2e_handler_job_${Date.now()}`;
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

  test('TC-82a~e: Handler 列表可查询，任务支持新增、详情、编辑、删除', async ({ adminPage }) => {
    const defaultGroup = await getDefaultGroup(api);
    const jobPage = new JobPage(adminPage);
    await jobPage.goto();

    await jobPage.openCreate();
    await jobPage.fillCommonFields({
      groupName: defaultGroup.name,
      name: jobName,
      description: 'handler job before update',
      cronExpr: '0 0 1 1 *',
    });
    await jobPage.selectHandler('等待指定时长 (host:wait)');
    await jobPage.fillHandlerParam('等待秒数', '1');
    await jobPage.save();

    await jobPage.fillSearchKeyword(jobName);
    await jobPage.clickSearch();
    await expect(await jobPage.hasJob(jobName)).toBe(true);

    const createdList = await listJobs(api, jobName);
    const created = createdList.list.find((item) => item.name === jobName);
    expect(created).toBeTruthy();
    jobId = created!.id;

    const detail = await getJob(api, jobId);
    expect(detail.name).toBe(jobName);
    expect(detail.handlerRef).toBe('host:wait');
    expect(detail.status).toBe('disabled');
    expect(JSON.parse(detail.params).seconds).toBe(1);

    await jobPage.openEditSearchedJob();
    await jobPage.fillCommonFields({
      description: 'handler job after update',
    });
    await jobPage.fillHandlerParam('等待秒数', '2');
    await jobPage.save();

    const updatedDetail = await getJob(api, jobId);
    expect(updatedDetail.description).toBe('handler job after update');
    expect(JSON.parse(updatedDetail.params).seconds).toBe(2);

    await jobPage.deleteSearchedJob();
    jobId = 0;

    await jobPage.fillSearchKeyword(jobName);
    await jobPage.clickSearch();
    await expect(await jobPage.hasJob(jobName)).toBe(false);
  });
});
