-- Mock data: 部门演示数据
INSERT IGNORE INTO sys_dept (id, parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
VALUES (1, 0, '0', 'Lina科技', 'lina', 0, 1, NOW(), NOW());
INSERT IGNORE INTO sys_dept (id, parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
VALUES (2, 1, '0,1', '研发部门', 'dev', 1, 1, NOW(), NOW());
INSERT IGNORE INTO sys_dept (id, parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
VALUES (3, 1, '0,1', '市场部门', 'market', 2, 1, NOW(), NOW());
INSERT IGNORE INTO sys_dept (id, parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
VALUES (4, 1, '0,1', '测试部门', 'qa', 3, 1, NOW(), NOW());
INSERT IGNORE INTO sys_dept (id, parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
VALUES (5, 1, '0,1', '财务部门', 'finance', 4, 1, NOW(), NOW());
INSERT IGNORE INTO sys_dept (id, parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
VALUES (6, 1, '0,1', '运维部门', 'ops', 5, 1, NOW(), NOW());
