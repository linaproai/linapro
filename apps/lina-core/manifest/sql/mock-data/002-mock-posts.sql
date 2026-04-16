-- Mock data: 岗位演示数据
INSERT IGNORE INTO sys_post (id, dept_id, code, name, sort, status, created_at, updated_at)
VALUES (1, 1, 'CEO', '总经理', 1, 1, NOW(), NOW());
INSERT IGNORE INTO sys_post (id, dept_id, code, name, sort, status, created_at, updated_at)
VALUES (2, 2, 'CTO', '技术总监', 2, 1, NOW(), NOW());
INSERT IGNORE INTO sys_post (id, dept_id, code, name, sort, status, created_at, updated_at)
VALUES (3, 2, 'PM', '项目经理', 3, 1, NOW(), NOW());
INSERT IGNORE INTO sys_post (id, dept_id, code, name, sort, status, created_at, updated_at)
VALUES (4, 2, 'DEV', '开发工程师', 4, 1, NOW(), NOW());
INSERT IGNORE INTO sys_post (id, dept_id, code, name, sort, status, created_at, updated_at)
VALUES (5, 4, 'QA', '测试工程师', 5, 1, NOW(), NOW());
