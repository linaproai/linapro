# Runtime I18N Sample

This directory stores the delivery locale bundles for the `plugin-demo-dynamic` sample plugin.

The host snapshots direct JSON files under `manifest/i18n/<locale>/` into one per-locale dynamic-plugin artifact asset so the runtime i18n API can merge plugin-owned messages after install, enable, upgrade, disable, and uninstall actions.

API-documentation translations for the plugin live under `manifest/i18n/<locale>/apidoc/**/*.json`. They are embedded into the dynamic-plugin artifact separately from runtime UI messages and are only merged when `/api.json` is rendered.

Included normalized keys cover:

- plugin metadata such as `plugin.plugin-demo-dynamic.name`
- menu metadata such as `menu.plugin:plugin-demo-dynamic:main-entry.title`
- embedded and standalone page copy such as `plugin.plugin-demo-dynamic.page.*`
- API-documentation metadata under `plugins.plugin_demo_dynamic.*` in `<locale>/apidoc/`

Runtime UI message files may use nested JSON or flat dotted keys. The host normalizes both forms into flat keys for aggregation and diagnostics, then returns nested objects to the frontend runtime.

API-documentation message files follow the same nested-or-flat authoring rule and normalize to stable `plugins.plugin_demo_dynamic.*` keys. Repeated standard response metadata is supplied by host-owned `core.common.*` fallback keys.

Use canonical locale directory names like `zh-CN` and `en-US`, and semantic runtime filenames such as `plugin.json`, `menu.json`, and `job.json`.
