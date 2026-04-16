-- Mock data: 用户-部门-岗位关联数据
-- admin 用户关联 Lina科技 部门和总经理岗位
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (1, 1);
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (1, 1);

-- ============================================================
-- 用户-部门关联（约 60% 的用户分配部门，40% 不分配）
-- 部门分布: 研发(2)=15人, 市场(3)=10人, 测试(4)=10人, 财务(5)=8人, 运维(6)=8人, Lina科技(1)=5人
-- user001~user005 → 研发部门(2)
-- ============================================================
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (2, 2);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (3, 2);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (4, 2);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (5, 2);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (6, 2);

-- user006~user010 → 研发部门(2)
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (7, 2);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (8, 2);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (9, 2);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (10, 2);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (11, 2);

-- user011~user015 → 研发部门(2)
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (12, 2);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (13, 2);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (14, 2);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (15, 2);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (16, 2);

-- user016~user025 → 市场部门(3)
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (17, 3);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (18, 3);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (19, 3);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (20, 3);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (21, 3);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (22, 3);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (23, 3);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (24, 3);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (25, 3);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (26, 3);

-- user026~user035 → 测试部门(4)
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (27, 4);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (28, 4);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (29, 4);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (30, 4);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (31, 4);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (32, 4);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (33, 4);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (34, 4);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (35, 4);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (36, 4);

-- user036~user043 → 财务部门(5)
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (37, 5);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (38, 5);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (39, 5);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (40, 5);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (41, 5);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (42, 5);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (43, 5);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (44, 5);

-- user044~user051 → 运维部门(6)
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (45, 6);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (46, 6);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (47, 6);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (48, 6);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (49, 6);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (50, 6);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (51, 6);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (52, 6);

-- user052~user056 → Lina科技(1)
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (53, 1);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (54, 1);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (55, 1);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (56, 1);
INSERT IGNORE INTO sys_user_dept (user_id, dept_id) VALUES (57, 1);

-- user057~user100 (id 58~101) → 未分配部门（共 44 个用户无部门关联）

-- ============================================================
-- 用户-岗位关联（约 40% 的用户分配岗位，60% 不分配）
-- 仅部分有部门的用户分配岗位
-- ============================================================

-- 研发部门用户 → 技术总监/项目经理/开发工程师
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (2, 2);
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (3, 3);
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (4, 4);
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (5, 4);
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (6, 4);
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (7, 4);
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (8, 4);
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (9, 3);
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (10, 4);
-- user011~user015 在研发部门但无岗位

-- 测试部门用户 → 测试工程师
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (27, 5);
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (28, 5);
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (29, 5);
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (30, 5);
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (31, 5);
-- user032~user035 在测试部门但无岗位

-- Lina科技用户 → 总经理
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (53, 1);
INSERT IGNORE INTO sys_user_post (user_id, post_id) VALUES (54, 1);
-- user055~user056 在Lina科技但无岗位
