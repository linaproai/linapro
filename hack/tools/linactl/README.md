# linactl

`linactl` is LinaPro's cross-platform development command entrypoint. It keeps the repository's long-lived task orchestration in Go so Windows, Linux, and macOS can run the same commands without depending on GNU Make or POSIX shell tools.

## Usage

```bash
cd hack/tools/linactl
go run . help
go run . status
go run . pack.assets
go run . wasm p=plugin-demo-dynamic
go run . plugins.status
go run . i18n.check
go run . init confirm=init
go run . tidy
go run . build platforms=linux/amd64,linux/arm64
go run . release.tag.check tag=v0.2.0
go run . release.tag.check print-version=1
```

## Windows Entry

The repository root also provides `make.cmd` as a thin Windows wrapper:

```cmd
make.cmd help
make.cmd status
make.cmd pack.assets
make.cmd plugins.status
make.cmd i18n.check
make.cmd init confirm=init
make.cmd tidy
make.cmd release.tag.check tag=v0.2.0
```

In PowerShell, run it with an explicit current-directory prefix:

```powershell
.\make.cmd help
.\make.cmd status
.\make.cmd pack.assets
.\make.cmd i18n.check
.\make.cmd release.tag.check tag=v0.2.0
```

## Parameters

`linactl` accepts the existing make-style `key=value` arguments to keep command migration low-friction.

| Parameter | Example | Purpose |
| --- | --- | --- |
| `confirm` | `confirm=init` | Confirms destructive bootstrap commands. |
| `rebuild` | `rebuild=true` | Rebuilds the configured database during `init`. |
| `platforms` | `platforms=linux/amd64,linux/arm64` | Selects build target platforms. |
| `plugins` | `plugins=0` | Overrides automatic plugin-full detection for build, dev, image, and Go test commands. |
| `tag` | `tag=v0.2.0` | Selects the release tag checked by `release.tag.check`. |
| `print-version` | `print-version=1` | Prints the validated `framework.version` for release automation. |
| `p` | `p=multi-tenant` | Selects one plugin for Wasm build or plugin workspace management commands. |
| `source` | `source=official` | Selects one configured plugin source for plugin workspace management commands. |
| `force` | `force=1` | Allows plugin install/update commands to overwrite existing or dirty plugin directories. |
| `verbose` | `verbose=1` | Shows child command output for build tasks. |

When `plugins` is omitted, build and dev commands enable plugin-full mode if `apps/lina-plugins` contains plugin manifests. Plugin-full mode generates or refreshes ignored `temp/go.work.plugins` from the host-only root `go.work`, then resolves source-plugin Go modules through `GOWORK`.

## Release Tag Check

`release.tag.check` reads `apps/lina-core/manifest/config/metadata.yaml` and verifies that the release tag exactly matches `framework.version`.

```bash
make.cmd release.tag.check tag=v0.2.0
make release.tag.check tag=v0.2.0
make release.tag.check metadata=apps/lina-core/manifest/config/metadata.yaml tag=v0.2.0
```

In GitHub Actions, the command also accepts `GITHUB_REF_NAME` as the tag source when `tag` is omitted.

## Plugin Workspace Commands

Plugin workspace management always uses the fixed `apps/lina-plugins` directory. Configure sources in `hack/config.yaml`:

```yaml
plugins:
  sources:
    official:
      repo: "https://github.com/linaproai/official-plugins.git"
      root: "."
      ref: "main"
      items:
        - "multi-tenant"
        - "org-center"
```

`items` only accepts plugin ID strings. Use the quoted string `"*"` to install every plugin directory directly under the source `root`; do not write bare `- *` because YAML treats it as alias syntax. If plugins from the same repository need different refs, split them into separate sources.

Common commands:

```bash
make plugins.init
make plugins.install
make plugins.install p=multi-tenant
make plugins.update source=official
make plugins.update force=1
make plugins.status
```

`plugins.init` converts `apps/lina-plugins` from a submodule into a normal directory while preserving files. `plugins.install` and `plugins.update` clone configured sources into `temp/`, copy plugin directories into `apps/lina-plugins/<plugin-id>`, and update the generated `apps/lina-plugins/.linapro-plugins.lock.yaml` lock file.

## Verification

```bash
cd hack/tools/linactl
go test ./...
go run . help
go run . wasm dry-run=true
go run . plugins.status
go run . release.tag.check tag=v0.2.0
```
