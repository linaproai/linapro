# Design

## Localization Sweep

Framework-delivered pages are checked in `en-US` for Chinese system copy, seed display leakage, localized role names, generated department nodes, built-in config metadata, and public workbench content.

## UI Layout

Forms and tables with long English labels receive layout adjustments so labels and critical table columns remain readable. Playwright screenshots cover representative English pages.

## Menu Governance

Dynamic plugin route permission buttons mount under the owning plugin menu, and menu tree rows expose direct title-click expansion with pointer affordance.

## Operational Safety

Scheduled-job Run Now requires confirmation before calling the trigger API. Demo-control blocks plugin governance writes when enabled. Built-in dictionaries and system parameters are editable but protected from deletion.

## Workbench

The dashboard workbench uses runtime i18n, local logo assets, LinaPro-specific quick links, project cards, activities, and todos. User theme preference takes precedence over public frontend defaults.

## Tests

Playwright tests cover English workbench, role/user seed display, dynamic plugin menu tree, organization and dictionary layout, service monitor disk table, job trigger confirmation, and follow-up feedback cases.
