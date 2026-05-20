## ADDED Requirements

### Requirement: Release tag must be consistent with framework metadata version

The system SHALL use `framework.version` in `apps/lina-core/manifest/config/metadata.yaml` as the sole version baseline for release tags. Any release tag publishing pipeline MUST validate that the Git tag name exactly matches `framework.version` before executing tests, building images, pushing to GHCR, or updating the floating tag.

#### Scenario: Tag name matches framework version and publishing continues
- **WHEN** GitHub Actions receives a release tag push event
- **AND** the tag name equals `framework.version` in `apps/lina-core/manifest/config/metadata.yaml`
- **THEN** the release workflow continues executing subsequent test and image publishing gates

#### Scenario: Tag name does not match framework version and publishing is blocked
- **WHEN** GitHub Actions receives a release tag push event
- **AND** the tag name does not equal `framework.version` in `apps/lina-core/manifest/config/metadata.yaml`
- **THEN** the release workflow fails before tests and image publishing
- **AND** the workflow MUST NOT execute GHCR login, `make image push=1`, `latest` tag publishing, or remote manifest inspect

#### Scenario: Invalid framework version format blocks publishing
- **WHEN** the release workflow validates `framework.version`
- **AND** `framework.version` is not a Docker tag-compatible release version format
- **THEN** the release workflow fails
- **AND** the error message states `framework.version`, the Git tag name, and the allowed version format

### Requirement: Release tag validation must be reusable through cross-platform tool

The system SHALL execute release tag and framework metadata version consistency validation through the repository's cross-platform tool entry point. GitHub Actions, local release checks, and subsequent release automation SHALL reuse the same validation command, avoiding duplicate YAML parsing and version format rules in different workflows or scripts.

#### Scenario: Local validation command reads framework version
- **WHEN** a maintainer runs the release tag validation command with a tag name
- **THEN** the command reads `apps/lina-core/manifest/config/metadata.yaml`
- **AND** the command compares the passed tag with `framework.version`
- **AND** the command outputs readable success or failure information

#### Scenario: Workflow reuses the same validation command
- **WHEN** the release workflow needs to validate a tag name
- **THEN** the workflow calls the repository's cross-platform validation command
- **AND** the workflow MUST NOT use independent ad hoc YAML parsing logic to replace the command

### Requirement: Controlled release entry must validate framework version before creating a tag

The system SHALL provide a controlled GitHub Actions manual release entry for reading `framework.version` and creating a matching release tag. The entry MUST run release tag validation before creating the tag; validation failure MUST NOT create or push a tag.

#### Scenario: Controlled entry successfully creates a matching tag
- **WHEN** a maintainer manually triggers the controlled release tag workflow
- **AND** `framework.version` is valid and the corresponding remote tag does not yet exist
- **THEN** the workflow creates a Git tag named equal to `framework.version`
- **AND** the workflow pushes the tag to the GitHub repository

#### Scenario: Controlled entry discovers the tag already exists
- **WHEN** a maintainer manually triggers the controlled release tag workflow
- **AND** a remote Git tag named equal to `framework.version` already exists
- **THEN** the workflow fails
- **AND** the workflow MUST NOT move, overwrite, or delete the existing tag

#### Scenario: Repository rules prevent manual release tag creation
- **WHEN** the repository has a GitHub tag ruleset configured for release tags
- **THEN** ordinary users MUST NOT directly create, update, or delete protected release tags
- **AND** only the controlled release actor can create release tags after validation passes
- **AND** the controlled release workflow MUST use repository variable `RELEASE_APP_CLIENT_ID` and repository secret `RELEASE_APP_PRIVATE_KEY` to generate a GitHub App installation token
- **AND** the ruleset bypass MUST be configured to the GitHub App actor, not to the token string itself
