## Context

`add-plugin-host-domain-capabilities`已经将宿主核心数据访问迁移为领域能力模型，并新增`usercap`、`dictcap`、`sessioncap`等组件。但当前公开包结构还保留了两类不一致：

- `capability.Services`同时返回`usercap.Service`等领域组件，以及`contract.APIDocService`、`contract.CacheService`、`contract.ConfigService`等具体服务接口。
- `capability`目录下同时存在`authzcap`、`dictcap`等`*cap`组件，以及`ai`、`bizctx`、`config`、`manifest`、`pluginstate`等非`*cap`命名组件。

这使`contract`包成为横向耦合点，也让包名无法稳定表达领域 owner。当前项目没有兼容负担，因此本变更采用一次性破坏性迁移，不保留旧包路径、旧接口别名或兼容转发层。

本变更属于`apps/lina-core`宿主通用能力和插件能力接缝重构，不属于工作台展示适配。实现必须继续保证源码插件和动态插件访问同一宿主能力时共享语义，动态`hostServices`仅作为 transport 和授权入口，不成为业务能力 owner。

## Goals / Non-Goals

**Goals:**

- 将`capability.Services`普通公开目录收敛为只返回各领域`*cap.Service`或对应组件的窄服务接口。
- 将`capability/contract`拆分为跨领域公共原语包和具体能力组件包，禁止公共原语包继续定义具体能力`Service`。
- 将`ai`、`bizctx`、`hostconfig`、`manifest`等公开能力组件统一重命名为`*cap`，将`tenantfilter`迁入`tenantcap`源码插件专用过滤子接口，并将插件自身配置、插件生命周期和插件状态能力收口到`plugincap`子领域。
- 将认证 token handoff 和授权能力收敛到`authcap`能力族，使用`authcap/token`和`authcap/authz`子领域维护窄接口，根目录只保留`Services.Auth()`入口。
- 将配置能力公开面收敛为两类：`Services.Plugins().Config()`读取当前插件自身配置，`Services.HostConfig()`读取宿主授权开放配置；根目录不再提供`Config()`或`PluginConfig()`。
- 保持动态插件`service + method`协议、授权快照、错误 envelope 和数据边界不变。
- 用静态检索、Go 编译门禁和 OpenSpec 严格校验确认旧`contract.*Service`具体服务引用和旧包路径不再进入公开能力目录。

**Non-Goals:**

- 不新增领域能力方法、动态插件`hostServices`方法、HTTP API 或数据库模型。
- 不改变`CapabilityContext`、批量缺失语义、数据权限、缓存一致性、`i18n`标签语义或领域错误语义。
- 不修改管理工作台页面、前端路由、菜单、按钮、语言包或 API 文档源文本。
- 不引入`capabilitycommon`、`capabilityutil`等语义模糊公共包。
- 不保留旧`capability/contract`具体服务兼容包，也不提供旧`capability/ai`到`aicap`的别名转发。
- 不向动态插件开放`*gdb.Model`、SQL 片段或数据库 query builder 形态的租户过滤能力；动态插件如需插件自有数据存储，后续必须通过单独的租户安全存储能力设计。

## Decisions

### 1. 公共原语包只承载跨领域值对象

新增或收敛一个小型公共原语包，例如`capability/capmodel`，用于承载跨领域且没有单一领域 owner 的值对象：

- `CapabilityContext`
- `DomainID`
- `CapabilityActor`
- `CapabilityAuthorizationSnapshot`
- `BatchResult`
- `PageRequest`
- `PageResult`
- `LocalizedLabel`
- `ProviderStatus`
- `CapabilityStatus`

该包不得定义`APIDocService`、`CacheService`、`ConfigService`、`ManifestService`、`PluginStateService`等具体能力服务接口。这样可以保留必要共享类型，同时避免把所有能力重新汇聚到另一个万能包。

拒绝方案是将`contract`简单改名为`capmodel`后保留全部内容。该方案只是换名，仍然让调用方依赖跨领域聚合包，不解决具体能力 owner 不清晰的问题。

### 2. 具体能力接口迁移到各自`*cap`组件

每个插件公开能力组件拥有自己的`Service`、`Factory`、DTO、错误语义和必要适配器。例如：

| 当前包 | 目标包 | 主要职责 |
|--------|--------|----------|
| `capability/ai` | `capability/aicap` | `AI`能力族聚合与子能力访问 |
| `contract.AuthService`和`capability/authzcap` | `capability/authcap/token`和`capability/authcap/authz`，由`capability/authcap`聚合 | 认证 token handoff 与授权能力族 |
| `capability/bizctx`和`contract.BizCtxService` | `capability/bizctxcap` | 当前请求业务上下文投影 |
| `capability/config`和`contract.ConfigService` | `capability/plugincap`子领域 | 当前插件静态配置读取 |
| `capability/hostconfig`和`contract.HostConfigService` | `capability/hostconfigcap` | 授权宿主配置读取 |
| `capability/manifest`和`contract.ManifestService` | `capability/manifestcap` | 当前插件`manifest/`资源读取 |
| `capability/pluginlifecycle`和`contract.PluginLifecycleService` | `capability/plugincap`子领域 | 插件生命周期治理调用 |
| `capability/pluginstate`和`contract.PluginStateService` | `capability/plugincap`子领域 | 插件状态和启用性查询 |
| `capability/tenantfilter`和`contract.TenantFilterService` | `capability/tenantcap`源码插件专用子接口 | 源码插件自有表租户过滤 |

`capability.Services`普通消费面不得暴露携带数据库 query builder 的租户过滤能力；`tenantcap.PluginTableFilterService`只通过`pluginhost.Services.TenantFilter()`等源码插件专用受控接缝提供。拒绝新增独立`tenantfiltercap`包，因为该能力没有脱离租户领域的独立 owner；也拒绝将其放入`Services.Tenant().Filter()`，因为`Services.Tenant()`属于普通插件和动态插件都可见的租户消费面，不能携带源码插件专用的`*gdb.Model`查询构造器。

认证和授权属于同一安全访问控制能力族，但 token 生命周期和权限治理仍是两个窄子领域。`authcap.Service`只作为领域聚合入口，返回`Token()`和`Authz()`子服务；业务服务仍应通过构造函数接收`authcap/token.Service`或`authcap/authz.Service`等最窄接口，不得为了目录收敛而长期保存整个认证授权聚合目录。

### 3. `Plugins()`承载插件领域子能力

插件自身配置、插件生命周期、插件状态和插件治理投影都属于插件领域。目标形态是根目录只暴露`Plugins()`插件领域命名空间，插件相关子能力通过该命名空间访问：

```go
type PluginService interface {
    Config() plugincap.ConfigService
    State() plugincap.StateService
    Lifecycle() plugincap.LifecycleService
    Registry() plugincap.RegistryService
}
```

`Plugins().Config()`只读取当前插件自身配置，例如当前插件作用域下的`config.yaml`或 artifact 内配置。`Plugins().State()`只读取插件启用状态和 provider 可用性快照。`Plugins().Lifecycle()`只承载插件生命周期前置校验和通知。`Plugins().Registry()`承载现有插件投影、租户插件列表等插件治理读取能力。

拒绝在根目录保留`PluginConfig()`、`PluginLifecycle()`和`PluginState()`，因为它们让插件领域能力分散在多个根入口，增加开发者理解成本。

### 4. 配置公开面只保留插件自身配置和宿主配置

配置能力公开面只保留两类：

| 入口 | 职责 | 作用域 |
|------|------|--------|
| `Services.Plugins().Config()` | 读取当前插件自身配置 | 当前`pluginID` |
| `Services.HostConfig()` | 读取宿主授权开放配置 | 宿主配置树 |

根目录不再提供`Config()`，也不再把运行时配置领域作为普通插件公开入口。若未来确实需要向插件公开“运行时配置中心”能力，必须使用新的明确领域名重新设计，例如`Settings()`或`RuntimeSettings()`，并单独说明数据权限、缓存和管理语义。

### 5. `Services`方法名按领域语义命名

Go 包名和领域入口名采用不同规则：

| 层次 | 命名规则 | 示例 |
|------|----------|------|
| 组件包名 | 单数领域名加`cap`后缀，表达能力 owner | `usercap`、`jobcap`、`plugincap` |
| 根服务入口 | 按插件开发者看到的领域语义命名 | `Users()`、`Jobs()`、`Plugins()`、`Tenant()` |
| 单一上下文或配置能力 | 使用单数或专有名词 | `BizCtx()`、`HostConfig()`、`AI()` |

拒绝为了和包名机械一致而将所有根入口改为单数。`Services.User()`容易被误解为当前用户对象，`Services.Plugin()`容易被误解为当前插件对象，而`Services.Users()`、`Services.Plugins()`更准确表达资源集合或领域命名空间。包名继续使用`*cap`统一能力组件归属，根入口只按领域消费语义保持直观。

### 6. `capability.Services`只聚合领域命名空间，不承载实现细节

目标形态是根目录只表达普通插件服务目录：

```go
type Services interface {
    APIDoc() apidoccap.Service
    Auth() authcap.Service
    AI() aicap.Service
    BizCtx() bizctxcap.Service
    Cache() cachecap.Service
    HostConfig() hostconfigcap.Service
    I18n() i18ncap.Service
    Manifest() manifestcap.Service
    Route() routecap.Service
    Users() usercap.Service
    Dict() dictcap.Service
    Files() filecap.Service
    Infra() infracap.Service
    Jobs() jobcap.Service
    Notifications() notifycap.Service
    Org() orgcap.Service
    Plugins() plugincap.Service
    Sessions() sessioncap.Service
    Tenant() tenantcap.Service
}
```

`Auth()`返回认证授权能力族聚合入口，目标形态是`Services.Auth().Token()`访问租户 token handoff 和 impersonation token 能力，`Services.Auth().Authz()`访问授权查询能力。`AdminServices`同步通过认证授权管理子目录暴露`Auth().Authz()`，而不是继续在管理根目录并列暴露`Authz()`。源码插件业务对象仍应通过构造函数接收最窄接口，而不是长期保存完整服务目录或认证授权聚合目录。

源码插件专用服务目录在`pluginhost.Services`扩展普通服务目录：

```go
type Services interface {
    capability.Services
    Admin() capability.AdminServices
    TenantFilter() tenantcap.PluginTableFilterService
}
```

`TenantFilter()`的接口归属在`tenantcap`，但入口保留在`pluginhost.Services`，用于表达它只服务源码插件自有表查询，不能进入普通`capability.Services.Tenant()`。

### 7. 动态插件租户过滤必须在 host service 边界隐式执行

动态插件不能直接使用`tenantcap.PluginTableFilterService`。该接口包含`*gdb.Model`，只能服务源码插件在宿主进程内查询插件自有表，不适合 WASM guest、语言无关协议或动态插件授权快照。

动态插件可使用普通`tenantcap.Service`读取当前租户和校验租户可见性：

```go
tenantID, err := guest.Default().Tenant().Current(ctx)
err = guest.Default().Tenant().EnsureTenantVisible(ctx, tenantID)
```

当动态插件调用用户、文件、通知、配置、插件列表等宿主能力时，对应 host service handler 必须基于调用身份、`pluginID`、当前`tenantID`、授权快照和既有数据权限边界在宿主侧完成过滤。动态插件不得拼接`tenant_id`条件，不得传入 SQL、DAO、`*gdb.Model`或 query builder。若未来需要动态插件自有数据存储，应新增类似`Plugins().Storage()`的租户安全存储能力，由宿主按`pluginID + tenantID`自动隔离。

### 8. 动态插件协议不随 Go 包重命名变化

Go 包重命名不得改变动态插件的语言无关协议。`plugin.yaml hostServices`仍使用`service: ai`、`service: user`、`service: config`等领域服务名，运行时授权快照仍按`service + method + resource`校验。

guest SDK 可以重命名 Go 入口，例如`guest.Default().AI()`对应`aicap`包，`guest.Default().Plugins().Config()`对应插件自身配置读取；但 transport payload、protobuf envelope、错误 envelope 和授权语义保持不变。动态协议中的`service: config`可继续表示插件自身配置读取，不能因为 Go 侧收口到`Plugins().Config()`而改成`service: plugincap`。

### 9. 不保留兼容层

本项目明确没有历史兼容负担。旧`capability/contract`具体服务接口、旧`capability/ai`包、旧`capability/config`包、根`Config()`、根`PluginConfig()`、根`PluginLifecycle()`和根`PluginState()`等迁移后直接删除，不通过类型别名、转发函数或空壳包保留。这能让遗漏调用点在编译阶段暴露，避免长期双路径。

### 10. 影响分析

| 领域 | 判断 |
|------|------|
| `i18n` | 不修改运行时用户可见文案、菜单、路由、API 文档源文本或语言包；实现记录中说明无影响。 |
| 数据权限 | 不新增读取、写入、导出、聚合或授权方法；领域方法既有`CapabilityContext`和数据边界不变。动态插件不直接获得租户过滤器，宿主 host service handler 继续在调用边界执行与宿主 API 等价的数据权限和租户边界。 |
| 缓存一致性 | 不新增缓存或失效机制；缓存敏感服务仍必须复用启动期共享实例或共享后端。 |
| SQL/DAO | 不新增 SQL、Seed、Mock 数据或 DAO 生成范围。 |
| HTTP API | 不修改路由、DTO、OpenAPI 元数据或权限标签。 |
| 前端 UI | 不修改页面、组件、菜单、交互或样式；不新增 E2E。 |
| 开发工具 | 默认不新增脚本或跨平台入口；若补治理扫描必须使用 Go 工具或现有跨平台入口。 |

## Risks / Trade-offs

- [Risk] 包迁移范围大，容易遗漏测试替身或官方插件导入。→ Mitigation：先迁移公共接口，再让编译错误驱动调用方迁移，并用静态检索阻断旧路径残留。
- [Risk] 公共原语包可能再次膨胀为万能契约包。→ Mitigation：规格明确公共原语包不得定义具体能力`Service`，新增具体能力时必须创建或使用对应`*cap`组件。
- [Risk] `tenantcap.PluginTableFilterService`携带数据库 query builder，若误放进普通`capability.Services.Tenant()`会污染普通插件和动态插件消费面。→ Mitigation：规格明确其只允许通过`pluginhost.Services.TenantFilter()`源码插件专用受控接缝暴露，并保留静态检查。
- [Risk] 将源码插件过滤接口迁入`tenantcap`可能让`tenantcap`包职责膨胀。→ Mitigation：`tenantcap.Service`继续只表达普通租户能力，`PluginTableFilterService`、`ScopeService`等专用接口以独立文件和清晰注释区分可见性，不新增`Tenant().Filter()`转发型抽象。
- [Risk] 动态 guest SDK Go 入口改名可能被误解为协议变更。→ Mitigation：设计和测试同时断言`hostServices`声明、`service`、`method`、授权快照和 envelope 不变。
- [Risk] 旧规格中存在`pluginservice`历史术语，容易让实现回退到旧包心智。→ Mitigation：本变更修改相关规格，将框架能力归属收敛到`pkg/plugin/capability/<domain>cap`。

## Migration Plan

1. 梳理`capability/contract`中所有类型，按公共原语、具体能力服务、内部 helper、测试 helper 分类。
2. 新增公共原语包并迁移跨领域值对象，更新所有领域`*cap`包引用。
3. 按能力逐个创建或重命名`*cap`组件，把具体服务接口、factory、adapter 和测试迁入对应包；插件自身配置、插件生命周期和插件状态迁入`plugincap`子领域；`tenantfilter`迁入`tenantcap.PluginTableFilterService`，不新增`tenantfiltercap`。
4. 调整`capability.Services`、`AdminServices`、`pluginhost.Services`、guest SDK、hostservices directory、`WASM`host service 和启动装配；认证授权能力使用`authcap`能力族子目录，源码插件通过`pluginhost.Services.TenantFilter()`使用租户过滤，动态插件不暴露该接口。
5. 迁移官方插件导入路径和测试替身；修改任一插件目录前先检查插件根目录`AGENTS.md`。
6. 删除旧`contract`具体服务、旧非`*cap`包、根`Config()`、根`PluginConfig()`、根`PluginLifecycle()`、根`PluginState()`和兼容转发层。
7. 运行变更范围 Go 编译门禁、静态检索、OpenSpec 严格校验和 diff 检查。

回滚方式是恢复旧包路径和`contract.*Service`引用。由于不涉及数据库、HTTP API 或前端资源，回滚不需要数据迁移。

## Open Questions

无。建议实现阶段直接采用`capmodel`作为公共原语包名；若实现时发现该命名与现有项目术语冲突，可以在不改变上述边界的前提下替换为更准确的短名。
