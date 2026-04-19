import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';

import {
  buildShellJobPayload,
  clearLogs,
  createAdminApiContext,
  createJob,
  getConfigByKey,
  getDefaultGroup,
  getLog,
  setCronShellEnabled,
  triggerJob,
  updateConfigValue,
} from './helpers';

test.describe('TC-94 Shell 输出截断可查看', () => {
  const jobName = `e2e_shell_output_${Date.now()}`;
  let api: APIRequestContext;
  let jobId = 0;
  let originalShellSwitch: { id: number; value: string } | null = null;

  test.beforeAll(async () => {
    api = await createAdminApiContext();
    originalShellSwitch = await getConfigByKey(api, 'cron.shell.enabled');
    await setCronShellEnabled(api, true);
    await expect
      .poll(async () => (await getConfigByKey(api, 'cron.shell.enabled')).value)
      .toBe('true');
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

  test('TC-94a~d: 超长 stdout/stderr 会被截断并保存在结果日志中', async () => {
    const defaultGroup = await getDefaultGroup(api);
    const created = await createJob(api, buildShellJobPayload({
      groupId: defaultGroup.id,
      name: jobName,
      shellCmd: "head -c 70000 /dev/zero | tr '\\000' 'a'; head -c 70000 /dev/zero | tr '\\000' 'b' 1>&2",
      timeoutSeconds: 30,
    }));
    jobId = created.id;

    const triggered = await triggerJob(api, jobId);
    expect(triggered.logId).toBeGreaterThan(0);

    await expect
      .poll(async () => (await getLog(api, triggered.logId)).status, {
        timeout: 10000,
        message: '超长输出脚本执行后应写入成功日志',
      })
      .toBe('success');

    const logDetail = await getLog(api, triggered.logId);
    const result = JSON.parse(logDetail.resultJson ?? '{}') as {
      stdout?: string;
      stderr?: string;
      exitCode?: number;
    };

    expect(result.exitCode).toBe(0);
    expect(result.stdout ?? '').toContain('[truncated]');
    expect(result.stderr ?? '').toContain('[truncated]');
    expect((result.stdout ?? '').length).toBeGreaterThan(1024);
    expect((result.stderr ?? '').length).toBeGreaterThan(1024);
  });
});
