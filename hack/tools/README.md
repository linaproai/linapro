# Hack Tools

This directory stores repository-level development tools that are implemented as standalone executables.

## Contents

| Directory | Purpose |
| --- | --- |
| `build-wasm/` | Builds dynamic plugin `Wasm` runtime artifacts from source plugins. |
| `runtime-i18n/` | Scans runtime-visible hard-coded copy and validates host/plugin i18n key coverage. |
| `upgrade-source/` | Runs repository-level framework and source-plugin upgrade flows. |

## Placement Rules

- Put standalone development tools under `hack/tools/` when they are invoked through commands such as `go run`, keep their own `go.mod`, or need focused internal packages.
- Put short shell, `PowerShell`, or Python automation under `hack/scripts/`; long-lived verification tools should move to `hack/tools/` when Go gives stronger typing, tests, or repository integration.
- Put `Makefile` fragments under `hack/makefiles/`.
- Put verification assets and end-to-end test code under `hack/tests/`.
- Every tool directory under `hack/tools/` must maintain both `README.md` and `README.zh_CN.md` with usage, options, examples, outputs, and verification notes.

## Maintenance Notes

- Keep each tool self-contained and avoid coupling tool internals back into runtime service packages.
- Update repository entry points such as `go.work`, root `Makefile`, and related tests whenever a tool path changes.
