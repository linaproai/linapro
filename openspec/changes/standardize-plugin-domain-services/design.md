## Context

当前插件领域能力已经形成`pkg/plugin/capability`、`pluginhost`、`pluginbridge`和动态`hostServices`授权模型，但基线规范仍要求`Service`/`AdminService`分层，且部分要求把领域宿主实现集中归属到`internal/service/plugin/internal/capabilityhost`。这会让源码插件开发者同时理解普通目录和管理目录，也会让动态插件的真实授权边界与 Go 接口分层重复。

本次设计以`localdocs/plugin-domain-service-unification-design.md`中的讨论结论为输入：项目没有历史兼容负担，允许一次性破坏性收敛；插件管理复杂度不应因为新增方法可声明性字段而扩大；动态插件可声明性应由动态`host service registry`是否注册该方法表达。

## Goals / Non-Goals

**Goals:**

- 废除插件可见`AdminService`和`Services.Admin()`，每个领域只保留一个插件可见`Service`入口。
- 原管理方法并入对应领域`Service`，并由方法级元数据表达风险、授权资源、上下文、数据权限、性能和缓存影响。
- 让动态插件调用链统一为：已注册到动态`host service registry`、已在`plugin.yaml hostServices`声明、安装或启用时已授权、运行时`service + method + resource`校验通过。
- 保持源码插件可信源码扩展模型：源码插件直接使用类型化统一`Service`，不要求维护`hostServices`声明。
- 将领域实现归属真实领域 owner，`capabilityhost`和 WASM host service 只保留薄适配职责。
- 一次性标准化动态 wire method，不保留旧方法兼容别名。

**Non-Goals:**

- 不引入`dynamic-auth`、`source-only`、`reserved`或等价方法可声明性字段。
- 不把`Runtime`、`Network`、`Data.RecordStore`、`Secret`、`Event`、`Queue`并入源码插件`capability.Services`。
- 不通过`pluginhost.Services`顶层分散入口或`pluginbridge`平行业务接口暴露治理能力；治理能力必须归属到`Plugins()`、`Tenant()`等对应领域组件。
- 不把源码插件和动态插件的声明期生命周期 callback facade 当作运行期领域能力暴露。
- 不改变插件菜单权限、安装授权和领域数据权限三者分层，只调整领域方法入口形态。
- 不新增数据库表或迁移，不改变宿主 HTTP API 的资源语义。

## Decisions

### 统一领域`Service`，删除`AdminService`

所有插件可见领域能力统一为每个领域一个`Service`入口。读取、校验、执行、写入、管理动作都在该领域`Service`下表达，但必须具备方法级治理元数据。

选择该方案的原因是动态插件真实安全边界已经是`service + method + resource`授权快照，Go 接口名中的`AdminService`不能替代运行时授权、数据权限和状态机校验。源码插件作为可信源码扩展，也不应为了方法风险分类承担双目录学习成本。

备选方案是继续保留`AdminService`作为风险外观。该方案被放弃，因为它会继续保留`Services.Admin()`和动态授权模型之间的双重概念，并让插件治理复杂度上升。

### 动态可声明性只由 registry 注册事实表达

动态插件可声明的方法必须存在于动态`host service registry`。不存在的`service + method`在构建、安装、启用或运行时都必须被拒绝，且不得进入领域 owner。

不新增单独的方法可声明性字段。Go `Service`中未注册到动态 registry 的方法自然不能被动态插件声明；这已经足以表达源码插件专用或暂不开放的能力，避免在插件管理界面和授权模型里引入第二套可见性状态。

### 动态领域能力按暴露风险分层推进

对`Manifest`、`Storage`、`Users`、`Dict`、`BizCtx`、`HostConfig.SysConfig`、`Sessions`、`Files`、`Notifications`、`Cache`、`Lock`、`AI`和`Jobs`的对账结论是：动态 registry 已覆盖大多数只读、批量、可见性和基础执行投影，但示例动态插件没有完整展示这些已发布投影；同时部分统一 Go `Service`方法因写入、执行、内容流、配置修改或跨领域治理风险，尚不应直接镜像为动态 wire method。

后续补齐动态插件暴露时按三层推进：

| 层级 | 目标范围 | 处理口径 |
| --- | --- | --- |
| 低风险示例投影 | `BizCtx`、`Cache`、`Lock`、`Manifest`批量/列表、`Storage`批量/游标/元数据 | 只补齐`linapro-demo-dynamic`声明和 smoke 示例，不新增运行期方法。 |
| 中风险只读对齐 | `Users`、`Dict`、`Sessions`、`Files`、`Notifications`只读投影、`HostConfig.SysConfig`只读、`AI`状态 helper | 仅在领域 owner 已有批量、可见性、数据权限和数量上限闭环时开放或对齐 guest helper，并补 dispatcher、catalog、guest client 和示例测试。 |
| 高风险写入/执行 | 用户写入与角色关联、字典写入与刷新、会话撤销、文件内容读取/更新/删除、任务运行管理、系统配置写入、`Storage.ProviderStatuses` | 另行设计事务、审计、缓存失效、租户隔离、幂等、资源成本和错误暴露策略，不在示例补齐中顺带开放。 |

某个统一 Go `Service`方法没有动态投影不默认视为缺陷。评估时必须记录它属于已发布但示例未声明、只读可补齐、源码插件专用、高风险待设计或缺少稳定 owner 中的哪一类；该分类只存在于 OpenSpec、README 或任务记录中，不新增运行时可声明性字段。

### 风险等级是治理元数据，不是功能开关

`risk`用于授权展示、升级风险、测试和审查门禁。它不决定方法是否对动态插件可声明，也不替代`hostServices`授权快照、数据权限、租户边界、状态机和资源校验。

这样可以避免出现“风险等级低所以自动开放”或“风险等级高所以需要额外开关”的隐式规则。方法是否可动态声明只有 registry 注册事实一个来源。

### 领域实现归属真实 owner

`usercap`、`dictcap`、`filecap`、`tenantcap`、`orgcap`等领域能力应由对应领域 owner 或稳定 owner 适配实现。owner 内部应复用既有业务逻辑、数据权限、事务、缓存失效和错误处理路径，不为插件领域能力实现第二套业务逻辑。

`capabilityhost`、WASM host service 和 dispatcher 只承担标准业务上下文桥接、动态授权、请求响应编解码和错误映射。它们不得构造额外插件调用上下文传给领域 owner，也不得直接访问其他领域`DAO`、`DO`、`Entity`、私有缓存或内部 helper。

### 动态 wire method 破坏性标准化

动态 wire method 按领域和子资源一次性标准化，例如`users.batch_get`、`dict.type.create`、`plugins.registry.list`、`org.department.list_tree`。旧 wire method 不保留兼容别名。

项目没有历史兼容负担，保留别名会制造长期测试矩阵和治理扫描复杂度。实现阶段必须同步更新 protocol、guest client、dispatcher、catalog、示例插件和 README。

### Auth 领域动态命名保持全局一致

认证授权能力在源码插件目录中以`Services.Auth()`聚合入口发布，`Token()`和`Authz()`只是`authcap.Service`下的子能力。动态插件协议必须保持同样的顶层领域边界，使用`service: auth`声明认证和授权方法，不再把授权子能力拆成独立顶层`service: authz`。

动态授权粒度通过方法名和派生能力区分：token 方法使用`token.*`前缀并派生`host:auth:token`，授权方法使用`authz.*`前缀并派生`host:auth:authz`。这样安装授权和运行时授权仍然可以区分 token 与 authz 风险，但插件开发者看到的顶层领域目录与源码插件`Services.Auth().Token()`、`Services.Auth().Authz()`保持一致。

### 治理能力内聚到对应领域`Service`

`PluginLifecycle()`、`PluginState()`、`TenantPluginGovernance()`和`TenantFilter()`不再挂在`pluginhost.Services`顶层，因为这些名称会把领域治理能力伪装成源码插件目录能力。治理能力必须按领域职责内聚：插件生命周期归属`Plugins().Lifecycle()`，插件启用状态读取归属`Plugins().State()`；租户插件启停和默认供给归属`Tenant().Plugins()`，租户过滤上下文归属`Tenant().Filter()`。`pluginhost.Services`镜像普通`capability.Services`，不再定义独立`TenantService`或表过滤顶层入口；同进程源码插件和宿主 adapter 如需改写 GoFrame 查询，必须显式调用`tenantspi.ApplyPluginTableFilter(ctx, filter, model, qualifier)`。

动态插件可以使用已注册的治理领域方法，但必须和普通领域方法一样经过动态`host service registry`、`plugin.yaml hostServices`声明、安装或启用授权、运行时`service + method + resource`校验，以及领域 owner 内部的数据权限、租户边界、状态机、缓存一致性和数量上限治理。未注册的治理方法即使存在于 Go `Service`接口中，也不能被动态插件声明或调用。
### capability 读模型命名取向

公开 capability DTO 优先采用业务含义更直接的命名，避免在主资源契约里继续使用过多`Projection`后缀。对于单条展示、可见实体或可缓存快照，优先使用`Info`、`Detail`、`Summary`、`Status`或领域专有名词；对于标签解析结果优先使用`Label`；对于导出或页面专用、且仅在同一领域内部稳定消费的子资源，可以保留更明确的`Projection`命名。该决定的目标不是清除所有 projection 语义，而是让插件开发者在主契约层看到更自然、专业的领域词汇。

现阶段优先处理 `sessioncap`、`usercap`、`authcap`、`filecap`、`dictcap`、`notifycap`、`hostconfigcap`、`orgcap`、`tenantcap` 和 `aicap` 的公开读模型；如果某个领域的现有名称已经等于最自然的业务词汇，则保持不动，不为统一而制造二次歧义。

源码插件可以直接通过类型化`Services.Plugins()`和`Services.Tenant()`使用这些治理子能力。`pluginhost.Services`不再保留`PluginLifecycle()`、`PluginState()`、`TenantPluginGovernance()`、`TenantFilter()`或`TenantTableFilter()`顶层转发，也不定义与`tenantcap.Service`重复的`TenantService`。新增源码插件调用点必须注入对应领域的最窄子能力。需要改写插件自有表查询时，调用点应接收普通`tenantcap.FilterService`，并通过`tenantspi.ApplyPluginTableFilter(...)`把上下文应用到插件自有`*gdb.Model`。

`TenantFilter`中接受`*gdb.Model`的查询构造器辅助只适用于同进程源码插件和宿主 owner adapter，并且只能作为`tenantspi`包级 helper 存在，不得进入普通`tenantcap.FilterService`、`pluginhost.Services`或 WASM 协议。动态插件的等价租户隔离必须通过`RecordStore`、领域 host service 参数或领域 owner 内部过滤完成，传输协议只暴露可序列化的上下文、资源和过滤意图。

## Risks / Trade-offs

- **一次性破坏性收敛导致实现面较大** → 通过 OpenSpec 任务拆分为契约、实现归属、动态 registry、README 和测试扫描多个步骤，且每步保留静态检索门禁。
- **旧规范仍残留`AdminService`语义** → 增量规范使用`MODIFIED`/`REMOVED`覆盖冲突要求，归档时替换基线。
- **领域 owner 迁移可能诱发跨模块依赖混乱** → 所有新增运行期依赖必须通过构造函数显式注入，任务记录包含 owner、创建位置、传递路径和共享实例判断。
- **动态插件方法开放边界被误解为 risk 控制** → 规范明确动态可声明性只由 registry 注册事实表达，`risk`仅用于治理。
- **高频读取方法出现`N+1`或前端瀑布式调用** → 每个插件可见读取、列表、树形、批量和聚合方法必须记录分页、数量上限、投影和批量装配策略。
- **缓存敏感写入遗漏跨实例失效** → 权限、插件状态、运行时配置、字典、组织、租户等关键数据写入必须记录权威数据源、事务后失效和跨实例同步机制。

## Migration Plan

1. 更新 OpenSpec 基线要求，删除`AdminService`、`Services.Admin()`和插件生命周期暴露口径。
2. 重构`pkg/plugin/capability`接口：删除`AdminServices`，把原管理方法并入对应领域`Service`，补齐方法注释和治理元数据。
3. 调整`pluginhost.Services`：删除`Admin()`，源码插件调用点改为注入最窄统一`*cap.Service`。
4. 将领域能力实现迁回真实 owner 或 owner adapter；`capabilityhost`只保留动态调用薄适配。
5. 更新动态`host service registry`、dispatcher、guest client 和 protocol wire method，拒绝未注册、未声明、未授权或资源不匹配的方法。
6. 删除`pluginhost.Services.PluginLifecycle()`和`PluginState()`顶层入口，源码插件调用点改用`Plugins().Lifecycle()`和`Plugins().State()`；需要动态开放的治理方法由领域 owner 注册到动态`host service registry`。
7. 同步`apps/lina-core/pkg/plugin`中英文 README、示例插件、治理扫描和测试。
8. 运行 OpenSpec strict、Go 编译门禁、动态 host service 单测、启动装配测试、静态检索和 README 同步检查。

## Open Questions

无。讨论结论已经确认：废除`AdminService`，不增加方法可声明性字段，动态 wire method 不保留兼容别名，插件治理生命周期不向插件暴露，领域实现归属真实 owner。
