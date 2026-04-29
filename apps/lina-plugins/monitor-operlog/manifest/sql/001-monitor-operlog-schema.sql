-- 001: monitor-operlog schema

CREATE TABLE IF NOT EXISTS plugin_monitor_operlog (
    id              INT PRIMARY KEY AUTO_INCREMENT COMMENT  'Log ID',
    title           VARCHAR(50)   NOT NULL DEFAULT '' COMMENT  'Module title',
    oper_summary    VARCHAR(200)  NOT NULL DEFAULT '' COMMENT  'Operation summary',
    route_owner     VARCHAR(100)  NOT NULL DEFAULT '' COMMENT  'Route owner: core or plugin ID',
    route_method    VARCHAR(20)   NOT NULL DEFAULT '' COMMENT  'Route request method',
    route_path      VARCHAR(255)  NOT NULL DEFAULT '' COMMENT  'Route path',
    route_doc_key   VARCHAR(255)  NOT NULL DEFAULT '' COMMENT  'API documentation structured key',
    oper_type       VARCHAR(20)   NOT NULL DEFAULT 'other' COMMENT  'Operation type: create=create, update=update, delete=delete, export=export, import=import, other=other',
    method          VARCHAR(200)  NOT NULL DEFAULT '' COMMENT  'Method name',
    request_method  VARCHAR(10)   NOT NULL DEFAULT '' COMMENT  'Request method: GET/POST/PUT/DELETE',
    oper_name       VARCHAR(50)   NOT NULL DEFAULT '' COMMENT  'Operator',
    oper_url        VARCHAR(500)  NOT NULL DEFAULT '' COMMENT  'Request URL',
    oper_ip         VARCHAR(50)   NOT NULL DEFAULT '' COMMENT  'Operation IP address',
    oper_param      TEXT          NOT NULL             COMMENT  'Request parameters',
    json_result     TEXT          NOT NULL             COMMENT  'Response parameters',
    status          TINYINT       NOT NULL DEFAULT 0   COMMENT  'Operation status: 0=succeeded, 1=failed',
    error_msg       TEXT          NOT NULL             COMMENT  'Error message',
    cost_time       INT           NOT NULL DEFAULT 0   COMMENT  'Duration in milliseconds',
    oper_time       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT  'Operation time'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Operation log table';

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
