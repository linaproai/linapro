## Why

The current relative storage path for ordinary file uploads includes a `t/<tenantId>/...` prefix, where `t` is just an abbreviation for tenant, which is not intuitive enough for callers and operations troubleshooting. Users want to remove this extra directory layer while preserving the physical organization partitioned by tenant ID.

## What Changes

- New uploaded files' relative storage path changes from `t/<tenantId>/<yyyy>/<MM>/<filename>` to `<tenantId>/<yyyy>/<MM>/<filename>`.
- Preserve existing historical paths recorded in `sys_file.path`, not performing historical file migration.
- Download and URL access continue to use the relative path from database records, so old `t/...` files and new path files can coexist for access.
- Update file upload path examples, unit tests, and implementation task records.

## Capabilities

### New Capabilities

- `file-upload-storage-path`: Constrain ordinary file upload tenant-partitioned relative path, historical path compatibility, and verification requirements.

### Modified Capabilities

- None.

## Impact

- Affects `apps/lina-core/internal/service/file` local storage path generation logic.
- Affects file upload related unit tests and API documentation examples.
- Does not change public upload, download, access API paths, does not change database schema, does not migrate existing files.
- Does not add user-visible copy, does not affect frontend runtime language packs or plugin manifest/i18n.
