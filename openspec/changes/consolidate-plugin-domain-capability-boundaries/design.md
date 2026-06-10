## Context

`apps/lina-core`当前已经具备统一领域能力契约：`pkg/plugin/capability`定义`capability.Services`和各`*cap`领域接口，源码插件通过`pkg/plugin/pluginhost.Services`消费，动态插件通过`pluginbridge`的`hostServices`协议和`guest`客户端进入同一组领域语义。

现有问题主要来自实现与 transport 边界命名和遗留入口：

- 宿主侧领域能力实现位于`internal/service/plugin/internal/hostservices`，该名称容易和动态插件`plugin.yaml hostServices`协议混淆。
- `WASM`分发层已有`ConfigureDomainHostServices(capability.Services)`，但仍保留`ConfigureAITextHostService`、`ConfigureUserHostService`、`ConfigureOrgHostService`和`ConfigureTenantHostService`等领域专用全局入口。
- `pluginbridge/protocol`和`pluginbridge/internal/hostservice`需要继续作为动态`hostServices`协议目录存在，但必须明确`protocol`公开暴露协议名并拥有 payload DTO 和 codec，`internal/hostservice`拥有 descriptor、授权推导、资源形态和清单规范化治理，二者都不拥有领域业务契约。
- `pluginbridge`大部分普通领域能力已通过`pluginbridge/internal/domainhostcall`返回`capability/*cap`接口，`AI`仍在`guest_hostcall_ai.go`中维护 guest 专用接口集合，和`aicap.Service`平行。

本变更是架构边界收敛。项目没有兼容性负担，因此动态插件公开 service 名、`plugin.yaml hostServices`声明和 guest 公共目录可以按领域边界重构；数据库结构、HTTP API、前端用户体验和运行时用户可见文案不在本次范围内。

## Goals / Non-Goals

**Goals:**

- 明确`pkg/plugin/capability`是唯一领域能力契约入口。
- 明确`internal/service/plugin/internal/capabilityhost`是宿主侧`capability.Services`和`pluginhost.Services`实现组件。
- 明确`pkg/plugin/pluginhost`是源码插件消费能力的固定入口。
- 明确动态插件宿主侧普通领域能力只通过一个`ConfigureDomainHostServices(capability.Services)`进入`WASM`分发层。
- 明确`pluginbridge/protocol`负责暴露动态`hostServices`公开协议描述，并拥有 payload DTO 和 payload 编解码；`pluginbridge/internal/hostservice`负责 descriptor、授权推导、资源形态和清单规范化治理。
- 明确`pluginbridge`是动态插件公共消费入口，`pluginbridge/internal/domainhostcall`是普通领域 hostcall 代理实现位置，并收敛`AI`代理到`aicap.Service`语义。
- 统一集合型领域的协议服务名、能力字符串、源码文件名和引用常量，使其与`capability.Services`领域目录名称保持一致。
- 将`Plugins().Lifecycle()`明确为受治理的插件领域生命周期编排能力，使源码插件和动态插件在显式授权下使用同一个`plugincap.Service`领域对象。
- 将插件自身配置读取归属到`plugins`领域，动态插件通过`pluginbridge.Services.Plugins().Config()`和`plugins.config.get`访问，不再暴露独立`config`host service。
- 将通知读取和发送统一归属到`notifications`领域，动态插件通过`notifications.messages.batch_get`和`notifications.messages.send`访问，不再暴露独立`notify`host service。
- 将动态插件 cron 注册和源码插件`Cron`注册入口统一收敛到`jobs`领域，动态插件旧 cron 声明能力迁移为`jobs.register`发现期声明能力，避免定时任务存在多套公开领域对象。
- 增加治理验证，防止后续新增领域能力时出现新的平行实现、平行 guest 接口或领域专用 WASM 全局配置入口。

**Non-Goals:**

- 不新增、删除或重命名`capability.Services`中的顶层领域目录方法。
- 不改变`pluginbridge` protobuf/wire envelope 字段、host call envelope 或错误状态码；本变更会调整公开 service/method 名和 guest helper 目录。
- 不修改数据库、SQL、HTTP API、前端 UI、菜单、运行时文案或插件生命周期资源。
- 不引入通用 DI 容器、代码生成体系或新的动态插件能力发现框架。

## Decisions

### 1. 宿主实现组件命名为`capabilityhost`

将`internal/service/plugin/internal/hostservices`迁移到`internal/service/plugin/internal/capabilityhost`。该组件只负责把启动期共享的宿主服务实例适配成`capability.Services`和源码插件需要的`pluginhost.Services`，并继续通过`internal/service/plugin.NewHostServices`暴露给启动层。

选择`capabilityhost`而不是继续使用`hostservices`，是为了避免和动态插件`hostServices`协议目录混淆。选择迁移内部包而不是新增一层 wrapper，是因为当前问题是命名和归属不清，不需要为了重命名再制造额外转发层。

迁移后导入方向保持不变：

```text
internal/cmd
  -> internal/service/plugin.NewHostServices
      -> internal/service/plugin/internal/capabilityhost.New
          -> pkg/plugin/capability
          -> pkg/plugin/pluginhost
```

`internal/cmd`仍不得直接导入`plugin/internal/capabilityhost`。

### 2. 动态普通领域能力只保留一个能力目录配置入口

`WASM`普通领域 host service dispatcher 只保留`ConfigureDomainHostServices(capability.Services)`和`capabilityServicesForHostCall`。`AI`、`User`、`Org`和`Tenant`分发代码必须直接使用该共享目录，不再维护专用包级变量、专用`Configure*HostService`函数或 fallback 目录。

替代方案是保留专用入口但在内部转发到`domainHostServices`。该方案仍会让新增领域能力时出现“应该新增专用 Configure 还是复用 Domain Configure”的歧义，因此拒绝。

`data` host service 若为了数据范围过滤需要组织能力，也必须通过同一个`domainHostServices`按`pluginID`绑定后获取`Org()`，不得依赖组织专用全局目录。

### 3. 动态协议目录保留，但不承载领域业务语义

`pkg/plugin/pluginbridge/internal/hostservice`继续维护 host service descriptor 表，覆盖 service、method、capability、资源类型、payload 名称、guest client 和 dispatcher 同步点。`pkg/plugin/pluginbridge/protocol`继续作为公开协议出口，并实际拥有 payload DTO、wire payload struct、marshal 和 unmarshal 实现。

该目录的职责是动态插件 transport 和授权目录，不是第二套领域能力契约。新增领域方法时，应先在`pkg/plugin/capability/<domain>cap`确定领域接口和 DTO，再在`pluginbridge`增加对应 transport 描述和编解码。

### 4. 动态 guest 领域代理收敛到`pluginbridge/internal/domainhostcall`

动态插件开发者继续只依赖`pkg/plugin/pluginbridge`公共目录。普通领域能力的 hostcall 实现固定在`pluginbridge/internal/domainhostcall`下，公共`pluginbridge.Services`方法返回`capability/*cap.Service`或等价领域接口。

`AI`是当前例外。目标是让`pluginbridge.Services.AI()`返回或至少实现`aicap.Service`，并将各子能力 transport 实现迁入`domainhostcall`。如果某个`aicap`子接口包含`Available`、`Status`或`MethodStatus`等当前 guest 未实现的方法，guest 实现必须通过已有`AI`host service method、通用状态响应或安全降级错误补齐，不能继续定义一套平行的`pluginbridge.AITextService`、`pluginbridge.AIImageService`等领域接口。

### 5. 治理测试覆盖边界回归

新增或更新 Go 治理测试，至少覆盖以下回归：

- 生产代码中不得出现`ConfigureAITextHostService`、`ConfigureUserHostService`、`ConfigureOrgHostService`、`ConfigureTenantHostService`等领域专用 WASM 配置入口。
- `WASM`普通领域 dispatcher 获取`capability.Services`只能通过`ConfigureDomainHostServices`维护的共享目录。
- `internal/cmd`不得导入`plugin/internal/capabilityhost`。
- `pluginbridge`普通领域 hostcall 实现不得在公共包里重新定义与`capability/*cap`平行的领域接口；公共包只能保留目录方法、声明期启动 facade、资源型 host service client 和必要 transport facade。
- `pluginbridge/internal/hostservice`descriptor 仍覆盖 public protocol symbol、guest client、非 WASI stub 和 host dispatcher。

### 6. 集合型领域协议服务名与领域目录名称一致

动态`hostServices`协议中的普通领域服务名必须与`capability.Services`领域目录名称保持一致。对于集合型领域，协议 service、能力字符串、descriptor、guest 调用、`WASM`dispatcher 文件名和示例清单统一使用复数领域名：

| 领域目录 | 协议 service | 能力字符串 |
|----------|--------------|------------|
| `Users()` | `users` | `host:users` |
| `Files()` | `files` | `host:files` |
| `Jobs()` | `jobs` | `host:jobs` |
| `Notifications()` | `notifications` | `host:notifications` |
| `Plugins()` | `plugins` | `host:plugins` |
| `Sessions()` | `sessions` | `host:sessions` |

`auth`、`authz`、`apidoc`、`bizctx`、`dict`、`i18n`、`infra`、`route`、`ai`、`org`和`tenant`等本身就是命名空间或单一领域名，继续保持现有 service 名。由于本项目没有兼容性负担，不保留旧的`user`、`file`、`job`、`notification`、`plugin`或`session`别名，避免动态插件协议长期存在两套领域名。

### 7. 插件生命周期编排归属`plugins`领域能力

`plugincap.LifecycleService`不是插件自身生命周期回调注册接口，而是由租户或插件治理模块触发的宿主生命周期编排能力。该能力可影响租户删除、租户插件禁用等跨插件治理流程，因此必须作为`plugins`领域下的受治理方法暴露，而不能作为普通隐式能力无条件开放。

动态插件可以通过`hostServices`显式声明以下`plugins`方法：

| 方法 | 语义 |
|------|------|
| `lifecycle.tenant_plugin_disable.ensure` | 执行租户插件禁用前置检查 |
| `lifecycle.tenant_plugin_disabled.notify` | 执行租户插件禁用后置通知 |
| `lifecycle.tenant_delete.ensure` | 执行租户删除前置检查 |
| `lifecycle.tenant_deleted.notify` | 执行租户删除后置通知 |

运行时仍先校验`host:plugins`能力和授权快照中的精确 method，再进入`plugincap.LifecycleService`。源码插件和动态插件都通过同一个`plugincap.Service`领域对象访问`Config()`、`Registry()`、`State()`和`Lifecycle()`；动态`guest`公共包不得继续保留独立`PluginService`接口。

### 8. 插件自身配置读取归属`plugins`领域能力

插件自身配置不是独立宿主资源，也不是通用 transport 能力，而是插件领域对象的一项子能力。动态插件公共调用入口保持为`pluginbridge.Services.Plugins().Config()`，协议授权收敛为`service: plugins`和`method: config.get`，能力字符串统一使用`host:plugins`。

宿主侧删除公开`service: config`、`host:config`、`pluginbridge.ConfigHostService`和`dispatchConfigHostService`；`WASM`内部可以继续复用`plugincap.ConfigServiceFactory`作为`plugins.config.get`的实现来源，并必须保留 active artifact 默认配置通过`WithArtifactConfig(pluginID, artifactDefaultConfig)`覆盖读取的语义。

选择该方案而不是保留独立`config`别名，是因为插件开发者已经通过`Plugins().Config()`理解该能力，继续暴露`config`会形成“插件配置到底属于 config 还是 plugins”的长期歧义。

### 9. 通知发送归属`notifications`领域能力

通知读取和发送都属于`notifications`领域。动态插件读取消息使用`messages.batch_get`，发送消息使用`messages.send`；公共能力字符串统一使用`host:notifications`。`messages.batch_get`不需要资源引用，`messages.send`必须绑定授权的通知渠道资源引用，避免任何插件绕过渠道治理发送通知。

宿主侧删除公开`service: notify`、`host:notify`和`pluginbridge.Notify()`；guest 侧通过`pluginbridge.Services.Notifications()`返回的`notifycap.Service`完成读取和发送。由于同一领域内存在无资源读取和资源化发送，descriptor 必须使用方法级资源类型，不能把整个`notifications`服务粗暴标记为资源型。

### 10. 定时任务归属`jobs`领域能力

动态插件不再保留独立 cron 注册或发现期 host-call。定时任务的声明、生命周期、管理投影、执行和状态控制统一归属`jobs`领域，不能再通过`service: cron`、`pluginbridge.Services.Cron()`、包级`pluginbridge.Cron()`或内部 reserved `cron.register`协议表达。

旧动态插件 cron 的“声明式注册内置定时任务”能力必须迁移为`jobs`领域下的发现期方法：`service: jobs`、`method: jobs.register`。该方法只允许宿主在动态任务发现执行源中调用，用来收集插件内置 Jobs 声明；普通业务运行期调用`jobs.register`必须被拒绝。动态插件 guest 侧应通过`RegisterPlugin(plugin pluginbridge.Declarations)`中的`plugin.Jobs().Register(...)`声明任务，声明结果继续进入宿主 Jobs 管理投影、状态控制、handler 发布和调度执行链。

源码插件也必须通过`pluginhost.Jobs()`和`ExtensionPointJobsRegister`声明内置定时任务，不再暴露`pluginhost.Cron()`、`CronRegistrar`、`RegisterCron`或`cron.register`扩展点。源码插件的注册结果继续由宿主调度组件执行，但公开领域对象、扩展点名称、宿主管理投影和 handler 引用统一使用 Jobs 语义。动态插件如需交付内置任务，后续必须补齐正式`jobs`领域声明或安装期资源同步契约，而不是恢复独立 cron host service。

## Risks / Trade-offs

- [Risk] 包迁移会触发较多导入路径和测试 fixture 变更。  
  Mitigation：先做纯路径迁移并保持构造签名不变，再删除 WASM 遗留入口，最后收敛 guest AI；每一步运行对应包测试。

- [Risk] `AI`guest client 实现`aicap.Service`可能需要补齐状态类方法。  
  Mitigation：优先复用现有`aicap`子能力降级语义；若某些状态方法当前 transport 不支持，返回结构化不可用状态并在规范中记录不改变 wire 行为。

- [Risk] 删除领域专用 WASM 全局入口会影响现有测试直接配置局部服务的方式。  
  Mitigation：测试统一构造`capability.Services`替身并调用`ConfigureDomainHostServices`，每个测试保存并恢复包级全局状态，保持顺序无关。

- [Risk] 治理扫描过严可能误伤协议目录或资源型 host service client。  
  Mitigation：扫描只阻断普通领域能力分叉，明确允许`runtime`、`storage`、`network`、`data`、`cache`、`lock`、`hostconfig`、`manifest`等资源型或 transport 型 client；`config`和`notify`不再作为允许的公开资源型 client。

## Migration Plan

1. 迁移`hostservices`到`capabilityhost`，更新导入、包注释、测试包名和根 facade 注释。
2. 调整`ConfigureWasmHostServices`，只配置一次`ConfigureDomainHostServices`；删除`AI/User/Org/Tenant`专用 Configure 函数和全局变量。
3. 调整`wasm`分发和`datahost`相关能力解析，统一通过`capabilityServicesForHostCall`或等价共享目录获取插件作用域能力。
4. 将 guest 侧`AI`能力迁入`pluginbridge/internal/domainhostcall`并对齐`aicap.Service`。
5. 将集合型领域的动态协议 service 和能力字符串统一为`users`、`files`、`jobs`、`notifications`、`plugins`和`sessions`，同步 descriptor、guest 调用、WASM dispatcher、示例清单和文档。
6. 将动态`pluginbridge.Services.Plugins()`收敛为`plugincap.Service`，并通过`plugins` host service 方法级授权暴露`Lifecycle()`。
7. 将插件配置读取从独立`config`迁移到`plugins.config.get`，同步 descriptor、public protocol alias、guest 配置 adapter、`WASM`dispatcher、示例清单和文档。
8. 将通知发送从独立`notify.send`迁移到`notifications.messages.send`，同步方法级资源授权、guest 通知能力和`WASM`dispatcher。
9. 将定时任务注册从独立动态`cron`和源码插件`Cron`入口移出，源码插件改用`pluginhost.Jobs()`，动态插件改用`jobs.register`发现期声明能力，不恢复独立 cron host service。
10. 更新`pluginbridge`descriptor 覆盖测试、capability 边界治理测试和 WASM host service 测试。
11. 更新`apps/lina-core/pkg/plugin`README 中的边界说明。
12. 运行 OpenSpec 严格校验、相关 Go 包测试和必要静态检索。

回滚策略是按提交或文件迁移步骤回退到旧路径与旧 Configure 入口；由于不改变数据库和 wire 协议，回滚不需要数据迁移。

## 影响分析

- `i18n`：不新增运行时用户可见文案、菜单、路由、API 文档源文本或语言包；仅可能更新技术 README 和 OpenSpec 文档。确认无运行时`i18n`资源影响。
- 缓存一致性：不新增缓存、快照或失效机制；变更要求继续复用启动期共享`capability.Services`和其中已有的缓存敏感服务。确认无新的缓存一致性机制影响。
- 数据权限：集合型领域只修改动态插件授权 service 键名；新增`plugins` lifecycle governance 方法属于执行类治理动作，运行时先校验`host:plugins`和精确 method 授权，再进入既有`plugincap.LifecycleService`，具体租户、插件状态和跨插件生命周期前置检查继续由领域服务负责。`notifications.messages.send`必须同时校验`host:notifications`、精确 method 和授权渠道资源引用；`plugins.config.get`只读取当前插件自身配置；`jobs.register`仅在宿主动态 Jobs 发现期收集插件自身声明，不读取或写入租户业务数据，不新增跨插件或宿主业务数据访问，确认无数据权限绕行。
- 开发工具跨平台：不修改`Makefile`、脚本、`linactl`或 CI；治理验证使用 Go 测试或`openspec validate`，确认无开发工具跨平台影响。
- 测试策略：属于内部架构和插件桥接重构，无用户可观察 UI 变化，不新增 E2E。需要更新或新增 Go 单元测试、治理测试、静态检索、动态示例插件服务测试和相关包编译门禁。
- 模块启停：不改变组织、租户、`AI`等能力的启停与降级语义；缺失 provider 时仍由对应`*cap.Service`返回既有状态或不可用错误。
- 接口性能：不新增列表、批量、树形、聚合或导出接口；收敛后仍要求领域能力使用既有批量和投影契约，避免通过动态 dispatcher 循环调用单项服务。
- DI 来源：不新增运行期依赖 owner；`plugins.config.get`继续复用启动期注入的`plugincap.ConfigServiceFactory`，`notifications.messages.send`继续复用启动期共享通知能力或同一通知服务后端，不允许在 host call 路径临时创建独立服务图。

## Open Questions

无。实施阶段若发现`aicap.Service`中某个状态方法缺少动态 transport 支撑，应优先选择安全降级实现，并在任务记录中说明不改变 wire 行为的依据。
