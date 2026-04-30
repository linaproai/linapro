# LinaPro Installer

This directory contains the repository-backed implementation for the single LinaPro installation entry point:

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```

The hosted `/install.sh` content must match `hack/scripts/install/bootstrap.sh`. The bootstrap script is self-contained: it resolves the target version, clones the requested tag, and dispatches to the platform script that ships inside the cloned repository.

## Supported Platforms

| Platform | Runtime |
| --- | --- |
| `macOS` | `bash` on Darwin |
| `Linux` | `bash` on Linux distributions and WSL |
| `Windows` | Git Bash or WSL only |

Windows users must run the command from Git Bash or WSL. Native PowerShell and `cmd.exe` are not supported entry points.

## Directory Layout

```text
hack/scripts/install/
  bootstrap.sh          hosted curl|bash entrypoint
  install-macos.sh      macOS post-clone setup
  install-linux.sh      Linux and WSL post-clone setup
  install-windows.sh    Windows Git Bash post-clone setup
  checks/prereq.sh      shared prerequisite checks
  lib/_common.sh        shared installer helpers
  README.md             English documentation
  README.zh_CN.md       Simplified Chinese mirror
```

## Environment Variables

| Variable | Default | Meaning | Example |
| --- | --- | --- | --- |
| `LINAPRO_VERSION` | Latest stable GitHub release | Target version tag to clone. The installer fails if it cannot resolve a tag automatically. | `LINAPRO_VERSION=v0.5.0 curl -fsSL https://linapro.ai/install.sh \| bash` |
| `LINAPRO_DIR` | `./linapro` | Target directory for the cloned project. | `LINAPRO_DIR=~/Workspace/my-linapro curl -fsSL https://linapro.ai/install.sh \| bash` |
| `LINAPRO_NON_INTERACTIVE` | unset | Skips interactive confirmations where a platform script needs one. | `LINAPRO_NON_INTERACTIVE=1 ...` |
| `LINAPRO_SKIP_MOCK` | unset | Runs `make init` but skips `make mock`. | `LINAPRO_SKIP_MOCK=1 ...` |
| `LINAPRO_SHALLOW` | unset | Uses `git clone --depth 1`. The first upgrade later requires `git fetch --unshallow`. | `LINAPRO_SHALLOW=1 ...` |

`LINAPRO_FORCE=1` is a hidden recovery switch for intentionally replacing a non-empty target directory.

## Local Equivalent

From an existing repository checkout, run the same bootstrap source locally:

```bash
bash hack/scripts/install/bootstrap.sh
```

The command still clones the requested version into `LINAPRO_DIR` or `./linapro`; it does not install over the current checkout unless you explicitly set `LINAPRO_DIR`.

## What the Installer Does

1. Resolves `LINAPRO_VERSION` or follows the GitHub `releases/latest` redirect.
2. Refuses to overwrite a non-empty target directory unless `LINAPRO_FORCE=1` is set.
3. Runs `git clone --branch <tag> https://github.com/linaproai/linapro.git "$LINAPRO_DIR"`.
4. Dispatches to `install-macos.sh`, `install-linux.sh`, or `install-windows.sh`.
5. Checks `go >= 1.22`, `node >= 20`, `pnpm >= 8`, `git`, `make`, the MySQL client, and ports `5666` / `8080`.
6. Runs backend `go mod download`, frontend `pnpm install`, `make init confirm=init`, and `make mock confirm=mock` unless `LINAPRO_SKIP_MOCK=1`.
7. Prints the project path, default `admin` / `admin123` credentials, and the `make dev` next step.

## Diagnostics and Retry

- If latest release resolution fails, rerun with `LINAPRO_VERSION=v0.x.y`.
- If clone fails, verify network access and that the selected tag exists in GitHub Releases.
- If prerequisites fail, install the missing tools with the platform-specific hints printed by `checks/prereq.sh`, then rerun the platform script inside the cloned repository.
- If ports `5666` or `8080` are occupied, stop the conflicting process before running `make dev`.
- If database initialization fails, check `apps/lina-core/manifest/config/config.yaml` and MySQL connectivity, then rerun `make init confirm=init`.

## Deployment to linapro.ai

Publishing the remote entry point is an operations task outside this repository change.

1. CI/CD copies `hack/scripts/install/bootstrap.sh` to the `linapro.ai` CDN path `/install.sh`.
2. `/install.ps1` is reserved for a future PowerShell entry point, but no PowerShell installer is published by this flow today.
3. CDN cache must be invalidated whenever a new stable LinaPro tag is released.
4. After publishing, verify from a clean environment:

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```
