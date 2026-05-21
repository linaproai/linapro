## 1. Implementation

- [x] 1.1 Adjust ordinary file local storage path generation, removing `t` directory layer from new upload paths.
- [x] 1.2 Synchronously update upload path comments and API examples, avoiding continued display of old path format.

## 2. Testing

- [x] 2.1 Supplement or update file service unit tests, covering new upload path format.
- [x] 2.2 Supplement or confirm old `t/...` paths still accessible by database record.

## 3. Verification and Review

- [x] 3.1 Run `openspec validate --strict`.
- [x] 3.2 Run `cd apps/lina-core && go test ./internal/service/file -count=1`.
- [x] 3.3 Run backend startup binding compile smoke `cd apps/lina-core && go test ./internal/cmd -count=1`.
- [x] 3.4 Record i18n, cache consistency, data permission, development tools cross-platform impact assessment, and execute `lina-review` review.

## 4. Completion Records

- Implementation: New uploaded files' local relative storage path changed from `t/<tenantId>/<yyyy>/<MM>/<filename>` to `<tenantId>/<yyyy>/<MM>/<filename>`; historical `sys_file.path` records not migrated, reads continue by database record path access.
- Testing: Added `TestLocalStoragePutUsesTenantIDWithoutTenantPrefix` covering new path format; added `TestOpenByPathPreservesLegacyTenantPrefixPath` covering old `t/...` path compatible read.
- i18n: No new, modified, or deleted user-visible runtime copy, frontend language packs, plugin manifest/i18n, or apidoc i18n resources; only synchronized Go API DTO English example values.
- Cache consistency: No new or modified cache; file access still uses `sys_file.path` database record as authoritative path, no new cache consistency risk in distributed environment.
- Data permission: No new or modified data operation interfaces; upload records still write to current tenant, reads still depend on existing metadata query, tenant filtering, and data permission validation.
- Development tools cross-platform: No new or modified development tools, scripts, CI entries, or default development commands.
- Verification: `openspec validate --strict` passed; `cd apps/lina-core && go test ./internal/service/file -count=1` passed; `cd apps/lina-core && go test ./api/file/v1 ./api/user/v1 -count=1` passed; `cd apps/lina-core && go test ./internal/cmd -count=1` passed.
- Review: Checked this change's OpenSpec records, Go backend path generation, API examples, old path compatibility tests, Go compile gate, i18n, cache consistency, data permission, and development tools cross-platform impact per `lina-review` scope, no blocking issues found.
