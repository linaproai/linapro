## ADDED Requirements

### Requirement: Menu capability must return localized titles for the current language
The system SHALL return localized menu titles in menu trees, parent menu display, role menu trees, and dynamic route projections according to the current request language. Menu localization MUST derive translation keys from stable `menu_key` anchors. When the current language lacks a translation, the system MUST fall back to the default language or the menu default name. The directly editable `name` field in menu edit forms MUST keep the database value to avoid writing display titles back into governance data during language switching.

#### Scenario: Menu tree returns English titles
- **WHEN** a user requests the menu tree or current-user dynamic routes with `en-US`
- **THEN** menu titles in the response use localized values for that language
- **AND** the same menu remains consistent between tree node names and route `meta.title`
- **AND** English menu titles use concise, natural product wording instead of literal translations from Chinese names

#### Scenario: Button titles under resource menus use short action words
- **WHEN** a user views button menus in the menu management page, parent menu tree, or role menu tree with `en-US`
- **THEN** button titles under resource menus use short action words such as `Query`, `Create`, `Update`, `Delete`, and `Export`
- **AND** button titles avoid repeating the parent resource menu name, such as `Users / Create` instead of `Users / Create User`
- **AND** buttons that cannot be clearly expressed by generic actions may keep short business actions such as `Reset Password`, `Force Logout`, or `Run Now`

#### Scenario: Missing menu translations fall back to default names
- **WHEN** a menu has no translation for the current language
- **THEN** the system falls back to the default-language title or menu default name
- **AND** menu structure, permissions, and sort order remain unaffected

#### Scenario: Administrator edits menus in English
- **WHEN** an administrator opens menu detail or edit drawer in an `en-US` environment
- **THEN** the menu list tree, parent menu display name, and parent menu selector tree show localized titles for the current language
- **AND** the editable `name` field in the form keeps the original database value

### Requirement: Menu management must support stable business keys for i18n copy
The system SHALL preserve stable business keys as i18n anchors in menu governance and allow host and plugin resources to maintain menu titles through unified translation resources instead of hard-coding the same menu copy in multiple pages.

#### Scenario: Plugin menus integrate with i18n resources
- **WHEN** a plugin declares a menu with a stable `menu_key` and provides matching translation resources
- **THEN** the host uses the menu's localized title in the menu management page, left navigation, and role authorization tree
- **AND** administrators do not need to configure multiple translation mappings for the same plugin menu

#### Scenario: Administrator views menu detail
- **WHEN** an administrator views a menu detail in the current language
- **THEN** the system returns the localized parent menu display title and related read-only copy for the current language
- **AND** the menu's editable fields continue to keep database values while the stable business key remains available for later translation maintenance and diagnostics
