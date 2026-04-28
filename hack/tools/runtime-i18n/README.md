# Runtime I18n Tool

`runtime-i18n` provides repository-level runtime i18n verification. It scans high-risk source-code locations for hard-coded runtime-visible copy and validates runtime i18n key coverage across host and plugin scopes.

## Usage

Preferred repository entry points:

```bash
make check-runtime-i18n
make check-runtime-i18n-messages
```

Direct tool invocation:

```bash
go run ./hack/tools/runtime-i18n scan
go run ./hack/tools/runtime-i18n scan --format json
go run ./hack/tools/runtime-i18n messages
```

## Commands

| Command | Description |
| --- | --- |
| `scan` | Scans Go, Vue, and TypeScript files for high-risk runtime-visible hard-coded copy. |
| `messages` | Validates host and plugin runtime i18n JSON key coverage and duplicate runtime keys. |

## Scan Options

| Option | Default | Description |
| --- | --- | --- |
| `--format` | `text` | Output format. Supported values: `text`, `json`. |
| `--allowlist` | `hack/tools/runtime-i18n/allowlist.json` | JSON allowlist file used to document accepted findings. |

## Exit Codes

- `0`: The selected check passed.
- `1`: The tool could not run, the selected check found issues, or invalid arguments were provided.

When invoked through `make`, GNU Make reports the non-zero tool exit as a Makefile failure.

## Notes

- Runtime JSON checks read direct files under `manifest/i18n/<locale>/*.json` and do not recurse into `apidoc/`.
- The current runtime copy governance cleanup is still in progress, so `scan` may intentionally report existing findings until the related modules are cleaned.
- Every allowlist entry must include the path, rule, category, and reason.
