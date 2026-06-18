## Context

`expand-plugin-domain-capabilities`已经完成阶段 0 冻结和阶段 1 第一批能力，当前基线包含`Users.Current`、`Users.BatchResolve`、`Authz.BatchHasPermissions`、`Dict.EnsureValuesVisible`和`Sessions.Current`，并已同步动态`hostServices`、WASM dispatcher、guest client、测试和 README。`localdocs/plugin-domain-capability-expansion-design.md`的剩余内容仍覆盖阶段 1.5 至阶段 5，包括候选搜索、组织/租户/插件治理投影、插件私有资源批量、动态 runtime state、多键 cache、manifest 批量、通知类型化和`AI`状态。

当前仓库还有多个已完成但未归档的活跃 OpenSpec 变更：`expand-plugin-domain-capabilities`、`reuse-goframe-shutdown-config`和`move-plugin-codegen-config-to-root-hack`。本变更以这些已完成变更作为当前工作区基线，不回滚、不重写其任务记录；实现时必须继续隔离无关工作区改动，避免把 shutdown/config/hack deploy 相关改动混入本变更审查范围。

本变更触及插件能力契约、动态协议、宿主 adapter、WASM host service、数据权限、缓存一致性和 README 文档。它不新增工作台 UI、静态 HTTP API 或插件业务目录；若实现中发现必须新增 SQL、API 或 UI，应先更新本设计和任务，再按对应规则补充验证。

## Goals / Non-Goals

**Goals:**

- 将`localdocs`阶段 1.5 至阶段 5 中可稳定冻结的能力全部转化为 OpenSpec 契约、实现任务和验证门禁。
- 继续保持`capability.Services`作为普通领域能力 owner，动态`hostServices`只作为 transport 和授权发布层。
- 搜索、候选、树形数据、批量投影、按来源读取和批量私有资源操作必须具备规模上限、数据库侧过滤或 provider 批量路径，避免`N+1`和存在性泄露。
- 动态发布的方法必须同步 catalog、协议常量、guest client、dispatcher、测试和 README。
- 插件私有资源批量方法必须继续使用插件 ID、租户、资源授权和 owner 共享实例，不得开放宿主核心表或跨插件资源。
- 任务记录必须覆盖`i18n`、缓存一致性、数据权限、数据库、开发工具跨平台、测试策略和 E2E 影响判断。

**Non-Goals:**

- 不开放动态`AdminServices`、Provider SPI、宿主核心表、官方能力插件内部表、原始 SQL、JOIN 或任意表达式查询。
- 不恢复动态`i18n`host service。
- 不实现`data.exists`和`data.count`；二者必须另开小设计冻结过滤 DSL、字段白名单、软删除、租户和数据权限注入、聚合不泄露策略后再决定。
- 不实现无边界`Storage.DeletePrefix`、缓存 scan/list、runtime state scan/list 或跨插件资源枚举。
- 不通过本变更新增默认工作台页面、路由、表单、表格或 E2E 用例；若实际实现触发用户可观察 UI，再按前端和 E2E 规则扩展任务。

## Decisions

### 分批实现但同一变更内保证技术方案闭环

本变更覆盖剩余全部阶段，但任务执行按阶段 1.5、阶段 2、阶段 3、阶段 4、阶段 5 顺序推进。每一批先完成契约和动态发布矩阵，再实现 adapter/dispatcher/client/test，避免同时修改所有领域导致协议漂移无法定位。

备选方案是继续为每个阶段创建独立 OpenSpec 变更。该方案能降低单次变更大小，但与当前用户“保证技术方案内容被全部实现完成”的目标不匹配，会让技术方案完成状态跨多个活跃变更漂移。本变更仍可在任务层按阶段拆分验证，降低审查风险。

### 只把稳定跨插件能力推入核心宿主契约

候选搜索、树形数据、批量状态和批量私有资源能力只有在具备跨插件通用场景、稳定 DTO、上限和数据权限语义时才进入`capability.Services`或动态`hostServices`。如果某个实现点实际只服务工作台展示，例如某个表格列或下拉字段组合，则该点应停在工作台适配层，不进入核心领域契约。

备选方案是按`localdocs`方法表机械实现所有方法。该方案会把部分展示适配需求混入`lina-core`核心契约，违反宿主通用性边界。

### 动态普通领域方法继续使用 JSON envelope

新增普通领域方法默认使用`HostServiceJSONRequest`和`HostServiceJSONResponse`。已有专用 codec 的资源型能力，例如 storage、data、manifest、runtime state 或`AI`，可以复用既有专用 payload 模式；不得为了单个普通领域方法新增 per-domain codec。catalog 仍是公开描述源，测试必须覆盖 catalog、public alias、guest client 和 dispatcher 同步。

备选方案是为高频能力提前设计专用 wire codec。当前缺少性能证据证明 JSON envelope 成为瓶颈，新增 codec 会扩大同步面和测试成本。

### 数据权限在 owner 或 provider 批量路径完成

用户、文件、任务、在线会话、通知、插件治理和租户查询必须在数据库查询阶段或 owner 批量服务阶段注入租户和数据权限。组织 provider 能力禁用时返回中性空投影、空页或 unavailable，禁止放开全量组织数据。批量校验类方法默认任一目标不可见整体拒绝；批量读取类方法使用`MissingIDs`隐藏不存在与不可见差异。

备选方案是查询后在 adapter 中内存过滤。该方案容易泄露分页总数、候选数量或排序差异，且不满足数据权限规则。

### 缓存和运行时状态只复用现有 owner

本变更不新增权威缓存。权限、字典、插件状态、租户、组织、session、manifest、`AI` provider 状态复用既有 owner、共享实例、修订号或共享后端。`Cache.GetMany/SetMany/DeleteMany`和 runtime state 多键方法只减少 host call 次数，不改变缓存或 runtime state 的非权威/插件私有语义。

备选方案是在 adapter 或 dispatcher 中维护本地聚合缓存。该方案在集群模式下容易退化为节点本地状态，违反缓存一致性规则。

### `Storage`provider 选择保持零配置

`Storage`provider 选择不进入`config.yaml`或宿主主配置契约。插件使用者只依赖`Storage`接口；项目管理员通过安装并启用 storage provider 插件改变后端实现。宿主在已注册 provider 中选择唯一一个平台可服务插件；没有可服务 provider 插件时使用内置本地文件 provider；多个 provider 插件同时可服务时拒绝请求，避免 durable storage 静默落到不确定后端。

备选方案是在主配置中维护`activeProviderPluginId`或`allowLocalProviderInCluster`。该方案会让宿主配置理解具体插件实现方，并形成配置状态和插件治理状态两套选择来源，增加模板、文档、测试和故障诊断成本。

### `data.batch_get`只补插件自有表批量读取

本批只实现`data.batch_get`，用于动态插件在已授权插件自有表中按主键集合读取记录。它必须复用现有`data`资源授权、表白名单、租户过滤、字段投影和 record store plan 约束。`exists`和`count`保留为后续独立设计候选。

备选方案是一次性补齐`exists`和`count`。这会引入聚合统计不泄露、过滤 DSL 表达能力、字段白名单和软删除语义的新风险，不能在本批安全完成。

### 通知类型化采用破坏性收敛，不保留旧动态别名

全新项目无兼容负担。`messages.batch_get`应返回稳定`notifycap.MessageProjection`或等价类型化 DTO，而不是`map[string]any`。若旧代码已有 map 形态，实施时直接收敛到类型化 DTO，并同步 guest client、dispatcher 和 README。

备选方案是保留旧 map 响应并新增 parallel 方法。该方案会增加长期协议负担，且与全新项目无历史兼容负担的顶层要求不一致。

### README 和治理测试作为每批完成门禁

每个动态发布批次完成后必须同步`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`，并更新 host service catalog/guest/dispatcher 覆盖测试。阶段 5 不作为“只写文档”的收尾，而是验证此前每批没有遗漏 README、协议和治理测试同步。

备选方案是全部实现后一次性更新文档。该方案会在中途留下可调用协议和文档不一致的窗口，难以及时审查。

## Risks / Trade-offs

- 变更范围大 → 按阶段任务拆分，每批完成后运行对应包测试和 OpenSpec 校验；发现无法冻结的方法时更新设计并记录不实现边界。
- 搜索/候选路径可能缺少索引 → 优先复用现有有索引查询；如发现新增索引需求，先补 SQL 设计和数据库验证，不在 Go 代码中绕过性能风险。
- provider 能力禁用语义不一致 → 在 spec 中逐领域规定空投影、空页、unavailable 或拒绝策略，并补单元测试。
- 动态协议同步点多 → 通过 catalog、alias、guest client、dispatcher、README 覆盖测试和静态检索阻断遗漏。
- `Cache`和 runtime state 多键写入失败语义复杂 → 本批默认整体拒绝或原子批量；如 owner 不支持原子性，必须在任务记录中明确逐项语义并测试。
- `data.batch_get`误变成核心表访问入口 → 静态检索和 hostservice/datahost 测试必须证明核心表授权继续失败，插件只能访问自有表。

## Migration Plan

1. 创建并校验本 OpenSpec 变更文档和增量规范，记录与已完成阶段 1 变更的关系。
2. 阶段 1.5：实现候选、搜索和可见性补充能力，并同步动态发布路径和测试。
3. 阶段 2：实现组织、租户和插件治理投影，覆盖 provider 禁用/缺失、批量装配和分页边界。
4. 阶段 3：实现 storage/cache/manifest/runtime state/data 批量能力，覆盖资源授权、插件/租户作用域和共享后端复用。
5. 阶段 4：实现通知类型化和`AI`方法状态，覆盖 provider 状态、不可用和权限边界。
6. 阶段 5：同步 README、治理测试、静态检索、任务记录和完整验证。
7. 每批完成后执行必要 Go 编译门禁；最终运行 OpenSpec strict、目标 Go 包测试、启动绑定 smoke、`git diff --check`和`lina-review`。

## Open Questions

无必须用户澄清的问题。若实现阶段发现某个领域 owner 当前缺少安全的批量能力、必要索引或 provider 禁用语义，必须先更新本设计和任务记录，再决定补底层 owner、补 SQL，或把该方法标记为无法在本批安全实现。
