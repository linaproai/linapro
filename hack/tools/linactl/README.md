# linactl

`linactl` is LinaPro's cross-platform development command entrypoint. It keeps the repository's long-lived task orchestration in Go so Windows, Linux, and macOS can run the same commands without depending on GNU Make or POSIX shell tools.

## Usage

```bash
go run ./hack/tools/linactl help
go run ./hack/tools/linactl status
go run ./hack/tools/linactl prepare-packed-assets
go run ./hack/tools/linactl wasm p=plugin-demo-dynamic
go run ./hack/tools/linactl init confirm=init
go run ./hack/tools/linactl build platforms=linux/amd64,linux/arm64
```

## Windows Entry

The repository root also provides `make.cmd` as a thin Windows wrapper:

```cmd
make.cmd help
make.cmd status
make.cmd init confirm=init
```

In PowerShell, run it with an explicit current-directory prefix:

```powershell
.\make.cmd help
.\make.cmd status
```

## Parameters

`linactl` accepts the existing make-style `key=value` arguments to keep command migration low-friction.

| Parameter | Example | Purpose |
| --- | --- | --- |
| `confirm` | `confirm=init` | Confirms destructive bootstrap commands. |
| `rebuild` | `rebuild=true` | Rebuilds the configured database during `init`. |
| `platforms` | `platforms=linux/amd64,linux/arm64` | Selects build target platforms. |
| `p` | `p=plugin-demo-dynamic` | Builds a specific dynamic plugin. |
| `verbose` | `verbose=1` | Shows child command output for build tasks. |

## Verification

```bash
go test ./hack/tools/linactl
go run ./hack/tools/linactl help
go run ./hack/tools/linactl wasm dry-run=true
```

