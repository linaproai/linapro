## Why

The system configuration management layer needed several coordinated improvements that were completed across closely related implementation windows:

1. The host consumed settings such as JWT expiry, session timeout, upload limits, and login IP blacklists, but the parameter-management layer did not provide a clear protected registry, runtime-safe validation, or multi-instance-friendly cache behavior. Runtime configuration existed as editable key-value records but did not reliably drive host behavior.

2. GoFrame's built-in `/api.json` output could not distinguish source-plugin routes from host routes, could not project source-plugin routes by enablement state, and would have required a duplicate route declaration model if the project had tried to solve the problem through `plugin.yaml`.

3. The default upload size limit was inconsistent across the host initialization SQL, the config template, and the backend static fallback value. The repository carried both a 10 MB and a 16 MB baseline, causing new environments, default runtime paths, and upload error messages to behave differently depending on which path was used.

4. The config-management component had unit-test coverage of only 71.9%, well below the repository's 80% target. Without stronger automated regression protection, later refactors or new config keys could easily introduce subtle regressions.

## What Changes

- Registered built-in runtime configuration parameters and protected public frontend settings in the host config layer.
- Added runtime validation, immutable key protection, import safeguards, and SQL seed metadata for protected system parameters.
- Made JWT expiry, session timeout, upload size, and login IP blacklist read through the host runtime configuration service.
- Added process-local snapshot caching plus shared revision synchronization for configuration and permission hot paths, with single-node and cluster-aware strategies.
- Added a public frontend configuration endpoint so the login page and admin workspace can safely consume branding and theme settings.
- Replaced direct reliance on GoFrame's built-in `/api.json` output with a host-managed OpenAPI builder.
- Added source-plugin route ownership capture through a host-observable `RouteGroup` facade without requiring route prefixes or duplicate route declarations in `plugin.yaml`.
- Projected enabled source-plugin routes and enabled dynamic-plugin routes into the host-managed API document while excluding internal non-business routes.
- Unified the host upload-size default at 20 MB across SQL seed, config template, and static fallback.
- Updated upload-limit validation and friendly error-message tests so the default 20 MB baseline behaves consistently across file upload and transport-limit enforcement.
- Added unit tests for the config-management component to cover plugin dynamic storage paths, protected public-frontend config helpers, runtime-parameter snapshot caching, and the clustered revision controller, bringing coverage from 71.9% to 83.0%.
- Added plugin detail dialog and host-service presentation refinements for plugin governance in the default admin workspace.
- Added structured logging switch support and unified HTTP server logs with business logs.
- Moved host-specific `server` and `logger` extensions under explicit `extensions` namespaces instead of borrowing GoFrame-owned config fields.
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

- Host backend services under `apps/lina-core/internal/service/config`, `auth`, `session`, `role`, `cron`, `plugin`, `file`, and related packages.
- Host HTTP boot and API document generation under `apps/lina-core/internal/cmd` and `apps/lina-core/internal/service/apidoc`.
- Source plugin registration contracts under `apps/lina-core/pkg/pluginhost`.
- Shared plugin bridge and plugin database support packages.
- Default admin workspace behavior for plugin details, public frontend settings, and plugin-facing regression coverage.
- Host initialization SQL, config templates, and upload-config fallback logic.
- OpenSpec workflow governance, especially archived artifact language requirements.
