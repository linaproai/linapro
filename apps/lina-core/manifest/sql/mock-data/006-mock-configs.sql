-- Mock data: config parameter examples.
-- 模拟数据：参数设置示例数据。

INSERT INTO sys_config ("name", "key", "value", "remark", "created_at", "updated_at") VALUES
('演示-支持邮箱', 'demo.support.email', 'support@example.com', '仅用于演示自定义参数能力，不被宿主运行时直接消费。', NOW(), NOW()),
('演示-首页公告文案', 'demo.notice.banner', '欢迎使用 LinaPro', '仅用于演示自定义参数能力，不被宿主运行时直接消费。', NOW(), NOW())
ON CONFLICT DO NOTHING;
