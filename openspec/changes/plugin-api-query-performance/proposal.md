## Why

访问插件管理页时，`GET /api/v1/plugins` 会执行插件扫描、同步和大量数据库写入，导致只读页面访问产生大量 SQL，并违反 RESTful 查询接口无副作用的约束。同时，插件宿主服务表注释查询会触发 `SHOW FULL COLUMNS FROM TABLES` 错误日志，在线会话活跃时间也会在短时间内被频繁更新。

## What Changes

- 将插件列表查询改为纯读投影接口，不再在 `GET /api/v1/plugins` 中扫描源码插件或写入插件治理表。
- 保留 `POST /api/v1/plugins/sync` 作为显式插件同步入口，负责扫描、同步和更新插件治理状态。
- 修复插件宿主服务数据表注释查询方式，避免 GoFrame 对 `information_schema.TABLES` 触发错误的字段探测日志。
- 对受保护接口认证链路中的 `sys_online_session.last_active_time` 更新增加节流策略，减少同一会话短时间内的重复写入，同时继续保证超时判断准确。
- 补充针对插件列表纯读行为、表注释查询降级行为和会话活跃时间节流的自动化测试。

## Capabilities

### New Capabilities

- 无

### Modified Capabilities

- `plugin-manifest-lifecycle`: 插件列表查询必须是无副作用的读操作，插件扫描和同步仅通过显式同步动作触发。
- `online-user`: 受保护请求仍需校验会话有效性和超时，但允许在短时间窗口内跳过重复的 `last_active_time` 写入。

## Impact

- 后端插件服务：`apps/lina-core/internal/service/plugin/` 和 `apps/lina-core/internal/controller/plugin/`
- 后端会话服务：`apps/lina-core/internal/service/session/`
- 插件管理 API：`GET /api/v1/plugins`、`POST /api/v1/plugins/sync`
- 前端插件管理页保持现有调用方式，但页面加载时不再触发插件治理表写入。
- i18n 影响：本次不新增或修改用户可见文案，不需要调整运行时语言包、插件 manifest i18n 或 apidoc i18n 资源。
