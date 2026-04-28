## 1. Define Panic Usage Boundaries

- [x] 1.1 Establish a production backend `panic` allowlist and document retained reasons such as startup, registration, `Must*`, or unknown panic rethrow paths.
- [x] 1.2 Add a static check script or test that scans production Go files and blocks `panic` calls outside the allowlist.

## 2. Convert Runtime Paths to Error Returns

- [x] 2.1 Convert unnecessary `panic` calls in Excel cell coordinate and file-closing helpers into explicit error returns.
- [x] 2.2 Split dynamic plugin hostServices normalization into an error-returning runtime path and a necessary Must path.
- [x] 2.3 Return explicit `error` values for runtime configuration parsing failures instead of silently degrading around the rule.

## 3. Tests and Verification

- [x] 3.1 Add Go unit tests for Excel helpers, invalid hostServices input, invalid runtime configuration values, and the panic allowlist check.
- [x] 3.2 Run affected backend package tests and OpenSpec validation, and confirm that no i18n resource changes are required.
- [x] 3.3 Run `lina-review` for this change and resolve the findings.

## Feedback

- [x] **FB-1**: Cron runtime configuration reads should return explicit errors instead of degrading through logs.
- [x] **FB-2**: `closeutil` and `excelutil` close-error logs should explain nil error pointer misuse and receive the caller context.
- [x] **FB-3**: Project rules and `lina-review` should require logging calls to propagate `ctx` through the call chain to preserve tracing.
- [x] **FB-4**: The panic allowlist check should move from the `lina-core` root into the `internal/cmd` test directory and should not treat test helpers as production panic boundaries.
- [x] **FB-5**: The panic allowlist test should reduce coupling to custom string concatenation and scanning logic to improve maintainability.
