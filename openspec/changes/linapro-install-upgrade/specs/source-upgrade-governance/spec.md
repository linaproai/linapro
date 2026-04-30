## MODIFIED Requirements

### Requirement: Source upgrades must provide a unified development-time entry point

The framework SHALL provide a unified development-time source upgrade entry point through the `lina-upgrade` Claude Code skill located at `.claude/skills/lina-upgrade/`. The skill MUST support both **framework upgrades** (covering changes to `apps/lina-core/` and `apps/lina-vben/`) and **source-plugin upgrades** (covering changes to `apps/lina-plugins/<plugin-id>/`). The implementation MUST be entirely script-driven (`bash` + `git` + `make`) and MUST NOT introduce a new binary CLI, a new project-level configuration file, or runtime commands inside `lina-core`.

The skill MUST be invocable by AI tools (`Claude Code`, `Codex`, etc.) so that the user only needs to state the upgrade intent in natural language ("upgrade framework to v0.6.0", "upgrade plugin plugin-demo to v0.5.0") and the AI orchestrates the workflow.

#### Scenario: Run a framework upgrade through the skill

- **WHEN** a user instructs the AI to run a framework upgrade to a specified target version
- **THEN** the AI invokes the `lina-upgrade` skill in framework-upgrade sub-flow
- **AND** the skill performs version checks, plan generation, code merge, code regeneration, SQL migration, and verification

#### Scenario: Run a source-plugin upgrade through the skill

- **WHEN** a user instructs the AI to upgrade a specific source plugin (e.g., `plugin-demo`) to a specified target version
- **THEN** the AI invokes the `lina-upgrade` skill in source-plugin-upgrade sub-flow
- **AND** the skill enters source-plugin upgrade planning and execution without triggering framework merge logic

#### Scenario: Skill rejects scope ambiguity

- **WHEN** a user gives an ambiguous upgrade instruction that does not clearly indicate framework or source-plugin scope
- **THEN** the AI asks the user to clarify which scope to upgrade
- **AND** does not start any merge or migration action until the scope is confirmed

### Requirement: Source upgrades must complete safety checks before execution

The unified upgrade skill SHALL verify that the Git worktree is clean before any upgrade starts. If tracked files are modified, unstaged, or uncommitted, the skill MUST refuse to continue and instruct the user to commit or stash. The skill MUST also remind the user to back up the database before continuing for any upgrade that involves SQL migration.

#### Scenario: The worktree contains local changes before an upgrade

- **WHEN** a user invokes the upgrade skill while the Git worktree contains modified, unstaged, or uncommitted files
- **THEN** the skill refuses to continue
- **AND** it tells the user to commit or stash the current changes first

#### Scenario: The skill prompts for database backup before SQL migration

- **WHEN** the skill detects that the upgrade introduces new SQL files (numbered higher than the current maximum)
- **THEN** the skill prints a database backup reminder before executing migrations
- **AND** in non-interactive mode (`LINAPRO_NON_INTERACTIVE=1`), the skill records the reminder in its output report instead of waiting for user confirmation

## REMOVED Requirements

### Requirement: Framework upgrades must read upgrade metadata only from hack config

**Reason**: The new upgrade workflow uses `apps/lina-core/manifest/config/metadata.yaml.framework.version` as the single source of truth for the current framework baseline version. The previous `apps/lina-core/hack/config.yaml.frameworkUpgrade.version` field is removed because (1) it duplicates information already maintained in `metadata.yaml`, (2) it requires synchronization with system-info display values, and (3) it does not align with the AI-driven upgrade flow that needs the same value for both display and upgrade reasoning.

**Migration**: Any code that previously read `frameworkUpgrade.version` from `apps/lina-core/hack/config.yaml` MUST be updated to read `framework.version` from `apps/lina-core/manifest/config/metadata.yaml`. The `frameworkUpgrade` section of `hack/config.yaml` SHOULD be removed entirely; if `frameworkUpgrade.repositoryUrl` was used to suggest a default upstream URL, that suggestion SHOULD be hard-coded into the install scripts or removed (the install scripts default to `https://github.com/linaproai/linapro.git`). The upgrade skill MUST NOT read host runtime configuration for upgrade metadata.

### Requirement: Framework upgrades must replay all host SQL from the first file

**Reason**: Replaying every SQL file from the first one is destructive when the user has populated the database with real data — it forces re-running DDL and seed inserts that may conflict with existing rows even with `INSERT IGNORE` semantics, and it is significantly slower than necessary. The new upgrade workflow uses incremental migration: only SQL files with index higher than the current maximum (computed from the local `manifest/sql/` directory before merge) are executed. Idempotency requirements on SQL files (per project convention) ensure new SQL files are safe to re-run if needed.

**Migration**: Any operator workflow or CI step that previously assumed full SQL replay MUST be updated to use the new incremental migration approach. The `make init` command continues to handle initial database setup; for upgrades, the skill computes the diff between the SQL files at the baseline tag and at HEAD, and runs only the new files in numerical order.

## ADDED Requirements

### Requirement: Upgrade skill must perform four-layer baseline validation before any merge

The upgrade skill SHALL invoke `.claude/skills/lina-upgrade/scripts/upgrade-baseline-check.sh` as the first step after passing the worktree clean check. The script MUST perform four layers of validation against `apps/lina-core/manifest/config/metadata.yaml.framework.version` (the declared baseline):

| Layer | Check | Failure Code |
| --- | --- | --- |
| 1. Existence | The declared version is a real upstream tag | `ERR_TAG_NOT_FOUND` |
| 2. Reachability | HEAD descends from the declared tag commit (`git merge-base --is-ancestor`) | `ERR_HEAD_NOT_DESCENDANT` |
| 3. Identity (soft) | SQL count and key path presence are consistent with the tag (warning only) | (warning) |
| 4. Summary | Outputs `OK_BASELINE_CONFIRMED` plus `commits_ahead`, `core_changed`, `sql_at_tag`, `sql_at_head` metrics | (none) |

The script MUST be a side-effect-free pure function: it reads `git` state plus `metadata.yaml` and writes nothing.

#### Scenario: Baseline matches an upstream tag with HEAD as descendant

- **WHEN** the skill invokes `upgrade-baseline-check.sh`
- **AND** `metadata.yaml.framework.version` matches a real upstream tag
- **AND** HEAD is a descendant of that tag commit
- **THEN** the script outputs `OK_BASELINE_CONFIRMED` plus the summary metrics
- **AND** the skill proceeds to the upgrade plan stage

#### Scenario: Declared version does not exist on upstream

- **WHEN** the skill invokes `upgrade-baseline-check.sh`
- **AND** `metadata.yaml.framework.version` does not match any upstream tag
- **THEN** the script outputs `ERR_TAG_NOT_FOUND` along with the three most recent stable upstream tags as candidates
- **AND** the AI asks the user to confirm the actual baseline before proceeding

#### Scenario: HEAD is not a descendant of the declared tag

- **WHEN** the skill invokes `upgrade-baseline-check.sh`
- **AND** `metadata.yaml.framework.version` matches an upstream tag
- **AND** HEAD is not a descendant of that tag commit
- **THEN** the script outputs `ERR_HEAD_NOT_DESCENDANT` with the tag commit hash and the HEAD commit hash
- **AND** the AI asks the user whether they have done a `rebase` or `reset` and what the actual baseline should be

### Requirement: Upgrade skill must classify file changes by stability tier

The upgrade skill SHALL use `.claude/skills/lina-upgrade/scripts/upgrade-classify.sh` to map every changed file path into one of three stability tiers and apply tier-specific conflict resolution strategy:

| Tier | Path patterns | Conflict resolution |
| --- | --- | --- |
| Tier 1 (stable contract) | `apps/lina-core/pkg/bizerr/**`, `apps/lina-core/pkg/logger/**`, `apps/lina-core/pkg/contract/**`, public plugin runtime API surfaces, `apps/lina-plugins/<your-plugin>/**` (user-owned plugin directories) | Should not conflict; on conflict, escalate to human |
| Tier 2 (user-modifiable, conflicts on the user) | `apps/lina-core/internal/**` (excluding generated paths), `apps/lina-vben/apps/web-antd/src/**` (excluding generated paths), `apps/lina-core/manifest/config/*.yaml` (excluding the `framework.version` field) | Three-way merge; if AI confidence is low, escalate to human |
| Tier 3 (auto-generated, regenerate on upgrade) | `apps/lina-core/internal/dao/**`, `apps/lina-core/internal/model/{do,entity}/**`, `apps/lina-core/internal/controller/**` (skeleton portions), and the equivalent paths inside source plugins | `git checkout --theirs <path>`, then run `make dao` / `make ctrl` to regenerate |

#### Scenario: Conflict in Tier 3 path is auto-resolved by regeneration

- **WHEN** `git merge` produces a conflict in `apps/lina-core/internal/dao/sys_user_dao.go`
- **THEN** the skill runs `git checkout --theirs apps/lina-core/internal/dao/sys_user_dao.go`
- **AND** subsequently invokes `make dao` to regenerate the file from the updated SQL schema

#### Scenario: Conflict in Tier 1 path triggers escalation

- **WHEN** `git merge` produces a conflict in `apps/lina-core/pkg/bizerr/code.go`
- **THEN** the skill stops automatic resolution
- **AND** reports the conflict to the user with the rationale "Tier 1 contract change requires manual review"

#### Scenario: Tier 2 conflict with low AI confidence triggers escalation

- **WHEN** `git merge` produces a conflict in `apps/lina-core/internal/service/auth/auth.go`
- **AND** the AI cannot confidently determine a three-way merge result
- **THEN** the skill leaves the conflict markers in place
- **AND** reports the file path and conflict context to the user for manual resolution

### Requirement: Upgrade skill must follow a fixed ten-step workflow

The upgrade skill SHALL execute the following ten steps in strict order on every invocation, halting on any step that requires human intervention:

1. **Pre-flight guard**: verify `git status` is clean, `metadata.yaml` exists, target version is greater than declared baseline.
2. **Baseline validation**: invoke `upgrade-baseline-check.sh`; on failure, dialogue with the user.
3. **Plan generation**: invoke `upgrade-plan.sh` to produce a list of changes (changelog excerpts from `CHANGELOG.md` and `openspec/changes/archive/`, modified file list from `git diff baseline...HEAD`, new SQL files); present the plan to the user for confirmation.
4. **Merge execution**: run `git merge --no-commit upstream/<target>` (or equivalent `cherry-pick` strategy if explicitly requested).
5. **Conflict handling**: classify conflicts per `upgrade-classify.sh`; auto-resolve where possible per tier rules; escalate to human otherwise.
6. **Regeneration**: invoke `upgrade-regenerate.sh` to run `make dao` and `make ctrl`.
7. **Database migration**: run `make init` to apply incremental SQL files (those numbered higher than the previous maximum).
8. **Verification**: invoke `upgrade-verify.sh` to run `go build`, `pnpm typecheck`, `pnpm lint`, and an e2e smoke subset (login + simple CRUD + plugin loading); on failure, halt and escalate.
9. **Commit**: run `git commit -m "chore: upgrade to v<target>"`. The `metadata.yaml.framework.version` is already overwritten by the upstream merge.
10. **Report**: produce a summary listing what was auto-resolved, what required manual handling, and what remains for the user to verify.

#### Scenario: Successful upgrade completes all ten steps automatically

- **WHEN** the upgrade target has no Tier 1 conflicts and Tier 2 conflicts are all auto-resolvable
- **THEN** the skill walks through all ten steps without prompting the user beyond the plan-confirmation step
- **AND** produces a final report listing the changes applied

#### Scenario: Upgrade halts on verification failure

- **WHEN** any of `go build`, `pnpm typecheck`, `pnpm lint`, or the e2e smoke subset fails after all conflicts are resolved
- **THEN** the skill stops at step 8 without committing
- **AND** reports the failure with stack trace / log snippet for the user to investigate
- **AND** does not auto-revert the merged state, leaving the user to either fix or `git merge --abort`

### Requirement: Upgrade skill must publish escalation rules document

The repository SHALL maintain `.claude/skills/lina-upgrade/references/escalation-rules.md` listing the explicit conditions under which the skill must escalate to human intervention. The document MUST include at least the following five rules:

1. Conflict in any Tier 1 path
2. Tier 2 three-way merge with low AI confidence or semantic ambiguity
3. SQL migration that risks destroying user data (e.g., `DROP COLUMN` against a column with existing values)
4. e2e smoke failure that auto-rollback cannot recover
5. User-declared modified paths overlap with upstream changes for the same paths

#### Scenario: Document exists and lists all five rules

- **WHEN** the upgrade skill is invoked
- **THEN** the skill references `escalation-rules.md` for the canonical list of escalation conditions
- **AND** the document at `.claude/skills/lina-upgrade/references/escalation-rules.md` contains all five rules with examples

### Requirement: Upgrade skill must publish file-tier classification document

The repository SHALL maintain `.claude/skills/lina-upgrade/references/tier-classification.md` listing the canonical Tier 1 / Tier 2 / Tier 3 path patterns. The document MUST be machine-readable enough that `upgrade-classify.sh` can be derived from or validated against it.

#### Scenario: New path added to repository is automatically classified

- **WHEN** an upgrade introduces a new path (e.g., `apps/lina-core/pkg/cache/`)
- **THEN** the skill consults `tier-classification.md` to determine the tier
- **AND** falls back to Tier 2 (the safest default for unknown paths under `apps/lina-core/internal/`) when no rule matches

### Requirement: Upgrade skill must publish changelog conventions document

The repository SHALL maintain `.claude/skills/lina-upgrade/references/changelog-conventions.md` describing how the upgrade plan stage parses framework changes from two sources:

1. `CHANGELOG.md` at the repository root (if present)
2. `openspec/changes/archive/<change-id>/proposal.md` for changes whose archived directory falls within the version range being upgraded

The document MUST specify how to detect breaking changes (markers like `**BREAKING**` in proposals) and how to surface Tier 1 changes prominently in the upgrade plan.

#### Scenario: Plan stage extracts changelog from OpenSpec archive

- **WHEN** the upgrade target is `v0.6.0` and the baseline is `v0.5.0`
- **THEN** `upgrade-plan.sh` enumerates archived changes whose archive timestamp falls between the two tag dates
- **AND** extracts each archived change's "Why" / "What Changes" sections as part of the plan output

#### Scenario: Breaking change marker is surfaced

- **WHEN** an archived change in the upgrade range contains `**BREAKING**` markers in its `What Changes` section
- **THEN** the upgrade plan output prominently lists the breaking changes
- **AND** when the breaking change touches Tier 1 paths, the plan flags it for mandatory user review
