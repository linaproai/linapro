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
  interval: 30s
watermark:
  enabled: true
  text: LinaPro 水印测试
  fontSize: 42
  color: "#ffffff"
  align: bottomRight
  opacity: 0.75',
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
  profile: mobile-preview
watermark:
  enabled: true
  text: 门店预览
  fontSize: 32
  color: "#ffffff"
  align: bottomRight
  opacity: 0.65',
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
  text: 园区安防
  fontSize: 40
  color: "#ffffff"
  align: bottomRight
  opacity: 0.7',
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

INSERT INTO media_node (
    "node_num",
    "name",
    "qn_url",
    "basic_url",
    "dn_url",
    "creator_id",
    "create_time",
    "updater_id",
    "update_time"
)
SELECT
    1,
    '华东媒体节点',
    'https://qn-east.example.com',
    'https://basic-east.example.com',
    'https://dn-east.example.com',
    admin."id",
    '2026-05-13 10:12:00',
    admin."id",
    '2026-05-13 10:12:00'
FROM sys_user admin
WHERE admin."username" = 'admin'
  AND NOT EXISTS (
      SELECT 1
      FROM media_node existing
      WHERE existing."node_num" = 1
  );

INSERT INTO media_node (
    "node_num",
    "name",
    "qn_url",
    "basic_url",
    "dn_url",
    "creator_id",
    "create_time",
    "updater_id",
    "update_time"
)
SELECT
    2,
    '华北媒体节点',
    'https://qn-north.example.com',
    'https://basic-north.example.com',
    'https://dn-north.example.com',
    admin."id",
    '2026-05-13 10:14:00',
    admin."id",
    '2026-05-13 10:14:00'
FROM sys_user admin
WHERE admin."username" = 'admin'
  AND NOT EXISTS (
      SELECT 1
      FROM media_node existing
      WHERE existing."node_num" = 2
  );

INSERT INTO media_device_node (
    "device_id",
    "node_num"
)
SELECT
    '34020000001320000001',
    1
WHERE EXISTS (
    SELECT 1
    FROM media_node node
    WHERE node."node_num" = 1
)
  AND NOT EXISTS (
      SELECT 1
      FROM media_device_node existing
      WHERE existing."device_id" = '34020000001320000001'
  );

INSERT INTO media_device_node (
    "device_id",
    "node_num"
)
SELECT
    '34020000001320000002',
    2
WHERE EXISTS (
    SELECT 1
    FROM media_node node
    WHERE node."node_num" = 2
)
  AND NOT EXISTS (
      SELECT 1
      FROM media_device_node existing
      WHERE existing."device_id" = '34020000001320000002'
  );

INSERT INTO media_tenant_stream_config (
    "tenant_id",
    "max_concurrent",
    "node_num",
    "enable",
    "creator_id",
    "create_time",
    "updater_id",
    "update_time"
)
SELECT
    'tenant-retail-east',
    80,
    1,
    1,
    admin."id",
    '2026-05-13 10:16:00',
    admin."id",
    '2026-05-13 10:16:00'
FROM sys_user admin
WHERE admin."username" = 'admin'
  AND EXISTS (
      SELECT 1
      FROM media_node node
      WHERE node."node_num" = 1
  )
  AND NOT EXISTS (
      SELECT 1
      FROM media_tenant_stream_config existing
      WHERE existing."tenant_id" = 'tenant-retail-east'
  );

INSERT INTO media_tenant_stream_config (
    "tenant_id",
    "max_concurrent",
    "node_num",
    "enable",
    "creator_id",
    "create_time",
    "updater_id",
    "update_time"
)
SELECT
    'tenant-park-security',
    160,
    2,
    1,
    admin."id",
    '2026-05-13 10:18:00',
    admin."id",
    '2026-05-13 10:18:00'
FROM sys_user admin
WHERE admin."username" = 'admin'
  AND EXISTS (
      SELECT 1
      FROM media_node node
      WHERE node."node_num" = 2
  )
  AND NOT EXISTS (
      SELECT 1
      FROM media_tenant_stream_config existing
      WHERE existing."tenant_id" = 'tenant-park-security'
  );

INSERT INTO media_tenant_white (
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
      FROM media_tenant_white existing
      WHERE existing."tenant_id" = 'tenant-retail-east'
        AND existing."ip" = '10.8.1.24'
  );

INSERT INTO media_tenant_white (
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
      FROM media_tenant_white existing
      WHERE existing."tenant_id" = 'tenant-park-security'
        AND existing."ip" = '172.16.20.8'
  );

INSERT INTO media_tenant_white (
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
      FROM media_tenant_white existing
      WHERE existing."tenant_id" = 'tenant-temporary-event'
        AND existing."ip" = '203.0.113.18'
  );
