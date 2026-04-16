-- ============================================================
-- 宿主插件表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_plugin (
    id            INT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    plugin_id     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '插件唯一标识（kebab-case）',
    name          VARCHAR(128) NOT NULL DEFAULT '' COMMENT '插件名称',
    version       VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '插件版本号',
    type          VARCHAR(32)  NOT NULL DEFAULT 'source' COMMENT '插件一级类型（source/dynamic）',
    installed     TINYINT      NOT NULL DEFAULT 0 COMMENT '安装状态（1=已安装 0=未安装）',
    status        TINYINT      NOT NULL DEFAULT 0 COMMENT '启用状态（1=启用 0=禁用）',
    desired_state VARCHAR(32)  NOT NULL DEFAULT 'uninstalled' COMMENT '宿主期望状态（uninstalled/installed/enabled）',
    current_state VARCHAR(32)  NOT NULL DEFAULT 'uninstalled' COMMENT '宿主当前状态（uninstalled/installed/enabled/reconciling/failed）',
    generation    BIGINT       NOT NULL DEFAULT 1 COMMENT '宿主当前生效代际号',
    release_id    INT          NOT NULL DEFAULT 0 COMMENT '宿主当前生效 release ID',
    manifest_path VARCHAR(255) NOT NULL DEFAULT '' COMMENT '插件清单文件路径',
    checksum      VARCHAR(128) NOT NULL DEFAULT '' COMMENT '插件包校验值',
    installed_at  DATETIME                          COMMENT '安装时间',
    enabled_at    DATETIME                          COMMENT '最后一次启用时间',
    disabled_at   DATETIME                          COMMENT '最后一次禁用时间',
    remark        VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
    created_at    DATETIME                          COMMENT '创建时间',
    updated_at    DATETIME                          COMMENT '更新时间',
    deleted_at    DATETIME                          COMMENT '删除时间',
    UNIQUE KEY uk_plugin_id (plugin_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='插件注册表';

-- ============================================================
-- 插件发布记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_plugin_release (
    id                INT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    plugin_id         VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '插件唯一标识（kebab-case）',
    release_version   VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '插件版本号',
    type              VARCHAR(32)   NOT NULL DEFAULT 'source' COMMENT '插件一级类型（source/dynamic）',
    runtime_kind      VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '运行时产物类型（当前仅 wasm）',
    schema_version    VARCHAR(32)   NOT NULL DEFAULT '' COMMENT 'plugin.yaml 清单 schema 版本',
    min_host_version  VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '宿主最小兼容版本',
    max_host_version  VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '宿主最大兼容版本',
    status            VARCHAR(32)   NOT NULL DEFAULT '' COMMENT 'release 状态（prepared/installed/active/uninstalled/failed）',
    manifest_path     VARCHAR(255)  NOT NULL DEFAULT '' COMMENT '插件清单路径',
    package_path      VARCHAR(255)  NOT NULL DEFAULT '' COMMENT '插件源码目录或运行时产物路径',
    checksum          VARCHAR(128)  NOT NULL DEFAULT '' COMMENT '插件清单或产物校验值',
    manifest_snapshot LONGTEXT                           COMMENT '插件清单与资源摘要快照（YAML，不保存具体 SQL/前端文件路径）',
    created_at        DATETIME                           COMMENT '创建时间',
    updated_at        DATETIME                           COMMENT '更新时间',
    deleted_at        DATETIME                           COMMENT '删除时间',
    UNIQUE KEY uk_plugin_release (plugin_id, release_version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='插件发布记录表';

-- ============================================================
-- 插件迁移执行记录表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_plugin_migration (
    id              INT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    plugin_id       VARCHAR(64)   NOT NULL DEFAULT '' COMMENT '插件唯一标识（kebab-case）',
    release_id      INT           NOT NULL DEFAULT 0 COMMENT '所属插件 release ID',
    phase           VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '迁移阶段（install/uninstall/upgrade/rollback）',
    migration_key   VARCHAR(255)  NOT NULL DEFAULT '' COMMENT '迁移执行键（如 install-step-001，不保存具体 SQL 路径）',
    checksum        VARCHAR(128)  NOT NULL DEFAULT '' COMMENT '迁移文件校验值',
    execution_order INT           NOT NULL DEFAULT 0 COMMENT '执行顺序（从1开始）',
    status          VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '执行状态（pending/succeeded/failed/skipped）',
    executed_at     DATETIME                           COMMENT '执行时间',
    error_message   VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '失败原因或补充说明',
    created_at      DATETIME                           COMMENT '创建时间',
    updated_at      DATETIME                           COMMENT '更新时间',
    UNIQUE KEY uk_plugin_migration (plugin_id, release_id, phase, migration_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='插件迁移执行记录表';

-- ============================================================
-- 插件资源引用表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_plugin_resource_ref (
    id             INT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    plugin_id      VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '插件唯一标识（kebab-case）',
    release_id     INT          NOT NULL DEFAULT 0 COMMENT '所属插件 release ID',
    resource_type  VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '资源类型（manifest/sql/frontend/menu/permission 等）',
    resource_key   VARCHAR(255) NOT NULL DEFAULT '' COMMENT '资源唯一键',
    resource_path  VARCHAR(255) NOT NULL DEFAULT '' COMMENT '资源定位补充信息（默认留空，不保存具体前端/SQL 路径）',
    owner_type     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '宿主对象类型（file/menu/route/slot 等）',
    owner_key      VARCHAR(255) NOT NULL DEFAULT '' COMMENT '宿主对象稳定标识',
    remark         VARCHAR(512) NOT NULL DEFAULT '' COMMENT '备注',
    created_at     DATETIME                          COMMENT '创建时间',
    updated_at     DATETIME                          COMMENT '更新时间',
    deleted_at     DATETIME                          COMMENT '删除时间',
    UNIQUE KEY uk_plugin_resource_ref (plugin_id, release_id, resource_type, resource_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='插件资源引用表';

-- ============================================================
-- 插件节点状态表
-- ============================================================
CREATE TABLE IF NOT EXISTS sys_plugin_node_state (
    id                INT PRIMARY KEY AUTO_INCREMENT COMMENT '主键ID',
    plugin_id         VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '插件唯一标识（kebab-case）',
    release_id        INT          NOT NULL DEFAULT 0 COMMENT '所属插件 release ID',
    node_key          VARCHAR(128) NOT NULL DEFAULT '' COMMENT '节点唯一标识',
    desired_state     VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '节点期望状态',
    current_state     VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '节点当前状态',
    generation        BIGINT       NOT NULL DEFAULT 0 COMMENT '插件代际号',
    last_heartbeat_at DATETIME                          COMMENT '最近一次心跳时间',
    error_message     VARCHAR(512) NOT NULL DEFAULT '' COMMENT '节点错误信息',
    created_at        DATETIME                          COMMENT '创建时间',
    updated_at        DATETIME                          COMMENT '更新时间',
    UNIQUE KEY uk_plugin_node_state (plugin_id, node_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='插件节点状态表';
