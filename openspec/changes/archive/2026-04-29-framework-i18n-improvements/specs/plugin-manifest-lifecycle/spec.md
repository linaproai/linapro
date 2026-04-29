## ADDED Requirements

### Requirement: Plugin manifest and lifecycle must support zero-code extension when adding new languages
The plugin manifest and lifecycle SHALL automatically cover new language runtime UI translation resources and apidoc translation resources when the host adds a new built-in language, without modifying host code or individual plugin source code. Source plugins SHALL append that language's resources in their own `manifest/i18n/<locale>/*.json` and `manifest/i18n/<locale>/apidoc/**/*.json`; dynamic plugins SHALL write that language's resources into release custom sections during the packaging phase; the host automatically discovers, loads, and cleans up these new language resources following existing rules during loading, enable/disable, upgrade, and uninstall flows.

#### Scenario: Plugin resources auto-integrate after enabling Traditional Chinese
- **WHEN** the host enables `zh-TW` as a built-in language
- **AND** the source plugin `apps/lina-plugins/<plugin-id>/manifest/i18n/zh-TW/*.json` exists
- **THEN** when the plugin is enabled, its `zh-TW` translation resources automatically join the runtime translation aggregation result
- **AND** when the plugin is disabled or uninstalled, `zh-TW` translation resources are synchronously removed from the aggregation result
- **AND** the entire flow requires no host code modifications and no other plugin code modifications

#### Scenario: Dynamic plugins carry Traditional Chinese resources via release
- **WHEN** a dynamic plugin adds `manifest/i18n/zh-TW/*.json` in a new version and repackages
- **AND** the host upgrades the plugin to the new release
- **THEN** after upgrade, the Traditional Chinese translation resources take effect, and the old version's resources are no longer used
- **AND** during upgrade, the host only clears sector caches related to that plugin, without affecting other languages or other plugins
