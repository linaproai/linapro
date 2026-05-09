-- ============================================================
-- Online session table (persistent table, tracks currently online users)
-- 在线会话表（持久表，用于跟踪当前在线用户）
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_online_session (
    "token_id"         VARCHAR(64) NOT NULL,
    "user_id"          INT NOT NULL DEFAULT 0,
    "username"         VARCHAR(64) NOT NULL DEFAULT '',
    "dept_name"        VARCHAR(100) NOT NULL DEFAULT '',
    "ip"               VARCHAR(64) NOT NULL DEFAULT '',
    "browser"          VARCHAR(100) NOT NULL DEFAULT '',
    "os"               VARCHAR(100) NOT NULL DEFAULT '',
    "login_time"       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "last_active_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("token_id")
);

COMMENT ON TABLE sys_online_session IS 'Online session table';
COMMENT ON COLUMN sys_online_session."token_id" IS 'Session token ID (UUID)';
COMMENT ON COLUMN sys_online_session."user_id" IS 'User ID';
COMMENT ON COLUMN sys_online_session."username" IS 'Login account';
COMMENT ON COLUMN sys_online_session."dept_name" IS 'Department name';
COMMENT ON COLUMN sys_online_session."ip" IS 'Login IP';
COMMENT ON COLUMN sys_online_session."browser" IS 'Browser';
COMMENT ON COLUMN sys_online_session."os" IS 'Operating system';
COMMENT ON COLUMN sys_online_session."login_time" IS 'Login time';
COMMENT ON COLUMN sys_online_session."last_active_time" IS 'Last active time';

CREATE INDEX IF NOT EXISTS idx_sys_online_session_user_id ON sys_online_session ("user_id");
CREATE INDEX IF NOT EXISTS idx_sys_online_session_username ON sys_online_session ("username");
CREATE INDEX IF NOT EXISTS idx_sys_online_session_last_active_time ON sys_online_session ("last_active_time");

