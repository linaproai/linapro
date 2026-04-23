## ADDED Requirements

### Requirement: Official source plugin usage field-capability plugin ID

The system SHALL uses domain-capability style `kebab-case` flags without the `plugin-` prefix for official source plugins to improve readability and avoid semantic duplication.

#### Scenario: Define the official source plugin identifier
- **WHEN** team named the official source plugin in the open source stage
- **THEN** Plugin ID uses `org-center`, `content-notice`, `monitor-online`, `monitor-server`, `monitor-operlog`, `monitor-loginlog`
- **AND** does not require the `plugin-` prefix

#### Scenario: Verify the validity of the plugin ID
- **WHEN** The host parses the `plugin.yaml` of the above official plugin
- **THEN** These plugin IDs only need to satisfy globally unique and `kebab-case` rules
- **AND** not illegal due to missing `plugin-` prefix

### Requirement: The source plugin menu MUST be mounted to the host stable directory

The system SHALL requires the official source plugin to point to the host stable directory key through `parent_key` in the manifest menu statement to ensure the long-term stability of the background navigation structure.

#### Scenario: Organization and content plugin declares parent directory
- **WHEN** `org-center` or `content-notice` declare menu metadata
- **THEN** Its top-level menu `parent_key` points to the host directory keys `org` and `content` respectively
- **AND** The internal submenu of the plugin can still continue to refer to the parent menu key declared by the same plugin.

#### Scenario: Monitoring plugin declares parent directory
- **WHEN** `monitor-online`, `monitor-server`, `monitor-operlog`, `monitor-loginlog` declare menu metadata
- **THEN** Its top menu `parent_key` points to the host directory key `monitor`
- **AND** The host presses the parent key to complete menu synchronization and start-stop linkage visibility management

#### Scenario: Official plugin uses fixed parent directory key mapping
- **WHEN** The host verifies the official source plugin manifest
- **THEN** The top-level `parent_key` of `org-center` MUST be `org`
- **AND** The top-level `parent_key` of `content-notice` MUST be `content`
- **AND** The top-level `parent_key` of `monitor-online`, `monitor-server`, `monitor-operlog`, `monitor-loginlog` MUST be `monitor`

#### Scenario: The official plugin declares an unsupported top-level mount key
- **WHEN** The above official source plugin uses `parent_key` that is inconsistent with the convention in its top-level menu declaration.
- **THEN** The host refuses to synchronize the plugin menu
- **AND** Provide administrators with diagnosable mount verification errors

### Requirement: The source plugin backend directory structure MUST converge to backend/internal

The system SHALL requires the source plugin to converge the backend business implementation under `backend/internal/`, avoid directly exposing the business service directory in the `backend/` root directory, and ensure that the private implementation boundary of the plugin is clear and consistent with the host agreement.

#### Scenario: Planning the standard directory of source plugins
- **WHEN** Team creates or refactors a source plugin
- **THEN** The plugin backend is organized by at least `backend/api/`, `backend/plugin.go`, `backend/internal/controller/`, `backend/internal/service/`
- **AND** Plugin frontend pages remain in `frontend/pages/`
- **AND** Plugin manifests and embedded resources are kept in `plugin.yaml`, `plugin_embed.go`, `manifest/sql/` and `manifest/sql/uninstall/`

#### Scenario: Place the plugin service component
- **WHEN** team adds or migrates business services for source plugins
- **THEN** All service components MUST be placed in `backend/internal/service/<component>/`
- **AND** MUST NOT CREATE `backend/service/<component>/`
- **AND** Non-`internal` directories such as `backend/provider/` are only used for stable capability provider / adapter and do not carry main business orchestration

#### Scenario: Plugin requires local ORM artifacts
- **WHEN** Source code plugin needs to access the database
- **THEN** `backend/hack/config.yaml` serves as the plugin’s local `gf gen dao` configuration entry
- **AND** The generated results fall into `backend/internal/dao/`, `backend/internal/model/do/` and `backend/internal/model/entity/`
- **AND** Access to the host shared table also continues to use the plugin's local generation of artifacts
