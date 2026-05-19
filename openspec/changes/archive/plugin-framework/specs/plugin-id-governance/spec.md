## Requirements

### Requirement: Plugin ID Must Satisfy Runtime Safety Boundary

The system SHALL only enforce basic safety boundaries for plugin IDs at runtime: the ID must not be empty, total length MUST NOT exceed 64 characters, and MUST use lowercase letters, digits, and hyphens组成的 kebab-case text, to ensure safe use in URL paths, dynamic asset paths, filenames, database keys, menu keys, permission strings, i18n namespaces, and apidoc namespaces. The host runtime SHALL NOT enforce `<author>-<domain>-<capability>` structure, domain whitelists, official capability reservations, or old official ID rejection tables.

#### Scenario: Accepts officially recommended structured plugin ID
- **WHEN** a plugin manifest declares `id: linapro-content-notice`
- **THEN** the system accepts the plugin ID
- **AND** the ID can continue serving as derived namespace for menus, permissions, resources, i18n, and apidoc

#### Scenario: Accepts non-three-segment extension plugin ID
- **WHEN** a plugin manifest declares `id: demo-control`
- **THEN** the system accepts the plugin ID
- **AND** the system must not reject the manifest because it does not satisfy the `<author>-<domain>-<capability>` recommended structure

#### Scenario: Accepts custom domain segment
- **WHEN** a plugin manifest declares `id: acme-random-report`
- **THEN** the system accepts the plugin ID
- **AND** the system must not reject the manifest because `random` is not in the host built-in list

#### Scenario: Rejects unsafe characters
- **WHEN** a plugin manifest declares `id: Acme_Report`
- **THEN** the system rejects the manifest
- **AND** the error states the plugin ID must use kebab-case lowercase letters and digits

#### Scenario: Rejects overly long plugin ID
- **WHEN** a plugin manifest declares an ID exceeding 64 characters
- **THEN** the system rejects the manifest
- **AND** the error states the plugin ID length must not exceed 64 characters

### Requirement: Official Plugin IDs Must Use Normalized Mapping

The system SHALL normalize LinaPro official plugin IDs shipped with the repository to the following mapping, and must not continue using old official IDs in official plugin runtime configuration, manifests, source registration, menus, permissions, cron, i18n, apidoc, tests, or documentation forward paths. This mapping only constrains LinaPro official plugin assets and does not serve as a runtime rejection table for third-party plugin IDs.

| Old ID | New ID |
|--------|--------|
| `content-notice` | `linapro-content-notice` |
| `monitor-loginlog` | `linapro-monitor-loginlog` |
| `monitor-operlog` | `linapro-monitor-operlog` |
| `monitor-online` | `linapro-monitor-online` |
| `monitor-server` | `linapro-monitor-server` |
| `multi-tenant` | `linapro-tenant-core` |
| `org-center` | `linapro-org-core` |
| `plugin-demo-dynamic` | `linapro-demo-dynamic` |
| `plugin-demo-source` | `linapro-demo-source` |
| `demo-control` | `linapro-ops-demo-guard` |

#### Scenario: Official plugin list uses new ID
- **WHEN** the host scans `apps/lina-plugins/linapro-org-core/plugin.yaml`
- **THEN** the manifest ID is `linapro-org-core`
- **AND** the host must not discover `org-center` in the official plugin list

#### Scenario: Official auto-enable config uses new ID
- **WHEN** the host reads the repository default `plugin.autoEnable`
- **THEN** official plugin entries use the normalized new IDs
- **AND** the default config must not continue using `multi-tenant`, `org-center`, or other old official IDs

### Requirement: Plugin Runtime Identity Must Use Current Plugin ID

The system SHALL use the current plugin ID at runtime identity boundaries and must not provide alias, redirect, or compatible lookup for old official IDs. This boundary includes plugin management API, extension API, dynamic frontend asset URL, menu keys, permission strings, cron handlerRefs, plugin state tables, release tables, migration tables, resource reference tables, node state tables, plugin KV state tables, and host service authorization records.

#### Scenario: New extension API path uses current ID
- **WHEN** dynamic plugin `linapro-demo-dynamic` exposes an extension API
- **THEN** the host public path uses `/api/v1/extensions/linapro-demo-dynamic/...`
- **AND** the host must not expose the same official plugin via `/api/v1/extensions/plugin-demo-dynamic/...`

#### Scenario: New dynamic asset path uses current ID
- **WHEN** dynamic plugin `linapro-demo-dynamic` provides frontend assets
- **THEN** the host asset path uses `/plugin-assets/linapro-demo-dynamic/<version>/...`
- **AND** the host must not expose the same official plugin assets via `/plugin-assets/plugin-demo-dynamic/<version>/...`

#### Scenario: New cron handlerRef uses current ID
- **WHEN** plugin `linapro-monitor-server` registers built-in scheduled tasks
- **THEN** handlerRef uses `plugin:linapro-monitor-server/cron:<name>`
- **AND** the system must not continue generating `plugin:monitor-server/cron:<name>`

### Requirement: Repository Governance Scan Must Verify Official Plugin ID Consistency

The system SHALL provide automated verification ensuring official plugin directory names, manifest IDs, source plugin registration IDs, dynamic artifact manifests, dependency declarations, menu keys, runtime i18n keys, apidoc i18n keys, configuration, and test fixtures use the same current plugin ID. Verification failure must block the change.

#### Scenario: Directory name inconsistent with manifest ID
- **WHEN** plugin directory is `apps/lina-plugins/linapro-content-notice`
- **AND** the directory's `plugin.yaml` declares `id: content-notice`
- **THEN** governance verification fails
- **AND** the error states the directory name is inconsistent with the manifest ID

#### Scenario: i18n namespace uses old ID
- **WHEN** plugin `linapro-content-notice` runtime language pack contains `plugin.content-notice.name`
- **THEN** governance verification fails
- **AND** the error states runtime i18n keys must use the `plugin.linapro-content-notice.` prefix

#### Scenario: apidoc namespace uses current ID
- **WHEN** plugin `linapro-demo-dynamic` apidoc language pack contains `plugins.linapro_demo_dynamic`
- **THEN** governance verification passes
