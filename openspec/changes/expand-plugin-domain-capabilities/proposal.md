## Why

插件开发已经形成`capability.Services`、源码插件`pluginhost.Services`和动态插件`hostServices`三层能力边界，但当前普通领域能力仍缺少部分稳定的当前上下文投影、批量解析和批量权限判断方法。若不提前冻结方法发布、错误语义、规模上限和动态授权矩阵，后续业务插件容易反复要求修改`lina-core`或绕过领域能力直接读取宿主核心表。

`localdocs/plugin-domain-capability-expansion-design.md`的方案边界已经足够清晰，无必须回问的需求缺口；需要推进为正式 OpenSpec 变更，并以第一批高频只读能力完成可验证实现。

## What Changes

- 创建`expand-plugin-domain-capabilities`变更，正式冻结插件领域能力扩展的阶段 0 门禁。
- 将第一批实现范围限定为高频只读、批量解析和校验能力：
  - `Users.Current`：返回当前 actor 用户投影。
  - `Users.BatchResolve`：按用户 ID、用户名、手机号或邮箱批量解析可见用户。
  - `Auth().Authz().BatchHasPermissions`：一次判断多个权限 key。
  - `Dict.EnsureValuesVisible`：从内部 scope 能力提升到普通消费面，保持现有实现语义。
  - `Sessions.Current`：返回当前 token 对应在线会话投影。
- 校准已有动态`hostServices`能力，避免重复新增已存在的`dict.labels.resolve`、`authz.permissions.batch_get`、`users.batch_get`等语义。
- 同步扩展动态插件 host service catalog、guest client、WASM dispatcher、治理测试和`apps/lina-core/pkg/plugin`中英文 README。
- 明确候选搜索、组织/租户投影、插件治理搜索、私有资源批量、通知和`AI`状态等能力仍属于后续阶段，不进入本次实现批次。
- 不新增 HTTP API、前端页面、SQL schema、插件清单或运行时用户可见文案。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `plugin-host-domain-capabilities`：补充插件领域能力扩展的阶段 0 冻结矩阵、阶段 1 方法清单、动态发布矩阵和高频能力性能/权限语义。
- `pluginbridge-subcomponent-architecture`：要求本次新增的动态普通领域方法继续使用公开`protocol/hostservices`catalog 和统一 JSON envelope，并由 catalog、guest client、dispatcher 覆盖测试阻断同步遗漏。
- `plugin-capability-boundary-governance`：明确普通能力补充不得扩大宿主核心表、Provider SPI、管理命令或动态`data`服务边界。

## Impact

- 影响代码：`apps/lina-core/pkg/plugin/capability`领域契约、`internal/service/plugin/internal/capabilityhost`宿主适配器、`pkg/plugin/pluginbridge/internal/domainhostcall`动态 guest client、`pkg/plugin/pluginbridge/protocol/hostservices`catalog、`internal/service/plugin/internal/wasm`dispatcher 和相关测试。
- 影响文档：`apps/lina-core/pkg/plugin/README.md`、`README.zh-CN.md`和本 OpenSpec 变更文档。
- 影响协议：动态插件普通领域 host service 方法目录会扩展，但只扩展已冻结的阶段 1 方法。
- 数据权限：所有新增读取、解析和校验方法必须在查询或服务 owner 阶段应用租户、actor 和数据权限边界。
- 缓存一致性：权限、session、字典和用户投影必须复用现有 owner 共享实例、快照或修订机制；不得在 adapter 或 dispatcher 中新增本地权威缓存。
- 数据库：本次默认不新增 SQL；如实现发现需要新增搜索索引或字段，必须先更新设计并按数据库规则补充 SQL 与验证。
