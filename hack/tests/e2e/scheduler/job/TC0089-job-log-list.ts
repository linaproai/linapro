import type { APIRequestContext } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import { JobLogPage } from "../../../pages/JobLogPage";

import {
  buildShellJobPayload,
  clearLogs,
  createAdminApiContext,
  createJob,
  getConfigByKey,
  getDefaultGroup,
  getLog,
  listLogs,
  setCronShellEnabled,
  triggerJob,
  updateConfigValue,
} from "../../../support/api/job";

test.describe("TC-89 执行日志查询与清理", () => {
  const jobName = `e2e_job_log_${Date.now()}`;
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
      await clearLogs(api, jobId);
      await api.delete(`job/${jobId}`);
    }
    if (originalShellSwitch) {
      await updateConfigValue(api, originalShellSwitch.id, originalShellSwitch.value);
    }
    await api.dispose();
  });

  test("TC-89a~e: 日志支持列表筛选、详情查看、批量删除，并可按任务维度清空", async ({
    adminPage,
  }) => {
    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(
      api,
      buildShellJobPayload({
        groupId: defaultGroup.id,
        name: jobName,
        status: "enabled",
      }),
    );
    jobId = created.id;

    const triggered = await triggerJob(api, jobId);
    expect(triggered.logId).toBeGreaterThan(0);
    const triggeredSecond = await triggerJob(api, jobId);
    expect(triggeredSecond.logId).toBeGreaterThan(0);

    await expect
      .poll(
        async () => {
          const [firstDetail, secondDetail] = await Promise.all([
            getLog(api, triggered.logId),
            getLog(api, triggeredSecond.logId),
          ]);
          return `${firstDetail.status}:${secondDetail.status}`;
        },
        {
          timeout: 10000,
          message: "两条日志详情应在触发后进入成功状态",
        },
      )
      .toBe("success:success");

    const logList = await listLogs(api, jobId);
    expect(logList.total).toBeGreaterThanOrEqual(2);
    expect(
      logList.list.some((item) => item.id === triggered.logId),
    ).toBeTruthy();
    expect(
      logList.list.some((item) => item.id === triggeredSecond.logId),
    ).toBeTruthy();

    const logDetail = await getLog(api, triggered.logId);
    expect(logDetail.jobId).toBe(jobId);
    expect(logDetail.jobName).toBe(jobName);
    expect(logDetail.trigger).toBe("manual");
    expect(logDetail.status).toBe("success");

    const jobLogPage = new JobLogPage(adminPage);
    await jobLogPage.goto();
    await jobLogPage.selectJob(jobName);
    await jobLogPage.clickSearch();
    await expect(await jobLogPage.getVisibleRowCount()).toBeGreaterThan(0);

    await jobLogPage.openFirstDetail();
    await expect(await jobLogPage.detailContains(jobName)).toBe(true);
    await expect(await jobLogPage.detailContains("manual")).toBe(true);
    await expect(await jobLogPage.detailContains("success")).toBe(true);

    await jobLogPage.selectFirstRow();
    await jobLogPage.deleteSelectedLogs();

    await expect
      .poll(
        async () => {
          const remained = await listLogs(api, jobId);
          return remained.total;
        },
        {
          timeout: 5000,
          message: "批量删除后应只剩余 1 条执行日志",
        },
      )
      .toBe(1);

    await jobLogPage.clearLogs();

    await expect
      .poll(
        async () => {
          const cleared = await listLogs(api, jobId);
          return cleared.total;
        },
        {
          timeout: 5000,
          message: "按任务清空日志后应看不到历史记录",
        },
      )
      .toBe(0);
  });
});
