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

## 5. Documentation and Verification

- [x] 5.1 Add `README.md` and `README.zh_CN.md` for `hack/tests/` to document directory boundaries, execution entrypoints, manifest mechanics, and governance-script usage.
- [x] 5.2 Run the governance validator, `pnpm test:smoke`, at least two module-scoped regressions, and `pnpm test` / `pnpm test:full`, then record timing and stability baselines before and after the migration.
  - Completed `pnpm run test:validate`, `pnpm test:smoke`, `pnpm run test:module -- iam:user`, `pnpm run test:module -- settings:config`, `pnpm run test:module -- settings:dict`, and `pnpm run test:full`.
  - `pnpm run test:full` passed on 2026-04-23 with 80 parallel-pool test files passing and 262 serial-pool test files passing.
  - For the second fixed-wait cleanup and plugin-regression fixes, additional targeted regressions were rerun for notice, department, menu, online users, operation/login logs, dictionary cascading delete, login failure, message panel, user selector, dictionary export, host-boundary regression, and scheduling scenarios, including `TC0002`, `TC0024`, `TC0025`, `TC0026`, `TC0031`, `TC0037~TC0040`, `TC0043`, `TC0048~TC0051`, `TC0056`, `TC0059`, `TC0060`, `TC0063`, `TC0081`, `TC0082`, `TC0084`, `TC0085`, `TC0089`, `TC0090`, `TC0097`, and `TC0099`; all passed.
- [x] 5.3 Cross-check the delivered directory structure, execution layering, auth-state reuse, and parallel boundaries against the OpenSpec design and specs so the change is ready for `/opsx:apply` completion and implementation review.
  - The suite self-check for directory boundaries, execution manifest behavior, `storageState` reuse, real-login coverage for auth scenarios, and serial/parallel layering has been completed.
  - The only previously noted leftovers were existing plugin failures in the full run and a second round of fixed-wait cleanup; both were resolved by the completed feedback work recorded below.

## Feedback

- [x] **FB-1**: After enabling a source plugin, public `ping` and protected `summary` routes in `TC-66d` returned `404`.
- [x] **FB-2**: The bundled-runtime plugin upload probe in `TC-67m` did not return a success payload when the request body exceeded 8 MB.
