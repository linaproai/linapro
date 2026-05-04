## Why

`apps/lina-core/pkg/pluginservice/config` had started exposing plugin-specific strongly typed configuration through `GetMonitor()`. That meant each new plugin or plugin configuration shape would require another change to a host public component. Plugins need to read different parts of the configuration file, but the configuration service itself should remain business-neutral and provide only stable, general, read-only access.

## What Changes

- Add a general read-only configuration service for source plugins so plugins can read, scan, and parse configuration content by arbitrary key.
- Remove direct exposure of plugin business configuration structures from `pluginservice/config`, including plugin-specific APIs such as `MonitorConfig` and `GetMonitor()`.
- Keep each plugin's configuration structure, defaults, validation, and business meaning inside that plugin. The host public component only handles generic reads and basic type parsing.
- Keep configuration access read-only. The plugin configuration service does not provide write, save, or runtime mutation methods.
- Document the trust boundary clearly: source plugins can read the full host configuration, while dynamic or third-party plugins must use host service authorization and auditing before reusing this capability.
- **BREAKING**: source plugins no longer call `configsvc.New().GetMonitor(ctx)` to obtain monitor settings. They must read the relevant keys through the generic accessor and perform structured parsing inside the plugin.

## Capabilities

### New Capabilities

- `plugin-config-service`: defines the general read-only configuration access available to plugins, including arbitrary key reads, section scanning, basic type parsing, and `time.Duration` parsing.

### Modified Capabilities

- None.

## Impact

- Changes the public API design and implementation of `apps/lina-core/pkg/pluginservice/config`.
- Affects source plugins that depended on `pluginservice/config.GetMonitor()`, primarily `apps/lina-plugins/monitor-server`.
- Requires plugin configuration service unit tests for arbitrary key reads, struct scanning, defaults, missing keys, duration parsing, and error handling.
- No database change is required, so no SQL file is added.
- No frontend UI text, menu, route, page interaction, runtime i18n, manifest i18n, or apidoc i18n resources are affected.
- No runtime mutable configuration cache is added. Static configuration reads continue to reuse the host's existing process-local static configuration access and do not introduce distributed cache consistency concerns.
