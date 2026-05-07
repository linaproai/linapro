import { existsSync, statSync } from "node:fs";
import path from "node:path";

import { test, expect } from "../../fixtures/auth";
import { createAdminApiContext } from "../../fixtures/plugin";
import {
  expectApiSuccess,
  requireSQLiteE2E,
} from "../../support/sqlite-e2e";

const sqliteDbPath = path.resolve(
  process.cwd(),
  "..",
  "..",
  "apps",
  "lina-core",
  "temp",
  "sqlite",
  "linapro.db",
);

test.describe("TC-166 SQLite mode rebuild and reseed", () => {
  requireSQLiteE2E();

  test("TC-166a: rebuilt SQLite database exists and seeded admin plus mock user are queryable", async () => {
    expect(existsSync(sqliteDbPath)).toBe(true);
    expect(statSync(sqliteDbPath).size).toBeGreaterThan(0);

    const api = await createAdminApiContext();
    try {
      const adminUsers = await expectApiSuccess<{ list: Array<{ username: string }> }>(
        await api.get("user", { params: { username: "admin" } }),
        "query admin after SQLite rebuild",
      );
      expect(adminUsers.list.some((item) => item.username === "admin")).toBe(
        true,
      );

      const mockUsers = await expectApiSuccess<{ list: Array<{ username: string }> }>(
        await api.get("user", { params: { username: "user001" } }),
        "query mock user after SQLite mock load",
      );
      expect(mockUsers.list.some((item) => item.username === "user001")).toBe(
        true,
      );
    } finally {
      await api.dispose();
    }
  });
});

