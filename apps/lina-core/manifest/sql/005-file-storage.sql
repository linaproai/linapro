-- 005: File Management

-- ============================================================
-- 文件管理表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_file (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT  'File ID',
    name        VARCHAR(255)  NOT NULL DEFAULT '' COMMENT  'Stored file name',
    original    VARCHAR(255)  NOT NULL DEFAULT '' COMMENT  'Original file name',
    suffix      VARCHAR(32)   NOT NULL DEFAULT '' COMMENT  'File suffix',
    scene       VARCHAR(64)   NOT NULL DEFAULT 'other' COMMENT  'Usage scene: avatar=user avatar, notice_image=notice image, notice_attachment=notice attachment, other=other',
    size        BIGINT        NOT NULL DEFAULT 0  COMMENT  'File size in bytes',
    hash        VARCHAR(64)   NOT NULL DEFAULT '' COMMENT  'File SHA-256 hash for deduplication',
    url         VARCHAR(512)  NOT NULL DEFAULT '' COMMENT  'File access URL',
    path        VARCHAR(512)  NOT NULL DEFAULT '' COMMENT  'File storage path',
    engine      VARCHAR(32)   NOT NULL DEFAULT 'local' COMMENT  'Storage engine: local=local storage',
    created_by  BIGINT        NOT NULL DEFAULT 0  COMMENT  'Uploader user ID',
    created_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT  'Creation time',
    updated_at  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT  'Update time',
    deleted_at  DATETIME      NULL     DEFAULT NULL COMMENT  'Deletion time',
    INDEX idx_engine (engine),
    INDEX idx_created_by (created_by),
    INDEX idx_suffix (suffix),
    INDEX idx_hash (hash),
    INDEX idx_scene (scene)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'File management table';
