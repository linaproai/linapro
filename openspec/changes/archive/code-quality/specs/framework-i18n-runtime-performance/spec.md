## ADDED Requirements

### Requirement: Language switching MUST NOT reload the full permission, menu, and route state

When the user switches language, the frontend SHALL only refresh local state that is strongly tied to language, including public configuration synchronization and dictionary cache reset. Language switching MUST NOT trigger full permission reload flows such as `refreshAccessibleState`, which refetch menus and regenerate routes. Menu and route titles MUST update automatically through reactive `$t(...)` usage or `meta.i18nKey` values carried by the first menu response. The frontend MUST NOT bake the current-language text into static strings during route generation and lose local redraw capability.

#### Scenario: Language switching only updates public config and dict cache

- **WHEN** the user switches `preferences.app.locale` in the UI
- **THEN** the frontend MUST call `syncPublicFrontendSettings(locale)` to synchronize public configuration
- **AND** MUST call `useDictStore().resetCache()` to reset dictionary cache
- **AND** MUST NOT call `refreshAccessibleState(router)` to regenerate routes

#### Scenario: Menu titles update reactively after language switching

- **WHEN** the user switches language and stays on the current page
- **THEN** menus and breadcrumbs MUST automatically display text in the new language
- **AND** this process MUST NOT refetch `/api/v1/user/info` or any menu API
- **AND** `meta.i18nKey` from the initial menu route response MUST be sufficient for the frontend to re-resolve menu titles from local runtime language packs

#### Scenario: Route meta.title MUST reference i18n keys

- **WHEN** any route configuration defines `meta.title`
- **THEN** the field MUST be an i18n key or `() => $t(...)`
- **AND** MUST NOT be evaluated once during route initialization into a string for a specific language
