## Context

当前代码已经把插件相关公共组件迁入`apps/lina-core/pkg/plugin`命名空间，但职责边界仍不够硬：

- `pkg/plugin/capability`是插件消费宿主能力目录，但`orgcap.Service`、`tenantcap.Service`等普通消费接口仍暴露`*gdb.Model`、`*ghttp.Request`、数据权限查询注入、用户组织/租户关系写入、自动开通和启动一致性等宿主内部能力。
- `pkg/plugin/pluginhost`是源码插件贡献入口，但仍通过`HostServices = capability.Services`和`HostServicesForPlugin`表达宿主能力目录语义。
- `pkg/plugin/pluginbridge`是动态插件 ABI 和 transport 层，但根包仍大量暴露业务 host service DTO、capability 常量和 guest ability client，容易被当作宿主能力 owner。
- `internal/service/plugin`是宿主内部插件运行时服务，但它和公共`pkg/plugin`之间还缺少足够明确的“谁拥有契约、谁拥有实现、谁负责装配”的边界。

本变更以“无历史兼容负担”为前提：旧 facade、旧方法、过宽接口和误导性类型可以直接删除，不保留兼容包或转发方法。

需求澄清结论：当前迭代的目标是“破坏性公共契约收敛”。旧兼容面直接删除，普通插件消费面只保留只读、DTO、状态、批量投影和可降级能力；`pluginbridge`根包只保留动态插件协议与 transport 能力；本迭代不主动改变 HTTP API、前端页面、数据库 schema、运行时文案或插件清单字段。若实现阶段发现插件确实缺少必要读能力，只能新增 DTO 化、批量化、只读方法，不得恢复`*gdb.Model`、`*ghttp.Request`、DAO、DO、Entity、写入接口、数据范围注入或宿主内部治理接口。

## Goals / Non-Goals

**Goals:**

- 将`pkg/plugin/capability`收敛为唯一插件消费宿主能力目录，普通消费面只暴露状态、DTO、批量投影、只读查询和可降级能力。
- 将`orgcap`、`tenantcap`拆为普通消费接口、provider-facing接口和宿主内部治理接口，禁止普通插件通过能力目录拿到宿主数据库模型、HTTP请求对象或写入/数据范围注入能力。
- 将 provider factory 的环境从`Services any`改为每个 capability 的强类型窄环境。
- 将`pluginhost`收敛为源码插件贡献 API，删除`HostServices`历史别名和重复 scoped factory。
- 将`pluginbridge`收敛为动态插件 ABI、host call、host service wire、artifact section、codec 和 route dispatcher；动态插件业务能力 client 归属`capability/guest`。
- 增加治理扫描和编译验证，确保包边界和导入方向不回退，并确保旧入口名、旧方法名、旧业务 client、`contract.ProviderEnv.Services`和插件导入宿主`internal/**`均无生产残留。

**Non-Goals:**

- 不改变动态插件`plugin.yaml`中的`hostServices`字段名和授权快照语义。
- 不重写插件安装、启用、禁用、卸载、升级的业务流程。
- 不新增通用 DI 容器、服务定位器或运行时反射注册中心。
- 不在本变更中改变 HTTP API 响应字段、前端插件管理页面交互或数据库表结构，除非实现阶段发现删除旧契约必须同步清理相关投影。
- 不为旧`pkg/pluginservice`、旧`pkg/pluginhost`、旧`pkg/pluginbridge`、旧`pkg/plugindb`或旧方法保留兼容 facade。

## Decisions

### 1. 能力目录只暴露普通插件消费面

`capability.Services`只返回插件安全消费接口，例如：

```go
type Services interface {
    APIDoc() contract.APIDocService
    Auth() contract.AuthService
    BizCtx() contract.BizCtxService
    Cache() contract.CacheService
    Config() contract.ConfigService
    HostConfig() contract.HostConfigService
    I18n() contract.I18nService
    Manifest() contract.ManifestService
    Notify() contract.NotifyService
    Org() orgcap.Service
    PluginLifecycle() contract.PluginLifecycleService
    PluginState() contract.PluginStateService
    Route() contract.RouteService
    Session() contract.SessionService
    Tenant() tenantcap.Service
}
```

普通`orgcap.Service`和`tenantcap.Service`不得包含`*gdb.Model`、`*ghttp.Request`、写入关系、启动一致性、自动开通、底层查询注入或 provider 生命周期控制方法。

备选方案是继续用一个宽接口并依靠注释约束调用方。该方案无法在编译期阻止普通插件误用，也会让动态 guest、源码插件和宿主内部测试共享过大的替身，因此不采用。

### 2. `orgcap`和`tenantcap`拆成多个窄接口

目标接口组如下：

- `orgcap.Service`：普通消费面，包含`Available`、`Status`、批量用户部门投影、部门树、岗位选项等 DTO 方法。
- `orgcap.Provider`：provider 插件实现面，保留 provider 必须实现的组织能力。
- `orgcap.ScopeService`：宿主内部数据权限接缝，承载`ApplyUserDeptScope`、`BuildUserDeptScopeExists`等`*gdb.Model`方法。
- `orgcap.AssignmentService`：宿主内部用户组织关系写入接缝。
- `tenantcap.Service`：普通消费面，包含当前租户、可用性、状态、租户列表、切换校验和可见性校验等 DTO 方法。
- `tenantcap.RequestResolver`：宿主中间件内部 HTTP 租户解析接缝，承载`*ghttp.Request`。
- `tenantcap.ScopeService`：宿主内部租户数据过滤接缝，承载`*gdb.Model`方法。
- `tenantcap.UserMembershipService`：宿主用户、角色、通知等模块使用的租户成员关系投影和写入接缝。
- `tenantcap.PluginProvisioningService`：宿主插件运行时使用的租户插件自动开通接缝。
- `tenantcap.StartupConsistencyService`：宿主启动一致性检查接缝。

宿主内部 service 通过构造函数显式注入需要的窄接口。测试替身也按窄接口构造，避免每个测试为了满足大接口实现大量无关方法。

### 3. Provider factory 使用强类型窄环境

删除`contract.ProviderEnv.Services any`。每个 capability 自己定义 provider 构造环境，例如：

```go
type orgcap.ProviderEnv struct {
    PluginID string
    TenantFilter contract.TenantFilterService
    I18n contract.I18nService
}

type tenantcap.ProviderEnv struct {
    PluginID string
    BizCtx contract.BizCtxService
    PluginLifecycle contract.PluginLifecycleService
}
```

provider adapter 只能拿到其声明需要的能力，不能拿到完整`capability.Services`后自行扩展依赖。

备选方案是把`Services any`改成`capability.Services`。该方案虽然去掉了类型断言，但仍把完整能力目录交给 provider adapter，无法约束 provider 依赖扩张，因此不采用。

### 4. `pluginhost`不再拥有 HostServices 语义

`pluginhost`保留源码插件贡献职责：assets、routes、hooks、cron、lifecycle、governance和 provider factory declaration。registrar 和 payload 中需要访问宿主能力时，直接返回`pluginhost.Services`源码插件运行期服务目录，方法名使用`Services()`。

删除：

- `type HostServices = capability.Services`
- `ScopedHostServicesFactory`
- `HostServicesForPlugin`
- registrar/payload 中的`HostServices()`方法

所有源码插件更新为使用`registrar.Services()`或 callback payload 的`Services()`，避免把服务目录访问器命名为`Capabilities()`而混淆动态插件能力授权集合。

### 5. `pluginbridge`不再发布业务能力 client

`pluginbridge`保留动态插件协议和 transport 职责：

- ABI 常量和 bridge contract。
- request/response envelope codec。
- WASM custom section 读写。
- host call 协议。
- host service wire DTO、service/method字符串、授权声明校验和编解码。
- guest route dispatcher 和 controller binding。

动态插件业务能力 client，例如`Runtime()`、`Storage()`、`Network()`、`Config()`、`Manifest()`、`Org()`、`Tenant()`，统一归属`pkg/plugin/capability/guest`。数据能力通过`guest.Directory.Data()`统一获取受治理的`capability/data` facade。旧`pluginbridge.Runtime()`等入口直接删除。

`pluginbridge`根包可保留最小协议 facade，但不得转发业务能力 client，不得被宿主能力目录反向依赖。

### 6. 宿主内部插件运行时保留实现 owner

`apps/lina-core/internal/service/plugin`继续拥有插件 catalog、runtime、lifecycle、sourceupgrade、datahost、wasm host functions、resourcefs 和管理端投影。它可以使用`pkg/plugin`公共契约，但公共`pkg/plugin`不得导入`internal/service/plugin`。

`internal/service/pluginhostservices`作为宿主运行期适配器组装层保留，但其对外返回值应是`capability.Services`，并由启动期显式注入共享服务实例。它不得创建第二套 cache、plugin state、tenant、org 或 i18n 服务实例。

### 7. 数据权限和性能以宿主内部窄接口承载

普通插件消费面只暴露批量 DTO 和投影方法，不暴露数据库模型注入。宿主内部需要数据库侧过滤时，通过`ScopeService`等内部接口注入 provider 实现，并继续在数据库查询阶段完成数据权限过滤。

动态插件访问数据统一从`capability/guest.Directory.Data()`取得`capability/data` facade，并继续通过`hostServices`授权快照执行，不允许 raw SQL、`*gdb.Model`、DAO、DO 或 Entity 出现在 guest SDK、普通 capability DTO 或 pluginbridge 业务 facade 中。

### 8. 治理扫描作为完成门禁

实现阶段必须补充静态治理，至少覆盖：

- `apps/lina-core/pkg/plugin/**`不得 import `apps/lina-core/internal/service/plugin/**`。
- `apps/lina-plugins/**`不得 import `apps/lina-core/internal/**`，官方插件受控启动装配和测试 fixture 除外。
- 普通插件目录不得调用`orgcap.ScopeService`、`tenantcap.ScopeService`等宿主内部接口。
- `pluginbridge`根包不得暴露`Runtime()`、`Storage()`、`Network()`、`Data()`、`Cache()`、`Config()`等业务能力 client。
- 生产代码不得继续引用`pluginhost.HostServices`、`HostServicesForPlugin`或`contract.ProviderEnv.Services`。
- 生产代码不得继续引用旧`HostServices()`访问器、旧`pluginbridge`业务 client 或旧 provider env 兼容字段；如扫描结果只剩 OpenSpec 记录、负向测试或注释，必须在任务记录中说明范围。

治理入口优先通过 Go 工具或`linactl`实现，避免新增平台专属脚本。

## Risks / Trade-offs

- **迁移面大，编译失败点多** → 按接口拆分、宿主注入、官方插件迁移、动态 guest 迁移、旧面删除的顺序推进，并用 Go 编译门禁覆盖宿主、公共包、官方源码插件和动态插件样例。
- **普通消费面过窄影响插件开发体验** → 只移除底层实现和宿主内部治理能力；常用只读能力提供 DTO、批量投影和状态方法，避免插件为了展示列表而逐项调用宿主详情。
- **provider adapter 需要更多显式环境字段** → 每个 provider 只声明自身需要的能力，构造签名变长但依赖更可见，符合当前显式依赖注入规则。
- **`pluginbridge`删除业务能力入口会影响动态插件样例和构建器** → 同步迁移到`capability/guest`，并用 WASI 构建 smoke 验证 guest 依赖图。
- **缓存和 provider 可用性可能出现第二套状态** → provider 可用性继续只读取插件 enabled snapshot 和 runtime revision，禁止新增独立 active 标记；验证单机和集群路径不创建孤立缓存实例。

## Migration Plan

1. 新增拆分后的`orgcap`、`tenantcap`窄接口和强类型 provider env，先让宿主内部实现同时满足新接口。
2. 调整`capability.Services`返回普通消费接口；更新`pluginhostservices`目录构造和 scoped 目录实现。
3. 将宿主内部`datascope`、`user`、`role`、`notify`、`middleware`、`plugin`等调用方改为显式注入`ScopeService`、`UserMembershipService`、`ProvisioningService`等窄接口。
4. 更新`linapro-org-core`、`linapro-tenant-core` provider adapter，去掉`env.Services.(capability.Services)`类型断言，改用强类型 env。
5. 删除`pluginhost.HostServices`别名、scoped factory和`HostServices()`方法，源码插件 registrar 调用迁移到`Services()`。
6. 将动态插件业务能力调用迁移到`capability/guest`，删除`pluginbridge`根包业务能力入口。
7. 删除旧接口、旧方法、旧别名和过渡代码，运行静态治理扫描确认无生产残留。
8. 执行 Go 编译门禁、官方源码插件编译、动态插件 WASI 构建 smoke、linactl 构建器测试、OpenSpec 严格校验和格式检查。

回滚策略：本项目不保留兼容路径。若实现阶段发现某个接口拆分不合理，修正目标窄接口后继续，不恢复旧宽接口或旧 facade。

## Open Questions

1. 普通插件是否需要读取组织/租户能力的写前校验类方法，例如“某些用户是否属于某租户”或“某部门是否可见”？建议只提供 DTO 化批量校验方法，不暴露数据库模型注入或底层关系写入。
2. `pluginbridge`根包是否允许保留纯协议 facade，例如`NormalizeHostServiceSpecs`和`ReadCustomSection`？建议允许保留协议 facade，但删除所有业务能力 client。
3. `TenantFilter`是否继续放在普通`capability.Services`中？建议保留为源码插件路由内部受治理能力，但不要放入动态 guest 普通消费面，也不要暴露`*gdb.Model`之外的任意查询拼接能力。
