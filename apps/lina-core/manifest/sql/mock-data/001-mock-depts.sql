-- Mock data: 部门演示数据
INSERT IGNORE INTO sys_dept (parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
SELECT 0, '0', 'Lina科技', 'lina', 0, 1, NOW(), NOW()
FROM DUAL
WHERE NOT EXISTS (
    SELECT 1
    FROM sys_dept
    WHERE code = 'lina'
);

INSERT IGNORE INTO sys_dept (parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
SELECT parent.id, CONCAT('0,', parent.id), '研发部门', 'dev', 1, 1, NOW(), NOW()
FROM sys_dept parent
WHERE parent.code = 'lina'
  AND NOT EXISTS (
      SELECT 1
      FROM sys_dept
      WHERE code = 'dev'
  );

INSERT IGNORE INTO sys_dept (parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
SELECT parent.id, CONCAT('0,', parent.id), '市场部门', 'market', 2, 1, NOW(), NOW()
FROM sys_dept parent
WHERE parent.code = 'lina'
  AND NOT EXISTS (
      SELECT 1
      FROM sys_dept
      WHERE code = 'market'
  );

INSERT IGNORE INTO sys_dept (parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
SELECT parent.id, CONCAT('0,', parent.id), '测试部门', 'qa', 3, 1, NOW(), NOW()
FROM sys_dept parent
WHERE parent.code = 'lina'
  AND NOT EXISTS (
      SELECT 1
      FROM sys_dept
      WHERE code = 'qa'
  );

INSERT IGNORE INTO sys_dept (parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
SELECT parent.id, CONCAT('0,', parent.id), '财务部门', 'finance', 4, 1, NOW(), NOW()
FROM sys_dept parent
WHERE parent.code = 'lina'
  AND NOT EXISTS (
      SELECT 1
      FROM sys_dept
      WHERE code = 'finance'
  );

INSERT IGNORE INTO sys_dept (parent_id, ancestors, name, code, order_num, status, created_at, updated_at)
SELECT parent.id, CONCAT('0,', parent.id), '运维部门', 'ops', 5, 1, NOW(), NOW()
FROM sys_dept parent
WHERE parent.code = 'lina'
  AND NOT EXISTS (
      SELECT 1
      FROM sys_dept
      WHERE code = 'ops'
  );
