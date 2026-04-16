-- sys_user table
CREATE TABLE IF NOT EXISTS sys_user (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT '用户ID',
    username    VARCHAR(64)  NOT NULL          COMMENT '用户账号',
    password    VARCHAR(256) NOT NULL          COMMENT '密码',
    nickname    VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '用户昵称',
    email       VARCHAR(128) NOT NULL DEFAULT '' COMMENT '邮箱',
    phone       VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '手机号码',
    sex         TINYINT      NOT NULL DEFAULT 0  COMMENT '性别（0未知 1男 2女）',
    avatar      VARCHAR(512) NOT NULL DEFAULT '' COMMENT '头像地址',
    status      TINYINT      NOT NULL DEFAULT 1  COMMENT '状态（0停用 1正常）',
    remark      VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
    login_date  DATETIME                         COMMENT '最后登录时间',
    created_at  DATETIME                         COMMENT '创建时间',
    updated_at  DATETIME                         COMMENT '更新时间',
    deleted_at  DATETIME                         COMMENT '删除时间',
    UNIQUE(username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户信息表';

-- Default admin user (password: admin123, bcrypt hash)
INSERT IGNORE INTO sys_user (username, password, nickname, status, created_at, updated_at)
VALUES ('admin', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '管理员', 1, NOW(), NOW());
