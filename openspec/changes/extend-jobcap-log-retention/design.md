# 设计说明

## 能力归属

`Jobs()` 是插件可消费的宿主定时任务领域能力，契约 owner 位于 `apps/lina-core/pkg/plugin/capability/jobcap`。本次变更扩展该领域能力的输入值对象和只读投影，不新增领域方法、不新增运行期依赖，也不改变定时任务 owner。

## 发布矩阵

| 项目 | 决策 |
| --- | --- |
| Go 领域契约 | 同步发布，`jobcap.SaveInput` 增加 `LogRetentionOverride` |
| Go 查询投影 | 同步发布，`jobcap.JobInfo` 增加 `LogRetentionOverride` |
| 源码插件 | 同步发布，通过 `pluginhost.Services().Jobs()` 消费 |
| 动态插件 | 同步发布，复用现有 `service: jobs` 的 `jobs.batch_get`、`jobs.list`、`jobs.visible.ensure`、`jobs.create`、`jobs.update`、`jobs.delete`、`jobs.run`、`jobs.status.set` 和 `jobs.register` |
| 新增动态方法 | 不新增 |
| 新增授权资源 | 不新增新的 host service family、wire method 或资源边界，动态示例插件补齐既有 `jobs.*` 方法声明 |
| 数据库结构 | 不变，继续使用 `sys_job.log_retention_override` |
| 规模上限 | `BatchGet` 复用既有 ID 上限，`List` 复用既有分页上限 |

## 字段语义

新增输入字段和查询投影语义与宿主 HTTP API 的 `logRetentionOverride` 保持一致：

- `nil` 表示不设置任务级覆盖，任务日志跟随系统参数 `sys.cron.log.retention`。
- `{mode:"days",value:N}` 表示按天数保留，`N` 必须大于 `0`。
- `{mode:"count",value:N}` 表示按条数保留，`N` 必须大于 `0`。
- `{mode:"none",value:0}` 表示该任务不按任务级策略清理日志。

字段校验复用 `jobmgmt.normalizeRetentionOptionJSON` 和 `jobmeta.RetentionOption` 既有路径，查询投影复用 `jobmeta.ParseRetentionOption` 解析既有持久化 JSON，错误语义不新增。

## 数据权限与性能

创建任务仍写入当前租户和当前调用者上下文；更新任务仍先校验目标任务可见性后写入。查询任务继续复用现有 `Get`、`BatchGet` 和 `List` 可见性过滤，只在已有任务投影中增加 `log_retention_override` 字段，不改变分页或批量上限，不增加逐行数据库访问，也不产生 `N+1` 风险。

## 影响判断

- `i18n`：动态示例插件 `plugin.yaml` 仅补齐机器可读 host service 方法值，README 仅同步技术说明；无运行时 UI、API 文档源文本或语言包资源变更。
- 缓存一致性：无缓存读写或失效路径变更。
- 数据权限：复用现有任务创建、更新和查询可见性边界，无新增例外。
- 开发工具跨平台：不修改脚本、构建、代码生成或 CI。
- 测试：补充 Go 单元测试，覆盖策略从 `jobcap` 输入传递到宿主 owner、查询投影返回给源码插件能力调用方、动态 host service JSON 调用、源码示例插件和动态示例插件完整 `jobcap.Service` 方法使用，以及动态示例插件 `jobs.*` 授权声明。
