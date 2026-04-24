# Installer Scripts

This directory contains the official bootstrap entrypoints for downloading the `LinaPro` source tree from a GitHub archive and unpacking it into a local workspace.

## Directory Layout

```text
hack/scripts/install/
  install.sh       installer for macOS and Linux
  install.ps1      installer for Windows PowerShell
  test_install.py  smoke test for installer behavior and option contracts
  README.md        English documentation
  README.zh_CN.md  Simplified Chinese mirror
```

## Purpose

The installer scripts focus on a narrow bootstrap workflow:

- download a source archive from `GitHub/Codeload`
- extract it in a temporary directory first
- copy the project into the final target directory
- resolve the latest stable tag automatically when the caller does not provide `--ref` or `-Ref`
- protect non-empty directories unless the caller explicitly allows overlay mode
- print an environment check for `Go`, `Node.js`, `pnpm`, `MySQL`, and `make`
- print the next recommended commands such as `make init confirm=init` and `make dev`

The scripts do not install system dependencies and do not execute bootstrap commands automatically.

## Official Entry Points

| Platform | Remote Entrypoint | Repository Script |
| --- | --- | --- |
| `macOS` / `Linux` | `curl -fsSL https://linapro.ai/install.sh \| bash` | `hack/scripts/install/install.sh` |
| `Windows PowerShell` | `irm https://linapro.ai/install.ps1 \| iex` | `hack/scripts/install/install.ps1` |

The hosted URLs should remain thin wrappers over the repository-backed scripts so the installation behavior does not drift.

By default, a no-argument install resolves the latest stable semver tag from the target repository and falls back to `main` when no stable tag is available.

## Usage

### `macOS` and `Linux`

```bash
bash ./hack/scripts/install/install.sh
bash ./hack/scripts/install/install.sh --ref v0.1.0 --dir ~/Workspace/linapro
bash ./hack/scripts/install/install.sh --current-dir --force
```

### `Windows PowerShell`

```powershell
.\hack\scripts\install\install.ps1
.\hack\scripts\install\install.ps1 -Ref v0.1.0 -Dir C:\Workspace\linapro
.\hack\scripts\install\install.ps1 -CurrentDir -Force
```

## Parameters

| Shell | PowerShell | Meaning |
| --- | --- | --- |
| `--repo` | `-Repo` | Override the default GitHub repository, for example `owner/name`. |
| `--ref` | `-Ref` | Select a branch, tag, or commit-like reference to download. |
| `--dir` | `-Dir` | Install into an explicit target directory. |
| `--name` | `-Name` | Create a child directory under the current working path. |
| `--current-dir` | `-CurrentDir` | Unpack directly into the current directory. |
| `--force` | `-Force` | Allow overlay install into a non-empty target directory. |
| `--help` | `-Help` | Print the built-in usage guide. |

## Local Archive Override

Both scripts support the environment variable `LINAPRO_INSTALL_ARCHIVE_PATH`.
It points the installer at a local archive file so you can test the workflow without downloading from the network.
The Shell script expects a local `.tar.gz` archive, and the PowerShell script expects a local `.zip` archive.

Both scripts also support `LINAPRO_INSTALL_STABLE_REF` to override the auto-detected stable tag. This is primarily useful for tests and controlled wrapper scripts.

## Validation

Use the shared smoke test entrypoint when you update installer behavior or wire the script into `CI`:

```bash
make test-install
```

This target runs `python3 hack/scripts/install/test_install.py` and currently validates:

- local archive installation into a named directory
- current-directory installation mode
- rejection of non-empty targets without `--force`
- option contract consistency between `install.sh` and `install.ps1`
