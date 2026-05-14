-- ------------------------------------------------------------
-- 001 water mock data SQL file
-- Purpose: Water mock data enriches media strategy mock rows because water reads media_* tables.
-- ------------------------------------------------------------

DO $$
BEGIN
    IF to_regclass('public.media_strategy') IS NOT NULL THEN
        -- Demo mock row patch: add a complete watermark node to the default strategy when it was loaded before this plugin existed.
        -- 演示数据补丁：为此前已加载的默认策略补齐完整水印节点。
        UPDATE media_strategy
        SET
            "enable" = 1,
            "strategy" = 'record:
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
            "update_time" = CURRENT_TIMESTAMP(3)
        WHERE "name" = '默认直播录制策略'
          AND "create_time" = TIMESTAMP '2026-05-13 09:00:00'
          AND ("strategy" NOT LIKE '%watermark:%' OR "strategy" NOT LIKE '%fontSize:%');

        -- Demo mock row patch: the default water page device is bound to this strategy, so it must include watermark settings.
        -- 演示数据补丁：水印页面默认设备命中该策略，因此需要携带水印配置。
        UPDATE media_strategy
        SET
            "enable" = 1,
            "strategy" = 'record:
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
            "update_time" = CURRENT_TIMESTAMP(3)
        WHERE "name" = '门店低延迟预览策略'
          AND "create_time" = TIMESTAMP '2026-05-13 09:10:00'
          AND ("strategy" NOT LIKE '%watermark:%' OR "strategy" NOT LIKE '%fontSize:%');

        -- Demo mock row patch: expand the park-security strategy from the old minimal watermark node.
        -- 演示数据补丁：将园区安防策略从旧版简略水印节点补齐为完整配置。
        UPDATE media_strategy
        SET
            "enable" = 1,
            "strategy" = 'record:
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
            "update_time" = CURRENT_TIMESTAMP(3)
        WHERE "name" = '园区安防留存策略'
          AND "create_time" = TIMESTAMP '2026-05-13 09:20:00'
          AND ("strategy" NOT LIKE '%watermark:%' OR "strategy" NOT LIKE '%fontSize:%');

        -- Demo fallback: keep one usable global demo strategy only when no global strategy exists.
        -- 演示兜底：仅在当前没有全局策略时恢复默认演示策略为全局策略。
        UPDATE media_strategy
        SET
            "global" = 1,
            "enable" = 1,
            "update_time" = CURRENT_TIMESTAMP(3)
        WHERE "name" = '默认直播录制策略'
          AND "create_time" = TIMESTAMP '2026-05-13 09:00:00'
          AND NOT EXISTS (
              SELECT 1
              FROM media_strategy existing
              WHERE existing."global" = 1
          );
    END IF;
END $$;
