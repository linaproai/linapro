# Plugin API Query Performance

## Why

Plugin management list queries and authentication request paths performed unnecessary write or metadata work. A `GET /api/v1/plugins` call could trigger plugin governance synchronization, host-service table comment lookup could emit schema probing errors, and session active time updates wrote too frequently.

## What Changes

- Make plugin list queries read-only and move synchronization to explicit `POST /api/v1/plugins/sync` actions.
- Make host-service table comment lookup use safe read-only metadata access and degrade to raw table names on failure.
- Throttle online-session `last_active_time` writes over a short window while preserving timeout checks.
- Add focused backend tests for read-only plugin lists, metadata lookup, and session activity throttling.

## I18n Impact

No runtime i18n resources are required. The change only adjusts backend read/write behavior, metadata lookup, and session write frequency.
