# Sub-Agent Prompt Template

Use this template when Stage 1 launches a sub agent. Keep the final prompt under 5KB. Include only the assigned endpoints, the minimal fixture subset, and short source pointers. If a module is too large, split it by endpoint group or small module shard and launch multiple sub agents.

## Template

```text
You are auditing LinaPro backend API performance and read-request side effects for one isolated shard.

Scope:
- module: {{module}}
- shard: {{shard_name_or_endpoint_group}}
- run_dir: {{run_dir}}
- log_path: {{log_path}}
- token: {{token}}
- fixtures: {{fixtures_json_subset}}
- endpoints[]: {{endpoint_json_subset}}

Hard limits:
- Do not run make stop, make init, make mock, setup-audit-env.sh, prepare-builtin-plugins.sh, or stress-fixture.sh.
- Do not modify source code, API DTOs, SQL files, frontend files, OpenSpec files, or scripts.
- Write only your assigned audit file under {{run_dir}}/audits/.
- Keep your final audit markdown under 5KB unless the assigned shard genuinely needs more evidence.

Process:
1. Audit endpoints serially in the order provided.
2. For each endpoint, build request data from fixtures. If required fixture data is missing, record SKIPPED with the missing key.
3. For destructive endpoints, use the autonomous fixture flow:
   - If a same-module create endpoint exists, create a dedicated audit resource.
   - Call the destructive endpoint only on that created resource.
   - Attempt cleanup even if the destructive call fails.
   - Mark "autonomous fixture completed, shared data not polluted".
   - If no same-module create endpoint exists, mark "SKIPPED: no matching create endpoint, manual follow-up required".
4. Call the endpoint with Authorization: Bearer {{token}} and capture response headers, body size, status code, and elapsed time.
5. Read the GoFrame Trace-ID response header. Use it to grep {{log_path}} for SQL and request log lines.
6. If Trace-ID is unavailable, search by call timestamp window +/- 2s and request URL. Mark "trace ID unavailable, evidence quality reduced".
7. Count SQL statements for the request. Capture short SQL excerpts only; do not paste long log sections.
8. For every GET/read/query endpoint, check the traced SQL for write statements whose first significant token is INSERT, UPDATE, DELETE, REPLACE, TRUNCATE, ALTER, DROP, or CREATE. Count only SQL for the audited endpoint trace, not setup/login/stress-fixture/autonomous fixture calls.
   - If the trace also contains read SQL and every write statement only touches `sys_online_session` and/or `plugin_monitor_operlog`, treat the session heartbeat or operation-log write as an expected operational side effect. Record the endpoint as PASS with a note such as `expected operational side effect: sys_online_session/plugin_monitor_operlog`, and do not create a HIGH finding.
   - If any write statement touches another table, or if the trace has only writes and no read SQL, record a HIGH finding with anti_pattern_signature prefix `read-write-side-effect`.
9. Read the relevant controller/service source only as needed to confirm patterns such as loop dao calls, missing pagination, repeated reads, missing indexes, blocking calls in loops, or read endpoints that mutate persistent state.
10. Classify findings as HIGH, MEDIUM, or LOW using references/severity-rubric.md.
11. Write {{run_dir}}/audits/{{module_or_shard_file}}.md using references/report-template.md.

Required finding fields:
- module
- endpoint method and path
- severity
- anti_pattern_signature
- trace_id or fallback marker
- status code and elapsed time
- SQL count and key SQL excerpts
- write SQL count, write target table names, and write SQL excerpts when the endpoint is expected to be read-only and the write is reportable
- PASS note for expected operational side effects that only touch `sys_online_session` and/or `plugin_monitor_operlog`
- source references as repository-relative file:line
- concrete remediation steps

If no issue is found for an endpoint, record a concise PASS entry with trace ID, SQL count, and elapsed time.
```

## Sharding Guidance

- Prefer module shards when the module prompt stays under 5KB.
- Use endpoint shards when a module has many endpoints, mixed resources, or large fixture requirements.
- Keep destructive endpoints in the same shard as their matching create endpoint when practical.
- Do not split one destructive endpoint away from the fixture create endpoint if that would force cross-shard coordination.
- Name shard files predictably, for example `user.md`, `plugin-job-list.md`, or `dict-data-write.md`.
