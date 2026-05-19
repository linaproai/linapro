## MODIFIED Requirements

### Requirement: Image build command must support multi-architecture Docker image building and pushing

The system SHALL allow operators to build and push multi-architecture Docker images via `make image platforms=linux/amd64,linux/arm64 registry=<registry> push=1`. Multi-architecture image building SHALL use the host binary corresponding to each target platform rather than reusing a single platform binary for all image platforms. Multi-platform image pushing SHALL use `Docker buildx` to generate and push remote multi-architecture manifests. When push is not enabled, multi-platform image building SHALL fail fast and prompt that `push=1` is required. Image building implementation SHALL be executed by the `linactl` internal image building component rather than depending on the independent `hack/tools/image-builder` tool module.

#### Scenario: Build and push amd64 and arm64 multi-architecture images

- **WHEN** an operator runs `make image platforms=linux/amd64,linux/arm64 registry=ghcr.io/linaproai tag=v1.0.0 push=1`
- **THEN** the build flow first prepares `linux/amd64` and `linux/arm64` host binaries separately
- **AND** the `Dockerfile` selects the corresponding platform host binary in each platform build context
- **AND** the image build flow pushes the multi-architecture manifest for `ghcr.io/linaproai/linapro:v1.0.0` via `Docker buildx`

#### Scenario: Multi-platform image build without push fails fast

- **WHEN** an operator runs `make image platforms=linux/amd64,linux/arm64 push=0`
- **THEN** the image build flow fails before calling `Docker buildx`
- **AND** the error message states that multi-platform image building requires `push=1`

#### Scenario: Image command no longer calls independent image-builder tool

- **WHEN** an operator runs `linactl image` or `linactl image.build`
- **THEN** the command completes image building or staging through the `linactl` internal image building component
- **AND** the command MUST NOT execute `go run ./hack/tools/image-builder`

### ADDED Requirements

### Requirement: Release workflow must reuse shared test verification suite and run brief test gates

The system SHALL provide a `Release Test and Build` GitHub Actions workflow to replace the image-only release workflow. After tag push, the workflow SHALL reuse the shared test verification suite like `Nightly Test and Build` and adopt Main CI's brief test scope: host-only and plugin-full Windows command smoke, Go unit tests, frontend unit tests, plugin command smoke, common make command smoke, Redis integration, SQLite smoke, host-only build smoke, and Redis cluster smoke. The release workflow SHALL NOT run host-only E2E or plugin-full E2E; full E2E verification is covered by the nightly workflow.

#### Scenario: Release tag triggers brief tests before image publishing

- **WHEN** GitHub Actions receives a tag push event
- **THEN** the release workflow first completes release tag and framework version validation
- **AND** the release workflow calls `.github/workflows/reusable-test-verification-suite.yml`
- **AND** the shared test suite uses Main CI's brief test scope without E2E
- **AND** the release image publishing job depends on tag validation and shared test suite via `needs`
- **AND** GHCR login, `make image push=1`, `latest` tag publishing, and remote manifest inspect only execute after all test jobs succeed

#### Scenario: Release reuses shared test template

- **WHEN** nightly workflow and main CI workflow orchestrate verification jobs through the shared test suite
- **THEN** the release workflow SHALL orchestrate verification jobs through the same `reusable-test-verification-suite.yml`
- **AND** the release workflow MUST NOT redundantly inline-expand verification jobs already encapsulated by the shared test suite

#### Scenario: Release does not run full E2E

- **WHEN** the release workflow calls the shared test suite
- **THEN** `include-host-only-e2e-tests` SHALL be `false`
- **AND** `include-plugin-full-e2e-tests` SHALL be `false`
- **AND** the release workflow's image publishing dependency SHALL NOT wait for separate E2E jobs

#### Scenario: Any brief test failure blocks release image push

- **WHEN** any test job in the release workflow fails, is cancelled, or times out
- **THEN** the release image publishing job MUST NOT execute
- **AND** the workflow MUST NOT push the release tag corresponding image
- **AND** the workflow MUST NOT update the `latest` floating image tag

#### Scenario: Release workflow name expresses test and build responsibility

- **WHEN** the repository maintains the release publishing workflow
- **THEN** the workflow filename uses `.github/workflows/release-test-and-build.yml`
- **AND** the workflow display name is `Release Test and Build`
- **AND** the old `release-build.yml` with overlapping responsibilities is not retained

### Requirement: Release plugin image must retain official plugin brief verification

The system SHALL treat the release publishing pipeline as a plugin-full image publishing path. Before publishing a full plugin image, the workflow SHALL run plugin-full Windows command smoke, Go unit tests, and frontend unit tests through the shared test suite, and use the official plugin workspace to build the plugin-full image during the image build phase.

#### Scenario: Official plugin workspace exists and brief verification continues

- **WHEN** the release workflow checkout completes
- **AND** `apps/lina-plugins` contains official plugin directories and `plugin.yaml`
- **THEN** the workflow continues executing plugin-full Windows command smoke, Go unit tests, frontend unit tests, and image build preflight verification
- **AND** official plugin source code is within the release brief verification scope

#### Scenario: Official plugin workspace missing causes fast failure

- **WHEN** the release workflow needs to publish a full plugin image
- **AND** `apps/lina-plugins` does not exist, is empty, or lacks official plugin manifests
- **THEN** plugin-full brief verification or image building fails before publishing
- **AND** the error message states that the official plugin workspace is currently missing
- **AND** the error message includes a hint to initialize the submodule

#### Scenario: Release image build includes official plugins

- **WHEN** the release workflow publishes multi-architecture images
- **THEN** the `release-image` job SHALL continue using plugin-full build configuration
- **AND** the published release tag and `latest` floating tag SHALL correspond to images containing official plugins

### Requirement: Release workflow must create GitHub Release after publishing gates complete

The system SHALL create a GitHub Release for the triggering Git tag after release tag validation, shared test verification suite, and GHCR image publishing all succeed. The Release title SHALL use the format `LinaPro Release <tag>` where `<tag>` is the current triggering tag name.

#### Scenario: GitHub Release created after successful publishing

- **WHEN** GitHub Actions receives a tag push event
- **AND** release tag validation succeeds
- **AND** shared test verification suite succeeds
- **AND** release image publishing succeeds
- **THEN** the release workflow creates a GitHub Release for the current tag
- **AND** the Release title is `LinaPro Release <tag>`

#### Scenario: GitHub Release not created when publishing gates fail

- **WHEN** release tag validation, shared test verification suite, or release image publishing fails, is cancelled, or times out
- **THEN** the release workflow MUST NOT create a GitHub Release

### Requirement: Nightly image publishing must support a manual no-test entrypoint

The system SHALL provide an independent GitHub Actions manual workflow for building and publishing nightly images. The workflow MUST be triggered only by `workflow_dispatch`, MUST directly call the unified image publish workflow, MUST NOT depend on the test verification suite, unit tests, E2E tests, smoke tests, or other pre-test jobs. The existing scheduled nightly workflow MUST continue to retain test gates.

#### Scenario: Manual trigger directly publishes nightly image

- **WHEN** a maintainer manually triggers the no-test nightly image publish workflow via GitHub Actions
- **THEN** the workflow directly calls the unified image publish workflow to build and push `linux/amd64` and `linux/arm64` multi-architecture images
- **AND** the workflow publishes the date-based `nightly-<yyyymmdd>` immutable tag and `nightly` floating tag
- **AND** the workflow does not wait for any test verification job to complete

#### Scenario: Scheduled nightly continues to be protected by test gates

- **WHEN** the existing nightly workflow is triggered by schedule
- **THEN** the workflow continues to first run the shared test verification suite
- **AND** nightly images are only published after the test verification suite succeeds

### Requirement: Nightly demo image must provide a memory-only Compose launcher

The system SHALL provide a Docker Compose launcher in `hack/deploy/docker-compose.yaml` for demonstrating the nightly image. The launcher MUST use the published `linapro` image, MUST use a PostgreSQL service as the demo database, MUST NOT mount host data directories or declare persistent volumes, MUST place application runtime data and PostgreSQL data directories in memory-only `tmpfs`, MUST maintain runtime configuration separately in `hack/deploy/config.yaml` and inject it as read-only config, MUST complete database initialization and mock demo data loading after PostgreSQL becomes healthy before starting the HTTP service, and MUST include necessary comments explaining image/port overrides, memory-only data, read-only config injection, database dependency, startup initialization order, and demo protection plugin usage.

#### Scenario: Start memory-only demo environment

- **WHEN** an evaluator runs `docker compose -f hack/deploy/docker-compose.yaml up`
- **THEN** Compose starts the `linapro` demo service and exposes port `8080`
- **AND** the application reads read-only runtime configuration from `hack/deploy/config.yaml`
- **AND** the application connects to the Compose-internal PostgreSQL service as its database
- **AND** application runtime data and PostgreSQL data are written to container-internal `tmpfs`
- **AND** stopping and removing containers does not preserve demo data on host disk volumes

### Requirement: Deployment test Compose must provide a manual development container

The system SHALL provide a Docker Compose test entry in `hack/deploy/tests/docker-compose.yaml` for manually verifying development commands. The entry MUST start a PostgreSQL service, MUST start a long-running development container based on `loads/ubuntu:24.04-npm`, MUST mount the current repository to the container working directory, MUST use `GF_GCFG_PATH` to point to the `hack/deploy/tests` config directory and read its `config.yaml`, MUST wait for PostgreSQL to become healthy before keeping the development container enterable, and MUST NOT automatically execute `lina init`, `lina mock`, or start the LinaPro HTTP service.

#### Scenario: Enter development container to manually execute verification commands

- **WHEN** a developer runs `docker compose -f hack/deploy/tests/docker-compose.yaml up -d`
- **THEN** Compose starts a one-time PostgreSQL service
- **AND** Compose starts a `loads/ubuntu:24.04-npm` development container with the current repository mounted
- **AND** the developer can enter the container via `docker compose -f hack/deploy/tests/docker-compose.yaml exec dev bash`
- **AND** the developer manually executes initialization, building, testing, or other development verification commands inside the container
