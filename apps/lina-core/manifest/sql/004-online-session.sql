-- ============================================================
-- Purpose: Persists active login sessions so online-user queries, heartbeat updates, and forced logout can work across restarts.
-- 用途：持久化活跃登录会话，支持在线用户查询、心跳更新与跨重启强制下线。
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_online_session (
    "tenant_id"        INT NOT NULL DEFAULT 0,
    "token_id"         VARCHAR(64) NOT NULL,
    "user_id"          INT NOT NULL DEFAULT 0,
    "username"         VARCHAR(64) NOT NULL DEFAULT '',
    "dept_name"        VARCHAR(100) NOT NULL DEFAULT '',
    "ip"               VARCHAR(64) NOT NULL DEFAULT '',
    "browser"          VARCHAR(100) NOT NULL DEFAULT '',
    "os"               VARCHAR(100) NOT NULL DEFAULT '',
    "login_time"       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "last_active_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_sys_online_session_token PRIMARY KEY ("token_id")
);

COMMENT ON TABLE sys_online_session IS 'Online session table';
COMMENT ON COLUMN sys_online_session."tenant_id" IS 'Owning tenant ID, 0 means PLATFORM';
COMMENT ON COLUMN sys_online_session."token_id" IS 'Session token ID (UUID)';
COMMENT ON COLUMN sys_online_session."user_id" IS 'User ID';
COMMENT ON COLUMN sys_online_session."username" IS 'Login account';
COMMENT ON COLUMN sys_online_session."dept_name" IS 'Department name';
COMMENT ON COLUMN sys_online_session."ip" IS 'Login IP';
COMMENT ON COLUMN sys_online_session."browser" IS 'Browser';
COMMENT ON COLUMN sys_online_session."os" IS 'Operating system';
COMMENT ON COLUMN sys_online_session."login_time" IS 'Login time';
COMMENT ON COLUMN sys_online_session."last_active_time" IS 'Last active time';

CREATE INDEX IF NOT EXISTS idx_sys_online_session_tenant_user ON sys_online_session ("tenant_id", "user_id");
CREATE INDEX IF NOT EXISTS idx_sys_online_session_user_tenant ON sys_online_session ("user_id", "tenant_id");
CREATE INDEX IF NOT EXISTS idx_sys_online_session_tenant_login_time ON sys_online_session ("tenant_id", "login_time");
CREATE INDEX IF NOT EXISTS idx_sys_online_session_username ON sys_online_session ("username");
CREATE INDEX IF NOT EXISTS idx_sys_online_session_last_active_time ON sys_online_session ("last_active_time");
