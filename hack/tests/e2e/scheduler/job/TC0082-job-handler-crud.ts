import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { JobPage } from '../../../pages/JobPage';

import {
  createAdminApiContext,
  getConfigByKey,
  listJobs,
  restoreCronShellEnabled,
  setCronShellEnabled,
} from '../../../support/api/job';

test.describe('TC-82 源码注册任务可见且只读', () => {
  let api: APIRequestContext;
  let originalShellSwitch: { id: number; value: string } | null = null;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
    originalShellSwitch = await getConfigByKey(api, 'cron.shell.enabled');
    await setCronShellEnabled(api, true);
  });

  test.afterAll(async () => {
    if (originalShellSwitch) {
      await restoreCronShellEnabled(api, originalShellSwitch);
    }
    await api.dispose();
  });

  test('TC-82a~e: 宿主内置任务应显示在列表中，公共新增入口不再支持 Handler，源码任务仅可查看详情', async ({
    adminPage,
  }) => {
    const expectedBuiltinNames = ['任务日志清理', '在线会话清理', '服务监控采集'];

    for (const name of expectedBuiltinNames) {
      const result = await listJobs(api, name);
      expect(
        result.list.some((item) => item.name === name && item.isBuiltin === 1),
      ).toBeTruthy();
    }

    const jobPage = new JobPage(adminPage);
    await jobPage.goto();

    await jobPage.openCreate();
    await expect(
      adminPage.getByTestId('job-form-tab-handler'),
    ).toHaveCount(0);
    await expect(
      adminPage.getByLabel('任务处理器', { exact: true }),
    ).toHaveCount(0);
    await jobPage.closeDialog();

    await jobPage.fillSearchKeyword('任务日志清理');
    await jobPage.clickSearch();
    await expect(await jobPage.hasJob('任务日志清理')).toBe(true);

    await jobPage.openEditSearchedJob();
    const builtinDetailCard = adminPage.getByTestId('job-builtin-detail-card');
    await expect(builtinDetailCard).toBeVisible();
    await expect(
      builtinDetailCard.getByText('宿主内置', { exact: true }),
    ).toBeVisible();
    await expect(
      builtinDetailCard.getByText('host:cleanup-job-logs', { exact: true }),
    ).toBeVisible();
    await expect(
      adminPage.getByRole('button', { name: /确\s*认/ }),
    ).toHaveCount(0);
  });
});
