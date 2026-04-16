-- 005: File Management

-- ============================================================
-- 文件管理表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_file (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '文件ID',
    name        VARCHAR(255)  NOT NULL DEFAULT '' COMMENT '存储文件名',
    original    VARCHAR(255)  NOT NULL DEFAULT '' COMMENT '原始文件名',
    suffix      VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '文件后缀',
    scene       VARCHAR(64)   NOT NULL DEFAULT 'other' COMMENT '使用场景：avatar=用户头像 notice_image=通知公告图片 notice_attachment=通知公告附件 other=其他',
    size        BIGINT        NOT NULL DEFAULT 0  COMMENT '文件大小（字节）',
    hash        VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '文件SHA-256散列值，用于去重',
    url         VARCHAR(512)  NOT NULL DEFAULT '' COMMENT '文件访问URL',
    path        VARCHAR(512)  NOT NULL DEFAULT '' COMMENT '文件存储路径',
    engine      VARCHAR(32)   NOT NULL DEFAULT 'local' COMMENT '存储引擎：local=本地',
    created_by  BIGINT        NOT NULL DEFAULT 0  COMMENT '上传者用户ID',
    created_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at  DATETIME      NULL     DEFAULT NULL COMMENT '删除时间',
    INDEX idx_engine (engine),
    INDEX idx_created_by (created_by),
    INDEX idx_suffix (suffix),
    INDEX idx_hash (hash),
    INDEX idx_scene (scene)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='文件管理表';

-- ============================================================
-- 通知公告表增加附件字段
-- ============================================================
ALTER TABLE sys_notice ADD COLUMN file_ids VARCHAR(500) NOT NULL DEFAULT '' COMMENT '附件文件ID列表，逗号分隔' AFTER content;
