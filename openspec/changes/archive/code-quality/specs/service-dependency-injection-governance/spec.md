## ADDED Requirements

### Requirement: Backend components must manage runtime dependencies through explicit dependency injection

The system SHALL require host and source-plugin production backend components to receive runtime dependencies through individual constructor parameters. Controllers, Middleware, Services, plugin host service adapters, and WASM host services MUST NOT implicitly create key service dependencies in business constructors, request processing, plugin callbacks, or host service call paths, and MUST NOT pass multiple interface-typed runtime dependencies through aggregate dependency structs.

#### Scenario: Service constructors receive interface dependencies individually

- **WHEN** a host service needs access to configuration, plugin, authorization, tenant, session, cache coordination, or i18n runtime dependencies
- **THEN** the constructor receives these interface-typed dependencies as individual parameters in its signature
- **AND** the constructor MUST NOT silently call other key services' `New()` to fill in missing dependencies

#### Scenario: Aggregate structs must not hide interface dependencies

- **WHEN** a backend component needs to receive multiple interface objects, service objects, or host capability adapters
- **THEN** these interface-typed dependencies MUST be split into individual constructor parameters
- **AND** they MUST NOT be passed through `Dependencies`, `Deps`, `Options`, or equivalent aggregate structs
- **AND** dependency additions, removals, or replacements MUST be exposed through Go compilation errors at all unsynchronized call sites

#### Scenario: Controller constructors receive service dependencies

- **WHEN** a host or source-plugin controller depends on one or more service components
- **THEN** the controller constructor receives these service instances through parameters
- **AND** the controller constructor MUST NOT create cache-sensitive or runtime-state-sensitive service instances on its own

#### Scenario: Request paths must not create key services temporarily

- **WHEN** an HTTP handler, middleware, plugin callback, or WASM host service is processing a runtime call
- **THEN** the path reuses dependencies injected at construction time
- **AND** the path MUST NOT temporarily call key service `New()` to create a new service graph

### Requirement: Must not bypass explicit dependencies through generic containers or global service locators

The system SHALL complete dependency management without introducing generic DI containers, global service locators, or new host-private assembly layers. Existing startup orchestration, route binding, and plugin registrar SHALL serve as explicit construction boundaries.

#### Scenario: Startup orchestration holds shared instances

- **WHEN** the HTTP runtime constructs host long-lifecycle services
- **THEN** these services are held by the existing startup orchestration structure and passed to route binding, plugin registration, and host service configuration
- **AND** business components MUST NOT query dependencies through a global registry at runtime

#### Scenario: Generic DI containers are prohibited

- **WHEN** developers design solutions for backend dependency management
- **THEN** the solution MUST NOT introduce third-party or self-developed generic DI containers
- **AND** dependency relationships MUST remain visible through Go type signatures

### Requirement: Cache-sensitive components must share runtime instances or shared backends

The system SHALL enforce sharing of the same runtime instance or shared backend for all components that hold caches, derived state, subscription state, session/token state, plugin runtime state, runtime configuration snapshots, permission snapshots, or cross-instance coordination dependencies.

#### Scenario: Middleware reuses authentication and authorization service instances

- **WHEN** host authentication, tenant, and authorization middleware are constructed
- **THEN** the middleware receives startup-constructed `auth`, `role`, `tenant`, `config`, `i18n`, `bizctx`, and `plugin` dependencies
- **AND** the middleware MUST NOT create another authentication, authorization, tenant, or plugin service graph on its own

#### Scenario: Plugin management and plugin runtime reuse the same plugin service

- **WHEN** plugin management controllers, plugin HTTP route dispatchers, plugin runtime cache, source route registrars, or dynamic plugin host services need plugin governance state
- **THEN** they reuse the same plugin service instance from startup or narrow interfaces published by that instance
- **AND** they MUST NOT create plugin service instances that hold independent enabled snapshots, route bindings, frontend bundles, runtime i18n, or revision observers

#### Scenario: Cache coordination backends remain consistent in cluster mode

- **WHEN** `cluster.enabled=true` and a component needs cachecoord, kvcache, lock, session hot state, or token state
- **THEN** the component uses the startup-injected coordination-backed service or the same shared coordination backend
- **AND** it MUST NOT fall back to a local default instance visible only to the current node

### Requirement: Source plugins must obtain host capabilities through host-published dependencies

The system SHALL provide source plugins with a stable host service directory through the source-plugin registrar or equivalent host-published context. Source-plugin Controllers and Services MUST receive host capability adapters through this directory, and MUST NOT construct host internal service graphs on their own in plugin production paths.

#### Scenario: Source plugin registers HTTP routes

- **WHEN** a source plugin constructs controllers and services in the `http.route.register` callback
- **THEN** the plugin obtains `bizctx`, `config`, `i18n`, `notify`, `auth`, `session`, `pluginstate`, and other host capabilities from the registrar-exposed host service directory
- **AND** plugin business services receive these capabilities through explicit dependency injection

#### Scenario: Plugin host service adapters are constructed by the host

- **WHEN** source plugins use host capabilities published by `pkg/pluginservice/*`
- **THEN** adapter instances are constructed by the host runtime and passed through the registrar
- **AND** plugin production paths MUST NOT call parameterless adapter constructors to create isolated host service graphs

### Requirement: Initialization and registration APIs must return errors for caller decision-making

The system SHALL require host and source-plugin runtime initialization, source-plugin registration, registrar, callback registration, route registration, Cron registration, and middleware registration APIs to return `error` when dependencies are missing, registration parameters are invalid, configuration sources are missing, backend creation fails, or validation fails. These APIs MUST NOT internally `panic` to handle expected errors; whether to abort the process, ignore, or degrade MUST be explicitly decided by the topmost entry point in the call stack.

#### Scenario: Source-plugin registration API returns errors

- **WHEN** a source plugin declares an invalid extension point, invalid execution mode, nil callback, or duplicate registration
- **THEN** the `pluginhost` registration API returns `error`
- **AND** the API MUST NOT internally `panic`

#### Scenario: Top-level static registration entry chooses failure exit

- **WHEN** a source-plugin package-level `init` calls a registration API and receives an error
- **THEN** the top-level static registration entry MAY explicitly `panic`
- **AND** the panic governance scan allowlist MUST record this as a top-level entry receiving an error and choosing to exit

#### Scenario: Runtime callbacks missing host dependencies

- **WHEN** HTTP, Cron, Hook, or middleware registration callbacks discover missing host-published dependencies during execution
- **THEN** the callback returns `error`
- **AND** the host caller decides whether to block startup, record the failure, or execute another degradation strategy

### Requirement: Dependency injection rules must be included in project specifications and lina-review

The system SHALL write explicit dependency injection, implicit construction prohibition, initialization/registration error return, and cache-sensitive shared instance requirements into project specifications and `lina-review` review standards. Reviews MUST cover host, source plugins, plugin host services, WASM host services, and test verification.

#### Scenario: Reviewing backend implementation changes

- **WHEN** `lina-review` reviews any backend Go change
- **THEN** the review checks whether newly added or modified components manage runtime dependencies through explicit dependency injection
- **AND** the review flags implicit key service construction in production paths

#### Scenario: Reviewing initialization and registration error handling

- **WHEN** `lina-review` reviews runtime initialization, source-plugin registration, registrar, callback registration, or startup assembly changes
- **THEN** the review confirms expected failures are returned as `error` to the caller
- **AND** the review flags APIs that internally `panic` to handle expected errors

#### Scenario: Reviewing aggregate interface dependency structs

- **WHEN** `lina-review` reviews constructor or dependency injection designs
- **THEN** the review flags implementations that pass multiple interface-typed runtime dependencies through aggregate structs
- **AND** the review requires interface-typed dependencies to be split into individual constructor parameters

#### Scenario: Reviewing cache-sensitive components

- **WHEN** `lina-review` reviews changes involving authentication, authorization, session, plugin, configuration, i18n, cachecoord, kvcache, lock, notify, or host services
- **THEN** the review requires explanation of how shared instances or shared backends ensure state consistency
- **AND** if the change has no cache impact, the review conclusion MUST explicitly state this

#### Scenario: Static scanning prevents regression

- **WHEN** change verification completes
- **THEN** the project executes static scanning or equivalent governance verification to identify key service `New()` calls outside test files and startup construction boundaries
- **AND** any newly added violating calls MUST be fixed or recorded with explicit exemption reasons
