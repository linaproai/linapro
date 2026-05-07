-- 001: org-center schema
-- 001：org-center 数据结构

CREATE TABLE IF NOT EXISTS plugin_org_center_dept (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT  'Department ID',
    parent_id   INT          NOT NULL DEFAULT 0  COMMENT  'Parent department ID',
    ancestors   VARCHAR(512) NOT NULL DEFAULT '' COMMENT  'Ancestor list',
    name        VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Department name',
    code        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT  'Department code',
    order_num   INT          NOT NULL DEFAULT 0  COMMENT  'Display order',
    leader      INT          NOT NULL DEFAULT 0  COMMENT  'Leader user ID',
    phone       VARCHAR(20)  NOT NULL DEFAULT '' COMMENT  'Contact phone number',
    email       VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Email address',
    status      TINYINT      NOT NULL DEFAULT 1  COMMENT  'Status: 0=disabled, 1=enabled',
    remark      VARCHAR(512) NOT NULL DEFAULT '' COMMENT  'Remark',
    created_at  DATETIME                         COMMENT  'Creation time',
    updated_at  DATETIME                         COMMENT  'Update time',
    deleted_at  DATETIME                         COMMENT  'Deletion time',
    UNIQUE KEY uk_plugin_org_center_dept_code ((NULLIF(code, ''))),
    KEY idx_plugin_org_center_dept_code (code),
    KEY idx_plugin_org_center_dept_parent_id (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Department table';

CREATE TABLE IF NOT EXISTS plugin_org_center_post (
    id          INT PRIMARY KEY AUTO_INCREMENT COMMENT  'Post ID',
    dept_id     INT          NOT NULL DEFAULT 0  COMMENT  'Owning department ID',
    code        VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Post code',
    name        VARCHAR(128) NOT NULL DEFAULT '' COMMENT  'Post name',
    sort        INT          NOT NULL DEFAULT 0  COMMENT  'Display order',
    status      TINYINT      NOT NULL DEFAULT 1  COMMENT  'Status: 0=disabled, 1=enabled',
    remark      VARCHAR(512) NOT NULL DEFAULT '' COMMENT  'Remark',
    created_at  DATETIME                         COMMENT  'Creation time',
    updated_at  DATETIME                         COMMENT  'Update time',
    deleted_at  DATETIME                         COMMENT  'Deletion time',
    UNIQUE KEY uk_plugin_org_center_post_code (code),
    KEY idx_plugin_org_center_post_dept_id (dept_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'Post information table';

CREATE TABLE IF NOT EXISTS plugin_org_center_user_dept (
    user_id INT NOT NULL COMMENT  'User ID',
    dept_id INT NOT NULL COMMENT  'Department ID',
    PRIMARY KEY (user_id, dept_id),
    KEY idx_plugin_org_center_user_dept_dept_user (dept_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'User-department relation table';

CREATE TABLE IF NOT EXISTS plugin_org_center_user_post (
    user_id INT NOT NULL COMMENT  'User ID',
    post_id INT NOT NULL COMMENT  'Post ID',
    PRIMARY KEY (user_id, post_id),
    KEY idx_plugin_org_center_user_post_post_id (post_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT= 'User-post relation table';
