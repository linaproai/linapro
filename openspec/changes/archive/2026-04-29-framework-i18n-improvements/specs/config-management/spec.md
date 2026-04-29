## ADDED Requirements

### Requirement: Config export and import headers must be resolved via translation keys by current language
The system SHALL resolve column headers (`name` / `key` / `value` / `remark` / `createdAt` / `updatedAt`) in the `Export configs to Excel` and `Import configs from Excel` pipelines by current request language through `config.field.<name>` translation keys, and MUST NOT maintain English/Chinese literal mapping tables in backend Go source. `apps/lina-core/manifest/i18n/<locale>/*.json` MUST contain translation values under the `config.field.*` namespace; adding a new language only requires appending resources, without modifying Go code.

#### Scenario: Traditional Chinese environment export contains Traditional Chinese headers
- **WHEN** an administrator requests `GET /config/export` in the `zh-TW` environment
- **THEN** the exported Excel column headers are displayed in Traditional Chinese
- **AND** the backend resolves headers via `config.field.name`, `config.field.key`, `config.field.value`, `config.field.remark`, `config.field.createdAt`, `config.field.updatedAt` translation keys
- **AND** no hardcoded literal mappings like `englishLabels` / `chineseLabels` exist in backend code

#### Scenario: Adding a new language requires no backend Go code changes
- **WHEN** the project enables a new built-in language and provides `manifest/i18n/<locale>/*.json` resources for that language
- **AND** the resource contains translation values under the `config.field.*` namespace
- **THEN** config import and export headers automatically display in that language
- **AND** no source code changes are needed in `apps/lina-core/internal/service/sysconfig/`
