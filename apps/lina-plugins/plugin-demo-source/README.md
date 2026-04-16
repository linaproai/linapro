# plugin-demo-source

`plugin-demo-source` is the source-plugin sample for Lina. It shows the smallest flow for a plugin that is developed in-repo and compiled into the host.

## Directory Layout

```text
plugin-demo-source/
  plugin.yaml
  backend/
  frontend/
  manifest/
```

## Manifest Scope

`plugin.yaml` keeps the plugin metadata and menu declarations. Pages, slots, and SQL assets follow directory conventions instead of being duplicated in metadata.

## Backend Integration

- implement backend entry points under `backend/`
- keep service logic under `backend/internal/service/`
- wire the plugin explicitly through the source-plugin registration entry used by the host build

## Front-end Integration

Front-end pages are discovered from the plugin's `frontend/` directory according to repository conventions.

## SQL Conventions

- install SQL lives under `manifest/sql/`
- uninstall SQL lives under `manifest/sql/uninstall/`

## Review Checklist

- metadata stays minimal and accurate
- host wiring remains explicit
- pages follow directory conventions
- plugin-owned SQL is kept separate from host SQL
