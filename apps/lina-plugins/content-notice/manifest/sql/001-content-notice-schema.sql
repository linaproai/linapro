-- 001: content-notice schema
-- 001：content-notice 数据结构

CREATE TABLE IF NOT EXISTS plugin_content_notice (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT  'Notice ID',
    title       VARCHAR(255)  NOT NULL DEFAULT '' COMMENT  'Notice title',
    type        TINYINT       NOT NULL DEFAULT 1  COMMENT  'Notice type: 1=notification, 2=announcement',
    content     LONGTEXT      NOT NULL             COMMENT  'Notice content',
    file_ids    VARCHAR(500)  NOT NULL DEFAULT '' COMMENT  'Attachment file ID list, comma-separated',
    status      TINYINT       NOT NULL DEFAULT 0  COMMENT  'Notice status: 0=draft, 1=published',
    remark      VARCHAR(500)  NOT NULL DEFAULT '' COMMENT  'Remark',
    created_by  BIGINT        NOT NULL DEFAULT 0  COMMENT  'Creator',
    updated_by  BIGINT        NOT NULL DEFAULT 0  COMMENT  'Updater',
    created_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT  'Creation time',
    updated_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT  'Update time',
    deleted_at  DATETIME      NULL DEFAULT NULL COMMENT  'Deletion time'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Notice table';

INSERT IGNORE INTO sys_dict_type (name, type, status, is_builtin, remark, created_at, updated_at)
VALUES ('通知类型', 'sys_notice_type', 1, 1, '通知公告类型列表', NOW(), NOW());
INSERT IGNORE INTO sys_dict_type (name, type, status, is_builtin, remark, created_at, updated_at)
VALUES ('公告状态', 'sys_notice_status', 1, 1, '通知公告状态列表', NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, is_builtin, created_at, updated_at)
VALUES ('sys_notice_type', '通知', '1', 1, 'primary', 1, 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, is_builtin, created_at, updated_at)
VALUES ('sys_notice_type', '公告', '2', 2, 'warning', 1, 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, is_builtin, created_at, updated_at)
VALUES ('sys_notice_status', '草稿', '0', 1, 'default', 1, 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, is_builtin, created_at, updated_at)
VALUES ('sys_notice_status', '已发布', '1', 2, 'success', 1, 1, NOW(), NOW());
