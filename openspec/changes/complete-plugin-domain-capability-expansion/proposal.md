## Why

`expand-plugin-domain-capabilities`已经完成`localdocs/plugin-domain-capability-expansion-design.md`中的阶段 0 和阶段 1 第一批能力，但该技术方案仍有候选搜索、组织/租户/插件治理投影、插件私有资源批量、通知类型化和`AI`状态等后续阶段未落地。若不继续以 OpenSpec 方式冻结并实现剩余阶段，插件开发仍会在高频列表装配、批量可见性校验、动态私有资源操作和 provider 状态降级上反复修改`lina-core`或绕过领域能力。

本变更继续推进该技术方案的剩余内容，目标是把所有可稳定冻结的后续能力纳入正式契约、动态发布矩阵、实现任务和验证门禁；对仍需要独立领域设计的`data.exists`、`data.count`和管理命令动态发布保持不实现并明确阻断边界。

## What Changes

- 基于已完成的`expand-plugin-domain-capabilities`作为阶段 1 基线，继续完成`localdocs`阶段 1.5 至阶段 5 的剩余能力。
- 补充候选、搜索和可见性能力：
  - 扩展`Users.Search`的稳定过滤和动态路径。
  - 新增`Dict.ListValues`候选分页。
  - 新增`Files.Search`、`Jobs.Search`、`Jobs.EnsureVisible`、`Sessions.BatchGetUserOnlineStatus`和`Sessions.EnsureVisible`。
- 补充组织、租户和插件治理投影能力：
  - `Org`批量组织档案、受限部门树、部门/岗位候选和可见性校验。
  - `Tenant`当前租户详情、租户批量投影、租户搜索、批量用户租户投影和批量可见性校验。
  - `Plugins`当前插件投影、插件搜索、分页租户插件列表和能力状态批量读取。
- 补充插件私有资源和动态专属批量能力：
  - `Storage.BatchStat`、`Storage.ListCursor`和`Storage.DeleteMany`。
  - `Cache.GetMany`、`Cache.SetMany`和`Cache.DeleteMany`。
  - `Manifest.GetMany`和`Manifest.List`。
  - 动态`Runtime state.get_many`、`state.set_many`和`state.delete_many`。
  - 动态`data.batch_get`仅限插件自有表和既有`data`资源授权，不开放宿主核心表。
- 补充通知和`AI`体验能力：
  - 将通知消息读取收敛为类型化`MessageProjection`，补充按来源批量读取和可见性校验。
  - 补齐`AI`文本方法级状态和跨子能力状态批量读取。
- 同步更新`protocol/hostservices`catalog、协议常量、guest client、WASM dispatcher、测试、`apps/lina-core/pkg/plugin`中英文 README 和任务记录。
- 明确不实现或单独延后：
  - 不开放动态`AdminServices`、Provider SPI、宿主核心表、官方能力插件内部表或动态`i18n`host service。
  - 不实现无边界`Storage.DeletePrefix`、缓存 scan/list、跨插件表查询、JOIN、原始 SQL 或任意表达式查询。
  - `data.exists`和`data.count`必须另开小设计冻结过滤 DSL、字段白名单、软删除、租户和数据权限注入后再决定是否实现。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `plugin-host-domain-capabilities`：补充阶段 1.5 至阶段 5 的方法发布矩阵、错误语义、规模上限、授权资源、数据权限和缓存复用要求。
- `pluginbridge-subcomponent-architecture`：要求新增动态普通领域方法、插件私有资源批量方法和动态专属 runtime state 方法继续由公开 catalog、统一 envelope 或既有专用 codec 统一治理，并由 guest client 与 dispatcher 覆盖测试阻断同步遗漏。
- `plugin-capability-boundary-governance`：扩大普通领域能力的覆盖范围但不得扩大宿主核心数据边界、管理命令边界、Provider SPI 边界或动态`data`核心表访问边界。
- `plugin-storage-service`：补充插件私有对象批量元数据、游标列表和批量删除契约。
- `plugin-cache-service`：补充插件私有缓存多键读写删契约，并保持非权威缓存语义。
- `plugin-data-service`：补充插件自有表`batch_get`契约，并明确`exists`和`count`不进入本批实现。
- `plugin-notify-service`：补充类型化消息投影、按来源批量读取和消息可见性校验。
- `framework-ai-capability-namespace`、`framework-ai-text-capability`：补充文本方法级状态和跨子能力状态批量读取要求。

## Impact

- 影响代码：`apps/lina-core/pkg/plugin/capability`领域契约、`internal/service/plugin/internal/capabilityhost`宿主适配器、`pkg/plugin/pluginbridge/internal/domainhostcall`动态 guest client、`pkg/plugin/pluginbridge/protocol/hostservices`catalog、`internal/service/plugin/internal/wasm`dispatcher、动态 runtime/data/storage/cache/manifest/notify/AI host service 和相关测试。
- 影响文档：`apps/lina-core/pkg/plugin/README.md`、`README.zh-CN.md`和本 OpenSpec 变更文档。
- 影响协议：动态插件`hostServices`目录继续扩展；全新项目无兼容负担，但必须通过 catalog、protocol alias、guest client、dispatcher 和 README 测试保持同步。
- 数据权限：搜索、候选、批量详情、树形数据、聚合状态、按来源批量读取和可见性校验必须在查询或 provider 阶段应用租户、actor 和数据权限边界。
- 缓存一致性：插件状态、字典、租户、组织、session、cache、runtime state、manifest 和`AI` provider 状态必须复用既有 owner、共享实例、共享后端、修订号或失效机制；不得在 adapter 或 dispatcher 中新增本地权威缓存。
- 数据库：默认优先复用现有表和索引；若实现发现新增搜索、排序、聚合或候选路径缺少必要索引，必须先更新设计并按数据库规则补充 SQL、`make db.init`、`make dao`和验证记录。
- 前端与 E2E：本变更默认不新增管理工作台页面或用户可观察 UI；如后续任务实际修改页面、demo 插件或表单候选交互，必须补充对应 E2E 并使用`lina-e2e`技能。
