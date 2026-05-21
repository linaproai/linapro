# Design

## Context

The monthly archive workflow already has deterministic candidate detection. The weak point is the base archive execution and later failure handling: earlier revisions allowed selected archive phases to fail and still continue by restoring a deterministic snapshot or emitting a manual PR URL. In the failed GitHub Actions run, GitHub Copilot CLI returned success but completed active changes remained under `openspec/changes/`, and a later run showed archive consolidation could fail while the workflow still finalized a PR without consolidation.

A local replay at the failed commit showed that `openspec archive -y` can archive most candidates, while `remove-sqlite-support` fails because one REMOVED requirement header does not exist in the current baseline spec. The monthly workflow should still keep archive execution inside the configured tool runtime so the base archive phase and the archive consolidation phase use the same Codex, Claude Code, or GitHub Copilot CLI execution model. It must also use GitHub Actions' default fail-fast behavior: once any archive, consolidation, validation, push, or PR step fails, no later archive steps should run.

## Approach

Add a shared `.github/prompts/monthly-openspec-auto-archive.zh-CN.md` prompt that instructs the selected AI coding tool to invoke the repository's `lina-auto-archive` skill. The prompt requires active-change scanning, conservative completion checks, `openspec archive -y "<change-name>"` for each safe candidate, post-archive active-list confirmation, `openspec/**`-only changes, and no human interaction in CI.

The tool-specific reusable workflows prepare their own runtime first, then run `Run Lina Auto Archive` through that runtime:

- Codex uses `loads/codex:latest` and `codex exec`.
- Claude Code uses `loads/cc:latest` and the existing `run-cc-task` wrapper.
- GitHub Copilot CLI uses `@github/copilot` and `copilot -p`.

After the selected tool completes auto-archive, the workflow detects OpenSpec diffs, validates the archive result, and only then runs archive consolidation through the same tool runtime. If auto-archive fails, the workflow stops at that tool step and does not continue to consolidation or PR finalization.

The tool-specific reusable workflows must not use `continue-on-error` around auto-archive, archive consolidation, temporary change cleanup, or validation. They must also avoid rollback-to-snapshot behavior after consolidation fails because that hides the actual failed phase and produces an archive PR that does not reflect the requested aggregation. Pull request creation and update failures are hard failures, including repository policy errors that block `GITHUB_TOKEN` from creating or editing pull requests.

## Scope

This is CI/OpenSpec governance only. It does not change runtime product behavior, HTTP APIs, backend Go production code, data permissions, runtime i18n, or cache behavior.

## Known Blocker Fix

`remove-sqlite-support` currently contains a REMOVED delta for a requirement title that is not present in `openspec/specs/cluster-coordination-config/spec.md`. The current baseline already expresses the PostgreSQL-only behavior in `Requirement: 非 PostgreSQL 数据库链接必须在 coordination 启动前失败`, so the delta should be changed from REMOVED to MODIFIED against that existing requirement.
