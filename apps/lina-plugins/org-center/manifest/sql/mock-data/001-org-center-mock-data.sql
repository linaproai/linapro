-- Mock data: organization departments, posts, and demo user bindings.
-- 模拟数据：组织部门、岗位和演示用户绑定。

INSERT INTO plugin_org_center_dept ("tenant_id", "parent_id", "ancestors", "name", "code", "order_num", "leader", "phone", "email", "status", "remark", "created_at", "updated_at")
SELECT 0, 0, '0', 'LinaPro.AI', 'linapro.ai', 0, admin."id", '021-55550000', 'office@example.com', 1, 'Mock organization root', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM sys_user admin
WHERE admin."username" = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_dept ("tenant_id", "parent_id", "ancestors", "name", "code", "order_num", "leader", "phone", "email", "status", "remark", "created_at", "updated_at")
SELECT 0, parent."id", '0,' || parent."id", '研发部门', 'dev', 1, COALESCE("leader"."id", 0), '021-55550100', 'dev@example.com', 1, 'Mock research and development department', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM plugin_org_center_dept parent
LEFT JOIN sys_user "leader" ON "leader"."username" = 'user002'
WHERE parent."tenant_id" = 0
  AND parent."code" = 'linapro.ai'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_dept ("tenant_id", "parent_id", "ancestors", "name", "code", "order_num", "leader", "phone", "email", "status", "remark", "created_at", "updated_at")
SELECT 0, parent."id", '0,' || parent."id", '市场部门', 'market', 2, COALESCE("leader"."id", 0), '021-55550200', 'market@example.com', 1, 'Mock marketing department', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM plugin_org_center_dept parent
LEFT JOIN sys_user "leader" ON "leader"."username" = 'user004'
WHERE parent."tenant_id" = 0
  AND parent."code" = 'linapro.ai'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_dept ("tenant_id", "parent_id", "ancestors", "name", "code", "order_num", "leader", "phone", "email", "status", "remark", "created_at", "updated_at")
SELECT 0, parent."id", '0,' || parent."id", '测试部门', 'qa', 3, COALESCE("leader"."id", 0), '021-55550300', 'qa@example.com', 1, 'Mock quality assurance department', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM plugin_org_center_dept parent
LEFT JOIN sys_user "leader" ON "leader"."username" = 'user008'
WHERE parent."tenant_id" = 0
  AND parent."code" = 'linapro.ai'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_dept ("tenant_id", "parent_id", "ancestors", "name", "code", "order_num", "leader", "phone", "email", "status", "remark", "created_at", "updated_at")
SELECT 0, parent."id", '0,' || parent."id", '财务部门', 'finance', 4, COALESCE("leader"."id", 0), '021-55550400', 'finance@example.com', 1, 'Mock finance department', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM plugin_org_center_dept parent
LEFT JOIN sys_user "leader" ON "leader"."username" = 'user011'
WHERE parent."tenant_id" = 0
  AND parent."code" = 'linapro.ai'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_dept ("tenant_id", "parent_id", "ancestors", "name", "code", "order_num", "leader", "phone", "email", "status", "remark", "created_at", "updated_at")
SELECT 0, parent."id", '0,' || parent."id", '运维部门', 'ops', 5, COALESCE("leader"."id", 0), '021-55550500', 'ops@example.com', 1, 'Mock operations department', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM plugin_org_center_dept parent
LEFT JOIN sys_user "leader" ON "leader"."username" = 'user009'
WHERE parent."tenant_id" = 0
  AND parent."code" = 'linapro.ai'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_dept ("tenant_id", "parent_id", "ancestors", "name", "code", "order_num", "leader", "phone", "email", "status", "remark", "created_at", "updated_at")
SELECT 0, parent."id", '0,' || parent."id", '归档部门', 'archive', 99, 0, '021-55550999', 'archive@example.com', 0, 'Disabled mock department for status filtering', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM plugin_org_center_dept parent
WHERE parent."tenant_id" = 0
  AND parent."code" = 'linapro.ai'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_post ("tenant_id", "dept_id", "code", "name", "sort", "status", "remark", "created_at", "updated_at")
SELECT 0, d."id", 'CEO', '总经理', 1, 1, 'Mock executive post', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM plugin_org_center_dept d
WHERE d."tenant_id" = 0
  AND d."code" = 'linapro.ai'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_post ("tenant_id", "dept_id", "code", "name", "sort", "status", "remark", "created_at", "updated_at")
SELECT 0, d."id", 'CTO', '技术总监', 2, 1, 'Mock technology leader post', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM plugin_org_center_dept d
WHERE d."tenant_id" = 0
  AND d."code" = 'dev'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_post ("tenant_id", "dept_id", "code", "name", "sort", "status", "remark", "created_at", "updated_at")
SELECT 0, d."id", 'PM', '项目经理', 3, 1, 'Mock project manager post', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM plugin_org_center_dept d
WHERE d."tenant_id" = 0
  AND d."code" = 'dev'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_post ("tenant_id", "dept_id", "code", "name", "sort", "status", "remark", "created_at", "updated_at")
SELECT 0, d."id", 'DEV', '开发工程师', 4, 1, 'Mock developer post', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM plugin_org_center_dept d
WHERE d."tenant_id" = 0
  AND d."code" = 'dev'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_post ("tenant_id", "dept_id", "code", "name", "sort", "status", "remark", "created_at", "updated_at")
SELECT 0, d."id", 'QA', '测试工程师', 5, 1, 'Mock quality engineer post', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM plugin_org_center_dept d
WHERE d."tenant_id" = 0
  AND d."code" = 'qa'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_post ("tenant_id", "dept_id", "code", "name", "sort", "status", "remark", "created_at", "updated_at")
SELECT 0, d."id", 'OPS', '运维工程师', 6, 1, 'Mock operations engineer post', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM plugin_org_center_dept d
WHERE d."tenant_id" = 0
  AND d."code" = 'ops'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_post ("tenant_id", "dept_id", "code", "name", "sort", "status", "remark", "created_at", "updated_at")
SELECT 0, d."id", 'FIN', '财务专员', 7, 1, 'Mock finance specialist post', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM plugin_org_center_dept d
WHERE d."tenant_id" = 0
  AND d."code" = 'finance'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_user_dept ("tenant_id", "user_id", "dept_id")
SELECT 0, u."id", d."id"
FROM sys_user u
JOIN plugin_org_center_dept d ON d."tenant_id" = 0 AND d."code" = 'linapro.ai'
WHERE u."username" = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_user_post ("tenant_id", "user_id", "post_id")
SELECT 0, u."id", p."id"
FROM sys_user u
JOIN plugin_org_center_post p ON p."tenant_id" = 0 AND p."code" = 'CEO'
WHERE u."username" = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_user_dept ("tenant_id", "user_id", "dept_id")
SELECT 0, u."id", d."id"
FROM sys_user u
JOIN plugin_org_center_dept d ON d."tenant_id" = 0 AND d."code" = 'dev'
WHERE u."username" IN ('user002', 'user014', 'user024', 'user036', 'user048', 'user060')
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_user_dept ("tenant_id", "user_id", "dept_id")
SELECT 0, u."id", d."id"
FROM sys_user u
JOIN plugin_org_center_dept d ON d."tenant_id" = 0 AND d."code" = 'market'
WHERE u."username" IN ('user004', 'user020', 'user032', 'user044', 'user056', 'user068')
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_user_dept ("tenant_id", "user_id", "dept_id")
SELECT 0, u."id", d."id"
FROM sys_user u
JOIN plugin_org_center_dept d ON d."tenant_id" = 0 AND d."code" = 'qa'
WHERE u."username" IN ('user008', 'user012', 'user018', 'user023', 'user029', 'user035')
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_user_dept ("tenant_id", "user_id", "dept_id")
SELECT 0, u."id", d."id"
FROM sys_user u
JOIN plugin_org_center_dept d ON d."tenant_id" = 0 AND d."code" = 'finance'
WHERE u."username" IN ('user011', 'user026', 'user039', 'user051', 'user063', 'user075')
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_user_dept ("tenant_id", "user_id", "dept_id")
SELECT 0, u."id", d."id"
FROM sys_user u
JOIN plugin_org_center_dept d ON d."tenant_id" = 0 AND d."code" = 'ops'
WHERE u."username" IN ('user009', 'user021', 'user033', 'user045', 'user057', 'user069')
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_user_post ("tenant_id", "user_id", "post_id")
SELECT 0, u."id", p."id"
FROM sys_user u
JOIN plugin_org_center_post p ON p."tenant_id" = 0 AND p."code" = 'DEV'
WHERE u."username" IN ('user014', 'user024', 'user036', 'user048', 'user060')
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_user_post ("tenant_id", "user_id", "post_id")
SELECT 0, u."id", p."id"
FROM sys_user u
JOIN plugin_org_center_post p ON p."tenant_id" = 0 AND p."code" = 'PM'
WHERE u."username" IN ('user002', 'user017', 'user030', 'user041', 'user053')
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_user_post ("tenant_id", "user_id", "post_id")
SELECT 0, u."id", p."id"
FROM sys_user u
JOIN plugin_org_center_post p ON p."tenant_id" = 0 AND p."code" = 'QA'
WHERE u."username" IN ('user008', 'user012', 'user018', 'user023', 'user029', 'user035')
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_user_post ("tenant_id", "user_id", "post_id")
SELECT 0, u."id", p."id"
FROM sys_user u
JOIN plugin_org_center_post p ON p."tenant_id" = 0 AND p."code" = 'OPS'
WHERE u."username" IN ('user009', 'user021', 'user033', 'user045', 'user057', 'user069')
ON CONFLICT DO NOTHING;

INSERT INTO plugin_org_center_user_post ("tenant_id", "user_id", "post_id")
SELECT 0, u."id", p."id"
FROM sys_user u
JOIN plugin_org_center_post p ON p."tenant_id" = 0 AND p."code" = 'FIN'
WHERE u."username" IN ('user011', 'user026', 'user039', 'user051', 'user063', 'user075')
ON CONFLICT DO NOTHING;
