-- 001: content-notice schema

CREATE TABLE IF NOT EXISTS plugin_content_notice (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '公告ID',
    title       VARCHAR(255)  NOT NULL DEFAULT '' COMMENT '公告标题',
    type        TINYINT       NOT NULL DEFAULT 1  COMMENT '公告类型（1通知 2公告）',
    content     LONGTEXT      NOT NULL             COMMENT '公告内容',
    file_ids    VARCHAR(500)  NOT NULL DEFAULT '' COMMENT '附件文件ID列表，逗号分隔',
    status      TINYINT       NOT NULL DEFAULT 0  COMMENT '公告状态（0草稿 1已发布）',
    remark      VARCHAR(500)  NOT NULL DEFAULT '' COMMENT '备注',
    created_by  BIGINT        NOT NULL DEFAULT 0  COMMENT '创建者',
    updated_by  BIGINT        NOT NULL DEFAULT 0  COMMENT '更新者',
    created_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at  DATETIME      NULL DEFAULT NULL COMMENT '删除时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='通知公告表';

INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('通知类型', 'sys_notice_type', 1, '通知公告类型列表', NOW(), NOW());
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('公告状态', 'sys_notice_status', 1, '通知公告状态列表', NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_notice_type', '通知', '1', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_notice_type', '公告', '2', 2, 'warning', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_notice_status', '草稿', '0', 1, 'default', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_notice_status', '已发布', '1', 2, 'success', 1, NOW(), NOW());
