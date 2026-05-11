-- 001: org-center schema
-- 001：org-center 数据结构

-- Purpose: Stores tenant-scoped department hierarchy, department profile fields, ordering, and status.
-- 用途：存储租户级部门层级、部门资料字段、排序与状态。
CREATE TABLE IF NOT EXISTS plugin_org_center_dept (
    "id"          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"   INT          NOT NULL DEFAULT 0,
    "parent_id"   INT          NOT NULL DEFAULT 0,
    "ancestors"   VARCHAR(512) NOT NULL DEFAULT '',
    "name"        VARCHAR(128) NOT NULL DEFAULT '',
    "code"        VARCHAR(64)  NOT NULL DEFAULT '',
    "order_num"   INT          NOT NULL DEFAULT 0,
    "leader"      INT          NOT NULL DEFAULT 0,
    "phone"       VARCHAR(20)  NOT NULL DEFAULT '',
    "email"       VARCHAR(128) NOT NULL DEFAULT '',
    "status"      SMALLINT     NOT NULL DEFAULT 1,
    "remark"      VARCHAR(512) NOT NULL DEFAULT '',
    "created_at"  TIMESTAMP,
    "updated_at"  TIMESTAMP,
    "deleted_at"  TIMESTAMP
);

-- Purpose: Stores tenant-scoped posts that can be associated with departments and assigned to users.
-- 用途：存储租户级岗位，可关联部门并分配给用户。
CREATE TABLE IF NOT EXISTS plugin_org_center_post (
    "id"          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"   INT          NOT NULL DEFAULT 0,
    "dept_id"     INT          NOT NULL DEFAULT 0,
    "code"        VARCHAR(128) NOT NULL DEFAULT '',
    "name"        VARCHAR(128) NOT NULL DEFAULT '',
    "sort"        INT          NOT NULL DEFAULT 0,
    "status"      SMALLINT     NOT NULL DEFAULT 1,
    "remark"      VARCHAR(512) NOT NULL DEFAULT '',
    "created_at"  TIMESTAMP,
    "updated_at"  TIMESTAMP,
    "deleted_at"  TIMESTAMP
);

-- Purpose: Stores tenant-scoped user-to-department membership relations.
-- 用途：存储租户级用户与部门的归属关系。
CREATE TABLE IF NOT EXISTS plugin_org_center_user_dept (
    "tenant_id" INT NOT NULL DEFAULT 0,
    "user_id" INT NOT NULL,
    "dept_id" INT NOT NULL,
    PRIMARY KEY ("tenant_id", "user_id", "dept_id")
);

-- Purpose: Stores tenant-scoped user-to-post assignment relations.
-- 用途：存储租户级用户与岗位的分配关系。
CREATE TABLE IF NOT EXISTS plugin_org_center_user_post (
    "tenant_id" INT NOT NULL DEFAULT 0,
    "user_id" INT NOT NULL,
    "post_id" INT NOT NULL,
    PRIMARY KEY ("tenant_id", "user_id", "post_id")
);

COMMENT ON TABLE plugin_org_center_dept IS 'Department table';
COMMENT ON COLUMN plugin_org_center_dept."id" IS 'Department ID';
COMMENT ON COLUMN plugin_org_center_dept."tenant_id" IS 'Owning tenant ID, 0 means PLATFORM';
COMMENT ON COLUMN plugin_org_center_dept."parent_id" IS 'Parent department ID';
COMMENT ON COLUMN plugin_org_center_dept."ancestors" IS 'Ancestor list';
COMMENT ON COLUMN plugin_org_center_dept."name" IS 'Department name';
COMMENT ON COLUMN plugin_org_center_dept."code" IS 'Department code';
COMMENT ON COLUMN plugin_org_center_dept."order_num" IS 'Display order';
COMMENT ON COLUMN plugin_org_center_dept."leader" IS 'Leader user ID';
COMMENT ON COLUMN plugin_org_center_dept."phone" IS 'Contact phone number';
COMMENT ON COLUMN plugin_org_center_dept."email" IS 'Email address';
COMMENT ON COLUMN plugin_org_center_dept."status" IS 'Status: 0=disabled, 1=enabled';
COMMENT ON COLUMN plugin_org_center_dept."remark" IS 'Remark';
COMMENT ON COLUMN plugin_org_center_dept."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_org_center_dept."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_org_center_dept."deleted_at" IS 'Deletion time';

COMMENT ON TABLE plugin_org_center_post IS 'Post information table';
COMMENT ON COLUMN plugin_org_center_post."id" IS 'Post ID';
COMMENT ON COLUMN plugin_org_center_post."tenant_id" IS 'Owning tenant ID, 0 means PLATFORM';
COMMENT ON COLUMN plugin_org_center_post."dept_id" IS 'Owning department ID';
COMMENT ON COLUMN plugin_org_center_post."code" IS 'Post code';
COMMENT ON COLUMN plugin_org_center_post."name" IS 'Post name';
COMMENT ON COLUMN plugin_org_center_post."sort" IS 'Display order';
COMMENT ON COLUMN plugin_org_center_post."status" IS 'Status: 0=disabled, 1=enabled';
COMMENT ON COLUMN plugin_org_center_post."remark" IS 'Remark';
COMMENT ON COLUMN plugin_org_center_post."created_at" IS 'Creation time';
COMMENT ON COLUMN plugin_org_center_post."updated_at" IS 'Update time';
COMMENT ON COLUMN plugin_org_center_post."deleted_at" IS 'Deletion time';

COMMENT ON TABLE plugin_org_center_user_dept IS 'User-department relation table';
COMMENT ON COLUMN plugin_org_center_user_dept."tenant_id" IS 'Owning tenant ID, 0 means PLATFORM';
COMMENT ON COLUMN plugin_org_center_user_dept."user_id" IS 'User ID';
COMMENT ON COLUMN plugin_org_center_user_dept."dept_id" IS 'Department ID';

COMMENT ON TABLE plugin_org_center_user_post IS 'User-post relation table';
COMMENT ON COLUMN plugin_org_center_user_post."tenant_id" IS 'Owning tenant ID, 0 means PLATFORM';
COMMENT ON COLUMN plugin_org_center_user_post."user_id" IS 'User ID';
COMMENT ON COLUMN plugin_org_center_user_post."post_id" IS 'Post ID';

CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_org_center_dept_tenant_code ON plugin_org_center_dept ("tenant_id", (NULLIF("code", '')));
CREATE INDEX IF NOT EXISTS idx_plugin_org_center_dept_tenant_code ON plugin_org_center_dept ("tenant_id", "code");
CREATE INDEX IF NOT EXISTS idx_plugin_org_center_dept_tenant_parent ON plugin_org_center_dept ("tenant_id", "parent_id");
CREATE UNIQUE INDEX IF NOT EXISTS uk_plugin_org_center_post_tenant_code ON plugin_org_center_post ("tenant_id", "code");
CREATE INDEX IF NOT EXISTS idx_plugin_org_center_post_tenant_dept ON plugin_org_center_post ("tenant_id", "dept_id");
CREATE INDEX IF NOT EXISTS idx_plugin_org_center_user_dept_tenant_dept_user ON plugin_org_center_user_dept ("tenant_id", "dept_id", "user_id");
CREATE INDEX IF NOT EXISTS idx_plugin_org_center_user_post_tenant_post ON plugin_org_center_user_post ("tenant_id", "post_id");
