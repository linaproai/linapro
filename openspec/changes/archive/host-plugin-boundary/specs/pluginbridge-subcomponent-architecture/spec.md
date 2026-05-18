## ADDED Requirements

### Requirement: pluginbridge MUST provide responsibility-scoped public subcomponents

The system SHALL organize `apps/lina-core/pkg/pluginbridge` into responsibility-scoped public subcomponent packages. Subcomponents must cover at least bridge contracts, bridge codec, WASM artifact helpers, host call protocol, host service protocol, and guest SDK. The root directory must not continue to host large numbers of cross-responsibility implementation files; the root package may only retain the facade, package documentation, and necessary compatibility entry points.

#### Scenario: Developer locates bridge capabilities by responsibility

- **WHEN** A developer needs to inspect dynamic plugin bridge contracts, codec, WASM artifact parsing, host call, host service, or guest SDK
- **THEN** The corresponding source code resides in semantically clear `pkg/pluginbridge/<subcomponent>/` subcomponent directories
- **AND** The number of production source files in the root package directory remains within the range of 1 to 3

#### Scenario: Subcomponent names express stable responsibilities

- **WHEN** The system completes pluginbridge subcomponentization
- **THEN** Subcomponent package names must use clear responsibility names
- **AND** Names such as `common`, `util`, or `helper` must not be used as catch-all packages for cross-domain logic

### Requirement: Root package facade MUST preserve existing stable call paths

The system SHALL preserve `lina-core/pkg/pluginbridge` root package as a backward-compatible facade. Existing public constants, types, and functions must remain accessible through the root package, delegating to the corresponding subcomponent implementations. The facade must not replicate protocol implementation logic; protocol implementation may exist in only one authoritative subcomponent.

#### Scenario: Old import paths continue to compile

- **WHEN** Host internal code, dynamic plugin samples, or user plugins continue to import `lina-core/pkg/pluginbridge`
- **THEN** Existing public API calls must continue to compile
- **AND** Return behavior must remain consistent with the pre-refactoring state

#### Scenario: Facade does not duplicate protocol logic

- **WHEN** A developer inspects the root package facade
- **THEN** Public entry points use type aliases, const aliases, or wrapper calls to subcomponents
- **AND** The root package must not maintain independent protobuf wire codec, WASM section traversal, or host service payload codec implementations

### Requirement: Subcomponent dependency direction MUST prevent circular dependencies

The system SHALL explicitly define the dependency direction for `pluginbridge` subcomponents. Low-level contract and protocol subcomponents must not depend on the root package facade or the guest SDK; the root package facade may depend on all subcomponents. Any `internal` implementation package sunk into a subcomponent must serve a clearly defined parent component and must not become a cross-component catch-all dependency.

#### Scenario: Subcomponent builds have no import cycles

- **WHEN** `go test ./pkg/pluginbridge/...` is executed
- **THEN** All subcomponent packages must pass compilation
- **AND** No import cycles may occur

#### Scenario: Low-level packages do not depend on root package

- **WHEN** The imports of `contract`, `codec`, `artifact`, `hostcall`, and `hostservice` subcomponents are inspected
- **THEN** These subcomponents must not import `lina-core/pkg/pluginbridge`
- **AND** They may only depend on subcomponents at the same or lower level in the dependency hierarchy

### Requirement: Host internal calls MUST prefer precise subcomponents

The system SHALL progressively migrate project-controlled host internal calls to precise subcomponent imports. Dynamic plugin guest code may continue to use the root package facade compatibility path, but host runtime, WASM host functions, artifact parsing, i18n/apidoc resource loading, and plugindb should use subcomponent packages that express the responsibility boundary.

#### Scenario: Host runtime uses precise subcomponents

- **WHEN** The host runtime parses dynamic plugin artifacts or executes Wasm bridge requests
- **THEN** Code preferentially imports `pluginbridge/artifact`, `pluginbridge/codec`, `pluginbridge/hostcall`, or `pluginbridge/hostservice`
- **AND** Does not import the entire root package facade just to access a single protocol capability

#### Scenario: Plugin-side compatibility path remains available

- **WHEN** Dynamic plugin guest code continues to call `pluginbridge.NewGuestRuntime`, `pluginbridge.BindJSON`, or `pluginbridge.Runtime()`
- **THEN** The system continues to provide compatibility entry points
- **AND** These entry points delegate to the guest subcomponent

### Requirement: Subcomponentization MUST NOT change bridge protocol behavior

The system SHALL guarantee that subcomponentization is a structural refactoring that does not change dynamic plugin bridge protocol behavior. ABI constants, WASM custom section names, protobuf field numbers, host call status codes, host service service/method strings, payload codec results, and guest helper behavior must remain unchanged.

#### Scenario: Bridge envelope codec remains unchanged

- **WHEN** The refactored API is used to encode and decode `BridgeRequestEnvelopeV1` or `BridgeResponseEnvelopeV1`
- **THEN** Round-trip results are equivalent to the pre-refactoring state
- **AND** Existing protocol tests must continue to pass

#### Scenario: Host service payload codec remains unchanged

- **WHEN** The refactored API is used to encode and decode runtime, storage, network, data, cache, lock, config, notify, or cron host service payloads
- **THEN** Round-trip results are equivalent to the pre-refactoring state
- **AND** Field numbers and default value semantics must not change

#### Scenario: Facade and subcomponent results are consistent

- **WHEN** The same call is executed through both the root package facade and the target subcomponent
- **THEN** Both return identical results or equivalent errors
- **AND** Tests must cover at least bridge envelope, WASM section, and host service payload representative entry points

### Requirement: Subcomponentization MUST have automated verification

The system SHALL provide automated verification for `pluginbridge` subcomponentization. Verification must cover root package compatibility, subcomponent compilation, host internal calls, dynamic plugin samples, and Wasm guest builds.

#### Scenario: pluginbridge subcomponent tests pass

- **WHEN** `go test ./pkg/pluginbridge/...` is executed
- **THEN** Root package facade and all subcomponent tests must pass

#### Scenario: Host plugin runtime tests pass

- **WHEN** Plugin runtime, WASM host function, and plugindb related Go tests are executed
- **THEN** Tests must pass
- **AND** No protocol behavior regressions due to import migration may occur

#### Scenario: Dynamic plugin samples can build

- **WHEN** Ordinary Go tests and `GOOS=wasip1 GOARCH=wasm` build are executed on dynamic plugin samples
- **THEN** Samples must pass compilation
- **AND** Guest-side bridge helper calls must be available
