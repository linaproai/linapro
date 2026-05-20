## ADDED Requirements

### Requirement: `linactl` must unify repository tool implementations

The system SHALL maintain image building, dynamic plugin Wasm packaging, and runtime i18n governance scanning implementations as internal subcomponents under `hack/tools/linactl/internal/`. `hack/tools/image-builder`, `hack/tools/build-wasm`, and `hack/tools/runtime-i18n` MUST NOT continue to exist as independent Go tool modules in the default development path.

#### Scenario: Image building executes through `linactl` internal component

- **WHEN** a developer runs `linactl image` or `linactl image.build`
- **THEN** the command directly calls the implementation in `hack/tools/linactl/internal/imagebuilder`
- **AND** the command MUST NOT call the independent tool via `go run ./hack/tools/image-builder`

#### Scenario: Dynamic plugin Wasm packaging executes through `linactl` internal component

- **WHEN** a developer runs `linactl wasm`
- **THEN** the command directly calls the implementation in `hack/tools/linactl/internal/wasmbuilder`
- **AND** the command MUST NOT call the independent tool by entering `hack/tools/build-wasm` and executing `go run .`

#### Scenario: Runtime i18n governance scanning executes through `linactl` internal component

- **WHEN** a developer runs `linactl i18n.check`
- **THEN** the command directly calls the implementation in `hack/tools/linactl/internal/runtimei18n`
- **AND** the command MUST NOT call the independent tool by entering `hack/tools/runtime-i18n` and executing `go run . scan` or `go run . messages`

### Requirement: Public development commands must remain stable

The system SHALL keep `make image`, `make image.build`, `make wasm`, `make i18n.check`, `linactl image`, `linactl image.build`, `linactl wasm`, and `linactl i18n.check` public entries stable. Tool implementation migration MUST NOT require developers to use new command names.

#### Scenario: Make entry continues to call `linactl`

- **WHEN** a developer runs `make image`, `make image.build`, `make wasm`, or `make i18n.check`
- **THEN** the root Makefile forwards to the corresponding `linactl` command
- **AND** the developer does not need to call old independent tool directories directly

### Requirement: Tool consolidation must update repository references

The system SHALL remove direct references to `hack/tools/image-builder`, `hack/tools/build-wasm`, and `hack/tools/runtime-i18n` from the default development path, including `go.work`, CI fixtures, test helpers, and tool documentation. Historical OpenSpec archive content may retain old paths as historical records.

#### Scenario: Default workspace no longer contains old tool modules

- **WHEN** a developer views the root `go.work`
- **THEN** the workspace only contains Go modules that still exist
- **AND** it does not contain deleted `hack/tools/image-builder`, `hack/tools/build-wasm`, or `hack/tools/runtime-i18n`

#### Scenario: Reference scan finds no old entry in default path

- **WHEN** a developer scans current code, documentation, and workflows for `go run ./hack/tools/image-builder`, `go run ./hack/tools/build-wasm`, `go run ./hack/tools/runtime-i18n`, `hack/tools/image-builder`, `hack/tools/build-wasm`, and `hack/tools/runtime-i18n`
- **THEN** except for historical archive descriptions, there are no references requiring use of old independent tool entries
