## Context

`apps/lina-core/pkg/plugin/capability`已经是源码插件和动态插件共享的领域能力契约层，`internal/service/plugin/internal/capabilityhost`负责宿主实现适配，`pkg/plugin/pluginbridge/protocol/hostservices`和`internal/service/plugin/internal/wasm`负责动态插件 transport。`localdocs/plugin-domain-capability-expansion-design.md`已经给出长期路线图，本变更只把其中阶段 0 和阶段 1 第一批落为正式 OpenSpec。

当前仓库仍存在一个已完成但未归档的活跃变更`move-plugin-codegen-config-to-root-hack`。该变更聚焦插件本地工具配置和构建聚合，不新增普通`capability`方法，也不扩展新的动态领域`host service`目录；其旧`config/cron`快照迁移不影响本变更的用户、权限、字典和会话能力语义。因此本变更可以并行创建，但实现和审查必须继续把该活跃变更视为当前基线。

当前代码基线已有部分阶段 1 能力：

| 领域 | 当前已有能力 | 本次处理 |
| --- | --- | --- |
| `Users` | `BatchGet`、`Search`、`EnsureVisible`，动态协议已有`users.batch_get`、`users.search`、`users.visible.ensure` catalog 声明，但 dispatcher 仅覆盖`users.batch_get`。 | 新增`Current`和`BatchResolve`，并补齐对应动态发布路径。 |
| `Authz` | `BatchGetPermissions`、`HasPermission`、`IsPlatformAdmin`，动态 dispatcher 已覆盖`permissions.batch_get`和`permissions.has`。 | 新增`BatchHasPermissions`，避免插件循环调用`HasPermission`。 |
| `Dict` | `ResolveLabels`普通消费面，`EnsureValuesVisible`目前只在`ScopeService`中。动态协议已有`labels.resolve`。 | 将`EnsureValuesVisible`提升到普通消费面，并同步动态发布为`values.visible.ensure`。 |
| `Sessions` | `Search`、`BatchGet`，动态协议 catalog 有`search`和`batch_get`，dispatcher 仅覆盖`sessions.batch_get`。 | 新增`Current`，返回当前 token 在线会话投影。 |

## Goals / Non-Goals

**Goals:**

- 冻结首批方法的发布矩阵、错误语义矩阵、规模上限矩阵和动态授权资源矩阵。
- 新增或校准阶段 1 方法，使普通源码插件和动态插件能够复用同一领域语义。
- 动态发布只覆盖冻结方法，并同步 catalog、protocol alias、guest client、dispatcher、测试和 README。
- 所有读取、批量解析和校验方法保持有界输入、批量装配和数据权限语义，避免`N+1`和存在性泄露。
- 任务记录明确`i18n`、缓存一致性、数据权限、数据库、开发工具跨平台和 E2E 影响判断。

**Non-Goals:**

- 不实现候选搜索、组织/租户投影、插件治理搜索、资源批量、通知类型化和`AI`方法状态等后续阶段。
- 不新增 HTTP API、前端页面、E2E 用例、SQL schema、插件清单或语言包。
- 不开放动态`AdminServices`、Provider SPI、宿主核心表、官方能力插件内部表或动态`data`服务核心表授权。
- 不恢复动态`i18n`host service。
- 不保留旧动态方法别名；全新项目无兼容负担，方法名以本变更冻结矩阵为准。

## Decisions

### 阶段 0 先冻结，阶段 1 只补缺口

本次实现不按`localdocs`路线图一次性增加所有方法。阶段 0 在 OpenSpec 规范中冻结四类矩阵；阶段 1 只实现当前用户、批量用户解析、批量权限判断、字典值可见性校验和当前会话。这样可以让每个方法都有明确数据权限、错误、规模和动态授权边界。

备选方案是直接扩展所有领域候选和搜索方法。该方案会把组织树、文件候选、任务状态、租户搜索和插件治理读模型混入同一批实现，难以证明不是工作台展示需求，也难以一次性验证索引、缓存和数据范围。

### 复用已有领域能力，不做重复命名

已有`Dict.ResolveLabels`就是方案中的`BatchResolveLabels`语义，本次不新增重复方法名。已有`Authz.BatchGetPermissions`继续保留为权限投影批量读取；新增`BatchHasPermissions`只表达多权限布尔判断，避免循环调用单权限判断。

备选方案是强制重命名现有方法以匹配设计文档术语。该方案会增加无业务收益的协议迁移和 README churn。

### 动态方法只发布冻结方法

动态发布矩阵如下：

| Go 方法 | 动态 service | 动态 method | 发布决策 | 资源授权 |
| --- | --- | --- | --- | --- |
| `Users.Current` | `users` | `users.current.get` | 同步动态发布 | `host:users` |
| `Users.BatchResolve` | `users` | `users.resolve.batch` | 同步动态发布 | `host:users` |
| `Authz.BatchHasPermissions` | `authz` | `permissions.batch_has` | 同步动态发布 | `host:authz` |
| `Dict.EnsureValuesVisible` | `dict` | `values.visible.ensure` | 同步动态发布 | `host:dict` |
| `Sessions.Current` | `sessions` | `sessions.current.get` | 同步动态发布 | `host:sessions` |
| `Users.Search`扩展、`Dict.ListValues`、`Files.Search`、`Jobs.Search`、`Sessions.BatchGetUserOnlineStatus` | 多个 | 多个 | 延后发布 | 后续冻结 |

动态方法使用现有普通领域 JSON envelope，不新增每领域专用 protowire codec。catalog 是唯一公开描述源，内部 descriptor、guest client 和 dispatcher 覆盖测试必须证明同步。

### 错误语义保持 fail-closed

阶段 1 默认错误矩阵如下：

| 场景 | 决策 |
| --- | --- |
| 方法需要当前用户 actor，但`CapabilityContext`缺少用户 actor | 返回`CodeCapabilityActorRequired`或等价`bizerr`包装错误。 |
| 系统上下文调用`Users.Current`或`Sessions.Current` | 返回`CodeCapabilityContextRequired`或`CodeCapabilityActorRequired`，不伪造用户或会话。 |
| 不存在、不可见、租户外或未授权的批量解析目标 | 进入`MissingIDs`，不区分具体原因。 |
| `EnsureValuesVisible`任一目标不可见或不存在 | 整体拒绝并返回`CodeCapabilityDenied`或领域现有等价错误。 |
| 输入数量或分页超限 | 返回`CodeCapabilityLimitExceeded`。 |
| 动态方法未授权 | dispatcher 在进入领域服务前拒绝，不泄露目标资源存在性。 |

### 规模上限使用领域常量

阶段 1 不新增全局万能上限。每个领域包定义自己的上限常量和 adapter 归一化逻辑：

| 方法 | 上限 |
| --- | --- |
| `Users.BatchResolve` | 每次最多 100 个 ID、100 个用户名、100 个手机号或邮箱，合计去重后不超过 300 个解析键。 |
| `Authz.BatchHasPermissions` | 每次最多 200 个权限 key。 |
| `Dict.EnsureValuesVisible` | 每个字典类型最多 200 个值。 |
| `Users.Current`、`Sessions.Current` | 单对象投影，无分页。 |

如实现发现现有包已定义更严格上限，以更严格上限为准，并在任务记录中说明。

### 数据权限和缓存复用 owner

用户解析必须复用现有用户 adapter 的租户过滤和`datascope.Service`用户范围过滤，不能先查全量再内存过滤。权限判断必须复用现有授权快照或权限服务，不得循环调用`HasPermission`装配结果。字典校验复用字典 owner 的可见性逻辑和字典缓存，不新增 adapter 本地缓存。当前会话复用启动期注入的共享 session store 或等价 owner 实例，不创建新的 session 缓存。

## Risks / Trade-offs

- 方法名与`localdocs`术语不完全一致 → 以当前代码已有契约为基线，OpenSpec 明确哪些方法是复用语义，避免重复接口。
- 动态协议同步点遗漏 → 通过 catalog、protocol alias、guest client、dispatcher 和 README 覆盖测试阻断。
- 用户批量解析可能诱导模糊枚举 → 解析结果仍按当前 actor 租户和数据权限过滤，不可见进入`MissingIDs`，不返回不存在与不可见差异。
- 批量权限判断误用单项循环 → 单元测试或静态审查必须证明实现使用快照/集合化路径，不循环调用`HasPermission`作为常规装配。
- 当前会话查找依赖 token 上下文 → 若当前请求上下文没有 token/session 标识，方法 fail-closed；不扫描在线会话表猜测当前用户最新会话。

## Migration Plan

1. 创建并校验本 OpenSpec 变更文档和增量规范。
2. 更新`capability`领域接口、DTO、常量和注释。
3. 更新`capabilityhost` adapter，复用现有共享服务实例和数据权限边界。
4. 更新动态`hostServices` catalog、protocol 常量、guest client、WASM dispatcher 和覆盖测试。
5. 更新`apps/lina-core/pkg/plugin`双语 README 的 host service 表和能力说明。
6. 运行 OpenSpec strict 校验、相关 Go 包测试、启动绑定包烟测和文档/静态检索验证。
7. 执行`lina-review`，确认规则域、任务状态和验证证据满足门禁。

## Open Questions

无必须用户澄清的问题。后续阶段的候选搜索、组织/租户投影、插件治理搜索和资源批量方法需要在进入对应阶段前另行冻结过滤字段、排序字段、索引路径、授权资源和错误语义。
