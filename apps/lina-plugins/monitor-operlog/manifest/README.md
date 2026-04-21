# Manifest Resources

This directory stores the install and uninstall SQL assets for `monitor-operlog`.

## Contents

- `sql/001-monitor-operlog-schema.sql`: creates the operation-log table and seeds operation dictionaries
- `sql/uninstall/001-monitor-operlog-schema.sql`: removes operation dictionaries and drops the table during uninstall purge
