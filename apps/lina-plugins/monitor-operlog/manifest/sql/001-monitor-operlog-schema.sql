-- 001: monitor-operlog schema

CREATE TABLE IF NOT EXISTS plugin_monitor_operlog (
    id              INT PRIMARY KEY AUTO_INCREMENT COMMENT '日志ID',
    title           VARCHAR(50)   NOT NULL DEFAULT '' COMMENT '模块标题',
    oper_summary    VARCHAR(200)  NOT NULL DEFAULT '' COMMENT '操作摘要',
    oper_type       VARCHAR(20)   NOT NULL DEFAULT 'other' COMMENT '操作类型（create新增 update修改 delete删除 export导出 import导入 other其他）',
    method          VARCHAR(200)  NOT NULL DEFAULT '' COMMENT '方法名称',
    request_method  VARCHAR(10)   NOT NULL DEFAULT '' COMMENT '请求方式（GET/POST/PUT/DELETE）',
    oper_name       VARCHAR(50)   NOT NULL DEFAULT '' COMMENT '操作人员',
    oper_url        VARCHAR(500)  NOT NULL DEFAULT '' COMMENT '请求URL',
    oper_ip         VARCHAR(50)   NOT NULL DEFAULT '' COMMENT '操作IP地址',
    oper_param      TEXT          NOT NULL             COMMENT '请求参数',
    json_result     TEXT          NOT NULL             COMMENT '返回参数',
    status          TINYINT       NOT NULL DEFAULT 0   COMMENT '操作状态（0成功 1失败）',
    error_msg       TEXT          NOT NULL             COMMENT '错误消息',
    cost_time       INT           NOT NULL DEFAULT 0   COMMENT '耗时（毫秒）',
    oper_time       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='操作日志记录表';

INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('操作类型', 'sys_oper_type', 1, '操作日志操作类型列表', NOW(), NOW());
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('操作状态', 'sys_oper_status', 1, '操作日志操作状态列表', NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '新增', 'create', 1, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '修改', 'update', 2, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '删除', 'delete', 3, 'danger', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '导出', 'export', 4, 'warning', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '导入', 'import', 5, 'processing', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_type', '其他', 'other', 6, 'default', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_status', '成功', '0', 1, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_oper_status', '失败', '1', 2, 'danger', 1, NOW(), NOW());
