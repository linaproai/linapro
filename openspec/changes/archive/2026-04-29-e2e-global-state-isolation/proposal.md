## Why

Full E2E regressions have exposed shared-state contention between tests: plugin lifecycle, runtime i18n caches, global configuration, dictionaries, menu permissions, and other cases can mutate the same database or cache state when executed in parallel, causing protocol-correct tests to produce false failures. This change captures those lessons as an executable E2E governance change to reduce instability in future full-regression runs.

## What Changes

- Establish global-state mutation classification rules for E2E test cases, clarifying which tests must enter a serial boundary and which can safely run in parallel.
- Extend test runner scripts and validation tooling to detect high-risk shared-state cases such as plugin lifecycle, system parameters, dictionaries, menu permissions, and runtime i18n caches.
- Consolidate tests that depend on plugin installation, enabling, mock data, and global prerequisites into unified fixtures to avoid implicit cross-file dependencies.
- Adjust assertion strategy for cache/ETag tests: validate protocol semantics rather than assuming global versions remain unchanged throughout the full regression.
- Establish stable data assertion conventions: prefer API stable fields, IDs, codes, or labelKeys for business state; test display copy separately.
- Add regression documentation recording identified conflict types, remediation approaches, classification rules, and ongoing maintenance requirements.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `e2e-suite-execution-efficiency`: Enhanced E2E serial/parallel boundary, shared global state isolation, cache revalidation assertions, and fixture-based prerequisite requirements.

## Impact

- Affects `hack/tests/scripts/run-suite.mjs`, `hack/tests/scripts/validate-e2e.mjs`, `hack/tests/fixtures/`, `hack/tests/support/`, and related E2E test cases.
- Affects E2E execution strategy: some high-risk global-state test cases will explicitly enter a serial boundary; independent read-only or local-data cases continue to run in parallel.
- No business API behavior changes; no database schema changes.
- i18n impact: this change does not introduce product runtime copy; it only involves test governance documentation and test assertions. When modifying language-switch or runtime i18n cache test cases, semantic validation of `i18n` behavior must be explicitly maintained.
