# Design

## Plugin List Read Path

The plugin list service reads discovered source manifests, dynamic plugin registry data, release snapshots, and current governance projection data without synchronizing governance tables. Explicit synchronization remains available through the sync action and can write registry, menu, resource, and permission governance data.

## Host-Service Metadata Lookup

Table comment lookup uses database metadata APIs or read-only queries that do not cause schema introspection against `information_schema.TABLES`. Lookup failures are non-fatal and fall back to raw table names.

## Session Activity Throttling

Authentication still checks session validity and timeout for every protected request. `last_active_time` writes are skipped when the previous update is inside the configured short throttle window, reducing write pressure without extending expired sessions.

## Tests

Backend tests cover plugin list no-write behavior, explicit sync behavior, metadata fallback, and throttled session active-time updates.
