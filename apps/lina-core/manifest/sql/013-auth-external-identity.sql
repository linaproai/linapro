-- 013: Authentication External Identity Linkage
-- 013：认证外部身份绑定

-- Purpose: Stores host-owned linkages between a verified third-party identity
-- (provider + immutable subject) and a local sys_user account, so source-plugin
-- OIDC/OAuth flows can resolve an external identity to a host session without a
-- host-side provider registry. The linkage is intentionally hard-delete only
-- (no deleted_at): a link either exists or does not, and unlinking must free the
-- (provider, subject) unique key for a future relink instead of leaving a
-- soft-deleted ghost row under the unique index. The table is platform-scoped
-- (no tenant_id) because the identity binding is a property of the user account;
-- tenant selection happens after login through the pre-login token flow.
-- 用途：存储宿主独占的「已验证第三方身份（provider + 不可变 subject）」与本地
-- sys_user 账号之间的绑定关系，使源码插件的 OIDC/OAuth 流程无需宿主侧 provider
-- 注册表即可把外部身份解析为宿主会话。该绑定刻意采用硬删除（无 deleted_at）：
-- 绑定要么存在要么不存在，解绑必须释放 (provider, subject) 唯一键以便未来重新绑定，
-- 而不是在唯一索引下遗留软删除幽灵行。表为平台级（无 tenant_id），因为身份绑定
-- 属于用户账号本身；租户在登录后经预登录令牌流程选择。
CREATE TABLE IF NOT EXISTS sys_user_external_identity (
    "id"             INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "user_id"        INT NOT NULL,
    "provider"       VARCHAR(64) NOT NULL,
    "subject"        VARCHAR(191) NOT NULL,
    "plugin_id"      VARCHAR(128) NOT NULL,
    "email_snapshot" VARCHAR(191) NOT NULL DEFAULT '',
    "created_at"     TIMESTAMPTZ,
    "updated_at"     TIMESTAMPTZ
);

-- PostgreSQL stores table and column comments through standalone COMMENT ON
-- statements. Unlike MySQL, PostgreSQL does not support inline COMMENT clauses
-- inside CREATE TABLE column definitions.
COMMENT ON TABLE sys_user_external_identity IS 'Verified external identity to local user linkage';
COMMENT ON COLUMN sys_user_external_identity."id" IS 'External identity linkage ID';
COMMENT ON COLUMN sys_user_external_identity."user_id" IS 'Linked local sys_user ID';
COMMENT ON COLUMN sys_user_external_identity."provider" IS 'Stable external provider ID owned by the declaring plugin, e.g. google, discord';
COMMENT ON COLUMN sys_user_external_identity."subject" IS 'Immutable provider-issued subject identifier, e.g. OIDC sub';
COMMENT ON COLUMN sys_user_external_identity."plugin_id" IS 'Source-plugin ID that owns the provider and created the linkage';
COMMENT ON COLUMN sys_user_external_identity."email_snapshot" IS 'Email captured at link time for audit only, never used as a resolution key';
COMMENT ON COLUMN sys_user_external_identity."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_user_external_identity."updated_at" IS 'Update time';

-- The (provider, subject) unique index is the authoritative resolution key used
-- by external login to find the linked local user in a single lookup.
CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_user_external_identity_provider_subject
    ON sys_user_external_identity ("provider", "subject");
-- Supports listing/revoking all external identities for one user.
CREATE INDEX IF NOT EXISTS idx_sys_user_external_identity_user
    ON sys_user_external_identity ("user_id");
