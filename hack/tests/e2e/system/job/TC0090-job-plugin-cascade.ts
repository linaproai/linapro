import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { JobPage } from '../../../pages/JobPage';

import {
  buildHandlerJobPayload,
  clearLogs,
  createAdminApiContext,
  createJob,
  disablePlugin,
  enablePlugin,
  expectBusinessError,
  getDefaultGroup,
  getJob,
  getLog,
  getPlugin,
  installPlugin,
  listHandlers,
  syncPlugins,
  triggerJob,
  uninstallPlugin,
  updateJobStatus,
} from './helpers';

test.describe('TC-90 插件处理器生命周期级联', () => {
  const pluginID = 'plugin-demo-source';
  const handlerRef = `plugin:${pluginID}/echo`;
  const jobName = `e2e_plugin_handler_job_${Date.now()}`;

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
      .poll(async () => {
        const handlers = await listHandlers(api);
        return handlers.list.some((item) => item.ref === handlerRef);
      }, {
        timeout: 10000,
        message: '启用源码插件后应注册其定时任务处理器',
      })
      .toBeTruthy();
  });

  test.afterAll(async () => {
    if (jobId) {
      await clearLogs(api, jobId);
      await api.delete(`job/${jobId}`);
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

  test('TC-90a~d: 插件禁用时任务应暂停，重新启用后应自动恢复并可继续执行', async ({ adminPage }) => {
    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(api, buildHandlerJobPayload({
      groupId: defaultGroup.id,
      name: jobName,
      handlerRef,
      params: {
        message: 'plugin lifecycle echo',
      },
      status: 'disabled',
    }));
    jobId = created.id;

    await updateJobStatus(api, jobId, 'enabled');
    let detail = await getJob(api, jobId);
    expect(detail.status).toBe('enabled');
    expect(detail.handlerRef).toBe(handlerRef);

    await disablePlugin(api, pluginID);

    await expect
      .poll(async () => {
        const current = await getJob(api, jobId);
        return `${current.status}:${current.stopReason}`;
      }, {
        timeout: 10000,
        message: '插件禁用后，关联任务应级联暂停并写入 plugin_unavailable',
      })
      .toBe('paused_by_plugin:plugin_unavailable');

    const jobPage = new JobPage(adminPage);
    await jobPage.goto();
    await jobPage.fillSearchKeyword(jobName);
    await jobPage.clickSearch();
    await expect(await jobPage.hasJob(jobName)).toBe(true);
    await expect(await jobPage.isPausedByPluginVisible()).toBe(true);
    await expect(await jobPage.isActionDisabled('job-enable-')).toBe(true);
    await expect(await jobPage.isActionDisabled('job-trigger-')).toBe(true);

    const handlersAfterDisable = await listHandlers(api);
    expect(handlersAfterDisable.list.some((item) => item.ref === handlerRef)).toBeFalsy();

    const triggerResponse = await api.post(`job/${jobId}/trigger`);
    await expectBusinessError(triggerResponse, '插件处理器当前不可用');

    const enableWhileMissing = await api.put(`job/${jobId}/status`, {
      data: { status: 'enabled' },
    });
    await expectBusinessError(enableWhileMissing, '任务处理器不存在');

    await enablePlugin(api, pluginID);

    await expect
      .poll(async () => {
        const current = await getJob(api, jobId);
        return current.status;
      }, {
        timeout: 10000,
        message: '插件重新启用后，关联任务应自动恢复为 enabled',
      })
      .toBe('enabled');

    const handlersAfterEnable = await listHandlers(api);
    expect(handlersAfterEnable.list.some((item) => item.ref === handlerRef)).toBeTruthy();

    const triggered = await triggerJob(api, jobId);
    expect(triggered.logId).toBeGreaterThan(0);

    await expect
      .poll(async () => {
        const logDetail = await getLog(api, triggered.logId);
        return logDetail.status;
      }, {
        timeout: 10000,
        message: '插件处理器任务重新启用后应可成功执行',
      })
      .toBe('success');

    const successLog = await getLog(api, triggered.logId);
    expect(successLog.resultJson ?? '').toContain(pluginID);
    expect(successLog.resultJson ?? '').toContain('plugin lifecycle echo');
  });
});
