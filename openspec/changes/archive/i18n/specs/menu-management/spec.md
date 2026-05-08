## ADDED Requirements

### Requirement: Menu capability must return localized titles for the current language
The system SHALL return localized menu titles in menu trees, parent menu display, role menu trees, and dynamic route projections according to the current request language. Menu localization MUST derive translation keys from stable `menu_key` anchors. When the current language lacks a translation, the system MUST fall back to the default language or the menu default name. The directly editable `name` field in menu edit forms MUST keep the database value.

#### Scenario: Menu tree returns English titles
- **WHEN** a user requests the menu tree or current-user dynamic routes with `en-US`
- **THEN** menu titles in the response use localized values for that language
- **AND** the same menu remains consistent between tree node names and route `meta.title`
- **AND** English menu titles use concise, natural product wording instead of literal translations

#### Scenario: Button titles under resource menus use short action words
- **WHEN** a user views button menus in the menu management page with `en-US`
- **THEN** button titles use short action words such as `Query`, `Create`, `Update`, `Delete`, and `Export`
- **AND** button titles avoid repeating the parent resource menu name

#### Scenario: Administrator edits menus in English
- **WHEN** an administrator opens menu detail or edit drawer in an `en-US` environment
- **THEN** the menu list tree, parent menu display, and parent menu selector tree show localized titles
- **AND** the editable `name` field in the form keeps the original database value

### Requirement: Dynamic route permission buttons must mount under their owning plugin menu

Button permission menus generated from dynamic plugin route declarations SHALL mount under the owning dynamic plugin page menu or plugin root menu.

#### Scenario: Dynamic plugin route buttons are children of plugin menu
- **WHEN** dynamic route permissions are synchronized
- **THEN** corresponding button permissions appear under the owning plugin menu

### Requirement: Menu tree expandable rows must be clickable

Menu management tree rows that can expand SHALL show a clickable pointer and allow clicking the node title area to expand or collapse.

#### Scenario: Expandable menu row pointer and click
- **WHEN** an administrator hovers and clicks an expandable row title
- **THEN** the node expands or collapses

### Requirement: Menu management must support stable business keys for i18n copy
The system SHALL preserve stable business keys as i18n anchors in menu governance and allow host and plugin resources to maintain menu titles through unified translation resources.

#### Scenario: Plugin menus integrate with i18n resources
- **WHEN** a plugin declares a menu with a stable `menu_key` and provides matching translation resources
- **THEN** the host uses the menu's localized title in the menu management page, left navigation, and role authorization tree
