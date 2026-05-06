# LinaPro Installer

This directory contains the repository-backed implementation for the single LinaPro source download entry point:

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```

The hosted `/install.sh` content must match `hack/scripts/install/install.sh`. The script is self-contained and only downloads the requested LinaPro repository source. Runtime dependency checks and tool installation belong to the `lina-doctor` skill.

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
  install.sh          hosted curl|bash entrypoint
  README.md             English documentation
  README.zh_CN.md       Simplified Chinese mirror
```

## Environment Variables

| Variable | Default | Meaning | Example |
| --- | --- | --- | --- |
| `LINAPRO_VERSION` | Highest stable Git tag from `origin` | Target version tag to clone. The installer fails if it cannot resolve a tag automatically. | `LINAPRO_VERSION=v0.5.0 curl -fsSL https://linapro.ai/install.sh \| bash` |
| `LINAPRO_DIR` | `./linapro` | Target directory for the cloned project. | `LINAPRO_DIR=~/Workspace/my-linapro curl -fsSL https://linapro.ai/install.sh \| bash` |
| `LINAPRO_SHALLOW` | unset | Uses a shallow clone. Default full clones are recommended because they can fetch later release tags directly for upgrades. | `LINAPRO_SHALLOW=1 ...` |
| `LINAPRO_FORCE` | unset | Allows replacing a non-empty target directory after built-in safety checks. | `LINAPRO_FORCE=1 ...` |

`LINAPRO_NON_INTERACTIVE` and `LINAPRO_SKIP_MOCK` are no longer used by the installer. The script does not prompt for environment setup and does not load mock data.

## Local Equivalent

From an existing repository checkout, run the same installer source locally:

```bash
bash hack/scripts/install/install.sh
```

The command still clones the requested version into `LINAPRO_DIR` or `./linapro`; it does not install over the current checkout unless you explicitly set `LINAPRO_DIR`.

## What The Installer Does

1. Detects whether the command is running on a supported `bash` platform.
2. Resolves `LINAPRO_VERSION` or selects the highest stable `vX.Y.Z` tag from the remote Git repository.
3. Refuses to overwrite a non-empty target directory unless `LINAPRO_FORCE=1` passes safety checks.
4. Runs a Git clone that keeps the `origin` remote and fetches release tags for later tag-based upgrades.
5. Checks out the selected tag and prints the project path, default `admin` / `admin123` credentials, and next steps.

## Next Steps After Clone

```bash
cd <project-dir>
# Ask your AI tool to run lina-doctor and fix missing development tools.
make init && make dev
```

Use `lina-doctor` before project initialization when Go, Node, pnpm, OpenSpec, GoFrame CLI, Playwright browsers, or the `goframe-v2` skill may be missing.

## Tag-Based Upgrades

The default install keeps a normal Git repository with the `origin` remote. To move an installed checkout to a newer release tag later:

```bash
git fetch --tags --force origin
git checkout --detach <new-version-tag>
```

Avoid `LINAPRO_SHALLOW=1` unless clone size is more important than upgrade ergonomics. If a shallow checkout cannot move to a newer tag, run `git fetch --unshallow --tags --force origin` once before checking out the new tag.

## Diagnostics And Retry

- If latest tag resolution fails, rerun with `LINAPRO_VERSION=v0.x.y`.
- If clone fails, verify network access and confirm that the selected tag exists in GitHub Releases.
- If the target directory is not empty, choose another `LINAPRO_DIR` or rerun with `LINAPRO_FORCE=1` after checking the target path.
- If development tools are missing after clone, invoke the `lina-doctor` skill through your AI tool.

## Deployment To linapro.ai

Publishing the remote entry point is an operations task outside this repository change.

1. CI/CD copies `hack/scripts/install/install.sh` to the `linapro.ai` CDN path `/install.sh`.
2. CDN cache must be invalidated whenever `install.sh` changes.
3. After publishing, verify from a clean environment:

```bash
curl -fsSL https://linapro.ai/install.sh | bash
```
