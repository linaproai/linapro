-- Mock data: demo roles and role bindings for access-control screens.

INSERT IGNORE INTO sys_role (name, `key`, sort, data_scope, status, remark, created_at, updated_at)
VALUES ('审计员', 'auditor', 10, 1, 1, 'Mock role for read-only audit and monitoring demos', NOW(), NOW());

INSERT IGNORE INTO sys_role (name, `key`, sort, data_scope, status, remark, created_at, updated_at)
VALUES ('运维人员', 'operator', 11, 2, 1, 'Mock role for operations and scheduler demos', NOW(), NOW());

INSERT IGNORE INTO sys_role (name, `key`, sort, data_scope, status, remark, created_at, updated_at)
VALUES ('停用演示角色', 'disabled-demo', 99, 3, 0, 'Disabled mock role for status filtering demos', NOW(), NOW());

INSERT IGNORE INTO sys_role_menu (role_id, menu_id)
SELECT r.id, m.id
FROM sys_role r
JOIN sys_menu m ON m.menu_key IN (
    'dashboard',
    'dashboard:analytics:list',
    'dashboard:workspace:list',
    'monitor',
    'about:system:list',
    'about:api:list'
)
WHERE r.`key` = 'auditor';

INSERT IGNORE INTO sys_role_menu (role_id, menu_id)
SELECT r.id, m.id
FROM sys_role r
JOIN sys_menu m ON m.menu_key IN (
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
WHERE r.`key` = 'operator';

INSERT IGNORE INTO sys_user_role (user_id, role_id)
SELECT u.id, r.id
FROM sys_user u
JOIN sys_role r ON r.`key` = 'auditor'
WHERE u.username IN ('user002', 'user011', 'user026');

INSERT IGNORE INTO sys_user_role (user_id, role_id)
SELECT u.id, r.id
FROM sys_user u
JOIN sys_role r ON r.`key` = 'operator'
WHERE u.username IN ('user009', 'user021', 'user057', 'user069');

INSERT IGNORE INTO sys_user_role (user_id, role_id)
SELECT u.id, r.id
FROM sys_user u
JOIN sys_role r ON r.`key` = 'disabled-demo'
WHERE u.username = 'user073';
