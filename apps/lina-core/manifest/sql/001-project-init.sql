-- Database bootstrap
-- 数据库初始化

-- sys_user table
-- sys_user 用户表
CREATE TABLE IF NOT EXISTS sys_user (
    "id"         INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "username"   VARCHAR(64) NOT NULL,
    "password"   VARCHAR(256) NOT NULL,
    "nickname"   VARCHAR(64) NOT NULL DEFAULT '',
    "email"      VARCHAR(128) NOT NULL DEFAULT '',
    "phone"      VARCHAR(20) NOT NULL DEFAULT '',
    "sex"        SMALLINT NOT NULL DEFAULT 0,
    "avatar"     VARCHAR(512) NOT NULL DEFAULT '',
    "status"     SMALLINT NOT NULL DEFAULT 1,
    "remark"     VARCHAR(512) NOT NULL DEFAULT '',
    "login_date" TIMESTAMP,
    "created_at" TIMESTAMP,
    "updated_at" TIMESTAMP,
    "deleted_at" TIMESTAMP
);

-- PostgreSQL stores table and column comments through standalone COMMENT ON
-- statements. Unlike MySQL, PostgreSQL does not support inline COMMENT clauses
-- inside CREATE TABLE column definitions.
COMMENT ON TABLE sys_user IS 'User information table';
COMMENT ON COLUMN sys_user."id" IS 'User ID';
COMMENT ON COLUMN sys_user."username" IS 'Username';
COMMENT ON COLUMN sys_user."password" IS 'Password';
COMMENT ON COLUMN sys_user."nickname" IS 'User nickname';
COMMENT ON COLUMN sys_user."email" IS 'Email address';
COMMENT ON COLUMN sys_user."phone" IS 'Mobile phone number';
COMMENT ON COLUMN sys_user."sex" IS 'Gender: 0=unknown, 1=male, 2=female';
COMMENT ON COLUMN sys_user."avatar" IS 'Avatar URL';
COMMENT ON COLUMN sys_user."status" IS 'Status: 0=disabled, 1=enabled';
COMMENT ON COLUMN sys_user."remark" IS 'Remark';
COMMENT ON COLUMN sys_user."login_date" IS 'Last login time';
COMMENT ON COLUMN sys_user."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_user."updated_at" IS 'Update time';
COMMENT ON COLUMN sys_user."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_user_username ON sys_user ("username");
CREATE INDEX IF NOT EXISTS idx_sys_user_status ON sys_user ("status");
CREATE INDEX IF NOT EXISTS idx_sys_user_phone ON sys_user ("phone");
CREATE INDEX IF NOT EXISTS idx_sys_user_created_at ON sys_user ("created_at");

-- Default admin user (password: admin123, bcrypt hash)
-- 默认管理员用户（密码：admin123，bcrypt 哈希）
INSERT INTO sys_user ("username", "password", "nickname", "status", "created_at", "updated_at")
VALUES ('admin', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', 'Administrator', 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
