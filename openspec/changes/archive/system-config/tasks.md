## 1. Protected Runtime Configuration

- [x] 1.1 Register built-in runtime parameters for JWT expiry, session timeout, upload size, and login IP blacklist.
- [x] 1.2 Register protected public frontend settings for branding, login-page copy, theme mode, layout, and watermark behavior.
- [x] 1.3 Add value validation, protected-key safeguards, SQL seed metadata, and import/update protection for host-owned settings.
- [x] 1.4 Wire runtime configuration into host behavior for authentication, online sessions, file upload, and frontend bootstrap.

## 2. Runtime Cache and Cluster Strategy

- [x] 2.1 Add process-local runtime snapshot caches for protected configuration reads.
- [x] 2.2 Add shared revision synchronization for multi-instance propagation.
- [x] 2.3 Optimize single-node mode to skip unnecessary shared-KV and watcher behavior.
- [x] 2.4 Refactor cache and watcher control flow toward constructor-time strategy wiring and add supporting tests.

## 3. Upload Size Unification

- [x] 3.1 Change the host initialization default for `sys.upload.maxSize` to 20 MB and update any related manifest or derived artifacts.
- [x] 3.2 Update the upload config template and backend static fallback default to 20 MB so the current 10 MB / 16 MB split disappears.
- [x] 3.3 Update file-upload validation, request-body size protection, and friendly error-message logic or assertions so the default baseline is consistently 20 MB.
- [x] 3.4 Update the affected backend automated tests to cover both the default 20 MB case and runtime override cases.
- [x] 3.5 Run the affected backend tests and any required initialization checks to confirm the initial value, runtime enforcement, and error messages all use a 20 MB baseline.
- [x] 3.6 If the build flow produces embedded or packaged manifest artifacts, verify that those artifacts also use the updated 20 MB default.

## 4. Plugin Governance and API Documentation

- [x] 4.1 Replace direct reliance on GoFrame's built-in `/api.json` output with a host-managed OpenAPI builder.
- [x] 4.2 Capture source-plugin route ownership and DTO documentation metadata at registration time through a host-observable route facade.
- [x] 4.3 Keep source-plugin middleware registration plugin-owned and unrestricted by route-prefix rules.
- [x] 4.4 Project enabled source-plugin routes and enabled dynamic-plugin routes into the host-managed API document.
- [x] 4.5 Exclude internal and non-business routes from the system API document.

## 5. Plugin UI and Operational Follow-up

- [x] 5.1 Add plugin detail dialog support and refine host-service presentation semantics in the default admin workspace.
- [x] 5.2 Improve plugin resource grouping, labels, empty-state behavior, and layout consistency between detail and authorization dialogs.
- [x] 5.3 Add pagination behavior for the dynamic plugin demo record list.
- [x] 5.4 Add structured logging switch support and align HTTP server logs with business log sinks.
- [x] 5.5 Move host-specific server and logger extensions under explicit `extensions` namespaces.

## 6. Config Management Unit Test Coverage

- [x] 6.1 Audit the current coverage details for `apps/lina-core/internal/service/config` and identify the low-coverage files and branches to prioritize in this iteration.
- [x] 6.2 Clean up the `config` package test fixtures and add paired reset helpers for static caches, runtime snapshots, revision state, and plugin-path overrides.
- [x] 6.3 If the existing implementation is unfriendly to test isolation, make the smallest possible testability cleanup without changing production semantics.
- [x] 6.4 Add tests for `config_plugin.go` that cover the default directory, `runtime.storagePath` compatibility fallback, override application, and cleanup behavior.
- [x] 6.5 Add tests for `config_public_frontend.go` that cover `PublicFrontendSettingSpecs` copy semantics, `IsProtectedConfigParam`, `ValidateProtectedConfigValue` dispatch, and time-zone parsing branches.
- [x] 6.6 Add tests for `config_runtime_params_revision.go` that cover clustered revision reads, synchronization, increments, and shared-KV error propagation.
- [x] 6.7 Add tests for `config_runtime_params_cache.go` that cover cache hits, rebuilds after revision changes, invalid cached value removal, exception fallback, and local TTL refresh behavior.
- [x] 6.8 Add tests for the remaining low-coverage getters/helpers such as `config_jwt.go`, `config_session.go`, `config_upload.go`, `config_login.go`, and `config_metadata.go`, including default-value, empty-object, and exception branches.
- [x] 6.9 Run `cd apps/lina-core && go test ./internal/service/config -cover` and confirm the package-level coverage reaches 80% or higher.
- [x] 6.10 If the first run misses the target, continue filling the remaining gaps and repeat verification until the threshold is reached.
- [x] 6.11 Record the final coverage result (83.0%) and the added test scope in the change record as review input.

## 7. Comment Conformance and Review

- [x] 7.1 Review comment coverage for configuration services, permission caches, command boot paths, host-managed OpenAPI code, source-plugin route capture code, and related tests.
- [x] 7.2 Complete backend comment-conformance cleanup across the affected host services and plugin backend samples.
- [x] 7.3 Run targeted Go verification for the touched host packages and plugin backend packages.
- [x] 7.4 Run an archive-time OpenSpec review and confirm that no critical issues remain.
