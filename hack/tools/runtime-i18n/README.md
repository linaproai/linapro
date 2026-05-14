# Runtime I18n Tool

`runtime-i18n` provides repository-level runtime i18n verification. It scans high-risk source-code locations for hard-coded runtime-visible copy, reports generated/test source statistics, and validates runtime i18n key coverage across host and plugin scopes.

## Usage

Preferred repository entry points:

```bash
make i18n.check
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

`make i18n.check` runs both direct tool commands.

The `scan` command blocks on non-allowlisted runtime-source findings. It also reports non-blocking statistics for generated source and test fixtures so review records can distinguish source-code violations from accepted generated/test data.

## Scan Options

| Option | Default | Description |
| --- | --- | --- |
| `--format` | `text` | Output format. Supported values: `text`, `json`. |
| `--allowlist` | `hack/tools/runtime-i18n/allowlist.json` | JSON allowlist file used to document accepted findings. |

JSON output uses a structured report:

```json
{
  "summary": {
    "violations": 0,
    "violationFiles": 0,
    "allowlistHits": 0,
    "generatedFiles": 0,
    "generatedItems": 0,
    "testFixtureFiles": 0,
    "testFixtureItems": 0,
    "byCategory": {}
  },
  "findings": [],
  "allowlistHits": []
}
```

## Allowlist Format

Each accepted finding must be documented with a path, rule, category, reason, and scope. Use `line` only when the acceptance is intentionally limited to one source line.

```json
{
  "entries": [
    {
      "path": "apps/lina-core/internal/service/example/example.go",
      "rule": "go-string-literal-han",
      "category": "Unclassified",
      "reason": "Stable user data fixture that is not rendered as system copy.",
      "scope": "single fixture value",
      "line": 42
    }
  ]
}
```

## Exit Codes

- `0`: The selected check passed.
- `1`: The tool could not run, the selected check found issues, or invalid arguments were provided.

When invoked through `make`, GNU Make reports the non-zero tool exit as a Makefile failure.

## Notes

- Runtime JSON checks read direct files under `manifest/i18n/<locale>/*.json` and do not recurse into `apidoc/`.
- Every allowlist entry must include the path, rule, category, reason, and scope.
