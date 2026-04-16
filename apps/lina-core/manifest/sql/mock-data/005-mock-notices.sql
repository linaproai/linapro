-- Mock data: 通知公告演示数据

INSERT IGNORE INTO sys_notice (id, title, type, content, status, remark, created_by, updated_by, created_at, updated_at)
VALUES (1, '系统升级通知', 1, '<p>系统将于本周六凌晨2:00-4:00进行升级维护，届时系统将暂停服务。请提前做好相关工作安排。</p><p><strong>升级内容：</strong></p><ul><li>性能优化</li><li>安全补丁更新</li><li>新功能发布</li></ul>', 1, '', 1, 1, NOW(), NOW());

INSERT IGNORE INTO sys_notice (id, title, type, content, status, remark, created_by, updated_by, created_at, updated_at)
VALUES (2, '关于规范使用系统的公告', 2, '<p>为保障系统安全稳定运行，请各位用户注意以下事项：</p><ol><li>请定期修改密码，密码长度不少于8位</li><li>不要将账号密码告知他人</li><li>离开工位时请锁定电脑屏幕</li></ol><p>感谢大家的配合！</p>', 1, '', 1, 1, NOW(), NOW());

INSERT IGNORE INTO sys_notice (id, title, type, content, status, remark, created_by, updated_by, created_at, updated_at)
VALUES (3, '新功能上线预告', 1, '<p>我们即将上线以下新功能：</p><ul><li>通知公告管理</li><li>消息中心</li><li>富文本编辑器</li></ul><p>敬请期待！</p>', 0, '草稿状态', 1, 1, NOW(), NOW());
