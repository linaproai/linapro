# Runtime I18N Sample

This directory stores the delivery locale bundles for the `plugin-demo-dynamic` sample plugin.

The host snapshots `manifest/i18n/<locale>.json` into the dynamic-plugin artifact so the runtime i18n API can merge plugin-owned messages after install, enable, upgrade, disable, and uninstall actions.

Included keys cover:

- plugin metadata such as `plugin.plugin-demo-dynamic.name`
- menu metadata such as `menu.plugin:plugin-demo-dynamic:main-entry.title`
- embedded page copy such as `plugin.plugin-demo-dynamic.page.*`

Keep keys flat and use canonical locale filenames like `zh-CN.json` and `en-US.json`.
