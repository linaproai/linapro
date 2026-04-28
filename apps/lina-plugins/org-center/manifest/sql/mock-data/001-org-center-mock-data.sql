-- Mock data: organization departments, posts, and demo user bindings.

INSERT IGNORE INTO plugin_org_center_dept (parent_id, ancestors, name, code, order_num, leader, phone, email, status, remark, created_at, updated_at)
SELECT 0, '0', 'LinaPro.AI', 'linapro.ai', 0, admin.id, '021-55550000', 'office@example.com', 1, 'Mock organization root', NOW(), NOW()
FROM sys_user admin
WHERE admin.username = 'admin'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_dept
      WHERE code = 'linapro.ai'
  );

INSERT IGNORE INTO plugin_org_center_dept (parent_id, ancestors, name, code, order_num, leader, phone, email, status, remark, created_at, updated_at)
SELECT parent.id, CONCAT('0,', parent.id), '研发部门', 'dev', 1, COALESCE(leader.id, 0), '021-55550100', 'dev@example.com', 1, 'Mock research and development department', NOW(), NOW()
FROM plugin_org_center_dept parent
LEFT JOIN sys_user leader ON leader.username = 'user002'
WHERE parent.code = 'linapro.ai'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_dept
      WHERE code = 'dev'
  );

INSERT IGNORE INTO plugin_org_center_dept (parent_id, ancestors, name, code, order_num, leader, phone, email, status, remark, created_at, updated_at)
SELECT parent.id, CONCAT('0,', parent.id), '市场部门', 'market', 2, COALESCE(leader.id, 0), '021-55550200', 'market@example.com', 1, 'Mock marketing department', NOW(), NOW()
FROM plugin_org_center_dept parent
LEFT JOIN sys_user leader ON leader.username = 'user004'
WHERE parent.code = 'linapro.ai'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_dept
      WHERE code = 'market'
  );

INSERT IGNORE INTO plugin_org_center_dept (parent_id, ancestors, name, code, order_num, leader, phone, email, status, remark, created_at, updated_at)
SELECT parent.id, CONCAT('0,', parent.id), '测试部门', 'qa', 3, COALESCE(leader.id, 0), '021-55550300', 'qa@example.com', 1, 'Mock quality assurance department', NOW(), NOW()
FROM plugin_org_center_dept parent
LEFT JOIN sys_user leader ON leader.username = 'user008'
WHERE parent.code = 'linapro.ai'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_dept
      WHERE code = 'qa'
  );

INSERT IGNORE INTO plugin_org_center_dept (parent_id, ancestors, name, code, order_num, leader, phone, email, status, remark, created_at, updated_at)
SELECT parent.id, CONCAT('0,', parent.id), '财务部门', 'finance', 4, COALESCE(leader.id, 0), '021-55550400', 'finance@example.com', 1, 'Mock finance department', NOW(), NOW()
FROM plugin_org_center_dept parent
LEFT JOIN sys_user leader ON leader.username = 'user011'
WHERE parent.code = 'linapro.ai'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_dept
      WHERE code = 'finance'
  );

INSERT IGNORE INTO plugin_org_center_dept (parent_id, ancestors, name, code, order_num, leader, phone, email, status, remark, created_at, updated_at)
SELECT parent.id, CONCAT('0,', parent.id), '运维部门', 'ops', 5, COALESCE(leader.id, 0), '021-55550500', 'ops@example.com', 1, 'Mock operations department', NOW(), NOW()
FROM plugin_org_center_dept parent
LEFT JOIN sys_user leader ON leader.username = 'user009'
WHERE parent.code = 'linapro.ai'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_dept
      WHERE code = 'ops'
  );

INSERT IGNORE INTO plugin_org_center_dept (parent_id, ancestors, name, code, order_num, leader, phone, email, status, remark, created_at, updated_at)
SELECT parent.id, CONCAT('0,', parent.id), '归档部门', 'archive', 99, 0, '021-55550999', 'archive@example.com', 0, 'Disabled mock department for status filtering', NOW(), NOW()
FROM plugin_org_center_dept parent
WHERE parent.code = 'linapro.ai'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_dept
      WHERE code = 'archive'
  );

INSERT IGNORE INTO plugin_org_center_post (dept_id, code, name, sort, status, remark, created_at, updated_at)
SELECT d.id, 'CEO', '总经理', 1, 1, 'Mock executive post', NOW(), NOW()
FROM plugin_org_center_dept d
WHERE d.code = 'linapro.ai'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_post
      WHERE code = 'CEO'
  );

INSERT IGNORE INTO plugin_org_center_post (dept_id, code, name, sort, status, remark, created_at, updated_at)
SELECT d.id, 'CTO', '技术总监', 2, 1, 'Mock technology leader post', NOW(), NOW()
FROM plugin_org_center_dept d
WHERE d.code = 'dev'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_post
      WHERE code = 'CTO'
  );

INSERT IGNORE INTO plugin_org_center_post (dept_id, code, name, sort, status, remark, created_at, updated_at)
SELECT d.id, 'PM', '项目经理', 3, 1, 'Mock project manager post', NOW(), NOW()
FROM plugin_org_center_dept d
WHERE d.code = 'dev'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_post
      WHERE code = 'PM'
  );

INSERT IGNORE INTO plugin_org_center_post (dept_id, code, name, sort, status, remark, created_at, updated_at)
SELECT d.id, 'DEV', '开发工程师', 4, 1, 'Mock developer post', NOW(), NOW()
FROM plugin_org_center_dept d
WHERE d.code = 'dev'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_post
      WHERE code = 'DEV'
  );

INSERT IGNORE INTO plugin_org_center_post (dept_id, code, name, sort, status, remark, created_at, updated_at)
SELECT d.id, 'QA', '测试工程师', 5, 1, 'Mock quality engineer post', NOW(), NOW()
FROM plugin_org_center_dept d
WHERE d.code = 'qa'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_post
      WHERE code = 'QA'
  );

INSERT IGNORE INTO plugin_org_center_post (dept_id, code, name, sort, status, remark, created_at, updated_at)
SELECT d.id, 'OPS', '运维工程师', 6, 1, 'Mock operations engineer post', NOW(), NOW()
FROM plugin_org_center_dept d
WHERE d.code = 'ops'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_post
      WHERE code = 'OPS'
  );

INSERT IGNORE INTO plugin_org_center_post (dept_id, code, name, sort, status, remark, created_at, updated_at)
SELECT d.id, 'FIN', '财务专员', 7, 1, 'Mock finance specialist post', NOW(), NOW()
FROM plugin_org_center_dept d
WHERE d.code = 'finance'
  AND NOT EXISTS (
      SELECT 1
      FROM plugin_org_center_post
      WHERE code = 'FIN'
  );

INSERT IGNORE INTO plugin_org_center_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN plugin_org_center_dept d ON d.code = 'linapro.ai'
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
WHERE u.username IN ('user002', 'user014', 'user024', 'user036', 'user048', 'user060');

INSERT IGNORE INTO plugin_org_center_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN plugin_org_center_dept d ON d.code = 'market'
WHERE u.username IN ('user004', 'user020', 'user032', 'user044', 'user056', 'user068');

INSERT IGNORE INTO plugin_org_center_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN plugin_org_center_dept d ON d.code = 'qa'
WHERE u.username IN ('user008', 'user012', 'user018', 'user023', 'user029', 'user035');

INSERT IGNORE INTO plugin_org_center_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN plugin_org_center_dept d ON d.code = 'finance'
WHERE u.username IN ('user011', 'user026', 'user039', 'user051', 'user063', 'user075');

INSERT IGNORE INTO plugin_org_center_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN plugin_org_center_dept d ON d.code = 'ops'
WHERE u.username IN ('user009', 'user021', 'user033', 'user045', 'user057', 'user069');

INSERT IGNORE INTO plugin_org_center_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN plugin_org_center_post p ON p.code = 'DEV'
WHERE u.username IN ('user014', 'user024', 'user036', 'user048', 'user060');

INSERT IGNORE INTO plugin_org_center_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN plugin_org_center_post p ON p.code = 'PM'
WHERE u.username IN ('user002', 'user017', 'user030', 'user041', 'user053');

INSERT IGNORE INTO plugin_org_center_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN plugin_org_center_post p ON p.code = 'QA'
WHERE u.username IN ('user008', 'user012', 'user018', 'user023', 'user029', 'user035');

INSERT IGNORE INTO plugin_org_center_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN plugin_org_center_post p ON p.code = 'OPS'
WHERE u.username IN ('user009', 'user021', 'user033', 'user045', 'user057', 'user069');

INSERT IGNORE INTO plugin_org_center_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN plugin_org_center_post p ON p.code = 'FIN'
WHERE u.username IN ('user011', 'user026', 'user039', 'user051', 'user063', 'user075');
