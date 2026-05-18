## Why

The management workbench has converged to stable capability boundaries such as `dashboard`, `iam`, `settings`, `scheduler`, `extension`, and `about`, plus plugin-owned capability areas. However, `hack/tests/e2e/` still keeps a large amount of the old workbench-oriented directory structure, with the oversized `system/` bucket carrying most test files. The suite no longer aligns with current menu boundaries or plugin ownership. Meanwhile, the Playwright suite runs with a single worker, repeats UI login in every test, relies heavily on `waitForTimeout`, and has no auditable shared-state isolation model, all of which drive full-regression duration upward and slow developer feedback.

Full E2E regressions have further exposed shared-state contention between tests: plugin lifecycle, runtime i18n caches, global configuration, dictionaries, menu permissions, and other cases mutate the same database or cache state when executed in parallel, causing protocol-correct tests to produce false failures. Without machine-readable isolation categories, heuristic detection, and governance documentation, the suite cannot safely scale parallelism or onboard new high-risk test cases.

This change upgrades the E2E suite from merely being runnable to being structurally clear, easy to navigate by module, layered for execution, materially faster to run, and governed by explicit shared-state isolation rules. That creates a stable testing foundation for future menu evolution, plugin expansion, and daily regression work.

## What Changes

- Reorganize `hack/tests/e2e/` by the current stable workbench capability boundaries and plugin ownership, split the overloaded `system/` bucket, and allow second-level module directories such as `scheduler/job/`.
- Move shared helpers and debug scripts that do not follow the test-file convention out of `e2e/`, and fix duplicate TC IDs and non-`TC*.ts` files mixed into the suite tree.
- Add layered Playwright execution entrypoints that at minimum support `smoke`, module-scoped execution, and `full`, so everyday development no longer depends on running the entire suite for every feedback loop.
- Introduce authenticated state reuse so high-frequency fixtures such as `adminPage` use a pre-generated login state instead of performing a full UI login in every test.
- Replace fixed-duration waits in high-frequency page objects and shared fixtures with deterministic state-based waiting, and establish explicit serial versus parallel safety boundaries for files.
- Establish global-state mutation classification rules for E2E test cases, clarifying which tests must enter a serial boundary and which can safely run in parallel.
- Extend test runner scripts and validation tooling to detect high-risk shared-state cases such as plugin lifecycle, system parameters, dictionaries, menu permissions, and runtime i18n caches.
- Consolidate tests that depend on plugin installation, enabling, mock data, and global prerequisites into unified fixtures to avoid implicit cross-file dependencies.
- Adjust assertion strategy for cache/ETag tests to validate protocol semantics rather than assuming global versions remain unchanged throughout the full regression.
- Establish stable data assertion conventions: prefer API stable fields, IDs, codes, or labelKeys for business state; test display copy separately.
- Add E2E suite governance documentation and validation scripts so future tests continue to respect directory ownership, TC uniqueness, helper placement, execution rules, isolation categories, and conflict governance.

## Capabilities

### New Capabilities

- `e2e-suite-organization`: Governs the E2E directory layout, module ownership, helper placement, and TC numbering so the suite structure matches the current workbench capability boundaries.
- `e2e-suite-execution-efficiency`: Governs layered execution, authenticated-state reuse, state-based waiting, parallel-safety boundaries, shared global state isolation, cache revalidation assertions, fixture-based prerequisites, and stable data assertion conventions to reduce both full-regression time and day-to-day development feedback cost.

### Modified Capabilities

- None.

## Impact

- **Test directories**: `hack/tests/e2e/`, internal imports, and helper/debug file locations are reorganized.
- **Test runner**: `hack/tests/playwright.config.ts`, `hack/tests/package.json`, `hack/tests/scripts/run-suite.mjs`, and `hack/tests/scripts/validate-e2e.mjs` gain layered execution entrypoints, login-state preparation, worker-strategy support, isolation category reporting, and high-risk heuristic detection.
- **Fixtures and page objects**: high-frequency login and wait logic in `hack/tests/fixtures/` and `hack/tests/pages/` is refactored; plugin and mock-data prerequisites are consolidated into idempotent fixtures.
- **Documentation and governance**: `hack/tests` documentation is updated and automated validation is added to prevent duplicate TC IDs, misplaced files, broken manifest references, unclassified high-risk patterns, and missing isolation categories.
- **Regression verification**: key module regressions and full-suite timing/stability baselines must be revalidated after the migration.
