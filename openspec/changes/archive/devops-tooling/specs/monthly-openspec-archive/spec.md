## ADDED Requirements

### Requirement: Monthly workflow must archive completed OpenSpec changes

The system SHALL provide a GitHub Actions monthly workflow for automatically scanning completed and unarchived active changes in `openspec/changes/` on the default branch and executing `lina-auto-archive` skill archiving through a configurable AI Coding tool.

#### Scenario: Scheduled archive run
- **WHEN** the monthly OpenSpec archive workflow triggers on schedule
- **AND** the current Asia/Shanghai date is the 1st of the month
- **THEN** the workflow checks out code on the repository default branch
- **AND** the workflow uses the AI Coding tool specified by `AI_CODING_TOOL` to run the archive task
- **AND** the workflow calls `lina-auto-archive` to scan and archive automatically processable completed changes

#### Scenario: Monthly schedule window
- **WHEN** GitHub Actions schedule event triggers in UTC
- **THEN** the workflow uses UTC month-end cron groups covering Asia/Shanghai 1st-of-month 00:00
- **AND** the workflow skips the `2/28 16:00 UTC` duplicate schedule event in leap years
- **AND** the workflow uses `2/28 16:00 UTC` to cover Asia/Shanghai March 1st 00:00 in non-leap years
- **AND** the workflow uses `2/29 16:00 UTC` when it exists to cover Asia/Shanghai March 1st 00:00 in leap years

#### Scenario: Manual archive run
- **WHEN** a maintainer manually triggers the monthly OpenSpec archive workflow via `workflow_dispatch`
- **THEN** the workflow is not restricted by the monthly schedule window
- **AND** the workflow continues executing default branch restriction, OpenSpec completion candidate pre-check, and AI Coding tool routing

#### Scenario: No completed active changes
- **WHEN** the monthly OpenSpec archive workflow triggers
- **AND** `openspec list --json` reports no `complete`, `completed`, or `done` status active changes
- **THEN** the workflow does not call the AI Coding tool archive task
- **AND** the workflow ends successfully without creating or updating an archive PR

### Requirement: Monthly workflow must consolidate only after new archive changes

The system SHALL only execute `lina-archive-consolidate` skill after this monthly auto-archive produces OpenSpec file changes, avoiding redundant rewriting of consolidated archive documents when no new archives exist.

#### Scenario: Archive produced changes
- **WHEN** `openspec/` has new file changes after `lina-auto-archive` executes
- **THEN** the workflow calls `lina-archive-consolidate` to consolidate archived changes
- **AND** the workflow continues executing OpenSpec validation after consolidation

#### Scenario: Archive produced no changes
- **WHEN** `lina-auto-archive` execution completes
- **AND** `openspec/` has no new file changes
- **THEN** the workflow skips `lina-archive-consolidate`
- **AND** the workflow does not create or update an archive PR

### Requirement: Monthly workflow must select the AI Coding tool from GitHub Variables

The system SHALL select the AI Coding tool for monthly OpenSpec archiving through `AI_CODING_TOOL` in GitHub Variables, and SHALL default to `codex` when the variable is not configured.

#### Scenario: Default Codex tool
- **WHEN** the monthly OpenSpec archive workflow triggers
- **AND** GitHub Variables does not configure `AI_CODING_TOOL`
- **THEN** the main workflow calls the Codex reusable workflow
- **AND** the Codex reusable workflow uses `loads/codex:latest` and `codex exec` to run AI tasks

#### Scenario: Explicit Claude Code tool
- **WHEN** the monthly OpenSpec archive workflow triggers
- **AND** GitHub Variables `AI_CODING_TOOL` is `cc`
- **THEN** the main workflow calls the Claude Code reusable workflow
- **AND** the Claude Code reusable workflow uses `loads/cc:latest` and `claude -p` to run AI tasks

#### Scenario: Explicit GitHub Copilot CLI tool
- **WHEN** the monthly OpenSpec archive workflow triggers
- **AND** GitHub Variables `AI_CODING_TOOL` is `copilot`
- **THEN** the main workflow calls the GitHub Copilot CLI reusable workflow
- **AND** the GitHub Copilot CLI reusable workflow uses `@github/copilot` and `copilot -p` to run AI tasks
- **AND** the GitHub Copilot CLI reusable workflow uses the `COPILOT_MODEL` variable for model configuration, defaulting to `auto` when not configured
- **AND** the GitHub Copilot CLI reusable workflow uses the `COPILOT_REASONING_EFFORT` variable for reasoning effort configuration, not passing explicit reasoning effort when not configured
- **AND** the workflow only accepts empty, `low`, `medium`, `high`, or `xhigh` as Copilot reasoning effort

#### Scenario: Unsupported tool value
- **WHEN** the monthly OpenSpec archive workflow triggers
- **AND** GitHub Variables `AI_CODING_TOOL` is not `codex`, `cc`, or `copilot`
- **THEN** the main workflow fails before executing any tool reusable workflow
- **AND** the workflow does not create or update an archive PR

### Requirement: Monthly workflow must isolate tool implementations in reusable workflows

The system SHALL encapsulate different AI Coding tools' runtime preparation, image invocation, authentication configuration, and log upload details in tool-specific reusable workflows, and SHALL let the main workflow only handle triggering, candidate detection, and routing.

#### Scenario: Only one tool workflow runs
- **WHEN** the monthly OpenSpec archive workflow triggers
- **AND** `AI_CODING_TOOL` is any valid value
- **THEN** the workflow only runs the matching tool's reusable workflow
- **AND** the workflow does not run other tools' reusable workflows

### Requirement: Monthly workflow must share prompt files across AI tools

The system SHALL maintain monthly OpenSpec auto-archive and archive consolidation prompts as public files under `.github/prompts/`, and SHALL let all tool-specific reusable workflows reference the same prompt content.

### Requirement: Monthly workflow must stream AI tool execution logs

The system SHALL stream the AI Coding tool process's stdout and stderr to GitHub Actions step logs during monthly OpenSpec auto-archive and archive consolidation execution, while continuing to preserve artifact logs for post-hoc review.

### Requirement: Monthly workflow must fail fast after each archive phase

The system SHALL execute deterministic phase checks after the monthly OpenSpec auto-archive and archive consolidation phases; when any phase fails, produces invalid OpenSpec state, or does not achieve the archiving results it should, the workflow MUST immediately fail and stop subsequent phases.

#### Scenario: Auto archive leaves completed changes active
- **WHEN** `openspec list --json` still reports `complete`, `completed`, or `done` status active changes after auto-archive
- **THEN** the workflow fails before executing change detection or archive consolidation
- **AND** the workflow outputs the names, statuses, and task counts of unarchived changes

#### Scenario: Auto archive produces invalid OpenSpec state
- **WHEN** `openspec validate --all` fails after auto-archive
- **THEN** the workflow fails before executing change detection or archive consolidation

### Requirement: Monthly workflow must inject AI tool credentials and endpoint at runtime

The system SHALL generate the selected AI Coding tool's authentication configuration and inject provider `base_url` at runtime through GitHub Secrets and Variables, and SHALL NOT write real API keys/tokens or real `base_url` into `.github/codex` or `.github/cc` configuration files in the repository.

### Requirement: Monthly workflow must guard generated changes

The system SHALL verify the change scope before creating or updating an archive PR, and only allow OpenSpec archive governance files to be modified by the monthly automated task.

#### Scenario: Allowed OpenSpec changes
- **WHEN** workspace changes only contain `openspec/**`
- **THEN** the workflow can create or update an archive PR targeting the repository default branch

#### Scenario: Unexpected file changes
- **WHEN** workspace changes exist outside the allowed scope
- **THEN** the workflow fails
- **AND** the workflow does not create or update an archive PR

### Requirement: Monthly workflow must validate OpenSpec artifacts before creating a pull request

The system SHALL execute OpenSpec validation before creating or updating an archive PR; validation failure MUST stop PR write-back.
