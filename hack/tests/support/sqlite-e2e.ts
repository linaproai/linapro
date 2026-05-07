import type { APIResponse } from "@playwright/test";

import { expect, test } from "@playwright/test";

import { readFileSync } from "node:fs";
import path from "node:path";

const sqliteModeEnv = "LINAPRO_E2E_SQLITE_MODE";
const sqliteBackendLogEnv = "LINAPRO_E2E_SQLITE_BACKEND_LOG";

export const sqliteSourcePluginId = "plugin-demo-source";

export function requireSQLiteE2E() {
  test.skip(
    process.env[sqliteModeEnv] !== "1",
    "SQLite E2E cases run only through scripts/run-sqlite-suite.mjs, which writes database.default.link into the backend config file before startup.",
  );
}

export function sqliteBackendLogPath() {
  return (
    process.env[sqliteBackendLogEnv] ??
    path.resolve(process.cwd(), "../..", "temp", "lina-core.log")
  );
}

export async function expectSQLiteStartupWarnings() {
  const logPath = sqliteBackendLogPath();
  const requiredMessages = [
    "SQLite mode is active",
    "SQLite mode only supports single-node deployment",
    "do not use SQLite mode in production",
    "Switch database.default.link back to a MySQL link",
  ];
  await expect
    .poll(
      () => {
        try {
          const content = readFileSync(logPath, "utf8");
          return requiredMessages.filter((message) => content.includes(message))
            .length;
        } catch {
          return 0;
        }
      },
      {
        message: `expected SQLite startup warnings in ${logPath}`,
        timeout: 15_000,
      },
    )
    .toBe(requiredMessages.length);
}

export async function expectApiSuccess<T>(
  response: APIResponse,
  message: string,
): Promise<T> {
  expect(response.ok(), `${message}, status=${response.status()}`).toBeTruthy();
  const payload = (await response.json()) as {
    code?: number;
    data?: T;
    message?: string;
  };
  expect(
    payload.code,
    `${message}, business code=${payload.code}, message=${payload.message ?? ""}`,
  ).toBe(0);
  return payload.data as T;
}
