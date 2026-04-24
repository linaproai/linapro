## 1. Test baseline and fixture cleanup

- [x] 1.1 Audit the current coverage details for `apps/lina-core/internal/service/config` and identify the low-coverage files and branches to prioritize in this iteration
- [x] 1.2 Clean up the `config` package test fixtures and add paired reset helpers for static caches, runtime snapshots, revision state, and plugin-path overrides
- [x] 1.3 If the existing implementation is unfriendly to test isolation, make the smallest possible testability cleanup without changing production semantics

## 2. Add unit tests for the low-coverage config submodules

- [x] 2.1 Add tests for `config_plugin.go` that cover the default directory, `runtime.storagePath` compatibility fallback, override application, and cleanup behavior
- [x] 2.2 Add tests for `config_public_frontend.go` that cover `PublicFrontendSettingSpecs` copy semantics, `IsProtectedConfigParam`, `ValidateProtectedConfigValue` dispatch, and time-zone parsing branches
- [x] 2.3 Add tests for `config_runtime_params_revision.go` that cover clustered revision reads, synchronization, increments, and shared-KV error propagation
- [x] 2.4 Add tests for `config_runtime_params_cache.go` that cover cache hits, rebuilds after revision changes, invalid cached value removal, exception fallback, and local TTL refresh behavior
- [x] 2.5 Add tests for the remaining low-coverage getters/helpers such as `config_jwt.go`, `config_session.go`, `config_upload.go`, `config_login.go`, and `config_metadata.go`, including default-value, empty-object, and exception branches

## 3. Coverage verification and result capture

- [x] 3.1 Run `cd apps/lina-core && go test ./internal/service/config -cover` and confirm the package-level coverage reaches `80%` or higher
- [x] 3.2 If the first run misses the target, continue filling the remaining gaps and repeat verification until the threshold is reached
- [x] 3.3 Record the final coverage result and the added test scope in the change record as review input

## 4. Current result

- `2026-04-23`: Added `config_plugin_test.go`, `config_protected_settings_test.go`, and `config_runtime_params_revision_additional_test.go` to cover critical branches in plugin config, protected config helpers, clustered revision logic, and runtime snapshot cache behavior.
- `2026-04-23`: Reused and standardized the existing `setTestConfigContent`, `resetRuntimeParamCacheTestState`, and `SetPluginDynamicStoragePathOverride` fixtures to control isolation without adding production refactors.
- `2026-04-23`: Ran `cd apps/lina-core && go test ./internal/service/config -cover` and obtained `coverage: 83.0% of statements`, meeting the `80%+` requirement for this change.
