## ADDED Requirements

### Requirement: pluginbridge must provide responsibility-scoped public subcomponents

The system SHALL organize `apps/lina-core/pkg/pluginbridge` into responsibility-scoped public subcomponent packages. Subcomponents must cover at least bridge contracts, bridge encoding/decoding, WASM artifact helpers, host call protocol, host service protocol, and guest SDK. The root directory must not continue carrying large numbers of cross-responsibility implementation files; the root package is only allowed to retain a facade, package documentation, and necessary compatibility entries.

#### Scenario: Developer locates bridge capabilities by responsibility

- **WHEN** a developer needs to view dynamic plugin bridge contracts, encoding/decoding, WASM artifact parsing, host call, host service, or guest SDK
- **THEN** the corresponding source code is in semantically clear `pkg/pluginbridge/<subcomponent>/` subcomponent directories
- **AND** the number of production source files in the root package directory stays within the 1-3 range

#### Scenario: Subcomponent names express stable responsibilities

- **WHEN** the system completes pluginbridge subcomponent organization
- **THEN** subcomponent package names must use clear responsibility names
- **AND** MUST NOT use fallback package names like `common`, `util`, or `helper` to carry cross-domain logic

### Requirement: Root package facade must preserve existing stable call paths

The system SHALL retain `lina-core/pkg/pluginbridge` root package as a compatibility facade. Existing public constants, types, and functions must remain accessible through the root package, delegating to corresponding subcomponent implementations. The facade must not duplicate protocol implementation logic; protocol implementation must exist in only one authoritative subcomponent.

#### Scenario: Old import paths continue to compile

- **WHEN** host internal code, dynamic plugin examples, or user plugins continue to import `lina-core/pkg/pluginbridge`
- **THEN** existing public API calls must continue to compile
- **AND** return behavior must remain consistent with pre-refactor behavior

#### Scenario: Facade does not re-implement protocol logic

- **WHEN** a developer inspects the root package facade
- **THEN** public entries use type alias, const alias, or wrapper calls to subcomponents
- **AND** the root package must not maintain independent protobuf wire encoding/decoding, WASM section traversal, or host service payload encoding/decoding implementations

### Requirement: Subcomponent dependency direction must prevent import cycles

The system SHALL clearly define the dependency direction of `pluginbridge` subcomponents. Bottom-level contract and protocol subcomponents must not depend on the root package facade or guest SDK; the root package facade may depend on subcomponents. Any subcomponent's `internal` implementation package must serve its explicit parent component and must not become a cross-component fallback dependency.

#### Scenario: Subcomponent builds have no import cycle

- **WHEN** `go test ./pkg/pluginbridge/...` is executed
- **THEN** all subcomponent packages must compile
- **AND** no import cycles may appear

#### Scenario: Bottom-level packages do not depend on root package

- **WHEN** inspecting `contract`, `codec`, `artifact`, `hostcall`, `hostservice` subcomponent imports
- **THEN** these subcomponents must not import `lina-core/pkg/pluginbridge`
- **AND** may only depend on subcomponents at the same or lower level

### Requirement: Host internal calls must prefer precise subcomponents

The system SHALL gradually migrate project-controlled host internal calls to precise subcomponent imports. Dynamic plugin guest code may continue using the root package facade compatibility path, but host runtime, WASM host function, artifact parsing, i18n/apidoc resource loading, and plugindb should use subcomponent packages that express responsibility boundaries.

#### Scenario: Host runtime uses precise subcomponents

- **WHEN** the host runtime parses dynamic plugin artifacts or executes Wasm bridge requests
- **THEN** code preferentially imports `pluginbridge/artifact`, `pluginbridge/codec`, `pluginbridge/hostcall`, or `pluginbridge/hostservice`
- **AND** does not import the entire root package facade when only a single protocol capability is needed

#### Scenario: Plugin-side compatibility path remains available

- **WHEN** dynamic plugin guest code continues calling `pluginbridge.NewGuestRuntime`, `pluginbridge.BindJSON`, or `pluginbridge.Runtime()`
- **THEN** the system continues providing compatibility entries
- **AND** these entries delegate to the guest subcomponent

### Requirement: Subcomponent organization must not change bridge protocol behavior

The system SHALL guarantee that subcomponent organization is a structural refactor that does not change dynamic plugin bridge protocol behavior. ABI constants, WASM custom section names, protobuf field numbers, host call status codes, host service service/method strings, payload encoding/decoding results, and guest helper behavior must remain unchanged.

#### Scenario: Bridge envelope encoding/decoding remains unchanged

- **WHEN** the refactored API is used to encode and decode `BridgeRequestEnvelopeV1` or `BridgeResponseEnvelopeV1`
- **THEN** the round trip result is equivalent to pre-refactor
- **AND** existing protocol tests must continue to pass

#### Scenario: Host service payload codec remains unchanged

- **WHEN** the refactored API is used to encode and decode runtime, storage, network, data, cache, lock, config, notify, or cron host service payloads
- **THEN** the round trip result is equivalent to pre-refactor
- **AND** field numbers and default value semantics must not change

#### Scenario: Facade and subcomponent results are consistent

- **WHEN** the same call is executed through both the root package facade and the target subcomponent
- **THEN** both return the same result or equivalent error
- **AND** tests must cover at least bridge envelope, WASM section, and host service payload representative entries

### Requirement: Subcomponent organization must have automated verification

The system SHALL provide automated verification for `pluginbridge` subcomponent organization. Verification must cover root package compatibility, subcomponent compilation, host internal calls, dynamic plugin examples, and Wasm guest builds.

#### Scenario: pluginbridge subcomponent tests pass

- **WHEN** `go test ./pkg/pluginbridge/...` is executed
- **THEN** root package facade and all subcomponent tests must pass

#### Scenario: Host plugin runtime tests pass

- **WHEN** plugin runtime, WASM host function, and plugindb related Go tests are executed
- **THEN** tests must pass
- **AND** no protocol behavior regression from import migration

#### Scenario: Dynamic plugin example can build

- **WHEN** normal Go test and `GOOS=wasip1 GOARCH=wasm` build are executed on the dynamic plugin example
- **THEN** the example must compile
- **AND** guest-side bridge helper calls must be available
