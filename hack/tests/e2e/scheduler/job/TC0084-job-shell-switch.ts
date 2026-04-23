import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import { JobPage } from '../../../pages/JobPage';

import {
  buildShellJobPayload,
  createAdminApiContext,
  expectBusinessError,
  getConfigByKey,
  setCronShellEnabled,
  updateConfigValue,
} from '../../../support/api/job';

test.describe('TC-84 Shell 全局开关关闭时拒绝写入', () => {
  let api: APIRequestContext;
  let originalShellSwitch: { id: number; value: string } | null = null;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
    originalShellSwitch = await getConfigByKey(api, 'cron.shell.enabled');
  });

  test.afterAll(async () => {
    if (originalShellSwitch) {
      await updateConfigValue(api, originalShellSwitch.id, originalShellSwitch.value);
    }
    await api.dispose();
  });

  test('TC-84a~b: 当 shell 开关关闭时，前端隐藏 Shell 类型选项且后端拒绝创建 Shell 任务', async ({ adminPage }) => {
    await setCronShellEnabled(api, false);
    await expect
      .poll(async () => (await getConfigByKey(api, 'cron.shell.enabled')).value)
      .toBe('false');

    const jobPage = new JobPage(adminPage);
    await jobPage.goto();
    await expect(adminPage.getByTestId('job-add')).toHaveCount(0);

    await expectBusinessError(
      await api.post('job', {
        data: buildShellJobPayload({
          name: `e2e_shell_switch_${Date.now()}`,
        }),
      }),
      '当前环境未启用 Shell 任务',
    );
  });
});
