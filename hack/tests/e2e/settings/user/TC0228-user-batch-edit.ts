import type { APIRequestContext } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import { UserPage } from "../../../pages/UserPage";
import {
  createAdminApiContext,
  createUser,
  deleteUser,
  expectSuccess,
} from "../../../support/api/job";

type UserListItem = {
  id: number;
  status: number;
  username: string;
};

async function listUsers(api: APIRequestContext, username: string) {
  return expectSuccess<{ list: UserListItem[]; total: number }>(
    await api.get(
      `user?pageNum=1&pageSize=20&username=${encodeURIComponent(username)}`,
    ),
  );
}

test.describe("TC-228 用户管理批量编辑", () => {
  let adminApi: APIRequestContext;
  const userIds: number[] = [];
  const suffix = Date.now().toString();
  const usernames = [
    `e2e_batch_edit_a_${suffix}`,
    `e2e_batch_edit_b_${suffix}`,
  ];

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    for (const username of usernames) {
      const created = await createUser(adminApi, {
        username,
        password: "test123456",
        nickname: `E2E ${username}`,
      });
      userIds.push(created.id);
    }
  });

  test.afterAll(async () => {
    for (const userId of userIds) {
      await deleteUser(adminApi, userId).catch(() => {});
    }
    await adminApi?.dispose();
  });

  test("TC-228a: 选中用户后可批量禁用状态", async ({ adminPage }) => {
    const userPage = new UserPage(adminPage);
    await userPage.goto();

    await userPage.searchByUsernameKeyword(`e2e_batch_edit_`);
    for (const username of usernames) {
      await expect(userPage.getUserRow(username)).toBeVisible();
    }

    await userPage.selectVisibleUserRows(usernames);
    await userPage.expectToolbarPrimaryActionsDistinct();
    await userPage.batchUpdateSelectedStatus("停用");

    const result = await listUsers(adminApi, `e2e_batch_edit_`);
    const rows = result.list.filter((item) => usernames.includes(item.username));
    expect(rows).toHaveLength(2);
    expect(rows.every((item) => item.status === 0)).toBeTruthy();
  });
});
