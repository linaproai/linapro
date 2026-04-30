-- Database bootstrap
-- 数据库初始化
CREATE DATABASE IF NOT EXISTS `linapro` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE `linapro`;

-- sys_user table
-- sys_user 用户表
CREATE TABLE IF NOT EXISTS sys_user (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT  'User ID',
    username    VARCHAR(64)  NOT NULL          COMMENT  'Username',
    password    VARCHAR(256) NOT NULL          COMMENT  'Password',
    nickname    VARCHAR(64)  NOT NULL DEFAULT '' COMMENT  'User nickname',
    email       VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Email address',
    phone       VARCHAR(20)  NOT NULL DEFAULT '' COMMENT  'Mobile phone number',
    sex         TINYINT      NOT NULL DEFAULT 0  COMMENT  'Gender: 0=unknown, 1=male, 2=female',
    avatar      VARCHAR(512) NOT NULL DEFAULT '' COMMENT  'Avatar URL',
    status      TINYINT      NOT NULL DEFAULT 1  COMMENT  'Status: 0=disabled, 1=enabled',
    remark      VARCHAR(512) NOT NULL DEFAULT '' COMMENT  'Remark',
    login_date  DATETIME                         COMMENT  'Last login time',
    created_at  DATETIME                         COMMENT  'Creation time',
    updated_at  DATETIME                         COMMENT  'Update time',
    deleted_at  DATETIME                         COMMENT  'Deletion time',
    UNIQUE(username),
    KEY idx_status (status),
    KEY idx_phone (phone),
    KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'User information table';

-- Default admin user (password: admin123, bcrypt hash)
-- 默认管理员用户（密码：admin123，bcrypt 哈希）
INSERT IGNORE INTO sys_user (username, password, nickname, status, created_at, updated_at)
VALUES ('admin', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', 'Administrator', 1, NOW(), NOW());
