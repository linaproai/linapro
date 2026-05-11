-- ------------------------------------------------------------
-- 015 distributed cache consistency SQL file
-- 015 分布式缓存一致性 SQL 文件
-- Purpose: Stores monotonic cache revisions by tenant, domain, and explicit scope for distributed invalidation.
-- 用途：按租户、缓存域和显式作用域存储单调递增缓存修订号，用于分布式失效协调。
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS sys_cache_revision (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT NOT NULL DEFAULT 0,
    "domain"     VARCHAR(64) NOT NULL DEFAULT '',
    "scope"      VARCHAR(128) NOT NULL DEFAULT '',
    "revision"   BIGINT NOT NULL DEFAULT 0,
    "reason"     VARCHAR(255) NOT NULL DEFAULT '',
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE sys_cache_revision IS 'Persistent cache revision coordination table';
COMMENT ON COLUMN sys_cache_revision."id" IS 'Primary key ID';
COMMENT ON COLUMN sys_cache_revision."tenant_id" IS 'Owning tenant ID, 0 means PLATFORM';
COMMENT ON COLUMN sys_cache_revision."domain" IS 'Cache domain, such as runtime-config, permission-access, or plugin-runtime';
COMMENT ON COLUMN sys_cache_revision."scope" IS 'Explicit invalidation scope, such as global, plugin:<id>, locale:<locale>, or user:<id>';
COMMENT ON COLUMN sys_cache_revision."revision" IS 'Monotonic cache revision for this domain and scope';
COMMENT ON COLUMN sys_cache_revision."reason" IS 'Latest change reason used for diagnostics';
COMMENT ON COLUMN sys_cache_revision."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_cache_revision."updated_at" IS 'Update time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_cache_revision_tenant_domain_scope ON sys_cache_revision ("tenant_id", "domain", "scope");
CREATE INDEX IF NOT EXISTS idx_sys_cache_revision_domain_updated_at ON sys_cache_revision ("domain", "updated_at");
