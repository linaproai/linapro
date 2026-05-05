---
name: lina-upgrade
description: Upgrade LinaPro framework or source plugins through an AI-guided, script-driven workflow covering upgrade, linapro, framework, and plugin tasks.
---

# Purpose

`lina-upgrade` guides AI tools through development-time upgrades for `LinaPro`.
It replaces the removed `make upgrade` entry point and keeps upgrades script-driven with `bash`, `git`, and `make`.

The skill has two scopes:

- `framework`: upgrade `apps/lina-core/`, `apps/lina-vben/`, shared manifests, and generated host artifacts.
- `source-plugin`: upgrade one source plugin or all source plugins under `apps/lina-plugins/`.

It does not upgrade dynamic plugins and does not introduce a new binary `CLI` or project-level configuration file.

# When to Invoke

Invoke this skill when the user asks for any of these intents:

- "upgrade LinaPro to `v0.6.0`"
- "upgrade the framework"
- "run framework upgrade"
- "升级 LinaPro 框架到 `v0.6.0`"
- "升级源码插件 `plugin-demo-source`"
- "upgrade source plugin `plugin-demo-source`"
- "upgrade all source plugins"

If the request does not clearly say `framework` or `source-plugin`, ask one concise clarification question before running scripts.

# Inputs the AI Must Collect from the User

Prerequisite: ensure `gf` and `openspec` are installed; invoke `lina-doctor` first if any tool is missing.

| Input | Required When | Notes |
| --- | --- | --- |
| Target version | `framework` upgrades | Must be greater than `apps/lina-core/manifest/config/metadata.yaml.framework.version`. |
| Scope | Always | Must be `framework` or `source-plugin`. |
| Plugin ID | `source-plugin` upgrades | Use a concrete plugin ID or `all` for bulk source-plugin upgrade. |
| Database backup confirmation | SQL migrations are detected | Print the backup reminder even in non-interactive mode. |

# Workflow

Run the following steps in order. Stop when a step requires human intervention.

1. **Pre-flight guard**: verify `git status --short` is clean, `metadata.yaml` exists, and target version is greater than the declared baseline. Failure mode: dirty worktree or ambiguous target; ask the user to commit, stash, or clarify.
2. **Baseline validation**: run `scripts/upgrade-baseline-check.sh`. Failure mode: `ERR_TAG_NOT_FOUND` or `ERR_HEAD_NOT_DESCENDANT`; show the script output and ask the user to confirm the actual baseline.
3. **Plan generation**: run `scripts/upgrade-plan.sh <target-version>`. Failure mode: missing target tag or unreadable changelog; fix inputs before continuing. Present the plan to the user for confirmation.
4. **Merge execution**: run `git merge --no-commit upstream/<target>` or the equivalent remote ref selected in the plan. Failure mode: merge conflicts; continue to conflict handling.
5. **Conflict handling**: classify each conflicted path with `scripts/upgrade-classify.sh <path>`. Tier 1 conflicts always escalate. Tier 2 conflicts require AI confidence; low confidence escalates. Tier 3 conflicts use `git checkout --theirs <path>`.
6. **Regeneration**: run `scripts/upgrade-regenerate.sh`. Failure mode: `make dao` or `make ctrl` fails; report the log path and stop.
7. **Database migration**: remind the user to back up the database, then run `make init confirm=init` for the incremental SQL set. Failure mode: migration risk or destructive SQL; escalate.
8. **Verification**: run `scripts/upgrade-verify.sh`. Failure mode: build, typecheck, lint, or smoke test failure; stop and report the failing command.
9. **Commit**: run `git commit -m "chore: upgrade to <target-version>"` after verification passes. Failure mode: user asks not to commit; leave changes staged or unstaged as requested.
10. **Report**: summarize automatic resolutions, manual decisions, migration status, test results, and remaining user checks.

# Source-Plugin Sub-flow

For `source-plugin` upgrades, skip framework merge steps. Inspect `apps/lina-plugins/<plugin-id>/plugin.yaml` or all plugin manifests, compare discovered versions with the effective installed versions reported by host governance data, then execute the existing explicit source-plugin upgrade service path through host commands or focused scripts. Dynamic plugins are ignored.

# Outputs the AI Must Produce

- Upgrade plan before merge: target version, baseline version, changed paths grouped by tier, SQL migration summary, and breaking-change notes.
- Final report after execution: commands run, conflicts resolved, migrations run, verification results, and manual action list.
- Escalation report when blocked: exact path, tier, failure code, and why the skill cannot safely proceed.

# References

- `references/tier-classification.md`
- `references/conflict-resolution.md`
- `references/escalation-rules.md`
- `references/changelog-conventions.md`

# Allowed Tools

- `Bash`
- `Read`
- `Edit`
- `Grep`
- `Glob`
