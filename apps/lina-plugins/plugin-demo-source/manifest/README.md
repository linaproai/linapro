# Manifest Resources

This directory stores plugin-owned `SQL` lifecycle assets for the source-plugin sample.

## Contents

- `sql/001-plugin-demo-source-records.sql`: creates the plugin-owned demo table during install
- `sql/mock-data/001-plugin-demo-source-mock-data.sql`: provides optional local demo records
- `sql/uninstall/001-plugin-demo-source-records.sql`: drops the plugin-owned demo table during uninstall when the user chooses to purge storage data

Menus stay in `plugin.yaml`; `SQL` assets are reserved for plugin-owned data lifecycle changes.
