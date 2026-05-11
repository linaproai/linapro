-- 014: Scheduled Job Management
-- 014：定时任务管理
-- Includes scheduled job groups, scheduled jobs, execution logs, runtime parameters, menu permissions, and dictionary seeds.
-- 包含：定时任务分组、定时任务、执行日志、运行时参数、菜单权限与字典种子

-- ============================================================
-- Purpose: Stores tenant-scoped scheduled job groups used to organize jobs and define the default group.
-- 用途：存储租户级定时任务分组，用于组织任务并定义默认分组。
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_job_group (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT NOT NULL DEFAULT 0,
    "code"       VARCHAR(64) NOT NULL DEFAULT '',
    "name"       VARCHAR(128) NOT NULL DEFAULT '',
    "remark"     VARCHAR(512) NOT NULL DEFAULT '',
    "sort_order" INT NOT NULL DEFAULT 0,
    "is_default" SMALLINT NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP,
    "updated_at" TIMESTAMP,
    "deleted_at" TIMESTAMP
);

COMMENT ON TABLE sys_job_group IS 'Scheduled job group table';
COMMENT ON COLUMN sys_job_group."id" IS 'Job group ID';
COMMENT ON COLUMN sys_job_group."tenant_id" IS 'Owning tenant ID, 0 means PLATFORM';
COMMENT ON COLUMN sys_job_group."code" IS 'Group code';
COMMENT ON COLUMN sys_job_group."name" IS 'Group name';
COMMENT ON COLUMN sys_job_group."remark" IS 'Remark';
COMMENT ON COLUMN sys_job_group."sort_order" IS 'Display order';
COMMENT ON COLUMN sys_job_group."is_default" IS 'Default group flag: 1=yes, 0=no';
COMMENT ON COLUMN sys_job_group."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_job_group."updated_at" IS 'Update time';
COMMENT ON COLUMN sys_job_group."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_job_group_tenant_code ON sys_job_group ("tenant_id", "code");
CREATE INDEX IF NOT EXISTS idx_sys_job_group_is_default ON sys_job_group ("is_default");

-- ============================================================
-- Purpose: Stores scheduled job definitions, execution policy, cron settings, handler or shell payload, and lifecycle status.
-- 用途：存储定时任务定义、执行策略、Cron 配置、Handler 或 Shell 载荷与生命周期状态。
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_job (
    "id"                     BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"              INT NOT NULL DEFAULT 0,
    "group_id"               BIGINT NOT NULL DEFAULT 0,
    "name"                   VARCHAR(128) NOT NULL DEFAULT '',
    "description"            VARCHAR(512) NOT NULL DEFAULT '',
    "task_type"              VARCHAR(32) NOT NULL DEFAULT 'handler',
    "handler_ref"            VARCHAR(255) NOT NULL DEFAULT '',
    "params"                 TEXT,
    "timeout_seconds"        INT NOT NULL DEFAULT 300,
    "shell_cmd"              TEXT,
    "work_dir"               VARCHAR(512) NOT NULL DEFAULT '',
    "env"                    TEXT,
    "cron_expr"              VARCHAR(128) NOT NULL DEFAULT '',
    "timezone"               VARCHAR(64) NOT NULL DEFAULT '',
    "scope"                  VARCHAR(32) NOT NULL DEFAULT 'master_only',
    "concurrency"            VARCHAR(32) NOT NULL DEFAULT 'singleton',
    "max_concurrency"        INT NOT NULL DEFAULT 1,
    "max_executions"         INT NOT NULL DEFAULT 0,
    "executed_count"         BIGINT NOT NULL DEFAULT 0,
    "stop_reason"            VARCHAR(64) NOT NULL DEFAULT '',
    "log_retention_override" TEXT,
    "status"                 VARCHAR(32) NOT NULL DEFAULT 'disabled',
    "is_builtin"             SMALLINT NOT NULL DEFAULT 0,
    "seed_version"           INT NOT NULL DEFAULT 0,
    "created_by"             BIGINT NOT NULL DEFAULT 0,
    "updated_by"             BIGINT NOT NULL DEFAULT 0,
    "created_at"             TIMESTAMP,
    "updated_at"             TIMESTAMP,
    "deleted_at"             TIMESTAMP
);

COMMENT ON TABLE sys_job IS 'Scheduled job table';
COMMENT ON COLUMN sys_job."id" IS 'Job ID';
COMMENT ON COLUMN sys_job."tenant_id" IS 'Owning tenant ID, 0 means PLATFORM';
COMMENT ON COLUMN sys_job."group_id" IS 'Owning group ID';
COMMENT ON COLUMN sys_job."name" IS 'Job name';
COMMENT ON COLUMN sys_job."description" IS 'Job description';
COMMENT ON COLUMN sys_job."task_type" IS 'Job type: handler/shell';
COMMENT ON COLUMN sys_job."handler_ref" IS 'Unique handler reference';
COMMENT ON COLUMN sys_job."params" IS 'Handler parameters JSON';
COMMENT ON COLUMN sys_job."timeout_seconds" IS 'Execution timeout in seconds';
COMMENT ON COLUMN sys_job."shell_cmd" IS 'Shell script content';
COMMENT ON COLUMN sys_job."work_dir" IS 'Working directory';
COMMENT ON COLUMN sys_job."env" IS 'Environment variables JSON';
COMMENT ON COLUMN sys_job."cron_expr" IS 'Cron expression';
COMMENT ON COLUMN sys_job."timezone" IS 'Timezone identifier';
COMMENT ON COLUMN sys_job."scope" IS 'Scheduling scope: master_only/all_node';
COMMENT ON COLUMN sys_job."concurrency" IS 'Concurrency policy: singleton/parallel';
COMMENT ON COLUMN sys_job."max_concurrency" IS 'Maximum concurrency';
COMMENT ON COLUMN sys_job."max_executions" IS 'Maximum executions, 0 means unlimited';
COMMENT ON COLUMN sys_job."executed_count" IS 'Executed count';
COMMENT ON COLUMN sys_job."stop_reason" IS 'Stop reason';
COMMENT ON COLUMN sys_job."log_retention_override" IS 'Log retention override JSON';
COMMENT ON COLUMN sys_job."status" IS 'Job status: enabled/disabled/paused_by_plugin';
COMMENT ON COLUMN sys_job."is_builtin" IS 'Built-in job flag: 1=yes, 0=no';
COMMENT ON COLUMN sys_job."seed_version" IS 'Seed version number';
COMMENT ON COLUMN sys_job."created_by" IS 'Creator user ID';
COMMENT ON COLUMN sys_job."updated_by" IS 'Updater user ID';
COMMENT ON COLUMN sys_job."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_job."updated_at" IS 'Update time';
COMMENT ON COLUMN sys_job."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_job_tenant_group_name ON sys_job ("tenant_id", "group_id", "name");
CREATE INDEX IF NOT EXISTS idx_sys_job_tenant_status ON sys_job ("tenant_id", "status");
CREATE INDEX IF NOT EXISTS idx_sys_job_group_id ON sys_job ("group_id");
CREATE INDEX IF NOT EXISTS idx_sys_job_task_type ON sys_job ("task_type");
CREATE INDEX IF NOT EXISTS idx_sys_job_handler_ref ON sys_job ("handler_ref");
CREATE INDEX IF NOT EXISTS idx_sys_job_is_builtin ON sys_job ("is_builtin");

-- ============================================================
-- Purpose: Stores scheduled job execution history, snapshots, timing, node information, result payloads, and failures.
-- 用途：存储定时任务执行历史、快照、耗时、节点信息、结果载荷与失败原因。
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_job_log (
    "id"              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"       INT NOT NULL DEFAULT 0,
    "job_id"          BIGINT NOT NULL DEFAULT 0,
    "job_snapshot"    TEXT,
    "node_id"         VARCHAR(128) NOT NULL DEFAULT '',
    "trigger"       VARCHAR(32) NOT NULL DEFAULT 'cron',
    "params_snapshot" TEXT,
    "start_at"        TIMESTAMP,
    "end_at"          TIMESTAMP,
    "duration_ms"     BIGINT NOT NULL DEFAULT 0,
    "status"          VARCHAR(64) NOT NULL DEFAULT 'running',
    "err_msg"         VARCHAR(1000) NOT NULL DEFAULT '',
    "result_json"     TEXT,
    "created_at"      TIMESTAMP
);

COMMENT ON TABLE sys_job_log IS 'Scheduled job execution log table';
COMMENT ON COLUMN sys_job_log."id" IS 'Log ID';
COMMENT ON COLUMN sys_job_log."tenant_id" IS 'Owning tenant ID, 0 means PLATFORM';
COMMENT ON COLUMN sys_job_log."job_id" IS 'Owning job ID';
COMMENT ON COLUMN sys_job_log."job_snapshot" IS 'Job snapshot JSON at execution time';
COMMENT ON COLUMN sys_job_log."node_id" IS 'Execution node identifier';
COMMENT ON COLUMN sys_job_log."trigger" IS 'Trigger type: cron/manual';
COMMENT ON COLUMN sys_job_log."params_snapshot" IS 'Parameter snapshot JSON at execution time';
COMMENT ON COLUMN sys_job_log."start_at" IS 'Start time';
COMMENT ON COLUMN sys_job_log."end_at" IS 'End time';
COMMENT ON COLUMN sys_job_log."duration_ms" IS 'Execution duration in milliseconds';
COMMENT ON COLUMN sys_job_log."status" IS 'Execution status';
COMMENT ON COLUMN sys_job_log."err_msg" IS 'Error summary';
COMMENT ON COLUMN sys_job_log."result_json" IS 'Execution result JSON';
COMMENT ON COLUMN sys_job_log."created_at" IS 'Creation time';

CREATE INDEX IF NOT EXISTS idx_sys_job_log_job_id_start_at ON sys_job_log ("job_id", "start_at");
CREATE INDEX IF NOT EXISTS idx_sys_job_log_tenant_job_start ON sys_job_log ("tenant_id", "job_id", "start_at");
CREATE INDEX IF NOT EXISTS idx_sys_job_log_status ON sys_job_log ("status");
CREATE INDEX IF NOT EXISTS idx_sys_job_log_start_at ON sys_job_log ("start_at");

-- ============================================================
-- Default group and runtime parameters
-- 默认分组与运行时参数
-- ============================================================
INSERT INTO sys_job_group ("tenant_id", "code", "name", "remark", "sort_order", "is_default", "created_at", "updated_at")
VALUES (0, 'default', 'Default Group', 'The system default job group. Jobs are moved here when other groups are deleted.', 0, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_config ("name", "key", "value", "is_builtin", "remark", "created_at", "updated_at")
VALUES (
    '定时任务-Shell 模式全局开关',
    'cron.shell.enabled',
    'true',
    1,
    '控制 Shell 类型任务是否允许创建、修改、触发与终止，可选值：true、false。',
    NOW(),
    NOW()
)
ON CONFLICT DO NOTHING;

INSERT INTO sys_config ("name", "key", "value", "is_builtin", "remark", "created_at", "updated_at")
VALUES (
    '定时任务-执行日志保留策略',
    'cron.log.retention',
    '{"mode":"days","value":30}',
    1,
    '控制定时任务执行日志默认保留策略，使用 JSON：{"mode":"days|count|none","value":N}。',
    NOW(),
    NOW()
)
ON CONFLICT DO NOTHING;

-- ============================================================
-- Dictionary types and dictionary data
-- 字典类型与字典数据
-- ============================================================
INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('定时任务状态', 'cron_job_status', 1, 1, '定时任务状态列表', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('定时任务类型', 'cron_job_task_type', 1, 1, '定时任务类型列表', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('定时任务调度范围', 'cron_job_scope', 1, 1, '定时任务调度范围列表', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('定时任务并发策略', 'cron_job_concurrency', 1, 1, '定时任务并发策略列表', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('定时任务触发方式', 'cron_job_trigger', 1, 1, '定时任务触发方式列表', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('定时任务日志状态', 'cron_job_log_status', 1, 1, '定时任务执行日志状态列表', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('定时任务日志保留模式', 'cron_log_retention_mode', 1, 1, '定时任务日志保留模式列表', NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_status', '启用', 'enabled', 1, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_status', '停用', 'disabled', 2, 'default', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_status', '不可用', 'paused_by_plugin', 3, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_task_type', 'Handler 任务', 'handler', 1, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_task_type', 'Shell 任务', 'shell', 2, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_scope', '仅主节点执行', 'master_only', 1, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_scope', '所有节点执行', 'all_node', 2, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_concurrency', '单例执行', 'singleton', 1, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_concurrency', '允许并行执行', 'parallel', 2, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_trigger', 'Cron 调度', 'cron', 1, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_trigger', '手动触发', 'manual', 2, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_log_status', '运行中', 'running', 1, 'processing', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_log_status', '成功', 'success', 2, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_log_status', '失败', 'failed', 3, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_log_status', '已取消', 'cancelled', 4, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_log_status', '超时', 'timeout', 5, 'orange', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_log_status', '非主节点跳过', 'skipped_not_primary', 6, 'default', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_log_status', '单例冲突跳过', 'skipped_singleton', 7, 'cyan', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_job_log_status', '并发上限跳过', 'skipped_max_concurrency', 8, 'purple', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_log_retention_mode', '按天保留', 'days', 1, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_log_retention_mode', '按条数保留', 'count', 2, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('cron_log_retention_mode', '不清理', 'none', 3, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ============================================================
-- Menus and button permissions
-- 菜单与按钮权限
-- ============================================================
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'scheduler') AS parent), 'system:job:list', '任务管理', '/system/job', 'system/job/index', 'system:job:list', 'lucide:clock-3', 'M', 1, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'scheduler') AS parent), 'system:jobgroup:list', '分组管理', '/system/job-group', 'system/job-group/index', 'system:jobgroup:list', 'lucide:blocks', 'M', 2, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'scheduler') AS parent), 'system:joblog:list', '执行日志', '/system/job-log', 'system/job-log/index', 'system:joblog:list', 'lucide:scroll-text', 'M', 3, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:job:list') AS parent), 'system:job:add', '任务新增', '', '', 'system:job:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:job:list') AS parent), 'system:job:edit', '任务修改', '', '', 'system:job:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:job:list') AS parent), 'system:job:remove', '任务删除', '', '', 'system:job:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:job:list') AS parent), 'system:job:status', '任务启停', '', '', 'system:job:status', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:job:list') AS parent), 'system:job:trigger', '立即执行', '', '', 'system:job:trigger', '', 'B', 6, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:job:list') AS parent), 'system:job:reset', '重置计数', '', '', 'system:job:reset', '', 'B', 7, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:job:list') AS parent), 'system:job:shell', 'Shell 任务权限', '', '', 'system:job:shell', '', 'B', 8, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:jobgroup:list') AS parent), 'system:jobgroup:add', '分组新增', '', '', 'system:jobgroup:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:jobgroup:list') AS parent), 'system:jobgroup:edit', '分组修改', '', '', 'system:jobgroup:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:jobgroup:list') AS parent), 'system:jobgroup:remove', '分组删除', '', '', 'system:jobgroup:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:joblog:list') AS parent), 'system:joblog:remove', '日志清空', '', '', 'system:joblog:remove', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:joblog:list') AS parent), 'system:joblog:cancel', '日志终止', '', '', 'system:joblog:cancel', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
