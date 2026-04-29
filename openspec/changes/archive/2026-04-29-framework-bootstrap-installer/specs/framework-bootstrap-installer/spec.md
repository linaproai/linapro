## ADDED Requirements

### Requirement: Installation entry point must provide consistent cross-platform capabilities
The system SHALL provide a quick installation script entry point for the entire `LinaPro` repository under the repository root `hack/scripts/install/`, where `install.sh` is for `macOS/Linux` and `install.ps1` is for `Windows PowerShell`. The two entry points MUST share consistent core parameter semantics, covering at least repository source, source reference, target directory mode, and overwrite protection behaviors, rather than implementing incompatible interaction models separately.

#### Scenario: Unix-like user executes installation via Shell entry point
- **WHEN** a user executes `hack/scripts/install/install.sh` on `macOS` or `Linux`
- **THEN** the script can parse core parameters related to quick installation
- **AND** the script begins executing the source code download and deployment flow for the entire `LinaPro` repository

#### Scenario: Windows user executes installation via PowerShell entry point
- **WHEN** a user executes `hack/scripts/install/install.ps1` on `Windows PowerShell`
- **THEN** the script can parse core installation parameters equivalent to the Shell entry point
- **AND** the script begins executing the source code download and deployment flow for the entire `LinaPro` repository

### Requirement: Installation flow must be based on source archive download rather than Git clone
The system SHALL support downloading `LinaPro` source archives based on user-specified repository and `ref`, and use the archive contents to complete local project deployment. The main installation flow MUST NOT depend on `git clone` or Git being pre-installed locally; on `macOS/Linux` it should prefer archive formats suitable for Shell environments, and on `Windows` it should prefer archive formats suitable for native PowerShell extraction.

#### Scenario: Download source archive using default repository and default reference
- **WHEN** a user executes the installation script without explicitly specifying the repository or `ref` parameters
- **THEN** the script downloads the source archive from the official default repository with the default reference
- **AND** the default reference preferably resolves to the latest stable semantic tag version of the repository
- **AND** the local installation flow does not require the user to have Git pre-installed

#### Scenario: Download source archive using specified repository and specified reference
- **WHEN** a user explicitly passes a repository address and `ref` to the installation script
- **THEN** the script downloads the source archive from that repository at the corresponding reference
- **AND** then uses the archive contents to perform extraction and local directory deployment

#### Scenario: Fall back to main branch when default stable tag is missing
- **WHEN** the user does not explicitly specify `ref`, and the target repository has no identifiable stable semantic tag version
- **THEN** the script falls back to the agreed default main branch reference to continue source archive download
- **AND** the output clearly displays the final resolved reference value

### Requirement: Installation script must safely handle current directory and specified directory modes
The system SHALL support extracting source code to the current directory or an explicitly specified directory, and by default avoid directly overwriting into non-empty directories. The script MUST first complete archive extraction and top-level directory identification in a temporary directory, then move the actual project files to the final target location; when the target directory is non-empty and the user has not explicitly allowed overwriting, the script MUST refuse to continue execution.

#### Scenario: User explicitly requests deployment to current directory
- **WHEN** a user executes the installation script with the current directory mode parameter
- **THEN** the script deploys the extracted project contents to the current working directory
- **AND** if the current directory already has content not created by the installation script and overwriting is not explicitly allowed, the script refuses to continue execution

#### Scenario: User explicitly requests deployment to specified directory
- **WHEN** a user executes the installation script with a target directory parameter
- **THEN** the script deploys the extracted project contents to the specified directory
- **AND** if the target directory does not exist, the script creates the directory or its parent directories before completing the installation

### Requirement: Post-installation output must include environment health check and next-step guidance
The system SHALL check key dependencies required for the `LinaPro` development flow after source code deployment is complete, and output clear post-installation guidance. The health check results MUST cover at least the presence or version information of `Go`, `Node.js`, `pnpm`, `MySQL`, and `make`; the installation success output MUST inform the user of the project path and the recommended key commands to execute next.

#### Scenario: Output guidance to proceed when all key dependencies are present
- **WHEN** the installation script completes source code deployment and detects that all key dependencies are satisfied
- **THEN** the script outputs the project directory
- **AND** the script prompts the user to continue executing subsequent commands such as `make init`, `make mock`, `make dev`

#### Scenario: Output diagnostic results when dependencies are missing
- **WHEN** the installation script completes source code deployment but finds one or more key dependencies missing
- **THEN** the script outputs each missing or unsatisfied dependency item
- **AND** the script still clearly informs the user that the project has been successfully deployed and provides direction for supplementing the environment
