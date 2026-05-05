# LinaPro Installer

This directory contains the repository-backed implementation for the single LinaPro source download entry point:

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```

The hosted `/install.sh` content must match `hack/scripts/install/bootstrap.sh`. The script is self-contained and only downloads the requested LinaPro repository source. Runtime dependency checks and tool installation belong to the `lina-doctor` skill.

## Supported Platforms

| Platform | Runtime |
| --- | --- |
| `macOS` | `bash` on Darwin |
| `Linux` | `bash` on Linux distributions and WSL |
| `Windows` | Git Bash or WSL |

Windows users must run the command from Git Bash or WSL. Native PowerShell and `cmd.exe` are not supported entry points.

## Directory Layout

```text
hack/scripts/install/
  bootstrap.sh          hosted curl|bash entrypoint
  README.md             English documentation
  README.zh_CN.md       Simplified Chinese mirror
```

## Environment Variables

| Variable | Default | Meaning | Example |
| --- | --- | --- | --- |
| `LINAPRO_VERSION` | Latest stable GitHub release | Target version tag to clone. The installer fails if it cannot resolve a tag automatically. | `LINAPRO_VERSION=v0.5.0 curl -fsSL https://linapro.ai/install.sh \| bash` |
| `LINAPRO_DIR` | `./linapro` | Target directory for the cloned project. | `LINAPRO_DIR=~/Workspace/my-linapro curl -fsSL https://linapro.ai/install.sh \| bash` |
| `LINAPRO_SHALLOW` | unset | Uses `git clone --depth 1`. The first upgrade later requires `git fetch --unshallow`. | `LINAPRO_SHALLOW=1 ...` |
| `LINAPRO_FORCE` | unset | Allows replacing a non-empty target directory after built-in safety checks. | `LINAPRO_FORCE=1 ...` |

`LINAPRO_NON_INTERACTIVE` and `LINAPRO_SKIP_MOCK` are no longer used by the installer. The script does not prompt for environment setup and does not load mock data.

## Local Equivalent

From an existing repository checkout, run the same bootstrap source locally:

```bash
bash hack/scripts/install/bootstrap.sh
```

The command still clones the requested version into `LINAPRO_DIR` or `./linapro`; it does not install over the current checkout unless you explicitly set `LINAPRO_DIR`.

## What The Installer Does

1. Detects whether the command is running on a supported `bash` platform.
2. Resolves `LINAPRO_VERSION` or follows the GitHub `releases/latest` redirect.
3. Refuses to overwrite a non-empty target directory unless `LINAPRO_FORCE=1` passes safety checks.
4. Runs `git clone --branch <tag> https://github.com/linaproai/linapro.git "$LINAPRO_DIR"`.
5. Prints the project path, default `admin` / `admin123` credentials, and next steps.

## Next Steps After Clone

```bash
cd <project-dir>
# Ask your AI tool to run lina-doctor and fix missing development tools.
make init && make dev
```

Use `lina-doctor` before project initialization when Go, Node, pnpm, OpenSpec, GoFrame CLI, Playwright browsers, or the `goframe-v2` skill may be missing.

## Diagnostics And Retry

- If latest release resolution fails, rerun with `LINAPRO_VERSION=v0.x.y`.
- If clone fails, verify network access and confirm that the selected tag exists in GitHub Releases.
- If the target directory is not empty, choose another `LINAPRO_DIR` or rerun with `LINAPRO_FORCE=1` after checking the target path.
- If development tools are missing after clone, invoke the `lina-doctor` skill through your AI tool.

## Deployment To linapro.ai

Publishing the remote entry point is an operations task outside this repository change.

1. CI/CD copies `hack/scripts/install/bootstrap.sh` to the `linapro.ai` CDN path `/install.sh`.
2. CDN cache must be invalidated whenever `bootstrap.sh` changes.
3. After publishing, verify from a clean environment:

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```
