# Design

## Context

LinaPro 插件平台经过多轮演进，从初始的清单契约和生命周期模型逐步扩展到动态`WASM`运行时、宿主服务授权、能力目录、领域能力模型、能力组件治理、运行时复杂度治理、清单资源读取统一和管理页面性能优化。当前平台已覆盖源码插件和动态插件的完整生命周期，具备统一的`plugin.yaml`清单契约、`/x/{plugin-id}/api/v1/...`路由命名空间、`/x-assets/{plugin-id}/{version}/...`公开资产托管、`hostServices`授权快照、`*cap`能力组件体系、`Manifest()`资源读取、`CapabilityContext`上下文治理和分页摘要列表读模型。

本项目定位为全新项目，不考虑旧接口兼容。所有设计决策均以一次性迁移和强治理为目标，不保留旧接口、旧协议方法或旧插件调用方式作为生产兼容入口。

## Goals / Non-Goals

**Goals:**

- 建立插件访问宿主数据的统一领域能力模型，所有插件只能通过领域接口、源码插件`pluginhost.Services`或动态插件`hostServices`协议访问宿主业务数据和治理状态。
- 将插件公开能力组件统一为`*cap`命名体系，删除`capability/contract`万能聚合包，收敛配置、认证授权、插件自能力和租户过滤等领域的公开入口。
- 将插件资源读取统一为`Manifest()`，覆盖`manifest/`下所有实际打包资源，删除旧`Metadata`服务语义。
- 移除源码插件宿主配置读取的 key 白名单限制，保持动态插件授权快照约束。
- 消除 runtime state 列表路径中的 N+1 manifest 扫描和 artifact 解析风险，确保 WASM host service 和 runtime 缓存敏感依赖来自启动期共享实例。
- 将插件管理列表首屏改为轻量分页摘要查询，详情和治理弹窗数据按需加载。
- 保持动态插件`service + method`协议、授权快照、错误 envelope 和数据边界不变。
- 用静态检索、Go 编译门禁和 OpenSpec 严格校验确认旧路径和旧接口不再进入公开能力目录。

**Non-Goals:**

- 不提供宿主核心表的通用表查询、通用写入、通用 SQL 或跨插件表白名单能力。
- 不保留旧`orgcap`、旧`tenantcap`、旧`ai`根方法、旧动态`host service`方法或插件侧宿主表`DAO`生成作为生产兼容层。
- 不把具体管理工作台页面的列、筛选项或路由装配形状固化到`apps/lina-core`核心领域契约中。
- 不改变动态插件 ABI、WASM custom section、host service service/method 字符串、payload wire 字段编号或 guest SDK 公开行为。
- 不新增 HTTP API、前端页面、数据库表、DAO、插件清单字段或运行时用户可见文案。

## Decisions

### 插件契约和生命周期

插件平台以`plugin.yaml`为统一入口，源码插件使用`apps/lina-plugins/<plugin-id>/`目录，动态插件使用发布产物和 active release 快照。清单承载插件身份、类型、版本、菜单、权限、`hostServices`、依赖声明、运行期配置和 manifest 资源索引；未声明的能力不得由运行时自行推断。

生命周期从"源码目录扫描 + 动态产物装载"演进为显式治理模型。源码插件可被发现、安装、启用、禁用、升级和通过启动引导自动启用；动态插件覆盖上传、安装、授权确认、启用、禁用、卸载、升级、同版本刷新和 active release 切换。插件业务入口在`pending_upgrade`、`abnormal`或`upgrade_failed`状态下必须受控，插件管理和升级入口仍保持可访问。

生命周期回调统一使用`Before*`、`Upgrade`、`Uninstall`和`After*`命名，替代旧`Can*`或 guard 风格契约。`Before*`可阻断操作并返回稳定原因键，`After*`作为 best-effort 通知；卸载相关回调携带`purgeStorageData`，使插件区分保留数据和清理数据两种卸载策略。动态插件生命周期契约由构建期自动发现 guest controller 方法生成，运行时只信任产物内显式生命周期 custom section。

插件生命周期 SQL、mock SQL、upgrade SQL、rollback SQL 和`sys_plugin_migration`账本必须保持事务一致。失败诊断需要保留原始失败、rollback 失败、失败阶段、错误码、消息键和 manifest 快照，避免只记录 warning 后丢失恢复依据。

### 动态运行时和路由资产边界

动态插件运行时以`WASM`产物自定义段为权威输入。`pluginbridge`集中解析`WASM`自定义段，`i18n`、`apidoc`和插件运行时不得维护重复解析器。运行时资源视图绑定 active release checksum 或 generation，包含 manifest、默认配置、前端资产、路由契约、生命周期契约、`hostServices`授权快照和语言资源。

路由边界收敛到统一插件 API 命名空间。源码插件和动态插件 API 均使用`/x/{plugin-id}/api/v1/...`，`/x`只表示插件 API 命名空间。源码插件公开页面、门户、自管静态资源和 fallback 使用非保留路径。公开资产由`plugin.yaml public_assets`声明，并通过`/x-assets/{plugin-id}/{version}/...`托管，不再保留`/plugin-assets`兼容入口。

管理工作台入口从默认根路径收敛为可配置入口，默认`/admin`。工作台 SPA fallback 只覆盖工作台入口及其子路径，不吞掉源码插件公开路由；`/api`、`/x`、`/x-assets`和已注册插件路由优先级高于工作台 fallback。

所有动态`WASM`执行入口共享宿主资源边界。路由、cron、生命周期和发现执行在调用方已有 deadline 时不得放宽；没有 deadline 时使用宿主默认超时；内存分配超过上限时拒绝或终止本次执行。

### 领域能力作为唯一宿主数据边界

插件需要的是用户、授权、组织、租户、字典、会话、文件、通知、插件状态和任务等领域能力，不是宿主表行。因此所有宿主核心表和官方能力插件表都视为领域 owner 私有事实；插件公共契约只暴露领域接口、输入输出 DTO、投影结构和稳定错误语义。动态`data`服务只允许访问当前插件自有表；`sys_*`表和官方能力插件表不得通过`data`服务对普通插件开放。

每个领域定义自己的命名 ID 类型，动态协议统一编码为字符串；领域实现内部可以映射到数据库主键、业务键或组合键。`BatchGet*`返回可见`Items`和不可解释的`MissingIDs`，不区分真实不存在和不可见。

### 能力组件统一为 *cap 命名体系

`capability.Services`普通公开目录收敛为只返回各领域`*cap.Service`或对应组件的窄服务接口。`capability/contract`拆分为跨领域公共原语包和具体能力组件包，公共原语包不得定义具体能力 Service。

认证 token 和授权能力收敛到`authcap`能力族，使用`authcap/token`和`authcap/authz`子领域维护窄接口。插件自身配置、插件生命周期和插件状态收口到`plugincap`子领域，通过`Services.Plugins()`命名空间访问。租户过滤接口归属`tenantcap.PluginTableFilterService`，只通过`pluginhost.Services.TenantFilter()`源码插件专用接缝暴露。动态插件受治理数据 SDK 收口到`capability/recordstore`。

Go 包重命名不得改变动态插件协议。`plugin.yaml hostServices`仍使用`service: ai`、`service: user`、`service: config`等领域服务名，运行时授权快照仍按`service + method + resource`校验。

### 配置和清单类只读宿主服务

配置能力公开面只保留两类：`Services.Plugins().Config()`读取当前插件自身配置，`Services.HostConfig()`读取宿主授权开放配置。源码插件通过`HostServices.HostConfig()`读取宿主配置时不受公开 key 白名单限制；动态插件仍受`hostServices`授权快照约束。

`Manifest()`成为插件自有`manifest/`目录的只读原始资源视图，覆盖`config/`、`sql/`、`i18n/`等专用目录。`metadata.yaml`降级为普通可选资源。源码插件可读取自身嵌入的完整`manifest/`资源；动态插件按`hostServices`授权快照读取 active release artifact 中的资源。专用管线继续负责资源生效：`Config()`读取有效配置，SQL 由生命周期管线执行，i18n 由资源管线加载。

### 宿主服务和能力目录

动态插件不直接获得宿主连接、`DAO`、`gdb.Model`、请求对象或内部 service。所有宿主能力通过版本化 host service envelope 进入`pkg/plugin/capability`能力目录或受控适配器。源码插件和动态插件共享同一能力语义，只允许 transport 不同。

源码插件通过`pluginhost.Services.Admin()`获得完整类型化 AdminService 目录；管理方法不使用字符串式授权声明，但必须执行`CapabilityContext`解析、租户边界、目标数据边界、状态机、数量上限、系统 actor 和审计治理。动态插件领域方法通过安装授权快照调用，运行时不再额外校验当前用户是否拥有某个工作台菜单或按钮权限。

数据能力通过结构化 data service 和受限 DSL 承载。动态插件按表、方法、字段、分页、排序、用户上下文和数据权限执行治理，不暴露原始 SQL。列表和投影接口必须批量化、有界装配，避免 N+1 查询。

缓存、锁、网络、通知和存储服务均以插件作用域授权为边界。缓存是有损缓存，不作为权限、配置、插件状态、业务数据或修订号事实源。集群模式使用 coordination 后端，单机模式可用 SQL table 后端。

### 包边界和内部实现

插件公共契约统一收敛到`pkg/plugin`命名空间。`pluginhost`只负责源码插件贡献入口，`pluginbridge`只负责动态插件 ABI、transport 和公开协议出口，`capability`只负责插件消费宿主能力。`capability`是最底层契约层，其非测试代码不得 import `pluginbridge`或`pluginhost`；`pluginhost`非测试代码不得 import `pluginbridge`。依赖方向由随`go test`执行的治理测试持续验证。

`ManifestSnapshotV1`类型定义从`pluginbridge/contract`迁入`capability/capmodel`公共原语包，`pluginbridge/contract`保留类型别名以维持 protocol facade 别名转发惯例。recordstore SDK 从`capability/recordstore`迁移到`pluginbridge/recordstore`，承认其动态插件专属 guest SDK 的事实定位。

`pluginbridge`子组件按职责拆分，公开协议出口唯一于`pluginbridge/protocol`。host service 协议样板由单一描述源覆盖，新增 method 时自动化验证所有同步点。动态插件公开入口从`pluginbridge/guest`收敛到`pluginbridge`根包，与源码插件`pluginhost`入口保持根包对称。声明期入口统一为`Declarations`，与运行期`Services`区分。

宿主插件运行时治理收敛到`apps/lina-core/internal/service/plugin`及其职责明确的子组件。宿主领域能力实现组件从`internal/service/plugin/internal/hostservices`重命名为`capabilityhost`，避免和动态插件`hostServices`协议目录混淆。启动层只依赖`plugin`根 facade 或明确允许的受控子包。框架能力 provider 使用窄接口和强类型 provider env。

### Host Service 协议 Catalog 与 Registry Dispatch

动态插件 host service 的领域桥接边界收敛为公开协议 catalog、guest typed client 和宿主 registry dispatch 三个稳定接缝。`protocol/hostservices`作为公开 catalog 集中维护 service、method、capability、资源类型、payload 形态、guest client 发布状态和 host dispatcher 发布状态。`internal/hostservice`descriptor 从公开 catalog 派生，不再维护第二份手写表。

普通领域 host service 默认使用统一 JSON envelope，减少每个领域新增专用`protowire`codec 的需求。`storage`、`cache`、`lock`、`data/recordstore`、`network`等有明确性能或资源需求的服务保留专用 codec。

宿主 WASM host service dispatch 改为显式注册的 registry 驱动。`wasm_host_service.go`入口只负责 envelope 解码、授权校验、上下文构造、registry lookup 和统一错误响应，不再维护 service 级大 switch。领域 handler 通过父包显式注册适配层接入 registry，避免扩大 WASM 私有执行上下文公开面。registry 注册使用显式装配函数，不得使用`init()`隐式注册。

descriptor 覆盖治理改为双向校验：descriptor 中声明发布的 guest client 和 dispatcher method 必须有对应实现；实现中出现的 service/method 也必须反向存在于 descriptor。README host service 表格由 descriptor 渲染器生成，漂移测试阻断未刷新文档的提交。

### 领域能力边界收敛

宿主侧领域能力实现组件从`hostservices`重命名为`capabilityhost`。动态普通领域能力只保留一个`ConfigureDomainHostServices(capability.Services)`配置入口，删除 AI、User、Org、Tenant 等领域专用`Configure*HostService`全局入口和 fallback 目录。

`pluginbridge/protocol`负责暴露动态`hostServices`公开协议描述并拥有 payload DTO 和 codec；`pluginbridge/internal/hostservice`负责 descriptor、授权推导、资源形态和清单规范化治理。二者都不拥有领域业务契约。

集合型领域的动态协议 service 名必须与`capability.Services`领域目录名称保持一致：`Users()`→`users`、`Files()`→`files`、`Jobs()`→`jobs`、`Notifications()`→`notifications`、`Plugins()`→`plugins`、`Sessions()`→`sessions`。不保留旧单数别名。

插件生命周期编排归属`plugins`领域能力，通过`host:plugins`能力和方法级授权暴露。插件自身配置读取从独立`config`收敛到`plugins.config.get`。通知发送从独立`notify`收敛到`notifications.messages.send`。定时任务统一归属`jobs`领域，动态插件旧 cron 声明能力迁移为`jobs.register`发现期声明能力，源码插件`Cron`入口迁移为`pluginhost.Jobs()`。

### 消费契约与 Provider SPI 分离

普通消费领域能力契约与源码插件 provider SPI、宿主内部 scope 接缝分离。`tenantcap`和`orgcap`父包只暴露普通消费`Service`、领域 DTO、值对象、错误码和常量；provider SPI、scope helper、request resolver 迁入`tenantspi`和`orgspi`子包。子包可以 import 父包复用 DTO；父包不得 import 子包。

`routecap.DynamicRouteMetadata`从`*ghttp.Request`改为`context.Context`；`apidoccap`删除依赖`ghttp`的 handler helper。非`*spi`的`capability/**`生产代码不得 import `gdb`或`ghttp`，`pluginbridge/**`生产代码不得 import 任何`*spi`子包。

provider factory 声明入口从能力包级`Provide()`和包级`defaultManager`迁移到`pluginhost.Declarations`的强类型 provider 声明分组。provider manager 由宿主启动装配层创建、持有并通过构造函数显式注入，不得由包级默认单例持有。

### AI 授权模型简化

动态插件`ai` host service 从`purpose`资源授权模型调整为`service + method`方法授权模型。`plugin.yaml`只声明`service: ai`和允许调用的`methods`，不再使用`resources`声明`purpose`和策略属性。`purpose`、`tier`、`maxOutputTokens`等参数由请求 DTO 提交，由 AI 能力服务及`linapro-ai-core`治理。主框架负责校验`service + method`授权、DTO 编解码、可信`pluginID`来源注入、结构化错误和错误脱敏。

动态插件普通领域 host service 覆盖源码插件`capability.Services`普通消费面中的领域能力。这些领域服务均采用方法授权型声明，运行时统一从宿主注入的同一个`capability.Services`目录进入对应`*cap.Service`。

### Guest 传输单轨化

guest host service client 统一为 invoker 注入式结构。根目录残留的`pluginbridge_hostcall_*_wasip1.go`逐域单例客户端、adapter 和镜像 stub 迁入`internal/domainhostcall`注入式客户端构造。`pluginbridge_directory.go`通过统一 invoker 装配基础能力和领域能力 guest client。非 WASI 不可用行为收敛到传输层`InvokeHostService`统一 stub。`recordstore`保持现有注入式执行文件，因为它承载查询计划执行领域逻辑。

### 资源能力领域能力化

`Cache`、`Lock`和`Storage`统一纳入`pkg/plugin/capability`领域能力目录。源码插件通过`pluginhost.Services`消费`cachecap.Service`、`lockcap.Service`和`storagecap.Service`；动态插件通过`pluginbridge`消费同一组领域接口。动态插件`hostServices`协议继续保留`service: cache`、`service: lock`和`service: storage`作为授权和 transport service 名，但这些协议名不再拥有业务契约。

动态插件分发层在完成`hostServices`授权校验后调用当前插件作用域的领域服务，不得直接调用底层`kvcache`、`hostlock`或本地文件目录实现。源码插件默认全信任，不要求在`plugin.yaml hostServices`中声明资源边界，但领域服务仍按插件 ID 和租户上下文隔离。

`Storage`新增`storagecap.Provider`和`storagecap.ProviderFactory`，允许主框架和源码插件提供对象存储后端。默认使用主框架内置本地磁盘 provider；配置 active provider plugin ID 时，只有该插件启用且 provider 构造成功才使用该 provider，不静默回退本地。本地磁盘 provider 在集群模式下必须提供明确诊断或阻断策略。`Storage.List`必须有明确 limit 上限，provider 实现不得无界遍历。

WASM 配置入口删除`ConfigureCacheHostService`、`ConfigureLockHostService`和`ConfigureStorageHostService`专用入口，统一通过`ConfigureDomainHostServices(capability.Services)`注入领域能力目录。

### 插件 UI、菜单和管理读模型

插件页面支持`iframe`、新标签页和宿主嵌入式挂载。插件菜单复用宿主菜单与角色授权体系，以`menu_key`作为生命周期锚点。源码插件通过代码注册的公开 HTTP 路由不自动投影为菜单、权限节点或 OpenAPI 路径。

插件管理列表拆分为分页摘要 DTO 和详情 DTO。`GET /plugins`只返回表格与行级操作可直接使用的最小字段；依赖检查、`hostServices`、授权快照、动态路由和 cron 声明等高成本数据只在打开详情或执行相关治理动作时按需加载。前端弹窗组件按需异步加载。缓存复用`plugin-runtime` revision/event 协调。

### 启动引导、依赖和升级

`plugin.autoEnable`从字符串列表演进为结构化条目，包含`id`和可选`withMockData`。启动引导在插件路由、cron 和动态前端包预热前执行，语义是按需安装再启用；列出的插件失败必须阻塞启动。

依赖治理收敛为硬插件依赖和框架版本约束。安装、启用、卸载、升级和发布切换前校验依赖存在性、版本范围、循环和反向依赖；插件清单不再维护软依赖或自动安装策略字段。

运行时升级显式分离发现版本和有效版本。启动扫描只标记`normal`、`pending_upgrade`、`abnormal`或失败状态，不自动升级。升级通过只读预览和有副作用执行 API 完成，执行链路按锁、状态重读、依赖校验、`BeforeUpgrade`、`Upgrade`、upgrade SQL、治理同步、release 切换、缓存失效和`AfterUpgrade`顺序推进。

### 运行时复杂度治理

runtime state 列表批量读取 manifest，不在 registry 循环中逐项调用`ScanManifests`。WASM host service 的 cache、config、session 等依赖由启动期显式注入，生产路径不得使用包级默认`New()`实例。runtime 必需依赖在构造或启动阶段校验，缺失时 fail fast。

### 工作区、构建和验证边界

官方插件工作区可选。host-only 模式下，宿主构建、测试和源码插件发现不要求`apps/lina-plugins`存在；plugin-full 模式下才要求官方插件工作区。插件工作区管理命令负责从`hack/config.yaml`来源安装、更新、诊断和锁定插件目录。

动态插件构建器支持`go:embed`资源声明，并在迁移期保留目录扫描回退。构建产物收敛到仓库根`temp/output/`或显式输出目录。发布产物必须保留与源码插件一致的 manifest、配置、metadata、SQL、i18n 和 public asset 路径语义。

### 运行时一致性和故障恢复

插件运行时变化必须按插件、sector、locale 或 global scope 精细失效 frontend bundle、runtime i18n、WASM 编译缓存、manifest 资源视图和默认配置视图。集群模式通过统一 coordination revision/event 唤醒其他节点和 reconciler；freshness 不可确认且超过窗口时采用 conservative-hide。

动态插件协调器在集群模式下按插件 ID 串行化共享副作用。协调器需要恢复 stale `reconciling`状态，并在 tick 边界隔离 panic。

## Risks / Trade-offs

- [Risk] 一次性全领域契约范围大，容易出现任务上下文过长或局部实现漂移。→ Mitigation：任务按契约、协议、领域适配、插件迁移、治理扫描和验证分层拆分。
- [Risk] 动态插件管理方法安装授权后可能被误解为绕过数据权限。→ Mitigation：规范明确安装授权只替代插件级菜单/RBAC 方法校验，领域方法仍强制执行租户、数据可见性、目标状态和审计校验。
- [Risk] 清零插件`sys_*`生产访问会影响多个官方插件并导致短期编译失败。→ Mitigation：先建立领域契约和宿主适配器，再逐插件迁移。
- [Risk] 包迁移范围大，容易遗漏测试替身或官方插件导入。→ Mitigation：先迁移公共接口，再让编译错误驱动调用方迁移，并用静态检索阻断旧路径残留。
- [Risk] 公共原语包可能再次膨胀为万能契约包。→ Mitigation：规格明确公共原语包不得定义具体能力 Service。
- [Risk] 列表和详情 DTO 拆分会触碰前端类型与多个弹窗入参。→ Mitigation：先建立清晰的类型边界，再逐个弹窗改为接收 detail 数据。
- [Risk] 协议生成引入工具复杂度。→ Mitigation：第一阶段允许先用 Go 描述表和覆盖测试，不强制一次性完成所有代码生成。
- [Risk] 缓存失效遗漏会让管理页显示过期插件状态。→ Mitigation：缓存键绑定`plugin-runtime` revision，写路径复用既有 revision/event。
- [Risk] 大规模移动代码可能掩盖行为回归。→ Mitigation：本次避免改变生命周期和 runtime upgrade 的控制流。

## Cross-Domain Impacts

- `cluster-deployment-mode`、`cluster-topology-boundaries`、`leader-election`和`distributed-locker`为插件运行时提供主节点、副本、revision/event 和 per-plugin 锁边界；当前契约由`openspec/specs`对应能力承载，历史 owner 为`archive/distributed-infra`。
- `cron-jobs`影响插件 cron 声明、主节点任务分类和动态任务接线；当前契约由`openspec/specs/cron-jobs/spec.md`与调度相关规范承载，历史 owner 为`archive/scheduled-jobs`。
- `menu-management`、`role-management`和`plugin-permission-governance`共同定义插件菜单、按钮权限和角色授权关系；插件侧历史保留`plugin-permission-governance`，菜单与角色通用契约由`archive/user-management`及主规范承载。
- `user-auth`只作为插件登录/登出 hook、鉴权上下文和失败隔离的交叉影响；当前认证契约由`openspec/specs/user-auth/spec.md`承载，历史 owner 为`archive/user-auth`。
- `online-user`和`server-monitor`只保留时长配置、会话触碰节流、监控采集和清理任务与插件运行时的交叉影响。
- `project-setup`、`release-image-build`和`e2e-suite-organization`只保留 host-only/plugin-full、工作台入口、插件 API 路径、发布和测试范围的交叉影响。
- `core-host-boundary-governance`、`module-decoupling`和`service-dependency-injection-governance`提供宿主边界、可选插件降级和显式依赖注入约束。
- `source-upgrade-governance`和`system-api-docs`分别影响源码插件升级内部化和动态插件 OpenAPI 投影。
- `config-duration-unification`和`demo-control-guard`分别影响插件配置读取、启动引导和演示只读插件安装方式。
