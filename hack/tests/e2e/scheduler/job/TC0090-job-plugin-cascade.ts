import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { JobPage } from '../../../pages/JobPage';

import {
  createAdminApiContext,
  disablePlugin,
  enablePlugin,
  expectBusinessError,
  getJob,
  getLog,
  getPlugin,
  installPlugin,
  listHandlers,
  listJobs,
  triggerJob,
  uninstallPlugin,
  syncPlugins,
} from '../../../support/api/job';

test.describe('TC-90 插件内置任务生命周期级联', () => {
  const pluginID = 'plugin-demo-source';
  const jobName = '源码插件回显巡检';
  const cronHandlerName = 'source-plugin-echo-inspection';
  const handlerRef = `plugin:${pluginID}/cron:${cronHandlerName}`;
  const removedGenericHandlerRef = `plugin:${pluginID}/echo`;

  let api: APIRequestContext;
  let jobId = 0;
  let originalInstalled = 0;
  let originalEnabled = 0;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
    await syncPlugins(api);

    const plugin = await getPlugin(api, pluginID);
    originalInstalled = plugin.installed;
    originalEnabled = plugin.enabled;

    if (plugin.installed !== 1) {
      await installPlugin(api, pluginID);
    }
    if (plugin.enabled !== 1) {
      await enablePlugin(api, pluginID);
    }

    await expect
      .poll(
        async () => {
          const handlers = await listHandlers(api);
          const hasCronHandler = handlers.list.some((item) => item.ref === handlerRef);
          const hasRemovedGenericHandler = handlers.list.some(
            (item) => item.ref === removedGenericHandlerRef,
          );
          return `${hasCronHandler}:${hasRemovedGenericHandler}`;
        },
        {
          timeout: 10000,
          message: '启用源码插件后应只注册其内置 cron 处理器，不再暴露 plugin.jobs 处理器',
        },
      )
      .toBe('true:false');

    await expect
      .poll(
        async () => {
          const jobs = await listJobs(api, jobName);
          const builtinJob = jobs.list.find(
            (item) => item.name === jobName && item.isBuiltin === 1,
          );
          jobId = builtinJob?.id ?? 0;
          return builtinJob
            ? `${builtinJob.status}:${builtinJob.handlerRef}:${builtinJob.isBuiltin}`
            : '';
        },
        {
          timeout: 10000,
          message: '插件内置定时任务应投影到 sys_job 并恢复为 enabled',
        },
      )
      .toBe(`enabled:${handlerRef}:1`);
  });

  test.afterAll(async () => {
    if (!api) {
      return;
    }

    if (originalInstalled !== 1) {
      await uninstallPlugin(api, pluginID);
    } else if (originalEnabled !== 1) {
      await disablePlugin(api, pluginID);
    } else {
      await enablePlugin(api, pluginID);
    }

    await api.dispose();
  });

  test('TC-90a~f: 插件禁用时内置任务应暂停，恢复后应自动启用并支持手动执行', async ({
    adminPage,
  }) => {
    const originalDetail = await getJob(api, jobId);
    expect(originalDetail.isBuiltin).toBe(1);
    expect(originalDetail.taskType).toBe('handler');
    expect(originalDetail.handlerRef).toBe(handlerRef);

    const jobPage = new JobPage(adminPage);
    await jobPage.goto();
    await jobPage.fillSearchKeyword(jobName);
    await jobPage.clickSearch();

    const initialRowText = await jobPage.getJobRowText(jobName);
    expect(initialRowText).toContain('插件内置');

    await jobPage.openEditSearchedJob();
    const detailCard = adminPage.getByTestId('job-builtin-detail-card');
    await expect(detailCard).toBeVisible();
    await expect(detailCard.getByText('插件内置', { exact: true })).toBeVisible();
    await expect(detailCard.getByText(pluginID, { exact: true })).toBeVisible();
    await expect(
      adminPage.getByRole('button', { name: /确\s*认/ }),
    ).toHaveCount(0);
    await jobPage.closeDialog();

    await disablePlugin(api, pluginID);

    await expect
      .poll(
        async () => {
          const current = await getJob(api, jobId);
          return `${current.status}:${current.stopReason}`;
        },
        {
          timeout: 10000,
          message: '插件禁用后，内置任务应级联暂停并写入 plugin_unavailable',
        },
      )
      .toBe('paused_by_plugin:plugin_unavailable');

    const handlersAfterDisable = await listHandlers(api);
    expect(
      handlersAfterDisable.list.some((item) => item.ref === handlerRef),
    ).toBeFalsy();
    expect(
      handlersAfterDisable.list.some(
        (item) => item.ref === removedGenericHandlerRef,
      ),
    ).toBeFalsy();

    await jobPage.goto();
    await jobPage.fillSearchKeyword(jobName);
    await jobPage.clickSearch();
    await expect(await jobPage.hasJob(jobName)).toBe(true);
    await expect(await jobPage.isPausedByPluginVisible()).toBe(true);
    await jobPage.hoverPausedStatusTag();
    await expect(
      await jobPage.isTooltipVisible('该任务依赖插件提供的处理器'),
    ).toBe(true);
    await expect(await jobPage.hasAction('job-enable-')).toBe(false);
    await expect(await jobPage.isActionDisabled('job-trigger-')).toBe(true);

    const triggerWhilePaused = await api.post(`job/${jobId}/trigger`);
    await expectBusinessError(triggerWhilePaused, '插件处理器当前不可用');

    const enableWhileBuiltinReadonly = await api.put(`job/${jobId}/status`, {
      data: { status: 'enabled' },
    });
    await expectBusinessError(enableWhileBuiltinReadonly, '源码注册任务不允许修改状态');

    await enablePlugin(api, pluginID);

    await expect
      .poll(
        async () => {
          const current = await getJob(api, jobId);
          return current.status;
        },
        {
          timeout: 10000,
          message: '插件重新启用后，内置任务应自动恢复为 enabled',
        },
      )
      .toBe('enabled');

    const handlersAfterEnable = await listHandlers(api);
    expect(
      handlersAfterEnable.list.some((item) => item.ref === handlerRef),
    ).toBeTruthy();
    expect(
      handlersAfterEnable.list.some(
        (item) => item.ref === removedGenericHandlerRef,
      ),
    ).toBeFalsy();

    const triggered = await triggerJob(api, jobId);
    expect(triggered.logId).toBeGreaterThan(0);

    await expect
      .poll(
        async () => {
          const logDetail = await getLog(api, triggered.logId);
          return logDetail.status;
        },
        {
          timeout: 10000,
          message: '插件内置任务恢复后应可成功手动执行',
        },
      )
      .toBe('success');

    const successLog = await getLog(api, triggered.logId);
    expect(successLog.resultJson ?? '').toContain('"executed":true');
  });
});
