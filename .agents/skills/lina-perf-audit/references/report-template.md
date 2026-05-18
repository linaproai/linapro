# Report Templates

Use these templates for run-level reports under `temp/lina-perf-audit/<run-id>/`.

## Module Audit Template

```markdown
# Audit: <module-or-shard>

## Scope

- Run ID: `<run-id>`
- Module: `<module>`
- Shard: `<shard>`
- Endpoint count: `<count>`
- Log path: `<run-dir>/server.log`

## Result Summary

- HIGH: `<count>`
- MEDIUM: `<count>`
- LOW: `<count>`
- PASS: `<count>`
- SKIPPED: `<count>`

## Findings

### <severity> - <method> <path> - <short title>

- Module: `<module>`
- Endpoint: `<method> <path>`
- Severity: `<HIGH|MEDIUM|LOW>`
- Anti-pattern signature: `<signature>`
- Trace ID: `<trace-id or fallback marker>`
- Status / elapsed: `<status-code> / <milliseconds>ms`
- SQL count: `<count>`
- Write SQL count: `<count, 0 when read-only endpoint has no write SQL>`
- Source: `<relative-file>:<line>`

Evidence:

```sql
<short SQL excerpts only>
```

Analysis:

<Describe the growth pattern, source cause, and impact.>

Recommendation:

1. <Concrete remediation step.>
2. <Concrete validation step.>

## Passed Endpoints

| Endpoint | Trace ID | SQL count | Elapsed | Notes |
|---|---|---:|---:|---|
| `<method> <path>` | `<trace-id>` | `<count>` | `<ms>ms` | `<brief note>` |
| `<method> <path>` | `<trace-id>` | `<count>` | `<ms>ms` | `expected operational side effect: writes only sys_online_session/plugin_monitor_operlog` |

## Skipped Endpoints

| Endpoint | Reason | Follow-up |
|---|---|---|
| `<method> <path>` | `<reason>` | `<manual action>` |

## Destructive Endpoint Handling

| Endpoint | Handling | Cleanup result |
|---|---|---|
| `<method> <path>` | `<autonomous fixture or skipped>` | `<result>` |
```

## Summary Template

```markdown
# LinaPro Performance Audit Summary

## Run Metadata

- Run ID: `<run-id>`
- Git commit: `<commit>`
- Started at: `<timestamp>`
- Finished at: `<timestamp>`
- Run directory: `temp/lina-perf-audit/<run-id>/`
- Stress fixture: `<enabled|disabled>`
- Read side-effect violations: `<count>`
- Sub agents: `<count>`
- Logger path: `<value>`
- Logger file: `<value>`
- Restore result: `<success|failure>`
- Delivery code modified: `no`

## Scope

- Host API catalog: `<count>` endpoints
- Built-in plugin API catalog: `<count>` endpoints
- Skipped plugins: `<plugin: reason>`

## HIGH

| Issue | Endpoint | Module | Card |
|---|---|---|---|
| `<title>` | `<method> <path>` | `<module>` | `perf-issues/<file>.md` |

## Read Request Side-Effect Violations

| Endpoint | Module | Write SQL count | Card |
|---|---|---:|---|
| `<method> <path>` | `<module>` | `<count>` | `perf-issues/<file>.md` |

## MEDIUM

| Issue | Endpoint | Module | Card |
|---|---|---|---|
| `<title>` | `<method> <path>` | `<module>` | `perf-issues/<file>.md` |

## LOW

| Issue | Endpoint | Module | Card |
|---|---|---|---|
| `<title>` | `<method> <path>` | `<module>` | `perf-issues/<file>.md` |

## Manual Follow-Up Required

| Endpoint | Reason | Audit file |
|---|---|---|
| `<method> <path>` | `<reason>` | `temp/lina-perf-audit/<run-id>/audits/<file>.md` |

## Audit Files

| Module or shard | Path | Status |
|---|---|---|
| `<module>` | `temp/lina-perf-audit/<run-id>/audits/<file>.md` | `<status>` |

## Notes

This run did not modify delivery code, API DTOs, SQL delivery files, frontend code, or OpenSpec artifacts.
```

## meta.json Template

```json
{
  "run_id": "YYYYMMDD-HHMMSS",
  "started_at": "2026-05-01T10:00:00+08:00",
  "finished_at": "2026-05-01T10:30:00+08:00",
  "git_commit": "<sha>",
  "run_dir": "temp/lina-perf-audit/<run-id>",
  "stress_fixture_enabled": true,
  "logger": {
    "original_path": "<value or empty>",
    "original_file": "<value or empty>",
    "audit_path": "temp/lina-perf-audit/<run-id>",
    "audit_file": "server.log",
    "restore_result": "success"
  },
  "sub_agents": {
    "count": 0,
    "completed": 0,
    "failed": 0
  },
  "skipped_plugins": [
    {
      "plugin": "<plugin-id>",
      "reason": "no backend API"
    }
  ]
}
```
