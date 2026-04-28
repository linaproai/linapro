# plugin-demo-source

`plugin-demo-source` is the source-plugin sample for `LinaPro`. It demonstrates an in-repo plugin that is discovered by the host, installed explicitly from plugin management, and mounted into the default workspace.

## Directory Layout

```text
plugin-demo-source/
  plugin.yaml
  plugin_embed.go
  backend/
    api/
    internal/
      controller/
      service/
      dao/
      model/do/
      model/entity/
    hack/config.yaml
    plugin.go
  frontend/
    pages/
  manifest/
    sql/
    sql/uninstall/
```

## What This Sample Demonstrates

- install SQL under `manifest/sql/` creates the plugin-owned table `plugin_demo_source_record`
- mock SQL under `manifest/sql/mock-data/` provides optional demo records for local sample data
- uninstall SQL under `manifest/sql/uninstall/` drops the plugin-owned table when the user confirms storage purge
- the sample page in `frontend/pages/sidebar-entry.vue` performs CRUD against the plugin-owned table and supports attachment upload/download
- plugin-owned attachment files are stored under the host upload root in the `plugin-demo-source/` namespace
- disabling the plugin hides menus and routes but keeps table data and stored files
- uninstalling the plugin opens a confirmation dialog that lets the user choose whether to purge plugin-owned table data and stored files

## Manifest Scope

`plugin.yaml` keeps the plugin metadata, menu declarations, and button permissions. Pages and SQL assets still follow directory conventions instead of being duplicated in metadata.

## Backend Integration

- implement backend entry points under `backend/`
- keep service logic under `backend/internal/service/`
- keep plugin-local ORM codegen output under `backend/internal/dao` and `backend/internal/model/{do,entity}` when the plugin accesses database tables
- register install, enable, disable, and uninstall lifecycle hooks through the source-plugin registration entry used by the host build
- keep plugin-owned cleanup logic in the plugin service so uninstall can optionally purge files before uninstall SQL drops the table

## Front-end Integration

- front-end pages are discovered from the plugin's `frontend/` directory according to repository conventions
- the sample page keeps the existing summary card and adds a `VXE` grid plus modal form for record CRUD
- the uninstall choice is surfaced by the host plugin-management page, not by the plugin page itself

## SQL Conventions

- install SQL lives under `manifest/sql/`
- uninstall SQL lives under `manifest/sql/uninstall/`
- install SQL should be idempotent so reinstall after a non-purge uninstall can preserve data
- uninstall SQL should be paired with plugin cleanup hooks when stored files must be removed together with table data

## Review Checklist

- metadata stays minimal and accurate
- host wiring remains explicit
- pages follow directory conventions
- plugin-owned SQL is kept separate from host SQL
- disable keeps plugin-owned data intact
- uninstall supports both retain-data and purge-data paths
