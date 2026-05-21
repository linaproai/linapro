## ADDED Requirements

### Requirement: Plugin File Storage Path Tenant Prefix
Plugin uploading files through host storage service SHALL auto-inject `tenant=<id>` in path prefix: `/storage/t/<tenant_id>/plugin-<plugin-id>/...`.

### Requirement: Cross-Tenant Access Requires Explicit Platform Interface
Plugin file read SHALL validate `bizctx.TenantId` matches `sys_file.tenant_id`; mismatch returns 403. Platform admin can only cross-tenant access through explicit `/platform/*` read-only interface or dedicated platform host service.

### Requirement: Platform Shared File Path
Plugin storing cross-tenant shared files SHALL through `storage.SaveAsPlatform(...)` explicitly write to `/storage/t/0/...`, requiring platform context with all data permissions.
