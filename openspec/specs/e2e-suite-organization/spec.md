# E2E Suite Organization

## Purpose

Define the directory ownership, helper placement, and TC-governance rules of the Playwright E2E suite so the test tree remains aligned with stable LinaPro capability boundaries and stays easy to maintain.

## Requirements

### Requirement: E2E test cases MUST be organized by stable capability boundaries
The E2E suite SHALL organize test directories by the current stable workbench capability boundaries and plugin ownership. It MUST NOT continue to pile most capability tests into an overloaded legacy catch-all directory. Second-level directories MAY be used for finer-grained capability splits, but the first-level directories MUST still reflect stable capability boundaries.

#### Scenario: Host-owned capability tests land in the matching capability directory
- **WHEN** a team adds or migrates a host-owned capability test file
- **THEN** that file MUST land in a directory aligned with the current workbench capability boundary, such as `iam/`, `settings/`, `scheduler/`, `extension/`, `dashboard/`, or `about/`

#### Scenario: Plugin-owned capability tests land in the matching plugin capability directory
- **WHEN** a team adds or migrates a plugin capability test file
- **THEN** that file MUST land in a directory that expresses the plugin capability boundary, such as `monitor/operlog/`, `monitor/loginlog/`, `org/dept/`, or `content/notice/`

#### Scenario: Second-level directories express subdomains instead of reviving legacy buckets
- **WHEN** a capability contains multiple clear subdomains
- **THEN** the suite MAY use second-level directories to express those subdomains
- **AND** it MUST NOT reintroduce a new overloaded catch-all directory in place of the stable capability boundary

### Requirement: Non-test files MUST NOT be mixed into the E2E test tree
The `hack/tests/e2e/` directory tree SHALL contain only real test-case files. Shared helpers, wait utilities, debug scripts, and execution-governance scripts MUST live in dedicated support directories and MUST NOT be mixed with `TC*.ts` files.

#### Scenario: Shared helpers live in support directories
- **WHEN** tests need shared API helpers, wait utilities, or data builders
- **THEN** those files MUST live in `fixtures/`, `support/`, `scripts/`, or an equivalent dedicated support directory
- **AND** they MUST NOT live under `hack/tests/e2e/`

#### Scenario: Debug scripts do not pollute test discovery
- **WHEN** the team adds a temporary debug or investigation script
- **THEN** that file MUST live in a dedicated debug directory
- **AND** it MUST NOT appear in the E2E discovery scope

### Requirement: TC numbering and directory ownership MUST be automatically validated
The E2E suite SHALL provide automated inventory and validation to check TC naming, global uniqueness, and directory ownership so duplicate TC IDs, invalid files, and misplaced tests do not linger in the repository.

#### Scenario: TC identifiers are globally unique
- **WHEN** the validator scans all `TC*.ts` files
- **THEN** the system MUST detect and report any duplicate TC identifier

#### Scenario: Invalid files are reported automatically
- **WHEN** the validator scans `hack/tests/e2e/`
- **THEN** the system MUST report any file that does not follow the `TC{NNNN}-{brief-name}.ts` convention
- **AND** it MUST report any test file that lives outside the allowed capability-directory mapping
