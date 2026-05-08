## 1. Baseline Inventory and Governance Skeleton

- [x] 1.1 Inventory the current `hack/tests/e2e/` file tree, output the migration mapping from legacy directories to target capability directories, and record duplicate TC IDs, non-`TC*.ts` files, and high-frequency `waitForTimeout` hotspots.
- [x] 1.2 Create the new execution manifest and support-directory skeleton under `hack/tests/` (for example `config/`, `support/`, `scripts/`, and `debug/`), and define where `smoke`, module scopes, and serial-pool declarations live.
- [x] 1.3 Add an E2E governance validator that automatically checks global TC uniqueness, misplaced files, wrong directory ownership, and broken execution-manifest references.

## 2. Directory Reorganization and Numbering Governance

- [x] 2.1 Create the target capability directory tree, split `hack/tests/e2e/system/` into stable capability boundaries such as `iam`, `settings`, `org`, `content`, and `scheduler`, and repair imports.
- [x] 2.2 Further split directories such as `monitor/` and `plugin/` by plugin capability and subdomain so monitoring and extension-center cases can be located by capability.
- [x] 2.3 Move non-test files such as `hack/tests/e2e/system/job/helpers.ts` and `hack/tests/e2e/debug/export-debug.ts` out of `e2e/` into dedicated support locations.
- [x] 2.4 Fix existing duplicate TC IDs and naming conflicts, and update the affected imports, documentation, and manifest references.

## 3. Execution Entrypoints and Auth-State Reuse

- [x] 3.1 Add pre-generated administrator login-state preparation and `storageState` artifact management for Playwright, and wire the preparation flow into `playwright.config.ts`.
- [x] 3.2 Refactor `hack/tests/fixtures/auth.ts` so high-frequency fixtures such as `adminPage` reuse the prepared login state by default while authentication-focused cases still keep a real login path.
- [x] 3.3 Add `test:smoke`, `test:module`, and `test:full` scripts to `hack/tests/package.json` while preserving the full-regression meaning of `pnpm test`.
- [x] 3.4 Implement parallel-pool and serial-pool scheduling from the execution manifest so full regression runs in two phases: parallel-safe files first, then high-risk serial files.

## 4. Wait Strategy and Speed Improvements

- [x] 4.1 Add shared state-based wait helpers for tables, dialogs, toasts, and route readiness in `hack/tests/support/` or an equivalent shared layer.
- [x] 4.2 Refactor the page objects with the densest fixed waits first, at minimum covering menus, roles, dictionaries, users, and config pages, and replace the main `waitForTimeout` usage with state-based waits.
- [x] 4.3 Review plugin governance, permission governance, import/export, runtime configuration, and other obvious shared-state test files to classify what must stay serial and what can safely move into the parallel pool after isolation fixes.
- [x] 4.4 Run a second cleanup pass over the remaining fixed waits, keep only the few waits that have an explicit business justification, and document that justification with comments.
  - A second fixed-wait cleanup pass was completed for `NoticePage`, `DeptPage`, `JobPage`, `JobLogPage`, `JobGroupPage`, `FilePage`, `RolePage`, `PluginPage`, `MenuPage`, and `TC0002`, `TC0010`, `TC0015`, `TC0017`, `TC0020`, `TC0024`, `TC0025`, `TC0026~TC0036`, `TC0038`, `TC0040`, `TC0046`, `TC0048`, `TC0049`, `TC0050`, `TC0051`, `TC0052`, `TC0056`, `TC0057`, `TC0059`, `TC0060`, `TC0061`, `TC0063`, `TC0064`, `TC0066`, `TC0099`, plus `hack/tests/debug/export-debug.ts`.
  - The repository now has 0 remaining `waitForTimeout` usages in business tests and debug scripts.

## 5. Global State Isolation and Conflict Classification

- [x] 5.1 Review the most recent full E2E regression log, catalog shared-state conflict cases, affected test files, root causes, and remediation directions.
- [x] 5.2 Scan `hack/tests/e2e` for high-risk operations including plugin lifecycle, system parameters, dictionaries, menu permissions, runtime i18n caches, and public config caches.
- [x] 5.3 For each hit, determine the isolation category, whether it must be serial, and whether the conflict can be resolved via fixture or semantic assertion.
- [x] 5.4 Extend `hack/tests/config/execution-manifest.json` to add machine-readable isolation categories for serial entries or high-risk files.
- [x] 5.5 Update `hack/tests/scripts/run-suite.mjs` to output parallel file count, serial file count, worker count, and serial isolation category summary in full, smoke, and module modes.
- [x] 5.6 Ensure module mode continues to apply the same serial/parallel split so that module regressions cannot bypass isolation rules.

## 6. E2E Validation Gate Enhancement

- [x] 6.1 Extend `hack/tests/scripts/validate-e2e.mjs` to validate isolation category format, referenced path existence, and serial entry classification completeness.
- [x] 6.2 Add high-risk heuristic detection covering plugin install/enable/disable/uninstall/sync/upload/upgrade, system parameter writes, dictionary import/modify, menu permission modifications, and runtime i18n ETag cache assertions.
- [x] 6.3 Output actionable error messages for files that hit high-risk patterns but are not in the serial boundary or lack a declared classification.
- [x] 6.4 Add an allowlist structure with documented reasons for justified parallel-safe exceptions and enforce reason entry during validation.
- [x] 6.5 Add script-level unit tests or executable verification cases for the runner covering full/module splitting and category summary output.

## 7. Fixture and Test Case Remediation

- [x] 7.1 Consolidate source-plugin state preparation logic to ensure tests depending on plugin pages/APIs/mock data use idempotent fixtures.
- [x] 7.2 Review local files, attachments, plugin tables, and mock-data cleanup logic in dynamic-plugin and source-plugin tests to ensure single-file independence.
- [x] 7.3 Adjust cache/ETag tests to assert that requests carry conditional headers and accept either `304` or a legitimate `200 + new ETag + body`.
- [x] 7.4 Adjust tests that infer business state from localized UI text; switch business assertions to stable IDs, codes, labelKeys, permission keys, or API counters.
- [x] 7.5 Perform minimal-scope remediation on known conflict-related files, covering at least `TC0124-runtime-i18n-etag.ts`, plugin lifecycle cases, organization department tree count cases, and content notice dependency cases.

## 8. Documentation and Verification

- [x] 8.1 Add `README.md` and `README.zh_CN.md` for `hack/tests/` to document directory boundaries, execution entrypoints, manifest mechanics, governance-script usage, isolation categories, serial boundaries, fixture prerequisites, and cache semantic assertion rules.
- [x] 8.2 Add an E2E conflict governance record listing conflict types, representative cases, remediation approaches, and a checklist for adding new test cases.
- [x] 8.3 Run the governance validator, `pnpm test:smoke`, at least two module-scoped regressions, and `pnpm test` / `pnpm test:full`, then record timing and stability baselines before and after the migration.
  - Completed `pnpm run test:validate`, `pnpm test:smoke`, `pnpm run test:module -- iam:user`, `pnpm run test:module -- settings:config`, `pnpm run test:module -- settings:dict`, and `pnpm test:full`.
  - `pnpm run test:full` passed on 2026-04-23 with 80 parallel-pool test files passing and 262 serial-pool test files passing.
  - For the second fixed-wait cleanup and plugin-regression fixes, additional targeted regressions were rerun for notice, department, menu, online users, operation/login logs, dictionary cascading delete, login failure, message panel, user selector, dictionary export, host-boundary regression, and scheduling scenarios; all passed.
  - `pnpm test:validate` validated 138 E2E files, 28 scopes, 7 smoke files, and 103 serial files.
  - `node scripts/run-suite.mjs module i18n --list` outputs module split and isolation category summary.
  - Affected subset: `TC0066-source-plugin-lifecycle.ts`, `TC0124-runtime-i18n-etag.ts`, `TC0021-user-dept-tree-count.ts`, `TC0037-notice-crud.ts` all passed; `TC0124` passed 5/5 after fix.
  - Full regression second round: parallel phase `102/102` passed; serial phase executed 326 assertions, 260 passed, 6 skipped, 4 not run.
- [x] 8.4 Cross-check the delivered directory structure, execution layering, auth-state reuse, parallel boundaries, isolation categories, and conflict governance against the OpenSpec design and specs so the change is ready for archive.

## Feedback

- [x] **FB-1**: After enabling a source plugin, public `ping` and protected `summary` routes in `TC-66d` returned `404`.
- [x] **FB-2**: The bundled-runtime plugin upload probe in `TC-67m` did not return a success payload when the request body exceeded 8 MB.
