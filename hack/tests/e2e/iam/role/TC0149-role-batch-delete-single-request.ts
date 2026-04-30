import type { APIRequestContext, Request } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import { RolePage } from "../../../pages/RolePage";
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

test.describe("TC-149 Role batch delete single request", () => {
  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  test("TC-149a: selected roles are deleted through one DELETE /role request", async ({
    adminPage,
  }) => {
    const suffix = Date.now();
    const roles = [
      { key: `e2e_batch_role_a_${suffix}`, name: `E2E Batch Role A ${suffix}` },
      { key: `e2e_batch_role_b_${suffix}`, name: `E2E Batch Role B ${suffix}` },
    ];
    const createdIds: number[] = [];
    for (const role of roles) {
      const created = await expectSuccess<CreatedID>(
        await adminApi.post("role", {
          data: {
            dataScope: 1,
            key: role.key,
            name: role.name,
            sort: 999,
            status: 1,
          },
        }),
      );
      createdIds.push(created.id);
    }

    try {
      const rolePage = new RolePage(adminPage);
      const deleteRequests: Request[] = [];
      const requestListener = (request: Request) => {
        const url = new URL(request.url());
        if (
          request.method() === "DELETE" &&
          url.pathname.endsWith("/api/v1/role")
        ) {
          deleteRequests.push(request);
        }
      };
      adminPage.on("request", requestListener);

      await rolePage.goto();
      await rolePage.searchRole(String(suffix));
      await rolePage.selectVisibleRoleRows(roles.map((role) => role.name));

      const responsePromise = adminPage.waitForResponse((response) => {
        const url = new URL(response.url());
        return (
          response.request().method() === "DELETE" &&
          url.pathname.endsWith("/api/v1/role") &&
          response.status() === 200
        );
      });

      await rolePage.confirmSelectedRoleBatchDelete();
      await responsePromise;
      adminPage.off("request", requestListener);

      expect(deleteRequests).toHaveLength(1);
      expect(getRequestedIds(deleteRequests[0]!)).toEqual(
        [...createdIds].sort(),
      );

      const result = await expectSuccess<ListResult>(
        await adminApi.get(
          `role?page=1&size=100&key=${encodeURIComponent(String(suffix))}`,
        ),
      );
      expect(result.list.map((item) => item.id)).not.toEqual(
        expect.arrayContaining(createdIds),
      );
    } finally {
      if (createdIds.length > 0) {
        await adminApi
          .delete(`role?${buildRepeatedIdsQuery(createdIds)}`)
          .catch(() => {});
      }
    }
  });
});
