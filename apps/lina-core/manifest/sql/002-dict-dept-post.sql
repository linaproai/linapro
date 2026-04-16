-- 002: Dict Management, Dept Management, Post Management, User-Dept-Post Association

-- ============================================================
-- 字典类型表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_dict_type (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT '字典类型ID',
    name        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '字典名称',
    type        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '字典类型',
    status      TINYINT      NOT NULL DEFAULT 1  COMMENT '状态（0停用 1正常）',
    remark      VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
    created_at  DATETIME                         COMMENT '创建时间',
    updated_at  DATETIME                         COMMENT '更新时间'
    UNIQUE(type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='字典类型表';

-- ============================================================
-- 字典数据表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_dict_data (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT '字典数据ID',
    dict_type   VARCHAR(128) NOT NULL DEFAULT '' COMMENT '字典类型',
    label       VARCHAR(128) NOT NULL DEFAULT '' COMMENT '字典标签',
    value       VARCHAR(128) NOT NULL DEFAULT '' COMMENT '字典键值',
    sort        INT          NOT NULL DEFAULT 0  COMMENT '显示排序',
    tag_style   VARCHAR(128) NOT NULL DEFAULT '' COMMENT '标签样式（primary/success/danger/warning等）',
    css_class   VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'CSS样式类名',
    status      TINYINT      NOT NULL DEFAULT 1  COMMENT '状态（0停用 1正常）',
    remark      VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
    created_at  DATETIME                         COMMENT '创建时间',
    updated_at  DATETIME                         COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='字典数据表';

-- ============================================================
-- 部门表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_dept (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT '部门ID',
    parent_id   INT          NOT NULL DEFAULT 0  COMMENT '父部门ID',
    ancestors   VARCHAR(512) NOT NULL DEFAULT '' COMMENT '祖级列表',
    name        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '部门名称',
    code        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '部门编码',
    order_num   INT          NOT NULL DEFAULT 0  COMMENT '显示排序',
    leader      INT          NOT NULL DEFAULT 0  COMMENT '负责人用户ID',
    phone       VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '联系电话',
    email       VARCHAR(128) NOT NULL DEFAULT '' COMMENT '邮箱',
    status      TINYINT      NOT NULL DEFAULT 1  COMMENT '状态（0停用 1正常）',
    remark      VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
    created_at  DATETIME                         COMMENT '创建时间',
    updated_at  DATETIME                         COMMENT '更新时间',
    deleted_at  DATETIME                         COMMENT '删除时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='部门表';

-- ============================================================
-- 岗位表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_post (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT '岗位ID',
    dept_id     INT          NOT NULL DEFAULT 0  COMMENT '所属部门ID',
    code        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '岗位编码',
    name        VARCHAR(128) NOT NULL DEFAULT '' COMMENT '岗位名称',
    sort        INT          NOT NULL DEFAULT 0  COMMENT '显示排序',
    status      TINYINT      NOT NULL DEFAULT 1  COMMENT '状态（0停用 1正常）',
    remark      VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
    created_at  DATETIME                         COMMENT '创建时间',
    updated_at  DATETIME                         COMMENT '更新时间',
    deleted_at  DATETIME                         COMMENT '删除时间',
    UNIQUE(code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='岗位信息表';

-- ============================================================
-- 用户-部门关联表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_user_dept (
    user_id     INT NOT NULL COMMENT '用户ID',
    dept_id     INT NOT NULL COMMENT '部门ID',
    PRIMARY KEY (user_id, dept_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户与部门关联表';

-- ============================================================
-- 用户-岗位关联表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_user_post (
    user_id     INT NOT NULL COMMENT '用户ID',
    post_id     INT NOT NULL COMMENT '岗位ID',
    PRIMARY KEY (user_id, post_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户与岗位关联表';

-- ============================================================
-- 字典初始化数据（系统必需）
-- ============================================================

-- 字典类型: 状态开关
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('状态开关', 'sys_normal_disable', 1, '状态开关列表', NOW(), NOW());

-- 字典类型: 用户性别
INSERT IGNORE INTO sys_dict_type (name, type, status, remark, created_at, updated_at)
VALUES ('用户性别', 'sys_user_sex', 1, '用户性别列表', NOW(), NOW());

-- 字典数据: 状态开关
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_normal_disable', '正常', '1', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_normal_disable', '停用', '0', 2, 'danger', 1, NOW(), NOW());

-- 字典数据: 用户性别
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_user_sex', '男', '1', 1, 'primary', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_user_sex', '女', '2', 2, 'danger', 1, NOW(), NOW());
INSERT IGNORE INTO sys_dict_data (dict_type, label, value, sort, tag_style, status, created_at, updated_at)
VALUES ('sys_user_sex', '未知', '0', 3, 'default', 1, NOW(), NOW());
