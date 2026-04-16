# plugin-demo-dynamic

`plugin-demo-dynamic` is the dynamic WASM plugin sample for Lina. It demonstrates the smallest end-to-end path for a governed runtime plugin.

## What It Demonstrates

- one menu entry rendered inside the default management workspace
- one standalone static page that does not depend on the host UI framework
- demo backend routes executed through the dynamic plugin bridge
- governed access to `runtime`, `storage`, `network`, and `data` host services

## Directory Layout

```text
plugin-demo-dynamic/
  main.go
  plugin_embed.go
  plugin.yaml
  backend/
  frontend/
  manifest/
```

## Build

Build all dynamic plugin artifacts:

```bash
make wasm
```

Build only this sample:

```bash
make wasm p=plugin-demo-dynamic
```

The runtime artifact is written to `temp/output/plugin-demo-dynamic.wasm`.

## Backend Contract

The sample exposes governed routes under the dynamic plugin public prefix and keeps business logic inside `backend/internal/service/`.

## Host Services

The sample requests the following host services in `plugin.yaml`:

- `runtime`
- `storage`
- `network`
- `data`

These declarations are reviewed and authorized by the host during plugin lifecycle operations.

## Review Checklist

- `plugin.yaml` keeps metadata and host-service declarations clear.
- frontend assets match the declared plugin access mode.
- the built WASM artifact is reproducible from the source tree.
- backend logic stays inside service components instead of controllers.
