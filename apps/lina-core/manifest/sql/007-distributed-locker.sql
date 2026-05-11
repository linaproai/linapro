-- Distributed lock table (persistent table)
-- 分布式锁表（持久表）
-- Purpose: Coordinates leader election and distributed lock ownership with persistent holders and expirations.
-- 用途：通过持久化持有者与过期时间协调主节点选举和分布式锁归属。
-- Lock state is retained across service restarts and expires through expire_time.
-- 服务重启后保留锁状态，并通过 expire_time 自然过期

CREATE TABLE IF NOT EXISTS sys_locker (
    "id"          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "name"        VARCHAR(64) NOT NULL,
    "reason"      VARCHAR(255) DEFAULT '',
    "holder"      VARCHAR(64) DEFAULT '',
    "expire_time" TIMESTAMP NOT NULL,
    "created_at"  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    "updated_at"  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE sys_locker IS 'Distributed lock table';
COMMENT ON COLUMN sys_locker."id" IS 'Primary key ID';
COMMENT ON COLUMN sys_locker."name" IS 'Lock name, unique identifier';
COMMENT ON COLUMN sys_locker."reason" IS 'Reason for acquiring the lock';
COMMENT ON COLUMN sys_locker."holder" IS 'Lock holder identifier (node name)';
COMMENT ON COLUMN sys_locker."expire_time" IS 'Lock expiration time';
COMMENT ON COLUMN sys_locker."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_locker."updated_at" IS 'Update time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_locker_name ON sys_locker ("name");
CREATE INDEX IF NOT EXISTS idx_sys_locker_expire_time ON sys_locker ("expire_time");
