-- Mock data: scheduled job groups, jobs, and execution logs.
-- 模拟数据：定时任务分组、任务与执行日志。
-- Static scheduled job execution logs use exact existence checks so mock loading is idempotent.

INSERT INTO sys_job_group ("code", "name", "remark", "sort_order", "is_default", "created_at", "updated_at")
VALUES ('mock-maintenance', 'Mock Maintenance', 'Mock job group for scheduler management demos', 10, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_job (
    "group_id",
    "name",
    "description",
    "task_type",
    "handler_ref",
    "params",
    "timeout_seconds",
    "cron_expr",
    "timezone",
    "scope",
    "concurrency",
    "max_concurrency",
    "max_executions",
    "executed_count",
    "status",
    "is_builtin",
    "seed_version",
    "created_by",
    "updated_by",
    "created_at",
    "updated_at"
)
SELECT
    g."id",
    'Demo cache cleanup',
    'Disabled mock handler job used by the scheduler list and detail pages.',
    'handler',
    'system:mock:cache-cleanup',
    '{"scope":"demo","dryRun":true}',
    120,
    '0 */30 * * * *',
    'Asia/Shanghai',
    'master_only',
    'singleton',
    1,
    0,
    12,
    'disabled',
    0,
    1,
    admin."id",
    admin."id",
    '2026-04-20 08:00:00',
    '2026-04-20 08:00:00'
FROM sys_job_group g
JOIN sys_user admin ON admin."username" = 'admin'
WHERE g."code" = 'mock-maintenance'
ON CONFLICT DO NOTHING;

INSERT INTO sys_job (
    "group_id",
    "name",
    "description",
    "task_type",
    "handler_ref",
    "params",
    "timeout_seconds",
    "shell_cmd",
    "cron_expr",
    "timezone",
    "scope",
    "concurrency",
    "max_concurrency",
    "max_executions",
    "executed_count",
    "status",
    "is_builtin",
    "seed_version",
    "created_by",
    "updated_by",
    "created_at",
    "updated_at"
)
SELECT
    g."id",
    'Demo disk inspection',
    'Disabled mock shell job used to demonstrate shell task configuration.',
    'shell',
    '',
    '{}',
    60,
    'df -h',
    '0 0 2 * * *',
    'Asia/Shanghai',
    'master_only',
    'singleton',
    1,
    0,
    3,
    'disabled',
    0,
    1,
    admin."id",
    admin."id",
    '2026-04-20 08:10:00',
    '2026-04-20 08:10:00'
FROM sys_job_group g
JOIN sys_user admin ON admin."username" = 'admin'
WHERE g."code" = 'mock-maintenance'
ON CONFLICT DO NOTHING;

INSERT INTO sys_job_log (
    "job_id",
    "job_snapshot",
    "node_id",
    "trigger",
    "params_snapshot",
    "start_at",
    "end_at",
    "duration_ms",
    "status",
    "err_msg",
    "result_json",
    "created_at"
)
SELECT
    j."id",
    '{"name":"Demo cache cleanup","taskType":"handler","mock":true}',
    'linapro-dev-01',
    'manual',
    '{"scope":"demo","dryRun":true}',
    '2026-04-20 09:00:00',
    '2026-04-20 09:00:03',
    3200,
    'success',
    '',
    '{"removed":128,"unit":"entries"}',
    '2026-04-20 09:00:03'
FROM sys_job j
JOIN sys_job_group g ON g."id" = j."group_id"
WHERE g."code" = 'mock-maintenance'
  AND j."name" = 'Demo cache cleanup'
  AND NOT EXISTS (
      SELECT 1
      FROM sys_job_log existing
      WHERE existing."job_id" = j."id"
        AND existing."node_id" = 'linapro-dev-01'
        AND existing."trigger" = 'manual'
        AND existing."start_at" = '2026-04-20 09:00:00'
  );

INSERT INTO sys_job_log (
    "job_id",
    "job_snapshot",
    "node_id",
    "trigger",
    "params_snapshot",
    "start_at",
    "end_at",
    "duration_ms",
    "status",
    "err_msg",
    "result_json",
    "created_at"
)
SELECT
    j."id",
    '{"name":"Demo disk inspection","taskType":"shell","mock":true}',
    'linapro-dev-02',
    'cron',
    '{}',
    '2026-04-21 02:00:00',
    '2026-04-21 02:00:01',
    1240,
    'failed',
    'Shell jobs are disabled in the mock environment',
    '{"exitCode":1}',
    '2026-04-21 02:00:01'
FROM sys_job j
JOIN sys_job_group g ON g."id" = j."group_id"
WHERE g."code" = 'mock-maintenance'
  AND j."name" = 'Demo disk inspection'
  AND NOT EXISTS (
      SELECT 1
      FROM sys_job_log existing
      WHERE existing."job_id" = j."id"
        AND existing."node_id" = 'linapro-dev-02'
        AND existing."trigger" = 'cron'
        AND existing."start_at" = '2026-04-21 02:00:00'
  );
