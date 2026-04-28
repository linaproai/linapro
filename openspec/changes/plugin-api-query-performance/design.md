## Context

当前插件管理页加载时会调用 `GET /api/v1/plugins`。该接口内部复用了 `SyncAndList`，会扫描源码插件清单、同步 `sys_plugin` / `sys_plugin_release` / `sys_plugin_resource_ref` / `sys_menu` 等治理表，并刷新插件启用快照。这使普通列表查询产生大量写 SQL，也让 GET 接口拥有副作用。

日志还显示插件列表投影在解析动态插件 hostServices 数据表说明时，会通过 GoFrame model 查询 `information_schema.TABLES`，触发错误的 `SHOW FULL COLUMNS FROM TABLES` 元数据探测。受保护接口认证链路则会在每次请求时读取并更新 `sys_online_session.last_active_time`，页面请求并发时会放大写入量。

## Goals / Non-Goals

**Goals:**

- 让 `GET /api/v1/plugins` 成为纯读接口，不扫描插件目录，不写插件治理表。
- 保留 `POST /api/v1/plugins/sync` 作为显式同步入口，继续承担扫描和治理表同步职责。
- 修复 `information_schema.TABLES` 查询导致的错误日志。
- 在不降低会话超时判断准确性的前提下，减少同一会话短时间内的 `last_active_time` 重复更新。
- 补充后端单元测试覆盖关键行为。

**Non-Goals:**

- 不改变插件安装、启用、禁用、卸载的业务流程。
- 不调整插件管理页的用户可见文案和布局。
- 不引入新的缓存基础设施或外部依赖。
- 不改变数据库结构或 SQL 初始化文件。

## Decisions

1. 插件列表查询从“同步后投影”拆分为“只读投影”。

   `GET /api/v1/plugins` 调用新的只读列表方法，从现有注册表和已发现清单构造列表项。显式的 `POST /api/v1/plugins/sync` 继续调用同步流程。这样保持管理页现有 API 路径不变，同时恢复 GET 的无副作用语义。

   备选方案是保留自动同步但跳过无变化更新。该方案仍然需要扫描目录和尝试写库，不符合 RESTful 约束，因此不采用。

2. 表注释查询使用原始 SQL 读取 `information_schema.TABLES`。

   该查询是元数据读操作，不需要 GoFrame 对目标表做字段探测。改为参数化 raw SQL 后，可以避免 `SHOW FULL COLUMNS FROM TABLES` 错误，同时保留 MySQL/TiDB/MariaDB 下的表注释展示能力。

   备选方案是静默忽略该错误。该方案仍会产生多余 SQL 和错误路径开销，因此不采用。

3. 会话活跃时间按最小写入窗口更新。

   认证链路仍然每次读取会话并判断超时；仅当 `last_active_time` 距当前时间超过一个短窗口时才写入新时间。该窗口使用代码内常量，避免新增配置项和 i18n 文案。

   备选方案是完全跳过活跃时间更新，无法支持真实在线活跃判断和超时延长，因此不采用。

## Risks / Trade-offs

- 插件目录新增或清单变更后，列表页不会自动写入治理表 → 管理员需要使用“同步插件”按钮或启动同步流程刷新注册表。
- 会话活跃时间短窗口内不更新 → 在线用户页面看到的最后活跃时间可能有短暂延迟，但超时判断仍基于读取到的持久化时间和后续窗口写入。
- 原始 SQL 需要保持数据库方言边界 → 仅对 MySQL、MariaDB、TiDB 执行元数据查询，其他数据库继续降级为空注释。

## Migration Plan

无需数据库迁移。部署后，插件管理页加载只读列表；需要发现新源码插件或刷新清单时，管理员继续通过 `POST /api/v1/plugins/sync` 对应的页面操作触发同步。若出现回滚，旧版本会恢复列表查询自动同步行为。
