## Context

Ordinary file uploads use `LocalStorage.Put` to generate relative storage paths and write that path to `sys_file.path`. The current rule is `t/<tenantId>/<yyyy>/<MM>/<generatedName>`, where `t` is the abbreviation for tenant. When accessing files, the system takes the relative path from the URL, queries `sys_file.path`, and reads the file through the storage backend, so historical record accessibility depends on the database path, not a fixed new path rule.

## Goals / Non-Goals

**Goals:**

- Remove the extra `t` directory layer from new uploads.
- Preserve tenant ID directory, continuing to support physical file partitioning by tenant.
- Do not migrate historical `t/...` files, ensuring old records continue to be accessible through existing database paths.
- Cover new path rule and old path compatibility boundaries with unit tests.

**Non-Goals:**

- Do not delete tenant ID directory.
- Do not modify database schema, DAO, entity, or existing file records.
- Do not adjust plugin host storage service path rules.
- Do not change `/api/v1/uploads/*path` access interface, permission model, or data permission boundaries.

## Decisions

### Only remove literal t prefix

New path adopts `<tenantId>/<yyyy>/<MM>/<filename>`. This satisfies removing the `t` directory requirement while preserving tenant isolation and the directory hierarchy needed for operations locating.

Alternative is completely removing tenant directory, only keeping `<yyyy>/<MM>/<filename>`. This approach reduces physical directory-level tenant identifiability and makes subsequent per-tenant cleanup or migration more difficult, so not adopted.

### Do not migrate historical files

Historical `sys_file.path` has already saved complete relative paths, and old files can continue to be accessed through `t/...` paths. Migrating historical files would involve filesystem moves, database batch updates, failure rollback, and duplicate hash record handling, exceeding this low-risk path rule adjustment scope.

### Maintain hash deduplication reuse status

Upload flow continues to check duplicates by `hash + tenant_id`. If upload content matches historical file, system continues reusing existing historical path; only when truly writing new physical file does it adopt new path rule. This avoids duplicate storage of the same file and avoids breaking existing deduplication semantics for path appearance.

## Risks / Trade-offs

- Old and new file path formats will coexist for a period. Mitigation: Clearly document compatibility strategy and confirm old `t/...` paths still accessible through tests.
- Same content repeated upload may still return old `t/...` path. Mitigation: Document this behavior as part of non-migration strategy, avoiding misjudgment as bug.
- Comments and API examples may be inconsistent with implementation. Mitigation: Synchronously update related examples and run validation.

## Required Assessments

- i18n: No new, modified, or deleted user-visible copy; no need to maintain frontend runtime language packs, plugin manifest/i18n, or apidoc i18n.
- Cache consistency: No new or modified cache; file access still uses database `sys_file.path` as authoritative path, no new cache consistency risk in distributed environment.
- Data permission: No new or modified data operation interfaces; upload records still write to current tenant, reads still through existing metadata and tenant filtering validation.
- Development tools cross-platform: No new or modified development tools, scripts, CI entries, or default development commands.
