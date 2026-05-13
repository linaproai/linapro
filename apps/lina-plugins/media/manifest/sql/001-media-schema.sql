-- ------------------------------------------------------------
-- 001 media schema SQL file
-- Purpose: Stores media strategies, strategy bindings, stream aliases, and tenant whitelist entries for the media source plugin.
-- ------------------------------------------------------------

CREATE TABLE IF NOT EXISTS media_strategy (
    "id"         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "name"       VARCHAR(255) NOT NULL DEFAULT '',
    "strategy"   TEXT NOT NULL DEFAULT '',
    "global"     INT NOT NULL DEFAULT 2,
    "enable"     INT NOT NULL DEFAULT 1,
    "creator_id" BIGINT NOT NULL DEFAULT 0,
    "updater_id" BIGINT NOT NULL DEFAULT 0,
    "create_time" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    "update_time" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    CONSTRAINT ck_media_strategy_global CHECK ("global" IN (1, 2)),
    CONSTRAINT ck_media_strategy_enable CHECK ("enable" IN (1, 2))
);

COMMENT ON TABLE media_strategy IS '媒体策略记录表';
COMMENT ON COLUMN media_strategy."id" IS '策略ID';
COMMENT ON COLUMN media_strategy."name" IS '策略名称';
COMMENT ON COLUMN media_strategy."strategy" IS 'YAML格式策略内容';
COMMENT ON COLUMN media_strategy."global" IS '是否全局策略：1是，2否';
COMMENT ON COLUMN media_strategy."enable" IS '启用状态：1开启，2关闭';
COMMENT ON COLUMN media_strategy."creator_id" IS '创建人ID';
COMMENT ON COLUMN media_strategy."updater_id" IS '修改人ID';
COMMENT ON COLUMN media_strategy."create_time" IS '创建时间';
COMMENT ON COLUMN media_strategy."update_time" IS '修改时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_media_strategy_single_global ON media_strategy ("global") WHERE "global" = 1;
CREATE INDEX IF NOT EXISTS idx_media_strategy_enable ON media_strategy ("enable");

CREATE OR REPLACE FUNCTION media_touch_update_time()
RETURNS TRIGGER AS $$
BEGIN
    NEW."update_time" = CURRENT_TIMESTAMP(3);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_media_strategy_update_time ON media_strategy;
CREATE TRIGGER trg_media_strategy_update_time
    BEFORE UPDATE ON media_strategy
    FOR EACH ROW
    EXECUTE FUNCTION media_touch_update_time();

CREATE TABLE IF NOT EXISTS media_strategy_device (
    "device_id"   VARCHAR(255) NOT NULL,
    "strategy_id" BIGINT NOT NULL,
    PRIMARY KEY ("device_id"),
    CONSTRAINT fk_media_strategy_device_strategy FOREIGN KEY ("strategy_id") REFERENCES media_strategy ("id") ON DELETE RESTRICT
);

COMMENT ON TABLE media_strategy_device IS '设备策略表';
COMMENT ON COLUMN media_strategy_device."device_id" IS '设备国标ID';
COMMENT ON COLUMN media_strategy_device."strategy_id" IS '策略ID';

CREATE INDEX IF NOT EXISTS idx_media_strategy_device_strategy ON media_strategy_device ("strategy_id");

CREATE TABLE IF NOT EXISTS media_strategy_device_tenant (
    "tenant_id"   VARCHAR(255) NOT NULL,
    "device_id"   VARCHAR(255) NOT NULL,
    "strategy_id" BIGINT NOT NULL,
    PRIMARY KEY ("tenant_id", "device_id"),
    CONSTRAINT fk_media_strategy_device_tenant_strategy FOREIGN KEY ("strategy_id") REFERENCES media_strategy ("id") ON DELETE RESTRICT
);

COMMENT ON TABLE media_strategy_device_tenant IS '租户设备策略表';
COMMENT ON COLUMN media_strategy_device_tenant."tenant_id" IS '租户ID';
COMMENT ON COLUMN media_strategy_device_tenant."device_id" IS '设备国标ID';
COMMENT ON COLUMN media_strategy_device_tenant."strategy_id" IS '策略ID';

CREATE INDEX IF NOT EXISTS idx_media_strategy_device_tenant_strategy ON media_strategy_device_tenant ("strategy_id");
CREATE INDEX IF NOT EXISTS idx_media_strategy_device_tenant_device ON media_strategy_device_tenant ("device_id");

CREATE TABLE IF NOT EXISTS media_strategy_tenant (
    "tenant_id"   VARCHAR(255) NOT NULL,
    "strategy_id" BIGINT NOT NULL,
    PRIMARY KEY ("tenant_id"),
    CONSTRAINT fk_media_strategy_tenant_strategy FOREIGN KEY ("strategy_id") REFERENCES media_strategy ("id") ON DELETE RESTRICT
);

COMMENT ON TABLE media_strategy_tenant IS '租户策略表';
COMMENT ON COLUMN media_strategy_tenant."tenant_id" IS '租户ID';
COMMENT ON COLUMN media_strategy_tenant."strategy_id" IS '策略ID';

CREATE INDEX IF NOT EXISTS idx_media_strategy_tenant_strategy ON media_strategy_tenant ("strategy_id");

CREATE TABLE IF NOT EXISTS media_stream_alias (
    "id"          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    "alias"       VARCHAR(255) NOT NULL,
    "auto_remove" INT NOT NULL DEFAULT 0,
    "stream_path" VARCHAR(255) NOT NULL DEFAULT '',
    "create_time" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    CONSTRAINT ck_media_stream_alias_auto_remove CHECK ("auto_remove" IN (0, 1))
);

COMMENT ON TABLE media_stream_alias IS '流别名表';
COMMENT ON COLUMN media_stream_alias."id" IS 'ID';
COMMENT ON COLUMN media_stream_alias."alias" IS '流别名';
COMMENT ON COLUMN media_stream_alias."auto_remove" IS '是否自动移除：1是，0否';
COMMENT ON COLUMN media_stream_alias."stream_path" IS '真实流路径';
COMMENT ON COLUMN media_stream_alias."create_time" IS '创建时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_media_stream_alias_alias ON media_stream_alias ("alias");

CREATE TABLE IF NOT EXISTS hg_tenant_white (
    "tenant_id"   VARCHAR(64) NOT NULL,
    "ip"          VARCHAR(32) NOT NULL,
    "description" VARCHAR(32),
    "enable"      SMALLINT NOT NULL DEFAULT 1,
    "creator_id"  INTEGER,
    "create_time" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updater_id"  INTEGER,
    "update_time" TIMESTAMP,
    CONSTRAINT uk_hg_tenant_white_tenant_ip UNIQUE ("tenant_id", "ip"),
    CONSTRAINT ck_hg_tenant_white_enable CHECK ("enable" IN (0, 1))
);

COMMENT ON TABLE hg_tenant_white IS '租户白名单表';
COMMENT ON COLUMN hg_tenant_white."tenant_id" IS '租户ID';
COMMENT ON COLUMN hg_tenant_white."ip" IS '白名单地址';
COMMENT ON COLUMN hg_tenant_white."description" IS '白名单描述';
COMMENT ON COLUMN hg_tenant_white."enable" IS '1开启，0关闭';
COMMENT ON COLUMN hg_tenant_white."creator_id" IS '创建人ID';
COMMENT ON COLUMN hg_tenant_white."create_time" IS '创建时间';
COMMENT ON COLUMN hg_tenant_white."updater_id" IS '修改人ID';
COMMENT ON COLUMN hg_tenant_white."update_time" IS '修改时间';

CREATE INDEX IF NOT EXISTS idx_hg_tenant_white_enable ON hg_tenant_white ("enable");
CREATE INDEX IF NOT EXISTS idx_hg_tenant_white_ip ON hg_tenant_white ("ip");
