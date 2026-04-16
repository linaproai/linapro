---
name: openspec-review
description: >-
  Code and specification review for OpenSpec workflow. Triggers automatically after /opsx:apply
  task completion, after /opsx:feedback task completion, and before /opsx:archive. Use when
  user requests code review, spec compliance check, or when explicitly invoked via /openspec-review.
compatibility: Requires OpenSpec CLI, GoFrame v2 skill, openspec-e2e skill.
---

# OpenSpec Review

Structured code and specification review for the OpenSpec development workflow.

**Spec Source**: `CLAUDE.md` is the single source of truth for all review criteria.

---

## When This Skill Activates

**Automatic triggers:**
- After completing each task in `/opsx:apply`
- After completing each task in `/opspec-feedback`
- Before executing `/opsx:archive`

**Manual trigger:**
- User explicitly requests: "review this code", "check spec compliance", "/openspec-review"

---

## Review Workflow

### 1. Identify Scope

Determine what needs to be reviewed:

1. **After task completion** — Review files modified/created by the completed task
2. **Before archive** — Review all changes in the current OpenSpec change
3. **Manual invocation** — Ask user to specify scope or use current change

Run `openspec status --change "<name>" --json` to understand the current change state.

### 2. Load Specifications

Read `CLAUDE.md` to load all specifications. This is the single source of truth.

### 3. Backend Code Review

**Trigger**: Changes to files under `apps/lina-core` directory

1. Invoke `goframe-v2` skill for GoFrame framework conventions
2. Check against `CLAUDE.md` backend code specifications

### 4. RESTful API Review

**Trigger**: Any API endpoint changes

Check against `CLAUDE.md` API design specifications.

### 5. Project Specification Review

**Trigger**: Any implementation changes

Check against `CLAUDE.md` architecture design specifications and code development specifications.

### 6. E2E Test Review

**Trigger**: New or modified E2E test files in `hack/tests/e2e/` directory

1. Invoke `openspec-e2e` skill for test conventions
2. Check against `CLAUDE.md` E2E test specifications

### 7. Generate Review Report

```markdown
## OpenSpec Review Report

**Change:** <change-name>
**Scope:** <task-specific / full change>
**Files Reviewed:** <count>

### Backend Code Review
✓ All checks passed / ⚠ N issues found

### RESTful API Review
✓ All endpoints compliant / ⚠ N violations found

### Project Spec Review
✓ Compliant with CLAUDE.md / ⚠ N violations found

### E2E Test Review
✓ Tests follow conventions / ⚠ N issues found

### Summary
- **Critical:** N (must fix before archive)
- **Warnings:** N (recommended to fix)

### Recommended Actions
1. [Specific action with CLAUDE.md reference]
```

---

## Issue Severity

| Level | Behavior |
|-------|----------|
| **Critical** | Block archive, must fix |
| **Warning** | Show but allow proceed |

---

## Integration Points

| Workflow Step | Behavior |
|---------------|----------|
| `/opsx:apply` task done | Review, offer to fix issues before next task |
| `/opspec-feedback` task done | Review, fix before marking complete |
| `/opsx:archive` | Review all changes, block on critical issues |

---

## Guardrails

- **CLAUDE.md is the single source of truth** — All spec references point to it
- Only check categories relevant to changed files
- Don't block on warnings — only critical issues block archive
- Include file paths and line numbers in issue reports
- Offer to fix issues automatically when straightforward