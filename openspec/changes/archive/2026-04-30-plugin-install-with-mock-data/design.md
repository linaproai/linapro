# Design

## Install Option

Manual install accepts `installMockData`. The frontend shows the checkbox only when plugin metadata reports mock data availability. The checkbox is unchecked by default and acts as explicit user opt-in.

## Transaction Model

Install SQL remains outside the mock transaction because delivery DDL cannot be reliably rolled back across MySQL variants. After install SQL succeeds, all mock SQL files and their `sys_plugin_migration` rows execute in a single transaction. Any mock failure rolls back mock data and mock ledger rows while preserving installed plugin state.

## Source and Dynamic Plugins

Source plugins read mock SQL from embedded or source file systems. Dynamic plugin packaging adds mock SQL assets to artifacts so the runtime scanner can use the same path and naming rules.

## Startup Bootstrap

`plugin.autoEnable` entries use structured objects with `id` and optional `withMockData`. Mock data loads only during first-time installation and only when explicitly enabled.

## Frontend

Plugin management displays a dedicated mock-data availability column and help tooltip. The install modal places the mock-data option near plugin base information. Uninstall cleanup warning is visually emphasized.

## Tests

Backend tests cover transactional rollback, startup mock opt-in, artifact packaging, and resource governance. E2E tests cover install with and without mock data and plugin list indicators.
