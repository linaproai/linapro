-- Mock data: demo roles and role bindings for access-control screens.
-- 模拟数据：访问控制页面使用的演示角色和角色绑定。

INSERT INTO sys_role ("tenant_id", "name", "key", "sort", "data_scope", "status", "remark", "created_at", "updated_at")
VALUES (0, '审计员', 'auditor', 10, 1, 1, 'Mock role for read-only audit and monitoring demos', NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_role ("tenant_id", "name", "key", "sort", "data_scope", "status", "remark", "created_at", "updated_at")
VALUES (0, '运维人员', 'operator', 11, 3, 1, 'Mock role for operations and scheduler demos', NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_role ("tenant_id", "name", "key", "sort", "data_scope", "status", "remark", "created_at", "updated_at")
VALUES (0, '停用演示角色', 'disabled-demo', 99, 4, 0, 'Disabled mock role for status filtering demos', NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO sys_role_menu ("tenant_id", "role_id", "menu_id")
SELECT 0, r."id", m."id"
FROM sys_role r
JOIN sys_menu m ON m."menu_key" IN (
    'dashboard',
    'dashboard:analytics:list',
    'dashboard:workspace:list',
    'monitor',
    'about:system:list',
    'about:api:list'
)
WHERE r."tenant_id" = 0
  AND r."key" = 'auditor'
ON CONFLICT DO NOTHING;

INSERT INTO sys_role_menu ("tenant_id", "role_id", "menu_id")
SELECT 0, r."id", m."id"
FROM sys_role r
JOIN sys_menu m ON m."menu_key" IN (
    'dashboard',
    'dashboard:workspace:list',
    'monitor',
    'scheduler',
    'system:job:list',
    'system:job:query',
    'system:jobgroup:list',
    'system:joblog:list',
    'system:file:list'
)
WHERE r."tenant_id" = 0
  AND r."key" = 'operator'
ON CONFLICT DO NOTHING;

INSERT INTO sys_user_role ("tenant_id", "user_id", "role_id")
SELECT 0, u."id", r."id"
FROM sys_user u
JOIN sys_role r ON r."tenant_id" = 0 AND r."key" = 'auditor'
WHERE u."username" IN ('user002', 'user011', 'user026')
ON CONFLICT DO NOTHING;

INSERT INTO sys_user_role ("tenant_id", "user_id", "role_id")
SELECT 0, u."id", r."id"
FROM sys_user u
JOIN sys_role r ON r."tenant_id" = 0 AND r."key" = 'operator'
WHERE u."username" IN ('user009', 'user021', 'user057', 'user069')
ON CONFLICT DO NOTHING;

INSERT INTO sys_user_role ("tenant_id", "user_id", "role_id")
SELECT 0, u."id", r."id"
FROM sys_user u
JOIN sys_role r ON r."tenant_id" = 0 AND r."key" = 'disabled-demo'
WHERE u."username" = 'user073'
ON CONFLICT DO NOTHING;
