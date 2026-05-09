-- 001: monitor-loginlog schema
-- 001：monitor-loginlog 数据结构

CREATE TABLE IF NOT EXISTS plugin_monitor_loginlog (
    "id"          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
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
COMMENT ON COLUMN plugin_monitor_loginlog."user_name" IS 'Login account';
COMMENT ON COLUMN plugin_monitor_loginlog."status" IS 'Login status: 0=succeeded, 1=failed';
COMMENT ON COLUMN plugin_monitor_loginlog."ip" IS 'Login IP address';
COMMENT ON COLUMN plugin_monitor_loginlog."browser" IS 'Browser type';
COMMENT ON COLUMN plugin_monitor_loginlog."os" IS 'Operating system';
COMMENT ON COLUMN plugin_monitor_loginlog."msg" IS 'Prompt message';
COMMENT ON COLUMN plugin_monitor_loginlog."login_time" IS 'Login time';
