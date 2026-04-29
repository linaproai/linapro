## ADDED Requirements

### Requirement: Business modules must own their own localization projection rules
The host system SHALL let business modules such as menu, dictionary, config, scheduled tasks, roles, and plugin runtime maintain localization projection rules within their own module boundaries. `internal/service/i18n` SHALL only provide foundational capabilities such as language resolution, translation lookup, resource loading, caching, and missing checks, and MUST NOT reference business entities, business protection rules, or business translation key derivation logic. Business modules can depend on underlying capabilities like `ResolveLocale`, `Translate`, `TranslateSourceText` through narrow interfaces, but MUST NOT require the i18n foundation service to reverse-know business modules.

#### Scenario: Business modules complete projection within their own boundaries
- **WHEN** `menu` / `dict` / `sysconfig` / `jobmgmt` / `role` / `plugin runtime` and other modules need to localize query results for display
- **THEN** the module derives translation keys, determines whether to skip default language, and determines whether records are protected built-in records in its own `*_i18n.go` or equivalent module
- **AND** `internal/service/i18n` does not import these business modules or business entities

#### Scenario: i18n foundation service only provides underlying capabilities
- **WHEN** a business module needs to translate a display field
- **THEN** the business module calls `Translate` / `TranslateSourceText` and other underlying methods with its own key and fallback
- **AND** the i18n service does not expose business-entity-named methods like `ProjectMenu`, `ProjectDictType`, `ProjectBuiltinJob`

### Requirement: Judgment rules for protected built-in records must remain in the owning business module
Business modules SHALL maintain judgment rules and translation key conventions for protected built-in records in their own services. For example: the display projection of the built-in `admin` role is maintained by the `role` module, the display projection of default task groups and built-in tasks is maintained by the `jobmgmt` module, and the display projection of plugin metadata is maintained by the plugin runtime module. The generic i18n foundation service must not contain these business determination constants.

#### Scenario: Built-in admin role displays localized name in English environment read-only list
- **WHEN** an administrator queries the role list in the `en-US` environment
- **THEN** the `role` module only provides localized display for the built-in `admin` role via the `role.builtin.admin.name` translation key
- **AND** other editable roles' `name` fields retain their original database values

#### Scenario: User-created custom tasks retain original values
- **WHEN** an administrator queries the scheduled task list in any non-default language
- **THEN** the `jobmgmt` module does not apply translation projection to user-created tasks where `is_builtin = 0`
- **AND** the task's `name` / `description` display their original database values
