import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { JobPage } from '../../../pages/JobPage';
import {
  buildShellJobPayload,
  createAdminApiContext,
  createJob,
  deleteJob,
  getConfigByKey,
  getDefaultGroup,
  setCronShellEnabled,
  updateConfigValue,
} from '../../../support/api/job';

test.describe('TC-143 Scheduled job manual trigger confirmation', () => {
  const jobName = `e2e_job_trigger_confirm_${Date.now()}`;

  let api: APIRequestContext;
  let jobId = 0;
  let originalShellSwitch: { id: number; value: string } | null = null;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
    originalShellSwitch = await getConfigByKey(api, 'cron.shell.enabled');
    await setCronShellEnabled(api, true);

    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(
      api,
      buildShellJobPayload({
        groupId: defaultGroup.id,
        name: jobName,
        shellCmd: "printf 'manual trigger confirmation'",
        status: 'disabled',
      }),
    );
    jobId = created.id;
  });

  test.afterAll(async () => {
    if (jobId) {
      await deleteJob(api, jobId);
    }
    if (originalShellSwitch) {
      await updateConfigValue(
        api,
        originalShellSwitch.id,
        originalShellSwitch.value,
      );
    }
    await api.dispose();
  });

  test('TC-143a~c: Run Now opens confirmation, cancel skips API call, confirm triggers once', async ({
    adminPage,
    mainLayout,
  }) => {
    let triggerCalls = 0;
    await adminPage.route(`**/api/v1/job/${jobId}/trigger`, async (route) => {
      triggerCalls += 1;
      await route.fulfill({
        body: JSON.stringify({
          code: 0,
          data: { logId: 987_654 },
          message: 'OK',
        }),
        contentType: 'application/json',
        status: 200,
      });
    });

    await mainLayout.switchLanguage('English');
    const jobPage = new JobPage(adminPage);
    await jobPage.goto();
    await jobPage.fillSearchKeyword(jobName);
    await jobPage.clickSearch();

    const firstConfirm = await jobPage.openTriggerConfirmForSearchedJob();
    await expect(firstConfirm).toContainText('Run this job now?');
    await jobPage.cancelTriggerConfirm();
    expect(triggerCalls).toBe(0);

    const secondConfirm = await jobPage.openTriggerConfirmForSearchedJob();
    await expect(secondConfirm).toContainText('Run this job now?');
    await jobPage.confirmTriggerConfirm();

    await expect.poll(() => triggerCalls).toBe(1);
  });
});
