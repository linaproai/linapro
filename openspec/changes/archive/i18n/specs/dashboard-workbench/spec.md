## ADDED Requirements

### Requirement: Workbench page copy must use runtime i18n resources

The workbench SHALL load delivered page content through runtime i18n resources so English environments do not display Chinese default copy.

#### Scenario: Workbench displays English default copy
- **WHEN** an administrator opens the workbench in `en-US`
- **THEN** titles, metrics, shortcuts, project cards, activities, and todos use English runtime copy

### Requirement: Workbench page must display real navigation entries

The system SHALL converge default workbench quick entries, project cards, and operational todos into real navigation and operational semantics matching the LinaPro management backend, and MUST NOT retain inaccessible template demo links.

#### Scenario: Click workbench quick entry
- **WHEN** an administrator clicks a quick entry on the workbench homepage
- **THEN** the host navigates to an accessible LinaPro internal page or a clear external official resource

#### Scenario: Browse workbench homepage content
- **WHEN** an administrator opens the workbench homepage
- **THEN** displayed projects, activities, and todos are organized around core host service, default management workspace, plugin extension capabilities, and development collaboration flows
- **AND** page copy remains consistent with current project positioning

### Requirement: Analysis page must provide clear metric semantics and time range switching

The system SHALL provide switchable time range perspectives in the default analysis page, and use clear, non-duplicated titles for key metric sections and chart sections.

#### Scenario: Switch analysis time range
- **WHEN** an administrator switches a preset time range on the analysis page
- **THEN** the page refreshes overview metrics, trend summaries, and insights with corresponding time range data

#### Scenario: Browse analysis chart sections
- **WHEN** an administrator views analysis page chart cards
- **THEN** each card title accurately expresses the content it displays
- **AND** the page no longer has two meaningfully different charts with the same title

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
