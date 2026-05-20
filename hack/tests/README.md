# E2E Test Suite

This directory contains the Playwright `E2E` suite for the default LinaPro management workbench and host-plugin integration flows.

## Prerequisites

Before running E2E tests for the first time, install frontend dependencies and the Playwright browser:

```bash
make dev.setup
```

This command runs `pnpm install` for the frontend workspace (if missing) and downloads the Chromium browser binary with required system libraries (`libnss3`, `libatk-bridge2.0-0`, etc. on Linux). On macOS and Windows it only downloads the browser. You only need to run this once (or again after upgrading dependencies).

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

Source plugins may keep their own test surface inside the plugin directory:

```text
apps/lina-plugins/<plugin-id>/
  hack/tests/e2e/       plugin-owned TC test cases
  hack/tests/pages/     plugin-owned page objects
  hack/tests/support/   optional plugin-owned E2E helpers
```

The `e2e/` tree is organized by stable capability boundaries instead of the legacy `system/` bucket:

- `auth/`, `dashboard/`, `about/`
- `iam/`
- `settings/`
- `scheduler/`
- `extension/`

Source-plugin-owned business coverage lives in each plugin directory instead of the host `e2e/` tree. The root suite only exposes generic plugin selection:

- `plugins` maps to every `apps/lina-plugins/<plugin-id>/hack/tests/e2e/` tree.
- `plugin:<plugin-id>` maps to one `apps/lina-plugins/<plugin-id>/hack/tests/e2e/` tree.
- Nested plugin paths are resolved from the owning plugin `hack/tests/e2e/` tree when needed.

## Naming Rules

- Test files must use `TC{NNN}-{brief-name}.ts`.
- `TC` identifiers are module-local: each owning E2E directory starts at `TC001` and increments continuously inside that directory.
- Do not reserve or skip numbers because another host module or plugin module has already used them.
- Only real `TC` files may live under `hack/tests/e2e/`.
- Shared helpers must live in `fixtures/`, `support/`, `scripts/`, or `debug/`.

## Execution Entrypoints

| Command | Purpose |
| --- | --- |
| `pnpm test` | Run the full layered suite. |
| `pnpm test:full` | Run the full layered suite explicitly. |
| `pnpm test:host` | Run host-owned E2E cases with plugin-environment exclusions. |
| `pnpm test:host:module -- <scope>` | Run one host-only module scope with plugin-environment exclusions. |
| `pnpm test:smoke` | Run the curated high-value smoke pack. |
| `pnpm test:module -- <scope>` | Run a module scope from the execution manifest. |
| `pnpm test:validate` | Validate `TC` uniqueness, directory ownership, and manifest references. |
| `pnpm report` | Open the Playwright HTML report. |

Example module scopes:

- `iam:user`
- `settings:config`
- `scheduler:job`
- `extension:plugin`
- `plugins`
- `plugin:<plugin-id>` for a source plugin with tests under `apps/lina-plugins/<plugin-id>/hack/tests/e2e/`

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
Full-suite runs include plugin-owned tests through the generic `plugins` entry.
Plugin-owned module selection intentionally stays generic: use `plugins` for all source-plugin-owned E2E files or `plugin:<plugin-id>` for one source plugin.
Use `pnpm test:host:module -- <scope>` to run only the host-owned portion of a scope without requiring `apps/lina-plugins`.
Individual source plugins can be run without editing the manifest:

```bash
pnpm test:module -- plugin:<plugin-id>
```

The root `hack/tests` tree must not hardcode concrete source-plugin IDs, plugin-owned routes, plugin-specific mock data, plugin-specific test configuration, plugin-specific baseline data, plugin-specific i18n keys, or plugin-specific page objects. Keep plugin behavior, plugin test data, plugin configuration, and plugin POMs under the owning `apps/lina-plugins/<plugin-id>/hack/tests/` tree. The root suite may keep only generic plugin discovery and runner mechanics.

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
Plugin-owned tests may use `fixtures/plugin.ts` to sync source plugins, install or enable their owning plugin, refresh the frontend projection, and load matching plugin mock SQL when present.
Tests that create users, departments, posts, notices, files, plugins, import rows, or export artifacts should use unique names or stable test prefixes and clean up their own data in `finally`, `afterEach`, or `afterAll`.

## Cache Revalidation

Cache and ETag tests should validate protocol semantics instead of assuming the resource version stays unchanged during a full regression.
A conditional request must prove that the request carries the expected precondition.
It may accept `304 Not Modified` when the ETag still matches, or `200 OK` only when the response includes a new ETag that differs from the cached value and a valid response body.

## Authentication Reuse

Most logged-in back-office tests reuse a pre-generated admin `storageState`.
The file is regenerated by `global-setup.ts` before each suite run and stored in `temp/storage-state/admin.json`.
Authentication-focused tests still keep their own real login flows when they need to verify login behavior directly.

## User-Flow Testing

`CRUD` and other user-visible workflows should be written as real workbench interactions, not as API-only checks.
Use page objects to click visible buttons, fill labeled form controls, operate dialogs, select table rows, confirm destructive actions, and assert the resulting page state.

Prefer accessible and user-facing locators such as `getByRole`, `getByLabel`, `getByText`, and stable `data-testid` hooks when the component does not expose reliable accessible names.
Direct API calls or database helpers are appropriate for fixture setup, cleanup, and hard-to-reach boundary state, but the behavior under test should normally pass through the UI.

For management-page `CRUD` coverage, include the core path when it is relevant to the feature:

- page loads and the table becomes ready
- search and reset
- create through the visible form
- find the created row from the list
- edit through the visible action
- delete with the real confirmation control
- verify the row is no longer visible
- validate required-field, uniqueness, or permission boundaries when the feature owns them

Screenshots are diagnostic artifacts, not the primary assertion for `CRUD` correctness.
Assert observable business outcomes such as success messages, table rows, field values, empty states, disabled or hidden actions, and network responses for export/import flows.

## Failure Artifacts

Playwright keeps `screenshot`, `trace`, and `video` artifacts when a test fails.
Use `pnpm report` to inspect the `HTML` report and open the trace for the exact click sequence, `DOM` snapshot, console output, and network activity.

The suite retries once in `CI` by default and can be controlled locally with `E2E_RETRIES`.

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

- duplicate `TC` identifiers in one module directory
- non-continuous module-local `TC` numbering
- legacy four-digit global `TC` filenames
- non-`TC` files under `e2e/`
- files outside the allowed module scopes
- broken smoke and serial manifest references
- missing serial isolation categories
- high-risk shared-state patterns that are still in the parallel pool
- parallel isolation allowlist entries without documented reasons

When adding a new test case, update `config/execution-manifest.json` if the new file must join the smoke pack, serial pool, or a new module scope.
