import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';

import {
  buildShellJobPayload,
  cancelLog,
  clearLogs,
  createAdminApiContext,
  createApiContext,
  createJob,
  createRole,
  createUser,
  deleteRole,
  deleteUser,
  expectBusinessError,
  getConfigByKey,
  getDefaultGroup,
  getLog,
  getMenuIdsByPerms,
  setCronShellEnabled,
  triggerJob,
  updateConfigValue,
} from '../../../support/api/job';

test.describe('TC-96 Shell 终止权限校验', () => {
  const jobName = `e2e_shell_cancel_permission_${Date.now()}`;
  const roleSuffix = Date.now().toString().slice(-8);
  const limitedRoleName = `ejscp_role_${roleSuffix}`;
  const limitedRoleKey = `ejscp_${roleSuffix}`;
  const limitedUsername = `e2e_shell_cancel_user_${Date.now()}`;
  const limitedPassword = 'test123456';

  let adminApi: APIRequestContext;
  let limitedApi: APIRequestContext | null = null;
  let roleId = 0;
  let userId = 0;
  let jobId = 0;
  let originalShellSwitch: { id: number; value: string } | null = null;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    originalShellSwitch = await getConfigByKey(adminApi, 'cron.shell.enabled');
    await setCronShellEnabled(adminApi, true);
    await expect
      .poll(async () => (await getConfigByKey(adminApi, 'cron.shell.enabled')).value)
      .toBe('true');

    const menuIds = await getMenuIdsByPerms(adminApi, ['system:joblog:cancel']);
    const createdRole = await createRole(adminApi, {
      name: limitedRoleName,
      key: limitedRoleKey,
      menuIds,
      remark: 'E2E shell cancel permission coverage',
    });
    roleId = createdRole.id;

    const createdUser = await createUser(adminApi, {
      username: limitedUsername,
      password: limitedPassword,
      nickname: 'E2E Shell Cancel User',
      roleIds: [roleId],
    });
    userId = createdUser.id;

    limitedApi = await createApiContext(limitedUsername, limitedPassword);
  });

  test.afterAll(async () => {
    if (jobId) {
      await clearLogs(adminApi, jobId);
      await adminApi.delete(`job/${jobId}`);
    }
    if (limitedApi) {
      await limitedApi.dispose();
    }
    if (userId) {
      await deleteUser(adminApi, userId);
    }
    if (roleId) {
      await deleteRole(adminApi, roleId);
    }
    if (originalShellSwitch) {
      await updateConfigValue(adminApi, originalShellSwitch.id, originalShellSwitch.value);
    }
    await adminApi.dispose();
  });

  test('TC-96a~d: 缺少 system:job:shell 时禁止终止运行中的 Shell 实例', async () => {
    if (!limitedApi) {
      throw new Error('limitedApi should be initialized in beforeAll');
    }
    const limitedUserApi = limitedApi;

    const defaultGroup = await getDefaultGroup(adminApi);
    const created = await createJob(adminApi, buildShellJobPayload({
      groupId: defaultGroup.id,
      name: jobName,
      shellCmd: 'sleep 30',
      timeoutSeconds: 60,
    }));
    jobId = created.id;

    const triggered = await triggerJob(adminApi, jobId);
    expect(triggered.logId).toBeGreaterThan(0);

    await expect
      .poll(async () => (await getLog(adminApi, triggered.logId)).status, {
        timeout: 5000,
        message: '权限校验前 Shell 任务应先进入 running 状态',
      })
      .toBe('running');

    await expectBusinessError(
      await limitedUserApi.post(`job/log/${triggered.logId}/cancel`),
      'system:job:shell',
    );

    await expect
      .poll(async () => (await getLog(adminApi, triggered.logId)).status, {
        timeout: 3000,
        message: '未授权取消请求不应改变运行中实例状态',
      })
      .toBe('running');

    await cancelLog(adminApi, triggered.logId);
    await expect
      .poll(async () => (await getLog(adminApi, triggered.logId)).status, {
        timeout: 10000,
        message: '管理员终止后日志状态应变为 cancelled',
      })
      .toBe('cancelled');
  });
});
