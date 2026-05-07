import { request } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";
import { config } from "../../fixtures/config";
import {
  expectApiSuccess,
  expectSQLiteStartupWarnings,
  requireSQLiteE2E,
} from "../../support/sqlite-e2e";

test.describe("TC-164 SQLite mode startup", () => {
  requireSQLiteE2E();

  test("TC-164a: SQLite startup warnings, single-node health, and admin login are available", async ({
    adminPage,
  }) => {
    await expectSQLiteStartupWarnings();

    const api = await request.newContext({
      baseURL: new URL("/api/v1/", config.baseURL).toString(),
    });
    try {
      const health = await expectApiSuccess<{ mode: string; status: string }>(
        await api.get("health"),
        "query health in SQLite mode",
      );
      expect(health.status).toBe("ok");
      expect(health.mode).toBe("single");
    } finally {
      await api.dispose();
    }

    await expect(adminPage).toHaveURL(/\/dashboard\/analytics/);
  });
});

