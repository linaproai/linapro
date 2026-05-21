## Why

The current E2E regression runtime has become a delivery bottleneck: host-only mode takes approximately 36 minutes, and plugin-full mode takes approximately 2 hours. Logs show that the main problem with plugin-full is a large number of serial files concentrated in a single CI job, and the main problem with host-only is that many page test cases repeatedly bear the cost of logged-in page initialization, default dashboard loading, and business page navigation.

This change systematically reduces E2E wall clock time while preserving isolation semantics for high-risk scenarios such as plugin lifecycle, permissions, i18n, menus, and shared data.

## What Changes

- Adjust plugin-full E2E from a single full-volume job to CI shard execution based on generic entry points, covering host plugin framework generic cases and all source plugin owned cases.
- Clarify the responsibility boundary between plugin-full and host-only, avoiding plugin-full indiscriminately repeating the complete host suite; root directory E2E only retains host plugin framework, dynamic test plugin, and generic plugin governance coverage that does not depend on specific official plugins.
- Source plugin case selection only retains generic entry points: `plugins` covers all source plugins, `plugin:<plugin-id>` covers a single source plugin, no longer maintaining long-term alias scopes named by official plugin business modules.
- Add host-only single module entry point, allowing running only specified host modules in the main framework environment without `apps/lina-plugins`.
- Add lightweight fixture for logged-in pages, allowing tests to directly navigate to target business routes without loading the default dashboard for each case.
- Add idempotent suite-level baseline setup capability for plugin E2E, reducing repeated synchronization, installation, enablement, mock data loading, and plugin projection refresh in ordinary plugin page test cases.
- Refactor the most time-consuming plugin lifecycle test cases, so complete UI lifecycle only covers representative plugins, while other official plugins use API/contract smoke plus page accessibility verification.
- Preserve per-E2E test case timing records and continue uploading Playwright report, test-results, and service logs as CI artifacts for optimization verification.
- Fix CI PostgreSQL health check authentication parameters, reducing invalid health check logs and potential wait jitter.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `e2e-suite-execution-efficiency`: Add plugin-full shard execution, lightweight authentication fixture, plugin baseline, and timing verification requirements.
- `e2e-suite-organization`: Clarify the responsibility boundary between plugin-full and host-only, requiring complete verification chain coverage for official plugins without indiscriminately repeating the complete host suite.

## Impact

- Affects CI workflow: `.github/workflows/reusable-test-verification-suite.yml`, `.github/workflows/reusable-e2e-tests.yml`.
- Affects E2E runner, manifest, and test documentation: `hack/tests/scripts/run-suite.mjs`, `hack/tests/scripts/execution-governance.mjs`, `hack/tests/config/execution-manifest.json`, `hack/tests/README.md`, `hack/tests/README.zh-CN.md`.
- Affects Playwright fixtures and support tools: `hack/tests/fixtures/auth.ts`, `hack/tests/fixtures/plugin.ts`, `hack/tests/support/ui.ts`.
- Affects some high-time-cost E2E test cases, focusing on dynamic runtime lifecycle, official source plugin lifecycle, menu CRUD, file management, and similar pattern host page cases.
- Does not change production API, database schema, runtime cache semantics, or user-visible functionality; no new i18n resources are needed for this change.
