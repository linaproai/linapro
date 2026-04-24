## Why

The current authorization review dialog that appears before a dynamic plugin is installed or enabled only shows the host-service request list. It does not show which dynamic routes the plugin is about to expose. For administrators, `hostServices` explains what host resources a plugin wants to access, while dynamic routes explain which external entry points the plugin will add. Without the latter, administrators cannot fully evaluate the plugin's external exposure before granting approval, especially for `public` routes, login-governed routes, and routes that bind to specific permissions.

This iteration brings declared dynamic-route contracts into the authorization review view so administrators can inspect both sides of plugin governance before installation or enablement: what the plugin will access and what the plugin will expose.

## What Changes

- Add a read-only "Route Information" section to the dynamic-plugin installation and enablement review dialog, placed after the "Host Service Authorization Scope" section.
- Show only the first two routes by default and reveal the remainder through an explicit expand action so the dialog does not become overloaded.
- Show each route's method, real host public path, access level (for example `public` or `login`), permission key, and summary.
- Keep route exposure clearly separate from host-service authorization. The two sections are presented side by side in the same governance dialog, but routes remain a review artifact rather than an authorization item.
- Extend the backend plugin-management projection so the frontend dialog can consume route review data directly from list and detail payloads without assembling temporary client-side state.
- Reuse the same route review structure in the plugin detail dialog so administrators can inspect the current release after installation.
- Improve plugin detail presentation by showing long descriptions as a dedicated full-width row in the base information table.
- Strengthen the visual hierarchy of governance section titles in both the authorization dialog and the detail dialog.
- Remove the redundant "Authorization Requirement" field from the plugin detail dialog and keep the plugin description inside the base information table.
- Add backend, frontend, and E2E verification for empty routes, `login` routes, `public` routes, and permission-bound routes.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `plugin-manifest-lifecycle`: extend the dynamic-plugin authorization and detail views so administrators can review route exposure alongside host-service authorization.
- `demo-control-guard`: align demo read-only activation with the plugin's enabled state and clarify the write-blocking response contract.

## Impact

- **Backend APIs**: plugin-management DTOs and list/detail projections now include route-review fields for dynamic plugins.
- **Backend services**: affects dynamic-plugin manifest and release projection logic under `apps/lina-core/internal/service/plugin/` and `apps/lina-core/internal/controller/plugin/`.
- **Frontend UI**: affects the dynamic-plugin authorization dialog, plugin detail dialog, and related view models under `apps/lina-vben/apps/web-antd/src/views/system/plugin/`.
- **Demo governance**: affects the `demo-control` runtime guard and its error serialization behavior.
- **Tests**: requires backend projection tests and E2E coverage for the authorization dialog, plugin detail dialog, and related regressions.
