## Why

当前`apps/lina-core/pkg/plugin/capability`已经完成领域能力扩展，但公开能力目录仍混用`*cap`领域组件和`contract.*Service`具体接口，`contract`包因此成为事实上的万能耦合点。继续保留这种形态会削弱本次领域对象重构的宿主边界，使源码插件、动态插件和宿主适配器难以从包名直接判断能力 owner、数据边界和错误语义。

项目没有历史兼容负担，适合通过独立变更一次性把插件公开能力包命名和公共原语归属收敛到清晰、可治理的目标结构。

## What Changes

- **BREAKING**：删除`capability/contract`作为具体能力接口聚合包的角色，具体能力接口迁移到各自职责明确的`*cap`组件包。
- **BREAKING**：将现有非`*cap`领域能力组件统一重命名或收口到明确领域包，例如`ai`迁移为`aicap`，`bizctx`迁移为`bizctxcap`，`hostconfig`迁移为`hostconfigcap`，`manifest`迁移为`manifestcap`；`tenantfilter`不再作为独立包保留，迁入`tenantcap`源码插件专用过滤子接口；插件自身配置、插件生命周期和插件状态能力收口到`plugincap`子领域。
- **BREAKING**：根`capability.Services`不再暴露`Config()`、`PluginConfig()`、`PluginLifecycle()`或`PluginState()`；插件相关能力通过`Services.Plugins()`命名空间访问，其中插件自身配置为`Services.Plugins().Config()`，宿主配置读取保留为`Services.HostConfig()`。
- 新增或收敛一个小型公共领域原语包，用于承载`CapabilityContext`、`DomainID`、`BatchResult`、`PageRequest`、`CapabilityStatus`等跨领域值对象；该包不得定义具体能力`Service`接口。
- 调整`capability.Services`和`pluginhost.Services`，使普通公开能力目录只返回各领域能力命名空间或`*cap.Service`，不再返回`contract.*Service`，也不再暴露含混的根`Config()`入口；`Services`方法名按领域语义命名，资源集合领域可继续使用`Users()`、`Jobs()`、`Plugins()`等复数入口，组件包名继续使用`usercap`、`jobcap`、`plugincap`等单数`*cap`包名。
- 将源码插件自有表`tenant_id`过滤接口归属到`tenantcap.PluginTableFilterService`，入口继续保留在`pluginhost.Services.TenantFilter()`；该能力不得进入普通`capability.Services.Tenant()`，也不得作为动态插件 host service 暴露。
- 调整源码插件 hostservices 适配器、动态插件`WASM`host service、guest SDK、官方插件和测试桩的导入路径与接口实现。
- 保持动态插件`hostServices`语言无关`service + method`协议、授权快照、错误 envelope、数据权限、缓存一致性和运行时语义不变；动态插件租户隔离由普通`tenantcap.Service`和宿主 host service handler 在调用边界隐式执行，不向 guest 暴露`*gdb.Model`过滤器。
- 不新增数据库迁移、HTTP API、前端页面、菜单、运行时用户文案或插件资源。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `plugin-capability-boundary-governance`：要求`capability`公开能力目录按`*cap`组件包返回能力接口，并禁止`contract`包承载具体能力服务。
- `plugin-package-boundary-governance`：要求`pkg/plugin/capability`子包命名统一为职责明确的`*cap`能力组件，公共原语包不得成为能力服务聚合点，并要求`tenantfilter`收敛到`tenantcap`而不是新增独立`tenantfiltercap`。
- `plugin-host-service-extension`：要求源码插件 hostservices 与动态插件 host service handler 适配到新的`*cap`能力组件，同时保持动态协议语义不变，并明确动态插件不直接消费源码插件专用租户过滤器。
- `framework-capability-registry`：将框架能力归属从历史`pluginservice`表述收敛到当前`pkg/plugin/capability/<domain>cap`组件体系。

## Impact

- 影响`apps/lina-core/pkg/plugin/capability/**`公开包、`capability.Services`、`capability.AdminServices`、`pluginhost.Services`和`capability/guest`目录。
- 影响`apps/lina-core/internal/service/plugin/internal/hostservices`、`apps/lina-core/internal/service/plugin/internal/wasm`、`internal/cmd`启动装配、测试替身和相关静态治理测试。
- 影响官方源码插件中导入`capability/contract`、`capability/ai`、`capability/config`、`capability/bizctx`等旧路径的生产代码和测试代码；修改插件目录前必须按插件本地`AGENTS.md`优先级检查本地规范。
- 动态插件协议层不改变`plugin.yaml hostServices`声明、`service`、`method`、授权快照或 protobuf envelope；只调整 Go 公共包和 guest SDK 入口。
- `i18n`、数据权限、缓存一致性、SQL、HTTP API、前端 UI 和开发工具跨平台预期无行为影响；实现和审查必须记录无影响判断，并以 Go 编译门禁、静态检索、OpenSpec 严格校验和必要单元测试验证。
