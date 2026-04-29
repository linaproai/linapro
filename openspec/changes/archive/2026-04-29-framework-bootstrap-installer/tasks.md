## 1. Cross-platform Installation Script Entry Points

- [x] 1.1 Add `install.sh` and `install.ps1` under the repository root `hack/scripts/install/`, uniformly defining core parameter semantics and help messages for repository, `ref`, current directory / specified directory, overwrite protection, etc.
- [x] 1.2 Implement source archive download, temporary directory extraction, dynamic top-level directory identification, and target directory deployment logic in both scripts, ensuring the main flow does not depend on `git clone`
- [x] 1.3 Implement safe directory policy: default deployment to a safe directory, explicit support for current directory mode, and refusal to continue execution when the target directory is non-empty and overwrite is not allowed

## 2. Post-installation Guidance and Documentation

- [x] 2.1 Add environment health check output for key dependencies such as `Go`, `Node.js`, `pnpm`, `MySQL`, `make`, as well as project path hints, at the end of the installation scripts
- [x] 2.2 Output unified post-installation next-step guidance, clearly listing recommended operations such as `make init`, `make mock`, `make dev` and related notes
- [x] 2.3 Update the repository root `README.md` and `README.zh_CN.md`, adding quick install examples for `macOS/Linux` and `Windows`, parameter descriptions, and official entry point mapping

## 3. Verification and Regression

- [x] 3.1 Add automated verification or minimal executable smoke tests for the installation scripts, covering core behaviors such as parameter parsing, archive download URL generation, directory protection, and error prompts
- [x] 3.2 Run script-related verification to confirm that the `macOS/Linux` and `Windows` entry points maintain consistency in core parameter contracts

## Feedback

- [x] **FB-1**: Change the installation script directory convention from `hack/scripts/` to `hack/scripts/install/`, and synchronize path references in proposal, design, spec, and tasks
- [x] **FB-2**: Add reusable local/CI execution entry points for installation script smoke tests, avoiding verification methods that only rely on manual commands
- [x] **FB-3**: Add Chinese and English documentation under `hack/scripts/install/`, describing the installation script's purpose, parameters, usage, and verification methods
- [x] **FB-4**: When the user does not explicitly pass `ref`, default to resolving and installing the latest stable tag version of the repository; if no stable tag exists, fall back to the default main branch, and display the final resolved reference value in the output
- [x] **FB-5**: Migrate repository-level standalone Go tools to `hack/tools/`, and synchronize directory references in build, upgrade, test, and specification documents
- [x] **FB-6**: Remove SQL seed data for the built-in cleanup task from `sys_job`, changing it to only be generated through source code registration and startup projection synchronization
- [x] **FB-7**: Split the core HTTP startup function and add key logic comments to reduce the single-function complexity of `cmd_http.go`
- [x] **FB-8**: Unify development build and backend runtime relative paths to the repository root `temp/`, avoiding the generation of duplicate `temp` directories under `apps/lina-core`
- [x] **FB-9**: Fix the issue where the cron expression column in the cron job list has insufficient contrast and is hard to read in dark theme
