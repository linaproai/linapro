# Manifest Resources

This directory stores the install and uninstall SQL assets for `org-management`.

## Contents

- `sql/001-org-management-schema.sql`: creates department, post, user-dept, and user-post tables
- `sql/uninstall/001-org-management-schema.sql`: drops the plugin-owned organization tables during uninstall purge
