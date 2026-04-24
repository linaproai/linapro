# Hack Tools

This directory stores repository-level development tools that are implemented as standalone executables.

## Contents

| Directory | Purpose |
| --- | --- |
| `build-wasm/` | Builds dynamic plugin `Wasm` runtime artifacts from source plugins. |
| `upgrade-source/` | Runs repository-level framework and source-plugin upgrade flows. |

## Placement Rules

- Put standalone development tools under `hack/tools/` when they are invoked through commands such as `go run`, keep their own `go.mod`, or need focused internal packages.
- Put shell, `PowerShell`, or Python automation under `hack/scripts/`.
- Put `Makefile` fragments under `hack/makefiles/`.
- Put verification assets and end-to-end test code under `hack/tests/`.

## Maintenance Notes

- Keep each tool self-contained and avoid coupling tool internals back into runtime service packages.
- Update repository entry points such as `go.work`, root `Makefile`, and related tests whenever a tool path changes.
