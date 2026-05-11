-- 003: File Storage
-- 003：文件存储

-- ============================================================
-- Purpose: Stores uploaded file metadata, storage location, deduplication hash, business scene, and tenant ownership.
-- 用途：存储上传文件元数据、存储位置、去重哈希、业务场景与租户归属。
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_file (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT NOT NULL DEFAULT 0,
    "name"       VARCHAR(255) NOT NULL DEFAULT '',
    "original"   VARCHAR(255) NOT NULL DEFAULT '',
    "suffix"     VARCHAR(32) NOT NULL DEFAULT '',
    "scene"      VARCHAR(64) NOT NULL DEFAULT 'other',
    "size"       BIGINT NOT NULL DEFAULT 0,
    "hash"       VARCHAR(64) NOT NULL DEFAULT '',
    "url"        VARCHAR(512) NOT NULL DEFAULT '',
    "path"       VARCHAR(512) NOT NULL DEFAULT '',
    "engine"     VARCHAR(32) NOT NULL DEFAULT 'local',
    "created_by" BIGINT NOT NULL DEFAULT 0,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP NULL DEFAULT NULL
);

COMMENT ON TABLE sys_file IS 'File management table';
COMMENT ON COLUMN sys_file."id" IS 'File ID';
COMMENT ON COLUMN sys_file."tenant_id" IS 'Owning tenant ID, 0 means PLATFORM';
COMMENT ON COLUMN sys_file."name" IS 'Stored file name';
COMMENT ON COLUMN sys_file."original" IS 'Original file name';
COMMENT ON COLUMN sys_file."suffix" IS 'File suffix';
COMMENT ON COLUMN sys_file."scene" IS 'Usage scene: avatar=user avatar, notice_image=notice image, notice_attachment=notice attachment, other=other';
COMMENT ON COLUMN sys_file."size" IS 'File size in bytes';
COMMENT ON COLUMN sys_file."hash" IS 'File SHA-256 hash for deduplication';
COMMENT ON COLUMN sys_file."url" IS 'File access URL';
COMMENT ON COLUMN sys_file."path" IS 'File storage path';
COMMENT ON COLUMN sys_file."engine" IS 'Storage engine: local=local storage';
COMMENT ON COLUMN sys_file."created_by" IS 'Uploader user ID';
COMMENT ON COLUMN sys_file."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_file."updated_at" IS 'Update time';
COMMENT ON COLUMN sys_file."deleted_at" IS 'Deletion time';

CREATE INDEX IF NOT EXISTS idx_sys_file_engine ON sys_file ("engine");
CREATE INDEX IF NOT EXISTS idx_sys_file_tenant_created_by ON sys_file ("tenant_id", "created_by");
CREATE INDEX IF NOT EXISTS idx_sys_file_suffix ON sys_file ("suffix");
CREATE INDEX IF NOT EXISTS idx_sys_file_hash ON sys_file ("hash");
CREATE INDEX IF NOT EXISTS idx_sys_file_tenant_scene ON sys_file ("tenant_id", "scene");
