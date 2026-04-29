## Why

The current local onboarding flow for `LinaPro` still leans toward "developers who are already familiar with the repository structure": users need to locate the repository, manually download or clone the source code, decide on an extraction directory themselves, and then separately verify whether their machine has the required dependencies such as `Go`, `Node.js`, `pnpm`, `MySQL`, and `make`. For first-time users of the framework, the environment preparation path is too long and varies significantly across platforms, making it easy to get stuck before actually starting to use the framework.

As an "AI-native full-stack framework for sustainable delivery", `LinaPro` should provide a unified, low-barrier, cross-platform entry point for quickly getting the source code up and running, allowing users to download a specified repository version to a local target directory with a single installation command and immediately receive follow-up initialization guidance and environment check results.

## What Changes

- Add a repository-level "quick install / source code bootstrap" capability, providing an official installation script entry point under `hack/scripts/install/` for quickly downloading and deploying the entire `LinaPro` repository source code, rather than covering only a single sub-project.
- Provide `install.sh` for `macOS` / `Linux` and `install.ps1` for `Windows`, uniformly supporting downloading compressed archives from a specified code repository at a specified branch, tag, or commit reference, rather than relying on `git clone`.
- The installation script defaults to distributing source code via GitHub/Codeload archives, preferring the lighter archive download method, and based on user parameters decides whether to extract the project into a new subdirectory under the current directory or into an explicitly specified directory.
- The installation script needs a built-in safe directory policy: detecting whether the target directory already exists, whether it is empty, and whether overwriting is allowed, and by default avoids directly scattering source code into a non-empty directory.
- After the source code is deployed, the installation script outputs a unified environment health check and next-step guidance, including but not limited to dependency check results for `Go`, `Node.js`, `pnpm`, `MySQL`, `make`, as well as subsequent command hints like `make init` and `make dev`.
- Update the repository root `README.md` and `README.zh_CN.md` with quick install instructions, providing recommended usage for `macOS/Linux` and `Windows` respectively, and clarifying the mapping between the official hosting entry point and in-repository scripts.

## Capabilities

### New Capabilities
- `framework-bootstrap-installer`: Provides cross-platform source code quick download, target directory deployment, safe extraction, environment health check, and post-installation guidance capabilities.

### Modified Capabilities
- None.

## Impact

- **Repository Scripts**: Affects the repository root directory `hack/scripts/install/`, requiring new cross-platform installation scripts and a unified parameter model.
- **Project Documentation**: Affects the repository root `README.md` and `README.zh_CN.md` quick start and installation instructions.
- **Developer Experience**: Affects the environment preparation path, directory initialization method, and cross-platform onboarding experience for first-time `LinaPro` users.
- **Distribution Conventions**: Requires clarifying GitHub/Codeload archive download rules, target directory policies, and consistency constraints between the official installation entry point and repository script versions.
