-- 002: Dict Management, Dept Management, Post Management, User-Dept-Post Association
-- 002：字典管理、部门管理、岗位管理、用户-部门-岗位关联

-- ============================================================
-- Purpose: Stores dictionary type definitions, including tenant ownership and whether tenant overrides are allowed.
-- 用途：存储字典类型定义，包括租户归属以及是否允许租户覆盖。
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_dict_type (
    "id"         INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT NOT NULL DEFAULT 0,
    "name"       VARCHAR(128) NOT NULL DEFAULT '',
    "type"     VARCHAR(128) NOT NULL DEFAULT '',
    "status"     SMALLINT NOT NULL DEFAULT 1,
    "is_builtin" SMALLINT NOT NULL DEFAULT 0,
    "allow_tenant_override" BOOL NOT NULL DEFAULT FALSE,
    "remark"     VARCHAR(512) NOT NULL DEFAULT '',
    "created_at" TIMESTAMP,
    "updated_at" TIMESTAMP,
    "deleted_at" TIMESTAMP DEFAULT NULL
);

COMMENT ON TABLE sys_dict_type IS 'Dictionary type table';
COMMENT ON COLUMN sys_dict_type."id" IS 'Dictionary type ID';
COMMENT ON COLUMN sys_dict_type."tenant_id" IS 'Owning tenant ID, 0 means PLATFORM default';
COMMENT ON COLUMN sys_dict_type."name" IS 'Dictionary name';
COMMENT ON COLUMN sys_dict_type."type" IS 'Dictionary type';
COMMENT ON COLUMN sys_dict_type."status" IS 'Status: 0=disabled, 1=enabled';
COMMENT ON COLUMN sys_dict_type."is_builtin" IS 'Built-in record flag: 1=yes, 0=no';
COMMENT ON COLUMN sys_dict_type."allow_tenant_override" IS 'Whether tenants may override this dictionary type';
COMMENT ON COLUMN sys_dict_type."remark" IS 'Remark';
COMMENT ON COLUMN sys_dict_type."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_dict_type."updated_at" IS 'Update time';
COMMENT ON COLUMN sys_dict_type."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_dict_type_tenant_type ON sys_dict_type ("tenant_id", "type");

-- ============================================================
-- Purpose: Stores dictionary option values, display labels, ordering, tag styles, and tenant-scoped overrides.
-- 用途：存储字典选项值、展示标签、排序、标签样式与租户级覆盖数据。
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_dict_data (
    "id"         INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT NOT NULL DEFAULT 0,
    "dict_type"  VARCHAR(128) NOT NULL DEFAULT '',
    "label"      VARCHAR(128) NOT NULL DEFAULT '',
    "value"    VARCHAR(128) NOT NULL DEFAULT '',
    "sort"       INT NOT NULL DEFAULT 0,
    "tag_style"  VARCHAR(128) NOT NULL DEFAULT '',
    "css_class"  VARCHAR(128) NOT NULL DEFAULT '',
    "status"     SMALLINT NOT NULL DEFAULT 1,
    "is_builtin" SMALLINT NOT NULL DEFAULT 0,
    "remark"     VARCHAR(512) NOT NULL DEFAULT '',
    "created_at" TIMESTAMP,
    "updated_at" TIMESTAMP,
    "deleted_at" TIMESTAMP DEFAULT NULL
);

COMMENT ON TABLE sys_dict_data IS 'Dictionary data table';
COMMENT ON COLUMN sys_dict_data."id" IS 'Dictionary data ID';
COMMENT ON COLUMN sys_dict_data."tenant_id" IS 'Owning tenant ID, 0 means PLATFORM default';
COMMENT ON COLUMN sys_dict_data."dict_type" IS 'Dictionary type';
COMMENT ON COLUMN sys_dict_data."label" IS 'Dictionary label';
COMMENT ON COLUMN sys_dict_data."value" IS 'Dictionary value';
COMMENT ON COLUMN sys_dict_data."sort" IS 'Display order';
COMMENT ON COLUMN sys_dict_data."tag_style" IS 'Tag style: primary/success/danger/warning, etc.';
COMMENT ON COLUMN sys_dict_data."css_class" IS 'CSS class name';
COMMENT ON COLUMN sys_dict_data."status" IS 'Status: 0=disabled, 1=enabled';
COMMENT ON COLUMN sys_dict_data."is_builtin" IS 'Built-in record flag: 1=yes, 0=no';
COMMENT ON COLUMN sys_dict_data."remark" IS 'Remark';
COMMENT ON COLUMN sys_dict_data."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_dict_data."updated_at" IS 'Update time';
COMMENT ON COLUMN sys_dict_data."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_dict_data_tenant_type_value ON sys_dict_data ("tenant_id", "dict_type", "value");

-- ============================================================
-- Dictionary seed data required by the host core
-- 字典初始化数据（宿主核心必需）
-- ============================================================

-- Dictionary type: status switch
-- 字典类型: 状态开关
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, '状态开关', 'sys_normal_disable', 1, 1, '状态开关列表', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Dictionary type: user gender
-- 字典类型: 用户性别
INSERT INTO sys_dict_type ("tenant_id", "name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES (0, '用户性别', 'sys_user_sex', 1, 1, '用户性别列表', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Dictionary data: status switch
-- 字典数据: 状态开关
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'sys_normal_disable', '正常', '1', 1, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'sys_normal_disable', '停用', '0', 2, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Dictionary data: user gender
-- 字典数据: 用户性别
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'sys_user_sex', '男', '1', 1, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'sys_user_sex', '女', '2', 2, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("tenant_id", "dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES (0, 'sys_user_sex', '未知', '0', 3, 'default', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
