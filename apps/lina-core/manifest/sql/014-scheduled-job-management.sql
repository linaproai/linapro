-- 014: Scheduled Job Management
-- 包含：定时任务分组、定时任务、执行日志、运行时参数、菜单权限与字典种子

-- ============================================================
-- 定时任务分组表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_job_group (
    id          BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '任务分组ID',
    code        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '分组编码',
    name        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '分组名称',
    remark      VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
    sort_order  INT          NOT NULL DEFAULT 0  COMMENT '显示排序',
    is_default  TINYINT      NOT NULL DEFAULT 0  COMMENT '是否默认分组（1=是 0=否）',
    created_at  DATETIME                         COMMENT '创建时间',
    updated_at  DATETIME                         COMMENT '更新时间',
    deleted_at  DATETIME                         COMMENT '删除时间',
    UNIQUE KEY uk_sys_job_group_code (code),
    INDEX idx_sys_job_group_is_default (is_default)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='定时任务分组表';

-- ============================================================
-- 定时任务表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_job (
    id                      BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '任务ID',
    group_id                BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '所属分组ID',
    name                    VARCHAR(128)    NOT NULL DEFAULT '' COMMENT '任务名称',
    description             VARCHAR(512)    NOT NULL DEFAULT '' COMMENT '任务描述',
    task_type               VARCHAR(32)     NOT NULL DEFAULT 'handler' COMMENT '任务类型（handler/shell）',
    handler_ref             VARCHAR(255)    NOT NULL DEFAULT '' COMMENT 'Handler 唯一引用',
    params                  LONGTEXT                            COMMENT 'Handler 参数JSON',
    timeout_seconds         INT             NOT NULL DEFAULT 300 COMMENT '执行超时时间（秒）',
    shell_cmd               LONGTEXT                            COMMENT 'Shell 脚本内容',
    work_dir                VARCHAR(512)    NOT NULL DEFAULT '' COMMENT '工作目录',
    env                     LONGTEXT                            COMMENT '环境变量JSON',
    cron_expr               VARCHAR(128)    NOT NULL DEFAULT '' COMMENT 'Cron 表达式',
    timezone                VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '时区标识',
    scope                   VARCHAR(32)     NOT NULL DEFAULT 'master_only' COMMENT '调度范围（master_only/all_node）',
    concurrency             VARCHAR(32)     NOT NULL DEFAULT 'singleton' COMMENT '并发策略（singleton/parallel）',
    max_concurrency         INT             NOT NULL DEFAULT 1 COMMENT '并发上限',
    max_executions          INT             NOT NULL DEFAULT 0 COMMENT '最大执行次数（0=无限）',
    executed_count          BIGINT          NOT NULL DEFAULT 0 COMMENT '已执行次数',
    stop_reason             VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '停止原因',
    log_retention_override  LONGTEXT                            COMMENT '日志保留策略覆盖JSON',
    status                  VARCHAR(32)     NOT NULL DEFAULT 'disabled' COMMENT '任务状态（enabled/disabled/paused_by_plugin）',
    is_builtin              TINYINT         NOT NULL DEFAULT 0 COMMENT '是否内置任务（1=是 0=否）',
    seed_version            INT             NOT NULL DEFAULT 0 COMMENT '种子版本号',
    created_by              BIGINT          NOT NULL DEFAULT 0 COMMENT '创建者用户ID',
    updated_by              BIGINT          NOT NULL DEFAULT 0 COMMENT '更新者用户ID',
    created_at              DATETIME                            COMMENT '创建时间',
    updated_at              DATETIME                            COMMENT '更新时间',
    deleted_at              DATETIME                            COMMENT '删除时间',
    UNIQUE KEY uk_sys_job_group_name (group_id, name),
    INDEX idx_sys_job_status (status),
    INDEX idx_sys_job_group_id (group_id),
    INDEX idx_sys_job_task_type (task_type),
    INDEX idx_sys_job_handler_ref (handler_ref),
    INDEX idx_sys_job_is_builtin (is_builtin),
    CONSTRAINT fk_sys_job_group_id FOREIGN KEY (group_id) REFERENCES sys_job_group (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='定时任务表';

-- ============================================================
-- 定时任务执行日志表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_job_log (
    id               BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '日志ID',
    job_id           BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '所属任务ID',
    job_snapshot     LONGTEXT                           COMMENT '执行时任务快照JSON',
    node_id          VARCHAR(128)    NOT NULL DEFAULT '' COMMENT '执行节点标识',
    `trigger`        VARCHAR(32)     NOT NULL DEFAULT 'cron' COMMENT '触发方式（cron/manual）',
    params_snapshot  LONGTEXT                           COMMENT '执行时参数快照JSON',
    start_at         DATETIME                           COMMENT '开始时间',
    end_at           DATETIME                           COMMENT '结束时间',
    duration_ms      BIGINT          NOT NULL DEFAULT 0 COMMENT '执行耗时（毫秒）',
    status           VARCHAR(64)     NOT NULL DEFAULT 'running' COMMENT '执行状态',
    err_msg          VARCHAR(1000)   NOT NULL DEFAULT '' COMMENT '错误摘要',
    result_json      LONGTEXT                           COMMENT '执行结果JSON',
    created_at       DATETIME                           COMMENT '创建时间',
    INDEX idx_sys_job_log_job_id_start_at (job_id, start_at),
    INDEX idx_sys_job_log_status (status),
    INDEX idx_sys_job_log_start_at (start_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='定时任务执行日志表';

-- ============================================================
-- 默认分组与运行时参数
-- ============================================================
INSERT IGNORE INTO sys_job_group (code, name, remark, sort_order, is_default, created_at, updated_at)
VALUES ('default', 'Default Group', 'The system default job group. Jobs are moved here when other groups are deleted.', 0, 1, NOW(), NOW());

INSERT IGNORE INTO sys_config (`name`, `key`, `value`, `remark`, `created_at`, `updated_at`)
VALUES (
    '定时任务-Shell 模式全局开关',
    'cron.shell.enabled',
    'false',
    '控制 Shell 类型任务是否允许创建、修改、触发与终止，可选值：true、false。',
    NOW(),
    NOW()
);

INSERT IGNORE INTO sys_config (`name`, `key`, `value`, `remark`, `created_at`, `updated_at`)
VALUES (
    '定时任务-执行日志保留策略',
    'cron.log.retention',
    '{"mode":"days","value":30}',
    '控制定时任务执行日志默认保留策略，使用 JSON：{"mode":"days|count|none","value":N}。',
    NOW(),
    NOW()
);

-- ============================================================
-- 字典类型与字典数据
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('定时任务状态', 'cron_job_status', 1, '定时任务状态列表', NOW(), NOW());
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('定时任务类型', 'cron_job_task_type', 1, '定时任务类型列表', NOW(), NOW());
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('定时任务调度范围', 'cron_job_scope', 1, '定时任务调度范围列表', NOW(), NOW());
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('定时任务并发策略', 'cron_job_concurrency', 1, '定时任务并发策略列表', NOW(), NOW());
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('定时任务触发方式', 'cron_job_trigger', 1, '定时任务触发方式列表', NOW(), NOW());
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('定时任务日志状态', 'cron_job_log_status', 1, '定时任务执行日志状态列表', NOW(), NOW());
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('定时任务日志保留模式', 'cron_log_retention_mode', 1, '定时任务日志保留模式列表', NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_status', '启用', 'enabled', 1, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_status', '停用', 'disabled', 2, 'default', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_status', '不可用', 'paused_by_plugin', 3, 'danger', 1, NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_task_type', 'Handler 任务', 'handler', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_task_type', 'Shell 任务', 'shell', 2, 'warning', 1, NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_scope', '仅主节点执行', 'master_only', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_scope', '所有节点执行', 'all_node', 2, 'success', 1, NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_concurrency', '单例执行', 'singleton', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_concurrency', '允许并行执行', 'parallel', 2, 'warning', 1, NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_trigger', 'Cron 调度', 'cron', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_trigger', '手动触发', 'manual', 2, 'success', 1, NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_log_status', '运行中', 'running', 1, 'processing', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_log_status', '成功', 'success', 2, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_log_status', '失败', 'failed', 3, 'danger', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_log_status', '已取消', 'cancelled', 4, 'warning', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_log_status', '超时', 'timeout', 5, 'danger', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_log_status', '非主节点跳过', 'skipped_not_primary', 6, 'default', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_log_status', '单例冲突跳过', 'skipped_singleton', 7, 'default', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_job_log_status', '并发上限跳过', 'skipped_max_concurrency', 8, 'default', 1, NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_log_retention_mode', '按天保留', 'days', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_log_retention_mode', '按条数保留', 'count', 2, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('cron_log_retention_mode', '不清理', 'none', 3, 'warning', 1, NOW(), NOW());

-- ============================================================
-- 菜单与按钮权限
-- ============================================================
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'scheduler') AS parent), 'system:job:list', '任务管理', '/system/job', 'system/job/index', 'system:job:list', 'lucide:clock-3', 'M', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'scheduler') AS parent), 'system:jobgroup:list', '分组管理', '/system/job-group', 'system/job-group/index', 'system:jobgroup:list', 'lucide:blocks', 'M', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'scheduler') AS parent), 'system:joblog:list', '执行日志', '/system/job-log', 'system/job-log/index', 'system:joblog:list', 'lucide:scroll-text', 'M', 3, 1, 1, 0, 0, NOW(), NOW());

INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:job:list') AS parent), 'system:job:add', '任务新增', '', '', 'system:job:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:job:list') AS parent), 'system:job:edit', '任务修改', '', '', 'system:job:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:job:list') AS parent), 'system:job:remove', '任务删除', '', '', 'system:job:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:job:list') AS parent), 'system:job:status', '任务启停', '', '', 'system:job:status', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:job:list') AS parent), 'system:job:trigger', '立即执行', '', '', 'system:job:trigger', '', 'B', 6, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:job:list') AS parent), 'system:job:reset', '重置计数', '', '', 'system:job:reset', '', 'B', 7, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:job:list') AS parent), 'system:job:shell', 'Shell 任务权限', '', '', 'system:job:shell', '', 'B', 8, 1, 1, 0, 0, NOW(), NOW());

INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:jobgroup:list') AS parent), 'system:jobgroup:add', '分组新增', '', '', 'system:jobgroup:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:jobgroup:list') AS parent), 'system:jobgroup:edit', '分组修改', '', '', 'system:jobgroup:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:jobgroup:list') AS parent), 'system:jobgroup:remove', '分组删除', '', '', 'system:jobgroup:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:joblog:list') AS parent), 'system:joblog:remove', '日志清空', '', '', 'system:joblog:remove', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:joblog:list') AS parent), 'system:joblog:cancel', '日志终止', '', '', 'system:joblog:cancel', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
