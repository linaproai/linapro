import type { APIRequestContext } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import {
  buildShellJobPayload,
  createAdminApiContext,
  createApiContext,
  createJob,
  createRole,
  createUser,
  deleteJob,
  deleteRole,
  deleteUser,
  expectBusinessError,
  getConfigByKey,
  getDefaultGroup,
  getMenuIdsByPermsWithAncestors,
  listJobs,
  setCronShellEnabled,
  updateConfigValue,
} from "../../../support/api/job";

const password = "test123456";

test.describe("TC-172 定时任务数据权限", () => {
  let adminApi: APIRequestContext;
  let limitedApi: APIRequestContext;
  let roleID = 0;
  let userID = 0;
  let visibleJobID = 0;
  let hiddenJobID = 0;
  let visibleJobName = "";
  let hiddenJobName = "";
  let originalShellSwitch: { id: number; value: string } | null = null;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    originalShellSwitch = await getConfigByKey(adminApi, "cron.shell.enabled");
    await setCronShellEnabled(adminApi, true);
    await expect
      .poll(async () => (await getConfigByKey(adminApi, "cron.shell.enabled")).value)
      .toBe("true");

    const suffix = Date.now().toString();
    const menuIds = await getMenuIdsByPermsWithAncestors(adminApi, [
      "system:job:list",
      "system:job:add",
      "system:job:trigger",
      "system:job:shell",
      "system:jobgroup:list",
    ]);
    roleID = (
      await createRole(adminApi, {
        name: `JobScope${suffix.slice(-10)}`,
        key: `e2e_job_self_${suffix}`,
        menuIds,
        dataScope: 3,
        sort: 970,
      })
    ).id;
    const username = `e2e_job_scope_${suffix}`;
    userID = (
      await createUser(adminApi, {
        username,
        password,
        nickname: "E2E Job Scope",
        roleIds: [roleID],
      })
    ).id;
    limitedApi = await createApiContext(username, password);

    const defaultGroup = await getDefaultGroup(adminApi);
    hiddenJobName = `e2e_hidden_job_${suffix}`;
    visibleJobName = `e2e_visible_job_${suffix}`;
    hiddenJobID = (
      await createJob(
        adminApi,
        buildShellJobPayload({
          groupId: defaultGroup.id,
          name: hiddenJobName,
        }),
      )
    ).id;
    visibleJobID = (
      await createJob(
        limitedApi,
        buildShellJobPayload({
          groupId: defaultGroup.id,
          name: visibleJobName,
        }),
      )
    ).id;
  });

  test.afterAll(async () => {
    for (const jobID of [visibleJobID, hiddenJobID]) {
      if (jobID > 0) {
        await deleteJob(adminApi, jobID).catch(() => {});
      }
    }
    await limitedApi?.post("auth/logout").catch(() => {});
    await limitedApi?.dispose();
    if (userID > 0) {
      await deleteUser(adminApi, userID).catch(() => {});
    }
    if (roleID > 0) {
      await deleteRole(adminApi, roleID).catch(() => {});
    }
    if (originalShellSwitch) {
      await updateConfigValue(
        adminApi,
        originalShellSwitch.id,
        originalShellSwitch.value,
      );
    }
    await adminApi?.dispose();
  });

  test("TC-172a~c: 用户创建任务过滤、内置任务可见、范围外触发拒绝", async () => {
    const scopedJobs = await listJobs(limitedApi, "e2e_");
    const scopedNames = scopedJobs.list.map((item) => item.name);
    expect(scopedNames).toContain(visibleJobName);
    expect(scopedNames).not.toContain(hiddenJobName);

    const allVisibleJobs = await listJobs(limitedApi);
    expect(
      allVisibleJobs.list.some((item) => item.isBuiltin === 1),
      "self-scope user should still see built-in job projection",
    ).toBeTruthy();

    await expectBusinessError(await limitedApi.get(`job/${hiddenJobID}`));
    await expectBusinessError(await limitedApi.post(`job/${hiddenJobID}/trigger`));
  });
});
