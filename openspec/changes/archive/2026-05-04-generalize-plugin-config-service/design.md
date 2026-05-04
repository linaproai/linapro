## Context

`apps/lina-core/pkg/pluginservice/config` is the public host configuration service entrypoint published for source plugins. Before this change, the component exposed the `monitor-server` plugin's configuration shape through a `MonitorConfig` type alias and a `GetMonitor()` method, which made the host public component own plugin business configuration schema.

That conflicted with the intended plugin boundary. Source plugins may reuse host capabilities, but plugin business configuration structures, defaults, and validation logic should belong to the plugin itself. The host public component should provide stable read access and should not keep adding plugin-specific `GetXxx()` methods as the number of plugins grows.

This change does not alter the overall configuration source. GoFrame's configuration system remains the authoritative data source, and no runtime configuration write capability is added.

## Goals / Non-Goals

**Goals:**

- Design `pluginservice/config` as a business-neutral general read-only configuration accessor.
- Allow source plugins to read arbitrary configuration keys instead of limiting reads to a plugin-specific prefix.
- Allow dynamic plugins to read the complete static configuration by declaring the `config` host service.
- Support struct scanning, raw value reads, existence checks, basic type reads, and `time.Duration` parsing.
- Move plugin-specific configuration structures, defaults, validation, and business semantics into the plugin.
- Remove the `MonitorConfig` alias and `GetMonitor()` plugin-specific method from the public component.
- Add unit coverage for configuration read errors, missing keys, and duration parsing.

**Non-Goals:**

- Do not provide configuration write, save, hot reload, or runtime configuration management capabilities.
- Do not change the database-backed system configuration management module.
- Do not introduce a new configuration file format, configuration center, or external dependency.
- Do not migrate the business meaning of existing configuration keys. For example, `monitor.interval` can still be read by the `monitor-server` plugin through the generic reader.

## Decisions

### Decision 1: Use Generic Key Access Instead of Business Methods

`pluginservice/config.Service` exposes these generic methods:

- `Get(ctx, key)`: read a raw GoFrame configuration value.
- `Exists(ctx, key)`: report whether a key exists.
- `Scan(ctx, key, target)`: scan a configuration section into a caller-provided struct.
- `String/Bool/Int/Duration(ctx, key, defaultValue)`: read basic types with default-value support.

Each plugin maintains its own `Config` structure and `Load(ctx)` method. For example, `monitor-server` can scan the `monitor` section inside the plugin, then read `monitor.interval` as a `time.Duration` and apply whole-second alignment validation.

Rejected alternative: keep adding `GetXxx()` methods to the public component for each plugin. That would make the host public component depend on plugin business schemas and grow with every plugin, so it is not used.

### Decision 2: Allow Arbitrary Key Reads While Keeping the Trusted Read-Only Boundary

Source plugins are trusted extensions built in the same process and repository as the host. The configuration service does not add prefix restrictions to keys, so a plugin can read the full configuration file.

Rejected alternative: force reads under a `plugins.<plugin-id>` prefix. That would provide stronger isolation, but it does not match this requirement and would prevent source plugins from reusing existing host configuration. Governance is handled through the read-only boundary, code review, and dynamic plugin host service authorization.

### Decision 3: Public Service Parses Duration, Plugins Own Business Validation

The public service parses configuration strings into `time.Duration` and keeps default-value semantics stable. Business constraints such as "must be greater than 0", "must be at least 1 second", and "must align to whole seconds" are validated by the plugin or caller in its own configuration loading method.

Rejected alternative: embed all duration validation strategies in the public service. That would mix business rules into a generic component and could not cover different configuration item requirements, so it is not used.

### Decision 4: Return Errors and Let Plugin Call Sites Choose the Failure Strategy

Generic read methods return `error` and do not directly `panic`. Plugin startup or cron registration paths can choose fail-fast behavior, while normal business paths can wrap errors as caller-visible business errors or internal errors.

Rejected alternative: reuse the host internal configuration service style that directly panics through helpers such as `mustScanConfig`. That style fits static host startup configuration loading, but it is too forceful for a generic plugin public interface.

### Decision 5: Dynamic Plugins Read Full Static Configuration Through the Config Host Service

Dynamic plugins cannot import `pkg/pluginservice/config` directly, so this change adds the `config` host service through `lina_env.host_call`. A dynamic plugin declares:

```yaml
hostServices:
  - service: config
```

This declaration grants the dynamic plugin the ability to read the complete static configuration. It no longer requires a resource allowlist of configuration keys. `methods` may be omitted; omission grants the current complete read-only config method set: `get`, `exists`, `string`, `bool`, `int`, and `duration`. A plugin that wants a narrower grant can explicitly declare a subset:

```yaml
hostServices:
  - service: config
    methods: [get, exists, string, bool, int, duration]
```

The request payload carries the key. `get` returns the configuration value as JSON, and an empty key or `.` returns the complete static configuration snapshot exposed by the GoFrame configuration system. `exists` returns a found flag. `string`, `bool`, `int`, and `duration` return string representations of their respective types. The wasip1 guest SDK helpers `Exists`, `String`, `Bool`, `Int`, and `Duration` call the corresponding host service methods directly, and dynamic plugins that manually build `host_call` requests can call the same read-only methods. The service does not provide write, save, hot reload, or runtime configuration management.

Rejected alternative: require dynamic plugins to declare each readable key or key pattern. That provides stronger isolation, but it does not match the current goal that dynamic plugins can also read the full configuration, so it is not used.

## Risks / Trade-offs

- Full configuration reads may expose sensitive configuration to source plugins and dynamic plugins. Dynamic plugins must declare the capability through `hostServices`; the grant is captured in the installation/enablement authorization snapshot and host_call audit path, and the service remains read-only.
- Removing `GetMonitor()` is a breaking API change. The project has no legacy compatibility constraint, so the `monitor-server` call sites and tests are migrated in the same change.
- Moving business validation into plugins may create inconsistent validation style. The `monitor-server` migration establishes a plugin-local configuration loading pattern that later plugins can reuse.
- The generic configuration service does not perform runtime cache invalidation. This change reads only static configuration files and adds no mutable cache. If runtime dynamic configuration support is added later, it must use cluster-mode revision numbers, broadcast invalidation, shared cache, or an equivalent mechanism.

## Migration Plan

1. Replace the public interface in `pkg/pluginservice/config` with generic read-only configuration access methods.
2. Remove the `MonitorConfig` type alias and `GetMonitor()` method.
3. Add private configuration loading logic inside the `monitor-server` plugin to read the existing `monitor` section and apply defaults, duration parsing, and business validation.
4. Update `monitor-server` cron registration and cleanup logic to use the plugin-local configuration loader.
5. Add or update unit tests for the generic configuration service and `monitor-server` configuration loading.
6. Add the dynamic plugin `config` host service constants, capability derivation, codec, guest helpers, and host dispatcher.
7. Run affected Go tests and the full backend test suite if needed.

## Open Questions

- None.
