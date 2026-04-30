# Changelog Conventions

The `lina-upgrade` plan stage summarizes changes from `CHANGELOG.md` and archived OpenSpec changes.

## `CHANGELOG.md`

When `CHANGELOG.md` exists:

1. Find headings that mention the baseline version and target version.
2. Extract entries between those headings.
3. Preserve version headings and bullet summaries.
4. Highlight entries containing `BREAKING`, `**BREAKING**`, `Tier 1`, or public contract package names.

If the file does not exist or the range cannot be detected, the plan should say so explicitly instead of fabricating release notes.

## OpenSpec Archive

Read archived proposals from:

```text
openspec/changes/archive/<archive-date>-<change-id>/proposal.md
```

For each archived proposal in the upgrade range:

1. Extract the `Why` section.
2. Extract the `What Changes` section.
3. Highlight bullets containing `**BREAKING**`.
4. If a breaking entry mentions Tier 1 paths, flag it as mandatory user review in the plan.

When tag dates are unavailable, list archive proposals as contextual changes and state that date-range filtering could not be performed.

## Output Shape

The upgrade plan should include:

- Baseline version and target version.
- Commit count from baseline to target.
- Changelog excerpts.
- Archived OpenSpec proposal excerpts.
- Changed files grouped by Tier 1 / Tier 2 / Tier 3 / unknown.
- New SQL files.
- Breaking-change and Tier 1 review warnings.
