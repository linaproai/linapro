# Dashboard Workbench

## Purpose

Define the default management workbench and analysis page behavior, including real LinaPro navigation, runtime i18n copy, project cards, logo treatment, and theme preference precedence.

## Requirements

### Requirement: Workbench home must expose real LinaPro management entries

The system SHALL align dashboard shortcuts, project cards, activities, and operational todos with real LinaPro management semantics. It MUST NOT keep unreachable template demo links or unrelated template narratives.

#### Scenario: Click workbench shortcuts
- **WHEN** an administrator clicks a workbench shortcut
- **THEN** the app navigates to an accessible LinaPro internal page or a clear official external resource
- **AND** template placeholder routes such as `/demos/...` are not used

#### Scenario: Browse workbench home content
- **WHEN** an administrator opens the workbench home page
- **THEN** projects, activity, and todos are organized around the core host, default workbench, plugin extension, and OpenSpec delivery workflow
- **AND** page copy stays aligned with LinaPro positioning

### Requirement: Analysis page must provide clear metric semantics and time-range switching

The default analysis page SHALL provide switchable time-range views and clear non-duplicated titles for metric and chart blocks.

#### Scenario: Switch analysis time range
- **WHEN** an administrator switches a preset time range
- **THEN** the page refreshes overview metrics, trend summaries, and insights for that range
- **AND** the selected range remains visibly active

#### Scenario: Browse analysis chart sections
- **WHEN** an administrator views analysis chart cards
- **THEN** each card title accurately describes its content
- **AND** different charts do not share the same ambiguous title

### Requirement: Workbench page copy must use runtime i18n resources

The workbench SHALL load delivered page content through runtime i18n resources so English environments do not display Chinese activities, todos, shortcuts, project cards, or explanatory copy.

#### Scenario: Workbench displays English default copy
- **WHEN** an administrator opens the workbench in `en-US`
- **THEN** titles, metrics, shortcuts, project cards, activities, and todos use English runtime copy
- **AND** no Chinese default system copy is shown

#### Scenario: Workbench copy changes with language switch
- **WHEN** an administrator switches from `zh-CN` to `en-US`
- **THEN** the workbench refreshes to English without requiring login again
- **AND** route title, tab title, and page content language remain consistent

#### Scenario: Workbench quick navigation targets core pages
- **WHEN** an administrator opens the workbench
- **THEN** quick navigation shows User Management, Menu Management, System Config, Extension Center, API Docs, and Scheduled Jobs
- **AND** those entries navigate to `/system/user`, `/system/menu`, `/system/config`, `/system/plugin`, `/about/api-docs`, and `/system/job`

#### Scenario: Workbench project cards reflect the delivered admin stack
- **WHEN** an administrator opens the workbench
- **THEN** project cards show LinaPro, GoFrame, Vue, Vben, Ant Design, and TypeScript
- **AND** each card links to the matching official site or documentation
- **AND** card dates show `2026-05-01`
- **AND** card descriptions remain short and use ellipsis when too long

### Requirement: Management workbench logo must provide subtle dark-mode edge depth

The management workbench SHALL add a very subtle cyan edge glow to the brand logo icon in dark mode, without affecting light mode, clickable area, or navigation layout stability.

#### Scenario: Dark-mode logo edge glow remains subtle and layout-safe
- **WHEN** an administrator opens the workbench in dark mode
- **THEN** the logo icon edge shows a subtle cyan glow following the icon pixels
- **AND** logo text, sidebar menu, and top navigation do not shift

#### Scenario: Light-mode logo keeps original treatment
- **WHEN** an administrator switches to light mode
- **THEN** the logo icon does not show the dark-mode-only cyan edge glow

### Requirement: User theme preference must override the public frontend default

The workbench SHALL preserve a user-explicit theme preference when synchronizing public frontend config. The system parameter `sys.ui.theme.mode` applies only when the user has no explicit local preference.

#### Scenario: Existing user theme preference survives page refresh
- **GIVEN** the user explicitly selected dark mode
- **AND** public frontend config declares `sys.ui.theme.mode=light`
- **WHEN** the user refreshes the workbench
- **THEN** the workbench remains in dark mode

#### Scenario: System default applies when no user theme preference exists
- **GIVEN** the user has no explicit local theme preference
- **AND** public frontend config declares `light`, `dark`, or `auto`
- **WHEN** the user opens or refreshes the workbench
- **THEN** the workbench uses that system-provided theme mode as default

