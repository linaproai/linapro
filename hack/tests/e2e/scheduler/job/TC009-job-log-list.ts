import type { APIRequestContext } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import { JobLogPage } from "../../../pages/JobLogPage";

import {
  buildShellJobPayload,
  clearLogs,
  createAdminApiContext,
  createJob,
  getConfigByKey,
  restoreCronShellEnabled,
  getDefaultGroup,
  getLog,
  listLogs,
  setCronShellEnabled,
  triggerJob,
} from "../../../support/api/job";

test.describe("TC-9 执行日志查询与删除", () => {
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
      await restoreCronShellEnabled(api, originalShellSwitch);
    }
    await api.dispose();
  });

  test("TC-9a~e: 日志支持列表筛选、详情查看，并通过范围删除弹窗发起删除", async ({
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
    await expect(await jobLogPage.detailContains(/success|成功/i)).toBe(true);

    await expect(adminPage.getByTestId("job-log-clear")).toHaveCount(0);
    await expect(adminPage.locator(".vxe-checkbox--icon")).toHaveCount(0);

    const deleteDialog = await jobLogPage.openDeleteDialog();
    await expect(deleteDialog).toContainText("删除执行日志");
    await expect(deleteDialog).toContainText("请选择执行日志删除方式");
    await expect(deleteDialog).toContainText("删除所有执行日志");
    await expect(deleteDialog.locator(".ant-picker-range")).toBeVisible();

    await deleteDialog
      .getByRole("button", { name: /确\s*(认|定)/ })
      .click();
    await expect(
      adminPage.getByText("请选择完整的执行日志日期范围"),
    ).toBeVisible();

    await deleteDialog.getByText("删除所有执行日志").click();
    await expect(
      deleteDialog.locator(".ant-picker-range input").first(),
    ).toBeDisabled();

    let cleanRequestUrl = "";
    await adminPage.route("**/job/log**", async (route) => {
      if (route.request().method() !== "DELETE") {
        await route.continue();
        return;
      }
      cleanRequestUrl = route.request().url();
      await route.fulfill({
        body: JSON.stringify({ code: 0, data: { deleted: 0 }, message: "OK" }),
        contentType: "application/json",
        status: 200,
      });
    });
    await jobLogPage.confirmDeleteDialog(deleteDialog);
    await expect.poll(() => cleanRequestUrl).not.toBe("");
    const deleteUrl = new URL(cleanRequestUrl);
    expect(deleteUrl.searchParams.has("beginTime")).toBe(false);
    expect(deleteUrl.searchParams.has("endTime")).toBe(false);

    const remained = await listLogs(api, jobId);
    expect(remained.total).toBeGreaterThanOrEqual(2);
  });
});
