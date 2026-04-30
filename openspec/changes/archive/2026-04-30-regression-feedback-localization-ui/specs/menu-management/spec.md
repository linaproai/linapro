## ADDED Requirements

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
