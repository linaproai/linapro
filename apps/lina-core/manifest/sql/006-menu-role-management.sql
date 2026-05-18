-- ============================================================
-- Purpose: Stores route, menu, button, permission, and display metadata used to assemble the management workbench.
-- 用途：存储用于装配管理工作台的路由、菜单、按钮、权限与展示元数据。
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_menu (
    "id"          INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "parent_id"   INT NOT NULL DEFAULT 0,
    "menu_key"    VARCHAR(128) NULL,
    "name"        VARCHAR(128) NOT NULL DEFAULT '',
    "path"        VARCHAR(255) NOT NULL DEFAULT '',
    "component"   VARCHAR(255) NOT NULL DEFAULT '',
    "perms"       VARCHAR(128) NOT NULL DEFAULT '',
    "icon"        VARCHAR(128) NOT NULL DEFAULT '',
    "type"      CHAR(1) NOT NULL DEFAULT 'M',
    "sort"        INT NOT NULL DEFAULT 0,
    "visible"     SMALLINT NOT NULL DEFAULT 1,
    "status"      SMALLINT NOT NULL DEFAULT 1,
    "is_frame"    SMALLINT NOT NULL DEFAULT 0,
    "is_cache"    SMALLINT NOT NULL DEFAULT 0,
    "query_param" VARCHAR(255) NOT NULL DEFAULT '',
    "remark"      VARCHAR(512) NOT NULL DEFAULT '',
    "created_at"  TIMESTAMP,
    "updated_at"  TIMESTAMP,
    "deleted_at"  TIMESTAMP
);

COMMENT ON TABLE sys_menu IS 'Menu permission table';
COMMENT ON COLUMN sys_menu."id" IS 'Menu ID';
COMMENT ON COLUMN sys_menu."parent_id" IS 'Parent menu ID, 0 means root menu';
COMMENT ON COLUMN sys_menu."menu_key" IS 'Stable menu business key';
COMMENT ON COLUMN sys_menu."name" IS 'Menu name with i18n support';
COMMENT ON COLUMN sys_menu."path" IS 'Route path';
COMMENT ON COLUMN sys_menu."component" IS 'Component path';
COMMENT ON COLUMN sys_menu."perms" IS 'Permission identifier';
COMMENT ON COLUMN sys_menu."icon" IS 'Menu icon';
COMMENT ON COLUMN sys_menu."type" IS 'Menu type: D=directory, M=menu, B=button';
COMMENT ON COLUMN sys_menu."sort" IS 'Display order';
COMMENT ON COLUMN sys_menu."visible" IS 'Visibility: 1=visible, 0=hidden';
COMMENT ON COLUMN sys_menu."status" IS 'Status: 0=disabled, 1=enabled';
COMMENT ON COLUMN sys_menu."is_frame" IS 'External link flag: 1=yes, 0=no';
COMMENT ON COLUMN sys_menu."is_cache" IS 'Cache flag: 1=yes, 0=no';
COMMENT ON COLUMN sys_menu."query_param" IS 'Route parameters in JSON format';
COMMENT ON COLUMN sys_menu."remark" IS 'Remark';
COMMENT ON COLUMN sys_menu."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_menu."updated_at" IS 'Update time';
COMMENT ON COLUMN sys_menu."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_menu_menu_key ON sys_menu ("menu_key");

-- ============================================================
-- Purpose: Stores tenant-scoped and platform roles, including permission keys, display order, status, and data scope.
-- 用途：存储租户级与平台级角色，包括权限标识、展示排序、状态与数据权限范围。
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_role (
    "id"         INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "tenant_id"  INT NOT NULL DEFAULT 0,
    "name"       VARCHAR(64) NOT NULL DEFAULT '',
    "key"      VARCHAR(64) NOT NULL DEFAULT '',
    "sort"       INT NOT NULL DEFAULT 0,
    "data_scope" SMALLINT NOT NULL DEFAULT 2,
    "status"     SMALLINT NOT NULL DEFAULT 1,
    "remark"     VARCHAR(512) NOT NULL DEFAULT '',
    "created_at" TIMESTAMP,
    "updated_at" TIMESTAMP,
    "deleted_at" TIMESTAMP
);

COMMENT ON TABLE sys_role IS 'Role information table';
COMMENT ON COLUMN sys_role."id" IS 'Role ID';
COMMENT ON COLUMN sys_role."tenant_id" IS 'Owning tenant ID, 0 means PLATFORM';
COMMENT ON COLUMN sys_role."name" IS 'Role name';
COMMENT ON COLUMN sys_role."key" IS 'Permission key';
COMMENT ON COLUMN sys_role."sort" IS 'Display order';
COMMENT ON COLUMN sys_role."data_scope" IS 'Data scope: 1=all, 2=tenant, 3=department, 4=self';
COMMENT ON COLUMN sys_role."status" IS 'Status: 0=disabled, 1=enabled';
COMMENT ON COLUMN sys_role."remark" IS 'Remark';
COMMENT ON COLUMN sys_role."created_at" IS 'Creation time';
COMMENT ON COLUMN sys_role."updated_at" IS 'Update time';
COMMENT ON COLUMN sys_role."deleted_at" IS 'Deletion time';

CREATE UNIQUE INDEX IF NOT EXISTS uk_sys_role_tenant_key ON sys_role ("tenant_id", "key");

-- ============================================================
-- Purpose: Stores role-to-menu and role-to-button grants for each tenant or platform context.
-- 用途：存储每个租户或平台上下文中的角色到菜单、按钮的授权关系。
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_role_menu (
    "tenant_id" INT NOT NULL DEFAULT 0,
    "role_id" INT NOT NULL,
    "menu_id" INT NOT NULL,
    CONSTRAINT pk_sys_role_menu_tenant PRIMARY KEY ("role_id", "menu_id", "tenant_id")
);

COMMENT ON TABLE sys_role_menu IS 'Role-menu relation table';
COMMENT ON COLUMN sys_role_menu."tenant_id" IS 'Role-menu relation tenant ID, 0 means PLATFORM';
COMMENT ON COLUMN sys_role_menu."role_id" IS 'Role ID';
COMMENT ON COLUMN sys_role_menu."menu_id" IS 'Menu ID';

CREATE INDEX IF NOT EXISTS idx_sys_role_menu_menu_id ON sys_role_menu ("menu_id");
CREATE INDEX IF NOT EXISTS idx_sys_role_menu_tenant_role ON sys_role_menu ("tenant_id", "role_id");

-- ============================================================
-- Purpose: Stores user role assignments within a tenant or platform context.
-- 用途：存储用户在租户或平台上下文中的角色分配关系。
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_user_role (
    "tenant_id" INT NOT NULL DEFAULT 0,
    "user_id" INT NOT NULL,
    "role_id" INT NOT NULL,
    CONSTRAINT pk_sys_user_role_tenant PRIMARY KEY ("user_id", "role_id", "tenant_id")
);

COMMENT ON TABLE sys_user_role IS 'User-role relation table';
COMMENT ON COLUMN sys_user_role."tenant_id" IS 'Role assignment tenant ID, 0 means PLATFORM';
COMMENT ON COLUMN sys_user_role."user_id" IS 'User ID';
COMMENT ON COLUMN sys_user_role."role_id" IS 'Role ID';

CREATE INDEX IF NOT EXISTS idx_sys_user_role_role_id ON sys_user_role ("role_id");
CREATE INDEX IF NOT EXISTS idx_sys_user_role_tenant_role ON sys_user_role ("tenant_id", "role_id");

-- ============================================================
-- Dictionary types and dictionary data
-- 字典类型与字典数据
-- ============================================================
INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('菜单状态', 'sys_menu_status', 1, 1, '菜单状态列表', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('显示状态', 'sys_show_hide', 1, 1, '显示状态列表', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('菜单类型', 'sys_menu_type', 1, 1, '菜单类型列表', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_type ("name", "type", "status", "is_builtin", "remark", "created_at", "updated_at")
VALUES ('数据权限范围', 'sys_data_scope', 1, 1, '数据权限范围列表', NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_menu_status', '正常', '1', 1, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_menu_status', '停用', '0', 2, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_show_hide', '显示', '1', 1, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_show_hide', '隐藏', '0', 2, 'danger', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_menu_type', '目录', 'D', 1, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_menu_type', '菜单', 'M', 2, 'success', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_menu_type', '按钮', 'B', 3, 'warning', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_data_scope', '全部数据', '1', 1, 'primary', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_data_scope', '本租户数据', '2', 2, 'orange', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_data_scope', '本部门数据', '3', 3, 'purple', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_dict_data ("dict_type", "label", "value", "sort", "tag_style", "status", "is_builtin", "created_at", "updated_at")
VALUES ('sys_data_scope', '本人数据', '4', 4, 'green', 1, 1, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ============================================================
-- Initial role data
-- 初始化角色数据
-- ============================================================
INSERT INTO sys_role ("tenant_id", "name", "key", "sort", "data_scope", "status", "remark", "created_at", "updated_at")
VALUES (0, '超级管理员', 'admin', 1, 1, 1, '超级管理员，拥有所有权限', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_role ("tenant_id", "name", "key", "sort", "data_scope", "status", "remark", "created_at", "updated_at")
VALUES (0, '普通用户', 'user', 2, 4, 1, '普通用户，仅查看本人数据', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ============================================================
-- Host stable top-level directory skeleton
-- 宿主稳定一级目录骨架
-- ============================================================
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES (0, 'dashboard', '工作台', 'dashboard', '', '', 'lucide:layout-dashboard', 'D', 1, 1, 1, 0, 0, '宿主稳定目录：工作台', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES (0, 'iam', '权限管理', 'iam', '', '', 'lucide:shield-check', 'D', 2, 1, 1, 0, 0, '宿主稳定目录：权限管理', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES (0, 'platform', '平台管理', 'platform', '', '', 'lucide:building-2', 'D', 3, 1, 1, 0, 0, '宿主稳定目录：平台管理', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES (0, 'org', '组织管理', 'org', '', '', 'lucide:network', 'D', 4, 1, 1, 0, 0, '宿主稳定目录：组织管理', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES (0, 'setting', '系统设置', 'setting', '', '', 'lucide:settings-2', 'D', 5, 1, 1, 0, 0, '宿主稳定目录：系统设置', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES (0, 'content', '内容管理', 'content', '', '', 'lucide:newspaper', 'D', 6, 1, 1, 0, 0, '宿主稳定目录：内容管理', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES (0, 'monitor', '系统监控', 'monitor', '', '', 'lucide:activity', 'D', 7, 1, 1, 0, 0, '宿主稳定目录：系统监控', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES (0, 'scheduler', '任务调度', 'scheduler', '', '', 'lucide:calendar-range', 'D', 8, 1, 1, 0, 0, '宿主稳定目录：任务调度', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES (0, 'extension', '扩展中心', 'extension', '', '', 'lucide:puzzle', 'D', 9, 1, 1, 0, 0, '宿主稳定目录：扩展中心', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES (0, 'developer', '开发中心', 'developer', '', '', 'lucide:flask-conical', 'D', 10, 1, 1, 0, 0, '宿主稳定目录：开发中心', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ============================================================
-- Workbench menus
-- 工作台菜单
-- ============================================================
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'dashboard') AS parent), 'dashboard:analytics:list', '分析页', 'analytics', 'dashboard/analytics/index', 'dashboard:analytics:list', 'lucide:area-chart', 'M', 1, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'dashboard') AS parent), 'dashboard:workspace:list', '工作台', 'workspace', 'dashboard/workspace/index', 'dashboard:workspace:list', 'carbon:workspace', 'M', 2, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ============================================================
-- Access-control menus
-- 权限管理菜单
-- ============================================================
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'iam') AS parent), 'system:user:list', '用户管理', '/system/user', 'system/user/index', 'system:user:list', 'ant-design:user-outlined', 'M', 1, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:user:list') AS parent), 'system:user:query', '用户查询', '', '', 'system:user:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:user:list') AS parent), 'system:user:add', '用户新增', '', '', 'system:user:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:user:list') AS parent), 'system:user:edit', '用户修改', '', '', 'system:user:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:user:list') AS parent), 'system:user:remove', '用户删除', '', '', 'system:user:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:user:list') AS parent), 'system:user:export', '用户导出', '', '', 'system:user:export', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:user:list') AS parent), 'system:user:import', '用户导入', '', '', 'system:user:import', '', 'B', 6, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:user:list') AS parent), 'system:user:resetPwd', '重置密码', '', '', 'system:user:resetPwd', '', 'B', 7, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'iam') AS parent), 'system:role:list', '角色管理', '/system/role', 'system/role/index', 'system:role:list', 'lucide:shield', 'M', 2, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:role:list') AS parent), 'system:role:query', '角色查询', '', '', 'system:role:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:role:list') AS parent), 'system:role:add', '角色新增', '', '', 'system:role:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:role:list') AS parent), 'system:role:edit', '角色修改', '', '', 'system:role:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:role:list') AS parent), 'system:role:remove', '角色删除', '', '', 'system:role:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:role:list') AS parent), 'system:role:auth', '角色授权', '', '', 'system:role:auth', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'iam') AS parent), 'system:menu:list', '菜单管理', '/system/menu', 'system/menu/index', 'system:menu:list', 'lucide:menu', 'M', 3, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:menu:list') AS parent), 'system:menu:query', '菜单查询', '', '', 'system:menu:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:menu:list') AS parent), 'system:menu:add', '菜单新增', '', '', 'system:menu:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:menu:list') AS parent), 'system:menu:edit', '菜单修改', '', '', 'system:menu:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:menu:list') AS parent), 'system:menu:remove', '菜单删除', '', '', 'system:menu:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ============================================================
-- System setting menus
-- 系统设置菜单
-- ============================================================
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'setting') AS parent), 'system:dict:list', '字典管理', '/system/dict', 'system/dict/index', 'system:dict:list', 'lucide:book-open', 'M', 1, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:dict:list') AS parent), 'system:dict:query', '字典查询', '', '', 'system:dict:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:dict:list') AS parent), 'system:dict:add', '字典新增', '', '', 'system:dict:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:dict:list') AS parent), 'system:dict:edit', '字典修改', '', '', 'system:dict:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:dict:list') AS parent), 'system:dict:remove', '字典删除', '', '', 'system:dict:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:dict:list') AS parent), 'system:dict:export', '字典导出', '', '', 'system:dict:export', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'setting') AS parent), 'system:config:list', '参数设置', '/system/config', 'system/config/index', 'system:config:list', 'lucide:sliders-horizontal', 'M', 2, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:config:list') AS parent), 'system:config:query', '参数查询', '', '', 'system:config:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:config:list') AS parent), 'system:config:add', '参数新增', '', '', 'system:config:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:config:list') AS parent), 'system:config:edit', '参数修改', '', '', 'system:config:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:config:list') AS parent), 'system:config:remove', '参数删除', '', '', 'system:config:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:config:list') AS parent), 'system:config:export', '参数导出', '', '', 'system:config:export', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'setting') AS parent), 'system:file:list', '文件管理', '/system/file', 'system/file/index', 'system:file:list', 'lucide:folder-open', 'M', 3, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:file:list') AS parent), 'system:file:query', '文件查询', '', '', 'system:file:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:file:list') AS parent), 'system:file:upload', '文件上传', '', '', 'system:file:upload', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:file:list') AS parent), 'system:file:download', '文件下载', '', '', 'system:file:download', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'system:file:list') AS parent), 'system:file:remove', '文件删除', '', '', 'system:file:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ============================================================
-- Extension center menus
-- 扩展中心菜单
-- ============================================================
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'extension') AS parent), 'extension:plugin:list', '插件管理', '/system/plugin', 'system/plugin/index', 'plugin:list', 'lucide:plug', 'M', 1, 1, 1, 0, 0, '插件管理菜单', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'extension:plugin:list') AS parent), 'extension:plugin:query', '插件查询', '', '', 'plugin:query', '', 'B', 1, 1, 1, 0, 0, '插件查询按钮', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'extension:plugin:list') AS parent), 'extension:plugin:enable', '插件启用', '', '', 'plugin:enable', '', 'B', 2, 1, 1, 0, 0, '插件启用按钮', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'extension:plugin:list') AS parent), 'extension:plugin:disable', '插件禁用', '', '', 'plugin:disable', '', 'B', 3, 1, 1, 0, 0, '插件禁用按钮', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'extension:plugin:list') AS parent), 'extension:plugin:edit', '插件配置', '', '', 'plugin:edit', '', 'B', 4, 1, 1, 0, 0, '插件配置按钮', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'extension:plugin:list') AS parent), 'extension:plugin:install', '插件安装', '', '', 'plugin:install', '', 'B', 5, 1, 1, 0, 0, '插件安装按钮', NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "remark", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'extension:plugin:list') AS parent), 'extension:plugin:uninstall', '插件卸载', '', '', 'plugin:uninstall', '', 'B', 6, 1, 1, 0, 0, '插件卸载按钮', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ============================================================
-- Developer center menus
-- 开发中心菜单
-- ============================================================
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'developer') AS parent), 'about:api:list', '接口文档', '/about/api-docs', 'about/api-docs/index', 'about:api:list', 'lucide:file-code', 'M', 1, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;
INSERT INTO sys_menu ("parent_id", "menu_key", "name", "path", "component", "perms", "icon", "type", "sort", "visible", "status", "is_frame", "is_cache", "created_at", "updated_at")
VALUES ((SELECT parent."id" FROM (SELECT "id" FROM sys_menu WHERE "menu_key" = 'developer') AS parent), 'about:system:list', '版本信息', '/about/system-info', 'about/system-info/index', 'about:system:list', 'lucide:server', 'M', 2, 1, 1, 0, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ============================================================
-- Role authorization and administrator binding
-- 角色授权与管理员绑定
-- ============================================================
INSERT INTO sys_role_menu ("role_id", "menu_id", "tenant_id")
SELECT r."id", m."id", 0
FROM sys_role r
CROSS JOIN sys_menu m
WHERE r."key" = 'admin'
  AND r."tenant_id" = 0
  AND m."menu_key" NOT LIKE 'plugin:%'
ON CONFLICT DO NOTHING;

INSERT INTO sys_user_role ("user_id", "role_id", "tenant_id")
SELECT u."id", r."id", 0
FROM sys_user u
JOIN sys_role r ON r."key" = 'admin' AND r."tenant_id" = 0
WHERE u."username" = 'admin'
ON CONFLICT DO NOTHING;
