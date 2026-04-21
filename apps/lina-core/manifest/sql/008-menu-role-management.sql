-- 008: Menu Management, Role Management, Core Navigation Skeleton

-- ============================================================
-- 重建菜单与角色治理表
-- 说明：当前项目处于全新开源阶段，初始化时直接重建菜单/角色治理表，
-- 以确保新的宿主目录骨架和权限结构能够完全覆盖旧初始化数据。
-- ============================================================
DROP TABLE IF EXISTS sys_role_menu;
DROP TABLE IF EXISTS sys_user_role;
DROP TABLE IF EXISTS sys_role;
DROP TABLE IF EXISTS sys_menu;

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
-- 字典类型与字典数据
-- ============================================================
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('菜单状态', 'sys_menu_status', 1, '菜单状态列表', NOW(), NOW());
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('显示状态', 'sys_show_hide', 1, '显示状态列表', NOW(), NOW());
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('菜单类型', 'sys_menu_type', 1, '菜单类型列表', NOW(), NOW());
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('数据权限范围', 'sys_data_scope', 1, '数据权限范围列表', NOW(), NOW());

INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_menu_status', '正常', '1', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_menu_status', '停用', '0', 2, 'danger', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_show_hide', '显示', '1', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_show_hide', '隐藏', '0', 2, 'danger', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_menu_type', '目录', 'D', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_menu_type', '菜单', 'M', 2, 'success', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_menu_type', '按钮', 'B', 3, 'warning', 1, NOW(), NOW());
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
-- 宿主稳定一级目录骨架
-- ============================================================
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, remark, created_at, updated_at)
VALUES (0, 'dashboard', '工作台', 'dashboard', '', '', 'lucide:layout-dashboard', 'D', 1, 1, 1, 0, 0, '宿主稳定目录：工作台', NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, remark, created_at, updated_at)
VALUES (0, 'iam', '权限管理', 'iam', '', '', 'lucide:shield-check', 'D', 2, 1, 1, 0, 0, '宿主稳定目录：权限管理', NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, remark, created_at, updated_at)
VALUES (0, 'org', '组织管理', 'org', '', '', 'lucide:network', 'D', 3, 1, 1, 0, 0, '宿主稳定目录：组织管理', NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, remark, created_at, updated_at)
VALUES (0, 'setting', '系统设置', 'setting', '', '', 'lucide:settings-2', 'D', 4, 1, 1, 0, 0, '宿主稳定目录：系统设置', NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, remark, created_at, updated_at)
VALUES (0, 'content', '内容管理', 'content', '', '', 'lucide:newspaper', 'D', 5, 1, 1, 0, 0, '宿主稳定目录：内容管理', NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, remark, created_at, updated_at)
VALUES (0, 'monitor', '系统监控', 'monitor', '', '', 'lucide:activity', 'D', 6, 1, 1, 0, 0, '宿主稳定目录：系统监控', NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, remark, created_at, updated_at)
VALUES (0, 'scheduler', '任务调度', 'scheduler', '', '', 'lucide:calendar-range', 'D', 7, 1, 1, 0, 0, '宿主稳定目录：任务调度', NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, remark, created_at, updated_at)
VALUES (0, 'extension', '扩展中心', 'extension', '', '', 'lucide:puzzle', 'D', 8, 1, 1, 0, 0, '宿主稳定目录：扩展中心', NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, remark, created_at, updated_at)
VALUES (0, 'developer', '开发中心', 'developer', '', '', 'lucide:flask-conical', 'D', 9, 1, 1, 0, 0, '宿主稳定目录：开发中心', NOW(), NOW());

-- ============================================================
-- 工作台菜单
-- ============================================================
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'dashboard') AS parent), 'dashboard:analytics:list', '分析页', '/analytics', 'dashboard/analytics/index', 'dashboard:analytics:list', 'lucide:area-chart', 'M', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'dashboard') AS parent), 'dashboard:workspace:list', '工作台', '/workspace', 'dashboard/workspace/index', 'dashboard:workspace:list', 'carbon:workspace', 'M', 2, 1, 1, 0, 0, NOW(), NOW());

-- ============================================================
-- 权限管理菜单
-- ============================================================
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'iam') AS parent), 'system:user:list', '用户管理', '/system/user', 'system/user/index', 'system:user:list', 'ant-design:user-outlined', 'M', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:user:list') AS parent), 'system:user:query', '用户查询', '', '', 'system:user:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:user:list') AS parent), 'system:user:add', '用户新增', '', '', 'system:user:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:user:list') AS parent), 'system:user:edit', '用户修改', '', '', 'system:user:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:user:list') AS parent), 'system:user:remove', '用户删除', '', '', 'system:user:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:user:list') AS parent), 'system:user:export', '用户导出', '', '', 'system:user:export', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:user:list') AS parent), 'system:user:import', '用户导入', '', '', 'system:user:import', '', 'B', 6, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:user:list') AS parent), 'system:user:resetPwd', '重置密码', '', '', 'system:user:resetPwd', '', 'B', 7, 1, 1, 0, 0, NOW(), NOW());

INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'iam') AS parent), 'system:role:list', '角色管理', '/system/role', 'system/role/index', 'system:role:list', 'lucide:shield', 'M', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:role:list') AS parent), 'system:role:query', '角色查询', '', '', 'system:role:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:role:list') AS parent), 'system:role:add', '角色新增', '', '', 'system:role:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:role:list') AS parent), 'system:role:edit', '角色修改', '', '', 'system:role:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:role:list') AS parent), 'system:role:remove', '角色删除', '', '', 'system:role:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:role:list') AS parent), 'system:role:auth', '角色授权', '', '', 'system:role:auth', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW());

INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'iam') AS parent), 'system:menu:list', '菜单管理', '/system/menu', 'system/menu/index', 'system:menu:list', 'lucide:menu', 'M', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:menu:list') AS parent), 'system:menu:query', '菜单查询', '', '', 'system:menu:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:menu:list') AS parent), 'system:menu:add', '菜单新增', '', '', 'system:menu:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:menu:list') AS parent), 'system:menu:edit', '菜单修改', '', '', 'system:menu:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:menu:list') AS parent), 'system:menu:remove', '菜单删除', '', '', 'system:menu:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

-- ============================================================
-- 系统设置菜单
-- ============================================================
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'setting') AS parent), 'system:dict:list', '字典管理', '/system/dict', 'system/dict/index', 'system:dict:list', 'lucide:book-open', 'M', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:dict:list') AS parent), 'system:dict:query', '字典查询', '', '', 'system:dict:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:dict:list') AS parent), 'system:dict:add', '字典新增', '', '', 'system:dict:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:dict:list') AS parent), 'system:dict:edit', '字典修改', '', '', 'system:dict:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:dict:list') AS parent), 'system:dict:remove', '字典删除', '', '', 'system:dict:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:dict:list') AS parent), 'system:dict:export', '字典导出', '', '', 'system:dict:export', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW());

INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'setting') AS parent), 'system:config:list', '参数设置', '/system/config', 'system/config/index', 'system:config:list', 'lucide:sliders-horizontal', 'M', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:config:list') AS parent), 'system:config:query', '参数查询', '', '', 'system:config:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:config:list') AS parent), 'system:config:add', '参数新增', '', '', 'system:config:add', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:config:list') AS parent), 'system:config:edit', '参数修改', '', '', 'system:config:edit', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:config:list') AS parent), 'system:config:remove', '参数删除', '', '', 'system:config:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:config:list') AS parent), 'system:config:export', '参数导出', '', '', 'system:config:export', '', 'B', 5, 1, 1, 0, 0, NOW(), NOW());

INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'setting') AS parent), 'system:file:list', '文件管理', '/system/file', 'system/file/index', 'system:file:list', 'lucide:folder-open', 'M', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:file:list') AS parent), 'system:file:query', '文件查询', '', '', 'system:file:query', '', 'B', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:file:list') AS parent), 'system:file:upload', '文件上传', '', '', 'system:file:upload', '', 'B', 2, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:file:list') AS parent), 'system:file:download', '文件下载', '', '', 'system:file:download', '', 'B', 3, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'system:file:list') AS parent), 'system:file:remove', '文件删除', '', '', 'system:file:remove', '', 'B', 4, 1, 1, 0, 0, NOW(), NOW());

-- ============================================================
-- 扩展中心菜单
-- ============================================================
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, remark, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'extension') AS parent), 'extension:plugin:list', '插件管理', '/system/plugin', 'system/plugin/index', 'plugin:list', 'lucide:plug', 'M', 1, 1, 1, 0, 0, '插件管理菜单', NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, remark, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'extension:plugin:list') AS parent), 'extension:plugin:query', '', '', '', 'plugin:query', '', 'B', 1, 1, 1, 0, 0, '插件查询按钮', NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, remark, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'extension:plugin:list') AS parent), 'extension:plugin:enable', '', '', '', 'plugin:enable', '', 'B', 2, 1, 1, 0, 0, '插件启用按钮', NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, remark, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'extension:plugin:list') AS parent), 'extension:plugin:disable', '', '', '', 'plugin:disable', '', 'B', 3, 1, 1, 0, 0, '插件禁用按钮', NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, remark, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'extension:plugin:list') AS parent), 'extension:plugin:install', '', '', '', 'plugin:install', '', 'B', 4, 1, 1, 0, 0, '插件安装按钮', NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, remark, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'extension:plugin:list') AS parent), 'extension:plugin:uninstall', '', '', '', 'plugin:uninstall', '', 'B', 5, 1, 1, 0, 0, '插件卸载按钮', NOW(), NOW());

-- ============================================================
-- 开发中心菜单
-- ============================================================
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'developer') AS parent), 'about:api:list', '接口文档', '/about/api-docs', 'about/api-docs/index', 'about:api:list', 'lucide:file-code', 'M', 1, 1, 1, 0, 0, NOW(), NOW());
INSERT IGNORE INTO sys_menu (parent_id, menu_key, name, path, component, perms, icon, type, sort, visible, status, is_frame, is_cache, created_at, updated_at)
VALUES ((SELECT parent.id FROM (SELECT id FROM sys_menu WHERE menu_key = 'developer') AS parent), 'about:system:list', '系统信息', '/about/system-info', 'about/system-info/index', 'about:system:list', 'lucide:server', 'M', 2, 1, 1, 0, 0, NOW(), NOW());

-- ============================================================
-- 角色授权与管理员绑定
-- ============================================================
INSERT IGNORE INTO sys_role_menu (role_id, menu_id)
SELECT r.id, m.id
FROM sys_role r
JOIN sys_menu m
WHERE r.`key` = 'admin'
  AND m.menu_key NOT LIKE 'plugin:%';

INSERT IGNORE INTO sys_user_role (user_id, role_id)
SELECT u.id, r.id
FROM sys_user u
JOIN sys_role r ON r.`key` = 'admin'
WHERE u.username = 'admin';
