-- 002: Dict Management, Dept Management, Post Management, User-Dept-Post Association
-- 002：字典管理、部门管理、岗位管理、用户-部门-岗位关联

-- ============================================================
-- Dictionary type table
-- 字典类型表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_dict_type (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT  'Dictionary type ID',
    name        VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Dictionary name',
    type        VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Dictionary type',
    status      TINYINT      NOT NULL DEFAULT 1  COMMENT  'Status: 0=disabled, 1=enabled',
    is_builtin  TINYINT      NOT NULL DEFAULT 0  COMMENT  'Built-in record flag: 1=yes, 0=no',
    remark      VARCHAR(512) NOT NULL DEFAULT '' COMMENT  'Remark',
    created_at  DATETIME                         COMMENT  'Creation time',
    updated_at  DATETIME                         COMMENT  'Update time',
    deleted_at  DATETIME DEFAULT NULL            COMMENT  'Deletion time',
    UNIQUE KEY uk_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Dictionary type table';

-- ============================================================
-- Dictionary data table
-- 字典数据表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_dict_data (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT  'Dictionary data ID',
    dict_type   VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Dictionary type',
    label       VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Dictionary label',
    value       VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Dictionary value',
    sort        INT          NOT NULL DEFAULT 0  COMMENT  'Display order',
    tag_style   VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Tag style: primary/success/danger/warning, etc.',
    css_class   VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'CSS class name',
    status      TINYINT      NOT NULL DEFAULT 1  COMMENT  'Status: 0=disabled, 1=enabled',
    is_builtin  TINYINT      NOT NULL DEFAULT 0  COMMENT  'Built-in record flag: 1=yes, 0=no',
    remark      VARCHAR(512) NOT NULL DEFAULT '' COMMENT  'Remark',
    created_at  DATETIME                         COMMENT  'Creation time',
    updated_at  DATETIME                         COMMENT  'Update time',
    deleted_at  DATETIME DEFAULT NULL            COMMENT  'Deletion time',
    UNIQUE KEY uk_dict_type_value (dict_type, value)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Dictionary data table';

-- ============================================================
-- Dictionary seed data required by the host core
-- 字典初始化数据（宿主核心必需）
-- ============================================================

-- Dictionary type: status switch
-- 字典类型: 状态开关
INSERT IGNORE INTO sys_dict_type (name, type, status, is_builtin, remark, created_at, updated_at)
VALUES ('状态开关', 'sys_normal_disable', 1, 1, '状态开关列表', NOW(), NOW());

-- Dictionary type: user gender
-- 字典类型: 用户性别
INSERT IGNORE INTO sys_dict_type (name, type, status, is_builtin, remark, created_at, updated_at)
VALUES ('用户性别', 'sys_user_sex', 1, 1, '用户性别列表', NOW(), NOW());

-- Dictionary data: status switch
-- 字典数据: 状态开关
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, is_builtin, created_at, updated_at)
VALUES ('sys_normal_disable', '正常', '1', 1, 'primary', 1, 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, is_builtin, created_at, updated_at)
VALUES ('sys_normal_disable', '停用', '0', 2, 'danger', 1, 1, NOW(), NOW());

-- Dictionary data: user gender
-- 字典数据: 用户性别
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, is_builtin, created_at, updated_at)
VALUES ('sys_user_sex', '男', '1', 1, 'primary', 1, 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, is_builtin, created_at, updated_at)
VALUES ('sys_user_sex', '女', '2', 2, 'danger', 1, 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, is_builtin, created_at, updated_at)
VALUES ('sys_user_sex', '未知', '0', 3, 'default', 1, 1, NOW(), NOW());
