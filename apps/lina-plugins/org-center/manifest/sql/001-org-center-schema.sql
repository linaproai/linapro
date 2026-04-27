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
    deleted_at  DATETIME                         COMMENT '删除时间'
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
    UNIQUE KEY uk_plugin_org_center_post_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='岗位信息表';

CREATE TABLE IF NOT EXISTS plugin_org_center_user_dept (
    user_id INT NOT NULL COMMENT '用户ID',
    dept_id INT NOT NULL COMMENT '部门ID',
    PRIMARY KEY (user_id, dept_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户与部门关联表';

CREATE TABLE IF NOT EXISTS plugin_org_center_user_post (
    user_id INT NOT NULL COMMENT '用户ID',
    post_id INT NOT NULL COMMENT '岗位ID',
    PRIMARY KEY (user_id, post_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户与岗位关联表';

INSERT IGNORE INTO plugin_org_center_dept (parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
SELECT 0, '0', 'LinaPro.AI', 'lina', 0, 1, NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_org_center_dept
    WHERE code = 'lina'
);

INSERT IGNORE INTO plugin_org_center_dept (parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
SELECT parent.id, CONCAT('0,', parent.id), '研发部门', 'dev', 1, 1, NOW(), NOW()
FROM plugin_org_center_dept parent
WHERE parent.code = 'lina'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_dept
      WHERE code = 'dev'
  );

INSERT IGNORE INTO plugin_org_center_dept (parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
SELECT parent.id, CONCAT('0,', parent.id), '市场部门', 'market', 2, 1, NOW(), NOW()
FROM plugin_org_center_dept parent
WHERE parent.code = 'lina'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_dept
      WHERE code = 'market'
  );

INSERT IGNORE INTO plugin_org_center_dept (parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
SELECT parent.id, CONCAT('0,', parent.id), '测试部门', 'qa', 3, 1, NOW(), NOW()
FROM plugin_org_center_dept parent
WHERE parent.code = 'lina'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_dept
      WHERE code = 'qa'
  );

INSERT IGNORE INTO plugin_org_center_dept (parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
SELECT parent.id, CONCAT('0,', parent.id), '财务部门', 'finance', 4, 1, NOW(), NOW()
FROM plugin_org_center_dept parent
WHERE parent.code = 'lina'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_dept
      WHERE code = 'finance'
  );

INSERT IGNORE INTO plugin_org_center_dept (parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
SELECT parent.id, CONCAT('0,', parent.id), '运维部门', 'ops', 5, 1, NOW(), NOW()
FROM plugin_org_center_dept parent
WHERE parent.code = 'lina'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_dept
      WHERE code = 'ops'
  );

INSERT IGNORE INTO plugin_org_center_post (dept_id, code, name, sort, status, created_at, updated_at)
SELECT d.id, 'CEO', '总经理', 1, 1, NOW(), NOW()
FROM plugin_org_center_dept d
WHERE d.code = 'lina'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_post
      WHERE code = 'CEO'
  );

INSERT IGNORE INTO plugin_org_center_post (dept_id, code, name, sort, status, created_at, updated_at)
SELECT d.id, 'CTO', '技术总监', 2, 1, NOW(), NOW()
FROM plugin_org_center_dept d
WHERE d.code = 'dev'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_post
      WHERE code = 'CTO'
  );

INSERT IGNORE INTO plugin_org_center_post (dept_id, code, name, sort, status, created_at, updated_at)
SELECT d.id, 'PM', '项目经理', 3, 1, NOW(), NOW()
FROM plugin_org_center_dept d
WHERE d.code = 'dev'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_post
      WHERE code = 'PM'
  );

INSERT IGNORE INTO plugin_org_center_post (dept_id, code, name, sort, status, created_at, updated_at)
SELECT d.id, 'DEV', '开发工程师', 4, 1, NOW(), NOW()
FROM plugin_org_center_dept d
WHERE d.code = 'dev'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_post
      WHERE code = 'DEV'
  );

INSERT IGNORE INTO plugin_org_center_post (dept_id, code, name, sort, status, created_at, updated_at)
SELECT d.id, 'QA', '测试工程师', 5, 1, NOW(), NOW()
FROM plugin_org_center_dept d
WHERE d.code = 'qa'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_post
      WHERE code = 'QA'
  );

INSERT IGNORE INTO plugin_org_center_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN plugin_org_center_dept d ON d.code = 'lina'
WHERE u.username = 'admin';

INSERT IGNORE INTO plugin_org_center_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN plugin_org_center_post p ON p.code = 'CEO'
WHERE u.username = 'admin';

INSERT IGNORE INTO plugin_org_center_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN plugin_org_center_dept d ON d.code = 'dev'
WHERE u.username IN (
    'user001', 'user002', 'user003', 'user004', 'user005',
    'user006', 'user007', 'user008', 'user009', 'user010',
    'user011', 'user012', 'user013', 'user014', 'user015'
);

INSERT IGNORE INTO plugin_org_center_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN plugin_org_center_dept d ON d.code = 'market'
WHERE u.username IN (
    'user016', 'user017', 'user018', 'user019', 'user020',
    'user021', 'user022', 'user023', 'user024', 'user025'
);

INSERT IGNORE INTO plugin_org_center_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN plugin_org_center_dept d ON d.code = 'qa'
WHERE u.username IN (
    'user026', 'user027', 'user028', 'user029', 'user030',
    'user031', 'user032', 'user033', 'user034', 'user035'
);

INSERT IGNORE INTO plugin_org_center_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN plugin_org_center_dept d ON d.code = 'finance'
WHERE u.username IN (
    'user036', 'user037', 'user038', 'user039',
    'user040', 'user041', 'user042', 'user043'
);

INSERT IGNORE INTO plugin_org_center_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN plugin_org_center_dept d ON d.code = 'ops'
WHERE u.username IN (
    'user044', 'user045', 'user046', 'user047',
    'user048', 'user049', 'user050', 'user051'
);

INSERT IGNORE INTO plugin_org_center_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN plugin_org_center_dept d ON d.code = 'lina'
WHERE u.username IN ('user052', 'user053', 'user054', 'user055', 'user056');

INSERT IGNORE INTO plugin_org_center_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN plugin_org_center_post p ON p.code = 'CTO'
WHERE u.username IN ('user001');

INSERT IGNORE INTO plugin_org_center_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN plugin_org_center_post p ON p.code = 'PM'
WHERE u.username IN ('user002', 'user008');

INSERT IGNORE INTO plugin_org_center_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN plugin_org_center_post p ON p.code = 'DEV'
WHERE u.username IN ('user003', 'user004', 'user005', 'user006', 'user007', 'user009');

INSERT IGNORE INTO plugin_org_center_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN plugin_org_center_post p ON p.code = 'QA'
WHERE u.username IN ('user026', 'user027', 'user028', 'user029', 'user030');

INSERT IGNORE INTO plugin_org_center_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN plugin_org_center_post p ON p.code = 'CEO'
WHERE u.username IN ('user052', 'user053');
