# Conflict Resolution

This document defines how the `lina-upgrade` skill handles merge conflicts after `git merge --no-commit`.

## Tier 1

Tier 1 paths are stable contracts. A conflict means the user and upstream both changed a public surface that downstream code may rely on.

Required behavior:

1. Stop automatic resolution immediately.
2. Report the conflicted path, current baseline, target version, and any changelog or OpenSpec entry that mentions the path.
3. Ask the user for a manual decision.

Do not run `git checkout --theirs` or invent compatibility wrappers for Tier 1 conflicts.

## Tier 2

Tier 2 paths may be validly edited by the user. The AI can perform a semantic three-way merge only when confidence is high.

High confidence examples:

- Import ordering changed around otherwise independent additions.
- Two adjacent methods were added without changing each other's behavior.
- A small config default changed and the user added unrelated comments or fields.

Low confidence examples:

- The conflict spans multiple files and shared business rules.
- Method signatures changed and callers also changed.
- The conflict changes authorization, data mutation, SQL migration, or plugin lifecycle behavior.

When confidence is low, leave the conflict markers untouched and escalate.

## Tier 3

Tier 3 paths are generated artifacts.

Resolution pattern:

```bash
git checkout --theirs <path>
make dao
make ctrl
```

Use the relevant plugin-local generator when the generated path belongs to a source plugin. Never hand-edit generated `DAO`, `DO`, `Entity`, or controller skeleton files.

## Common Cases

| Case | Preferred handling |
| --- | --- |
| Import reorder only | Accept the formatter result after semantic merge. |
| Adjacent method additions | Keep both methods if names and responsibilities do not overlap. |
| Signature changed upstream | Update local callers only when the intent is unambiguous; otherwise escalate. |
| SQL drops a populated column | Escalate because data loss may occur. |
