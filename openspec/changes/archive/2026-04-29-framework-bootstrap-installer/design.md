## Context

This change is not simply about adding a download script, but about establishing a "first landing entry point" for the entire `LinaPro` repository. The requirements span repository distribution methods, cross-platform script entry points, target directory safety policies, environment dependency health checks, and root repository documentation. Therefore, it is appropriate to clarify the boundaries at the design stage to avoid mixing "source code downloader", "environment installer", and "project initializer" into a single concern during implementation.

The repository already has a root `hack/scripts/` directory for repository-level scripts, but there is no standard entry point for "quickly obtaining source code and getting started." First-time `LinaPro` users need to decide on their own: should they `git clone`, download a ZIP, or copy an existing directory; at the same time, they need to determine whether dependencies like `Go`, `Node.js`, `pnpm`, `MySQL`, and `make` are available. This path is neither unified nor capable of providing a consistent experience across `macOS`, `Linux`, and `Windows`. Therefore, this installation entry point is further scoped to the `hack/scripts/install/` subdirectory to avoid being mixed with other repository-level scripts at the same level.

Additionally, the user's example entry point is `curl -fsSL https://linapro.ai/install.sh | bash`. This indicates an external desire for a "single-command start" official distribution method; but within the scope of the current repository implementation, what really needs to be established first is the in-repository scripts and parameter contracts. The official hosting address should reuse the repository script content, not maintain a separate set of logic outside the repository, to avoid version drift.

## Goals / Non-Goals

**Goals:**
- Establish a quick installation script entry point under the repository root `hack/scripts/install/` for the entire `LinaPro`, rather than just for the `apps/lina-core` sub-project.
- Provide directly executable script entry points for `macOS/Linux` and `Windows` respectively, maintaining consistent core parameter semantics.
- Uniformly adopt the GitHub/Codeload source archive download pattern, supporting repository address and `ref` parameters, avoiding reliance on `git clone`.
- Support "extract to current directory" and "extract to specified directory" modes, with a safe non-overwrite strategy as the default.
- Output local environment health check results and next-step suggestions after installation completes, shortening the first-time onboarding path.
- Establish formal quick install instructions in the root `README.md` and `README.zh_CN.md`.

**Non-Goals:**
- This change is not responsible for automatically installing system dependencies such as `Go`, `Node.js`, `pnpm`, `MySQL`, and `make`; it only checks and prompts.
- This change does not automatically execute commands that modify the local environment or database, such as `make init`, `make mock`, or `make dev`.
- This change does not implement `linapro.ai` site publishing infrastructure within the repository; it only establishes repository scripts as the content source for the official hosting entry point.
- This change does not introduce a `git clone`-based primary installation path, nor does it handle credential-based private repository login flows.

## Decisions

### Decision 1: Use dual entry point scripts instead of trying to support all three platforms with a single Shell file

**Choice**: Add two entry points under the repository root `hack/scripts/install/`: `install.sh` for `macOS/Linux`, and `install.ps1` for `Windows PowerShell`. Both share consistent core parameter semantics, such as repository, reference, target directory, overwrite protection, and current directory mode.

**Reason**:
- `curl | bash` works well for Unix-like environments but not for native Windows users; PowerShell requires native download and extraction capabilities, and forcing everything into a single script would hide platform differences behind extensive compatibility branches, resulting in higher maintenance costs.
- `Windows` natively supports `Invoke-WebRequest` / `Expand-Archive`, while `macOS/Linux` is better suited to `curl` / `wget` + `tar`. Splitting entry points makes script logic clearer and makes it easier for users to understand the recommended commands.

**Alternatives**:
- A single `bash` script supporting all three platforms. Rejected, because it would implicitly require `Windows` users to install `Git Bash` or `WSL`, which conflicts with the goal of "out-of-the-box support for mainstream operating systems."
- A single `Python` script. Not adopted, because we cannot assume users have a suitable `Python` environment before the first installation.

### Decision 2: Source code acquisition uniformly uses archive download; `git clone` is not the primary installation path

**Choice**: The installation script defaults to downloading archive files from GitHub/Codeload based on the platform, with Unix preferring `tar.gz` and Windows preferring `zip`, and supports overriding the repository and `ref` through parameters.

**Reason**:
- Archive download is lighter than `git clone`, does not require repository history, and does not depend on Git being pre-installed locally.
- `tar.gz` is more universal on `macOS/Linux`, while `zip` is better supported by native system commands on `Windows`.
- Through repository and `ref` parameters, it can simultaneously support the official default repository, specified branches, specified tags, and development version distribution, without needing to maintain multiple scripts for different branches.

**Alternatives**:
- Directly using `https://github.com/<repo>/archive/refs/heads/<branch>.zip`. Viable, but prioritizing `codeload.github.com` is more suitable as a pure download entry point, with fewer redirects and clearer distribution semantics.
- Preferring `git clone --depth 1`. Rejected, because it still requires Git to be available, and both the size and error surface are larger.

### Decision 3: Target directory uses "explicit mode selection + safe defaults"

**Choice**: By default, source code is deployed to a new subdirectory under the current working directory; when the user explicitly passes a "current directory mode" parameter, the extraction results are allowed to be placed directly in the current directory; when the user explicitly passes a target path, the script deploys the source code to the specified directory. In all modes, if the target directory is non-empty, the script defaults to refusing to continue unless the user explicitly passes an overwrite parameter.

**Reason**:
- Users have explicitly requested support for both the current directory and specified directory modes, but defaulting to directly extracting the archive contents into the current directory carries high risk and can easily pollute existing files.
- Using "new subdirectory under current directory" as the default mode better matches the safety expectations of a first installation and is compatible with non-interactive commands like `curl ... | bash`.
- The top-level directory name of GitHub archives changes with branches or tags; the script needs to dynamically identify the unique root directory after extraction to a temporary directory, then move it to the final position, rather than hardcoding names like `linapro-main` into the logic.

**Alternatives**:
- Default to directly extracting into the current directory. Not adopted, because the risk of accidental overwrite is too high.
- Only support `--dir`, not support "current directory mode". Not adopted, because it does not meet the user's explicit requirement for "current directory / specified directory" dual mode.

### Decision 4: The first version of the installation script only performs "environment health check + next-step guidance", not automatic system dependency installation

**Choice**: After source code deployment is complete, the script uniformly checks for the presence of key dependencies such as `Go`, `Node.js`, `pnpm`, `MySQL`, and `make`, and outputs version information or missing dependency prompts; but it does not automatically call package managers like `brew`, `apt`, `yum`, `choco`, or `winget` to install dependencies.

**Reason**:
- The common problem that actually hinders first-time onboarding is "not knowing what you're missing", not "having the installation script install everything for you."
- Automatically installing system dependencies would introduce administrator privileges, network mirrors, package manager differences, and enterprise environment restrictions, with complexity far exceeding the problem this change aims to solve.
- Doing a good job on dependency checking and next-step command guidance can already significantly reduce onboarding costs while keeping the script predictable and auditable.

**Alternatives**:
- Automatically install all dependencies by platform. Rejected, because cross-platform differences and risks are too high.
- Do not check dependencies at all, only prompt "please see README". Not adopted, because it would break the quick install experience midway.

### Decision 5: Repository scripts as the single source of truth; official distribution URL is only a thin wrapper

**Choice**: `hack/scripts/install/install.sh` and `hack/scripts/install/install.ps1` serve as the sole source of installation logic within the repository; the publicly available `https://linapro.ai/install.sh` / `install.ps1` should reuse the same content, at most doing a thin wrapper or redirect mapping, and the site-hosted script must not diverge from the repository script long-term.

**Reason**:
- The installation entry point is one of the first capabilities users encounter; if the site copy and repository script logic are inconsistent, it will amplify troubleshooting costs.
- Making the repository script the single source of truth is also more consistent with the governance approach of publishing stable versions through release/tag.

**Alternatives**:
- The website maintains a shorter bootstrap script separately, which then assembles the logic. Not adopted, because it easily forms a second implementation.

## Risks / Trade-offs

- [Risk] When dependencies are not automatically installed, some users still need to manually supplement their environment. → Mitigation: Output clear missing items, recommended versions, and next-step commands, prioritizing the resolution of "not understanding the blocker."
- [Risk] Differences in download and extraction tool availability across platforms. → Mitigation: Unix side prefers `curl` with `wget` fallback; Windows side uses native PowerShell commands; clear failure prompts are given when necessary capabilities are missing.
- [Risk] Current directory mode can easily damage existing files. → Mitigation: Design it as an explicitly triggered parameter, and default to refusing execution when the directory is non-empty.
- [Risk] If the official hosting URL is published with a delay, it may briefly be inconsistent with the repository script version. → Mitigation: Establish repository scripts as the source of truth, and constrain synchronization methods in documentation and release processes.

## Migration Plan

1. Add `install.sh` and `install.ps1` under the repository root `hack/scripts/install/`, establishing a unified parameter model and output format.
2. Implement source archive download, temporary directory extraction, dynamic top-level directory identification, and target directory deployment logic.
3. Add environment dependency check and next-step guidance output at the end of the installation flow.
4. Update the repository root `README.md` and `README.zh_CN.md`, adding quick install examples and platform instructions.
5. Add automated verification or minimal executable regression tests for the scripts, ensuring stable parameter parsing, directory policy, and error prompt behavior.
6. In the future, if an external `linapro.ai` hosting entry point is needed, the repository script content can be directly reused without defining a second implementation.

## Open Questions

- Does the first version need to support enterprise mirror source parameters beyond `repo` (such as GitHub Enterprise or self-hosted mirrors)? The current design focuses on public GitHub repositories first.
- Should environment health check results output a machine-parseable format (such as `json`) to support future installer UI or diagnostic tool reuse? The first version can remain as human-readable text.
