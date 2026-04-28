-- Mock data: inbox notification messages and deliveries.

INSERT IGNORE INTO sys_notify_message (
    plugin_id,
    source_type,
    source_id,
    category_code,
    title,
    content,
    payload_json,
    sender_user_id,
    created_at
)
SELECT
    '',
    'system',
    'mock-welcome',
    'notice',
    '欢迎使用 LinaPro',
    '这是一条用于演示站内信列表、未读状态和消息详情的 mock 通知。',
    '{"priority":"normal","mock":true}',
    admin.id,
    '2026-04-20 09:00:00'
FROM sys_user admin
WHERE admin.username = 'admin'
  AND NOT EXISTS (
      SELECT 1
      FROM sys_notify_message
      WHERE source_type = 'system'
        AND source_id = 'mock-welcome'
  );

INSERT IGNORE INTO sys_notify_message (
    plugin_id,
    source_type,
    source_id,
    category_code,
    title,
    content,
    payload_json,
    sender_user_id,
    created_at
)
SELECT
    'content-notice',
    'notice',
    'mock-maintenance',
    'announcement',
    '系统维护提醒',
    '内容公告插件发布维护公告后，可通过统一通知域投递到用户站内信。',
    '{"priority":"high","mock":true}',
    admin.id,
    '2026-04-21 10:30:00'
FROM sys_user admin
WHERE admin.username = 'admin'
  AND NOT EXISTS (
      SELECT 1
      FROM sys_notify_message
      WHERE source_type = 'notice'
        AND source_id = 'mock-maintenance'
  );

INSERT IGNORE INTO sys_notify_delivery (
    message_id,
    channel_key,
    channel_type,
    recipient_type,
    recipient_key,
    user_id,
    delivery_status,
    is_read,
    read_at,
    sent_at,
    created_at,
    updated_at
)
SELECT
    msg.id,
    'inbox',
    'inbox',
    'user',
    CAST(u.id AS CHAR),
    u.id,
    1,
    0,
    NULL,
    '2026-04-20 09:00:10',
    '2026-04-20 09:00:10',
    '2026-04-20 09:00:10'
FROM sys_notify_message msg
JOIN sys_user u ON u.username IN ('admin', 'user002', 'user060')
WHERE msg.source_type = 'system'
  AND msg.source_id = 'mock-welcome'
  AND NOT EXISTS (
      SELECT 1
      FROM sys_notify_delivery existing
      WHERE existing.message_id = msg.id
        AND existing.user_id = u.id
        AND existing.channel_key = 'inbox'
  );

INSERT IGNORE INTO sys_notify_delivery (
    message_id,
    channel_key,
    channel_type,
    recipient_type,
    recipient_key,
    user_id,
    delivery_status,
    is_read,
    read_at,
    sent_at,
    created_at,
    updated_at
)
SELECT
    msg.id,
    'inbox',
    'inbox',
    'user',
    CAST(u.id AS CHAR),
    u.id,
    1,
    CASE WHEN u.username = 'admin' THEN 1 ELSE 0 END,
    CASE WHEN u.username = 'admin' THEN '2026-04-21 11:00:00' ELSE NULL END,
    '2026-04-21 10:31:00',
    '2026-04-21 10:31:00',
    '2026-04-21 10:31:00'
FROM sys_notify_message msg
JOIN sys_user u ON u.username IN ('admin', 'user009', 'user021')
WHERE msg.source_type = 'notice'
  AND msg.source_id = 'mock-maintenance'
  AND NOT EXISTS (
      SELECT 1
      FROM sys_notify_delivery existing
      WHERE existing.message_id = msg.id
        AND existing.user_id = u.id
        AND existing.channel_key = 'inbox'
  );
