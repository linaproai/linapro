## Why

The host already supports auto-enabling source plugins during startup through `plugin.autoEnable`, but it still lacks a data-protection capability tailored for demo environments. A demo environment usually needs to preserve full browse/query behavior while preventing accidental or malicious writes to system data. We therefore need a source plugin whose behavior is governed directly by plugin enablement state and that can enforce read-only protection consistently in the request pipeline.

## What Changes

- Add an official source plugin named `demo-control` and connect it to the full-system request-governance chain through the host's published global HTTP middleware registration seam.
- When `demo-control` is enabled by the host, intercept write requests across `/*` by following RESTful `HTTP Method` semantics and keep only query-style requests available.
- Preserve a minimal whitelist for required session entry points such as login and logout so the demo environment remains usable.
- Use `plugin.autoEnable` in the host main config file as the switch that decides whether `demo-control` is auto-enabled during startup.

## Capabilities

### New Capabilities
- `demo-control-guard`: Define a demo-control source plugin governed by `plugin.autoEnable`, together with global write-interception rules based on `HTTP Method` semantics.

### Modified Capabilities

## Impact

- The affected code is mainly under `apps/lina-core/manifest/config/` and `apps/lina-plugins/`.
- A new source-plugin directory `apps/lina-plugins/demo-control/` is required, together with updates to the plugin workspace wiring entry and `go.work`.
- `plugin.autoEnable` config tests and demo-control middleware unit tests must cover the default-disabled case, explicit enablement, login whitelist behavior, and write interception.
