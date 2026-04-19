## Why

The repository ended up with two active OpenSpec changes at the same time:

- `config-runtime-params`
- `host-managed-plugin-openapi`

That state violated the project rule that one active iteration should be managed and archived as a single change. Both changes were already implemented and verified, but they still needed to be consolidated into one historical record before archiving.

The completed work spanned two closely related host-governance tracks:

1. Protected runtime configuration and public frontend settings that are consumed directly by the host at runtime.
2. Host-managed API documentation and source-plugin route ownership capture so plugin interfaces can be projected according to enablement state.

During the same implementation window, the iteration also absorbed a number of follow-up improvements around plugin detail presentation, structured logging, configuration extension namespacing, cache and permission comment conformance, and OpenSpec language governance.

## What Changed

- Registered built-in runtime configuration parameters and protected public frontend settings in the host config layer.
- Added runtime validation, immutable key protection, import safeguards, and SQL seed metadata for protected system parameters.
- Made JWT expiry, session timeout, upload size, and login IP blacklist read through the host runtime configuration service.
- Added process-local snapshot caching plus shared revision synchronization for configuration and permission hot paths, with single-node and cluster-aware strategies.
- Added a public frontend configuration endpoint so the login page and admin workspace can safely consume branding and theme settings.
- Added plugin detail dialog and host-service presentation refinements for plugin governance in the default admin workspace.
- Added structured logging switch support and unified HTTP server logs with business logs.
- Moved host-specific `server` and `logger` extensions under explicit `extensions` namespaces instead of borrowing GoFrame-owned config fields.
- Replaced direct reliance on GoFrame’s built-in `/api.json` output with a host-managed OpenAPI builder.
- Added source-plugin route ownership capture through a host-observable `RouteGroup` facade without requiring route prefixes or duplicate route declarations in `plugin.yaml`.
- Projected enabled source-plugin routes and enabled dynamic-plugin routes into the host-managed API document while excluding internal non-business routes.
- Completed repository-wide backend comment conformance cleanup for the affected host services, plugin integration components, and plugin backend samples.

## Capabilities

### Modified Capabilities

- `config-management`
- `online-user`
- `user-auth`
- `plugin-ui-integration`
- `spec-governance`
- `system-api-docs`
- `plugin-runtime-loading`
- `plugin-manifest-lifecycle`

## Impact

- Host backend services under `apps/lina-core/internal/service/config`, `auth`, `session`, `role`, `cron`, `plugin`, and related packages.
- Host HTTP boot and API document generation under `apps/lina-core/internal/cmd` and `apps/lina-core/internal/service/apidoc`.
- Source plugin registration contracts under `apps/lina-core/pkg/pluginhost`.
- Shared plugin bridge and plugin database support packages.
- Default admin workspace behavior for plugin details, public frontend settings, and plugin-facing regression coverage.
- OpenSpec workflow governance, especially archived artifact language requirements.

## Merged Change Record

This archive intentionally merges the following completed active changes into one archived iteration:

- `config-runtime-params`
- `host-managed-plugin-openapi`

The merged archive is used as the single historical record for the completed work, and both source active directories are removed after archival so the repository returns to a clean no-active-change state.
