# Severity Rubric

Classify evidence-backed performance risks and read-request side effects. Every finding needs runtime evidence, static source evidence, or an explicit note that it is static-only.

## Required Evidence

Each finding must include:

- module
- endpoint method and path
- severity
- anti-pattern signature
- trace ID or fallback marker
- SQL count and representative SQL excerpts
- write SQL count and representative write SQL excerpts when the endpoint is expected to be read-only
- elapsed time when available
- source reference as repository-relative `file:line`
- concrete remediation suggestion

## HIGH

Use `HIGH` when immediate remediation is justified.

| Pattern | Rule | Anti-pattern signature prefix |
|---|---|---|
| N+1 list/detail | SQL count grows approximately with returned row count plus a small constant, and source shows looped DAO or service calls | `n-plus-one` |
| Missing index | Query filters or joins on a field that is not indexed and the table can grow beyond small seed data | `missing-index` |
| Slow non-batch endpoint | Non-export, non-batch endpoint elapsed time is over 1s in audit data | `slow-endpoint` |
| Blocking loop work | Loop performs remote calls, file I/O, cross-transaction work, or repeated expensive computation | `loop-blocking` |
| Read request write side effect | GET/list/query/detail/count/options/tree/current endpoint trace executes write SQL such as INSERT, UPDATE, DELETE, REPLACE, TRUNCATE, ALTER, DROP, or CREATE against tables other than the expected operational side-effect tables | `read-write-side-effect` |

Examples:

- `n-plus-one:list:user:dao.Dept.Get`
- `missing-index:sys_login_log:login_time`
- `slow-endpoint:GET:/user`
- `loop-blocking:notice:send-file-io`
- `read-write-side-effect:user-message:update-read-status`

## MEDIUM

Use `MEDIUM` when the pattern is likely to become harmful as data grows.

| Pattern | Rule | Anti-pattern signature prefix |
|---|---|---|
| Small-sample N+1 | N+1 pattern exists but current observed row count is below 20 | `n-plus-one-small` |
| Missing pagination | List endpoint can return an unbounded full table or ignores page/pageSize | `unbounded-list` |
| Repeated read | Same request repeatedly reads identical reference data without request-level caching or batching | `duplicate-read` |
| Mergeable SELECTs | Multiple SELECT calls can be collapsed into JOIN, WHERE IN, or preloaded maps | `mergeable-selects` |

Examples:

- `n-plus-one-small:dict-data:dao.DictType.Get`
- `unbounded-list:sys_menu`
- `duplicate-read:config:i18n.locales`
- `mergeable-selects:user:roles-posts`

## LOW

Use `LOW` for observable but non-urgent inefficiency.

| Pattern | Rule | Anti-pattern signature prefix |
|---|---|---|
| Slightly high SQL count | SQL count is higher than ideal but queries are indexed and each is under 5ms | `extra-sql` |
| Application filter | Filtering that can be pushed to SQL is done in application code | `app-filter` |
| Static-only risk | Source suggests a risk but runtime evidence did not reproduce it | `static-risk` |
| Static-only read side effect | Source suggests a GET/read path may write but the sampled trace did not execute that branch | `static-read-write-risk` |

Examples:

- `extra-sql:dict-options:indexed`
- `app-filter:notice:status`
- `static-risk:job-log:loop-service-call`
- `static-read-write-risk:user-message:conditional-update`

## Tie-Breaking Rules

- Prefer the highest severity supported by evidence.
- Do not mark an issue `HIGH` for SQL count alone; connect it to returned row count, missing indexes, elapsed time, or source evidence.
- Export, import, and batch endpoints may legitimately exceed 1s. Classify them by growth pattern and blocking behavior instead of elapsed time alone.
- Do not count Stage 0 setup writes, login writes, stress fixture writes, or autonomous fixture create/delete writes as read-request side effects; count only statements inside the audited endpoint trace.
- If a read endpoint trace contains read SQL and its write SQL only touches `sys_online_session` and/or `plugin_monitor_operlog`, classify it as PASS with an expected operational side-effect note. These writes are normal session heartbeat and operation-log behavior, not performance findings.
- A non-operational write SQL statement in a GET/read/query endpoint is `HIGH` even if it is fast, because it violates read semantics and can corrupt cache, idempotency, retries, and crawler behavior.
- Trace-ID fallback lowers evidence quality but does not automatically lower severity.
- If source and runtime evidence disagree, report the disagreement and use the lower severity unless there is clear user-impact evidence.
