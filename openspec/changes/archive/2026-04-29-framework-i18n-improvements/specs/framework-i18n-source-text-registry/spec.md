## ADDED Requirements

### Requirement: Host must provide explicit registration mechanism for code-owned source text namespaces
The host system SHALL provide a `RegisterSourceTextNamespace(prefix, reason string)` registration function and corresponding read-only query capability in the `internal/service/i18n` package. Business modules MUST explicitly register their code-owned source text namespaces (e.g., `job.handler.`, `job.group.default.`) in their own `init()`. The `i18n` package MUST NOT hardcode any specific business module's namespace prefix in its own implementation. Missing translation checks, override source diagnostics, and import/export SHALL identify "namespaces whose translation keys are owned by code sources" by querying this registry.

#### Scenario: Business modules register code-owned namespaces via init
- **WHEN** the `jobmgmt` package executes `init()` at project startup
- **THEN** the package registers its namespace via `i18n.RegisterSourceTextNamespace("job.handler.", "code-owned cron handler display")`
- **AND** missing checks can identify these keys as code-owned without modifying the `i18n` package source

#### Scenario: Missing checks exempt code-owned namespaces based on registry
- **WHEN** the system calls `CheckMissingMessages` for any non-default target language (e.g., `en-US` or `zh-TW`) and some keys belong to registered code-owned namespaces
- **THEN** these keys do not appear in missing results
- **AND** the display fallback for these keys is handled by the owning module's code source text, without requiring each target language to redundantly maintain JSON keys

### Requirement: i18n package must no longer depend on specific business module namespace prefixes
The host system SHALL NOT allow any function in the `i18n` package (including helper determinations like `isSourceTextBackedRuntimeKey`) to hardcode `job.handler.`, `job.group.default.`, or other specific business module namespace prefixes. All determinations of "this key is owned by a code source" MUST be obtained by querying the namespace registry.

#### Scenario: Delete reverse dependency on jobmgmt in i18n package
- **WHEN** reviewing any source file in `apps/lina-core/internal/service/i18n/`
- **THEN** no hardcoded strings with business-module-specific prefixes like `job.handler.` or `job.group.default.` exist
- **AND** the file instead uses the namespace registry's query interface for determination
