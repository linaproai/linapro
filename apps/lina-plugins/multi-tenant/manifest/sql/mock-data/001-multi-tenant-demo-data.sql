-- Mock data: multi-tenant demo tenants, users, memberships, roles, and overrides.
-- 模拟数据：多租户演示租户、用户、成员、角色与覆盖配置。

-- Mock data: tenant directory rows for lifecycle, list, filtering, and
-- impersonation demos. The first three tenants carry scenario-specific active
-- and suspended states; the rest provide realistic list volume.
-- 模拟数据：租户主体目录，用于生命周期、列表、筛选和代操作演示。前三个租户
-- 覆盖活跃与暂停等特定场景，其余租户用于构造更真实的列表规模。
INSERT INTO plugin_multi_tenant_tenant ("code", "name", "status", "remark", "created_by", "updated_by", "created_at", "updated_at")
VALUES
    ('alpha-retail', '摸鱼科技有限公司', 'active', '活跃租户 mock 数据，用于租户成员、角色与代操作演示。', 0, 0, '2026-05-01 09:00:00', '2026-05-01 09:00:00'),
    ('beta-manufacturing', '精神股东科技有限公司', 'active', '活跃租户 mock 数据，用于跨租户成员与审计演示。', 0, 0, '2026-05-01 09:30:00', '2026-05-01 09:30:00'),
    ('gamma-sandbox', '打工人企业服务有限公司', 'suspended', '暂停租户 mock 数据，用于生命周期状态过滤演示。', 0, 0, '2026-05-01 10:00:00', '2026-05-01 10:00:00'),
    ('one-click-triple-media', '一键三连文化传媒有限公司', 'active', '租户管理列表 mock 数据。', 0, 0, '2026-05-01 10:30:00', '2026-05-01 10:30:00'),
    ('cyber-wellness-health', '赛博养生健康科技有限公司', 'active', '租户管理列表 mock 数据。', 0, 0, '2026-05-01 11:00:00', '2026-05-01 11:00:00'),
    ('clock-in-on-time-hr', '踩点到岗人力资源有限公司', 'active', '租户管理列表 mock 数据。', 0, 0, '2026-05-01 11:30:00', '2026-05-01 11:30:00'),
    ('crazy-thursday-catering', '疯狂星期四餐饮管理有限公司', 'active', '租户管理列表 mock 数据。', 0, 0, '2026-05-01 12:00:00', '2026-05-01 12:00:00'),
    ('deal-hunter-trading', '薅羊毛优选商贸有限公司', 'active', '租户管理列表 mock 数据。', 0, 0, '2026-05-01 12:30:00', '2026-05-01 12:30:00'),
    ('stay-calm-auto-service', '稳住别浪汽车服务有限公司', 'active', '租户管理列表 mock 数据。', 0, 0, '2026-05-01 13:00:00', '2026-05-01 13:00:00'),
    ('anti-involution-management', '不想内卷企业管理有限公司', 'active', '租户管理列表 mock 数据。', 0, 0, '2026-05-01 13:30:00', '2026-05-01 13:30:00'),
    ('drink-hot-water-health', '多喝热水健康管理有限公司', 'active', '租户管理列表 mock 数据。', 0, 0, '2026-05-01 14:00:00', '2026-05-01 14:00:00'),
    ('juejuezi-housekeeping', '绝绝子家政服务有限公司', 'active', '租户管理列表 mock 数据。', 0, 0, '2026-05-01 14:30:00', '2026-05-01 14:30:00'),
    ('eye-catching-brand-planning', '显眼包品牌策划有限公司', 'active', '租户管理列表 mock 数据。', 0, 0, '2026-05-01 15:00:00', '2026-05-01 15:00:00'),
    ('sudden-fortune-trading', '泼天富贵贸易有限公司', 'active', '租户管理列表 mock 数据。', 0, 0, '2026-05-01 15:30:00', '2026-05-01 15:30:00'),
    ('breakdown-repair-service', '破防维修服务有限公司', 'active', '租户管理列表 mock 数据。', 0, 0, '2026-05-01 16:00:00', '2026-05-01 16:00:00'),
    ('yep-yep-customer-service', '啊对对对客服外包有限公司', 'active', '租户管理列表 mock 数据。', 0, 0, '2026-05-01 16:30:00', '2026-05-01 16:30:00'),
    ('read-random-reply-service', '已读乱回客服外包有限公司', 'active', '租户管理列表 mock 数据。', 0, 0, '2026-05-01 17:00:00', '2026-05-01 17:00:00')
ON CONFLICT ("code") DO NOTHING;

-- Mock data: platform operator and auditor accounts for tenant lifecycle,
-- impersonation, and cross-tenant audit demos. These users stay in PLATFORM
-- tenant_id=0 so they can exercise platform control-plane permissions.
-- Demo login password for all platform mock users below: admin123.
-- 模拟数据：平台运营和审计账号，用于租户生命周期、代操作和跨租户审计演示。
-- 这些用户固定写入 PLATFORM tenant_id=0，以便验证平台控制面权限。
-- 以下所有平台 mock 用户的演示登录密码：admin123。
INSERT INTO sys_user ("tenant_id", "username", "password", "nickname", "email", "phone", "status", "remark", "created_at", "updated_at")
VALUES
    (0, 'platform_ops', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '平台 租户生命周期运营员', 'platform.ops@example.com', '13860000001', 1, 'Platform operator for multi-tenant demos', '2026-05-01 08:30:00', '2026-05-01 08:30:00'),
    (0, 'platform_auditor', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '平台 跨租户审计员', 'platform.auditor@example.com', '13860000002', 1, 'Platform auditor for multi-tenant demos', '2026-05-01 08:35:00', '2026-05-01 08:35:00')
ON CONFLICT ("username") DO NOTHING;

-- Mock data: tenant-scoped user accounts for login, tenant switching,
-- impersonation, and user-list membership demos. Each nickname starts with the
-- tenant name and ends with the account purpose so operators can identify the
-- related tenant and scenario from user tables. Every active demo-list tenant
-- has at least one active user so tenant filtering on the user list can be
-- demonstrated without empty result sets.
-- Demo login password for all tenant-scoped mock users below: admin123.
-- 模拟数据：租户范围用户账号，用于登录、租户切换、代操作和用户列表成员关系演示。
-- 每个 nickname 都以租户名称开头、以账号用途结尾，便于在用户表中直接识别
-- 所属租户和使用场景。每个活跃的列表演示租户至少有一个活跃用户,便于用户
-- 列表按租户筛选时不会出现空结果。
-- 以下所有租户范围 mock 用户的演示登录密码：admin123。
WITH v("tenant_code", "username", "password", "nickname", "email", "phone", "status", "remark", "created_at", "updated_at") AS (
    VALUES
        ('alpha-retail', 'tenant_alpha_admin', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '摸鱼科技 租户管理员', 'alpha.admin@example.com', '13860010001', 1, '摸鱼科技有限公司租户管理员', '2026-05-01 11:00:00', '2026-05-01 11:00:00'),
        ('alpha-retail', 'tenant_alpha_ops', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '摸鱼科技 运营用户', 'alpha.ops@example.com', '13860010002', 1, '摸鱼科技有限公司运营用户', '2026-05-01 11:05:00', '2026-05-01 11:05:00'),
        ('beta-manufacturing', 'tenant_beta_admin', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '精神股东 租户管理员', 'beta.admin@example.com', '13860020001', 1, '精神股东科技有限公司租户管理员', '2026-05-01 11:10:00', '2026-05-01 11:10:00'),
        ('beta-manufacturing', 'tenant_beta_auditor', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '精神股东 审计用户', 'beta.auditor@example.com', '13860020002', 1, '精神股东科技有限公司审计用户', '2026-05-01 11:15:00', '2026-05-01 11:15:00'),
        ('gamma-sandbox', 'tenant_gamma_admin', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '打工人 暂停租户管理员', 'gamma.admin@example.com', '13860030001', 0, '打工人企业服务有限公司暂停租户管理员', '2026-05-01 11:20:00', '2026-05-01 11:20:00'),
        ('one-click-triple-media', 'tenant_one_click_triple_media_user', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '一键三连 演示用户', 'one.click.triple.media.user@example.com', '13860040001', 1, '一键三连文化传媒有限公司用户列表演示用户', '2026-05-01 11:25:00', '2026-05-01 11:25:00'),
        ('cyber-wellness-health', 'tenant_cyber_wellness_health_user', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '赛博养生 演示用户', 'cyber.wellness.health.user@example.com', '13860050001', 1, '赛博养生健康科技有限公司用户列表演示用户', '2026-05-01 11:30:00', '2026-05-01 11:30:00'),
        ('clock-in-on-time-hr', 'tenant_clock_in_on_time_hr_user', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '踩点到岗 演示用户', 'clock.in.on.time.hr.user@example.com', '13860060001', 1, '踩点到岗人力资源有限公司用户列表演示用户', '2026-05-01 11:35:00', '2026-05-01 11:35:00'),
        ('crazy-thursday-catering', 'tenant_crazy_thursday_catering_user', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '疯狂星期四 演示用户', 'crazy.thursday.catering.user@example.com', '13860070001', 1, '疯狂星期四餐饮管理有限公司用户列表演示用户', '2026-05-01 11:40:00', '2026-05-01 11:40:00'),
        ('deal-hunter-trading', 'tenant_deal_hunter_trading_user', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '薅羊毛优选 演示用户', 'deal.hunter.trading.user@example.com', '13860080001', 1, '薅羊毛优选商贸有限公司用户列表演示用户', '2026-05-01 11:45:00', '2026-05-01 11:45:00'),
        ('stay-calm-auto-service', 'tenant_stay_calm_auto_service_user', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '稳住别浪 演示用户', 'stay.calm.auto.service.user@example.com', '13860090001', 1, '稳住别浪汽车服务有限公司用户列表演示用户', '2026-05-01 11:50:00', '2026-05-01 11:50:00'),
        ('anti-involution-management', 'tenant_anti_involution_management_user', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '不想内卷 演示用户', 'anti.involution.management.user@example.com', '13860100001', 1, '不想内卷企业管理有限公司用户列表演示用户', '2026-05-01 11:55:00', '2026-05-01 11:55:00'),
        ('drink-hot-water-health', 'tenant_drink_hot_water_health_user', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '多喝热水 演示用户', 'drink.hot.water.health.user@example.com', '13860110001', 1, '多喝热水健康管理有限公司用户列表演示用户', '2026-05-01 12:00:00', '2026-05-01 12:00:00'),
        ('juejuezi-housekeeping', 'tenant_juejuezi_housekeeping_user', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '绝绝子 演示用户', 'juejuezi.housekeeping.user@example.com', '13860120001', 1, '绝绝子家政服务有限公司用户列表演示用户', '2026-05-01 12:05:00', '2026-05-01 12:05:00'),
        ('eye-catching-brand-planning', 'tenant_eye_catching_brand_planning_user', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '显眼包 演示用户', 'eye.catching.brand.planning.user@example.com', '13860130001', 1, '显眼包品牌策划有限公司用户列表演示用户', '2026-05-01 12:10:00', '2026-05-01 12:10:00'),
        ('sudden-fortune-trading', 'tenant_sudden_fortune_trading_user', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '泼天富贵 演示用户', 'sudden.fortune.trading.user@example.com', '13860140001', 1, '泼天富贵贸易有限公司用户列表演示用户', '2026-05-01 12:15:00', '2026-05-01 12:15:00'),
        ('breakdown-repair-service', 'tenant_breakdown_repair_service_user', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '破防维修 演示用户', 'breakdown.repair.service.user@example.com', '13860150001', 1, '破防维修服务有限公司用户列表演示用户', '2026-05-01 12:20:00', '2026-05-01 12:20:00'),
        ('yep-yep-customer-service', 'tenant_yep_yep_customer_service_user', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '啊对对对 演示用户', 'yep.yep.customer.service.user@example.com', '13860160001', 1, '啊对对对客服外包有限公司用户列表演示用户', '2026-05-01 12:25:00', '2026-05-01 12:25:00'),
        ('read-random-reply-service', 'tenant_read_random_reply_service_user', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '已读乱回 演示用户', 'read.random.reply.service.user@example.com', '13860170001', 1, '已读乱回客服外包有限公司用户列表演示用户', '2026-05-01 12:30:00', '2026-05-01 12:30:00')
)
INSERT INTO sys_user ("tenant_id", "username", "password", "nickname", "email", "phone", "status", "remark", "created_at", "updated_at")
SELECT t."id", v."username", v."password", v."nickname", v."email", v."phone", v."status", v."remark", CAST(v."created_at" AS TIMESTAMP), CAST(v."updated_at" AS TIMESTAMP)
FROM v
JOIN plugin_multi_tenant_tenant t ON t."code" = v."tenant_code"
WHERE 1 = 1
ON CONFLICT ("username") DO NOTHING;

-- Mock data: user-to-tenant memberships for tenant switching, cross-tenant
-- visibility, and suspended-member scenarios. Several scenario users are
-- intentionally active in multiple tenants so list, switch, and permission
-- tests can verify realistic multi-tenant users.
-- 模拟数据：用户与租户成员关系，用于租户切换、跨租户可见性和暂停成员场景。
-- 多个场景用户特意同时加入多个租户，以便列表、切换和权限测试覆盖更真实的
-- 多租户用户。
WITH v("username", "tenant_code", "status", "joined_at") AS (
    VALUES
        ('tenant_alpha_admin', 'alpha-retail', 1, '2026-05-01 11:00:00'),
        ('tenant_alpha_ops', 'alpha-retail', 1, '2026-05-01 11:05:00'),
        ('tenant_beta_admin', 'beta-manufacturing', 1, '2026-05-01 11:10:00'),
        ('tenant_beta_auditor', 'beta-manufacturing', 1, '2026-05-01 11:15:00'),
        ('tenant_gamma_admin', 'gamma-sandbox', 0, '2026-05-01 11:20:00'),
        ('tenant_alpha_ops', 'beta-manufacturing', 1, '2026-05-03 09:00:00'),
        ('tenant_alpha_ops', 'one-click-triple-media', 1, '2026-05-03 09:10:00'),
        ('tenant_beta_auditor', 'alpha-retail', 1, '2026-05-03 09:20:00'),
        ('tenant_one_click_triple_media_user', 'one-click-triple-media', 1, '2026-05-01 11:25:00'),
        ('tenant_one_click_triple_media_user', 'alpha-retail', 1, '2026-05-03 09:30:00'),
        ('tenant_one_click_triple_media_user', 'cyber-wellness-health', 1, '2026-05-03 09:40:00'),
        ('tenant_cyber_wellness_health_user', 'cyber-wellness-health', 1, '2026-05-01 11:30:00'),
        ('tenant_cyber_wellness_health_user', 'beta-manufacturing', 1, '2026-05-03 09:50:00'),
        ('tenant_clock_in_on_time_hr_user', 'clock-in-on-time-hr', 1, '2026-05-01 11:35:00'),
        ('tenant_crazy_thursday_catering_user', 'crazy-thursday-catering', 1, '2026-05-01 11:40:00'),
        ('tenant_deal_hunter_trading_user', 'deal-hunter-trading', 1, '2026-05-01 11:45:00'),
        ('tenant_stay_calm_auto_service_user', 'stay-calm-auto-service', 1, '2026-05-01 11:50:00'),
        ('tenant_anti_involution_management_user', 'anti-involution-management', 1, '2026-05-01 11:55:00'),
        ('tenant_drink_hot_water_health_user', 'drink-hot-water-health', 1, '2026-05-01 12:00:00'),
        ('tenant_juejuezi_housekeeping_user', 'juejuezi-housekeeping', 1, '2026-05-01 12:05:00'),
        ('tenant_eye_catching_brand_planning_user', 'eye-catching-brand-planning', 1, '2026-05-01 12:10:00'),
        ('tenant_sudden_fortune_trading_user', 'sudden-fortune-trading', 1, '2026-05-01 12:15:00'),
        ('tenant_breakdown_repair_service_user', 'breakdown-repair-service', 1, '2026-05-01 12:20:00'),
        ('tenant_yep_yep_customer_service_user', 'yep-yep-customer-service', 1, '2026-05-01 12:25:00'),
        ('tenant_read_random_reply_service_user', 'read-random-reply-service', 1, '2026-05-01 12:30:00')
)
INSERT INTO plugin_multi_tenant_user_membership ("user_id", "tenant_id", "status", "joined_at", "created_by", "updated_by", "created_at", "updated_at")
SELECT u."id", t."id", v."status", CAST(v."joined_at" AS TIMESTAMP), 0, 0, NOW(), NOW()
FROM v
JOIN sys_user u ON u."username" = v."username"
JOIN plugin_multi_tenant_tenant t ON t."code" = v."tenant_code"
WHERE 1 = 1
ON CONFLICT ("user_id", "tenant_id") DO NOTHING;

-- Mock data: platform roles paired with the platform accounts above. The
-- operations role can manage tenants and impersonation; the auditor role is
-- read-only for tenant and user visibility checks.
-- 模拟数据：与上述平台账号配套的平台角色。运营角色用于租户管理和代操作；
-- 审计角色用于只读租户与用户可见性检查。
INSERT INTO sys_role ("tenant_id", "name", "key", "sort", "data_scope", "status", "remark", "created_at", "updated_at")
VALUES
    (0, '平台运营员', 'platform-ops', 30, 1, 1, 'Platform role for tenant lifecycle operations demos', NOW(), NOW()),
    (0, '平台租户审计员', 'platform-tenant-auditor', 31, 1, 1, 'Platform role for read-only tenant audit demos', NOW(), NOW())
ON CONFLICT ("tenant_id", "key") DO NOTHING;

-- Mock data: tenant-local roles for admin, operations, auditor, and suspended
-- tenant validation. Role data_scope values exercise tenant, department, user,
-- and inactive role paths.
-- 模拟数据：租户内角色，用于管理员、运营、审计和暂停租户校验。角色 data_scope
-- 覆盖本租户、本部门、本人和停用角色路径。
WITH v("tenant_code", "name", "key", "sort", "data_scope", "status", "remark") AS (
    VALUES
        ('alpha-retail', '摸鱼科技租户管理员', 'tenant-alpha-admin', 20, 2, 1, '摸鱼科技有限公司租户管理员角色'),
        ('alpha-retail', '摸鱼科技运营员', 'tenant-alpha-ops', 21, 3, 1, '摸鱼科技有限公司运营角色'),
        ('beta-manufacturing', '精神股东租户管理员', 'tenant-beta-admin', 20, 2, 1, '精神股东科技有限公司租户管理员角色'),
        ('beta-manufacturing', '精神股东审计员', 'tenant-beta-auditor', 22, 4, 1, '精神股东科技有限公司审计角色'),
        ('gamma-sandbox', '打工人暂停租户管理员', 'tenant-gamma-admin', 90, 2, 0, '打工人企业服务有限公司暂停租户角色'),
        ('one-click-triple-media', '一键三连演示用户', 'tenant-one-click-triple-media-user', 50, 4, 1, '一键三连文化传媒有限公司演示用户角色'),
        ('cyber-wellness-health', '赛博养生演示用户', 'tenant-cyber-wellness-health-user', 50, 4, 1, '赛博养生健康科技有限公司演示用户角色'),
        ('clock-in-on-time-hr', '踩点到岗演示用户', 'tenant-clock-in-on-time-hr-user', 50, 4, 1, '踩点到岗人力资源有限公司演示用户角色'),
        ('crazy-thursday-catering', '疯狂星期四演示用户', 'tenant-crazy-thursday-catering-user', 50, 4, 1, '疯狂星期四餐饮管理有限公司演示用户角色'),
        ('deal-hunter-trading', '薅羊毛优选演示用户', 'tenant-deal-hunter-trading-user', 50, 4, 1, '薅羊毛优选商贸有限公司演示用户角色'),
        ('stay-calm-auto-service', '稳住别浪演示用户', 'tenant-stay-calm-auto-service-user', 50, 4, 1, '稳住别浪汽车服务有限公司演示用户角色'),
        ('anti-involution-management', '不想内卷演示用户', 'tenant-anti-involution-management-user', 50, 4, 1, '不想内卷企业管理有限公司演示用户角色'),
        ('drink-hot-water-health', '多喝热水演示用户', 'tenant-drink-hot-water-health-user', 50, 4, 1, '多喝热水健康管理有限公司演示用户角色'),
        ('juejuezi-housekeeping', '绝绝子演示用户', 'tenant-juejuezi-housekeeping-user', 50, 4, 1, '绝绝子家政服务有限公司演示用户角色'),
        ('eye-catching-brand-planning', '显眼包演示用户', 'tenant-eye-catching-brand-planning-user', 50, 4, 1, '显眼包品牌策划有限公司演示用户角色'),
        ('sudden-fortune-trading', '泼天富贵演示用户', 'tenant-sudden-fortune-trading-user', 50, 4, 1, '泼天富贵贸易有限公司演示用户角色'),
        ('breakdown-repair-service', '破防维修演示用户', 'tenant-breakdown-repair-service-user', 50, 4, 1, '破防维修服务有限公司演示用户角色'),
        ('yep-yep-customer-service', '啊对对对演示用户', 'tenant-yep-yep-customer-service-user', 50, 4, 1, '啊对对对客服外包有限公司演示用户角色'),
        ('read-random-reply-service', '已读乱回演示用户', 'tenant-read-random-reply-service-user', 50, 4, 1, '已读乱回客服外包有限公司演示用户角色')
)
INSERT INTO sys_role ("tenant_id", "name", "key", "sort", "data_scope", "status", "remark", "created_at", "updated_at")
SELECT t."id", v."name", v."key", v."sort", v."data_scope", v."status", v."remark", NOW(), NOW()
FROM v
JOIN plugin_multi_tenant_tenant t ON t."code" = v."tenant_code"
WHERE 1 = 1
ON CONFLICT ("tenant_id", "key") DO NOTHING;

-- Mock data: full platform tenant-management permissions for the operations
-- role. This supports CRUD, impersonation, and user-list tenant membership demos.
-- 模拟数据：为平台运营角色授予完整租户管理权限，用于 CRUD、代操作和用户列表
-- 租户归属演示。
INSERT INTO sys_role_menu ("tenant_id", "role_id", "menu_id")
SELECT 0, r."id", m."id"
FROM sys_role r
JOIN sys_menu m ON m."perms" IN (
    'system:tenant:list',
    'system:tenant:query',
    'system:tenant:add',
    'system:tenant:edit',
    'system:tenant:remove',
    'system:tenant:impersonate',
    'system:user:list',
    'system:user:query'
)
WHERE r."tenant_id" = 0
  AND r."key" = 'platform-ops'
ON CONFLICT DO NOTHING;

-- Mock data: read-only tenant and user visibility permissions for the auditor
-- role. This role verifies cross-tenant audit screens without write actions.
-- 模拟数据：为平台审计角色授予只读租户与用户可见性权限，用于验证跨租户审计页面，
-- 不包含写操作权限。
INSERT INTO sys_role_menu ("tenant_id", "role_id", "menu_id")
SELECT 0, r."id", m."id"
FROM sys_role r
JOIN sys_menu m ON m."perms" IN (
    'system:tenant:list',
    'system:tenant:query',
    'system:user:list',
    'system:user:query'
)
WHERE r."tenant_id" = 0
  AND r."key" = 'platform-tenant-auditor'
ON CONFLICT DO NOTHING;

-- Mock data: grant tenant-admin roles access to tenant plugin and user
-- visibility menus for tenant administration demos.
-- 模拟数据：为租户管理员角色授予租户插件和用户可见性菜单权限，用于租户管理演示。
INSERT INTO sys_role_menu ("tenant_id", "role_id", "menu_id")
SELECT r."tenant_id", r."id", m."id"
FROM sys_role r
JOIN sys_menu m ON m."perms" IN (
    'system:tenant:plugin:list',
    'system:user:list',
    'system:user:query',
    'system:role:list',
    'system:role:query',
    'system:dict:list',
    'system:dict:query',
    'system:config:list',
    'system:config:query',
    'system:file:list',
    'system:file:query'
)
WHERE r."key" IN ('tenant-alpha-admin', 'tenant-beta-admin', 'tenant-gamma-admin')
ON CONFLICT DO NOTHING;

-- Mock data: grant operational and auditor roles read-oriented user
-- permissions plus several shared management menus so tenant data-scope and
-- permission differences are visible in demos. Demo-list tenant user roles also
-- get the same read-oriented permissions so each active mock tenant can open
-- several menus after login.
-- 模拟数据：为运营和审计角色授予偏只读的用户权限，使演示中可以看到不同租户
-- 数据权限的差异，并补充多个共享管理菜单权限。列表演示租户用户角色也授予
-- 同样偏只读的权限，以便每个活跃 mock 租户登录后都能打开多个菜单。
INSERT INTO sys_role_menu ("tenant_id", "role_id", "menu_id")
SELECT r."tenant_id", r."id", m."id"
FROM sys_role r
JOIN sys_menu m ON m."perms" IN (
    'system:user:list',
    'system:user:query',
    'system:role:list',
    'system:role:query',
    'system:dict:list',
    'system:dict:query',
    'system:config:list',
    'system:config:query',
    'system:file:list',
    'system:file:query'
)
WHERE r."key" IN (
    'tenant-alpha-ops',
    'tenant-beta-auditor',
    'tenant-one-click-triple-media-user',
    'tenant-cyber-wellness-health-user',
    'tenant-clock-in-on-time-hr-user',
    'tenant-crazy-thursday-catering-user',
    'tenant-deal-hunter-trading-user',
    'tenant-stay-calm-auto-service-user',
    'tenant-anti-involution-management-user',
    'tenant-drink-hot-water-health-user',
    'tenant-juejuezi-housekeeping-user',
    'tenant-eye-catching-brand-planning-user',
    'tenant-sudden-fortune-trading-user',
    'tenant-breakdown-repair-service-user',
    'tenant-yep-yep-customer-service-user',
    'tenant-read-random-reply-service-user'
)
ON CONFLICT DO NOTHING;

-- Mock data: bind tenant users to their scenario roles. These bindings drive
-- login, permission snapshot, and tenant data-scope demos.
-- 模拟数据：把租户用户绑定到对应场景角色，用于登录、权限快照和租户数据权限演示。
WITH v("username", "role_key") AS (
    VALUES
        ('tenant_alpha_admin', 'tenant-alpha-admin'),
        ('tenant_alpha_ops', 'tenant-alpha-ops'),
        ('tenant_alpha_ops', 'tenant-beta-auditor'),
        ('tenant_alpha_ops', 'tenant-one-click-triple-media-user'),
        ('tenant_beta_admin', 'tenant-beta-admin'),
        ('tenant_beta_auditor', 'tenant-beta-auditor'),
        ('tenant_beta_auditor', 'tenant-alpha-ops'),
        ('tenant_gamma_admin', 'tenant-gamma-admin'),
        ('tenant_one_click_triple_media_user', 'tenant-one-click-triple-media-user'),
        ('tenant_one_click_triple_media_user', 'tenant-alpha-ops'),
        ('tenant_one_click_triple_media_user', 'tenant-cyber-wellness-health-user'),
        ('tenant_cyber_wellness_health_user', 'tenant-cyber-wellness-health-user'),
        ('tenant_cyber_wellness_health_user', 'tenant-beta-auditor'),
        ('tenant_clock_in_on_time_hr_user', 'tenant-clock-in-on-time-hr-user'),
        ('tenant_crazy_thursday_catering_user', 'tenant-crazy-thursday-catering-user'),
        ('tenant_deal_hunter_trading_user', 'tenant-deal-hunter-trading-user'),
        ('tenant_stay_calm_auto_service_user', 'tenant-stay-calm-auto-service-user'),
        ('tenant_anti_involution_management_user', 'tenant-anti-involution-management-user'),
        ('tenant_drink_hot_water_health_user', 'tenant-drink-hot-water-health-user'),
        ('tenant_juejuezi_housekeeping_user', 'tenant-juejuezi-housekeeping-user'),
        ('tenant_eye_catching_brand_planning_user', 'tenant-eye-catching-brand-planning-user'),
        ('tenant_sudden_fortune_trading_user', 'tenant-sudden-fortune-trading-user'),
        ('tenant_breakdown_repair_service_user', 'tenant-breakdown-repair-service-user'),
        ('tenant_yep_yep_customer_service_user', 'tenant-yep-yep-customer-service-user'),
        ('tenant_read_random_reply_service_user', 'tenant-read-random-reply-service-user')
)
INSERT INTO sys_user_role ("tenant_id", "user_id", "role_id")
SELECT r."tenant_id", u."id", r."id"
FROM v
JOIN sys_user u ON u."username" = v."username"
JOIN sys_role r ON r."key" = v."role_key"
WHERE 1 = 1
ON CONFLICT DO NOTHING;

-- Mock data: bind the platform operations account to its lifecycle role.
-- 模拟数据：把平台运营账号绑定到租户生命周期管理角色。
INSERT INTO sys_user_role ("tenant_id", "user_id", "role_id")
SELECT 0, u."id", r."id"
FROM sys_user u
JOIN sys_role r ON r."tenant_id" = 0 AND r."key" = 'platform-ops'
WHERE u."username" = 'platform_ops'
ON CONFLICT DO NOTHING;

-- Mock data: bind the platform auditor account to its read-only audit role.
-- 模拟数据：把平台审计账号绑定到只读租户审计角色。
INSERT INTO sys_user_role ("tenant_id", "user_id", "role_id")
SELECT 0, u."id", r."id"
FROM sys_user u
JOIN sys_role r ON r."tenant_id" = 0 AND r."key" = 'platform-tenant-auditor'
WHERE u."username" = 'platform_auditor'
ON CONFLICT DO NOTHING;

-- Mock data: tenant-specific config overrides for fallback and tenant isolation
-- demos. Active tenants receive visible banner values; the suspended tenant
-- override is disabled to exercise filtering behavior.
-- 模拟数据：租户级配置覆盖，用于 fallback 和租户隔离演示。活跃租户提供可见
-- banner 值，暂停租户的覆盖配置保持禁用以验证过滤行为。
WITH v("tenant_code", "config_key", "config_value", "enabled") AS (
    VALUES
        ('alpha-retail', 'demo.notice.banner', '摸鱼科技有限公司工作台通知', TRUE),
        ('beta-manufacturing', 'demo.notice.banner', '精神股东科技有限公司工作台通知', TRUE),
        ('gamma-sandbox', 'demo.notice.banner', '打工人企业服务有限公司已暂停，用于生命周期演示', FALSE)
)
INSERT INTO plugin_multi_tenant_config_override ("tenant_id", "config_key", "config_value", "enabled", "created_at", "updated_at")
SELECT t."id", v."config_key", v."config_value", v."enabled", NOW(), NOW()
FROM v
JOIN plugin_multi_tenant_tenant t ON t."code" = v."tenant_code"
WHERE 1 = 1
ON CONFLICT ("tenant_id", "config_key") DO NOTHING;
