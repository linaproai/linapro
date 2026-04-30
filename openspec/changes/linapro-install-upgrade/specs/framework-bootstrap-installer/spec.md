## MODIFIED Requirements

### Requirement: Installation entry point must provide consistent cross-platform capabilities

The system SHALL provide a single quick installation entry point under `hack/scripts/install/bootstrap.sh`, which is also published to `https://linapro.ai/install.sh` so that users can run `curl -fsSL https://linapro.ai/install.sh | bash` to install on any supported platform. The `bootstrap.sh` MUST be a fully self-contained single-file `bash` script that detects the operating system and dispatches to one of three platform-specific bash scripts (`install-macos.sh`, `install-linux.sh`, `install-windows.sh`). The system MUST NOT maintain a separate PowerShell entry point; Windows users MUST execute the same single command in `Git Bash` (bundled with `Git for Windows`) or `WSL`.

The dispatcher MUST share consistent core parameter semantics across all three platforms, covering at least target version, target directory, mock data toggle, and non-interactive mode. Parameters MUST be exposed as environment variables (`LINAPRO_VERSION`, `LINAPRO_DIR`, `LINAPRO_SKIP_MOCK`, `LINAPRO_NON_INTERACTIVE`, `LINAPRO_SHALLOW`) rather than positional arguments, to remain compatible with the `curl | bash` invocation pattern.

#### Scenario: macOS user runs the unified installation command

- **WHEN** a user runs `curl -fsSL https://linapro.ai/install.sh | bash` on `macOS`
- **THEN** `bootstrap.sh` detects `Darwin` via `uname -s`
- **AND** dispatches to `hack/scripts/install/install-macos.sh` after the repository is cloned

#### Scenario: Linux user runs the unified installation command

- **WHEN** a user runs `curl -fsSL https://linapro.ai/install.sh | bash` on `Linux`
- **THEN** `bootstrap.sh` detects `Linux` via `uname -s`
- **AND** dispatches to `hack/scripts/install/install-linux.sh` after the repository is cloned

#### Scenario: Windows user runs the unified installation command in Git Bash

- **WHEN** a user runs `curl -fsSL https://linapro.ai/install.sh | bash` in `Git Bash` on Windows
- **THEN** `bootstrap.sh` detects `MINGW*` / `MSYS*` / `CYGWIN*` via `uname -s`
- **AND** dispatches to `hack/scripts/install/install-windows.sh` after the repository is cloned

#### Scenario: Windows user runs the unified installation command in PowerShell

- **WHEN** a user runs `curl -fsSL https://linapro.ai/install.sh | bash` in native `PowerShell` or `CMD` without `bash` available
- **THEN** the command fails because `bash` is not in `PATH`
- **AND** the system documentation guides the user to switch to `Git Bash` or `WSL` before retrying

### Requirement: Installation script must safely handle target directory selection

The system SHALL clone the `LinaPro` repository into a target directory determined by the `LINAPRO_DIR` environment variable. When `LINAPRO_DIR` is not set, the default target directory MUST be `./linapro` (a new subdirectory in the current working directory). The script MUST refuse to overwrite an existing non-empty directory unless the user explicitly opts in through a documented overwrite flag or environment variable.

#### Scenario: User installs into the default subdirectory

- **WHEN** a user runs the installation command without setting `LINAPRO_DIR`
- **THEN** the script clones the repository into `./linapro`
- **AND** if `./linapro` already exists with content, the script refuses to continue and prints the recovery action

#### Scenario: User installs into an explicit directory via environment variable

- **WHEN** a user runs the installation command with `LINAPRO_DIR=/path/to/my-app`
- **THEN** the script clones the repository into `/path/to/my-app`
- **AND** if `/path/to/my-app` does not exist, the script creates the directory and its parents before cloning

#### Scenario: User attempts to overwrite a non-empty directory without opt-in

- **WHEN** a user runs the installation command targeting a directory that already contains files not created by `LinaPro`
- **THEN** the script refuses to continue execution
- **AND** prints a message describing how to opt in to overwrite if that is intended

### Requirement: Post-installation output must include environment health check and next-step guidance

The system SHALL run `hack/scripts/install/checks/prereq.sh` after source code deployment to verify that key dependencies required for the `LinaPro` development flow are present. The health check results MUST cover at least the presence or version information of `Go` (>= 1.22), `Node.js` (>= 20), `pnpm` (>= 8), `git`, and `MySQL` client connectivity. The script MUST also probe TCP ports `5666` and `8080` for occupancy and warn (but not block) if they are in use.

The installation success output MUST inform the user of the installed project path, the default administrator credentials (`admin` / `admin123`), and the recommended next-step commands (`make dev`).

#### Scenario: Output guidance to proceed when all key dependencies are present

- **WHEN** the installation script completes source code deployment and detects that all key dependencies are satisfied
- **THEN** the script outputs the project directory path
- **AND** prompts the user to continue executing `make dev` to start the development environment

#### Scenario: Output diagnostic results when dependencies are missing

- **WHEN** the installation script completes source code deployment but finds one or more key dependencies missing
- **THEN** the script outputs each missing or unsatisfied dependency item with a platform-specific install hint (`brew` on macOS, `apt-get` / `yum` on Linux, `winget` / `scoop` on Windows)
- **AND** still informs the user that the project has been successfully cloned and provides direction for supplementing the environment

#### Scenario: Output port occupancy warning

- **WHEN** the installation script detects that port `5666` or `8080` is already in use
- **THEN** the script prints a warning identifying the occupied port and the suggested action
- **AND** continues to print success output, since port conflicts are diagnosed at `make dev` time

## REMOVED Requirements

### Requirement: Installation flow must be based on source archive download rather than Git clone

**Reason**: The new installation model uses `git clone` so that the installed project retains full git history, which the `lina-upgrade` skill relies on (for `git diff`, `git merge-base`, and `git merge` operations). Archive-based download cannot support the upgrade workflow without re-bootstrapping git state.

**Migration**: Installation flows previously based on archive download MUST now use `git clone --branch <tag> https://github.com/linaproai/linapro.git $LINAPRO_DIR` (or the equivalent for a custom repository specified by the user). Default behavior is full clone; users on bandwidth-constrained environments can opt in to `LINAPRO_SHALLOW=1` for shallow clone, but the upgrade workflow will then require running `git fetch --unshallow` before the first upgrade.

## ADDED Requirements

### Requirement: Hosted bootstrap script must be a self-contained single file

The system SHALL ensure that the published `https://linapro.ai/install.sh` content is byte-identical to the repository's `hack/scripts/install/bootstrap.sh`. The `bootstrap.sh` MUST NOT make additional network calls to fetch secondary scripts (no chained `curl` calls inside the bootstrap script). All platform-specific logic, helper functions, and dependency checks MUST be delivered as part of the cloned repository, not separately downloaded.

#### Scenario: Bootstrap script does not chain additional downloads

- **WHEN** a user runs `curl -fsSL https://linapro.ai/install.sh | bash`
- **THEN** `bootstrap.sh` only makes the network calls necessary to (1) probe the GitHub `releases/latest` redirect and (2) execute `git clone`
- **AND** does not invoke any additional `curl` or `wget` to fetch helper scripts

#### Scenario: Local bootstrap execution is equivalent to hosted execution

- **WHEN** a user has already cloned the repository and runs `bash hack/scripts/install/bootstrap.sh`
- **THEN** the script behaves identically to the hosted `curl | bash` invocation
- **AND** does not require network access if `LINAPRO_VERSION` is explicitly provided and the repository is already at the right ref

### Requirement: Default target version must resolve to the latest stable GitHub release

The system SHALL resolve the default installation target version by following the redirect of `https://github.com/linaproai/linapro/releases/latest` and extracting the tag name from the redirected URL. This mechanism uses GitHub's built-in semantics for "latest stable" (excluding pre-release tags) without consuming the GitHub API rate limit. The system MUST NOT silently fall back to the `main` branch when default resolution fails; the script MUST fail with a clear error message instructing the user to set `LINAPRO_VERSION` explicitly.

#### Scenario: Default version resolves to the latest stable tag

- **WHEN** a user runs the installation command without setting `LINAPRO_VERSION`
- **AND** the GitHub repository has at least one stable release
- **THEN** `bootstrap.sh` issues `curl -sIL https://github.com/linaproai/linapro/releases/latest` and parses the redirect
- **AND** extracts the tag (e.g., `v0.5.0`) as the target version
- **AND** clones that tag

#### Scenario: User overrides version via environment variable

- **WHEN** a user runs the installation command with `LINAPRO_VERSION=v0.4.0`
- **THEN** `bootstrap.sh` skips the GitHub redirect resolution
- **AND** clones the `v0.4.0` tag directly

#### Scenario: Default version resolution fails

- **WHEN** a user runs the installation command without setting `LINAPRO_VERSION`
- **AND** the GitHub redirect cannot be reached or returns no parseable tag
- **THEN** `bootstrap.sh` exits with a clear error message
- **AND** the error message includes the suggested workaround `LINAPRO_VERSION=<tag> curl -fsSL ... | bash`

### Requirement: Installation must support common environment variable overrides

The system SHALL support the following environment variables to control installation behavior, each with a documented default:

| Variable | Default | Effect |
| --- | --- | --- |
| `LINAPRO_VERSION` | (resolved from GitHub) | Target version tag to install |
| `LINAPRO_DIR` | `./linapro` | Target directory to clone into |
| `LINAPRO_NON_INTERACTIVE` | unset | When set, skips interactive prompts and uses defaults |
| `LINAPRO_SKIP_MOCK` | unset | When set, skips loading mock demo data after `make init` |
| `LINAPRO_SHALLOW` | unset | When set, performs shallow clone (`--depth 1`) instead of full clone |

The platform-specific `install-*.sh` scripts MUST read the same variables when invoked outside the bootstrap flow, to preserve behavioral consistency between fresh installation and re-runs.

#### Scenario: Non-interactive mode skips prompts

- **WHEN** a user runs the installation command with `LINAPRO_NON_INTERACTIVE=1`
- **THEN** the script does not prompt for confirmation at any step
- **AND** uses default behavior for all decision points

#### Scenario: Skip mock data flag is honored

- **WHEN** a user runs the installation command with `LINAPRO_SKIP_MOCK=1`
- **THEN** the script runs `make init` to apply DDL and seed data
- **AND** does not run `make mock`

#### Scenario: Shallow clone flag is honored

- **WHEN** a user runs the installation command with `LINAPRO_SHALLOW=1`
- **THEN** `bootstrap.sh` invokes `git clone --branch <tag> --depth 1`
- **AND** prints a notice that the `lina-upgrade` skill will require `git fetch --unshallow` on first upgrade

### Requirement: Installation scripts must enforce LF line endings via .gitattributes

The system SHALL include `.gitattributes` rules that force all `*.sh` files (and `bootstrap.sh` specifically) to use LF line endings, preventing Windows automatic CRLF conversion from corrupting bash scripts when the repository is cloned on Windows. The repository MUST contain a `.gitattributes` file at the root with at least the following entries:

```
*.sh text eol=lf
hack/scripts/install/bootstrap.sh text eol=lf
```

#### Scenario: Windows user clones repository and bash scripts have LF endings

- **WHEN** a Windows user clones the `LinaPro` repository
- **AND** their git config has `core.autocrlf=true`
- **THEN** `*.sh` files in the working tree retain LF line endings
- **AND** `bash bootstrap.sh` executes without `\r` parsing errors

### Requirement: Platform-specific install scripts must share common helper library

The system SHALL provide a shared bash helper library at `hack/scripts/install/lib/_common.sh` containing functions for logging, version comparison, retry logic, and structured error reporting. Each platform-specific install script (`install-macos.sh`, `install-linux.sh`, `install-windows.sh`) MUST source this library at the top and use its functions instead of duplicating logic.

#### Scenario: All platform scripts use shared logging functions

- **WHEN** any platform-specific install script needs to print an info / warning / error message
- **THEN** it calls `log_info`, `log_warn`, or `log_error` from `lib/_common.sh`
- **AND** does not implement its own coloring or prefix logic

#### Scenario: All platform scripts use shared version-comparison helpers

- **WHEN** any platform-specific install script needs to compare semantic versions (e.g., to validate `go` minimum version)
- **THEN** it calls a comparison function from `lib/_common.sh`
- **AND** does not duplicate version-parsing logic
