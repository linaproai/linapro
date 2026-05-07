# E2E Test Suite

This directory contains the Playwright `E2E` suite for the default LinaPro management workbench and host-plugin integration flows.

## Directory Layout

```text
hack/tests/
  config/        execution manifest and suite governance data
  debug/         ad-hoc debugging scripts kept out of the E2E tree
  e2e/           TC test cases only
  fixtures/      shared Playwright fixtures and auth helpers
  pages/         page objects
  scripts/       suite runner and validation scripts
  support/       shared helpers such as API utilities and UI wait helpers
  temp/          runtime-only artifacts such as generated storage state
```

The `e2e/` tree is organized by stable capability boundaries instead of the legacy `system/` bucket:

- `auth/`, `dashboard/`, `about/`
- `iam/`
- `settings/`
- `org/`
- `content/`
- `monitor/`
- `scheduler/`
- `extension/`

## Naming Rules

- Test files must use `TC{NNNN}-{brief-name}.ts`.
- `TC` identifiers are globally unique across the whole suite.
- Only real `TC` files may live under `hack/tests/e2e/`.
- Shared helpers must live in `fixtures/`, `support/`, `scripts/`, or `debug/`.

## Execution Entrypoints

| Command | Purpose |
| --- | --- |
| `pnpm test` | Run the full layered suite. |
| `pnpm test:full` | Run the full layered suite explicitly. |
| `pnpm test:smoke` | Run the curated high-value smoke pack. |
| `pnpm test:module -- <scope>` | Run a module scope from the execution manifest. |
| `pnpm test:sqlite` | Prepare a SQLite config-file database, restart the app, and run SQLite-only E2E cases. |
| `pnpm test:sqlite:e2e-smoke` | Run only the SQLite browser startup/login E2E case. |
| `pnpm test:validate` | Validate `TC` uniqueness, directory ownership, and manifest references. |
| `pnpm report` | Open the Playwright HTML report. |

Example module scopes:

- `iam:user`
- `settings:config`
- `monitor:operlog`
- `scheduler:job`
- `extension:plugin`
- `dialect`

`pnpm test:sqlite` is the dedicated full SQLite channel. The script backs up
`apps/lina-core/manifest/config/config.yaml`, writes
`database.default.link=sqlite::@file(./temp/sqlite/linapro.db)` into that config
file, runs `make init confirm=init rebuild=true`, `make mock confirm=mock`,
starts `make dev`, runs `TC0164` to `TC0166`, then restores the original config.
The backend still derives the active database dialect only from its config file.
`pnpm test:sqlite:e2e-smoke` uses the same preparation path but only runs
`TC0164`. Main CI does not install frontend dependencies or Playwright for
SQLite; it calls `hack/tests/scripts/run-sqlite-smoke.sh`, which starts only the
backend and checks SQLite startup warnings, health mode, and admin login.

## Execution Model

The suite uses `config/execution-manifest.json` as the single source of truth for:

- legacy-to-target directory mapping
- module scopes
- smoke file selection
- serial execution boundaries for shared-state scenarios
- serial isolation categories and documented parallel exceptions

`pnpm test`, `pnpm test:full`, `pnpm test:smoke`, and `pnpm test:module` all run through `scripts/run-suite.mjs`.
The runner splits the selected files into a parallel pool and a serial pool so global-state heavy scenarios still execute safely.
Every run prints the selected file count, parallel file count, serial file count, parallel worker count, and the isolation categories represented in the serial pool.

## Isolation Categories

Use `serialIsolation` in `config/execution-manifest.json` when a test file or directory mutates or depends on shared state that can affect other files.

| Category | Use for |
| --- | --- |
| `authSession` | Tests that verify shared authenticated browser state, such as logout. |
| `pluginLifecycle` | Plugin sync, install, enable, disable, uninstall, upload, or upgrade flows. |
| `runtimeI18nCache` | Runtime language bundle versions, ETag checks, and language-cache revalidation. |
| `systemConfig` | System parameter and public frontend configuration mutations. |
| `dictionaryData` | Dictionary type or dictionary data create, import, edit, delete, and cascade scenarios. |
| `permissionMatrix` | Menu, role, button permission, and plugin-generated permission matrix mutations. |
| `sharedDatabaseSeed` | Tests that depend on shared seed or mock data loaded by fixtures. |
| `filesystemArtifact` | Plugin package, runtime plugin, or other shared runtime artifact mutations. |

Keep read-only tests in the parallel pool when they use fixture-owned prerequisites and unique local data.
If a high-risk pattern is intentionally parallel safe, add a `parallelIsolationAllowlist` entry with the file, category, and reason.
The validator rejects missing categories and allowlist entries without reasons.

## Fixture-Owned Prerequisites

Test files must be independently runnable.
Source plugin prerequisites should go through `fixtures/plugin.ts`, which syncs source plugins, installs or enables them when needed, refreshes the frontend projection, and loads matching plugin mock SQL when present.
Tests that create users, departments, posts, notices, files, plugins, import rows, or export artifacts should use unique names or stable test prefixes and clean up their own data in `finally`, `afterEach`, or `afterAll`.

## Cache Revalidation

Cache and ETag tests should validate protocol semantics instead of assuming the resource version stays unchanged during a full regression.
A conditional request must prove that the request carries the expected precondition.
It may accept `304 Not Modified` when the ETag still matches, or `200 OK` only when the response includes a new ETag that differs from the cached value and a valid response body.

## Authentication Reuse

Most logged-in back-office tests reuse a pre-generated admin `storageState`.
The file is regenerated by `global-setup.ts` before each suite run and stored in `temp/storage-state/admin.json`.
Authentication-focused tests still keep their own real login flows when they need to verify login behavior directly.

## Wait Strategy

High-frequency page objects should use shared state-based helpers from `support/ui.ts` instead of fixed sleeps.
Prefer waiting for:

- route readiness
- table visibility and loading-mask removal
- dialog readiness and skeleton removal
- dropdown visibility
- confirmation overlays

Use fixed `waitForTimeout` calls only when a test has a clear business reason that cannot be modeled with deterministic UI signals.

## Governance

Run `pnpm test:validate` whenever you add, rename, or move test files.
The validator checks:

- duplicate `TC` identifiers
- non-`TC` files under `e2e/`
- files outside the allowed module scopes
- broken smoke and serial manifest references
- missing serial isolation categories
- high-risk shared-state patterns that are still in the parallel pool
- parallel isolation allowlist entries without documented reasons

When adding a new test case, update `config/execution-manifest.json` if the new file must join the smoke pack, serial pool, or a new module scope.
