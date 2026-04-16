-- 008: Menu Management, Role Management, User-Role Association

-- ============================================================
-- 菜单表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_menu (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT '菜单ID',
    parent_id   INT          NOT NULL DEFAULT 0  COMMENT '父菜单ID（0=根菜单）',
    menu_key    VARCHAR(128) NULL COMMENT '菜单稳定业务标识',
    name        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '菜单名称（支持i18n）',
    path        VARCHAR(255) NOT NULL DEFAULT '' COMMENT '路由地址',
    component   VARCHAR(255) NOT NULL DEFAULT '' COMMENT '组件路径',
    perms       VARCHAR(128) NOT NULL DEFAULT '' COMMENT '权限标识',
    icon        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '菜单图标',
    type        CHAR(1)      NOT NULL DEFAULT 'M' COMMENT '菜单类型（D=目录 M=菜单 B=按钮）',
    sort        INT          NOT NULL DEFAULT 0  COMMENT '显示排序',
    visible     TINYINT      NOT NULL DEFAULT 1  COMMENT '是否显示（1=显示 0=隐藏）',
    status      TINYINT      NOT NULL DEFAULT 1  COMMENT '状态（0=停用 1=正常）',
    is_frame    TINYINT      NOT NULL DEFAULT 0  COMMENT '是否外链（1=是 0=否）',
    is_cache    TINYINT      NOT NULL DEFAULT 0  COMMENT '是否缓存（1=是 0=否）',
    query_param VARCHAR(255) NOT NULL DEFAULT '' COMMENT '路由参数（JSON格式）',
    remark      VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
    created_at  DATETIME                         COMMENT '创建时间',
    updated_at  DATETIME                         COMMENT '更新时间',
    deleted_at  DATETIME                         COMMENT '删除时间',
    UNIQUE KEY uk_menu_key (menu_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='菜单权限表';

-- ============================================================
-- 角色表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_role (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT '角色ID',
    name        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '角色名称',
    `key`       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '权限字符',
    sort        INT          NOT NULL DEFAULT 0  COMMENT '显示排序',
    data_scope  TINYINT      NOT NULL DEFAULT 1  COMMENT '数据权限范围（1=全部 2=本部门 3=仅本人）',
    status      TINYINT      NOT NULL DEFAULT 1  COMMENT '状态（0=停用 1=正常）',
    remark      VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
    created_at  DATETIME                         COMMENT '创建时间',
    updated_at  DATETIME                         COMMENT '更新时间',
    deleted_at  DATETIME                         COMMENT '删除时间',
    UNIQUE(`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='角色信息表';

-- ============================================================
-- 角色-菜单关联表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_role_menu (
    role_id INT NOT NULL COMMENT '角色ID',
    menu_id INT NOT NULL COMMENT '菜单ID',
    PRIMARY KEY (role_id, menu_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='角色与菜单关联表';

-- ============================================================
-- 用户-角色关联表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_user_role (
    user_id INT NOT NULL COMMENT '用户ID',
    role_id INT NOT NULL COMMENT '角色ID',
    PRIMARY KEY (user_id, role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户与角色关联表';

-- ============================================================
-- 字典类型: 菜单状态
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('菜单状态', 'sys_menu_status', 1, '菜单状态列表', NOW(), NOW());

-- ============================================================
-- 字典类型: 显示状态
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('显示状态', 'sys_show_hide', 1, '显示状态列表', NOW(), NOW());

-- ============================================================
-- 字典类型: 菜单类型
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('菜单类型', 'sys_menu_type', 1, '菜单类型列表', NOW(), NOW());

-- ============================================================
-- 字典类型: 数据权限范围
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('数据权限范围', 'sys_data_scope', 1, '数据权限范围列表', NOW(), NOW());

-- ============================================================
-- 字典数据: 菜单状态
-- ============================================================
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_menu_status', '正常', '1', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_menu_status', '停用', '0', 2, 'danger', 1, NOW(), NOW());

-- ============================================================
-- 字典数据: 显示状态
-- ============================================================
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_show_hide', '显示', '1', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_show_hide', '隐藏', '0', 2, 'danger', 1, NOW(), NOW());

-- ============================================================
-- 字典数据: 菜单类型
-- ============================================================
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_menu_type', '目录', 'D', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_menu_type', '菜单', 'M', 2, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_menu_type', '按钮', 'B', 3, 'warning', 1, NOW(), NOW());

-- ============================================================
-- 字典数据: 数据权限范围
-- ============================================================
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_data_scope', '全部数据', '1', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_data_scope', '本部门数据', '2', 2, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_data_scope', '仅本人数据', '3', 3, 'warning', 1, NOW(), NOW());

-- ============================================================
-- 初始化角色数据
-- ============================================================
INSERT IGNORE INTO sys_role (name, `key`, sort, data_scope, status, remark, created_at, updated_at)
VALUES ('超级管理员', 'admin', 1, 1, 1, '超级管理员，拥有所有权限', NOW(), NOW());
INSERT IGNORE INTO sys_role (name, `key`, sort, data_scope, status, remark, created_at, updated_at)
VALUES ('普通用户', 'user', 2, 3, 1, '普通用户，仅查看本人数据', NOW(), NOW());

-- ============================================================
-- 初始化菜单数据
-- ============================================================

-- 清理旧菜单数据（包括角色-菜单关联）
DELETE FROM sys_role_menu;
DELETE FROM sys_menu WHERE id > 0;
ALTER TABLE sys_menu AUTO_INCREMENT = 1;

-- ========================================
-- 仪表盘（目录）
-- ========================================
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (1, 0, 'dashboard', '仪表盘', 'dashboard', '', '', 'ant-design:dashboard-outlined', 'D', 0, 1, 1, 0, 0, NOW(), NOW());

-- 仪表盘 -> 分析页（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (100, 1, 'dashboard:analytics:list', '分析页', 'analytics', 'dashboard/analytics/index', 'dashboard:analytics:list', 'ant-design:area-chart-outlined', 'M', 1, 1, 1, 0, 0, NOW(), NOW());

-- 仪表盘 -> 工作台（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (101, 1, 'dashboard:workspace:list', '工作台', 'workspace', 'dashboard/workspace/index', 'dashboard:workspace:list', 'ant-design:desktop-outlined', 'M', 2, 1, 1, 0, 0, NOW(), NOW());

-- ========================================
-- 系统管理（目录）
-- ========================================
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2, 0, 'system', '系统管理', 'system', '', '', 'ant-design:setting-outlined', 'D', 1, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 用户管理（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (200, 2, 'system:user:list', '用户管理', 'user', 'system/user/index', 'system:user:list', 'ant-design:user-outlined', 'M', 1, 1, 1, 0, 0, NOW(), NOW());

-- 用户管理 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2001, 200, 'system:user:query', '用户查询', '', '', 'system:user:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2002, 200, 'system:user:add', '用户新增', '', '', 'system:user:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2003, 200, 'system:user:edit', '用户修改', '', '', 'system:user:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2004, 200, 'system:user:remove', '用户删除', '', '', 'system:user:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2005, 200, 'system:user:export', '用户导出', '', '', 'system:user:export', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2006, 200, 'system:user:import', '用户导入', '', '', 'system:user:import', '', 'B', 6, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2007, 200, 'system:user:resetPwd', '重置密码', '', '', 'system:user:resetPwd', '', 'B', 7, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 角色管理（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (201, 2, 'system:role:list', '角色管理', 'role', 'system/role/index', 'system:role:list', 'ant-design:team-outlined', 'M', 2, 1, 1, 0, 0, NOW(), NOW());

-- 角色管理 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2011, 201, 'system:role:query', '角色查询', '', '', 'system:role:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2012, 201, 'system:role:add', '角色新增', '', '', 'system:role:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2013, 201, 'system:role:edit', '角色修改', '', '', 'system:role:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2014, 201, 'system:role:remove', '角色删除', '', '', 'system:role:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 菜单管理（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (202, 2, 'system:menu:list', '菜单管理', 'menu', 'system/menu/index', 'system:menu:list', 'ant-design:menu-outlined', 'M', 3, 1, 1, 0, 0, NOW(), NOW());

-- 菜单管理 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2021, 202, 'system:menu:query', '菜单查询', '', '', 'system:menu:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2022, 202, 'system:menu:add', '菜单新增', '', '', 'system:menu:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2023, 202, 'system:menu:edit', '菜单修改', '', '', 'system:menu:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2024, 202, 'system:menu:remove', '菜单删除', '', '', 'system:menu:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 部门管理（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (203, 2, 'system:dept:list', '部门管理', 'dept', 'system/dept/index', 'system:dept:list', 'ant-design:apartment-outlined', 'M', 4, 1, 1, 0, 0, NOW(), NOW());

-- 部门管理 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2031, 203, 'system:dept:query', '部门查询', '', '', 'system:dept:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2032, 203, 'system:dept:add', '部门新增', '', '', 'system:dept:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2033, 203, 'system:dept:edit', '部门修改', '', '', 'system:dept:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2034, 203, 'system:dept:remove', '部门删除', '', '', 'system:dept:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 岗位管理（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (204, 2, 'system:post:list', '岗位管理', 'post', 'system/post/index', 'system:post:list', 'ant-design:cluster-outlined', 'M', 5, 1, 1, 0, 0, NOW(), NOW());

-- 岗位管理 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2041, 204, 'system:post:query', '岗位查询', '', '', 'system:post:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2042, 204, 'system:post:add', '岗位新增', '', '', 'system:post:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2043, 204, 'system:post:edit', '岗位修改', '', '', 'system:post:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2044, 204, 'system:post:remove', '岗位删除', '', '', 'system:post:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2045, 204, 'system:post:export', '岗位导出', '', '', 'system:post:export', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 字典管理（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (205, 2, 'system:dict:list', '字典管理', 'dict', 'system/dict/index', 'system:dict:list', 'ant-design:book-outlined', 'M', 6, 1, 1, 0, 0, NOW(), NOW());

-- 字典管理 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2051, 205, 'system:dict:query', '字典查询', '', '', 'system:dict:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2052, 205, 'system:dict:add', '字典新增', '', '', 'system:dict:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2053, 205, 'system:dict:edit', '字典修改', '', '', 'system:dict:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2054, 205, 'system:dict:remove', '字典删除', '', '', 'system:dict:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2055, 205, 'system:dict:export', '字典导出', '', '', 'system:dict:export', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 通知公告（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (206, 2, 'system:notice:list', '通知公告', 'notice', 'system/notice/index', 'system:notice:list', 'ant-design:notification-outlined', 'M', 7, 1, 1, 0, 0, NOW(), NOW());

-- 通知公告 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2061, 206, 'system:notice:query', '公告查询', '', '', 'system:notice:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2062, 206, 'system:notice:add', '公告新增', '', '', 'system:notice:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2063, 206, 'system:notice:edit', '公告修改', '', '', 'system:notice:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2064, 206, 'system:notice:remove', '公告删除', '', '', 'system:notice:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 参数设置（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (207, 2, 'system:config:list', '参数设置', 'config', 'system/config/index', 'system:config:list', 'ant-design:tool-outlined', 'M', 8, 1, 1, 0, 0, NOW(), NOW());

-- 参数设置 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2071, 207, 'system:config:query', '参数查询', '', '', 'system:config:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2072, 207, 'system:config:add', '参数新增', '', '', 'system:config:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2073, 207, 'system:config:edit', '参数修改', '', '', 'system:config:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2074, 207, 'system:config:remove', '参数删除', '', '', 'system:config:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2075, 207, 'system:config:export', '参数导出', '', '', 'system:config:export', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 文件管理（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (208, 2, 'system:file:list', '文件管理', 'file', 'system/file/index', 'system:file:list', 'ant-design:folder-outlined', 'M', 9, 1, 1, 0, 0, NOW(), NOW());

-- 文件管理 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2081, 208, 'system:file:query', '文件查询', '', '', 'system:file:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2082, 208, 'system:file:upload', '文件上传', '', '', 'system:file:upload', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2083, 208, 'system:file:download', '文件下载', '', '', 'system:file:download', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (2084, 208, 'system:file:remove', '文件删除', '', '', 'system:file:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 消息列表（隐藏菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (209, 2, 'system:message:list', '消息列表', 'message', 'system/message/index', 'system:message:list', 'ant-design:message-outlined', 'M', 10, 0, 1, 0, 0, NOW(), NOW());

-- 系统管理 -> 角色授权用户（隐藏菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (210, 2, 'system:role:auth', '角色授权用户', 'role-auth/user', 'system/role-auth/index', 'system:role:auth', 'ant-design:usergroup-add-outlined', 'M', 11, 0, 1, 0, 0, NOW(), NOW());

-- ========================================
-- 系统监控（目录）
-- ========================================
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (3, 0, 'monitor', '系统监控', 'monitor', '', '', 'ant-design:monitor-outlined', 'D', 2, 1, 1, 0, 0, NOW(), NOW());

-- 系统监控 -> 在线用户（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (300, 3, 'monitor:online:list', '在线用户', 'online', 'monitor/online/index', 'monitor:online:list', 'ant-design:user-outlined', 'M', 1, 1, 1, 0, 0, NOW(), NOW());

-- 在线用户 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (3001, 300, 'monitor:online:query', '在线查询', '', '', 'monitor:online:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (3002, 300, 'monitor:online:forceLogout', '强制退出', '', '', 'monitor:online:forceLogout', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());

-- 系统监控 -> 服务监控（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (301, 3, 'monitor:server:list', '服务监控', 'server', 'monitor/server/index', 'monitor:server:list', 'ant-design:desktop-outlined', 'M', 2, 1, 1, 0, 0, NOW(), NOW());

-- 系统监控 -> 操作日志（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (302, 3, 'monitor:operlog:list', '操作日志', 'operlog', 'monitor/operlog/index', 'monitor:operlog:list', 'ant-design:form-outlined', 'M', 3, 1, 1, 0, 0, NOW(), NOW());

-- 操作日志 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (3021, 302, 'monitor:operlog:query', '日志查询', '', '', 'monitor:operlog:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (3022, 302, 'monitor:operlog:remove', '日志删除', '', '', 'monitor:operlog:remove', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (3023, 302, 'monitor:operlog:export', '日志导出', '', '', 'monitor:operlog:export', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (3024, 302, 'monitor:operlog:clear', '清空日志', '', '', 'monitor:operlog:clear', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

-- 系统监控 -> 登录日志（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (303, 3, 'monitor:loginlog:list', '登录日志', 'loginlog', 'monitor/loginlog/index', 'monitor:loginlog:list', 'ant-design:login-outlined', 'M', 4, 1, 1, 0, 0, NOW(), NOW());

-- 登录日志 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (3031, 303, 'monitor:loginlog:query', '日志查询', '', '', 'monitor:loginlog:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (3032, 303, 'monitor:loginlog:remove', '日志删除', '', '', 'monitor:loginlog:remove', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (3033, 303, 'monitor:loginlog:export', '日志导出', '', '', 'monitor:loginlog:export', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (3034, 303, 'monitor:loginlog:clear', '清空日志', '', '', 'monitor:loginlog:clear', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

-- ========================================
-- 系统信息（目录）
-- ========================================
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (4, 0, 'about', '系统信息', 'about', '', '', 'ant-design:info-circle-outlined', 'D', 3, 1, 1, 0, 0, NOW(), NOW());

-- 系统信息 -> 系统接口（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (400, 4, 'about:api:list', '系统接口', 'api-docs', 'about/api-docs/index', 'about:api:list', 'ant-design:file-text-outlined', 'M', 1, 1, 1, 0, 0, NOW(), NOW());

-- 系统信息 -> 版本信息（菜单）
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (401, 4, 'about:system:list', '版本信息', 'system-info', 'about/system-info/index', 'about:system:list', 'ant-design:desktop-outlined', 'M', 2, 1, 1, 0, 0, NOW(), NOW());

-- ========================================
-- 系统管理 -> 插件管理（菜单）
-- ========================================
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, query_param, remark, created_at, updated_at)
VALUES (900, 2, 'system:plugin:list', '插件管理', 'plugin', 'system/plugin/index', 'plugin:list', 'ant-design:appstore-outlined', 'M', 10, 1, 1, 0, 0, '', '插件管理菜单', NOW(), NOW());

-- 插件管理 -> 按钮权限
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, query_param, remark, created_at, updated_at)
VALUES (9011, 900, 'system:plugin:query', '插件查询', '', '', 'plugin:query', '', 'B', 1, 1, 1, 0, 0, '', '插件查询按钮', NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, query_param, remark, created_at, updated_at)
VALUES (9012, 900, 'system:plugin:enable', '插件启用', '', '', 'plugin:enable', '', 'B', 2, 1, 1, 0, 0, '', '插件启用按钮', NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, query_param, remark, created_at, updated_at)
VALUES (9013, 900, 'system:plugin:disable', '插件禁用', '', '', 'plugin:disable', '', 'B', 3, 1, 1, 0, 0, '', '插件禁用按钮', NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, query_param, remark, created_at, updated_at)
VALUES (9014, 900, 'system:plugin:install', '插件安装', '', '', 'plugin:install', '', 'B', 4, 1, 1, 0, 0, '', '插件安装按钮', NOW(), NOW());
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, query_param, remark, created_at, updated_at)
VALUES (9015, 900, 'system:plugin:uninstall', '插件卸载', '', '', 'plugin:uninstall', '', 'B', 5, 1, 1, 0, 0, '', '插件卸载按钮', NOW(), NOW());

-- ========================================
-- 个人中心（隐藏菜单，不属于任何目录）
-- ========================================
INSERT IGNORE INTO sys_menu (id, parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES (500, 0, 'profile:view', '个人中心', 'profile', '_core/profile/index', 'profile:view', 'ant-design:user-outlined', 'M', 99, 0, 1, 0, 0, NOW(), NOW());

INSERT IGNORE INTO sys_user_role (user_id, role_id) VALUES (1, 1);

-- ============================================================
-- 为 admin 角色分配所有菜单权限
-- ============================================================
INSERT IGNORE INTO sys_role_menu (role_id, menu_id)
SELECT 1, id FROM sys_menu;
