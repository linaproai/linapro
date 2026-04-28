## Why

The backend Go code contained several `panic` calls. Some were valid fail-fast paths during startup or source-plugin registration, but others appeared in runtime request handling, import/export flows, dynamic plugin input handling, and runtime configuration loading. Those runtime panics could bypass unified error responses, logging, and API response governance.

This change defines the allowed `panic` boundary and converts unnecessary runtime panics into explicit `error` returns or controlled local handling, so ordinary business paths do not trigger process-level failures for recoverable errors.

## What Changes

- Define that production backend code may use `panic` only for startup, initialization, unrecoverable critical paths, `Must*` semantic constructors, and unknown panic rethrow paths that truly need to be preserved.
- Convert unnecessary runtime panics in Excel cell coordinate conversion, shared resource closing, runtime configuration reading, and dynamic plugin hostServices normalization into explicit error returns. Controlled logging is retained only for cleanup paths that cannot return errors and do not affect the main flow.
- Preserve fail-fast behavior for source-plugin registration contracts, plugin extension declarations, DB driver registration, and other startup or registration-time failures, with tests to prevent accidental removal.
- Add backend panic governance unit tests and static checks to prevent ordinary business paths from reintroducing unnecessary `panic` calls.
- This change does not add external APIs, database schema changes, frontend runtime copy, or i18n resource changes.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `backend-conformance`: Adds production backend `panic` boundary requirements and explicit runtime error-return requirements.

## Impact

- Code impact: `apps/lina-core/internal/service/config`, `apps/lina-core/internal/service/{user,dict,sysconfig}`, `apps/lina-core/pkg/{excelutil,closeutil,pluginbridge}`, and related dynamic plugin and monitoring-plugin export logic.
- Test impact: Adds or updates Go unit tests for Excel helpers, invalid runtime configuration values, dynamic plugin hostServices validation, and allowed startup fail-fast paths.
- Governance impact: Adds a static check or test entry point that enforces an allowlist for new `panic` calls in production backend code.
- i18n impact: No user-visible copy is added, modified, or removed. Error messages continue through the existing backend error-return path, so no frontend locale bundle, manifest i18n, or apidoc i18n resource updates are required.
