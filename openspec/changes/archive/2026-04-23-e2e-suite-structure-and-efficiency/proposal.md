## Why

The management workbench has already converged to stable capability boundaries such as `dashboard`, `iam`, `settings`, `scheduler`, `extension`, and `about`, plus the corresponding plugin-owned capability areas. However, `hack/tests/e2e/` still keeps a large amount of the old workbench-oriented directory structure. In particular, the oversized `system/` bucket still carries most test files, so the suite no longer aligns with the current menu boundaries or plugin ownership model. At the same time, the Playwright suite still runs with a single worker, repeats UI login in every test, and relies heavily on `waitForTimeout`, which keeps driving full-regression duration upward and slows down developer feedback.

This change upgrades the E2E suite from merely being runnable to being structurally clear, easy to navigate by module, layered for execution, and materially faster to run. That creates a stable testing foundation for future menu evolution, plugin expansion, and daily regression work.

## What Changes

- Reorganize `hack/tests/e2e/` by the current stable workbench capability boundaries and plugin ownership, split the overloaded `system/` bucket, and allow second-level module directories such as `scheduler/job/`.
- Move shared helpers and debug scripts that do not follow the test-file convention out of `e2e/`, and fix duplicate TC IDs and non-`TC*.ts` files mixed into the suite tree.
- Add layered Playwright execution entrypoints that at minimum support `smoke`, module-scoped execution, and `full`, so everyday development no longer depends on running the entire suite for every feedback loop.
- Introduce authenticated state reuse so high-frequency fixtures such as `adminPage` use a pre-generated login state instead of performing a full UI login in every test.
- Replace fixed-duration waits in high-frequency page objects and shared fixtures with deterministic state-based waiting, and establish explicit serial versus parallel safety boundaries for files.
- Add E2E suite governance documentation and validation scripts so future tests continue to respect directory ownership, TC uniqueness, helper placement, and layered execution rules.

## Capabilities

### New Capabilities

- `e2e-suite-organization`: Governs the E2E directory layout, module ownership, helper placement, and TC numbering so the suite structure matches the current workbench capability boundaries.
- `e2e-suite-execution-efficiency`: Governs layered execution, authenticated-state reuse, state-based waiting, and parallel-safety boundaries to reduce both full-regression time and day-to-day development feedback cost.

### Modified Capabilities

- None.

## Impact

- **Test directories**: `hack/tests/e2e/`, internal imports, and helper/debug file locations are reorganized.
- **Test runner**: `hack/tests/playwright.config.ts` and `hack/tests/package.json` gain layered execution entrypoints, login-state preparation, and worker-strategy support.
- **Fixtures and page objects**: high-frequency login and wait logic in `hack/tests/fixtures/` and `hack/tests/pages/` is refactored.
- **Documentation and governance**: `hack/tests` documentation is updated and automated validation is added to prevent duplicate TC IDs, misplaced files, and broken manifest references.
- **Regression verification**: key module regressions and full-suite timing/stability baselines must be revalidated after the migration.
