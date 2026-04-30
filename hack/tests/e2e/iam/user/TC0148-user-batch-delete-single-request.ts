import type { APIRequestContext, Request } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import { UserPage } from "../../../pages/UserPage";
import { createAdminApiContext, expectSuccess } from "../../../support/api/job";

type CreatedID = {
  id: number;
};

type ListResult = {
  list: Array<{ id: number }>;
  total: number;
};

function getRequestedIds(request: Request) {
  return new URL(request.url())
    .searchParams.getAll("ids")
    .map(Number)
    .sort();
}

function buildRepeatedIdsQuery(ids: number[]) {
  const params = new URLSearchParams();
  for (const id of ids) {
    params.append("ids", String(id));
  }
  return params.toString();
}

let adminApi: APIRequestContext;

test.describe("TC-148 User batch delete single request", () => {
  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  test("TC-148a: selected users are deleted through one DELETE /user request", async ({
    adminPage,
  }) => {
    const suffix = Date.now();
    const users = [
      {
        nickname: "Batch Delete User A",
        username: `e2e_batch_user_a_${suffix}`,
      },
      {
        nickname: "Batch Delete User B",
        username: `e2e_batch_user_b_${suffix}`,
      },
    ];
    const createdIds: number[] = [];
    for (const user of users) {
      const created = await expectSuccess<CreatedID>(
        await adminApi.post("user", {
          data: {
            nickname: user.nickname,
            password: "test123456",
            status: 1,
            username: user.username,
          },
        }),
      );
      createdIds.push(created.id);
    }

    try {
      const userPage = new UserPage(adminPage);
      const deleteRequests: Request[] = [];
      const requestListener = (request: Request) => {
        const url = new URL(request.url());
        if (
          request.method() === "DELETE" &&
          url.pathname.endsWith("/api/v1/user")
        ) {
          deleteRequests.push(request);
        }
      };
      adminPage.on("request", requestListener);

      await userPage.goto();
      await userPage.searchByUsernameKeyword(String(suffix));
      await userPage.selectVisibleUserRows(users.map((user) => user.username));

      const responsePromise = adminPage.waitForResponse((response) => {
        const url = new URL(response.url());
        return (
          response.request().method() === "DELETE" &&
          url.pathname.endsWith("/api/v1/user") &&
          response.status() === 200
        );
      });

      await userPage.confirmSelectedUserBatchDelete();
      await responsePromise;
      adminPage.off("request", requestListener);

      expect(deleteRequests).toHaveLength(1);
      expect(getRequestedIds(deleteRequests[0]!)).toEqual(
        [...createdIds].sort(),
      );

      const result = await expectSuccess<ListResult>(
        await adminApi.get(
          `user?pageNum=1&pageSize=100&username=${encodeURIComponent(String(suffix))}`,
        ),
      );
      expect(result.list.map((item) => item.id)).not.toEqual(
        expect.arrayContaining(createdIds),
      );
    } finally {
      if (createdIds.length > 0) {
        await adminApi
          .delete(`user?${buildRepeatedIdsQuery(createdIds)}`)
          .catch(() => {});
      }
    }
  });
});
