# Lina Plugin System

`apps/lina-plugins/` is the primary reference entry for Lina's plugin system. It documents the conventions that are implemented in this repository today and points to the sample plugins that contributors should follow.

## Current Scope

Lina currently supports two plugin shapes:

- `source`: source plugins that live under `apps/lina-plugins/<plugin-id>/` and are compiled together with the host.
- `dynamic`: managed WASM plugins that ship as runtime artifacts with manifest, frontend assets, SQL assets, and governed host-service access.

The repository already includes:

- plugin discovery and governance metadata
- plugin pages and plugin slots
- plugin management flows
- install and uninstall SQL conventions
- sample source and dynamic plugins

## Directory Layout

```text
apps/lina-plugins/
  lina-plugins.go/        Explicit source-plugin wiring entry in the host build
  plugin-demo-source/     Source plugin sample
  plugin-demo-dynamic/    Dynamic WASM plugin sample
  OPERATIONS.md           Operations and review notes
```

## Design Principles

- **Convention over configuration**: pages, slots, and SQL assets are discovered from stable directory conventions.
- **Single source of truth**: plugin metadata lives in `plugin.yaml`; implementation details live in real source files.
- **Explicit wiring**: source plugins are wired explicitly so the host build graph remains visible and reviewable.
- **Governance first**: dynamic plugins are installed and executed through host-controlled contracts rather than unrestricted runtime code execution.

## Plugin Types

### Source Plugins

Source plugins are compiled into the host. Use them when the plugin should ship together with the repository and follow the same delivery pipeline as the host application.

### Dynamic Plugins

Dynamic plugins are built as governed WASM artifacts. Use them when the plugin lifecycle must be managed at runtime through upload, install, enable, disable, uninstall, and release reconciliation flows.

## Host Service Model for Dynamic Plugins

Dynamic plugins request access to host capabilities through structured `hostServices` declarations in `plugin.yaml`.

Current host service groups are:

- `runtime`
- `storage`
- `network`
- `data`

Each service is authorized by the host and constrained by declared methods and governed resources.

## Typical Development Flow

### Source Plugin Flow

1. Create `apps/lina-plugins/<plugin-id>/`.
2. Define metadata in `plugin.yaml`.
3. Add backend, frontend, and optional SQL assets.
4. Wire the plugin explicitly through the source-plugin registration entry.

### Dynamic Plugin Flow

1. Create the plugin source tree.
2. Embed manifest and static resources.
3. Build the runtime artifact with `make wasm`.
4. Upload or place the artifact in the configured storage path.
5. Install and manage it through the host lifecycle flows.

## References

- `plugin-demo-source/README.md`
- `plugin-demo-dynamic/README.md`
- `OPERATIONS.md`
