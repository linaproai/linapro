# plugin-ui-integration Specification

## Purpose
TBD - created by archiving change plugin-framework. Update Purpose after archive.
## Requirements
### Requirement: Plugin pages support multiple host integration modes
The system SHALL support three plugin page integration modes: `iframe`, new-tab navigation, and host-embedded mounting.

#### Scenario: Integrate a plugin page in iframe mode
- **WHEN** a plugin menu declares a page that uses `iframe` mode
- **THEN** the host loads the plugin page URL in an iframe within the unified layout
- **AND** the host still manages the page entry with LinaPro menus, permissions, and navigation

#### Scenario: Integrate a plugin page in new-tab mode
- **WHEN** a plugin menu declares a page that uses new-tab mode
- **THEN** the host opens the corresponding plugin page in a new browser tab when the menu is clicked
- **AND** the original host workspace navigation state remains unaffected

#### Scenario: Integrate a plugin page in host-embedded mode
- **WHEN** a plugin page declares host-embedded mounting mode
- **THEN** the host loads the plugin frontend mount entry into the designated container
- **AND** the plugin can use any frontend framework internally
- **AND** the frontend entry MUST satisfy the host-defined mount contract

### Requirement: The host manages plugin frontend assets
The system SHALL host frontend assets according to plugin type and provide stable asset access URLs for each page integration mode.

#### Scenario: Source plugins participate in the host build
- **WHEN** a source plugin provides frontend assets according to convention
- **THEN** those frontend assets participate in the host frontend build pipeline
- **AND** host-embedded pages for source plugins MUST be implemented with real frontend source files instead of declarative JSON-only page descriptions
- **AND** those source files are placed under `frontend/pages/` by default and discovered during the host build
- **AND** the final assets are embedded into the backend binary together with the host frontend

#### Scenario: Dynamic plugins provide frontend assets
- **WHEN** a runtime `wasm` plugin with assets is installed
- **THEN** the host extracts and hosts the plugin's static frontend assets
- **AND** the host generates a stable static asset path for the plugin
- **AND** plugin pages without correctly prepared frontend assets cannot be enabled

#### Scenario: Dynamic plugin frontend assets can come from embedded resource declarations
- **WHEN** a dynamic plugin author declares `frontend` directory assets with `go:embed`
- **THEN** the builder MUST extract frontend static assets from the embedded file system and write them into the runtime snapshot
- **AND** the host continues to serve `/plugin-assets/<plugin-id>/<version>/...` from that runtime snapshot
- **AND** the host does not need to read frontend asset contents item by item through guest runtime methods

### Requirement: Plugin pages participate in host routing and generation awareness
The system SHALL include enabled plugin pages in the host's dynamic routing system and protect the current user experience when plugin generations change.

#### Scenario: The host generates plugin page routes
- **WHEN** an enabled plugin declares visible menus and page entries
- **THEN** the host adds those entries to the current user's accessible menu set and dynamic routes
- **AND** plugin page visibility remains controlled by LinaPro role permissions

#### Scenario: A plugin inserts a top-level entry into the left navigation menu
- **WHEN** an enabled plugin declares a top-level entry for the left navigation menu
- **THEN** the host shows that plugin entry near the top of the left menu
- **AND** clicking the entry opens the plugin tab page in the main content area
- **AND** the page content clearly indicates that the entry comes from the plugin extension mechanism

#### Scenario: Refresh the left navigation menu immediately after enabling a plugin in the current session
- **WHEN** an administrator enables a source plugin that is installed but currently disabled in an already logged-in session
- **THEN** the host refreshes dynamic routes and the left navigation menu without requiring a new login
- **AND** the new top-level menu entry added by the plugin becomes visible immediately
- **AND** the current page remains inside the host workspace environment

#### Scenario: The current session regains focus after plugin state changes
- **WHEN** another operation outside the current logged-in session changes plugin enablement state and the current browser tab regains focus
- **THEN** the host automatically resynchronizes plugin-related dynamic routes and the left navigation menu
- **AND** entries for enabled plugins become visible again
- **AND** entries for disabled plugins are hidden in time

#### Scenario: A hot upgrade happens while the current user stays on a plugin page
- **WHEN** the current user is visiting a plugin page and that plugin upgrades to a new generation
- **THEN** the host shows a prompt telling the user that the plugin has been updated and the current page should be refreshed
- **AND** users who are not on that plugin page are not forcibly interrupted

#### Scenario: The current plugin page switches to the new generation after refresh
- **WHEN** the current user stays on a dynamic plugin page and that plugin has already completed its generation switch
- **THEN** after the user clicks `Refresh current page`, the host first rebuilds the accessible menus and dynamic routes for the current logged-in user
- **AND** the host does not forcibly navigate the user to another workspace page first
- **AND** the current plugin page then remounts with the new-generation assets and updates the related tab metadata

#### Scenario: Users on non-plugin pages remain unaffected after a hot upgrade
- **WHEN** a dynamic plugin completes a hot upgrade but the current user is not staying on that plugin page
- **THEN** the host does not show the plugin refresh prompt
- **AND** the host does not forcibly interrupt browsing on the current non-plugin page because of that upgrade
- **AND** users only start using the new-generation assets when they enter the plugin page later

### Requirement: The plugin management page focuses on key governance information
The system SHALL prioritize the key governance fields needed on the plugin management page and avoid crowding the list overview with low-value information.

#### Scenario: Show the governance category for source plugins
- **WHEN** the plugin type in the manifest is `source`
- **THEN** the plugin management page shows it as `Source Plugin` in the `Plugin Type` column

#### Scenario: Show the governance category for dynamic plugins
- **WHEN** the plugin type in the manifest is `dynamic`
- **THEN** the plugin management page consistently shows it as `Dynamic Plugin` in the `Plugin Type` column
- **AND** the current `dynamic` plugin is still shown and filtered as `Dynamic Plugin` in the governance list even when the concrete artifact is a `wasm` product

#### Scenario: Simplify the plugin management list overview fields
- **WHEN** an administrator views the plugin management list
- **THEN** the page no longer shows the `Delivery Mode`, `Integration Mode`, and `Entry` columns
- **AND** the page shows `Installed At` before `Updated At`

### Requirement: Plugin Management Supports a Detail Dialog
The system SHALL provide a detail dialog for plugin records in plugin management so administrators can inspect plugin governance information without leaving the list page.

#### Scenario: Administrator opens plugin details from the list
- **WHEN** an administrator clicks the detail action for a plugin record
- **THEN** the system opens a detail dialog
- **AND** the dialog shows the plugin’s identity, type, version, description, installation state, status, authorization requirements, authorization status, install time, and update time

### Requirement: Dynamic Plugin Details Merge Requested and Effective Host-Service Scope
The system SHALL present dynamic-plugin host-service information in a way that distinguishes declared intent from effective authorization only when the two differ.

#### Scenario: Requested and authorized host-service scopes are identical
- **WHEN** a dynamic plugin has the same requested and authorized host-service scope
- **THEN** the dialog merges them into one effective scope presentation
- **AND** the UI avoids duplicate explanatory text

#### Scenario: Requested and authorized host-service scopes differ
- **WHEN** a dynamic plugin’s authorized host-service scope differs from its requested scope
- **THEN** the dialog shows requested scope and authorized scope as distinct sections
- **AND** resource groups use one semantic heading plus a resource list instead of repeating the same prefix for every resource item

### Requirement: Empty Host-Service State Is Only Shown for Dynamic Plugins
The system SHALL show an empty host-service state only when a dynamic plugin has no host-service governance data.

#### Scenario: Source plugin without host-service governance data
- **WHEN** a source plugin has no host-service governance information
- **THEN** the detail dialog does not show an unnecessary host-service empty-state block

### Requirement: Dynamic Plugin Demo Records Support Pagination
The system SHALL let administrators page through demo records on the dynamic plugin sample page.

#### Scenario: Administrator switches demo-record pages
- **WHEN** an administrator changes the page in the dynamic plugin demo record list
- **THEN** the page reloads the target page and updates the range summary and current page state accordingly

### Requirement: Plugin installation dialog supports an install-and-enable shortcut

The system SHALL provide both `Install Only` and `Install and Enable` governance actions in the plugin installation dialog so administrators can complete immediate enablement after reviewing plugin information without returning to the list for a second step.

#### Scenario: Show the shortcut action when the user has install and enable permissions
- **WHEN** an administrator opens the installation dialog for a plugin that is not installed and the current account has both `plugin:install` and `plugin:enable`
- **THEN** the dialog shows the `Install and Enable` shortcut action
- **AND** the dialog still keeps the `Install Only` action so the administrator can explicitly choose installation without enablement

#### Scenario: Hide the shortcut action when the user only has install permission
- **WHEN** an administrator opens the installation dialog for a plugin that is not installed but the current account has `plugin:install` without `plugin:enable`
- **THEN** the dialog shows only the `Install Only` action
- **AND** the UI MUST NOT show the `Install and Enable` button so it does not imply access to a governance action outside the user's permission scope

#### Scenario: Show the real state when the second step of the composite action fails
- **WHEN** an administrator chooses `Install and Enable` in the installation dialog and the install step succeeds but the enable step fails
- **THEN** the refreshed plugin management page MUST show the plugin as `installed but disabled`
- **AND** the UI clearly tells the administrator that installation completed but enablement did not succeed and can be retried later

### Requirement: Host-embedded plugin pages must participate in host locale context and message refresh
The system SHALL let host-managed plugin pages participate in the host locale context and runtime message refresh flow. When the active language changes, host-embedded plugin pages and their route metadata MUST refresh to the new language without requiring plugin reinstallation.

#### Scenario: Plugin route titles refresh after language switch
- **WHEN** a user switches workspace language in an authenticated session
- **THEN** the host rebuilds plugin route titles and menu titles accessible to the current user
- **AND** enabled plugin pages switch their navigation and tab display copy to the new language

#### Scenario: Host-embedded plugin pages load runtime message bundles
- **WHEN** the host loads a plugin frontend page in embedded mode
- **THEN** the plugin page can access the host current-locale context and matching runtime translation resources
- **AND** the plugin does not need to reimplement language resolution rules detached from the host

### Requirement: Plugin language resource changes must be observed by the host promptly
The system SHALL refresh plugin-related runtime messages after plugin enablement, disablement, or upgrade so that the host UI does not continue showing stale plugin translations.

#### Scenario: Enabled plugin immediately displays plugin translations
- **WHEN** an administrator enables a plugin that has i18n resources in the current session
- **THEN** the host refreshes runtime message bundles and dynamic menus
- **AND** the plugin's navigation title and page copy can immediately display in the current language

#### Scenario: Plugin upgrade switches to new-version translation resources
- **WHEN** a plugin is upgraded to a release containing new translation resources and the switch becomes effective
- **THEN** subsequent runtime message bundles provided by the host use the new release resources
- **AND** old release translation messages are no longer consumed by the frontend

