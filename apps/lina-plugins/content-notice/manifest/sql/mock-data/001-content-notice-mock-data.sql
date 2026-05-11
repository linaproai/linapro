-- Mock data: notice and announcement records for content management demos.
-- 模拟数据：内容管理演示使用的通知与公告记录。

INSERT INTO plugin_content_notice ("tenant_id", "title", "type", "content", "status", "remark", "created_by", "updated_by", "created_at", "updated_at")
SELECT
    0,
    '系统升级通知',
    1,
    '<p>系统将于本周六凌晨2:00-4:00进行升级维护，届时系统将暂停服务。请提前做好相关工作安排。</p><p><strong>升级内容：</strong></p><ul><li>性能优化</li><li>安全补丁更新</li><li>新功能发布</li></ul>',
    1,
    'Mock published notice',
    admin."id",
    admin."id",
    '2026-04-20 09:00:00',
    '2026-04-20 09:00:00'
FROM sys_user admin
WHERE admin."username" = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_content_notice ("tenant_id", "title", "type", "content", "status", "remark", "created_by", "updated_by", "created_at", "updated_at")
SELECT
    0,
    '关于规范使用系统的公告',
    2,
    '<p>为保障系统安全稳定运行，请各位用户注意以下事项：</p><ol><li>请定期修改密码，密码长度不少于8位</li><li>不要将账号密码告知他人</li><li>离开工位时请锁定电脑屏幕</li></ol><p>感谢大家的配合！</p>',
    1,
    'Mock published announcement',
    admin."id",
    admin."id",
    '2026-04-21 10:30:00',
    '2026-04-21 10:30:00'
FROM sys_user admin
WHERE admin."username" = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_content_notice ("tenant_id", "title", "type", "content", "status", "remark", "created_by", "updated_by", "created_at", "updated_at")
SELECT
    0,
    '新功能上线预告',
    1,
    '<p>我们即将上线以下新功能：</p><ul><li>通知公告管理</li><li>消息中心</li><li>富文本编辑器</li></ul><p>敬请期待！</p>',
    0,
    'Mock draft notice',
    admin."id",
    admin."id",
    '2026-04-22 14:15:00',
    '2026-04-22 14:15:00'
FROM sys_user admin
WHERE admin."username" = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO plugin_content_notice ("tenant_id", "title", "type", "content", "status", "remark", "created_by", "updated_by", "created_at", "updated_at")
SELECT
    0,
    '安全巡检结果公告',
    2,
    '<p>本周安全巡检已完成，核心服务、插件运行时和数据库备份链路均正常。请各团队继续保持变更登记与权限最小化。</p>',
    1,
    'Mock governance announcement',
    admin."id",
    admin."id",
    '2026-04-23 16:45:00',
    '2026-04-23 16:45:00'
FROM sys_user admin
WHERE admin."username" = 'admin'
ON CONFLICT DO NOTHING;
