-- Mock data: 用户-部门-岗位关联数据
-- admin 用户关联 Lina科技 部门和总经理岗位
INSERT IGNORE INTO sys_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN sys_dept d ON d.code = 'lina'
WHERE u.username = 'admin';

INSERT IGNORE INTO sys_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN sys_post p ON p.code = 'CEO'
WHERE u.username = 'admin';

-- ============================================================
-- 用户-部门关联（约 60% 的用户分配部门，40% 不分配）
-- 部门分布: 研发=15人, 市场=10人, 测试=10人, 财务=8人, 运维=8人, Lina科技=5人
-- ============================================================
INSERT IGNORE INTO sys_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN sys_dept d ON d.code = 'dev'
WHERE u.username IN (
    'user001', 'user002', 'user003', 'user004', 'user005',
    'user006', 'user007', 'user008', 'user009', 'user010',
    'user011', 'user012', 'user013', 'user014', 'user015'
);

INSERT IGNORE INTO sys_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN sys_dept d ON d.code = 'market'
WHERE u.username IN (
    'user016', 'user017', 'user018', 'user019', 'user020',
    'user021', 'user022', 'user023', 'user024', 'user025'
);

INSERT IGNORE INTO sys_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN sys_dept d ON d.code = 'qa'
WHERE u.username IN (
    'user026', 'user027', 'user028', 'user029', 'user030',
    'user031', 'user032', 'user033', 'user034', 'user035'
);

INSERT IGNORE INTO sys_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN sys_dept d ON d.code = 'finance'
WHERE u.username IN (
    'user036', 'user037', 'user038', 'user039',
    'user040', 'user041', 'user042', 'user043'
);

INSERT IGNORE INTO sys_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN sys_dept d ON d.code = 'ops'
WHERE u.username IN (
    'user044', 'user045', 'user046', 'user047',
    'user048', 'user049', 'user050', 'user051'
);

INSERT IGNORE INTO sys_user_dept (user_id, dept_id)
SELECT u.id, d.id
FROM sys_user u
JOIN sys_dept d ON d.code = 'lina'
WHERE u.username IN ('user052', 'user053', 'user054', 'user055', 'user056');

-- user057~user100 → 未分配部门（共 44 个用户无部门关联）

-- ============================================================
-- 用户-岗位关联（约 40% 的用户分配岗位，60% 不分配）
-- ============================================================
INSERT IGNORE INTO sys_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN sys_post p ON p.code = 'CTO'
WHERE u.username IN ('user001');

INSERT IGNORE INTO sys_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN sys_post p ON p.code = 'PM'
WHERE u.username IN ('user002', 'user008');

INSERT IGNORE INTO sys_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN sys_post p ON p.code = 'DEV'
WHERE u.username IN ('user003', 'user004', 'user005', 'user006', 'user007', 'user009');

-- user010~user015 在研发部门但无岗位

INSERT IGNORE INTO sys_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN sys_post p ON p.code = 'QA'
WHERE u.username IN ('user026', 'user027', 'user028', 'user029', 'user030');

-- user031~user035 在测试部门但无岗位

INSERT IGNORE INTO sys_user_post (user_id, post_id)
SELECT u.id, p.id
FROM sys_user u
JOIN sys_post p ON p.code = 'CEO'
WHERE u.username IN ('user052', 'user053');

-- user054~user056 在 Lina科技 但无岗位
