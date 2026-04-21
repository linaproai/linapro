# Manifest Resources

This directory stores the install and uninstall SQL assets for `content-notice`.

## Contents

- `sql/001-content-notice-schema.sql`: creates the notice table and seeds notice dictionaries
- `sql/uninstall/001-content-notice-schema.sql`: removes notice dictionaries and drops the notice table during uninstall purge
