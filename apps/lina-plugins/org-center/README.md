# org-center

`org-center` is the official LinaPro source plugin for organization management.

## Scope

This plugin owns:

- department management
- post management
- the optional organization capability consumed by host user management

## Host Boundary

The host keeps user management, authentication, and menu governance. This plugin adds organization-specific APIs, menus, and storage tables under `manifest/sql/`.

## Directory Layout

```text
org-center/
  plugin.yaml
  backend/
  frontend/
  manifest/
```
