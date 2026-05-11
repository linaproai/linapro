-- Mock data: inbox notification messages and deliveries.
-- 模拟数据：站内信通知消息与投递记录。
-- Static notification history rows use exact existence checks so mock loading is idempotent.

INSERT INTO sys_notify_message (
    "tenant_id",
    "plugin_id",
    "source_type",
    "source_id",
    "category_code",
    "title",
    "content",
    "payload_json",
    "sender_user_id",
    "created_at"
)
SELECT
    0,
    '',
    'system',
    'mock-welcome',
    'notice',
    '欢迎使用 LinaPro',
    '这是一条用于演示站内信列表、未读状态和消息详情的 mock 通知。',
    '{"priority":"normal","mock":true}',
    admin."id",
    '2026-04-20 09:00:00'
FROM sys_user admin
WHERE admin."username" = 'admin'
  AND NOT EXISTS (
      SELECT 1
      FROM sys_notify_message existing
      WHERE existing."tenant_id" = 0
        AND existing."plugin_id" = ''
        AND existing."source_type" = 'system'
        AND existing."source_id" = 'mock-welcome'
  );

INSERT INTO sys_notify_message (
    "tenant_id",
    "plugin_id",
    "source_type",
    "source_id",
    "category_code",
    "title",
    "content",
    "payload_json",
    "sender_user_id",
    "created_at"
)
SELECT
    0,
    'content-notice',
    'notice',
    'mock-maintenance',
    'announcement',
    '系统维护提醒',
    '内容公告插件发布维护公告后，可通过统一通知域投递到用户站内信。',
    '{"priority":"high","mock":true}',
    admin."id",
    '2026-04-21 10:30:00'
FROM sys_user admin
WHERE admin."username" = 'admin'
  AND NOT EXISTS (
      SELECT 1
      FROM sys_notify_message existing
      WHERE existing."tenant_id" = 0
        AND existing."plugin_id" = 'content-notice'
        AND existing."source_type" = 'notice'
        AND existing."source_id" = 'mock-maintenance'
  );

INSERT INTO sys_notify_delivery (
    "tenant_id",
    "message_id",
    "channel_key",
    "channel_type",
    "recipient_type",
    "recipient_key",
    "user_id",
    "delivery_status",
    "is_read",
    "read_at",
    "sent_at",
    "created_at",
    "updated_at"
)
SELECT
    0,
    "msg"."id",
    'inbox',
    'inbox',
    'user',
    CAST(u."id" AS VARCHAR),
    u."id",
    1,
    0,
    NULL::TIMESTAMP,
    TIMESTAMP '2026-04-20 09:00:10',
    TIMESTAMP '2026-04-20 09:00:10',
    TIMESTAMP '2026-04-20 09:00:10'
FROM sys_notify_message "msg"
JOIN sys_user u ON u."username" IN ('admin', 'user002', 'user060')
WHERE "msg"."tenant_id" = 0
  AND "msg"."source_type" = 'system'
  AND "msg"."source_id" = 'mock-welcome'
  AND "msg"."id" = (
      SELECT latest."id"
      FROM sys_notify_message latest
      WHERE latest."tenant_id" = 0
        AND latest."source_type" = 'system'
        AND latest."source_id" = 'mock-welcome'
      ORDER BY latest."id" DESC
      LIMIT 1
  )
  AND NOT EXISTS (
      SELECT 1
      FROM sys_notify_delivery existing
      WHERE existing."tenant_id" = 0
        AND existing."message_id" = "msg"."id"
        AND existing."channel_key" = 'inbox'
        AND existing."recipient_type" = 'user'
        AND existing."user_id" = u."id"
  );

INSERT INTO sys_notify_delivery (
    "tenant_id",
    "message_id",
    "channel_key",
    "channel_type",
    "recipient_type",
    "recipient_key",
    "user_id",
    "delivery_status",
    "is_read",
    "read_at",
    "sent_at",
    "created_at",
    "updated_at"
)
SELECT
    0,
    "msg"."id",
    'inbox',
    'inbox',
    'user',
    CAST(u."id" AS VARCHAR),
    u."id",
    1,
    CASE WHEN u."username" = 'admin' THEN 1 ELSE 0 END,
    CASE WHEN u."username" = 'admin' THEN TIMESTAMP '2026-04-21 11:00:00' ELSE NULL::TIMESTAMP END,
    TIMESTAMP '2026-04-21 10:31:00',
    TIMESTAMP '2026-04-21 10:31:00',
    TIMESTAMP '2026-04-21 10:31:00'
FROM sys_notify_message "msg"
JOIN sys_user u ON u."username" IN ('admin', 'user009', 'user021')
WHERE "msg"."tenant_id" = 0
  AND "msg"."source_type" = 'notice'
  AND "msg"."source_id" = 'mock-maintenance'
  AND "msg"."id" = (
      SELECT latest."id"
      FROM sys_notify_message latest
      WHERE latest."tenant_id" = 0
        AND latest."source_type" = 'notice'
        AND latest."source_id" = 'mock-maintenance'
      ORDER BY latest."id" DESC
      LIMIT 1
  )
  AND NOT EXISTS (
      SELECT 1
      FROM sys_notify_delivery existing
      WHERE existing."tenant_id" = 0
        AND existing."message_id" = "msg"."id"
        AND existing."channel_key" = 'inbox'
        AND existing."recipient_type" = 'user'
        AND existing."user_id" = u."id"
  );
