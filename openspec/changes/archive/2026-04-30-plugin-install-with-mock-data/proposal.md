# Plugin Install With Mock Data

## Why

Plugins can ship demonstration SQL under `manifest/sql/mock-data/`, but manual installation did not execute those files. New users saw empty plugin pages after installation, and testers lacked ready-to-use sample data.

## What Changes

- Add an `installMockData` option to manual plugin installation.
- Execute mock SQL after install SQL succeeds, with all mock files and mock ledger writes in one transaction.
- Return actionable structured error details when mock loading fails and rolls back.
- Support source plugins and dynamic plugins through the same mock-data directory convention.
- Extend startup auto-enable entries so explicit `withMockData=true` can load mock data on first installation.
- Track mock SQL execution with `direction='mock'` ledger rows.
- Add plugin list metadata and UI indicators for mock-data availability.
- Improve uninstall warning copy and hard-delete semantics for plugin-owned storage cleanup.

## I18n Impact

Frontend runtime language packs, API documentation translations, and error resources were updated for the new checkbox, help text, list column, and rollback warning. English OpenAPI source text remains in DTOs; non-English apidoc resources carry translations.
