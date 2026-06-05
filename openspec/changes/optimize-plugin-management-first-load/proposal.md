## Why

插件管理页面首次进入仍然明显偏慢，当前链路在首屏同时承担插件目录扫描、动态 `wasm` manifest 解析、依赖检查、宿主服务授权、动态路由、cron 声明和多个弹窗组件加载，超出了列表概览所需的工作量。

这个问题现在需要作为接口和读模型契约修复，而不是只依赖启动预热；否则预热未完成、语言缓存键不命中、插件数量增长或动态插件产物增多时，首次请求仍会退化为同步构建完整治理读模型。

## What Changes

- 将插件管理列表首屏改为轻量、分页、有上限的摘要查询；`GET /plugins`只返回表格与行级操作可直接使用的最小字段，不再返回详情、安装、卸载、升级、授权确认弹窗专用字段。
- 保留并强化 `GET /plugins/{id}`及既有动作预览接口作为完整治理详情入口；依赖检查、`hostServices`、授权快照、动态路由、cron 声明等高成本数据只在打开详情或执行相关治理动作时按需加载。
- 后端拆分插件管理“摘要列表读模型”和“详情治理读模型”，列表路径不得通过构建完整 `PluginItem` 后裁剪字段实现，也不得逐插件调用依赖检查、cron 声明或 manifest 解析。
- 前端插件管理页面首屏只加载表格、筛选和常用行操作；详情、上传、安装授权、卸载、升级、生命周期前置条件等弹窗组件改为按需异步加载。
- 调整缓存和预热策略：列表摘要读模型与详情读模型分别管理缓存键、失效范围和预热边界；列表缓存必须绑定 locale、runtime bundle version 和 plugin runtime revision，并在集群模式下复用既有 `plugin-runtime` revision/event 失效。
- 补充性能验证：覆盖首屏 `GET /plugins`分页边界、无前端逐行详情瀑布、无列表依赖检查、无列表逐插件 cron 扫描、manifest/artifact 解析次数有界，以及首次进入页面的 E2E 行为。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `plugin-ui-integration`：插件管理页面首屏从完整治理读模型改为摘要列表，详情和治理弹窗数据改为按需加载，首屏组件包按需拆分。
- `plugin-runtime-loading`：插件管理列表的运行时 manifest/artifact 发现和预热必须满足首屏摘要读模型的有界成本，不得由列表路径重复触发完整运行时扫描。
- `plugin-dependency-management`：依赖检查结果仍必须通过 API 和 UI 可见，但可见位置调整为详情、安装、升级、卸载等需要决策的路径，不再要求随首屏列表返回。
- `distributed-cache-coordination`：插件管理摘要读模型和详情读模型作为插件运行时派生缓存的一部分，需要明确权威源、缓存键、失效范围、跨实例同步和故障降级策略。

## Impact

- 后端 API：`apps/lina-core/api/plugin/v1/plugin_list.go`需要增加分页请求契约，并拆分列表摘要响应与详情响应字段边界；`GET /plugins`仍为受保护 `plugin:query`只读接口。
- 后端实现：`apps/lina-core/internal/service/plugin/`和 `apps/lina-core/internal/controller/plugin/`需要新增或调整摘要列表路径、详情路径投影、缓存预热和性能测试，避免 `N+1`装配。
- 插件运行时扫描：`apps/lina-core/internal/service/plugin/internal/catalog/`和 cron 集成路径需要支持列表首屏复用 bounded snapshot 或批量契约，避免为每个插件重复扫描 runtime manifest。
- 前端：`apps/lina-vben/apps/web-antd/src/views/system/plugin/`和 `src/api/system/plugin/`需要调整列表模型、详情按需请求、弹窗异步加载和 E2E 覆盖。
- 缓存一致性：不新增独立缓存协调域，复用 `plugin-runtime` revision/event；实现时必须记录单机和集群失效策略。
- 数据权限：插件管理是平台治理控制面，继续由 `plugin:query`及相关治理权限保护；首屏列表减少敏感治理细节暴露，详情和动作路径继续按权限校验。
- 数据库：预期不新增 SQL 或索引；若实现中引入新的持久化查询路径，再按数据库规则追加幂等 SQL 和索引评估。
- `i18n`：预期不新增用户可见文案；若实现中新增 loading、错误、按钮或表格文案，必须同步宿主前端语言包和 API 文档本地化资源。
