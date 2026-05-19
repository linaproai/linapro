# Hack Tools

This directory stores repository-level development tools. Long-lived LinaPro
development commands are consolidated in `linactl` unless a focused standalone
module is still needed.

## Contents

| Directory | Purpose |
| --- | --- |
| `linactl/` | Provides the cross-platform LinaPro development command entrypoint used by Makefile and Windows `make.cmd` wrappers, including image builds, dynamic plugin `Wasm` artifact builds, and runtime `i18n` governance checks. |

## Placement Rules

- Put standalone development tools under `hack/tools/` when they are invoked through commands such as `go run`, keep their own `go.mod`, or need focused internal packages.
- Prefer adding maintained repository commands to `hack/tools/linactl/internal/<component>` with a `linactl` command wrapper when the capability only serves LinaPro.
- Put short shell, `PowerShell`, or Python automation under `hack/scripts/`; long-lived verification tools should move to `hack/tools/` when Go gives stronger typing, tests, or repository integration.
- Put `Makefile` fragments under `hack/makefiles/`.
- Put verification assets and end-to-end test code under `hack/tests/`.
- Every standalone tool directory under `hack/tools/` must maintain both `README.md` and `README.zh-CN.md` with usage, options, examples, outputs, and verification notes.

## Maintenance Notes

- Keep each tool self-contained and avoid coupling tool internals back into runtime service packages.
- Update repository entry points such as `go.work`, root `Makefile`, and related tests whenever a tool path changes.
- Keep tool lifecycle guidance in each tool's own README, and avoid documenting removed command paths from this directory.
