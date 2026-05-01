import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { JobPage } from '../../../pages/JobPage';
import {
  createAdminApiContext,
  disablePlugin,
  enablePlugin,
  expectBusinessError,
  getJob,
  getPlugin,
  installPlugin,
  listJobs,
  syncPlugins,
  triggerJob,
  uninstallPlugin,
} from '../../../support/api/job';

test.describe('TC-158 Built-in job execution boundary', () => {
  const pluginID = 'plugin-demo-source';
  const pluginJobName = '源码插件回显巡检';
  const pluginHandlerRef = `plugin:${pluginID}/cron:source-plugin-echo-inspection`;

  let api: APIRequestContext;
  let hostJobId = 0;
  let hostJobName = '';
  let pluginJobId = 0;
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

    const allJobs = await listJobs(api);
    const hostJob = allJobs.list.find(
      (item) => item.handlerRef === 'host:cleanup-job-logs' && item.isBuiltin === 1,
    );
    expect(hostJob).toBeTruthy();
    hostJobId = hostJob!.id;
    hostJobName = hostJob!.name;

    await expect
      .poll(
        async () => {
          const result = await listJobs(api, pluginJobName);
          const pluginJob = result.list.find(
            (item) => item.handlerRef === pluginHandlerRef && item.isBuiltin === 1,
          );
          pluginJobId = pluginJob?.id ?? 0;
          return pluginJob?.status ?? '';
        },
        {
          timeout: 10000,
          message: 'plugin built-in job should be enabled before boundary checks',
        },
      )
      .toBe('enabled');
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

  test('TC-158a~c: host built-ins stay read-only but manually triggerable', async ({
    adminPage,
  }) => {
    let triggerCalls = 0;
    await adminPage.route(`**/api/v1/job/${hostJobId}/trigger`, async (route) => {
      triggerCalls += 1;
      await route.fulfill({
        body: JSON.stringify({
          code: 0,
          data: { logId: 155_001 },
          message: 'OK',
        }),
        contentType: 'application/json',
        status: 200,
      });
    });

    const jobPage = new JobPage(adminPage);
    await jobPage.goto();
    await jobPage.fillSearchKeyword(hostJobName);
    await jobPage.clickSearch();

    await expect.poll(() => jobPage.hasAction('job-edit-')).toBe(true);
    await expect.poll(() => jobPage.hasAction('job-more-')).toBe(false);
    await expect.poll(() => jobPage.isActionDisabled('job-trigger-')).toBe(false);

    await jobPage.openTriggerConfirmForSearchedJob();
    await jobPage.confirmTriggerConfirm();
    await expect.poll(() => triggerCalls).toBe(1);
  });

  test('TC-158d~f: paused plugin built-ins are visible but not triggerable', async ({
    adminPage,
  }) => {
    await disablePlugin(api, pluginID);

    await expect
      .poll(
        async () => {
          const current = await getJob(api, pluginJobId);
          return `${current.status}:${current.stopReason}`;
        },
        {
          timeout: 10000,
          message: 'plugin built-in job should be projected as unavailable after disable',
        },
      )
      .toBe('paused_by_plugin:plugin_unavailable');

    const jobPage = new JobPage(adminPage);
    await jobPage.goto();
    await jobPage.fillSearchKeyword(pluginJobName);
    await jobPage.clickSearch();

    await expect.poll(() => jobPage.hasJob(pluginJobName)).toBe(true);
    await expect.poll(() => jobPage.isActionDisabled('job-trigger-')).toBe(true);
    await expect.poll(() => jobPage.hasAction('job-enable-')).toBe(false);

    await expectBusinessError(
      await api.post(`job/${pluginJobId}/trigger`),
      '插件处理器当前不可用',
    );

    await enablePlugin(api, pluginID);
    await expect
      .poll(
        async () => {
          const current = await getJob(api, pluginJobId);
          return current.status;
        },
        {
          timeout: 10000,
          message: 'plugin built-in job should recover after enable',
        },
      )
      .toBe('enabled');

    const triggered = await triggerJob(api, pluginJobId);
    expect(triggered.logId).toBeGreaterThan(0);
  });
});
