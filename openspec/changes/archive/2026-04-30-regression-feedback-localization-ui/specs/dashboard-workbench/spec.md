## ADDED Requirements

### Requirement: Workbench page copy must use runtime i18n resources

The workbench SHALL load delivered page content through runtime i18n resources so English environments do not display Chinese default copy.

#### Scenario: Workbench displays English default copy
- **WHEN** an administrator opens the workbench in `en-US`
- **THEN** titles, metrics, shortcuts, project cards, activities, and todos use English runtime copy

### Requirement: Management workbench logo must provide subtle dark-mode edge depth

The management workbench SHALL add a very subtle cyan edge glow to the brand logo icon in dark mode without affecting layout.

#### Scenario: Dark-mode logo edge glow remains subtle and layout-safe
- **WHEN** an administrator opens the workbench in dark mode
- **THEN** the logo icon edge shows a subtle cyan glow and navigation layout does not shift

### Requirement: User theme preference must override the public frontend default

The workbench SHALL preserve a user-explicit theme preference when synchronizing public frontend config.

#### Scenario: Existing user theme preference survives page refresh
- **WHEN** the user refreshes after explicitly selecting a theme
- **THEN** public frontend defaults do not overwrite that preference
