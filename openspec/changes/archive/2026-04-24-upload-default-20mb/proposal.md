## Why

The default upload size limit is currently inconsistent across the host initialization SQL, the config template, and the backend static fallback value. The repository still carries both a 10 MB and a 16 MB baseline. That split causes new environments, default runtime paths, and upload error messages to behave differently depending on which path is used. We need one unified 20 MB default so fresh initialization and default host behavior stay aligned.

## What Changes

- Change the host default value of built-in runtime parameter `sys.upload.maxSize` to 20 MB so config-management seed data matches real default behavior.
- Align the host static upload fallback, config-template default, and request-size protection path to remove the current 10 MB / 16 MB split.
- Update the affected upload-limit validation and friendly error-message tests so the default 20 MB baseline behaves consistently across file upload and transport-limit enforcement.

## Capabilities

### New Capabilities

### Modified Capabilities
- `config-management`: Add a unified 20 MB default-value constraint for built-in runtime parameter `sys.upload.maxSize` and require all host fallback behavior to match that default.

## Impact

- `apps/lina-core/manifest/sql/007-config-management.sql` needs to update its built-in config seed data.
- `apps/lina-core/manifest/config/config.template.yaml` and the host upload-config fallback logic must both move to 20 MB.
- Upload-limit logic and test assertions under `apps/lina-core/internal/service/config/`, `apps/lina-core/internal/service/file/`, and `apps/lina-core/internal/service/middleware/` are affected.
- Any embedded or packaged manifest artifacts must be regenerated so build outputs do not keep shipping the old default.
