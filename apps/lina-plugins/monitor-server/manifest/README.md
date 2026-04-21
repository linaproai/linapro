# Manifest Resources

This directory stores the install and uninstall SQL assets for `monitor-server`.

## Contents

- `sql/001-monitor-server-schema.sql`: creates the server-monitor snapshot table
- `sql/uninstall/001-monitor-server-schema.sql`: drops the server-monitor snapshot table during uninstall purge
