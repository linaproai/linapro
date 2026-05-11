-- 001: monitor-loginlog schema
-- 001：monitor-loginlog 数据结构

-- Purpose: Stores login audit records with tenant, impersonation, client, status, and login-time metadata.
-- 用途：存储登录审计记录，包括租户、代操作、客户端、状态与登录时间元数据。
CREATE TABLE IF NOT EXISTS plugin_monitor_loginlog (
    "id"          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"   INT          NOT NULL DEFAULT 0,
    "acting_user_id" INT       NOT NULL DEFAULT 0,
    "on_behalf_of_tenant_id" INT NOT NULL DEFAULT 0,
    "is_impersonation" BOOL   NOT NULL DEFAULT FALSE,
    "user_name"   VARCHAR(50)  NOT NULL DEFAULT '',
    "status"      SMALLINT                        NOT NULL DEFAULT 0,
    "ip"          VARCHAR(50)  NOT NULL DEFAULT '',
    "browser"     VARCHAR(200) NOT NULL DEFAULT '',
    "os"          VARCHAR(200) NOT NULL DEFAULT '',
    "msg"         VARCHAR(500) NOT NULL DEFAULT '',
    "login_time"  TIMESTAMP                       NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE plugin_monitor_loginlog IS 'System login log table';
COMMENT ON COLUMN plugin_monitor_loginlog."id" IS 'Log ID';
COMMENT ON COLUMN plugin_monitor_loginlog."tenant_id" IS 'Owning tenant ID, 0 means PLATFORM';
COMMENT ON COLUMN plugin_monitor_loginlog."acting_user_id" IS 'Actual acting user ID for platform operations or impersonation';
COMMENT ON COLUMN plugin_monitor_loginlog."on_behalf_of_tenant_id" IS 'Target tenant ID when a platform administrator acts on behalf of a tenant';
COMMENT ON COLUMN plugin_monitor_loginlog."is_impersonation" IS 'Whether this log was produced during tenant impersonation';
COMMENT ON COLUMN plugin_monitor_loginlog."user_name" IS 'Login account';
COMMENT ON COLUMN plugin_monitor_loginlog."status" IS 'Login status: 0=succeeded, 1=failed';
COMMENT ON COLUMN plugin_monitor_loginlog."ip" IS 'Login IP address';
COMMENT ON COLUMN plugin_monitor_loginlog."browser" IS 'Browser type';
COMMENT ON COLUMN plugin_monitor_loginlog."os" IS 'Operating system';
COMMENT ON COLUMN plugin_monitor_loginlog."msg" IS 'Prompt message';
COMMENT ON COLUMN plugin_monitor_loginlog."login_time" IS 'Login time';

CREATE INDEX IF NOT EXISTS idx_plugin_monitor_loginlog_tenant_time ON plugin_monitor_loginlog ("tenant_id", "login_time");
CREATE INDEX IF NOT EXISTS idx_plugin_monitor_loginlog_tenant_user ON plugin_monitor_loginlog ("tenant_id", "user_name");
CREATE INDEX IF NOT EXISTS idx_plugin_monitor_loginlog_impersonation ON plugin_monitor_loginlog ("tenant_id", "is_impersonation", "on_behalf_of_tenant_id");
