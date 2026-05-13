-- Mock data: media strategy, binding, and stream-alias examples.
-- 模拟数据：媒体策略、策略绑定和流别名案例。

INSERT INTO media_strategy (
    "name",
    "strategy",
    "global",
    "enable",
    "creator_id",
    "updater_id",
    "create_time",
    "update_time"
)
SELECT
    '默认直播录制策略',
    'record:
  enabled: true
  format: mp4
  retainDays: 7
stream:
  transport: tcp
  timeout: 10s
snapshot:
  enabled: true
  interval: 30s',
    1,
    1,
    admin."id",
    admin."id",
    '2026-05-13 09:00:00',
    '2026-05-13 09:00:00'
FROM sys_user admin
WHERE admin."username" = 'admin'
  AND NOT EXISTS (
      SELECT 1
      FROM media_strategy existing
      WHERE existing."name" = '默认直播录制策略'
  )
  AND NOT EXISTS (
      SELECT 1
      FROM media_strategy existing
      WHERE existing."global" = 1
  );

INSERT INTO media_strategy (
    "name",
    "strategy",
    "global",
    "enable",
    "creator_id",
    "updater_id",
    "create_time",
    "update_time"
)
SELECT
    '门店低延迟预览策略',
    'record:
  enabled: false
stream:
  transport: udp
  latencyMode: low
  timeout: 5s
transcode:
  enabled: true
  profile: mobile-preview',
    2,
    1,
    admin."id",
    admin."id",
    '2026-05-13 09:10:00',
    '2026-05-13 09:10:00'
FROM sys_user admin
WHERE admin."username" = 'admin'
  AND NOT EXISTS (
      SELECT 1
      FROM media_strategy existing
      WHERE existing."name" = '门店低延迟预览策略'
  );

INSERT INTO media_strategy (
    "name",
    "strategy",
    "global",
    "enable",
    "creator_id",
    "updater_id",
    "create_time",
    "update_time"
)
SELECT
    '园区安防留存策略',
    'record:
  enabled: true
  format: hls
  retainDays: 30
stream:
  transport: tcp
  timeout: 15s
watermark:
  enabled: true
  text: 园区安防',
    2,
    1,
    admin."id",
    admin."id",
    '2026-05-13 09:20:00',
    '2026-05-13 09:20:00'
FROM sys_user admin
WHERE admin."username" = 'admin'
  AND NOT EXISTS (
      SELECT 1
      FROM media_strategy existing
      WHERE existing."name" = '园区安防留存策略'
  );

INSERT INTO media_strategy_device (
    "device_id",
    "strategy_id"
)
SELECT
    '34020000001320000001',
    strategy."id"
FROM media_strategy strategy
WHERE strategy."name" = '门店低延迟预览策略'
  AND NOT EXISTS (
      SELECT 1
      FROM media_strategy_device existing
      WHERE existing."device_id" = '34020000001320000001'
  );

INSERT INTO media_strategy_tenant (
    "tenant_id",
    "strategy_id"
)
SELECT
    'tenant-retail-east',
    strategy."id"
FROM media_strategy strategy
WHERE strategy."name" = '门店低延迟预览策略'
  AND NOT EXISTS (
      SELECT 1
      FROM media_strategy_tenant existing
      WHERE existing."tenant_id" = 'tenant-retail-east'
  );

INSERT INTO media_strategy_device_tenant (
    "tenant_id",
    "device_id",
    "strategy_id"
)
SELECT
    'tenant-park-security',
    '34020000001320000002',
    strategy."id"
FROM media_strategy strategy
WHERE strategy."name" = '园区安防留存策略'
  AND NOT EXISTS (
      SELECT 1
      FROM media_strategy_device_tenant existing
      WHERE existing."tenant_id" = 'tenant-park-security'
        AND existing."device_id" = '34020000001320000002'
  );

INSERT INTO media_stream_alias (
    "alias",
    "auto_remove",
    "stream_path",
    "create_time"
)
SELECT
    'retail-east-entrance',
    0,
    'live/tenant-retail-east/entrance',
    '2026-05-13 10:00:00'
WHERE NOT EXISTS (
    SELECT 1
    FROM media_stream_alias existing
    WHERE existing."alias" = 'retail-east-entrance'
);

INSERT INTO media_stream_alias (
    "alias",
    "auto_remove",
    "stream_path",
    "create_time"
)
SELECT
    'park-gate-north',
    0,
    'live/tenant-park-security/gate-north',
    '2026-05-13 10:05:00'
WHERE NOT EXISTS (
    SELECT 1
    FROM media_stream_alias existing
    WHERE existing."alias" = 'park-gate-north'
);

INSERT INTO media_stream_alias (
    "alias",
    "auto_remove",
    "stream_path",
    "create_time"
)
SELECT
    'temporary-event-room',
    1,
    'live/events/temporary-room',
    '2026-05-13 10:10:00'
WHERE NOT EXISTS (
    SELECT 1
    FROM media_stream_alias existing
    WHERE existing."alias" = 'temporary-event-room'
);

INSERT INTO hg_tenant_white (
    "tenant_id",
    "ip",
    "description",
    "enable",
    "creator_id",
    "create_time",
    "updater_id",
    "update_time"
)
SELECT
    'tenant-retail-east',
    '10.8.1.24',
    '门店出口',
    1,
    admin."id",
    '2026-05-13 10:20:00',
    admin."id",
    '2026-05-13 10:20:00'
FROM sys_user admin
WHERE admin."username" = 'admin'
  AND NOT EXISTS (
      SELECT 1
      FROM hg_tenant_white existing
      WHERE existing."tenant_id" = 'tenant-retail-east'
        AND existing."ip" = '10.8.1.24'
  );

INSERT INTO hg_tenant_white (
    "tenant_id",
    "ip",
    "description",
    "enable",
    "creator_id",
    "create_time",
    "updater_id",
    "update_time"
)
SELECT
    'tenant-park-security',
    '172.16.20.8',
    '园区出口',
    1,
    admin."id",
    '2026-05-13 10:25:00',
    admin."id",
    '2026-05-13 10:25:00'
FROM sys_user admin
WHERE admin."username" = 'admin'
  AND NOT EXISTS (
      SELECT 1
      FROM hg_tenant_white existing
      WHERE existing."tenant_id" = 'tenant-park-security'
        AND existing."ip" = '172.16.20.8'
  );

INSERT INTO hg_tenant_white (
    "tenant_id",
    "ip",
    "description",
    "enable",
    "creator_id",
    "create_time",
    "updater_id",
    "update_time"
)
SELECT
    'tenant-temporary-event',
    '203.0.113.18',
    '临时活动',
    0,
    admin."id",
    '2026-05-13 10:30:00',
    admin."id",
    '2026-05-13 10:30:00'
FROM sys_user admin
WHERE admin."username" = 'admin'
  AND NOT EXISTS (
      SELECT 1
      FROM hg_tenant_white existing
      WHERE existing."tenant_id" = 'tenant-temporary-event'
        AND existing."ip" = '203.0.113.18'
  );
