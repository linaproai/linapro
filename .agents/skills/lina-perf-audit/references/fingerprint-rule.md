# Fingerprint Rule

Use fingerprints to avoid duplicate issue cards across audit runs.

## Formula

```text
fingerprint_input = normalized_module + ":" + normalized_method + ":" + normalized_path + ":" + normalized_severity + ":" + normalized_anti_pattern_signature
fingerprint = sha256(fingerprint_input)
```

The formula must match:

```text
sha256(module + ":" + method + ":" + path + ":" + severity + ":" + anti_pattern_signature)
```

after each component is normalized as described below.

## Normalization

- `module`: lower-case, trimmed, use catalog module ID, replace whitespace with `-`.
- `method`: upper-case HTTP method.
- `path`: strip scheme, host, query string, and trailing slash except for `/`; preserve path parameters as declared by the catalog, for example `/user/{id}`.
- `severity`: upper-case `HIGH`, `MEDIUM`, or `LOW`.
- `anti_pattern_signature`: lower-case canonical signature from `severity-rubric.md`; remove run IDs, trace IDs, timestamps, elapsed times, SQL literal values, and row counts.

Recommended anti-pattern signature shape:

```text
<pattern-prefix>:<resource-or-table>:<stable-cause>
```

Examples:

```text
n-plus-one:user:dao.dept.get
missing-index:sys_login_log:login_time
unbounded-list:sys_menu:no-page-size-limit
duplicate-read:config:i18n.locales
read-write-side-effect:user-message:update-read-status
```

Do not include volatile line numbers in the signature unless there is no more stable source anchor. Put `file:line` references in report evidence instead.

## Lookup

Before creating a card:

1. Search `perf-issues/*.md` for a frontmatter `fingerprint` equal to the calculated value.
2. If one card matches, update that card.
3. If more than one card matches, keep the oldest `first_seen_run` card, update it, and list the duplicate paths in `SUMMARY.md` as cleanup follow-up.
4. If no card matches, create a new card using `issue-card-template.md`.

## New Card Rules

For a new finding:

- File name: `perf-issues/<severity>-<module>-<slug>.md`
- `id`: `<severity>-<module>-<slug>`
- `status`: `open`
- `first_seen_run`: current run ID
- `last_seen_run`: current run ID
- `seen_count`: `1`
- `fingerprint`: calculated SHA-256 value

The slug should be short, stable, and derived from the pattern, for example `n-plus-1-list`, `missing-index-status`, or `unbounded-list`.

## Existing Card Update Rules

When the fingerprint matches an existing card:

- Preserve `id`, `first_seen_run`, and `fingerprint`.
- Set `last_seen_run` to the current run ID.
- Increment `seen_count` by 1.
- Append a `历史记录` entry with current run ID, audit file path, SQL count, trace ID or fallback marker, and notable change.
- Keep `status: open` as open.
- Keep `status: in-progress` as in-progress.
- Change `status: fixed` to `open` and append a `历史记录` entry containing `被再次观察到 (回归)`.
- Change `status: obsolete` to `open` if the same endpoint and pattern are observed again, and append a `历史记录` entry containing `被再次观察到 (回归)`.

If severity changes, the fingerprint changes because severity is part of the formula. Create a new card and mention the related previous card in both cards when known.

## INDEX.md Rules

After all cards are created or updated:

1. Read all `perf-issues/*.md` except `INDEX.md`.
2. Include only cards with `status: open` or `status: in-progress`.
3. Sort by severity order `HIGH`, `MEDIUM`, `LOW`, then module, then endpoint.
4. Write repository-relative links to card files.
5. Include `last_seen_run` and `seen_count` in the index table.

## Cross-Run References

- `SUMMARY.md` links to every card created or updated in the current run.
- Each card history entry links back to `temp/lina-perf-audit/<run-id>/audits/<module-or-shard>.md`.
- Cards remain under repository root `perf-issues/` and are not OpenSpec archive artifacts.
