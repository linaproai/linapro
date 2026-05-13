# media

`media` is a LinaPro source plugin for media strategy, strategy binding, and stream alias management.

This module is intentionally Chinese-only for user-facing copy. It does not ship runtime i18n JSON or apidoc i18n JSON.

## Capabilities

- Media strategy CRUD and global strategy selection
- Device, tenant, and tenant-device strategy bindings
- Effective strategy preview with priority resolution
- Stream alias CRUD

## Development

- Backend entry: `backend/plugin.go`
- API DTOs: `backend/api/media/v1/`
- Service implementation: `backend/internal/service/media/`
- Frontend page: `frontend/pages/media-management.vue`
- PostgreSQL install SQL: `manifest/sql/001-media-schema.sql`
- Mock demo data: `manifest/sql/mock-data/001-media-mock-data.sql`
