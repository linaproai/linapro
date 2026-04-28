-- 001: org-center schema

CREATE TABLE IF NOT EXISTS plugin_org_center_dept (
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
    deleted_at  DATETIME                         COMMENT '删除时间',
    UNIQUE KEY uk_plugin_org_center_dept_code ((NULLIF(code, ''))),
    KEY idx_plugin_org_center_dept_code (code),
    KEY idx_plugin_org_center_dept_parent_id (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='部门表';

CREATE TABLE IF NOT EXISTS plugin_org_center_post (
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
    UNIQUE KEY uk_plugin_org_center_post_code (code),
    KEY idx_plugin_org_center_post_dept_id (dept_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='岗位信息表';

CREATE TABLE IF NOT EXISTS plugin_org_center_user_dept (
    user_id INT NOT NULL COMMENT '用户ID',
    dept_id INT NOT NULL COMMENT '部门ID',
    PRIMARY KEY (user_id, dept_id),
    KEY idx_plugin_org_center_user_dept_dept_id (dept_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户与部门关联表';

CREATE TABLE IF NOT EXISTS plugin_org_center_user_post (
    user_id INT NOT NULL COMMENT '用户ID',
    post_id INT NOT NULL COMMENT '岗位ID',
    PRIMARY KEY (user_id, post_id),
    KEY idx_plugin_org_center_user_post_post_id (post_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户与岗位关联表';
