-- Mock data: 岗位演示数据
INSERT IGNORE INTO sys_post (dept_id, code, name, sort, status, created_at, updated_at)
SELECT d.id, 'CEO', '总经理', 1, 1, NOW(), NOW()
FROM sys_dept d
WHERE d.code = 'lina';

INSERT IGNORE INTO sys_post (dept_id, code, name, sort, status, created_at, updated_at)
SELECT d.id, 'CTO', '技术总监', 2, 1, NOW(), NOW()
FROM sys_dept d
WHERE d.code = 'dev';

INSERT IGNORE INTO sys_post (dept_id, code, name, sort, status, created_at, updated_at)
SELECT d.id, 'PM', '项目经理', 3, 1, NOW(), NOW()
FROM sys_dept d
WHERE d.code = 'dev';

INSERT IGNORE INTO sys_post (dept_id, code, name, sort, status, created_at, updated_at)
SELECT d.id, 'DEV', '开发工程师', 4, 1, NOW(), NOW()
FROM sys_dept d
WHERE d.code = 'dev';

INSERT IGNORE INTO sys_post (dept_id, code, name, sort, status, created_at, updated_at)
SELECT d.id, 'QA', '测试工程师', 5, 1, NOW(), NOW()
FROM sys_dept d
WHERE d.code = 'qa';
