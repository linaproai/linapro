## 1. 基线与冻结矩阵

- [x] 1.1 记录基线状态：`expand-plugin-domain-capabilities`已完成阶段 0/1；本变更接续`localdocs/plugin-domain-capability-expansion-design.md`阶段 1.5 至阶段 5；当前其他已完成未归档变更不与本变更领域协议冲突。
- [x] 1.2 为剩余阶段输出方法发布矩阵、错误语义矩阵、规模上限矩阵和动态授权资源矩阵，覆盖所有同步动态发布、源码插件专属、延后发布和不发布方法。
- [x] 1.3 运行`openspec validate complete-plugin-domain-capability-expansion --strict`，确认提案、设计、规格和任务产物可校验。

## 2. 阶段 1.5：候选搜索和可见性能力

- [x] 2.1 更新`usercap`契约和`capabilityhost`用户 adapter，扩展`Users.Search`稳定过滤，确保数据库侧租户和数据权限过滤、分页上限、字段投影和动态`users.search`同步。
- [x] 2.2 新增`dictcap.ListValues`普通能力、领域 DTO、上限常量和`capabilityhost`实现，支持字典类型候选分页、租户覆盖、状态过滤和动态`dict.values.list`。
- [x] 2.3 新增`filecap.Search`普通能力、领域 DTO、上限常量和`capabilityhost`实现，支持业务场景、名称关键字、MIME 类型和数据权限过滤，并同步动态`files.search`。
- [x] 2.4 新增`jobcap.Search`和`jobcap.EnsureVisible`普通能力、领域 DTO、上限常量和`capabilityhost`实现，并同步动态`jobs.search`和`jobs.visible.ensure`。
- [x] 2.5 新增`sessioncap.BatchGetUserOnlineStatus`和`sessioncap.EnsureVisible`普通能力、领域 DTO、上限常量和`capabilityhost`实现，复用共享`session.Store`批量路径，并同步动态`sessions.users.online.batch_get`和`sessions.visible.ensure`。
- [x] 2.6 为阶段 1.5 更新`pluginbridge/internal/domainhostcall`guest client、`protocol/hostservices`catalog、protocol alias、WASM dispatcher、注册表和覆盖测试。

## 3. 阶段 2：组织、租户和插件治理投影

- [x] 3.1 更新`orgcap`契约，新增批量用户组织档案、受限部门树、部门搜索、岗位候选、部门可见性校验和岗位可见性校验，保持 provider SPI 与普通消费面分离。
- [x] 3.2 实现`capabilityhost`组织 adapter 或 provider adapter，覆盖 provider 缺失/禁用安全降级、批量装配、节点上限、分页和数据权限边界。
- [x] 3.3 更新`tenantcap`契约，新增当前租户详情、批量租户投影、租户搜索、批量用户租户列表和批量租户可见性校验。
- [x] 3.4 实现`capabilityhost`租户 adapter，覆盖租户可见性、平台租户中性投影、批量装配、分页上限和数据权限边界。
- [x] 3.5 更新`plugincap`契约，新增当前插件投影、插件搜索、分页租户插件列表和能力状态批量读取。
- [x] 3.6 实现`capabilityhost`插件治理 adapter，复用插件治理读模型、enabled snapshot 或修订机制，避免重复扫描 manifest 和`N+1`状态读取。
- [x] 3.7 为阶段 2 更新动态 host service catalog、protocol alias、guest client、WASM dispatcher 和同步覆盖测试。

## 4. 阶段 3：插件私有资源和动态专属批量能力

- [x] 4.1 更新`storagecap`契约和动态 storage host service，实现`BatchStat`、`ListCursor`、`DeleteMany`，覆盖路径数量、路径长度、总字节数、资源授权、插件/租户作用域和 provider 错误语义。
- [x] 4.2 更新`cachecap`契约和动态 cache host service，实现`GetMany`、`SetMany`、`DeleteMany`，复用共享缓存后端、命名空间隔离、TTL 语义和集群模式边界。
- [x] 4.3 更新`manifestcap`契约和动态 manifest host service，实现`GetMany`和`List`，覆盖路径授权、前缀限制、总字节数和不可见资源缺失语义。
- [x] 4.4 更新动态 runtime state host service，实现`state.get_many`、`state.set_many`、`state.delete_many`，复用插件/租户作用域和共享 runtime state 存储。
- [x] 4.5 更新动态 data host service，实现`data.batch_get`，仅支持插件自有表、已授权资源表、字段投影、租户过滤、软删除和不透明缺失集合；确认`data.exists`和`data.count`未发布。
- [x] 4.6 为阶段 3 更新 protocol codec、public alias、guest client、recordstore 或 capability guest 入口、WASM dispatcher 和覆盖测试。

## 5. 阶段 4：通知类型化和 AI 状态

- [x] 5.1 更新`notifycap`契约和动态 notifications host service，将消息读取收敛为类型化`MessageProjection`，并覆盖`messages.batch_get`guest/client/dispatcher 响应。
- [x] 5.2 新增`Notifications.BatchGetBySource`和`Notifications.EnsureVisible`，实现按`SourceType + SourceIDs`集合化读取、可见性校验、数据权限和不透明拒绝语义，并同步动态`messages.by_source.batch_get`和`messages.visible.ensure`。
- [x] 5.3 更新`aicap`命名空间契约，实现跨子能力`MethodStatuses`批量读取，覆盖 provider 缺失、禁用、不可用和不暴露 provider 配置。
- [x] 5.4 更新`aitext`契约，实现`Text().MethodStatus`，并同步动态`text.method_status.get`和`ai.methods.status.batch_get`协议、guest client、dispatcher 和测试。

## 6. README、治理和静态边界

- [x] 6.1 更新`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`，同步普通领域能力说明和动态`hostServices`表，确认中英文事实一致。
- [x] 6.2 更新 catalog、guest client、dispatcher、JSON envelope 或专用 codec 覆盖测试，确保新增方法缺少任一同步点会失败。
- [x] 6.3 静态检索确认未新增动态`i18n`host service、per-domain 普通领域专用 codec、动态`data`核心表授权、Provider SPI 暴露或宿主核心表直接暴露。
- [x] 6.4 在任务记录中写明影响分析：HTTP API、前端 UI、SQL schema、插件清单、语言包、开发工具脚本、E2E、数据权限、缓存一致性、数据库索引和 DI 来源检查。

## 7. 验证与审查

- [x] 7.1 运行`openspec validate complete-plugin-domain-capability-expansion --strict`。
- [x] 7.2 运行覆盖变更包的 Go 编译门禁：`cd apps/lina-core && go test ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm -count=1`。
- [x] 7.3 运行资源型和动态服务相关 Go 测试：`cd apps/lina-core && go test ./internal/service/plugin/internal/datahost ./pkg/plugin/pluginbridge/recordstore ./internal/service/plugin/internal/wasm -count=1`。
- [x] 7.4 运行宿主启动绑定烟测：`cd apps/lina-core && go test ./internal/cmd -count=1`，若因外部依赖阻断则记录替代验证和剩余风险。
- [x] 7.5 运行`git diff --check`和阶段 6 静态边界检索。
- [x] 7.6 执行`lina-review`任务级审查，修复严重问题后再标记实现完成。

### 影响分析初始记录

| 领域 | 初始判断 |
|------|----------|
| 架构 | 有影响。新增跨模块插件能力契约、动态协议和 provider 降级边界。 |
| 插件 | 有影响。动态插件`hostServices`目录和源码插件能力目录继续扩展。 |
| 后端 Go | 有影响。涉及`capability`接口、`capabilityhost`adapter、`pluginbridge`协议、WASM dispatcher 和测试。 |
| API 契约 | 无静态 HTTP API 直接影响；动态`host service`协议有影响。 |
| 数据权限 | 有影响。搜索、候选、树形数据、批量读取、按来源读取和可见性校验必须显式接入。 |
| 缓存一致性 | 有影响。插件状态、cache、runtime state、manifest、session、字典、租户、组织和`AI`状态必须复用 owner 或共享后端。 |
| 数据库 | 当前文档不预设 SQL；实现中若新增索引或查询辅助字段，必须补 SQL、`make db.init`、`make dao`和验证。 |
| `i18n` | 有影响评估。默认无运行时 UI 文案；新增错误码、API 文档源文本、README 或字典文案时必须补治理记录。 |
| E2E | 默认不触发；若实际修改页面、demo 插件或工作台交互，则补 E2E 并使用`lina-e2e`。 |
| 开发工具 | 默认不触发；若新增治理脚本或生成入口，必须按`dev-tooling.md`记录跨平台影响。 |

### 1.1 基线记录

| 项目 | 记录 |
|------|------|
| 阶段 0/1 基线 | `expand-plugin-domain-capabilities`已完成 18/18 个任务，覆盖阶段 0 冻结和阶段 1 第一批能力：`Users.Current`、`Users.BatchResolve`、`Authz.BatchHasPermissions`、`Dict.EnsureValuesVisible`、`Sessions.Current`，并同步动态 catalog、protocol alias、guest client、WASM dispatcher、测试和`apps/lina-core/pkg/plugin`中英文 README。 |
| 本变更范围 | 本变更接续`localdocs/plugin-domain-capability-expansion-design.md`阶段 1.5 至阶段 5，覆盖候选搜索、组织/租户/插件治理投影、插件私有资源批量、动态 runtime state、`data.batch_get`、通知类型化和`AI`状态。 |
| 活跃变更判断 | 当前活跃变更包括`complete-plugin-domain-capability-expansion`、`expand-plugin-domain-capabilities`、`reuse-goframe-shutdown-config`和`move-plugin-codegen-config-to-root-hack`。后两个已完成变更分别聚焦 shutdown 配置复用和插件代码生成配置治理，不新增普通领域能力或动态`hostServices`方法，与本变更领域协议无冲突。 |
| 工作区隔离 | 当前工作区存在与本变更无关的`httpstartup`、`config`、`role`、`manifest/config`和`hack/deploy`改动。本变更实现和审查只覆盖插件能力、动态协议、README、OpenSpec 和相关测试，不回滚无关改动。 |

### 1.2 冻结矩阵

#### 方法发布矩阵

| 阶段 | 方法 | 动态方法 | 决策 |
|------|------|----------|------|
| 1.5 | `Users.Search`扩展 | `users.search` | 同步动态发布，复用现有普通领域方法和 JSON envelope。 |
| 1.5 | `Dict.ListValues` | `dict.values.list` | 同步动态发布，使用 JSON envelope。 |
| 1.5 | `Files.Search` | `files.search` | 同步动态发布，使用 JSON envelope。 |
| 1.5 | `Jobs.Search` | `jobs.search` | 同步动态发布，使用 JSON envelope。 |
| 1.5 | `Jobs.EnsureVisible` | `jobs.visible.ensure` | 同步动态发布，使用 JSON envelope。 |
| 1.5 | `Sessions.BatchGetUserOnlineStatus` | `sessions.users.online.batch_get` | 同步动态发布，使用 JSON envelope。 |
| 1.5 | `Sessions.EnsureVisible` | `sessions.visible.ensure` | 同步动态发布，使用 JSON envelope。 |
| 2 | `Org.BatchGetUserOrgProfiles` | `users.org_profiles.batch_get` | 同步动态发布，使用 JSON envelope，不暴露 provider SPI。 |
| 2 | `Org.ListDeptTree` | `depts.tree.list` | 同步动态发布，使用 JSON envelope，必须有节点上限或懒加载边界。 |
| 2 | `Org.SearchDepartments` | `depts.search` | 同步动态发布，使用 JSON envelope。 |
| 2 | `Org.ListPostOptions` | `posts.options.list` | 同步动态发布，使用 JSON envelope。 |
| 2 | `Org.EnsureDepartmentsVisible` | `depts.visible.ensure` | 同步动态发布，使用 JSON envelope。 |
| 2 | `Org.EnsurePostsVisible` | `posts.visible.ensure` | 同步动态发布，使用 JSON envelope。 |
| 2 | `Tenant.CurrentTenantInfo` | `tenants.current_info.get` | 同步动态发布，使用 JSON envelope。 |
| 2 | `Tenant.BatchGetTenants` | `tenants.batch_get` | 同步动态发布，使用 JSON envelope。 |
| 2 | `Tenant.SearchTenants` | `tenants.search` | 同步动态发布，使用 JSON envelope。 |
| 2 | `Tenant.BatchListUserTenants` | `users.tenants.batch_list` | 同步动态发布，使用 JSON envelope。 |
| 2 | `Tenant.EnsureTenantsVisible` | `tenants.visible.batch_ensure` | 同步动态发布，使用 JSON envelope。 |
| 2 | `Plugins.Current` | `plugins.current.get` | 同步动态发布，使用 JSON envelope。 |
| 2 | `Plugins.Search` | `plugins.search` | 同步动态发布，使用 JSON envelope。 |
| 2 | `Plugins.ListTenantPlugins`分页扩展 | `plugins.tenant.list` | 同步动态发布，允许破坏性收敛为分页响应。 |
| 2 | `Plugins.BatchGetCapabilityStatus` | `plugins.capabilities.status.batch_get` | 同步动态发布，只返回`CapabilityStatus`。 |
| 3 | `Storage.BatchStat` | `storage.stat.batch` | 同步动态发布，复用 storage 资源型 payload 或 JSON envelope。 |
| 3 | `Storage.ListCursor` | `storage.list.cursor` | 同步动态发布，复用 storage 资源型 payload 或 JSON envelope。 |
| 3 | `Storage.DeleteMany` | `storage.delete.batch` | 同步动态发布，复用 storage 资源型 payload 或 JSON envelope。 |
| 3 | `Cache.GetMany` | `cache.get_many` | 同步动态发布，复用 cache host service 语义。 |
| 3 | `Cache.SetMany` | `cache.set_many` | 同步动态发布，复用 cache host service 语义。 |
| 3 | `Cache.DeleteMany` | `cache.delete_many` | 同步动态发布，复用 cache host service 语义。 |
| 3 | `Manifest.GetMany` | `manifest.get_many` | 同步动态发布，复用 manifest host service 语义。 |
| 3 | `Manifest.List` | `manifest.list` | 同步动态发布，复用 manifest host service 语义。 |
| 3 | Runtime state 多键读写删 | `state.get_many`、`state.set_many`、`state.delete_many` | 动态专属同步发布，不进入`capability.Services`普通消费面。 |
| 3 | `data.batch_get` | `data.batch_get` | 动态专属同步发布，只允许当前插件自有授权表。 |
| 3 | `data.exists`、`data.count` | `data.exists`、`data.count` | 延后发布，本变更不实现。 |
| 4 | `Notifications.MessageProjection` | `messages.batch_get` | 同步动态发布，破坏性收敛为类型化响应。 |
| 4 | `Notifications.BatchGetBySource` | `messages.by_source.batch_get` | 同步动态发布，使用 JSON envelope 或既有 notify payload。 |
| 4 | `Notifications.EnsureVisible` | `messages.visible.ensure` | 同步动态发布，使用 JSON envelope。 |
| 4 | `AI.MethodStatuses` | `ai.methods.status.batch_get` | 同步动态发布，使用 JSON envelope 或既有 AI payload。 |
| 4 | `Text.MethodStatus` | `text.method_status.get` | 同步动态发布，使用 JSON envelope 或既有 AI payload。 |
| 全阶段 | 动态`i18n`、`AdminServices`、Provider SPI、管理命令、核心表 data 访问、JOIN、原始 SQL、任意表达式查询、cache/runtime scan/list、storage 无界`DeletePrefix` | 无 | 不发布。 |

#### 错误语义矩阵

| 场景 | 冻结语义 |
|------|----------|
| 需要用户 actor 但当前为系统上下文或缺少 actor | 返回`CodeCapabilityContextRequired`或`CodeCapabilityActorRequired`，不得伪造用户身份。 |
| provider 缺失或禁用且方法可安全中性降级 | 返回空页、空投影、中性平台租户投影或`CapabilityStatus.Available=false`，不得放开全量数据。 |
| provider 是方法必要依赖且无法安全中性降级 | 返回`CodeCapabilityUnavailable`。 |
| 多个 provider 同时可服务 | 返回`CodeCapabilityProviderConflict`，能力状态 reason 使用`provider_conflict`。 |
| 不存在、不可见、租户外或未授权的批量读取目标 | 统一进入`MissingIDs`或等价缺失集合，不区分具体原因。 |
| 写入、删除、执行或可见性校验中任一目标不可见 | 整体拒绝并返回`CodeCapabilityDenied`，除非具体方法记录 no-op 缺失删除语义。 |
| 批量数量、分页、节点数、path/key 长度或总字节数超限 | 返回`CodeCapabilityLimitExceeded`或领域已有结构化超限错误。 |
| 动态插件声明未授权`hostServices`、资源路径、表或资源引用 | 返回`CodeCapabilityDenied`或安装/启用治理失败，不暴露资源是否存在。 |
| cache 后端、runtime state 后端、storage provider 或 manifest 资源不可用 | 复用单键方法现有结构化错误；不得临时创建节点本地默认实例。 |
| 通知、租户、组织、插件治理和`AI`状态读取 | 只返回类型化投影或结构化 unavailable，不泄露内部存储、provider 配置、密钥或模型路由。 |

#### 规模上限矩阵

| 类别 | 冻结上限 |
|------|----------|
| 普通分页搜索和候选 | 默认复用`capabilityhost.DefaultPageSize=20`和`MaxPageSize=200`；具体领域可定义更窄上限。 |
| 普通批量 ID、key、source ID | 默认每方法定义领域级`MaxBatchSize`，优先使用`100`至`300`区间；阶段 1 已有上限继续沿用。 |
| 组织树 | 必须提供节点总数上限、懒加载父节点边界或 cursor，禁止无界整树返回。 |
| storage path | 复用现有`storagecap.MaxLogicalPathBytes=512`作为单路径基线；批量路径数和总 path 字节数在`storagecap`落常量。 |
| cache key 和 runtime state key | 必须定义 key 数量、单 key 字节数和总 value 字节数上限；不提供 scan/list。 |
| manifest path 和内容 | 必须定义路径数量、前缀范围、单资源字节数和总字节数上限；`List`只返回元数据或路径投影。 |
| `data.batch_get` | 必须定义主键数量上限和字段投影数量上限，数据库侧按授权表、租户和软删除过滤。 |
| 通知按来源读取 | `SourceIDs`和返回消息数均必须有上限，按集合查询并稳定排序。 |
| `AI`方法状态批量 | 子能力方法数量必须有上限，只返回状态投影。 |

#### 动态授权资源矩阵

| 方法组 | `hostServices`授权资源 | 资源约束 |
|--------|-----------------------|----------|
| `users.*` | `host:users` | 当前 actor、租户和数据权限范围内的用户投影。 |
| `dict.values.*` | `host:dict` | 字典类型和值范围；不得开放字典管理。 |
| `files.*` | `host:files` | 文件中心可见投影和引用校验，不含下载 URL、物理路径或 provider key。 |
| `jobs.*` | `host:jobs` | 任务投影和引用校验，不含执行、启停或管理动作。 |
| `sessions.*` | `host:sessions` | 当前会话、会话投影、在线状态和可见性校验，不含吊销。 |
| `org/depts/posts.*` | `host:org` | 组织 provider 稳定投影，不含范围 SQL 构造、写入或 provider SPI。 |
| `tenants.*` | `host:tenants` | 租户可见投影和引用校验，不含租户解析、自动开通或关系写入。 |
| `plugins.*` | `host:plugins` | 插件治理只读投影和状态，不含安装、启停、升级、卸载或 provider 配置。 |
| `storage.*` | `storage`资源 | `plugin.yaml resources`中的 storage path/prefix 约束，当前插件和租户作用域。 |
| `cache.*` | `cache`资源 | 插件私有 namespace 和租户作用域，非权威缓存语义。 |
| `manifest.*` | `manifest`资源 | `resources.paths`或等价前缀授权，禁止未授权路径枚举。 |
| `state.*` | `runtime`资源 | 当前动态插件和租户作用域，不提供跨插件枚举。 |
| `data.batch_get` | `data`资源 | 仅当前插件自有表和已授权资源表；核心表、官方能力插件内部表、跨插件表禁止。 |
| `messages.*` | `host:notifications` | 当前 actor 可见通知投影、来源和消息 ID 范围，不含管理删除动作。 |
| `ai.*`、`text.*` | `host:ai` | 能力和方法状态投影，不含 provider 密钥、endpoint、模型路由或配置。 |

### 1.3 验证记录

| 命令 | 结果 |
|------|------|
| `openspec validate complete-plugin-domain-capability-expansion --strict` | 通过，输出`Change 'complete-plugin-domain-capability-expansion' is valid`。 |

### 2.x 实施记录

| 任务 | 实现与验证记录 |
|------|----------------|
| 2.1 | `usercap.SearchInput`新增`Status`、`TenantID`和`EnabledOnly`稳定过滤；`capabilityhost_user.go`在数据库查询阶段继续先应用租户过滤和`datascope.Service.ApplyUserScope`，再按关键字、状态、启用状态和租户过滤，分页使用`NormalizePage`，投影只读取`id`、`tenant_id`、`username`、`nickname`、`avatar`、`status`。动态`users.search`guest request 和 WASM dispatcher 同步携带新过滤字段。 |
| 2.2 | `dictcap.Service`新增`ListValues`、`ListValuesInput`和`MaxListValuesPageSize`。`capabilityhost_dict.go`按字典类型、租户平台覆盖、可选状态过滤和稳定排序返回有界候选；实现先在数据库侧按当前页读取 distinct value，再只批量装配当前页租户覆盖行，避免无界读全量后分页。动态`dict.values.list`已同步 catalog、alias、guest client、dispatcher 和测试。 |
| 2.3 | `filecap.Service`新增`Search`和`SearchInput`。`capabilityhost_file.go`按业务场景、文件名关键字和粗粒度 MIME 类型过滤，数据库侧完成租户过滤、分页和字段投影，只返回文件 ID、名称、MIME、大小和业务场景，不暴露 URL、path、engine 或 provider key。动态`files.search`已同步。 |
| 2.4 | `jobcap.Service`新增`Search`、`EnsureVisible`、`SearchInput`、`MaxSearchPageSize`和`MaxEnsureVisible`。`capabilityhost_job.go`按任务名称/handler、任务组和状态数据库侧过滤分页，`EnsureVisible`复用`BatchGet`的不透明缺失集合并整体拒绝。动态`jobs.search`和`jobs.visible.ensure`已同步。 |
| 2.5 | `sessioncap.Service`新增`BatchGetUserOnlineStatus`、`EnsureVisible`、`UserOnlineStatusProjection`和上限常量。`session.Store`新增`BatchGetUserOnlineStatusScoped`窄接口，`DBStore`用一次按`user_id`聚合查询完成在线状态统计，`CoordinationStore`复用共享 PostgreSQL 投影 store。`capabilityhost_session.go`先通过启动期同一`usercap.Service`批量确认用户可见性，再调用共享`session.Store`批量状态路径；不可见和非法用户进入`MissingIDs`，不会被误报为离线。动态`sessions.users.online.batch_get`和`sessions.visible.ensure`已同步。 |
| 2.6 | 已更新`pluginbridge/internal/hostservice`常量、`protocol/hostservices`catalog、`protocol_hostservice_contract.go`公开 alias、`pluginbridge/internal/domainhostcall`guest client、`internal/service/plugin/internal/wasm`dispatcher 和注册表，并扩展 hostservice/wasm 覆盖测试。普通领域新增方法继续使用`HostServiceJSONRequest`/`HostServiceJSONResponse`，未新增 per-domain 专用 codec。 |

### 2.x 影响与验证记录

| 项目 | 记录 |
|------|------|
| 数据权限与性能 | 用户、文件、任务和会话查询均在数据库或 owner 批量路径注入租户/数据权限；字典候选按租户平台覆盖且只装配当前页；会话在线状态先批量校验用户可见，再一次聚合查询 session projection，避免逐用户查询。 |
| 缓存一致性 | 本批不新增权威缓存。会话复用启动期共享`session.Store`，协调模式继续委托共享 projection；字典只读取权威字典表和既有租户覆盖语义；动态 dispatcher 不缓存状态。 |
| DI 来源检查 | `newSessionCapabilityAdapter`新增`usercap.Service`运行期依赖。owner 为同一`capabilityhost.New`中创建的`userDomain`；创建位置为`userDomain := newUserCapabilityAdapter(tenantFilterSvc, scopeSvc)`；传递路径为`capabilityhost.New -> newSessionCapabilityAdapter(authSvc, bizCtxAdapter, userDomain, scopeSvc, sessionStore, tenantSvc)`；共享实例策略为复用启动期用户 adapter、`datascope.Service`、`session.Store`和`tenantspi.RuntimeService`，未在 session adapter 或 WASM dispatcher 内部临时创建关键服务。 |
| 文档与`i18n` | 已同步`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`的领域能力说明与可声明`hostServices`表。无运行时 UI 文案、语言包、插件清单或 API 文档源文本变更；`i18n`运行期资源无影响。 |
| SQL 与数据库索引 | 未新增 SQL、DAO、DO 或 Entity 文件。实现复用已有字段和索引假设；候选/搜索路径保持分页和投影。若后续压测发现索引不足，应另按数据库规则补 SQL。 |
| 开发工具与 E2E | 未新增或修改工具脚本、CI、Makefile、Node 脚本或前端 UI；开发工具和 E2E 不触发。 |
| 验证 | `cd apps/lina-core && go test ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... ./internal/service/session ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm -count=1`通过；`openspec validate complete-plugin-domain-capability-expansion --strict`通过。 |

### 3.x 实施记录

| 任务 | 实现与验证记录 |
|------|----------------|
| 3.1 | `orgcap.Service`新增`BatchGetUserOrgProfiles`、`ListDeptTree`、`SearchDepartments`、`ListPostOptionsPage`、`EnsureDepartmentsVisible`和`EnsurePostsVisible`，并定义用户档案、部门投影、树结果、部门搜索、岗位候选和上限常量。`UserDeptTree`和`ListPostOptions`继续只属于`orgspi.ProjectionService`，普通`orgcap.Service`未暴露工作台专用投影。 |
| 3.2 | `orgspi.Provider`新增对应 provider 方法；`orgspi.serviceImpl`在 provider 缺失时对读取类返回空档案、空树或空页，对可见性校验返回`CodeCapabilityUnavailable`，并按批量、节点和分页上限归一化。`linapro-org-core` provider 使用集合化查询装配用户部门、岗位名称、部门搜索、岗位候选和可见性计数校验，未暴露 DAO、Entity 或查询构造器。 |
| 3.3 | `tenantcap.Service`新增`CurrentTenantInfo`、`BatchGetTenants`、`SearchTenants`、`BatchListUserTenants`和`EnsureTenantsVisible`，并定义租户批量、搜索和用户租户批量上限。`tenantspi`将租户投影 provider 和用户租户批量方法作为可选 provider seam，未把 HTTP resolver、查询构造器或写入关系暴露给普通消费面。 |
| 3.4 | `tenantspi`默认实现对平台租户提供中性投影；provider 缺失时批量读取只返回平台租户并将其他目标放入`MissingIDs`，搜索和用户租户批量返回空集合，可见性校验只允许平台租户。`linapro-tenant-core` provider 使用集合化 membership/tenant 查询实现批量用户租户列表，并按当前 actor 可见边界处理批量租户读取、搜索和可见性校验。 |
| 3.5 | `plugincap.RegistryService`新增`Current`、`Search`、分页化`ListTenantPlugins(input)`和`BatchGetCapabilityStatus`，并定义插件搜索分页和能力状态批量上限。 |
| 3.6 | `capabilityhost`插件治理 adapter 复用`sys_plugin`治理读模型和`sys_plugin_state`租户 enabled 状态当前页集合查询；`Current`只读取`capCtx.PluginID`或已绑定插件 ID；`Search`数据库侧按插件 ID、名称、备注、类型和启用状态过滤分页；`ListTenantPlugins`对当前页插件一次读取租户 enabled 状态；`BatchGetCapabilityStatus`只返回`org`、`tenant`和`ai.text`的`CapabilityStatus`，不暴露 provider 配置。 |
| 3.7 | 已同步`pluginbridge/internal/hostservice`常量、`protocol/hostservices`catalog、`protocol_hostservice_contract.go`alias、`pluginbridge/internal/domainhostcall`guest client、`internal/service/plugin/internal/wasm`dispatcher 和注册表；阶段 2 新增动态方法继续使用`HostServiceJSONRequest`/`HostServiceJSONResponse`，未新增 per-domain 专用 codec。 |

### 3.x 影响与验证记录

| 项目 | 记录 |
|------|------|
| 数据权限与性能 | 组织、租户和插件治理新增路径均具备批量或分页上限。组织 provider 按用户、部门、岗位集合化查询并计数校验可见性；租户 provider 批量装配用户租户列表，批量读取不可见目标进入`MissingIDs`；插件治理搜索数据库侧分页，租户 enabled 状态只对当前页集合查询，避免重复扫描 manifest 或逐插件状态读取。 |
| 缓存一致性 | 本批不新增权威缓存。插件治理状态读取复用启动期注入的`plugincap.StateService`和现有`sys_plugin_state`修订机制；组织、租户和`AI`状态只通过各能力 owner 的`Status`读取，不在 adapter、guest client 或 WASM dispatcher 中新增本地缓存。 |
| DI 来源检查 | `newPluginCapabilityAdapter`新增`orgcap.Service`、`tenantcap.Service`和`aicap.Service`运行期依赖。owner 分别为`capabilityhost.New`入参中的`orgSvc`、`tenantSvc`和同一函数创建的`aiDomain := capabilityai.New(aiTextSvc)`；传递路径为`capabilityhost.New -> newPluginCapabilityAdapter(pluginConfigFactory, pluginState, pluginLifecycle, orgSvc, tenantSvc, aiDomain)`；共享实例策略为复用启动期组织、租户、插件状态、生命周期和`AI Text`服务，未在插件治理 adapter 或 WASM dispatcher 内部临时创建关键服务。 |
| 插件本地规范 | 修改`apps/lina-plugins/linapro-org-core`和`apps/lina-plugins/linapro-tenant-core`前已检查插件根目录`AGENTS.md`，两个插件均不存在本地规范文件，继续遵守顶层`AGENTS.md`和命中规则。 |
| 文档与`i18n` | 本批新增的是 Go 能力契约、动态协议和 OpenSpec 任务记录；无运行时 UI 文案、菜单、语言包、插件清单或 API 文档源文本变更。`i18n`运行期资源无影响。 |
| SQL 与数据库索引 | 未新增 SQL、DAO、DO 或 Entity 文件。组织、租户和插件治理查询复用现有表与字段；候选和搜索路径均保持分页、投影和集合化读取。若后续压测发现索引不足，应另按数据库规则补 SQL。 |
| 开发工具与 E2E | 未新增或修改工具脚本、CI、Makefile、Node 脚本或前端 UI；开发工具跨平台和 E2E 不触发。 |
| 验证 | `cd apps/lina-core && go test ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm -count=1`通过；`cd apps/lina-plugins/linapro-org-core && GOWORK=off go test ./backend/internal/provider/orgcapadapter -count=1`通过；`cd apps/lina-plugins/linapro-tenant-core && GOWORK=off go test ./backend/internal/service/provider -count=1`通过。插件包测试直接运行时被仓库根`go.work`未包含插件模块阻断，改用`GOWORK=off`在插件自身模块内验证。 |

### 4.x 实施记录

| 任务 | 实现与验证记录 |
|------|----------------|
| 4.1 | `storagecap.Service`已提供`BatchStat`、`ListCursor`和`DeleteMany`，并落地路径数量、单路径长度、总 path 字节数和列表分页上限。动态 storage dispatcher 在资源授权、插件 ID 和租户作用域内处理`storage.stat.batch`、`storage.list.cursor`和`storage.delete.batch`；批量 stat 返回不透明缺失集合，批量删除缺失对象按 no-op 处理，provider 冲突继续返回`CodeStorageProviderConflict`。 |
| 4.2 | `cachecap.Service`已提供`GetMany`、`SetMany`和`DeleteMany`，动态 cache host service 已同步`cache.get_many`、`cache.set_many`和`cache.delete_many`。实现复用启动期共享`kvcache`后端和既有插件命名空间、租户作用域、TTL 语义，不新增权威缓存或节点本地兜底。 |
| 4.3 | `manifestcap.Service`已提供`GetMany`和`List`，动态 manifest host service 已同步`manifest.get_many`和`manifest.list`。实现校验路径数量、路径长度、前缀授权和总字节数，`config/`等前缀授权可覆盖嵌套资源，未授权和不存在资源统一进入缺失语义。 |
| 4.4 | 动态 runtime state host service 已提供`state.get_many`、`state.set_many`和`state.delete_many`，所有方法限定当前动态插件和租户作用域。`get_many`使用一次`WhereIn`集合查询，`set_many`在一个事务内批量写入，`delete_many`使用一次`WhereIn`删除；同时限制 key 数量、单 key 长度和总 value 字节数。 |
| 4.5 | 动态 data host service 已提供`data.batch_get`，并在`datahost.ExecuteBatchGet`中按主键集合去重后执行一次集合查询。实现复用现有`data`资源表授权、字段投影、租户过滤、软删除和数据权限 scope；运行时额外拒绝空表名、`sys_*`和非`plugin_<plugin-id>*`表名，确认未新增`data.exists`或`data.count`动态发布。 |
| 4.6 | 已同步阶段 3 的 catalog、protocol alias、storage/cache/manifest/runtime JSON envelope 或 data 专用 codec、guest client、recordstore 入口、非 WASI stub、WASM dispatcher、注册表和覆盖测试。资源型能力未新增普通领域 per-domain 专用 codec；`data.batch_get`使用显式协议 codec。 |

### 4.x 影响与验证记录

| 项目 | 记录 |
|------|------|
| HTTP API 与前端 UI | 无静态 HTTP API、路由、Controller、前端页面、表单、表格或用户可观察 UI 变更；仅扩展插件领域能力和动态`hostServices`协议。 |
| 数据权限与性能 | `data.batch_get`只允许当前插件自有授权表，数据库侧注入租户、软删除和数据权限 scope，并用集合查询返回当前权限范围内记录；storage、manifest、cache 和 runtime state 均按插件/租户作用域和资源授权处理。批量路径使用集合化查询、共享后端批量接口或有界事务，避免逐项详情调用作为常规装配路径。 |
| 缓存一致性 | `Cache`批量能力复用共享`kvcache`后端和既有 TTL/删除语义，不成为权限、配置、业务记录或审计状态的权威来源。runtime state 复用共享`sys_plugin_state`存储；manifest 不新增缓存；data 不新增缓存。集群模式不新增节点本地默认实例。 |
| DI 来源检查 | 阶段 3 未新增运行期依赖 owner。storage、cache、manifest、runtime state 和 data dispatcher 均复用既有`capabilityhost`、WASM host service、DB handle、`kvcache`和 manifest/runtime 存储依赖；未在 adapter、guest client 或 dispatcher 中临时`New()`关键服务。 |
| SQL 与数据库索引 | 未新增生产 SQL、DAO、DO 或 Entity 文件。runtime state 继续使用既有`sys_plugin_state`表；`data.batch_get`使用插件自有授权表的主键集合查询。测试只创建临时插件自有表，不属于交付 SQL。若后续压测发现插件业务表索引不足，应由对应插件 SQL 按数据库规则维护。 |
| 文档与`i18n` | 无运行时 UI 文案、菜单、语言包、插件清单、API 文档源文本或字典文案变更；`i18n`运行期资源无影响。阶段 5 将统一同步`apps/lina-core/pkg/plugin`中英文 README。 |
| 开发工具与 E2E | 未新增或修改`Makefile`、`make.cmd`、脚本、CI、Node 工具、Go 工具或前端页面；开发工具跨平台和 E2E 不触发。 |
| 验证 | `cd apps/lina-core && go test ./pkg/plugin/capability/storagecap ./pkg/plugin/capability/cachecap ./pkg/plugin/capability/manifestcap ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/datahost ./pkg/plugin/pluginbridge/internal/domainhostcall ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/pluginbridge/protocol ./pkg/plugin/pluginbridge/protocol/hostservices ./pkg/plugin/pluginbridge/recordstore ./pkg/plugin/pluginbridge -run 'TestStorage|TestCache|TestManifest|TestRuntime|TestData|TestHostService|TestCatalog|Test.*Storage|Test.*Cache|Test.*Manifest|Test.*Runtime|Test.*Data|TestExecuteBatchGet|TestHandleHostServiceInvokeData' -count=1`通过。 |

### 5.x 实施记录

| 任务 | 实现与验证记录 |
|------|----------------|
| 5.1 | `notifycap.Service.BatchGet`已收敛为`capmodel.BatchResult[*notifycap.MessageProjection, notifycap.MessageID]`，`MessageProjection`只包含通知 ID、租户、插件、来源、分类、标题和 Unix 毫秒创建时间等稳定字段。动态`messages.batch_get`通过 JSON envelope 返回同一类型化投影，WASM dispatcher 和 guest client 已同步，未继续返回未治理`map[string]any`。 |
| 5.2 | `notifycap.Service`新增`BatchGetBySource`和`EnsureVisible`。宿主 adapter 使用`sys_notify_message`与`sys_notify_delivery`一次 join 查询，按当前 actor、租户和`SourceType + SourceIDs`集合过滤；`SourceIDs`和单次返回消息数均有上限，不存在、不可见和无消息来源统一进入缺失或拒绝语义。动态`messages.by_source.batch_get`和`messages.visible.ensure`已同步 catalog、protocol alias、guest client、dispatcher、注册表和测试。 |
| 5.3 | `aicap.Service`新增`MethodStatuses`和`MethodStatusQuery`，对跨子能力方法状态设置`MaxMethodStatusBatchSize`。实现按子能力调用各自`MethodStatus`，provider 缺失、禁用、未知子能力或方法不可用时返回结构化 unavailable 状态，不暴露 provider 配置、密钥、endpoint 或模型路由。 |
| 5.4 | `aitext.Service`新增`MethodStatus`，文本能力动态发布`text.method_status.get`，AI 命名空间动态发布`ai.methods.status.batch_get`。guest client、WASM dispatcher、注册表、catalog、public alias 和测试已同步，`AI`状态 transport 继续使用 JSON envelope，不新增 per-domain 普通领域专用 codec。 |

### 5.x 影响与验证记录

| 项目 | 记录 |
|------|------|
| HTTP API 与前端 UI | 无静态 HTTP API、路由、Controller、前端页面、表格、表单、交互或用户可观察 UI 变更；仅扩展插件领域能力与动态`hostServices`协议。 |
| 数据权限与性能 | 通知读取要求 actor 用户，使用 delivery join 在数据库阶段约束当前用户可见性和租户范围；按来源读取一次集合查询并限制`SourceIDs`和返回消息数，避免按来源循环列表。AI 状态只读取 provider 状态投影，不访问业务数据或泄露 provider 私有配置。 |
| 缓存一致性 | 本批不新增权威缓存。通知读取直接查询通知权威表；AI 状态复用已有 provider manager、插件启用状态和子能力 owner 的`Status/MethodStatus`路径，不在 adapter、guest client 或 dispatcher 中新增节点本地缓存。 |
| DI 来源检查 | 阶段 4 未新增启动期运行时依赖 owner。通知 adapter 继续复用`capabilityhost.New`中传入的共享`notify`服务；AI 命名空间复用同一函数创建并注入的`aiTextSvc`和已有子能力 fallback，WASM dispatcher 只通过已注入的`capability.Services`目录取用能力，未临时`New()`关键服务图。 |
| SQL 与数据库索引 | 未新增 SQL、DAO、DO 或 Entity 文件。通知按消息 ID 和来源集合读取复用`sys_notify_message`和`sys_notify_delivery`现有表，查询有上限和投影字段；若后续压测发现来源读取索引不足，应另按数据库规则补 SQL。 |
| 文档与`i18n` | 无运行时 UI 文案、菜单、语言包、插件清单、API 文档源文本或字典文案变更；`i18n`运行期资源无影响。README 中英文镜像在阶段 6 同步。 |
| 开发工具与 E2E | 未新增或修改`Makefile`、`make.cmd`、脚本、CI、Node 工具、Go 工具或前端页面；开发工具跨平台和 E2E 不触发。 |
| 验证 | `cd apps/lina-core && go test ./pkg/plugin/capability/notifycap ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm ./pkg/plugin/capability/aicap ./pkg/plugin/capability/aicap/aitext ./pkg/plugin/pluginbridge/internal/domainhostcall ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/pluginbridge/protocol ./pkg/plugin/pluginbridge/protocol/hostservices ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/integration -run 'Test.*AI|Test.*Notif|TestCatalog|TestHostServiceReadme|TestValidateHostServiceSpecs.*AI|TestValidateHostServiceSpecsAcceptsDomainServicesWithoutResources|TestHostServiceDescriptors|TestDomainCapabilityBoundaryGovernance|TestHandleHostServiceInvokeAI|TestHandleHostServiceInvokeNotifications|TestBatchGetBySourceRejectsSourceIDLimit' -count=1`通过。 |

### 6.x 实施记录

| 任务 | 实现与验证记录 |
|------|----------------|
| 6.1 | 已更新`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`，同步领域能力说明和可声明`hostServices`表，覆盖 runtime state、storage、data、cache、manifest、AI、notifications、plugins、org 和 tenant 新增方法。 |
| 6.2 | `HostServiceReadmeGeneratedBlocksMatchCatalog`新增中英文 README 生成块与公开 catalog 的一致性测试；现有 descriptor 测试继续覆盖 catalog、public alias、guest client selector、dispatcher selector、JSON envelope 或专用 codec 同步点。 |
| 6.3 | 静态检索已确认生产动态桥接和 dispatcher 未新增动态`i18n`host service；生产 bridge/host service 未新增`data.exists`或`data.count`；生产动态/readme 表面未暴露 Provider SPI 或 provider 配置；生产 datahost/recordstore/protocol/readme 表面未直接暴露核心表；普通领域方法未新增专用 protocol codec。 |
| 6.4 | 本任务记录已覆盖 HTTP API、前端 UI、SQL schema、插件清单、语言包、开发工具脚本、E2E、数据权限、缓存一致性、数据库索引和 DI 来源检查影响。 |

### 6.x 影响与验证记录

| 项目 | 记录 |
|------|------|
| HTTP API 与前端 UI | 无静态 HTTP API 或前端 UI 变更；README 只描述插件公开契约和动态`hostServices`。 |
| 插件清单与协议 | `hostServices`catalog 和 README 表同步，未修改`plugin.yaml`schema 或插件清单字段结构；新增方法沿用既有声明方法数组和资源声明形态。 |
| 数据权限与边界 | 静态扫描确认没有发布动态`i18n`、`data.exists`、`data.count`、Provider SPI、provider 配置或核心表直接访问；`data.batch_get`仍受插件自有表、资源表授权、租户和软删除边界约束。 |
| 缓存一致性 | README 和 catalog 治理测试不新增运行期缓存；无缓存失效、跨实例同步或共享后端变更。 |
| SQL 与数据库索引 | 无 SQL schema、DAO、DO、Entity 或索引变更。 |
| 文档与`i18n` | 中英文 README 镜像已同步，新增测试确保 generated host-services 块与 catalog 一致。无运行时语言包或 API 文档翻译资源影响。 |
| 开发工具与 E2E | 未新增或修改工具脚本、CI、Makefile、Node 工具或前端页面；开发工具跨平台和 E2E 不触发。 |
| 验证 | `cd apps/lina-core && go test ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/pluginbridge/protocol/hostservices -run 'TestHostServiceReadmeGeneratedBlocksMatchCatalog|TestHostServiceDescriptorsCoverProtocolGuestAndDispatcher|TestCatalogPayloadKinds|TestOrdinaryJSONServicesHaveNoDedicatedCapabilityCodecs' -count=1`通过；阶段 6 静态扫描均通过。 |

### 7.x 验证记录

| 任务 | 结果 |
|------|------|
| 7.1 | `openspec validate complete-plugin-domain-capability-expansion --strict`通过，输出`Change 'complete-plugin-domain-capability-expansion' is valid`。 |
| 7.2 | `cd apps/lina-core && go test ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm -count=1`通过。 |
| 7.3 | `cd apps/lina-core && go test ./internal/service/plugin/internal/datahost ./pkg/plugin/pluginbridge/recordstore ./internal/service/plugin/internal/wasm -count=1`通过。 |
| 7.4 | `cd apps/lina-core && go test ./internal/cmd -count=1`通过。 |
| 7.5 | `git diff --check`通过；生产动态桥接和 dispatcher 静态扫描确认未新增动态`i18n`host service、`data.exists`/`data.count`、Provider SPI 或 provider 配置暴露、核心表直接动态 data 暴露和普通领域专用 protocol codec。 |
| 7.6 | `lina-review`任务级审查完成。审查范围来自`git status --short`、`git ls-files --others --exclude-standard`、`apps/lina-plugins`子仓库状态和本 OpenSpec 变更上下文；插件本地规范检查确认`linapro-org-core`与`linapro-tenant-core`根目录均无`AGENTS.md`。已读取并适用`AGENTS.md`、`openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`api-contract.md`、`data-permission.md`、`cache-consistency.md`、`database.md`、`i18n.md`、`testing.md`、`documentation.md`、`dev-tooling.md`和 Markdown 格式指令。结论：未发现阻塞问题；确认无静态 HTTP API、前端 UI、SQL schema、插件清单字段、语言包、开发工具脚本或 E2E 影响，相关无影响判断已记录；`git diff --check`和受影响`capabilityhost`通知测试在审查清理后通过。 |

## Feedback

- [x] **FB-1**: 收敛`Storage`provider 选择为零配置，删除`activeProviderPluginId`和`allowLocalProviderInCluster`主配置项
- [x] **FB-2**: 多个`Storage`provider 同时可服务时返回专用冲突错误码
- [x] **FB-3**: 共享领域 provider 多实例冲突返回结构化冲突错误码

### FB-1 实施记录

| 项目 | 记录 |
|------|------|
| 根因 | 原实现把`Storage`provider 选择拆成`plugin.storage.activeProviderPluginId`和`plugin.storage.allowLocalProviderInCluster`两个主配置项，导致宿主配置需要理解具体 provider 插件实现方，并形成配置状态与插件治理状态两套来源。 |
| 实现 | 删除`PluginStorageConfig`、`GetPluginStorage`、`HostPluginStorageConfigReader`和本地 provider 集群例外参数；`Storage`provider 改为根据已注册 provider 插件和`plugincap.StateService.IsProviderEnabled`自动选择。没有可服务 provider 插件时回退内置本地 provider；恰好一个 provider 插件可服务时使用该插件；多个 provider 插件可服务时返回`CodeStorageProviderConflict`，不静默选择或回退。 |
| 文档 | 已更新`plugin-storage-service`增量规范、`design.md`决策和`apps/lina-core/pkg/plugin`中英文 README。 |
| 缓存一致性 | 有影响评估但不新增缓存。provider 可服务状态仍复用启动期注入的插件 enabled snapshot/治理状态读取路径，不在 storage adapter 或 dispatcher 中新增本地缓存。 |
| 数据权限 | 无数据权限影响；本次只调整 provider 选择，不改变对象路径授权、插件作用域或租户作用域。 |
| `i18n` | 无运行时 UI、插件清单、语言包或 API 文档源文本影响；新增说明仅为 OpenSpec 和 README 文档。 |
| 开发工具与 E2E | 无开发工具脚本或前端 UI 影响；不触发 E2E。 |
| 验证 | `cd apps/lina-core && go test ./pkg/plugin/capability/storagecap -count=1`通过；`cd apps/lina-core && go test ./internal/service/config -count=1`通过；`cd apps/lina-core && go test ./internal/service/role -run TestNonExistent -count=0`通过；`openspec validate complete-plugin-domain-capability-expansion --strict`通过；`git diff --check`通过。 |
| 剩余阻断 | `cd apps/lina-core && go test ./internal/service/plugin/internal/capabilityhost -run 'Test(StorageAdapter|LocalStorageProvider)' -count=1`、`cd apps/lina-core && go test ./internal/service/user -run TestNonExistent -count=0`和`cd apps/lina-core && go test ./internal/cmd -run TestNonExistent -count=0`均被当前工作区既有`internal/service/plugin/internal/wasm/wasm_host_service_plugins.go:42`调用`ListTenantPlugins`缺少`plugincap.TenantListInput`参数阻断，和本次`Storage`provider 配置收敛无直接关系。 |

### FB-2 实施记录

| 项目 | 记录 |
|------|------|
| 根因 | 多个`Storage`provider 同时可服务属于 provider 选择冲突，不是 provider 不存在或不可用；复用`CodeStorageProviderUnavailable`会让调用端和治理日志误判为“无可用 provider”。 |
| 实现 | 新增`CodeStorageProviderConflict`错误码，`ResolveProvider`在多个 provider 同时可服务时返回该错误码，并携带`providerIds`参数；单 provider 缺失、factory 缺失或返回 nil 的不可用场景继续使用`CodeStorageProviderUnavailable`。 |
| 规范与 README | 已更新`plugin-storage-service`增量规范，明确多个 provider 同时可服务时返回 storage provider 冲突错误。已审查`apps/lina-core/pkg/plugin`中英文 README，现有文本已说明多个 provider 时拒绝调用，未枚举具体错误码，本次无需同步。 |
| 缓存一致性 | 无缓存一致性影响；本次只调整错误语义，不改变 provider 可服务状态来源、缓存、修订号或失效策略。 |
| 数据权限 | 无数据权限影响；本次不改变 storage 对象路径授权、插件作用域、租户作用域或数据读取路径。 |
| `i18n` | 有错误消息影响；新增 bizerr 英文 fallback 和稳定 message key，未新增运行时 UI、插件清单、语言包或 API 文档源文本。 |
| 开发工具与 E2E | 无开发工具脚本、CI、Makefile、Node 脚本或前端 UI 影响；不触发 E2E。 |
| 验证 | `cd apps/lina-core && go test ./pkg/plugin/capability/storagecap -count=1`通过；`openspec validate complete-plugin-domain-capability-expansion --strict`通过；`git diff --check`通过。 |

### FB-3 实施记录

| 项目 | 记录 |
|------|------|
| 根因 | `storagecap`已把多 provider 同时可服务修正为`CodeStorageProviderConflict`，但`orgspi`、`tenantspi`和`aitext`共用的`capabilityregistry.ActiveProviderWithError`在同类 singleton provider 冲突时仍返回裸`gerror`；`aitext.GenerateText`还会把该冲突包装成`CodeTextProviderUnavailable`，导致调用端把“多个 provider 冲突”误判成“provider 不可用”。 |
| 实现 | 新增共享`CodeCapabilityProviderConflict`错误码；`capabilityregistry`在多个 enabled provider 时返回结构化冲突错误并携带`capability`和`providerIds`参数；`aitext.GenerateText`对该冲突错误直接透传，其他 provider 创建失败或不可用场景继续按原语义处理。 |
| 规范与 README | 已更新错误语义矩阵和`plugin-host-domain-capabilities`增量规范，明确多个 singleton provider 同时可服务返回`CodeCapabilityProviderConflict`且状态 reason 为`provider_conflict`。已审查`apps/lina-core/pkg/plugin`中英文 README，现有内容说明 provider manager 与 storage 多 provider 拒绝策略，本次共享错误码属于内部错误语义收敛，无需同步 README。 |
| 缓存一致性 | 无缓存一致性影响；本次只调整 provider 冲突错误语义，不改变 provider enablement 来源、状态快照、修订号、失效或跨实例同步。 |
| 数据权限 | 无数据权限影响；本次不新增数据读取、写入、批量装配或租户/组织可见性边界，只改变多 provider 冲突时的错误码。 |
| `i18n` | 有错误消息影响；新增 bizerr 英文 fallback 和稳定 message key，未新增运行时 UI、插件清单、语言包或 API 文档源文本。 |
| 开发工具与 E2E | 无开发工具脚本、CI、Makefile、Node 脚本或前端 UI 影响；不触发 E2E。 |
| 外部规则 | 已按`AGENTS.md`读取`openspec.md`、`backend-go.md`、`testing.md`、`architecture.md`、`plugin.md`、`i18n.md`、`documentation.md`和`cache-consistency.md`；确认 API 契约、数据库、前端 UI、开发工具和插件本地规范无影响。 |
| 验证 | 目标冲突路径子集测试通过，覆盖`StorageAdapter`、`DisabledOrganization`、`OrganizationCapabilityRejects`、`DisabledTenant`、`TenantCapabilityRejects`、`GenerateText`和`ActiveProvider`；`cd apps/lina-core && go test ./pkg/plugin/capability/internal/capabilityregistry ./pkg/plugin/capability/orgcap/orgspi ./pkg/plugin/capability/tenantcap/tenantspi ./pkg/plugin/capability/aicap/aitext ./internal/service/plugin/internal/capabilityhost -count=1`通过；`openspec validate complete-plugin-domain-capability-expansion --strict`通过；`git diff --check`通过。 |
