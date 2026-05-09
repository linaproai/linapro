-- 001: monitor-operlog schema
-- 001：monitor-operlog 数据结构

CREATE TABLE IF NOT EXISTS plugin_monitor_operlog (
    "id"              INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "title"           VARCHAR(50)   NOT NULL DEFAULT '',
    "oper_summary"    VARCHAR(200)  NOT NULL DEFAULT '',
    "route_owner"     VARCHAR(100)  NOT NULL DEFAULT '',
    "route_method"    VARCHAR(20)   NOT NULL DEFAULT '',
    "route_path"      VARCHAR(255)  NOT NULL DEFAULT '',
    "route_doc_key"   VARCHAR(255)  NOT NULL DEFAULT '',
    "oper_type"       VARCHAR(20)   NOT NULL DEFAULT 'other',
    "method"          VARCHAR(200)  NOT NULL DEFAULT '',
    "request_method"  VARCHAR(10)   NOT NULL DEFAULT '',
    "oper_name"       VARCHAR(50)   NOT NULL DEFAULT '',
    "oper_url"        VARCHAR(500)  NOT NULL DEFAULT '',
    "oper_ip"         VARCHAR(50)   NOT NULL DEFAULT '',
    "oper_param"      TEXT                             NOT NULL,
    "json_result"     TEXT                             NOT NULL,
    "status"          SMALLINT                         NOT NULL DEFAULT 0,
    "error_msg"       TEXT                             NOT NULL,
    "cost_time"       INT                              NOT NULL DEFAULT 0,
    "oper_time"       TIMESTAMP                        NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE plugin_monitor_operlog IS 'Operation log table';
COMMENT ON COLUMN plugin_monitor_operlog."id" IS 'Log ID';
COMMENT ON COLUMN plugin_monitor_operlog."title" IS 'Module title';
COMMENT ON COLUMN plugin_monitor_operlog."oper_summary" IS 'Operation summary';
COMMENT ON COLUMN plugin_monitor_operlog."route_owner" IS 'Route owner: core or plugin ID';
COMMENT ON COLUMN plugin_monitor_operlog."route_method" IS 'Route request method';
COMMENT ON COLUMN plugin_monitor_operlog."route_path" IS 'Route path';
COMMENT ON COLUMN plugin_monitor_operlog."route_doc_key" IS 'API documentation structured key';
COMMENT ON COLUMN plugin_monitor_operlog."oper_type" IS 'Operation type: create=create, update=update, delete=delete, export=export, import=import, other=other';
COMMENT ON COLUMN plugin_monitor_operlog."method" IS 'Method name';
COMMENT ON COLUMN plugin_monitor_operlog."request_method" IS 'Request method: GET/POST/PUT/DELETE';
COMMENT ON COLUMN plugin_monitor_operlog."oper_name" IS 'Operator';
COMMENT ON COLUMN plugin_monitor_operlog."oper_url" IS 'Request URL';
COMMENT ON COLUMN plugin_monitor_operlog."oper_ip" IS 'Operation IP address';
COMMENT ON COLUMN plugin_monitor_operlog."oper_param" IS 'Request parameters';
COMMENT ON COLUMN plugin_monitor_operlog."json_result" IS 'Response parameters';
COMMENT ON COLUMN plugin_monitor_operlog."status" IS 'Operation status: 0=succeeded, 1=failed';
COMMENT ON COLUMN plugin_monitor_operlog."error_msg" IS 'Error message';
COMMENT ON COLUMN plugin_monitor_operlog."cost_time" IS 'Duration in milliseconds';
COMMENT ON COLUMN plugin_monitor_operlog."oper_time" IS 'Operation time';

INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('操作类型', 'sys_oper_type', 1, 1, '操作日志操作类型列表', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('操作状态', 'sys_oper_status', 1, 1, '操作日志操作状态列表', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_oper_type', '新增', 'create', 1, 'success', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_oper_type', '修改', 'update', 2, 'primary', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_oper_type', '删除', 'delete', 3, 'danger', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_oper_type', '导出', 'export', 4, 'warning', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_oper_type', '导入', 'import', 5, 'processing', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_oper_type', '其他', 'other', 6, 'default', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_oper_status', '成功', '0', 1, 'success', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_oper_status', '失败', '1', 2, 'danger', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;
