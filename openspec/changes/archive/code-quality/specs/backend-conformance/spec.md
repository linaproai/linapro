## MODIFIED Requirements

### Requirement: Controller and service layer implementation constraints

Backend production code SHALL follow the GoFrame v2 layered conventions defined by the repository: controller dependencies are injected through constructor functions, and service components are organized by convention directories and naming. Controllers, services, Middleware, plugin adapters, and other runtime components SHALL receive key dependencies through explicit dependency injection and MUST NOT implicitly create cache-sensitive or runtime-state-sensitive service instances in business constructors or request paths.

#### Scenario: Controller dependency initialization

- **WHEN** a controller depends on one or more service components
- **THEN** those dependencies are explicitly passed in by the caller through the corresponding `_new.go` constructor function
- **AND** controller constructors and interface methods do not internally call key `service.New()` to create dependencies

#### Scenario: Service component file splitting

- **WHEN** a service component has multiple responsibility sub-modules
- **THEN** code is split into independent files by component prefix and sub-module suffix
- **AND** bare filenames unrelated to the component name are not used to host sub-module logic

#### Scenario: Service layer dependencies explicitly passed in

- **WHEN** a service component depends on configuration, authentication, authorization, tenant, plugin, i18n, session, cache coordination, KV cache, lock, or notification capabilities
- **THEN** the service constructor receives these interface-typed dependencies through individual parameters
- **AND** the service constructor MUST NOT create another equivalent runtime service graph internally
- **AND** the service constructor MUST NOT receive multiple interface-typed runtime dependencies through aggregate structs such as `Dependencies`, `Deps`, `Options`, or equivalents

#### Scenario: Middleware shares runtime dependencies

- **WHEN** HTTP Middleware is constructed for authentication, tenant, authorization, response, context, or request size control
- **THEN** the Middleware uses shared service instances passed in during startup
- **AND** the Middleware MUST NOT re-create authentication, configuration, i18n, plugin, role, or tenant services in its own constructor

#### Scenario: Initialization and registration errors returned to caller

- **WHEN** backend runtime initialization, source-plugin registration, registrar, callback registration, route/Cron/middleware registration, or startup assembly helper encounters missing dependencies, invalid registration parameters, missing configuration sources, backend creation failures, or validation failures
- **THEN** the API returns `error` to the caller
- **AND** the API MUST NOT internally `panic` to handle expected errors
- **AND** only the topmost entry point in the call stack may choose to abort the process after explicitly handling the error

#### Scenario: Stateless local construction must be auditable

- **WHEN** a backend component genuinely needs to locally construct a stateless helper or pure DAO query service
- **THEN** the construction MUST NOT hold caches, subscriptions, sessions, tokens, plugin state, runtime configuration, or cross-instance coordination state
- **AND** the review conclusion MUST be able to explain that the construction does not affect runtime consistency

### Requirement: ORM and soft-delete conformance

Backend production code SHALL use GoFrame-recommended ORM patterns for database access and follow automatic soft-delete and timestamp maintenance conventions.

#### Scenario: Querying soft-delete tables

- **WHEN** code queries a table that contains a `deleted_at` field
- **THEN** the query logic relies on GoFrame automatic soft-delete filtering
- **AND** production code does not hand-write `WhereNull(deleted_at)` or equivalent SQL conditions

#### Scenario: Updating and writing data

- **WHEN** code performs database writes, updates, or association maintenance
- **THEN** production code uses DO objects to pass `Data`
- **AND** does not manually maintain `created_at`, `updated_at`, or `deleted_at` fields that are handled by the framework

### Requirement: Exported symbol documentation completeness

Backend exported methods, structs, and key public fields SHALL carry comments that follow Go documentation conventions, suitable for doc generation and long-term maintenance. Interface method comments MUST describe function, key inputs, outputs, error returns, and applicable constraints; they MUST NOT only repeat the method name or only supplement implementation methods.

#### Scenario: Adding or modifying exported symbols

- **WHEN** code contains exported methods, exported structs, or key exported fields
- **THEN** their declarations are preceded by adjacent, semantically clear comments
- **AND** comments are recognizable by Go doc, not just separator remarks or detached notes

#### Scenario: Interface method comments describe usage contracts

- **WHEN** backend production code declares or refactors any `interface` method
- **THEN** each method declaration has an adjacent comment describing the method's function and side effects
- **AND** the comment describes key input parameter semantics, return value meanings, and empty-result or zero-value semantics where applicable
- **AND** the comment describes applicable business errors, permission errors, data-permission rejections, configuration errors, or underlying errors

#### Scenario: Interface method comments cover important constraints

- **WHEN** an interface method involves permissions, data permissions, tenant isolation, caching, i18n, transactions, idempotency, concurrency, or external resource access
- **THEN** the method comment explicitly describes constraints, visibility boundaries, cache consistency, or failure handling semantics the caller needs to understand
- **AND** the caller does not need to read the implementation to learn the method's main usage considerations

### Requirement: Runtime errors must not replace explicit error handling with panic

Production backend code SHALL use `panic` only for startup, initialization, unrecoverable critical paths, `Must*` semantic constructors, or unknown panic rethrow scenarios. Ordinary requests, import/export flows, dynamic plugin input, runtime configuration reads, and recoverable resource handling paths MUST use explicit `error` returns, unified error responses, or controlled degradation.

#### Scenario: Startup unrecoverable errors use fail-fast

- **WHEN** the backend detects an unrecoverable error during process startup, driver registration, command tree initialization, or source-plugin static registration
- **THEN** the code MAY use `panic` to fail the process fast
- **AND** the panic call site MUST be in the allowlist with a reason for retaining it

#### Scenario: Ordinary business requests return errors

- **WHEN** an ordinary HTTP request, file import/export, Excel generation, or resource close operation encounters a recoverable error
- **THEN** the service or controller MUST return `error` so the unified error handling chain can generate the response
- **AND** it MUST NOT use `panic` instead of returning the error

#### Scenario: Dynamic plugin input validation fails

- **WHEN** a dynamic plugin artifact, manifest, hostServices declaration, or authorization input is invalid
- **THEN** the host MUST return a validation error with context
- **AND** plugin-provided dynamic input MUST NOT trigger a production-code panic

#### Scenario: Invalid runtime configuration values return explicitly

- **WHEN** a protected runtime configuration value has a parsing error while a snapshot is being read
- **THEN** the backend MUST expose the configuration problem through an explicit `error` return or unified error response
- **AND** write paths MUST still keep strict validation so normal management entries cannot save invalid values

#### Scenario: New panics are constrained by static checks

- **WHEN** a developer adds a `panic` call in production backend Go code
- **THEN** automated checks MUST require the call site to match the allowlist
- **AND** the allowlist entry MUST document its category and retained reason

## ADDED Requirements

### Requirement: Public component main files must serve as stable contract entry points

`apps/lina-core/pkg` public components SHALL use main-file responsibility governance consistent with service components. `pkg/<component>/<component>.go` main files MUST serve as public component contract entry points, retaining public types, interfaces, constructors, and lightweight contract methods; complex implementation logic MUST migrate to responsibility files in the same package.

#### Scenario: Public component main files stay concise

- **WHEN** a developer adds or refactors `apps/lina-core/pkg/<component>/<component>.go`
- **THEN** the main file retains public component comments, exported types, public interfaces, constructors, and necessary lightweight methods
- **AND** parsing, encoding/decoding, validation, database access, runtime execution, bridge adaptation, and complex business logic migrate to other files in the same package

#### Scenario: Public component migration preserves external contracts

- **WHEN** public component implementation logic migrates out of the main file
- **THEN** package name, exported symbols, function signatures, error semantics, and calling behavior remain unchanged
- **AND** host and source-plugin callers do not need to modify import paths or calling conventions

### Requirement: Backend source file top comments must describe file purpose and caveats

All backend production Go source files SHALL provide comments at the top of the file sufficient to help developers understand the file's purpose. Main file comments MUST describe component-level responsibilities, boundaries, and reading entry points; non-main file comments MUST describe the implementation slice the file carries, its main logic, and caveats.

#### Scenario: Main files use component-level comments

- **WHEN** a backend main file declares its package
- **THEN** the package comment is adjacent to the `package` declaration
- **AND** the comment describes the component's overall responsibilities, main capability boundaries, and dependencies or key constraints callers need to know

#### Scenario: Non-main files use file-responsibility comments

- **WHEN** a backend non-main file declares its package
- **THEN** the file top comment has one blank line before the `package` declaration
- **AND** the comment describes the implementation slice the file carries, its main flow, and caveats
- **AND** the comment does not use the main file's component-level description as its sole content

#### Scenario: Comment quality avoids vagueness and line-by-line repetition

- **WHEN** a developer adds or refactors file top comments
- **THEN** the comment does not consist of a single generic sentence that cannot distinguish the file's responsibility
- **AND** the comment does not line-by-line repeat the implementation code
- **AND** the comment focuses on why the file exists, what logic it handles, and what important constraints apply

### Requirement: Backend source readability governance must be included in lina-review

`lina-review` SHALL include backend main-file responsibility, interface method comment completeness, file-level comment quality, and batch verification records in its review standards. Any backend Go change MUST undergo these checks; large-scale refactoring MUST proceed per-module in batches with per-batch verification evidence.

#### Scenario: Reviewing main-file responsibility

- **WHEN** `lina-review` reviews backend Go changes
- **THEN** the review identifies whether newly added or modified `internal/service` main files and `lina-core/pkg` main files only serve as contract entry points
- **AND** the review flags new or remaining complex business implementation logic in main files

#### Scenario: Reviewing interface method comments

- **WHEN** `lina-review` reviews newly added or modified backend interface definitions
- **THEN** the review confirms each interface method has an adjacent comment
- **AND** the review confirms complex method comments cover function, inputs, outputs, errors, and key constraints where applicable

#### Scenario: Reviewing file-level comments

- **WHEN** `lina-review` reviews newly added or modified backend Go files
- **THEN** the review confirms file-level comments describe file purpose, main logic, and caveats
- **AND** the review confirms main files and non-main files follow different comment-level responsibilities

#### Scenario: Reviewing batch refactoring verification

- **WHEN** a backend source readability refactoring task completes a module batch
- **THEN** the review confirms the task recorded the batch's modification scope, behavior-unchanged judgment, i18n impact, cache consistency impact, data permission impact, and Go compilation gate results
- **AND** batches without corresponding verification evidence MUST NOT be marked complete

#### Scenario: Reviewing linactl command file naming

- **WHEN** `lina-review` reviews newly added or modified command implementations under `hack/tools/linactl/`
- **THEN** the review confirms specific command implementation files are named `command_<command>.go` preserving dot-segment command semantics
- **AND** the review confirms commands that conflict with Go toolchain file suffix rules (e.g., `test`, `wasm`) use documented command-specific suffixes
- **AND** the review flags implementations where multiple unrelated commands are mixed into `command_ops.go` or similar catch-all files

#### Scenario: Reviewing linactl sub-component organization

- **WHEN** `lina-review` reviews newly added or modified shared implementation logic under `hack/tools/linactl/`
- **THEN** the review confirms complex implementations have migrated to `internal/<component>/` sub-packages and are referenced through package interfaces
- **AND** the review confirms newly added or remaining non-command files in the root directory belong to command registration, startup assembly, base types, or platform adaptation boundaries
- **AND** the review flags root directory files that continue to carry development services, plugin workspaces, GoFrame CLI, frontend dependencies, Playwright, image building, repository governance scanning, or file system tools
