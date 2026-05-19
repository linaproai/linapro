## MODIFIED Requirements

### Requirement: Archive residual governance

The system SHALL identify and govern half-finished main spec updates after archive failures or interruptions, preventing subsequent archiving from being blocked by duplicate additions or residual files. The system SHALL allow completed change archive governance to be executed through a controlled monthly automation process, but that process MUST execute OpenSpec validation and protect the change scope before creating or updating an archive PR.

#### Scenario: Half-finished main spec file exists
- **WHEN** the archive process is interrupted abnormally and leaves half-finished main spec files
- **THEN** the system or maintenance process can identify the residual
- **AND** complete cleanup or alignment before the next archive to avoid duplicate capability writes

#### Scenario: Monthly automated archive governance
- **WHEN** the monthly OpenSpec archive process automatically archives completed changes
- **THEN** the process executes OpenSpec validation before creating or updating an archive PR
- **AND** the process refuses to write file changes outside the archive governance allowed scope into the PR
