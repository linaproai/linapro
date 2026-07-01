## 1. 契约与规范同步

- [x] 1.1 读取并记录本次实现命中的`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`api-contract.md`、`data-permission.md`、`cache-consistency.md`、`i18n.md`和`testing.md`规则。
- [x] 1.2 更新`pkg/plugin/capability`根服务目录，删除`AdminServices`和各领域插件可见`AdminService`入口。
- [x] 1.3 建立领域迁移对账表，逐项覆盖`Users`、`Auth.Token`、`Authz`、`AI`、`APIDoc`、`BizCtx`、`Cache`、`Dict`、`Files`、`HostConfig`、`I18n`、`Infra`、`Jobs`、`Lock`、`Manifest`、`Notifications`、`Plugins`、`Org`、`Route`、`Sessions`、`Storage`和`Tenant`。
- [x] 1.4 将原`AdminService`方法并入对应领域统一`Service`接口，并补齐方法注释中的风险、授权资源、上下文、数据权限、性能和缓存影响说明。
- [x] 1.5 删除`pluginhost.Services.Admin()`和源码插件管理目录相关测试替身。
- [x] 1.6 静态检索确认生产代码、官方插件、测试替身和文档不再引用旧`AdminServices`、领域`AdminService`或`Services.Admin()`。

## 2. 领域实现归属

- [x] 2.1 梳理现有`capabilityhost`中承载的领域业务实现，按真实领域 owner 迁移到对应`internal/service/<domain>`或 owner 发布的稳定适配契约。
- [x] 2.2 调整启动装配，通过构造函数显式注入领域 owner 契约，记录新增运行期依赖的 owner、创建位置、传递路径和共享实例策略。
- [x] 2.3 保持`capabilityhost`和 WASM host service 仅承担标准业务上下文桥接、授权、编解码和错误映射。
- [x] 2.4 静态检索确认`capabilityhost`和 WASM host service 生产代码不直接导入其他领域`internal/dao`、`internal/model`或私有实现包。

## 3. 身份、认证与权限领域

- [x] 3.1 迁移`Users`领域统一`Service`，覆盖`Current`、`Get`、`BatchGet`、`BatchResolve`、`List`、`EnsureVisible`、`Create`、`Update`、`Delete`、`SetStatus`、`ResetPassword`和`Assignment().ReplaceRoles`，并验证用户数据权限、租户边界和批量投影。
- [x] 3.2 迁移`Auth.Token`领域统一`Service`，覆盖`SelectTenant`、`SwitchTenant`、`IssueImpersonationToken`和`RevokeImpersonationToken`，并验证 token/session 状态、系统调用和租户边界。
- [x] 3.3 迁移`Authz`领域统一`Service`，覆盖`BatchGetPermissions`、`HasPermission`、`BatchHasPermissions`、`IsPlatformAdmin`、`EnsurePermissionsVisible`和`ReplaceRolePermissions`，并验证权限快照复用和事务后缓存修订号。
- [x] 3.4 迁移`Sessions`领域统一`Service`，覆盖`Current`、`Get`、`BatchGet`、`List`、`BatchGetUserOnlineStatus`、`EnsureVisible`、`Revoke`和`RevokeMany`，并验证不存在与不可见不泄露。

## 4. 配置、上下文与框架支撑领域

- [x] 4.1 迁移`AI`领域统一`Service`，覆盖状态读取和`Text`、`Image`、`Embedding`、`Audio`、`Vision`、`Document`、`Safety`、`Video`子能力，确保不暴露 provider 密钥、模型路由和原始渠道响应。
- [x] 4.2 迁移`APIDoc`领域统一`Service`，覆盖`ResolveRouteText`、`ResolveRouteTexts`和`FindRouteTitleOperationKeys`，并验证批量解析不产生重复资源加载。
- [x] 4.3 迁移`BizCtx`领域统一`Service`，覆盖`Current`，并验证缺少请求上下文时的零值或结构化错误语义。
- [x] 4.4 迁移`HostConfig`领域统一`Service`，覆盖静态配置读取、`SysConfig()`读取、`SysConfig()`写入和`EnsureVisible`，并验证 key 授权、缓存失效和跨实例同步。
- [x] 4.5 迁移`I18n`领域统一`Service`，覆盖`GetLocale`、`Translate`和`FindMessageKeys`，并确认动态插件不恢复独立`i18n`host service。
- [x] 4.6 迁移`Infra`领域统一`Service`，覆盖`BatchGetStatus`、`ListComponents`、`EnsureComponentsVisible`、`RefreshStatus`和`RefreshAllStatus`，并验证敏感基础设施信息脱敏和刷新限流。

## 5. 资源、内容与任务领域

- [x] 5.1 迁移`Cache`领域统一`Service`，覆盖`Get`、`GetMany`、`Set`、`SetMany`、`Incr`、`Expire`、`Delete`和`DeleteMany`，并验证插件 ID、租户作用域和共享缓存后端复用。
- [x] 5.2 迁移`Dict`领域统一`Service`，覆盖`Type()`、`Value()`和`Refresh`方法，验证字典值数据权限、`labelKey`、当前 locale 标签和字典缓存失效。
- [x] 5.3 迁移`Files`领域统一`Service`，覆盖文件投影读取、内容读取、`EnsureVisible`、`Upload`、`CreateFromStorage`、`UpdateMetadata`、`Delete`和`DeleteMany`，并验证宿主文件中心与插件私有`Storage`边界。
- [x] 5.4 迁移`Jobs`领域统一`Service`，覆盖`Get`、`BatchGet`、`List`、`EnsureVisible`、`Create`、`Update`、`Delete`、`Run`和`SetStatus`，并验证`jobs.register`仅属于声明期 facade，不进入运行期`jobcap.Service`。
- [x] 5.5 迁移`Lock`领域统一`Service`，覆盖`Acquire`、`Renew`和`Release`，并验证锁名作用域、ticket 不透明和共享锁后端复用。
- [x] 5.6 迁移`Manifest`领域统一`Service`，覆盖`Get`、`GetMany`、`List`、`Exists`和`Scan`，并验证当前插件作用域、路径、单资源大小和总字节数限制。
- [x] 5.7 迁移`Notifications`领域统一`Service`，覆盖`Get`、`BatchGet`、`List`、`BatchGetBySource`、`EnsureVisible`、`Send`、`Delete`、`DeleteBySource`、`MarkRead`和`MarkUnread`，并验证类型化投影、channel 授权和租户边界。
- [x] 5.8 迁移`Storage`领域统一`Service`，覆盖`Put`、`Get`、`Stat`、`BatchStat`、`List`、`ListCursor`、`ProviderStatuses`、`Delete`和`DeleteMany`，并验证插件私有对象边界和禁止无边界`DeletePrefix`。

## 6. 插件、组织、租户与路由领域

- [x] 6.1 迁移`Plugins`领域统一`Service`，只保留`Registry()`、`Capability()`和`Config()`子能力，删除`Lifecycle()`和租户生命周期暴露。
- [x] 6.2 迁移`Org`领域统一`Service`，覆盖`Department()`、`Post()`和`Assignment()`子资源，验证 provider 缺失安全降级、组织数据权限和缓存失效。
- [x] 6.3 迁移`Route`领域统一`Service`，覆盖`GetMetadata`/`Metadata`，并验证只返回当前动态路由投影且不暴露宿主 router 内部对象。

- [x] 6.4 迁移`Tenant`领域统一`Service`，覆盖`Context()`、`Directory()`和`Membership()`子能力，验证租户 provider 缺失安全降级、数据权限和成员关系写入边界。

## 7. 动态专属服务

- [x] 7.1 更新动态`host service registry`，把动态可声明性收敛为 registry 注册事实，不新增额外方法可声明性字段或等价功能开关。
- [x] 7.2 标准化动态 wire method 名称，更新 protocol、guest client、dispatcher、catalog 和示例插件，不保留旧 method 兼容别名。
- [x] 7.3 调整动态插件清单校验，未注册、未声明、未授权或资源不匹配的`service + method`必须在构建、安装、启用或运行时拒绝。
- [x] 7.4 迁移`Runtime`动态专属服务，覆盖日志、状态读写、时间、UUID 和节点信息方法，并验证插件作用域、授权和数据隔离。
- [x] 7.5 迁移`Network`动态专属服务，覆盖`Request`，并验证 URL 白名单、方法、header、body 大小、超时和错误映射策略。
- [x] 7.6 迁移`Data.RecordStore`动态专属服务，覆盖`Get`、`BatchGet`、`List`、`Count`、`Create`、`Update`、`Delete`和`DeleteMany`，并验证表授权、租户隔离、分页和批量写入整体拒绝语义。
- [x] 7.7 明确`Secret`、`Event`和`Queue`仅为预留服务，本次不进入`capability.Services`，不注册动态 dispatcher。
- [x] 7.8 从动态 registry、guest SDK 和 dispatcher 中移除`Plugins.Lifecycle()`、租户生命周期钩子和插件治理生命周期动作。
- [x] 7.9 补充动态 host service 单元测试，覆盖已注册已授权、未注册、未声明、未授权和资源越界场景。

## 8. 数据权限与缓存

- [x] 8.1 为所有读取、列表、树形、批量和聚合类插件可见方法确认分页、数量上限、投影和批量装配策略，避免`N+1`查询和前端瀑布式调用。
- [x] 8.2 为所有写入、删除、状态变更和执行类插件可见方法确认目标可见性校验、批量整体拒绝或明确逐项语义。
- [x] 8.3 为权限、插件状态、运行时配置、字典、组织和租户等缓存敏感写入方法记录权威数据源、事务后失效、跨实例同步和故障降级策略。
- [x] 8.4 确认调用端可见错误使用稳定业务错误码、`messageKey`、参数和英文 fallback。

## 9. 文档与示例

- [x] 9.1 更新`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`，删除`AdminService`、`Services.Admin()`和插件生命周期暴露描述。
- [x] 9.2 更新源码插件和动态插件示例，源码插件使用统一`Service`，动态插件使用标准化 wire method。
- [x] 9.3 检查插件本地 README、OpenSpec 任务记录和设计文档中的`i18n`、缓存一致性、数据权限、开发工具跨平台和测试策略影响判断。

## 10. 验证

- [x] 10.1 运行`openspec validate standardize-plugin-domain-services --strict`。
- [x] 10.2 运行覆盖`pkg/plugin/capability`、`pkg/plugin/pluginhost`、`pkg/plugin/pluginbridge`和动态 host service 变更包的 Go 编译门禁。
- [x] 10.3 运行宿主启动装配测试，至少覆盖`cd apps/lina-core && go test ./internal/cmd -count=1`或等价更窄测试。
- [x] 10.4 运行新增或更新的动态 host service、领域 owner 和包边界治理测试。
- [x] 10.5 运行静态检索，确认旧`AdminService`入口、旧动态 wire method 兼容别名和插件生命周期动态 service 不再存在。
- [x] 10.6 完成实现后调用`lina-review`进行代码和规范审查。

## Feedback

- [x] **FB-1**: 补齐`pkg/plugin/capability`统一`Service`方法面，对齐`localdocs/plugin-domain-service-unification-design.md`方法矩阵。
- [x] **FB-33**: 将公开 capability 读模型中的`Projection`命名收敛为更自然的领域命名（如`Info`、`Detail`、`Label`或`Status`），减少插件开发者在主资源契约中接触生硬的技术词缀。
- [x] **FB-34**: 删除`pluginhost.Services.PluginLifecycle()`和`PluginState()`顶层入口，源码插件与测试替身统一改用`Services.Plugins().Lifecycle()`和`Services.Plugins().State()`。
- [x] **FB-36**: 删除`pluginhost.Services.TenantPluginGovernance()`和`TenantFilter()`顶层入口，源码插件与测试替身统一改用`Services.Tenant().Plugins()`和`Services.Tenant().Filter()`。
- [x] **FB-37**: 将`usercap.Service`中的用户与角色关联关系方法拆分到`Assignment()`子领域，保留用户基础方法不变，并同步更新宿主适配、动态 guest 适配和测试替身。
- [x] **FB-38**: 基于动态`host service`注册事实重新对账`Manifest`、`Storage`、`Users`、`Dict`、`BizCtx`、`HostConfig.SysConfig`、`Sessions`、`Files`、`Notifications`、`Cache`、`Lock`、`AI`和`Jobs`领域能力，分层补齐动态插件示例投影、只读动态方法与高风险写入方法设计。
- [x] **FB-39**: 将`orgspi`和`tenantspi`对主框架暴露的`Service`接口统一收口为同名公开契约，并同步`httpstartup`、控制器和测试替身引用。
- [x] **FB-40**: 收敛`orgspi.Service`方法面，移除内部 flat read 和 assignment 聚合，改由`Scope()`、`Workspace()`和普通`orgcap.Service`子资源提供窄接口。
- [x] **FB-41**: 收敛`tenantspi.Service`方法面，删除总线型公开接口并改用`tenantcap.Service`子资源和宿主窄接口注入。
- [x] **FB-42**: 恢复窄版`tenantspi.Service`公开接口，作为主框架租户运行期契约并避免`New`返回内部结构体。
- [x] **FB-43**: 进一步收敛`tenantcap`和`tenantspi`租户能力边界，移除插件公共面的租户生命周期写入、成员替换、动态无效`Apply`和未使用 resolver/batch SPI。
- [x] **FB-44**: 继续简化`tenantspi`设计，删除总线型聚合接口并通过`Services.TenantTableFilter()`显式注入源码插件表过滤能力。
- [x] **FB-45**: 恢复`tenantspi.New`返回公开窄接口，避免主框架装配层暴露内部`serviceImpl`类型。
- [x] **FB-46**: 合并`tenantspi`单方法宿主治理接口，使用`HostGovernanceService`收敛启动校验、租户可见性、平台绕过和租户插件供给能力。
- [x] **FB-47**: 合并整理`tenantspi`非主文件中的过细实现文件，保留主文件为契约入口并收拢同职责实现。
- [x] **FB-48**: 合并插件宿主装配中的重复单方法接口，删除相邻包同签名的`GetRaw`、`RevokeSession`和`Current`接口重复定义。
- [x] **FB-49**: 取消源码插件`Services.TenantTableFilter()`顶层入口，将源码插件同进程租户表过滤内聚到`Services.Tenant().Filter()`。
- [x] **FB-50**: 删除`pluginhost.TenantService`重复接口，保留`tenantcap.Service`作为唯一租户服务契约，并将同进程 GoFrame 表过滤收敛为`tenantspi.ApplyPluginTableFilter(...)`helper。
- [x] **FB-51**: 合并插件服务根包薄文件，删除重构残留空壳文件，减少根包文件数量。
- [x] **FB-52**: 修正插件服务根包与子包单元测试文件命名，恢复测试文件与源码文件的同名关联。
- [x] **FB-53**: 修正业务事务闭包中直接使用`tx.Model`和传递`gdb.TX`的问题，统一通过`dao.Ctx(ctx)`继承事务上下文。
- [x] **FB-54**: 修正插件领域写入绕过统一`cachecoord`修订号发布的问题，并合并`cachecoord`过细小文件。
- [x] **FB-56**: 将动态插件授权子能力从顶层`authz` service 收敛到`auth`领域方法族，保持源码插件和动态插件领域目录一致。
- [x] **FB-57**: 合并动态插件侧`authz`实现文件到`auth`领域能力相关源文件，避免文件组织继续暗示存在顶层`authz`动态领域。
- [x] **FB-2**: 标准化动态`host service`wire method、catalog、dispatcher和 guest client 命名，对齐领域子资源方法。
- [x] **FB-3**: 补齐领域 owner adapter、测试替身和源码插件调用点，确保新增方法复用真实 owner 或安全降级，不回流`capabilityhost`业务实现。
- [x] **FB-4**: 更新 README、OpenSpec 记录和验证证据，运行编译、静态检索、`openspec validate`和`lina-review`。
- [x] **FB-5**: 修复预留动态`Secret`、`Event`和`Queue`host service 仍可被`plugin.yaml hostServices`声明的问题，确保预留项仅保留为 catalog/README 占位，不进入运行时可声明方法、capability 派生或 capability 白名单。
- [x] **FB-6**: 调整`pkg/plugin/capability`主契约文件组织顺序，将组件关键接口定义置于主文件开头，支撑 DTO、输入输出和投影类型放在其后。
- [x] **FB-7**: 将插件领域能力的显式能力上下文参数收敛为`ctx`承载的调用上下文。
- [x] **FB-8**: 删除额外能力调用上下文模型，让领域能力仅依赖标准业务`ctx`和动态 dispatcher 授权结果。
- [x] **FB-9**: 将`HostConfig`中的`sys_config`参数能力收敛到`SysConfig()`子领域，并使用原始`sys_config.value`值契约。
- [x] **FB-10**: 将`Files`内容读取契约收敛为单一`Open`方法，删除与其语义和实现重复的`Download`方法。
- [x] **FB-11**: 为`HostConfig().SysConfig()`补充单 key `Get`方法，并复用`BatchGet`保持一致权限与缺失语义。
- [x] **FB-12**: 补齐`Files.Upload`和`Files.CreateFromStorage`写入能力，让插件可上传并记录到`sys_file`，同时保持宿主文件中心与插件私有`Storage`边界。
- [x] **FB-13**: 修复`Jobs`领域运行期能力实现，接入`jobmgmt` owner 并将`Register`收敛到声明期 facade。
- [x] **FB-14**: 将`Jobs`领域能力接口和内部状态参数改为复用`api/job/v1.Status`，避免在`internal/service`或`jobcap`中重复定义 Job status 枚举值。
- [x] **FB-15**: 将`Manifest`能力的宿主 factory 和读取实现迁移到插件服务内部组件，保持`manifestcap`仅暴露插件可见契约。
- [x] **FB-16**: 将`HostConfig().SysConfig()`的宿主适配实现迁移到插件服务内部组件，避免以开放独立`runtimeconfig` service 组件暴露插件能力细节。
- [x] **FB-17**: 回收`capabilitydomain`相关共享原语到既有能力契约包，删除独立耦合包，保留文件投影与缓存修订等局部耦合逻辑在原领域实现和事务边界内。
- [x] **FB-18**: 删除未实现、未引用的领域`ScopeService`空壳接口，保留`orgspi`和`tenantspi`真实宿主内部 SPI。
- [x] **FB-19**: 补齐`Notifications`动态插件接口实现，发布并接通`List`、`Delete`、`DeleteBySource`、`MarkRead`和`MarkUnread`动态 host-service 方法。
- [x] **FB-20**: 将`Org`和`Tenant`可见性校验收敛为单一`EnsureVisible(ctx, ids []...)`批量签名，删除并列的`EnsureVisibleMany`公共入口并同步测试替身。
- [x] **FB-21**: 将`plugincap.ConfigService`契约归位到`plugincap.go`，并按配置、生命周期和状态职责整合`plugincap`包文件命名与实现分布。
- [x] **FB-22**: 修正`plugincap`包聚合后源码文件命名，恢复符合项目风格的`plugincap_*.go`前缀。
- [x] **FB-23**: 为`plugincap.ConfigService.Get`补齐`defaultValue`参数，并同步实现、动态适配器和测试替身。
- [x] **FB-24**: 将插件配置宿主 factory 和读取实现迁移到插件服务内部组件，保持`plugincap`仅暴露插件可见`ConfigService`契约。
- [x] **FB-25**: 将`ScopedServicesFactory`和`ServicesForPlugin`迁移到插件服务内部`scoping`组件，并删除`pkg/plugin/capability`公开绑定辅助。

### FB-25 实施记录

- 将 `ScopedServicesFactory` 和 `ServicesForPlugin` 从 `pkg/plugin/capability` 迁移到 `internal/service/plugin/internal/scoping`，让插件作用域绑定只存在于宿主内部边界。
- 生产调用点改为 `plugin_integration.go`、`internal/wasm/wasm_host_service.go` 和 `internal/runtime/runtime_reconciler_uninstall.go` 引用新 `scoping` 组件。
- 测试替身与断言改为依赖新 `scoping` 组件，公开 `capability` 包不再暴露插件作用域绑定辅助。
- `i18n`、缓存一致性、数据权限和开发工具跨平台影响：本次仅搬迁宿主内部绑定辅助，无运行时文案、缓存后端、数据权限逻辑或脚本变更，确认无影响。
- DI 来源检查：未新增运行期依赖；`scoping` 仅消费既有的 `capability.Services`，仍由启动期共享服务图传入。

### FB-38 探索记录

- 根因：动态插件能力可声明性以`host service`registry/catalog发布事实为准，不等同于公开`capability.Service`方法完整面；当前`linapro-demo-dynamic`只声明`runtime`、`storage`部分、`network`、`data`、`plugins`部分、`jobs.register`、`manifest.get`、`hostConfig.get`、`org`和`tenant`，未展示`Users`、`Dict`、`BizCtx`、`Sessions`、`Files`、`Notifications`、`Cache`、`Lock`、`AI`等已发布或可评估能力，造成“框架有投影但示例不可见”的偏差。
- 已有动态投影：`Manifest`已发布`get`、`get_many`和`list`；`Storage`已发布对象读写、分片上传、删除、批量删除、列表、游标列表、元数据和批量元数据；`Users`已发布当前用户、批量、解析、列表和可见性；`Dict`已发布值标签解析、值列表和可见性；`BizCtx`已发布当前上下文；`Sessions`已发布当前会话、列表、批量、在线用户批量和可见性；`Files`已发布批量、列表、可见性、上传和从`Storage`创建；`Notifications`已发布批量、列表、来源查询、可见性、发送、删除和已读状态；`Cache`、`Lock`已基本完整发布；`AI`已发布多模态调用与根状态批量查询；`Jobs`已发布声明期`jobs.register`和运行期读投影。
- 当前缺口：`Storage.ProviderStatuses`未动态发布；`HostConfig.SysConfig()`未动态发布，只有静态`hostconfig.get`；`Users`写入、角色关联、密码和状态类方法未发布；`Dict.Type`、字典写入和刷新未发布；`Sessions.Revoke/RevokeMany`未发布；`Files.Detail/ListScenes/ListSuffixes/Open/UpdateMetadata/Delete/DeleteMany`未发布；`Jobs.Create/Update/Delete/Run/SetStatus`未发布；`AI`非`text`子能力`MethodStatus`guest helper未委托根`ai.methods.status.batch_get`。
- 缺口原因：只读投影大多已经通过领域 owner、批量/可见性校验和动态授权闭环完成；未发布项集中在写入、执行、内容流、配置修改和跨领域治理，涉及事务、数据权限、审计、缓存失效、租户隔离、幂等与资源成本控制，不能简单把`Service`方法镜像为动态 wire method。
- 推进策略：第一层优先补齐`linapro-demo-dynamic`的低风险 smoke 投影声明与示例接口，覆盖`BizCtx`、`Cache`、`Lock`、`Manifest`批量/列表和`Storage`批量/游标/元数据；第二层评估只读扩展与 guest helper 对齐，覆盖`Users`、`Dict`、`Sessions`、`Files`、`Notifications`、`HostConfig.SysConfig`只读和`AI`状态 helper；第三层把用户写入、字典写入、会话撤销、文件内容/删除、任务运行管理、系统配置写入和`Storage.ProviderStatuses`作为高风险能力另行设计，不在探索记录中直接实现。
- 本轮边界：当前仍处于`openspec-explore`，只沉淀评估结论与后续任务，不修改动态`host service`运行时代码、插件`plugin.yaml`或示例实现。
- OpenSpec 补充：已将动态领域能力按 registry、示例和风险分层治理的决策写入`design.md`和`plugin-host-service-extension`增量规范；`FB-38`仍保持未完成，等待退出探索后进入实现。
- 后续实施拆分：
  1. 低风险示例投影：检查`apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`是否存在；在`plugin.yaml`声明已注册但示例未覆盖的`bizctx.current.get`、`cache.get`、`cache.get_many`、`cache.set`、`cache.set_many`、`cache.delete`、`cache.delete_many`、`cache.incr`、`cache.expire`、`lock.acquire`、`lock.renew`、`lock.release`、`manifest.get_many`、`manifest.list`、`storage.delete.batch`、`storage.list.cursor`和`storage.stat.batch`；只复用现有 registry/catalog，不新增动态方法。
  2. 低风险示例接口：扩展`linapro-demo-dynamic`既有 host-call demo payload、service fake 和单测，展示`BizCtx`当前上下文、`Cache`读写/过期/批量、`Lock`获取续期释放、`Manifest`批量/列表和`Storage`批量删除/游标列表/批量元数据；保持接口为 smoke 展示，不引入新的业务工作流。
  3. 文档与声明一致性：同步更新`linapro-demo-dynamic`中英文 README、插件清单注释和`apps/lina-core/pkg/plugin`中英文 README 生成表；静态检索确认示例声明的方法均存在于 catalog，且未声明高风险待设计方法。
  4. 中风险只读对齐：单独评估`Users`、`Dict`、`Sessions`、`Files`、`Notifications`、`HostConfig.SysConfig`只读和`AI`状态 helper。只有在领域 owner、批量边界、数据权限、错误收敛和 guest helper 路径完整时，才进入动态 registry、dispatcher、catalog、guest client 和测试补齐。
  5. 高风险另行设计：用户写入与角色关联、字典写入与刷新、会话撤销、文件内容读取/更新/删除、任务运行管理、系统配置写入和`Storage.ProviderStatuses`不得混入 demo smoke 任务；需要先形成事务、审计、缓存失效、租户隔离、幂等、资源成本和错误暴露设计。
  6. 实现阶段验证门禁：运行`openspec validate standardize-plugin-domain-services --strict`、相关 Go 包测试、动态插件 demo 后端测试、插件清单/README 静态检索、`git diff --check`和必要的`linapro-demo-dynamic`构建或 smoke 验证；完成后再将`FB-38`标记为完成并调用`lina-review`。
- 影响判断：仅更新 OpenSpec 设计、增量规范和任务记录；无运行时`i18n`、缓存一致性、数据权限、开发工具跨平台、HTTP API、SQL、前端或测试 fixture 变更；后续进入实现时需要分别补充验证门禁和数据权限/缓存/审计影响分析。
- 验证策略：本轮仅运行`openspec validate standardize-plugin-domain-services --strict`和`git diff --check -- openspec/changes/standardize-plugin-domain-services/design.md openspec/changes/standardize-plugin-domain-services/specs/plugin-host-service-extension/spec.md openspec/changes/standardize-plugin-domain-services/tasks.md`；后续若补齐示例或动态方法，需要追加 Go 编译、动态 dispatcher/guest client 测试和 demo 插件 smoke 测试。

### FB-38 实施记录

- 将动态插件示例投影分层收敛为低风险 smoke 面：`bizctx.current.get`、`cache.get/get_many/set/set_many/delete/delete_many/incr/expire`、`lock.acquire/renew/release`、`manifest.get_many/list`、`storage.delete.batch/list.cursor/stat.batch`。
- 只补充示例插件的动态声明、payload、controller、服务 fake、README 和插件内文档翻译，不把`Users`、`Dict`、`Sessions`、`Files`、`Notifications`、`HostConfig.SysConfig`、`AI`和`Jobs`的高风险写入或执行面直接镜像到动态方法。
- 高风险能力保持为后续单独设计项，原因是它们会引入事务、权限、审计、缓存失效、租户边界和幂等语义的额外复杂度，不适合和 smoke 投影混在同一轮示例扩展中。
- `i18n`、缓存一致性、数据权限和开发工具跨平台影响：本次只更新 OpenSpec 任务记录与既有动态示例的能力对账结论，无运行时文案、缓存策略、数据权限逻辑或工具脚本变化，确认无影响。
- DI 来源检查：未新增运行期依赖；动态示例仍复用现有 `capability.Services`、guest invoker 和 demo 插件内部 fake service，不引入新的启动期共享服务图。

### FB-38 验证

- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check -- openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/internal/service/dynamic ./backend/internal/controller/dynamic ./backend/api/... -count=1`通过。
- `rg -n "bizctx\\.current\\.get|cache\\.(get|get_many|set|set_many|delete|delete_many|incr|expire)|lock\\.(acquire|renew|release)|manifest\\.(get_many|list)|storage\\.(delete\\.batch|list\\.cursor|stat\\.batch)" apps/lina-plugins/linapro-demo-dynamic -g '*.go' -g '*.yaml' -g '*.md'`命中对应插件声明、示例实现和文档，确认低风险动态投影已对齐。
- `rg -n "Users|Dict|Sessions|Files|Notifications|HostConfig\\.SysConfig|AI|Jobs" apps/lina-plugins/linapro-demo-dynamic/backend apps/lina-plugins/linapro-demo-dynamic/plugin.yaml -g '*.go' -g '*.yaml'`未发现高风险动态方法声明，确认其仍处于后续单独设计边界内。

### FB-39 实施记录

- 根因：`orgspi`和`tenantspi`对主框架暴露的可消费接口仍存在`RuntimeService`命名，而`httpstartup`中的组织能力字段和用户控制器构造仍引用窄投影式类型名，导致主框架注入面未与统一`Service`契约完全对齐。
- 将`orgspi.RuntimeService`改为`orgspi.Service`，并保留`WorkspaceViewService`作为`Service`内部的组织工作台窄 helper；`httpRuntime`、用户控制器和路由装配统一直接注入`orgspi.Service`。
- 将`tenantspi.RuntimeService`改为`tenantspi.Service`，并同步更新宿主启动装配、插件宿主服务适配器、会话适配器、测试 helper 和断言。
- `i18n`影响：未新增运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源，仅涉及 Go 公开契约命名和注入面收口，确认无影响。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径或共享修订号策略，仅是接口命名和类型收口，确认无影响。
- 数据权限影响：未新增数据读取、写入、存在性探测或租户/组织边界暴露，原有可见性和校验语义保持不变，确认无影响。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本或`linactl`入口，确认无影响。
- DI 来源检查：未新增运行期依赖；`orgspi.Service`和`tenantspi.Service`仍由启动期共享服务图创建并传递，owner、创建位置和传递路径仅做命名收敛，不引入新实例图。

### FB-39 验证

- `openspec validate standardize-plugin-domain-services --strict`通过。
- `cd apps/lina-core && go test ./pkg/plugin/capability/orgcap ./pkg/plugin/capability/orgcap/orgspi ./pkg/plugin/capability/tenantcap ./pkg/plugin/capability/tenantcap/tenantspi ./internal/controller/user ./internal/cmd/internal/httpstartup ./internal/cmd -count=1`通过。
- `rg -n "RuntimeService\\b|WorkspaceViewService|orgProjection\\b" apps/lina-core -g '*.go'`仅剩`WorkspaceViewService`作为`orgspi.Service`内部组成部分，确认`RuntimeService`与主框架窄注入引用已清理。
- `rg -n "tenantspi\\.RuntimeService|orgspi\\.RuntimeService" apps/lina-core -g '*.go'`无匹配，确认公开 SPI 旧命名已移除。

### FB-40 实施记录

- 根因：`orgspi.Service`在`FB-39`后仍把普通`orgcap.Service`、内部 flat read、assignment 写入、数据范围 query builder 和工作台投影聚合在同一个接口中，导致调用方只需要部门名称、数据范围或用户管理树时仍必须依赖过宽方法面。
- 将`orgspi.Service`收敛为`orgcap.Service`加`Scope() ScopeService`和`Workspace() WorkspaceViewService`，删除`ConsumerReadService`和`orgspi.AssignmentService`接口；普通组织读取和写入统一走`orgcap.Service.Assignment()`、`Department()`和`Post()`子资源。
- `datascope`、用户服务、认证服务、用户控制器、`httpstartup`和插件集成构造改为注入实际需要的窄接口：数据权限使用`orgCapSvc.Scope()`，用户工作台控制器使用`orgCapSvc.Workspace()`，插件资源数据范围使用`orgcap.AssignmentService`，用户组织写入复用`orgCapSvc.Assignment()`。
- 保留`serviceImpl`上的 provider 转发方法作为包内实现 helper，不再通过`orgspi.Service`暴露给消费者；fallback 测试改为覆盖`Assignment()`、`Scope()`、`Workspace()`和`Department()`子资源，防止 flat 方法重新进入服务接口。
- `apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`已审查：当前文档描述的是插件公开`Org`普通能力和 provider SPI 边界，未描述宿主内部`orgspi.Service` flat 方法面；本次不改变公开`capability.Services.Org()`、动态 host service 或源码插件 provider 声明，因此无需同步 README。
- 规则读取与技能：本轮已读取`lina-feedback`、`goframe-v2`、`AGENTS.md`、`.agents/rules/openspec.md`、`architecture.md`、`backend-go.md`、`plugin.md`、`documentation.md`、`data-permission.md`和`testing.md`；未修改 HTTP API、SQL、前端、缓存实现、运行时文案、语言包或开发工具脚本，确认`api-contract`、`database`、`frontend-ui`、`cache-consistency`、`i18n`和`dev-tooling`规则域无影响。
- `i18n`影响：未新增或修改运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源，确认无运行时`i18n`影响。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径、订阅状态或共享修订号策略；组织 provider 可用性和 fallback 语义保持不变，确认无缓存一致性影响。
- 数据权限影响：数据范围仍由`datascope`在数据库查询阶段注入过滤，只是依赖来源从完整`orgspi.Service`改为`ScopeService`；插件资源部门范围仍通过`orgcap.AssignmentService.GetUserDeptIDs`读取，未新增数据存在性探测面。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本、Shell/Node 工具或`linactl`入口，确认无跨平台工具影响。
- DI 来源检查：未新增运行期依赖；`orgspi.New`仍在启动期共享服务图创建，消费者只接收更窄的同源接口投影，未引入新的独立实例图或缓存后端。
- 测试策略：本次属于后端 Go 插件宿主内部契约和测试替身收敛，不涉及用户可观察前端行为，未触发新增 E2E；通过 Go 包测试、OpenSpec 校验、静态检索和`git diff --check`完成验证。

### FB-40 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/orgcap/orgspi ./internal/service/user ./internal/service/auth ./internal/cmd/internal/httpstartup ./internal/cmd ./internal/service/plugin/internal/integration ./internal/service/plugin ./internal/service/datascope ./internal/service/cron ./internal/service/file ./internal/controller/i18n ./internal/service/jobmgmt -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check -- apps/lina-core/pkg/plugin/capability/orgcap/orgspi/orgspi.go apps/lina-core/pkg/plugin/capability/orgcap/orgspi/orgspi_impl.go apps/lina-core/pkg/plugin/capability/orgcap/orgspi/orgspi_fallback_test.go apps/lina-core/internal/service/auth/auth.go apps/lina-core/internal/service/auth/auth_impl.go apps/lina-core/internal/service/user/user.go apps/lina-core/internal/service/user/user_impl.go apps/lina-core/internal/service/user/user_test_dependencies_test.go apps/lina-core/internal/controller/user/user_new.go apps/lina-core/internal/cmd/internal/httpstartup/httpstartup_runtime.go apps/lina-core/internal/cmd/internal/httpstartup/httpstartup_routes_test.go apps/lina-core/internal/service/plugin/internal/integration/integration.go apps/lina-core/internal/service/plugin/plugin.go apps/lina-core/internal/service/plugin/plugin_startup_consistency.go apps/lina-core/internal/service/plugin/internal/testutil/testutil_services.go apps/lina-core/internal/service/cron/cron_single_node_test.go apps/lina-core/internal/service/file/file_data_scope_test.go apps/lina-core/internal/service/file/file_runtime_params_test.go apps/lina-core/internal/controller/i18n/i18n_v1_runtime_test.go apps/lina-core/internal/service/jobmgmt/jobmgmt_test.go openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `rg -n "orgspi\\.AssignmentService|ConsumerReadService|orgspi\\.Service|RuntimeService\\b" apps/lina-core -g '*.go'`无匹配，确认旧接口名和过宽`orgspi.Service`生产引用已清理。
- `rg -n "datascope\\.New\\([^\\n]*orgCapSvc\\)|orgCapSvc, orgCapSvc,|orgSvc:\\s*orgCapSvc\\b" apps/lina-core/internal -g '*.go'`无匹配，确认数据范围与工作台注入点均使用窄接口。

### FB-41 实施记录

- 根因：`tenantspi.Service`在`FB-39`后仍把普通`tenantcap.Service`、插件表过滤、HTTP 租户解析、目录读取、成员读取、用户租户关系写入、插件供给和启动一致性检查聚合成一个导出大接口，导致`httpstartup`、`capabilityhost`和测试替身只需要少量能力时仍被迫依赖或实现整套方法面。
- 移除旧总线型`tenantspi.Service`中的 reader/provider 聚合面，删除`PluginTableFilterProvider`、`ContextReader`、`DirectoryReader`和`MembershipReader`导出接口；调用方改用`tenantcap.Service`子资源和宿主窄接口满足所需契约。
- `httpstartup`在 FB-41 阶段明确了启动期实际需要的`tenantcap.Service`、`RequestResolver`、`ScopeService`、`UserMembershipService`、`PluginProvisioningService`、`StartupConsistencyService`和`PlatformBypassEvaluator`组合；FB-42 进一步将该组合恢复为公开`tenantspi.Service`名称。
- `plugin.NewHostServices`和`capabilityhost.New`改为使用本地租户窄接口，只要求普通租户能力、租户查询作用域和平台绕过判断；会话能力适配器进一步收敛为`Directory()`加`ScopeService`，撤销会话前通过`Directory().EnsureVisible(...)`执行租户可见性校验。
- 清理`capabilityhost`会话测试、`auth`租户流程测试和`menu`平台守卫测试中为了旧大接口保留的空方法、旧接口断言和过宽返回类型。
- `apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`已审查：未描述`tenantspi.Service`、`ContextReader`、`DirectoryReader`、`MembershipReader`或`PluginTableFilterProvider`，本次不改变插件公开`Tenant()`普通能力、动态 host service 或 provider SPI 文档口径，无需同步 README。
- 规则读取与技能：本轮已读取`lina-feedback`、`goframe-v2`、`AGENTS.md`、`.agents/rules/openspec.md`、`architecture.md`、`backend-go.md`、`plugin.md`、`data-permission.md`、`testing.md`和`documentation.md`；未修改 HTTP API、SQL、前端、缓存实现、运行时文案、语言包或开发工具脚本，确认`api-contract`、`database`、`frontend-ui`、`cache-consistency`、`i18n`和`dev-tooling`规则域无影响。
- `i18n`影响：未新增或修改运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源，确认无运行时`i18n`影响。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径、订阅状态或共享修订号策略；租户 provider 可用性和 fallback 语义保持不变，确认无缓存一致性影响。
- 数据权限影响：未新增数据读取、写入或存在性探测；会话撤销前的租户可见性校验从旧`EnsureTenantVisible`并列方法改为等价的`Directory().EnsureVisible`子资源路径，租户和数据范围边界不变。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本、Shell/Node 工具或`linactl`入口，确认无跨平台工具影响。
- DI 来源检查：未新增运行期依赖；`tenantspi.New`仍在 HTTP 启动期用同一`tenantProviderManager`、`pluginRuntime`和`bizCtxSvc`创建共享实例，后续消费者只接收该实例的更窄接口投影，未引入新的独立服务图、缓存后端或请求路径临时`New()`。
- 测试策略：本次属于后端 Go 插件宿主内部契约和测试替身收敛，不涉及用户可观察前端行为，未触发新增 E2E；通过 Go 包测试、OpenSpec 校验、静态检索和`git diff --check`完成验证。

### FB-41 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/tenantcap ./pkg/plugin/capability/tenantcap/tenantspi ./pkg/plugin/capability ./pkg/plugin/pluginhost ./pkg/plugin/pluginbridge ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/integration ./internal/service/plugin ./internal/cmd/internal/httpstartup ./internal/cmd ./internal/service/auth ./internal/service/menu ./internal/service/user ./internal/service/role ./internal/service/notify ./internal/service/session ./internal/service/middleware -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check -- apps/lina-core/pkg/plugin/capability/tenantcap/tenantspi/tenantspi.go apps/lina-core/internal/cmd/internal/httpstartup/httpstartup_runtime.go apps/lina-core/internal/service/plugin/plugin_host_services.go apps/lina-core/internal/service/plugin/internal/capabilityhost/capabilityhost.go apps/lina-core/internal/service/plugin/internal/capabilityhost/capabilityhost_session.go apps/lina-core/internal/service/plugin/internal/capabilityhost/capabilityhost_session_test.go apps/lina-core/pkg/plugin/capability/capability_test.go apps/lina-core/internal/service/plugin/plugin_startup_consistency_test.go apps/lina-core/internal/service/auth/auth_tenant_flow_test.go apps/lina-core/internal/service/menu/menu_platform_guard_test.go openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `rg -n "PluginTableFilterProvider|ContextReader|DirectoryReader|MembershipReader|tenantRuntimeService" apps/lina-core -g '*.go'`无匹配，确认旧总线 reader/provider 接口和临时本地启动聚合已从 Go 代码清理。

### FB-42 实施记录

- 根因：FB-41 为消除旧总线接口时让`tenantspi.New`返回内部`*serviceImpl`，导致主框架装配层缺少可命名的公开租户运行期契约；这与`orgspi.New`返回公开`orgspi.Service`的模式不一致，也不利于启动期共享服务图表达依赖边界。
- 恢复窄版`tenantspi.Service`公开接口，组合`tenantcap.Service`、`RequestResolver`、`ScopeService`、`UserMembershipService`、`PluginProvisioningService`、`StartupConsistencyService`、`TenantVisibilityService`和`PlatformBypassEvaluator`，作为主框架租户运行期契约。
- `tenantspi.New`恢复为返回`tenantspi.Service`；`httpRuntime.tenantSvc`改为直接使用`tenantspi.Service`，删除临时`tenantRuntimeService`本地接口。
- 将主框架 auth/user 实际依赖的`ListUserTenants`和`ValidateUserInTenant`纳入`UserMembershipService`，并补充`TenantVisibilityService`承载禁用租户能力时主框架单租户可见性校验的宽松 fallback；未恢复`DirectoryReader`、`MembershipReader`、`ContextReader`、`PluginTableFilterProvider`或目录批量读取/search flat 方法。
- 更新`tenantspi` fallback 测试：公开`Service`继续验证主框架内部 fallback，普通目录读取冲突场景改通过`Directory()`子资源验证。
- 规则读取与技能：本轮已读取`lina-feedback`、`goframe-v2`、`AGENTS.md`、`.agents/rules/openspec.md`、`architecture.md`、`backend-go.md`、`plugin.md`、`data-permission.md`、`testing.md`和`documentation.md`；未修改 HTTP API、SQL、前端、缓存实现、运行时文案、语言包或开发工具脚本，确认`api-contract`、`database`、`frontend-ui`、`cache-consistency`、`i18n`和`dev-tooling`规则域无影响。
- `i18n`影响：未新增或修改运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源，确认无运行时`i18n`影响。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径、订阅状态或共享修订号策略；租户 provider 可用性和 fallback 语义保持不变，确认无缓存一致性影响。
- 数据权限影响：未新增数据读取、写入或存在性探测；只是恢复主框架可命名运行期接口并保留既有租户成员、查询过滤和可见性校验语义，数据权限边界不变。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本、Shell/Node 工具或`linactl`入口，确认无跨平台工具影响。
- DI 来源检查：未新增运行期依赖；`tenantspi.New`仍在 HTTP 启动期用同一`tenantProviderManager`、`pluginRuntime`和`bizCtxSvc`创建共享实例，只是将返回类型从内部结构体改回公开接口，未引入新的独立服务图、缓存后端或请求路径临时`New()`。
- 测试策略：本次属于后端 Go 主框架 SPI 契约和测试修正，不涉及用户可观察前端行为，未触发新增 E2E；通过 Go 包测试、OpenSpec 校验、静态检索和`git diff --check`完成验证。

### FB-42 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/tenantcap ./pkg/plugin/capability/tenantcap/tenantspi ./pkg/plugin/capability ./pkg/plugin/pluginhost ./pkg/plugin/pluginbridge ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/integration ./internal/service/plugin ./internal/cmd/internal/httpstartup ./internal/cmd ./internal/service/auth ./internal/service/menu ./internal/service/user ./internal/service/role ./internal/service/notify ./internal/service/session ./internal/service/middleware -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check -- apps/lina-core/pkg/plugin/capability/tenantcap/tenantspi/tenantspi.go apps/lina-core/pkg/plugin/capability/tenantcap/tenantspi/tenantspi_fallback_test.go apps/lina-core/internal/cmd/internal/httpstartup/httpstartup_runtime.go openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `rg -n "PluginTableFilterProvider|ContextReader|DirectoryReader|MembershipReader|tenantRuntimeService" apps/lina-core -g '*.go'`无匹配，确认未恢复旧总线 reader/provider 接口或临时本地启动聚合。

### FB-43 实施记录

- 根因：`tenantcap.DirectoryService`仍暴露租户创建、更新、状态变更和删除，`tenantcap.MembershipService`仍暴露成员关系替换，`tenantcap.FilterService`仍把`*gdb.Model`查询构造器带入普通插件公共契约；同时`tenantspi`中还保留未被生产调用的 resolver chain、批量用户租户读取和 provider CRUD 转发，导致`tenantspi.Service`和租户能力公共面继续偏重。
- 将`tenantcap.DirectoryService`收敛为`Get`、`BatchGet`、`List`和`EnsureVisible`；将`tenantcap.MembershipService`收敛为`ListByUser`和`Validate`；删除公共租户 CRUD 输入类型和动态 guest 侧 unsupported 空方法。
- 将`tenantcap.FilterService`收敛为只读`Context`，把同进程`Apply(ctx, *gdb.Model, qualifier)`移到`tenantspi.PluginTableFilterService`；动态 guest 侧不再提供 no-op `Apply`，源码插件入口在注入插件表过滤器时显式断言`tenantspi.PluginTableFilterService`。
- 删除`tenantspi.Resolver`、`ResolverName`和未使用的`tenantspi_resolver_chain.go`；删除`UserMembershipProvider.BatchListUserTenants`和`serviceImpl.BatchListUserTenants`；删除`DirectoryProvider`和`linapro-tenant-core`provider adapter 上的租户 CRUD 转发。
- `linapro-tenant-core`provider 构造函数删除未使用的`tenantSvc`依赖；租户生命周期写入仍由该插件自己的 HTTP controller 和内部`tenant`service 闭环处理，不进入通用插件 capability 或动态`host service`注册事实。
- 同步更新`apps/lina-core/pkg/plugin/README.md`、`README.zh-CN.md`、当前变更`design.md`和增量规范：普通`tenantcap`只提供可序列化过滤上下文，源码插件表过滤通过`tenantspi.PluginTableFilterService`窄接口注入，租户生命周期和成员替换不属于插件公共面。
- 插件本地规范检查：本轮修改的`linapro-tenant-core`、`linapro-content-notice`、`linapro-monitor-operlog`、`linapro-monitor-loginlog`、`linapro-monitor-online`、`linapro-org-core`、`linapro-demo-source`和`linapro-demo-dynamic`根目录均未发现本地`AGENTS.md`。
- 规则读取与技能：本轮已读取`lina-feedback`、`goframe-v2`、`AGENTS.md`、`.agents/rules/openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`data-permission.md`、`cache-consistency.md`、`documentation.md`、`testing.md`、`i18n.md`和`.agents/instructions/markdown-format.instructions.md`；未修改 HTTP API、SQL、前端 UI 或开发工具脚本，确认`api-contract`、`database`、`frontend-ui`和`dev-tooling`规则域无影响。
- `i18n`影响：未新增运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源；仅更新 README 和 OpenSpec 文档，确认无运行时`i18n`资源变更。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径、订阅状态或共享修订号策略；删除的是未使用或过度暴露的接口面，确认无缓存一致性影响。
- 数据权限影响：插件公共写入面减少，租户生命周期写入和成员关系替换留在租户 owner 或宿主内部`tenantspi.UserMembershipService`；现有目录读取、可见性校验、成员校验和列表路径保持原有数据边界，动态插件仍只能调用已注册的只读/校验/治理方法。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本、Shell/Node 工具或`linactl`入口，确认无跨平台工具影响。
- DI 来源检查：删除了`linapro-tenant-core`provider adapter 的`tenantSvc`构造依赖，未新增运行期依赖；`tenantspi.New`仍由 HTTP 启动期使用共享`tenantProviderManager`、`pluginRuntime`和`bizCtxSvc`创建，源码插件表过滤器仍由`capabilityhost.New`用启动期共享`bizCtxAdapter`和`tenantSvc`构造并注入。
- 测试策略：本次属于后端 Go 插件宿主公共契约、源码插件 SPI、动态 guest 适配和 README/OpenSpec 治理变更，不涉及用户可观察前端行为，未触发新增 E2E；通过 Go 编译门禁、源码插件后端测试、OpenSpec 校验、静态检索和`git diff --check`验证。

### FB-43 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/tenantcap ./pkg/plugin/capability/tenantcap/tenantspi -count=1`通过。
- `cd apps/lina-core && go test ./pkg/plugin/capability ./pkg/plugin/pluginbridge ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/integration ./internal/service/plugin ./internal/cmd/internal/httpstartup ./internal/cmd ./internal/service/auth ./internal/service/menu ./internal/service/user ./internal/service/role ./internal/service/notify ./internal/service/session ./internal/service/middleware -count=1`通过。
- `cd apps/lina-plugins/linapro-tenant-core && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-content-notice && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-operlog && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-loginlog && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-online && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-org-core && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-demo-source && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/... -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check`通过。
- `rg -n "tenantspi\\.Resolver|tenantspi\\.ResolverName|tenantspi\\.NewResolverChain|tenantcap\\.CreateInput|tenantcap\\.UpdateInput|tenantcap\\.SetStatusInput|BatchListUserTenants|CreateTenant\\(|UpdateTenant\\(|SetTenantStatus\\(|DeleteTenant\\(|ReplaceByUser\\(context\\.Context, int, \\[\\]tenantcap\\.TenantID\\)|Tenant\\(\\)\\.Filter\\(\\)\\.Apply|Services\\.Tenant\\(\\)\\.Filter\\(\\)\\.Apply" apps/lina-core apps/lina-plugins -g '*.go' -g '*.md' -g '*.yaml' -g '*.yml'`无生产残留，确认旧租户公共写入、未使用 SPI 和普通`Filter().Apply`调用已清理。

### FB-44 实施记录

- 根因：`tenantspi.Service`在`FB-42`后重新变成主框架聚合接口，且源码插件为了获得`*gdb.Model`表过滤能力仍需要从普通`Tenant().Filter()`做`tenantspi.PluginTableFilterService`类型断言；这会让普通租户过滤上下文和源码插件同进程查询构造器混在同一契约里。
- 删除公开`tenantspi.Service`聚合接口，`tenantspi.New`返回内部`*serviceImpl`并让启动装配、宿主适配和调用方依赖`tenantcap.Service`、`ScopeService`、`UserMembershipService`等真实窄接口。
- 将`pluginhost.Services`从`capability.Services`类型别名改为源码插件运行期接口，内嵌普通`capability.Services`并新增`TenantTableFilter() tenantspi.PluginTableFilterService`；普通`capability.Services`和动态插件协议仍不暴露`*gdb.Model`能力。
- `capabilityhost`显式保存并发布启动期共享的`tenantspi.PluginTableFilterService`，源码插件`linapro-content-notice`、`linapro-monitor-operlog`、`linapro-monitor-loginlog`、`linapro-monitor-online`、`linapro-org-core`和`linapro-demo-source`统一改用`services.TenantTableFilter()`，不再对`Tenant().Filter()`做类型断言。
- 删除`PluginGovernanceService`别名，`capabilityowner`和`linapro-tenant-core`直接依赖真实`tenantcap.PluginService`；删除插件表过滤中的`PlatformBypassEvaluator`注入，平台绕过统一来自`bizctxcap.CurrentContext.PlatformBypass`。
- 将`tenantspi`目录 provider SPI 方法名从`CurrentTenantInfo`、`BatchGetTenants`、`SearchTenants`和`EnsureTenantsVisible`收敛为`Info`、`BatchGet`、`List`和`EnsureVisible`，并将`serviceImpl`包内 helper 降为小写，减少无必要导出面。
- 同步更新`apps/lina-core/pkg/plugin/README.md`、`README.zh-CN.md`、当前变更`design.md`和增量规范，明确普通`Tenant().Filter()`只提供可序列化上下文，同进程表过滤只通过`Services.TenantTableFilter()`显式注入。
- 插件本地规范检查：本轮修改的`linapro-tenant-core`、`linapro-content-notice`、`linapro-monitor-operlog`、`linapro-monitor-loginlog`、`linapro-monitor-online`、`linapro-org-core`和`linapro-demo-source`根目录均未发现本地`AGENTS.md`。
- 规则读取与技能：本轮已读取`lina-feedback`、`karpathy-guidelines`、`goframe-v2`、`AGENTS.md`、`.agents/rules/openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`data-permission.md`、`cache-consistency.md`、`testing.md`、`i18n.md`和`documentation.md`；未修改 HTTP API、SQL、前端 UI 或开发工具脚本，确认`api-contract`、`database`、`frontend-ui`和`dev-tooling`规则域无影响。
- `i18n`影响：未新增运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源；仅更新 README 和 OpenSpec 文档，确认无运行时`i18n`资源变更。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径、订阅状态或共享修订号策略；删除的是聚合接口、别名和重复依赖注入，确认无缓存一致性影响。
- 数据权限影响：普通`tenantcap.FilterService`只保留上下文读取，源码插件表过滤仍由启动期共享`bizctxcap.Service`读取当前租户、actor 和平台绕过状态后在数据库查询阶段注入谓词；动态插件仍只能通过可序列化上下文或领域 owner/RecordStore 完成租户隔离。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本、Shell/Node 工具或`linactl`入口，确认无跨平台工具影响。
- DI 来源检查：未新增运行期依赖；`tenantspi.NewPluginTableFilter`从依赖`bizCtxSvc + tenantSvc`收敛为仅依赖启动期共享`bizCtxAdapter`，由`capabilityhost.New`创建一次并传入用户、字典、文件、任务等领域 adapter，同时通过`pluginhost.Services.TenantTableFilter()`暴露给源码插件。没有请求路径临时`New()`、独立服务图或新缓存后端。
- 测试策略：本次属于后端 Go 插件宿主契约、源码插件运行期接口、测试替身和 README/OpenSpec 治理变更，不涉及用户可观察前端行为，未触发新增 E2E；通过 Go 编译门禁、源码插件包测试、OpenSpec 校验、静态检索和`git diff --check`验证。

### FB-44 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/tenantcap ./pkg/plugin/capability/tenantcap/tenantspi ./pkg/plugin/pluginhost ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/integration ./internal/service/plugin -count=1`通过。
- `cd apps/lina-core && go test ./pkg/plugin/capability ./pkg/plugin/pluginbridge ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/plugin/internal/wasm ./internal/cmd/internal/httpstartup ./internal/cmd -count=1`通过。
- `cd apps/lina-plugins/linapro-content-notice && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-operlog && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-loginlog && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-online && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-org-core && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-demo-source && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-tenant-core && GOWORK=off go test ./backend/... -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check`通过。
- `git -C apps/lina-plugins diff --check`通过。
- `rg -n "PlatformBypassEvaluator|PluginGovernanceService|NewPluginTableFilter\\([^\\)]*,|tenantSvc\\.Filter\\(\\).*PluginTableFilterService|Tenant\\(\\)\\.Filter\\(\\)\\.Apply|Services\\.Tenant\\(\\)\\.Filter\\(\\)\\.Apply|PluginTableFilter\\(\\)|CurrentTenantInfo|BatchGetTenants|SearchTenants|EnsureTenantsVisible" apps/lina-core apps/lina-plugins -g '*.go' -g '*.md' -g '*.yaml' -g '*.yml'`无匹配，确认旧平台绕过注入、治理别名、旧构造签名、普通过滤类型断言和多余导出 helper 已清理。

### FB-45 实施记录

- 根因：`FB-44`为删除总线式聚合接口时过度收敛，导致`tenantspi.New`重新返回内部`*serviceImpl`，把实现类型暴露给主框架装配层；这不符合 SPI 构造函数应返回公开契约的边界。
- 恢复公开窄版`tenantspi.Service`接口作为`tenantspi.New`返回类型，组合`tenantcap.Service`、`RequestResolver`、`ScopeService`、`UserMembershipService`、`PluginProvisioningService`、`StartupConsistencyService`和`TenantVisibilityService`，并保留当前宿主守卫实际依赖的`PlatformBypass(ctx)`方法。
- 未恢复旧`RuntimeService`、`PluginTableFilterProvider`、`ContextReader`、`DirectoryReader`、`MembershipReader`或`PlatformBypassEvaluator`类型；`PluginTableFilterService`仍只通过`pluginhost.Services.TenantTableFilter()`或构造函数显式注入，不进入普通租户能力或`tenantspi.New`返回契约。
- `apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`已通过静态检索审查：文档未描述`tenantspi.New`返回类型或内部`serviceImpl`，且`FB-44`后已有普通`Tenant().Filter()`与`TenantTableFilter()`边界说明，本次无需同步 README。
- 规则读取与技能：本轮已读取`lina-feedback`、`karpathy-guidelines`、`goframe-v2`、`AGENTS.md`、`.agents/rules/openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`data-permission.md`、`cache-consistency.md`、`testing.md`、`i18n.md`和`documentation.md`；未修改 HTTP API、SQL、前端 UI 或开发工具脚本，确认`api-contract`、`database`、`frontend-ui`和`dev-tooling`规则域无影响。
- `i18n`影响：未新增运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源；仅更新 Go 类型签名和 OpenSpec 记录，确认无运行时`i18n`资源变更。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径、订阅状态或共享修订号策略；只是把返回类型从内部结构体改为公开接口，确认无缓存一致性影响。
- 数据权限影响：未新增数据读取、写入、存在性探测或租户边界变化；租户目录读取、成员校验、平台绕过和查询过滤语义保持不变。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本、Shell/Node 工具或`linactl`入口，确认无跨平台工具影响。
- DI 来源检查：未新增运行期依赖；`tenantspi.New`仍由 HTTP 启动期使用同一个`tenantProviderManager`、`pluginRuntime`和`bizCtxSvc`创建共享实例，返回类型变化不引入独立服务图、请求路径临时`New()`或新缓存后端。
- 测试策略：本次属于后端 Go 主框架 SPI 构造契约和 OpenSpec 治理修正，不涉及用户可观察前端行为，未触发新增 E2E；通过 Go 编译门禁、OpenSpec 校验、静态检索和`git diff --check`验证。

### FB-45 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/tenantcap ./pkg/plugin/capability/tenantcap/tenantspi -count=1`通过。
- `cd apps/lina-core && go test ./internal/cmd/internal/httpstartup ./internal/cmd ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin -count=1`通过。
- `cd apps/lina-core && go test ./pkg/plugin/capability ./pkg/plugin/pluginbridge ./internal/service/plugin/internal/wasm ./internal/service/menu ./internal/service/user ./internal/service/role ./internal/service/middleware ./internal/service/notify ./internal/service/usermsg -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check -- apps/lina-core/pkg/plugin/capability/tenantcap/tenantspi/tenantspi.go openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `rg -n 'func New\\(manager \\*Manager, runtime ProviderRuntime, bizCtxSvc bizctxcap\\.Service\\) \\*serviceImpl|RuntimeService\\b|PluginTableFilterProvider|ContextReader|DirectoryReader|MembershipReader|tenantRuntimeService|PlatformBypassEvaluator|PluginGovernanceService' apps/lina-core/pkg/plugin/capability/tenantcap/tenantspi apps/lina-core/internal apps/lina-core/pkg/plugin -g'*.go'`无匹配，确认`New`不再返回内部结构体且旧总线接口没有回到 Go 代码。

### FB-46 实施记录

- 根因：`FB-45`恢复公开`tenantspi.Service`后，把`PluginProvisioningService`、`StartupConsistencyService`和`TenantVisibilityService`三个单方法接口嵌入`Service`，同时把`PlatformBypass(ctx)`作为裸方法放在`Service`上；这让同属宿主治理面的能力在接口形态上不一致，也让`tenantspi`公开面看起来像按每个方法机械拆接口。
- 将上述宿主治理方法合并为公开`HostGovernanceService`，由`tenantspi.Service`统一嵌入；`PlatformBypass`、`EnsureTenantVisible`、`ProvisionAutoEnabledTenantPlugins`和`ValidateUserMembershipStartupConsistency`现在位于同一个宿主治理契约中。
- 删除仅承载单方法断言的`PluginProvisioningProvider`导出接口；`ProvisionAutoEnabledTenantPlugins`内部改用匿名可选 facet 判断 provider 是否支持该能力，`linapro-tenant-core`删除对应旧接口断言但保留方法实现。
- `plugin.New`和测试 helper 不再引用已删除的`tenantspi.PluginProvisioningService`，改用`plugin`包内消费侧窄接口接收启动 provisioning 依赖；这不扩大`tenantspi`公开面，也不强迫只需要 provisioning 的内部测试替身实现完整治理接口。
- `apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`已通过静态检索审查：文档只描述`tenantspi.PluginTableFilterService`的源码插件表过滤边界，未描述本次删除的宿主治理接口或 provider 断言，无需同步 README。
- 规则读取与技能：本轮已读取`lina-feedback`、`goframe-v2`、`AGENTS.md`、`.agents/rules/openspec.md`、`architecture.md`、`backend-go.md`、`plugin.md`、`data-permission.md`、`cache-consistency.md`、`testing.md`、`i18n.md`、`documentation.md`和`.agents/instructions/markdown-format.instructions.md`；未修改 HTTP API、SQL、前端 UI 或开发工具脚本，确认`api-contract`、`database`、`frontend-ui`和`dev-tooling`规则域无影响。`linapro-tenant-core`插件根目录不存在本地`AGENTS.md`。
- `i18n`影响：未新增运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源；仅调整 Go 类型签名、编译期断言和 OpenSpec 记录，确认无运行时`i18n`资源变更。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径、订阅状态或共享修订号策略；租户 provider 可用性和 fallback 语义保持不变。
- 数据权限影响：未新增数据读取、写入、存在性探测或租户边界变化；租户可见性、平台绕过、成员启动校验和租户插件供给仍复用原实现路径。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本、Shell/Node 工具或`linactl`入口，确认无跨平台工具影响。
- DI 来源检查：未新增运行期依赖；`tenantspi.New`仍由 HTTP 启动期使用同一个`tenantProviderManager`、`pluginRuntime`和`bizCtxSvc`创建共享实例。`plugin.New`的 provisioning 依赖只是从旧`tenantspi`单方法接口改为包内消费侧接口，传递路径和共享实例策略不变，未引入独立服务图、请求路径临时`New()`或新缓存后端。
- 测试策略：本次属于后端 Go 主框架 SPI 公开面简化和源码插件 provider 断言清理，不涉及用户可观察前端行为，未触发新增 E2E；通过 Go 包测试、宿主启动装配测试、源码插件 provider 测试、OpenSpec 校验、静态检索和`git diff --check`验证。

### FB-46 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/tenantcap ./pkg/plugin/capability/tenantcap/tenantspi -count=1`通过。
- `cd apps/lina-core && go test ./internal/service/plugin ./internal/service/plugin/internal/lifecycle -count=1`通过。
- `cd apps/lina-core && go test ./internal/cmd/internal/httpstartup -count=1`通过。
- `cd apps/lina-core && go test ./internal/cmd -count=1`通过。
- `cd apps/lina-plugins/linapro-tenant-core && GOWORK=off go test ./backend/internal/service/provider -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check -- apps/lina-core/pkg/plugin/capability/tenantcap/tenantspi/tenantspi.go apps/lina-core/pkg/plugin/capability/tenantcap/tenantspi/tenantspi_host_impl.go apps/lina-core/internal/service/plugin/plugin.go apps/lina-core/internal/service/plugin/plugin_test.go apps/lina-core/internal/service/plugin/plugin_startup_consistency.go openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `git -C apps/lina-plugins diff --check -- linapro-tenant-core/backend/internal/service/provider/provider.go`通过。
- `rg -n "type (PluginProvisioningService|StartupConsistencyService|TenantVisibilityService|PluginProvisioningProvider) interface|tenantspi\\.(PluginProvisioningService|StartupConsistencyService|TenantVisibilityService|PluginProvisioningProvider)" apps/lina-core apps/lina-plugins/linapro-tenant-core -g'*.go'`无匹配，确认旧`tenantspi`单方法公开接口和 provider 别名已从 Go 代码清理。
- `rg -n "HostGovernanceService|PlatformBypass\\(ctx context.Context\\) bool" apps/lina-core/pkg/plugin/capability/tenantcap/tenantspi apps/lina-core/internal/service/plugin -g'*.go'`仅命中`tenantspi.HostGovernanceService`、实现方法和插件治理本地 guard 接口，确认`PlatformBypass`已纳入统一宿主治理契约。
- `rg -n "tenantspi|HostGovernanceService|PluginProvisioningService|StartupConsistencyService|TenantVisibilityService|PluginProvisioningProvider" apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md`仅命中既有`tenantspi.PluginTableFilterService`描述，确认无需同步 README。

### FB-47 实施记录

- 根因：`tenantspi`目录在前序 SPI 收敛后仍保留`tenantspi_manager.go`、`tenantspi_plugin_filter_service.go`和`tenantspi_plugin_filter_apply.go`等薄文件，其中多个非主文件低于 50 行，职责拆分过细，不符合后端复杂度治理对小文件合并的要求。
- 将 provider 状态 DTO 转换辅助从`tenantspi_manager.go`合并到`tenantspi_host_impl.go`，让可用性、状态和 provider fallback 实现集中在同一 host runtime 实现文件内。
- 将`PluginTableFilterService`契约、`NewPluginTableFilter`构造函数、`pluginTableFilterService`实现、上下文转换和`Apply`查询过滤逻辑合并到`tenantspi_plugin_filter.go`，让源码插件表过滤能力保留一个非主职责文件。
- 删除`tenantspi_manager.go`、`tenantspi_plugin_filter_service.go`和`tenantspi_plugin_filter_apply.go`；整理后`tenantspi`目录仅保留主契约文件、host 实现文件、插件表过滤文件和对应测试文件。
- `apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`已评估：本次不修改插件公开能力、动态 host service、源码插件入口或文档描述的能力边界，无需同步 README。
- 规则读取与技能：本轮已读取`lina-feedback`、`goframe-v2`、`AGENTS.md`、`.agents/rules/openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`testing.md`、`documentation.md`和`.agents/instructions/markdown-format.instructions.md`；未修改 HTTP API、SQL、前端 UI、运行时文案、缓存实现或开发工具脚本，确认`api-contract`、`database`、`frontend-ui`、`i18n`、`cache-consistency`、`data-permission`和`dev-tooling`规则域无影响。
- `i18n`影响：未新增或修改运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源；仅移动 Go 同包实现和更新 OpenSpec 记录，确认无运行时`i18n`资源变更。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径、订阅状态或共享修订号策略；provider 状态转换只是同包移动，语义不变。
- 数据权限影响：未新增数据读取、写入、存在性探测或租户边界变化；租户表过滤`Context`和`Apply`实现只做同包合并，生成谓词和平台绕过语义保持不变。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本、Shell/Node 工具或`linactl`入口，确认无跨平台工具影响。
- DI 来源检查：未新增运行期依赖；`NewPluginTableFilter`仍只显式接收启动期共享的`bizctxcap.Service`，`tenantspi.New`依赖来源和共享实例策略不变。
- 测试策略：本次属于后端 Go 同包文件组织治理，不涉及用户可观察前端行为，未触发新增 E2E；通过 Go 包测试、OpenSpec 校验、静态检索、文件行数检查和`git diff --check`验证。

### FB-47 验证

- `wc -l apps/lina-core/pkg/plugin/capability/tenantcap/tenantspi/*.go`确认非主生产文件仅剩`tenantspi_host_impl.go`和`tenantspi_plugin_filter.go`，且`tenantspi_plugin_filter.go`为 89 行，不再存在低于 50 行的过细非主生产文件。
- `git diff --check -- apps/lina-core/pkg/plugin/capability/tenantcap/tenantspi openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `cd apps/lina-core && go test ./pkg/plugin/capability/tenantcap/tenantspi -count=1`通过。
- `cd apps/lina-core && go test ./pkg/plugin/capability/tenantcap ./internal/service/plugin/internal/capabilityhost -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。

### FB-48 实施记录

- 根因：插件宿主装配为了追求窄依赖在`plugin`、`capabilityhost`、`hostconfig`和`pluginconfig`相邻包中重复声明了`GetRaw`、`RevokeSession`和`Current`单方法接口；这些接口没有隔离新的变化点，反而增加了命名、类型断言和构造签名理解成本。
- 将原始配置读取契约归属到`config.RawReader`，并纳入`config.Service`；`plugin.NewHostConfigService`、`pluginconfig.NewFactoryWithHostStaticConfig`和`hostconfig.NewStaticCapabilityAdapter`统一接收该 owner 契约，启动装配不再对`configSvc`做插件侧接口类型断言。
- 将会话撤销契约归属到`auth.SessionRevoker`，`plugin.NewHostServices`、`capabilityhost.New`和 session capability adapter 直接使用该契约，删除`HostAuthSessionRevoker`、`AuthSessionRevoker`和`sessionAuthRevoker`重复定义。
- 将业务上下文当前值依赖收敛为已有`bizctxcap.Service`，删除`HostBizContextProvider`和`BizContextProvider`重复定义；`bizCtxAdapter`保留对`bizctxcap.CurrentFromContext`的 fallback 行为，避免破坏动态和测试上下文注入。
- 清理测试和 testutil 中旧接口类型断言，直接传入`config.Service`或`config.RawReader`测试替身。
- `apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`已评估：本次仅收敛宿主内部装配接口和测试替身，不改变插件公开 capability、动态 host service、源码插件入口或文档中的能力边界，无需同步 README。
- 规则读取与技能：本轮已读取`lina-feedback`、`goframe-v2`、`AGENTS.md`、`.agents/rules/openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`testing.md`和`documentation.md`；未修改 HTTP API、SQL、前端 UI、运行时文案、缓存实现或开发工具脚本，确认`api-contract`、`database`、`frontend-ui`、`i18n`、`cache-consistency`、`data-permission`和`dev-tooling`规则域无影响。
- `i18n`影响：未新增或修改运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源，确认无运行时`i18n`影响。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径、订阅状态或共享修订号策略；`configSvc`、`authSvc`和`bizCtxSvc`仍复用启动期共享实例。
- 数据权限影响：未新增数据读取、写入、存在性探测或租户边界变化；仅调整依赖接口归属，原始配置读取、会话撤销和业务上下文读取语义保持不变。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本、Shell/Node 工具或`linactl`入口，确认无跨平台工具影响。
- DI 来源检查：未新增运行期依赖；`config.RawReader`、`auth.SessionRevoker`和`bizctxcap.Service`均由既有启动期共享`configSvc`、`authSvc`和`bizCtxSvc`提供，创建位置、传递路径和共享实例策略保持不变。
- 测试策略：本次属于后端 Go 接口复杂度治理和启动装配编译面收敛，不涉及用户可观察前端行为，未触发新增 E2E；通过 Go 包测试、宿主启动包测试、OpenSpec 校验、静态检索和`git diff --check`验证。

### FB-48 验证

- `cd apps/lina-core && go test ./internal/service/config ./internal/service/auth ./internal/service/plugin/internal/hostconfig ./internal/service/plugin/internal/pluginconfig ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm ./internal/service/plugin ./internal/cmd/internal/httpstartup -count=1`通过。
- `cd apps/lina-core && go test ./internal/cmd -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check -- apps/lina-core/internal/service/config/config.go apps/lina-core/internal/service/config/config_raw.go apps/lina-core/internal/service/auth/auth.go apps/lina-core/internal/service/plugin/plugin_host_services.go apps/lina-core/internal/service/plugin/plugin_config.go apps/lina-core/internal/service/plugin/internal/capabilityhost/capabilityhost.go apps/lina-core/internal/service/plugin/internal/capabilityhost/capabilityhost_adapters.go apps/lina-core/internal/service/plugin/internal/capabilityhost/capabilityhost_session.go apps/lina-core/internal/service/plugin/internal/hostconfig/hostconfig_reader.go apps/lina-core/internal/service/plugin/internal/pluginconfig/pluginconfig.go apps/lina-core/internal/service/plugin/internal/pluginconfig/pluginconfig_test.go apps/lina-core/internal/cmd/internal/httpstartup/httpstartup_runtime.go apps/lina-core/internal/cmd/internal/httpstartup/httpstartup_routes_test.go apps/lina-core/internal/service/plugin/plugin_host_services_test.go apps/lina-core/internal/service/plugin/plugin_test.go apps/lina-core/internal/service/plugin/plugin_startup_consistency_test.go apps/lina-core/internal/service/plugin/internal/testutil/testutil_services.go apps/lina-core/internal/service/plugin/internal/wasm/wasm_host_service_hostconfig_test.go openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `rg -n "HostConfigRawReader|hostconfigadapter\\.RawConfigReader|pluginconfig\\.HostStaticConfigReader|PluginConfigHostStaticReader|HostAuthSessionRevoker|capabilityhost\\.AuthSessionRevoker|sessionAuthRevoker|HostBizContextProvider|capabilityhost\\.BizContextProvider|type HostStaticConfigReader interface|type RawConfigReader interface|type AuthSessionRevoker interface|type BizContextProvider interface" apps/lina-core -g '*.go'`无匹配，确认重复单方法接口已清理。

- [x] **FB-26**: 删除`Plugins().Capability().BatchGetStatus`聚合框架能力状态查询及动态`plugins.capability.status.batch_get`方法。
- [x] **FB-27**: 将`PluginLifecycle`、插件启用状态读取、`TenantPluginGovernance`和`TenantFilter`治理能力按领域归属收敛到`Plugins()`与`Tenant()`组件，并允许动态插件通过已注册领域治理方法按`hostServices`授权调用。
- [x] **FB-28**: 将`Route`领域动态路由元数据契约重命名为`GetMetadata`/`Metadata`，统一公开契约命名风格
- [x] **FB-29**: 消除`plugincap.CapabilityService`与`StateService`的领域职责重复，将`StateService`降级为源码插件兼容适配器。
- [x] **FB-30**: 审查并迁移普通`capability`包中的宿主实现和宿主类型泄漏，保持`hostconfigcap`与`tenantcap`等普通契约边界一致。
- [x] **FB-31**: 将插件启用状态普通领域能力命名统一为`StateService`，并将旧 bool 风格兼容接口改名为`StateCompatService`。
- [x] **FB-32**: 删除`bizctxcap`公共 service adapter，实现只保留宿主内部 adapter，保持`bizctxcap`为上下文契约与值 helper。
- [x] **FB-35**: 删除`plugincap.StateCompatService`公开兼容契约，将 bool 风格插件启用状态读取下沉为宿主内部私有接口并清理对应调用点。

### FB-37 实施记录

- 根因：`usercap.Service`顶层直接承载`ReplaceRoles`，会把用户基础读写、状态、凭证能力与用户角色关联关系混在同一方法面；而`orgcap`已经通过`Assignment()`聚合关联关系，用户领域缺少同样的子领域边界。
- 将`usercap.Service`中的角色关联入口调整为`Assignment() AssignmentService`，并让`usercap.AssignmentService`只包含`ReplaceRoles(ctx, id, roleIDs)`；后续用户角色关联关系方法继续归入该子领域。
- 宿主`internal/service/user/capabilityadapter`新增轻量`userAssignmentAdapter`，复用同一个启动期共享`userCapabilityAdapter`和 user owner；`ReplaceRoles`仍先调用`EnsureVisible`校验目标用户可见性，再通过既有`owner.Update(... RoleIds: roleIDs)`路径写入。
- 动态 guest 适配器保留`Assignment().ReplaceRoles`类型契约，但返回未发布动态方法错误`users.assignment.roles.replace`；本轮不新增动态`host service`注册、protocol 常量、catalog 条目或 WASM dispatcher 分支。
- 同步更新宿主和源码插件测试替身，让 fake user service 通过`Assignment()`返回只包含`ReplaceRoles`的子服务；更新 README 和增量规范，明确当前动态`users`服务只发布投影、批量、解析、列表和可见性方法。
- 插件本地规范检查：本轮涉及的`linapro-content-notice`、`linapro-org-core`和`linapro-tenant-core`根目录均未发现`AGENTS.md`。
- 规则读取与技能：本轮已读取`lina-feedback`、`goframe-v2`、`AGENTS.md`、`.agents/rules/openspec.md`、`architecture.md`、`backend-go.md`、`plugin.md`、`documentation.md`、`testing.md`、`data-permission.md`、`i18n.md`、`cache-consistency.md`和`.agents/instructions/markdown-format.instructions.md`；未修改 HTTP API、SQL、前端或开发工具脚本，确认`api-contract`、`database`、`frontend-ui`和`dev-tooling`规则域无影响。
- `i18n`影响：未新增运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源；仅更新 README 和 OpenSpec 文档，确认无运行时`i18n`资源变更。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径或共享修订号策略；角色替换仍走既有 user owner 更新路径，授权修订和缓存影响保持宿主 owner 既有语义。
- 数据权限影响：授权关系变更前仍通过`EnsureVisible`校验目标用户可见性；角色 ID 校验、事务和写入语义继续复用既有 user owner 更新路径，未新增数据存在性探测面。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本或`linactl`入口。
- DI 来源检查：未新增运行期依赖；`Assignment()`只包装当前`userCapabilityAdapter`，动态 guest stub 复用既有`baseService`invoker，测试替身也只在现有 fake service 内返回子服务。
- 测试策略：本次属于后端 Go 插件宿主契约、源码插件测试替身和 README/OpenSpec 治理变更，不涉及用户可观察前端行为，未触发新增 E2E；通过 Go 编译门禁、源码插件包测试、宿主启动装配测试、静态检索、`git diff --check`和`openspec validate`验证。

### FB-37 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/usercap ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/user/capabilityadapter ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/wasm ./internal/service/plugin -count=1`通过。
- `cd apps/lina-core && go test ./internal/cmd -count=1`通过。
- `cd apps/lina-plugins/linapro-content-notice && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-org-core && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-tenant-core && GOWORK=off go test ./backend/... -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check -- apps/lina-core/pkg/plugin/capability/usercap/usercap.go apps/lina-core/internal/service/user/capabilityadapter/user_capability.go apps/lina-core/pkg/plugin/pluginbridge/internal/domainhostcall/domainhostcall_users.go apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md openspec/changes/standardize-plugin-domain-services/tasks.md openspec/changes/standardize-plugin-domain-services/specs/plugin-host-domain-capabilities/spec.md`通过。
- `git -C apps/lina-plugins diff --check -- linapro-content-notice/backend/internal/service/notice/notice_usercap_test.go linapro-org-core/backend/plugin_provider_test.go linapro-tenant-core/backend/plugin_provider_test.go linapro-tenant-core/backend/internal/service/membership/membership_test.go linapro-tenant-core/backend/internal/service/impersonate/impersonate_test.go linapro-tenant-core/backend/internal/service/provider/provider_provisioning_test.go`通过。
- `rg -n "func \\([^)]*\\) ReplaceRoles\\(|ReplaceRoles\\(" apps/lina-core apps/lina-plugins -g "*.go"`仅命中`usercap.AssignmentService`、宿主 assignment adapter、动态 unsupported stub 和测试替身，确认顶层`usercap.Service`不再直接暴露`ReplaceRoles`。
- `rg -n "users\\.roles\\.replace|HostServiceMethodUsers.*Role|users\\.assignment\\.roles\\.replace" apps/lina-core apps/lina-plugins openspec/changes/standardize-plugin-domain-services --glob "*.go" --glob "*.md" --glob "*.yaml" --glob "*.yml" --glob "!**/public/**"`仅命中动态 guest unsupported stub 的`users.assignment.roles.replace`错误标识，确认未发布旧用户角色动态方法。

### FB-36 实施记录

- 根因：前一轮已将租户插件治理和租户过滤委托到`tenantcap.Service`，但`pluginhost.Services`仍保留`TenantPluginGovernance()`和`TenantFilter()`顶层快捷入口，导致租户领域能力和`plugincap`领域能力的收敛口径不一致。
- 删除`pluginhost.Services.TenantPluginGovernance()`和`TenantFilter()`顶层接口方法，`capabilityhost`目录和测试替身不再实现这两个顶层方法。
- 将`tenantcap.FilterService`扩展为租户领域过滤子能力，包含`Context()`和源码插件同进程使用的`Apply(ctx, model, qualifier)`；`tenantspi.PluginTableFilterService`改为该领域契约别名，动态 guest 侧`Apply`保持本地 no-op，不通过 WASM wire 传输`*gdb.Model`。
- 源码插件`linapro-content-notice`、`linapro-monitor-operlog`、`linapro-org-core`、`linapro-monitor-loginlog`、`linapro-monitor-online`、`linapro-demo-source`和`linapro-tenant-core`统一改用`services.Tenant().Filter()`或`services.Tenant().Plugins()`。
- 宿主 provider env 构造改为从`services.Tenant()`读取租户过滤和租户插件治理，README、当前变更设计与增量规范同步删除源码插件顶层快捷入口口径。
- 插件本地规范检查：本轮修改的`linapro-content-notice`、`linapro-monitor-operlog`、`linapro-org-core`、`linapro-monitor-loginlog`、`linapro-monitor-online`、`linapro-demo-source`和`linapro-tenant-core`根目录均未发现`AGENTS.md`。
- 规则读取与技能：本轮已读取`lina-feedback`、`goframe-v2`、`karpathy-guidelines`、`lina-review`、`AGENTS.md`、`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`api-contract.md`、`data-permission.md`、`cache-consistency.md`、`i18n.md`、`testing.md`、`dev-tooling.md`和`.agents/instructions/markdown-format.instructions.md`。
- `i18n`影响：未新增运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源；仅更新 README 和 OpenSpec 文档，确认无运行时`i18n`资源变更。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径、共享修订号或状态快照策略；租户插件治理继续进入既有插件 owner 和租户 owner 路径。
- 数据权限影响：无新增数据读取、写入或存在性探测；源码插件调用入口从顶层快捷方法迁移到同一`Tenant()`领域子能力，租户插件写入、默认供给和插件表过滤仍复用既有租户、权限和当前业务上下文边界。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本或`linactl`入口。
- DI 来源检查：未新增运行期依赖；`capabilityhost.New`继续创建启动期共享`tenantFilterSvc`，并将同一实例传入用户、字典、文件、任务等领域 adapter 和`Tenant().Filter()`领域 wrapper；`Tenant().Plugins()`继续复用启动期共享的插件领域 adapter。
- 治理扫描影响：`pkg/plugin/capability`根包治理测试已同步记录`tenantcap.FilterService.Apply`为唯一源码插件同进程`gdb.Model`例外，动态 wire 契约仍禁止传输`*gdb.Model`。
- 测试策略：本次属于后端 Go 插件宿主契约、源码插件调用点和 README/OpenSpec 治理变更，不涉及用户可观察前端行为，未触发新增 E2E；通过 Go 编译门禁、源码插件包测试、宿主启动装配测试、静态检索、`git diff --check`和`openspec validate`验证。

### FB-36 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability -count=1`通过。
- `cd apps/lina-core && go test ./pkg/plugin/capability/tenantcap ./pkg/plugin/capability/tenantcap/tenantspi ./pkg/plugin/pluginhost ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/wasm ./internal/service/plugin ./internal/cmd -count=1`通过。
- `cd apps/lina-plugins/linapro-content-notice && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-operlog && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-org-core && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-loginlog && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-online && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-demo-source && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-tenant-core && GOWORK=off go test ./backend/... -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check -- apps/lina-core/pkg/plugin apps/lina-core/internal/service/plugin apps/lina-plugins/linapro-content-notice apps/lina-plugins/linapro-monitor-operlog apps/lina-plugins/linapro-org-core apps/lina-plugins/linapro-monitor-loginlog apps/lina-plugins/linapro-monitor-online apps/lina-plugins/linapro-demo-source apps/lina-plugins/linapro-tenant-core openspec/changes/standardize-plugin-domain-services`通过。
- `rg -n "\b(services|sourceServices|hostServices|registrar\.Services\(\))\.TenantPluginGovernance\(|\b(services|sourceServices|hostServices|registrar\.Services\(\))\.TenantFilter\(|func \([^)]*\) TenantPluginGovernance|func \([^)]*\) TenantFilter|TenantPluginGovernance\(\) tenantspi|TenantFilter\(\) tenantspi" apps/lina-core apps/lina-plugins --glob "*.go" --glob "!**/public/**"`无匹配，确认旧顶层源码插件快捷入口的 Go 调用和实现已清理。

### FB-35 实施记录

- 根因：`StateCompatService`只是源码插件和宿主内部的 bool 读状态兼容面，公开后会和已存在的`plugincap.StateService`重复职责，也让公开契约继续暴露 fail-closed 兼容语义。
- 删除`apps/lina-core/pkg/plugin/capability/plugincap/plugincap_state.go`，把 bool 风格插件启用状态读取下沉为宿主内部`internal/service/plugin/internal/statecompat.Service`，由`plugin.Service`本身和宿主内部 provider/runtime 直接实现。
- `pluginhost.NewStorageProviderRuntime`、`capabilityhost.New`和`capabilityowner.NewCapabilityAdapter`已改为依赖宿主内部`statecompat.Service`，不再引用公开`plugincap.StateCompatService`。
- `i18n`影响：无运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源变更。
- 缓存一致性影响：无新增缓存后端或失效策略；仅将既有插件 enablement bool 读取从公开契约下沉到宿主内部接口，仍复用同一启动期共享状态服务。
- 数据权限影响：无新增数据读写、存在性探测或租户/组织边界；只调整状态读取契约的暴露边界。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本或`linactl`入口。
- DI 来源检查：未新增运行期依赖；`statecompat.Service`由`plugin.Service`及其运行期共享实例直接实现，`capabilityhost.New`和`pluginhost.NewStorageProviderRuntime`继续接收启动期共享的同一状态读取能力。
- 测试策略：本次属于后端 Go 和 OpenSpec 治理变更，不涉及用户可观察前端行为，未触发新增 E2E；将通过 Go 编译门禁、`openspec validate`、静态检索和`git diff --check`验证。

### FB-35 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/plugincap ./internal/service/plugin/internal/capabilityowner ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin -count=1`通过。
- `cd apps/lina-core && go test ./pkg/plugin/capability/... ./internal/service/plugin/internal/statecompat ./internal/service/plugin/internal/capabilityowner ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/testutil ./internal/service/plugin -count=1`通过。
- `cd apps/lina-core && go test ./pkg/plugin/pluginhost ./pkg/plugin/pluginbridge ./pkg/plugin/pluginbridge/internal/domainhostcall -count=1`通过。
- `cd apps/lina-core && go test ./internal/cmd ./internal/cmd/internal/httpstartup -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check -- apps/lina-core/pkg/plugin/capability/plugincap apps/lina-core/internal/service/plugin/internal/statecompat apps/lina-core/internal/service/plugin/plugin.go apps/lina-core/internal/service/plugin/plugin_host_services.go apps/lina-core/internal/service/plugin/plugin_host_services_test.go apps/lina-core/internal/service/plugin/internal/capabilityhost/capabilityhost.go apps/lina-core/internal/service/plugin/internal/capabilityowner/capabilityowner_plugin.go openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `rg -n "StateCompatService|NewStateCompat|plugincap\\.StateCompatService|plugincap\\.NewStateCompat" apps/lina-core -g "*.go"`无匹配，确认`lina-core` Go 源码不再引用公开兼容契约。

### FB-26 实施记录

- 删除`plugincap.CapabilityService.BatchGetStatus`、`CapabilityKey`、`CapabilityKeyOrg`、`CapabilityKeyTenant`、`CapabilityKeyAIText`和`MaxCapabilityStatusBatchSize`，`Plugins().Capability()`仅保留插件启用/provider 状态查询。
- 删除插件 owner 中对`org`、`tenant`和`AI`状态的聚合读取，`Plugins.Service`不再接收这些领域 owner 依赖；框架领域的可用性与诊断状态改由各自领域 owner 暴露。
- 删除动态`plugins.capability.status.batch_get`协议常量、catalog 条目、registry 注册、WASM dispatcher 分支、guest client 方法和 README 声明。
- 按既有`FB-20`批量可见性签名要求，同步清理验证中暴露的`tenant.directory.visible.ensure`单个动态方法残留，保留`tenant.directory.visible.ensure_many`作为唯一租户可见性动态入口。
- 同步`linapro-demo-dynamic/plugin.yaml`中的租户可见性 host service 声明，避免示例动态插件继续声明已删除方法。
- 插件本地规范检查：`linapro-tenant-core`和`linapro-demo-dynamic`根目录均未发现`AGENTS.md`。
- `i18n`影响：未新增或修改运行时用户可见文案、菜单、按钮、API 文档源文本、插件清单或语言包资源；仅更新 README 和 OpenSpec 文档。
- 缓存一致性影响：删除只读聚合状态方法，未新增缓存后端、缓存写入或失效路径；插件启用/provider 状态查询继续复用既有状态服务。
- 数据权限影响：移除跨领域聚合读取面，不新增数据读取、写入、存在性探测或租户/组织边界暴露；`org`、`tenant`、`AI`状态仍由对应领域 owner 自己治理。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本或`linactl`入口。
- DI 来源检查：未新增运行期依赖；`capabilityowner.NewCapabilityAdapter`删除`org`和`AI`依赖，仅保留插件配置 factory、插件状态服务和租户服务，启动装配路径同步减少依赖。

### FB-26 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/plugincap ./pkg/plugin/pluginbridge ./pkg/plugin/pluginbridge/internal/domainhostcall ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/pluginbridge/protocol/hostservices ./internal/service/plugin/internal/capabilityowner ./internal/service/plugin/internal/wasm -count=1`通过。
- `cd apps/lina-core && go test ./internal/cmd -count=1`通过。
- `cd apps/lina-plugins/linapro-tenant-core && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/... -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check`通过。
- `git -C apps/lina-plugins diff --check`通过。
- `rg -n '\bCapabilityKey\b|CapabilityKeyOrg|CapabilityKeyTenant|CapabilityKeyAIText|MaxCapabilityStatusBatchSize|Plugins\(\)\.Capability\(\)\.BatchGetStatus|plugins\.capability\.status\.batch_get|HostServiceMethodPluginsCapabilityBatchGetStatus|capabilityKeys|capability-status' apps/lina-core apps/lina-plugins openspec/changes/standardize-plugin-domain-services --glob '*.go' --glob '*.md' --glob '*.yaml' --glob '*.yml' --glob '!**/public/**'`仅命中本任务记录，确认旧插件聚合状态契约已清理。
- `rg -n 'tenant\.directory\.visible\.ensure(\b|$)|HostServiceMethodTenantEnsureVisible' apps/lina-core apps/lina-plugins openspec/changes/standardize-plugin-domain-services --glob '*.go' --glob '*.md' --glob '*.yaml' --glob '*.yml' --glob '!**/public/**'`仅命中本任务记录，确认单个租户可见性动态方法残留已清理。

### FB-27 实施记录

- 将`plugincap.Service`扩展为领域事实 owner，新增`Lifecycle()`子能力，并将插件启用状态读取统一到`Capability()`；`pluginhost.Services.PluginLifecycle()`保留为源码插件兼容快捷入口并委托到`Services.Plugins().Lifecycle()`，`PluginState()`保留为源码插件 bool 风格兼容适配器。
- 将`tenantcap.Service`扩展为领域事实 owner，新增`Plugins()`和`Filter()`子能力；`pluginhost.Services.TenantPluginGovernance()`与`TenantFilter()`在本轮曾作为源码插件兼容快捷入口保留并委托到`Services.Tenant()`。后续`FB-36`已删除这两个顶层快捷入口，最终状态以`FB-36`为准。
- 在`capabilityowner`、`capabilityhost`、`tenantspi`和测试替身中补齐领域子能力装配，插件生命周期继续复用启动期共享`pluginLifecycleRunner`，租户插件治理继续复用既有插件领域与租户 owner 路径。
- 发布动态治理方法到 registry、protocol、catalog、guest client 和 WASM dispatcher：`plugins.capability.enabled.check`、`plugins.capability.provider_enabled.check`、`plugins.capability.enabled_authoritative.check`、`plugins.lifecycle.tenant_plugin_disable.ensure`、`plugins.lifecycle.tenant_plugin_disabled.notify`、`plugins.lifecycle.tenant_delete.ensure`、`plugins.lifecycle.tenant_deleted.notify`、`tenant.plugins.enabled.set`、`tenant.plugins.defaults.provision`和`tenant.filter.context`。
- 明确`TenantFilter.Apply(*gdb.Model)`仅适用于源码插件同进程查询构造；动态插件只暴露可序列化`tenant.filter.context`，实际租户隔离必须通过`RecordStore`、领域 host service 参数或 owner 侧过滤完成。
- 更新`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`，将旧“源码插件专用顶层治理能力”口径改为领域内聚、顶层兼容转发和动态 registry 授权口径，并同步 generated host-service 表。
- 补充动态 WASM dispatcher 单元测试，覆盖`Plugins().Capability()`、`Plugins().Lifecycle()`、`Tenant().Plugins()`和`Tenant().Filter().Context()`调用路径；补齐`linapro-demo-dynamic`测试替身的`Tenant().Plugins()`/`Filter()`接口。
- 插件本地规范检查：本轮修改`apps/lina-plugins/linapro-demo-dynamic`前确认插件根目录不存在`AGENTS.md`。
- `i18n`影响：未新增运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源；仅更新 README 和 OpenSpec 文档，确认无运行时`i18n`资源变更。
- 缓存一致性影响：未新增缓存后端或独立本地缓存；插件状态、生命周期和租户插件启停继续通过既有 owner、启用状态快照和缓存失效路径执行。
- 数据权限影响：动态插件只能在方法已注册、`plugin.yaml hostServices`已声明、宿主已授权、运行时`service + method + resource`校验通过后进入领域 owner；租户插件写入和默认供给继续由领域 owner 按当前租户、状态机和授权边界治理。`tenant.filter.context`只返回当前 actor/租户/平台 bypass 元数据，不暴露数据库模型或跨租户数据。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本或`linactl`入口。
- DI 来源检查：新增领域子能力未引入独立运行期服务图；`PluginLifecycle`由`capabilityhost.New`使用启动期共享`pluginLifecycleRunner`构造后注入`capabilityowner.NewCapabilityAdapter`，插件启用状态读取继续使用启动期共享插件状态服务并通过`Plugins().Capability()`暴露，`PluginState()`仅作为源码插件兼容适配器保留；`Tenant().Plugins()`由`newTenantDomainService`委托既有`pluginDomain`，`Tenant().Filter()`委托启动期传入的`tenantFilterSvc`。`WASM`运行时仍通过启动期共享`capability.Services`目录和`scoping.ServicesForPlugin`获得当前插件作用域。

### FB-27 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/plugincap ./pkg/plugin/capability/tenantcap ./pkg/plugin/capability/tenantcap/tenantspi ./pkg/plugin/pluginbridge ./pkg/plugin/pluginbridge/internal/domainhostcall ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/pluginbridge/protocol/hostservices ./internal/service/plugin/internal/capabilityowner ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm -count=1`通过。
- `cd apps/lina-core && go test ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/integration ./internal/service/plugin ./internal/cmd -count=1`通过。
- `cd apps/lina-plugins/linapro-tenant-core && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/... -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check`通过。
- `git -C apps/lina-plugins diff --check`通过。
- `rg -n 'remain host-owned source-only|source-plugin-only entries such as|动态插件不会接收源码插件专用的插件生命周期|插件生命周期、插件状态和租户插件治理保留为宿主持有的源码插件专用接缝' apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md`无匹配，确认 README 旧口径已清理。

### FB-28 实施记录

- 将`Route`领域原动态路由元数据读取方法重命名为`GetMetadata(ctx)`，并将公开投影类型重命名为`routecap.Metadata`。
- 同步源码插件能力目录、动态 guest client、运行时路由上下文、响应记录、`WASM`host service、测试替身和`linapro-monitor-operlog`审计日志中间件调用点，避免保留旧公开符号或兼容别名。
- 更新主规格和当前变更增量规范中的`Route`领域能力描述，统一为`GetMetadata`/`Metadata`命名。
- `apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`已检查，未包含旧动态路由元数据命名描述，无需同步修改。
- 插件本地规范检查：本轮修改`apps/lina-plugins/linapro-monitor-operlog`前确认插件根目录不存在`AGENTS.md`。
- `i18n`影响：未新增运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源；仅调整 Go 公开契约命名和 OpenSpec 文档，确认无运行时`i18n`资源变更。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径或运行时路由缓存策略；仅重命名现有动态路由元数据读取契约，确认无缓存一致性影响。
- 数据权限影响：不新增数据读取、写入、存在性探测或租户/组织边界；审计日志仍基于当前请求上下文读取已匹配动态路由投影，确认无数据权限边界变化。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本或`linactl`入口。
- DI 来源检查：未新增运行期依赖；`Route`能力仍由既有`capabilityhost`适配器从请求上下文读取动态路由元数据并注入`capability.Services`目录。

### FB-28 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/routecap ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/integration ./internal/service/plugin -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-operlog && GOWORK=off go test ./backend/... -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check -- apps/lina-core apps/lina-plugins openspec/changes/standardize-plugin-domain-services openspec/specs/plugin-host-domain-capabilities/spec.md`通过。
- `rg -n "DynamicRouteMetadata|GetDynamicRouteMetadata|BuildDynamicRouteMetadata" apps/lina-core apps/lina-plugins openspec/specs openspec/changes/standardize-plugin-domain-services/specs -g '!**/node_modules/**'`无匹配，确认代码与规范中的旧公开命名已清理。

### FB-29 实施记录

- 从`plugincap.Service`删除`State()`子能力，保留`Capability()`作为插件领域唯一启用状态查询入口，避免`CapabilityService`与`StateService`在同一领域目录下表达重复职责。
- 将动态`pluginbridge.Services.Plugins()`中的`State()`guest client 删除，动态插件统一通过`Plugins().Capability()`和`plugins.capability.*`动态 host service 方法读取启用、provider 和权威启用状态。
- `StateService`保留为源码插件兼容适配器和宿主内部 provider runtime 输入，继续提供 bool 风格、失败即 false 的状态读取语义；`pluginhost.Services.PluginState()`改为返回`capabilityhost`启动期共享的`pluginState`适配器，不再委托`Plugins().State()`。
- 补充修正验证暴露的租户过滤子能力落点：普通`tenantcap.Service.Filter()`只暴露动态插件可用的可序列化`TenantFilterContext`读取，源码插件同进程`gdb.Model`过滤 helper 改由`tenantspi.RuntimeService.PluginTableFilter()`承载，`pluginbridge` guest 侧不再导入`tenantspi`。
- 同步删除`capabilityowner`、`pluginbridge`、`WASM`测试替身、`linapro-tenant-core`测试替身和宿主测试替身中的多余`State()`实现。
- 更新`apps/lina-core/pkg/plugin/README.md`、`README.zh-CN.md`、当前变更设计与增量规范、根`plugin-host-domain-capabilities`规格，明确插件生命周期归属`Plugins().Lifecycle()`，插件启用状态读取归属`Plugins().Capability()`，`PluginState()`仅为源码插件兼容适配器。
- 插件本地规范检查：本轮修改`apps/lina-plugins/linapro-tenant-core`前确认插件根目录不存在`AGENTS.md`。
- `i18n`影响：未新增运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源；仅更新 README 和 OpenSpec 文档，确认无运行时`i18n`资源变更。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径或状态快照刷新策略；启用状态读取继续复用既有插件状态服务、启用状态快照和缓存失效路径。
- 数据权限影响：未新增数据读取、写入、存在性探测或租户/组织边界；动态插件仍需经过`plugins.capability.*`方法的 registry 发布、`plugin.yaml hostServices`声明和运行时授权校验。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本或`linactl`入口。
- DI 来源检查：未新增运行期依赖；`capabilityhost.New`继续使用启动期传入的共享`pluginStateSvc`构造`plugincap.NewState(pluginStateSvc)`，并分别注入`capabilityowner.NewCapabilityAdapter`用于`Plugins().Capability()`和保存在`directory.pluginState`用于源码插件`PluginState()`兼容入口。

### FB-29 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/plugincap ./pkg/plugin/pluginbridge ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/plugin/internal/capabilityowner ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm -count=1`通过。
- `cd apps/lina-core && go test ./pkg/plugin/capability/tenantcap/... ./pkg/plugin/capability/plugincap ./pkg/plugin/pluginbridge ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm -count=1`通过。
- `cd apps/lina-core && go test ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/integration ./internal/service/plugin -count=1`通过。
- `cd apps/lina-core && go test ./internal/cmd -count=1`通过。
- `cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/internal/service/dynamic -count=1`通过。
- `cd apps/lina-plugins/linapro-tenant-core && GOWORK=off go test ./backend/... -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check`通过。
- `git -C apps/lina-plugins diff --check`通过。
- `rg -n 'Plugins\(\)\.State\(|func \(pluginDirectory\) State\(|PluginState\(invoker|pluginStateService|State\(\) StateService|State\(\) plugincap\.StateService|State\(\) capabilityplugincap\.StateService' apps/lina-core apps/lina-plugins openspec/changes/standardize-plugin-domain-services openspec/specs/plugin-host-domain-capabilities/spec.md --glob '*.go' --glob '*.md' --glob '*.yaml' --glob '*.yml' --glob '!**/public/**'`仅命中源码插件`PluginState()`兼容入口和本任务记录，确认`Plugins().State()`重复领域入口已清理。

### FB-30 实施记录

- 根因：`hostconfigcap`作为普通插件可消费能力包，除`Service`、`SysConfigService`和 DTO 契约外，还承载了静态宿主配置读取 adapter、`RawConfigReader`和构造函数，导致公共契约包同时拥有接口和宿主实现，与其他领域 owner 回流到`internal/service`的架构不一致。
- 横向审查：`manifestcap`和`plugincap.ConfigService`的宿主实现已迁移到插件服务内部组件；`authcap`为命名空间聚合；`aicap`、`orgspi`、`tenantspi`和`storagecap`属于 provider/SPI 或 provider 注册接缝；`bizctxcap`的上下文值 helper 可保留为公共契约辅助，但后续复核发现其公共`New(provider)`和`serviceAdapter`仍构成实现回流，已在`FB-32`独立清理。
- 将静态 host config adapter 迁移到`apps/lina-core/internal/service/plugin/internal/hostconfig`，公共`hostconfigcap`只保留插件可见契约；宿主通过`plugin.NewHostConfigService`暴露窄构造入口，HTTP 启动、插件测试夹具和 WASM hostConfig 测试均改为通过宿主内部 adapter 装配。
- 删除`apps/lina-core/pkg/plugin/capability/hostconfigcap/hostconfigcap_imp.go`和对应测试，把原静态读取测试迁移到宿主内部`hostconfig`组件。
- 补充处理横向扫描发现的`tenantcap`边界问题：当时普通`tenantcap.Service.Filter()`只暴露可序列化`TenantFilterContext`读取，`TenantFilterColumn`、`PluginTableFilterService`和`PlatformBypassEvaluator`保留在`tenantspi`边界，源码插件继续通过`pluginhost.Services.TenantFilter()`兼容入口和`tenantspi.RuntimeService.PluginTableFilter()`使用`gdb.Model`过滤 helper。后续`FB-36`先删除`pluginhost.Services.TenantFilter()`顶层入口，本轮`FB-43`进一步将源码插件同进程过滤收敛到显式注入的`tenantspi.PluginTableFilterService`，普通`tenantcap.FilterService`只保留可序列化上下文读取。
- 在`apps/lina-core/pkg/plugin/capability/capability_test.go`补充普通 capability 实现回流扫描：`hostconfigcap`、`bizctxcap`和`manifestcap`保持纯契约文件，普通 consumer capability 下的实现型文件名会被阻断；provider SPI、AI provider/fallback 和`storagecap`provider 注册作为明确例外记录。
- `apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`已检查，已有普通消费契约不得包含宿主私有实现状态、宿主实现位于`internal/service/plugin`的说明，无需同步修改。
- 插件本地规范检查：本轮修改`apps/lina-plugins/linapro-demo-dynamic`前补查插件根目录不存在`AGENTS.md`。
- `i18n`影响：未新增运行时用户可见文案、菜单、按钮、API 文档源文本、插件清单或语言包资源，仅迁移 Go adapter 和测试，确认无`i18n`资源影响。
- 缓存一致性影响：不修改静态配置读取语义、不新增缓存后端、缓存写入或失效路径；`SysConfig()`缓存与共享修订号语义不变。
- 数据权限影响：不新增数据读取、写入、存在性探测或租户/组织边界；动态 hostConfig key 授权和`SysConfig()`key 级治理语义保持不变。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本或`linactl`入口。
- DI 来源检查：未新增运行期依赖；`RawConfigReader`仍由启动期共享`config.Service`提供，HTTP 启动显式断言后传入`plugin.NewHostConfigService`，再进入插件服务、`capabilityhost`和 WASM runtime，不在请求路径临时创建独立服务图。
- 审查补充：`capabilityhost.New`对`hostConfigSvc`缺失执行 fail fast，并补充测试夹具显式传入 hostConfig 替身，避免源码插件 host service 构造在缺少静态配置适配器时静默降级；同步修正 WASM host-service 测试替身的插件状态字段命名残留。
- 测试策略：本次属于后端 Go、插件宿主服务适配器和治理边界变更；使用包级`go test`、宿主启动装配测试、静态检索、`git diff --check`和`openspec validate`作为验证。未涉及用户可观察前端行为，未触发新增 E2E。

### FB-30 验证

- `cd apps/lina-core && go test ./internal/service/plugin/internal/hostconfig ./pkg/plugin/capability -count=1`通过。
- `cd apps/lina-core && go test ./internal/service/plugin/internal/capabilityhost -count=1`通过。
- `cd apps/lina-core && go test ./pkg/plugin/capability/tenantcap ./pkg/plugin/capability/tenantcap/tenantspi ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/testutil ./internal/service/plugin ./internal/cmd/internal/httpstartup -count=1`通过。
- `cd apps/lina-core && go test ./internal/cmd -count=1`通过。
- `cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/internal/service/dynamic -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check -- apps/lina-core/internal/service/plugin/internal/hostconfig apps/lina-core/internal/service/plugin/plugin_host_services.go apps/lina-core/internal/cmd/internal/httpstartup apps/lina-core/internal/service/plugin/plugin_test.go apps/lina-core/internal/service/plugin/plugin_startup_consistency_test.go apps/lina-core/internal/service/plugin/internal/testutil/testutil_services.go apps/lina-core/internal/service/plugin/internal/wasm apps/lina-core/internal/service/plugin/internal/capabilityhost apps/lina-core/pkg/plugin/capability/capability_test.go apps/lina-core/pkg/plugin/capability/hostconfigcap apps/lina-core/pkg/plugin/capability/tenantcap apps/lina-core/pkg/plugin/pluginbridge/internal/domainhostcall/domainhostcall_tenant.go apps/lina-plugins/linapro-demo-dynamic/backend/internal/service/dynamic/dynamic_host_call_demo_test.go openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `rg -n "hostconfigcap\\.New\\(|RawConfigReader|serviceAdapter|staticConfigAdapter|NewStaticCapabilityAdapter" apps/lina-core/pkg/plugin/capability/hostconfigcap -g'*.go'`无匹配，确认`hostconfigcap`不再包含实现入口。
- `rg -n "github.com/gogf/gf/v2/database/gdb|\\*gdb\\.Model|PluginTableFilterService|TenantFilterColumn" apps/lina-core/pkg/plugin/capability/tenantcap/tenantcap.go`无匹配，确认普通`tenantcap`不再暴露 GoFrame 查询构造器。
- `rg -n "hostconfigcap\\.New\\(|capabilityhostconfig\\.New\\(|hostconfigcap\\.RawConfigReader|tenantcap\\.PluginTableFilterService|tenantcap\\.TenantFilterColumn" apps/lina-core apps/lina-plugins -g'*.go'`无匹配，确认旧构造入口和普通 tenant 过滤类型引用已清理。

### FB-31 实施记录

- 根因：此前为避免和旧源码插件 bool 风格`plugincap.StateService`重名，将普通插件启用状态领域能力临时命名为`CapabilityService`，导致`Plugins().Capability()`与插件状态治理语义不一致，也让动态插件治理方法落在`plugins.capability.*`命名空间。
- 将普通插件领域状态能力统一为`plugincap.StateService`和`Plugins().State()`，保留`IsEnabled`、`IsProviderEnabled`和`IsEnabledAuthoritative`三类带错误返回的状态读取方法。
- 曾将旧源码插件 bool 风格状态读取临时拆成兼容接口；后续`FB-35`已删除该公开兼容契约，并将 bool 风格状态读取下沉为宿主内部`statecompat.Service`。
- 将动态插件状态治理方法从`plugins.capability.enabled.check`、`plugins.capability.provider_enabled.check`和`plugins.capability.enabled_authoritative.check`统一改为`plugins.state.enabled.check`、`plugins.state.provider_enabled.check`和`plugins.state.enabled_authoritative.check`，并同步 registry、protocol、catalog、guest client、WASM dispatcher、测试替身、README 和 OpenSpec 规范文本。
- 恢复机械重命名误改的测试能力目录局部变量，避免把通用`capability.Services`误命名为`stateSvc`。
- 插件本地规范检查：本轮修改`apps/lina-plugins/linapro-tenant-core`和`apps/lina-plugins/linapro-demo-dynamic`相关引用前确认插件根目录不存在`AGENTS.md`。
- `i18n`影响：未新增运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源；仅调整 Go 公开契约命名、动态 host service 方法名和 README/OpenSpec 文档，确认无运行时`i18n`资源变更。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径或插件状态快照刷新策略；启用状态读取继续复用既有共享插件状态服务。
- 数据权限影响：未新增数据读取、写入、存在性探测或租户/组织边界；动态插件仍必须通过`plugins.state.*`方法的 registry 发布、`plugin.yaml hostServices`声明、安装或启用授权和运行时授权快照校验。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本或`linactl`入口。
- DI 来源检查：未新增运行期依赖；`capabilityhost.New`继续使用启动期传入的共享`pluginStateSvc`注入`capabilityowner.NewCapabilityAdapter`用于`Plugins().State()`，后续`FB-35`已将该 bool 风格读取接口收敛为宿主内部`statecompat.Service`。
- 测试策略：本次属于后端 Go 公开契约、动态插件 host service 协议和文档治理变更；使用相关 Go 包测试、宿主启动装配测试、插件子模块测试、静态检索、`git diff --check`和`openspec validate`验证。未涉及用户可观察前端行为，未触发新增 E2E。

### FB-31 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/plugincap ./pkg/plugin/pluginbridge ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/plugin/internal/capabilityowner ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm -count=1`通过。
- `cd apps/lina-core && go test ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/integration ./internal/service/plugin -count=1`通过。
- `cd apps/lina-core && go test ./internal/cmd -count=1`通过。
- `cd apps/lina-plugins/linapro-tenant-core && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/... -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check`通过。
- `git -C apps/lina-plugins diff --check`通过。
- `rg -n 'Plugins\(\)\.Capability\(|func \(.*\) Capability\(\) (plugincap|capabilityplugincap)\.StateService|plugincap\.CapabilityService|capabilityplugincap\.CapabilityService|PluginCapability|HostServiceMethodPluginsCapability|plugins\.capability\.enabled|plugins\.capability\.provider|plugins\.capability\.enabled_authoritative' apps/lina-core apps/lina-plugins openspec/changes/standardize-plugin-domain-services openspec/specs/plugin-host-domain-capabilities/spec.md --glob '*.go' --glob '*.md' --glob '*.yaml' --glob '*.yml' --glob '!**/public/**'`仅命中历史`FB-26`、`FB-27`和`FB-29`任务记录以及本次`FB-31`根因说明，确认生产代码和规范性文档已切换到`State()`与`plugins.state.*`。
- `rg -n 'NewState\(|plugincap\.StateService|capabilityplugincap\.StateService' apps/lina-core apps/lina-plugins --glob '*.go'`仅命中普通`Plugins().State()`领域接口、guest client 和测试替身，确认旧`NewState`构造函数已清理。

### FB-32 实施记录

- 根因：`bizctxcap`公共包除`Service`、`CurrentContext`和上下文值 helper 外，还暴露了`New(provider)`、`serviceAdapter`和仅服务该 adapter 的`ContextProvider`，使普通 consumer capability 包同时持有契约与宿主侧服务适配实现，和`hostconfigcap`问题同类。
- 删除`apps/lina-core/pkg/plugin/capability/bizctxcap/bizctxcap_imp.go`和旧实现测试，公共`bizctxcap`只保留插件可见`Service`契约、`CurrentContext`值对象、`WithCurrentContext`和`CurrentFromContext`辅助函数。
- 宿主侧业务上下文适配统一归属`apps/lina-core/internal/service/plugin/internal/capabilityhost`中的`bizCtxAdapter`；测试和插件调用点不再使用`bizctxcap.New(nil)`，改为局部测试替身或既有`tenantGuardBizCtx`。
- 在`apps/lina-core/pkg/plugin/capability/capability_test.go`将`bizctxcap`纳入 contract-only 扫描，只允许生产文件`bizctxcap.go`存在，避免后续重新引入公共实现文件。
- 插件本地规范检查：本轮修改`apps/lina-plugins/linapro-tenant-core`前确认插件根目录不存在`AGENTS.md`。
- `apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`已检查，未描述`bizctxcap.New`、`ContextProvider`或公共实现归属，无需同步修改。
- `i18n`影响：未新增运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源，确认无`i18n`资源影响。
- 缓存一致性影响：不新增缓存后端、缓存写入、失效路径或权限快照刷新策略；`CurrentContext`只承载请求上下文快照，确认无缓存一致性影响。
- 数据权限影响：不新增数据读取、写入、存在性探测或租户/组织边界；权限、数据权限和租户字段仍由宿主业务上下文写入`ctx`并由领域 owner 读取，确认无新增数据权限影响。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本或`linactl`入口。
- DI 来源检查：未新增运行期依赖；`capabilityhost.New`继续使用启动期共享`bizctx.Service`构造内部`bizCtxAdapter`并注入`capability.Services`，未在请求路径临时创建独立服务图。
- 测试策略：本次属于后端 Go 公共契约边界和治理扫描变更；使用变更包 Go 测试、插件子模块定向测试、静态检索、`git diff --check`和`openspec validate`验证。未涉及用户可观察前端行为，未触发新增 E2E。

### FB-32 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/bizctxcap ./pkg/plugin/capability ./pkg/plugin/capability/tenantcap/tenantspi -count=1`通过。
- `cd apps/lina-plugins/linapro-tenant-core && GOWORK=off go test ./backend/internal/service/tenant -count=1`通过。
- `cd apps/lina-core && go test ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm ./internal/service/plugin ./internal/cmd/internal/httpstartup -count=1`通过。
- `cd apps/lina-core && go test ./internal/cmd -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check -- apps/lina-core/pkg/plugin/capability/bizctxcap apps/lina-core/pkg/plugin/capability/capability_test.go apps/lina-core/pkg/plugin/capability/tenantcap/tenantspi/tenantspi_plugin_filter_test.go openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `git -C apps/lina-plugins diff --check -- linapro-tenant-core/backend/internal/service/tenant/tenant_create_test.go linapro-tenant-core/backend/internal/service/tenant/tenant_delete_test.go`通过。
- `find apps/lina-core/pkg/plugin/capability/bizctxcap -maxdepth 1 -type f -name '*.go' -print | sort`仅列出`bizctxcap.go`和`bizctxcap_test.go`。
- `rg -n 'bizctxcap\.New\(|bizctxcap\.ContextProvider|bizctxcap_imp|serviceAdapter|ContextProvider' apps/lina-core/pkg/plugin/capability/bizctxcap apps/lina-core/pkg/plugin/capability/tenantcap apps/lina-plugins/linapro-tenant-core/backend/internal/service/tenant -g'*.go'`无匹配，确认公共实现入口和相关测试残留已清理。

### FB-33 实施记录

- 根因：`Projection`来自跨模块只读装配和最小字段视图的技术术语，在服务内部或宿主工作台投影缝合点中语义成立；但放到`pkg/plugin/capability`公开主契约时，会让插件开发者直接接触数据库/实现层词汇，领域表达不够自然。
- 将公开 capability 主读模型命名统一收敛为更贴近调用方的领域词汇：`SessionInfo`、`UserOnlineStatus`、`UserInfo`、`PermissionInfo`、`FileInfo`、`DetailInfo`、`MessageInfo`、`TypeInfo`、`ValueInfo`、`LabelInfo`、`SysConfigInfo`、`DeptInfo`、`PostInfo`、`TenantMembershipInfo`和`ProviderInfo`等。
- 同步将公开注释中的“projection”表述改为`info`、`record`、`metadata`、`reference`或`status`等更具体词汇；明显面向 capability SPI 调用方的接口名也同步收敛为`ConsumerReadService`、`WorkspaceViewService`、`DirectoryProvider`、`ContextReader`、`DirectoryReader`和`MembershipReader`等更明确名称。插件管理等宿主内部实现仍可在文件名、局部变量或内部装配路径中保留 projection 语义，用于表达受限的展示投影或批量装配结果。
- 修正`capabilityhost`测试中旧的`PluginLifecycle()`和`PluginState()`断言，改为通过`Plugins().Lifecycle()`和`Plugins().State()`访问插件领域子能力；清理测试工具包中迁移后残留的无用`plugincap`导入。
- 同步更新`apps/lina-core/pkg/plugin/README.md`、`README.zh-CN.md`、`openspec/specs/plugin-host-domain-capabilities/spec.md`和本变更`design.md`/`tasks.md`中的公开类型名与命名取向说明。
- 规则读取与技能：本轮已读取`lina-feedback`、`goframe-v2`、`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/api-contract.md`、`.agents/rules/data-permission.md`、`.agents/rules/i18n.md`和`.agents/rules/testing.md`。
- `i18n`影响：未新增或修改运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源；仅调整 Go 标识符、注释、README 和 OpenSpec 文档，确认无运行时`i18n`资源变更。
- 缓存一致性影响：不新增缓存后端、缓存写入、失效路径、共享修订号或状态快照策略；只读 DTO 命名变更不改变缓存权威源和失效语义。
- 数据权限影响：不新增数据读取、写入、存在性探测或租户/组织边界；原有领域 owner 的可见性校验、批量上限和动态 host service 授权路径保持不变。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本或`linactl`入口。
- DI 来源检查：未新增运行期依赖；本次只调整公开类型名、调用点、测试断言和文档说明，不改变服务构造、启动装配或请求路径依赖图。
- 测试策略：本次属于后端 Go 公开契约命名和治理文档变更，不涉及用户可观察前端行为，未触发新增 E2E；使用相关 Go 包测试、宿主启动装配测试、静态检索、`git diff --check`和`openspec validate`验证。

### FB-33 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/... -count=1`通过。
- `cd apps/lina-core && go test ./pkg/plugin/pluginbridge ./pkg/plugin/pluginbridge/internal/domainhostcall ./pkg/plugin/pluginbridge/internal/hostservice ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/capabilityhost ./internal/service/user/capabilityadapter ./internal/service/dict ./internal/service/file ./internal/service/notify ./internal/service/role -count=1`通过。
- `cd apps/lina-core && go test ./internal/cmd -count=1`通过。
- `cd apps/lina-plugins/linapro-content-notice && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-loginlog && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-operlog && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-org-core && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-tenant-core && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/... -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check`通过。
- 静态检索旧公开`*Projection`类型名、`jobcap`旧读模型类型名和旧 SPI 名称无匹配，确认公开契约、README、OpenSpec 规范和本变更历史记录已清理。
- `rg -n "projection|Projection" apps/lina-core/pkg/plugin/capability -g '*.go' -g '!**/*spi/**' -g '!**/*_test.go'`无匹配，确认非 SPI 公开 capability 生产契约不再使用 projection 术语描述主读模型。

### FB-34 实施记录

- 根因：`plugincap.Service`已经提供`Lifecycle()`和`State()`子领域，`pluginhost.Services`继续暴露`PluginLifecycle()`和`PluginState()`会让源码插件绕过统一插件领域入口，且与当前 OpenSpec 治理能力内聚目标冲突。
- 删除`pluginhost.Services`、`capabilityhost`目录和测试替身上的`PluginLifecycle()`/`PluginState()`顶层方法，`TenantProviderEnv`改为从`Services.Plugins().Lifecycle()`获取生命周期能力。
- 将`linapro-tenant-core`路由装配改为使用`services.Plugins().Lifecycle()`，将`linapro-ops-demo-guard`改为使用`services.Plugins().State()`，并把其窄化状态接口同步为`plugincap.PluginID`和带`error`返回的新领域签名。
- 同步更新`apps/lina-core/pkg/plugin/README.md`、`README.zh-CN.md`、当前变更`proposal.md`、`design.md`、增量规范和根`plugin-host-domain-capabilities`规格，明确`pluginhost.Services`不再暴露这两个顶层方法。
- 插件本地规范检查：`linapro-tenant-core`和`linapro-ops-demo-guard`根目录均未发现`AGENTS.md`。
- 规则读取与技能：本轮已读取`lina-feedback`、`goframe-v2`、`karpathy-guidelines`、`AGENTS.md`、`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`api-contract.md`、`data-permission.md`、`cache-consistency.md`、`i18n.md`和`testing.md`。
- `i18n`影响：未新增或修改运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源；仅调整 Go 调用点、README 和 OpenSpec 文档，确认无运行时`i18n`资源变更。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径、共享修订号或状态快照策略；插件启用状态读取仍复用既有插件状态服务和`Plugins().State()`领域适配器。
- 数据权限影响：未新增数据读取、写入、存在性探测或租户/组织边界；源码插件只是切换到同一插件领域子能力入口，动态 host service 授权和领域 owner 内部数据权限语义不变。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本或`linactl`入口。
- DI 来源检查：未新增运行期依赖；生命周期和状态能力仍由`capabilityhost.New`使用启动期共享`pluginLifecycleRunner`和`pluginStateSvc`构造后注入`capabilityowner.NewCapabilityAdapter`，源码插件通过同一`capability.Services`目录下的`Plugins()`子领域消费。
- 测试策略：本次属于后端 Go 插件宿主服务接口和源码插件调用点变更，不涉及用户可观察前端行为，未触发新增 E2E；使用相关 Go 包测试、源码插件测试、静态检索、`git diff --check`和`openspec validate`验证。

### FB-34 验证

- `cd apps/lina-core && go test ./pkg/plugin/pluginhost ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/integration ./internal/service/plugin -count=1`通过。
- `cd apps/lina-core && go test ./internal/cmd -count=1`通过。
- `cd apps/lina-plugins/linapro-tenant-core && GOWORK=off go test ./backend/... -count=1`通过。
- 使用临时 workspace 包含`apps/lina-core`和`apps/lina-plugins/linapro-ops-demo-guard`后运行`go test .../linapro-ops-demo-guard/backend/... -count=1`通过；该插件未加入仓库根`go.work`且`GOWORK=off`无法解析本地`lina-core`模块。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check`覆盖本次`lina-core`变更文件通过。
- `git -C apps/lina-plugins diff --check -- linapro-tenant-core/backend/plugin.go linapro-ops-demo-guard/backend/plugin.go linapro-ops-demo-guard/backend/internal/service/middleware/middleware.go linapro-ops-demo-guard/backend/internal/service/middleware/middleware_guard.go linapro-ops-demo-guard/backend/internal/service/middleware/middleware_guard_test.go`通过。
- `rg -n "\b(services|sourceServices|hostServices|registrar\.Services\(\))\.PluginLifecycle\(|\b(services|sourceServices|hostServices|registrar\.Services\(\))\.PluginState\(" apps/lina-core apps/lina-plugins --glob '*.go' --glob '!**/public/**'`无匹配，确认旧顶层调用已清理。
- `rg -n "PluginLifecycle\(\) plugincap|PluginState\(\) plugincap|PluginLifecycle\(\) capabilityplugincap|PluginState\(\) capabilityplugincap|func \([^)]*\) PluginLifecycle|func \([^)]*\) PluginState" apps/lina-core apps/lina-plugins --glob '*.go' --glob '!**/public/**'`无匹配，确认旧顶层方法定义已清理。

## 实施记录

### 规则读取与影响判断

- 已读取规则：`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/api-contract.md`、`.agents/rules/data-permission.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/i18n.md`、`.agents/rules/testing.md`、`.agents/rules/database.md`和`.agents/instructions/markdown-format.instructions.md`。
- `i18n`影响：本次未新增运行时 UI 文案、菜单、按钮或 API 文档源文本；仅更新`apps/lina-core/pkg/plugin/README.md`、`README.zh-CN.md`以及`linapro-demo-dynamic`示例 README 中插件能力描述，确认无运行时`i18n`资源变更。
- 缓存一致性影响：保留原权限、字典、运行时配置、插件运行时状态等共享修订号写入语义；相关写入迁移到领域 owner 包后仍通过事务后`sys_cache_revision`推进，未新增缓存后端或本地孤立缓存。
- 数据权限影响：用户、会话、文件、任务、通知等读取和写入能力继续在领域 owner 或稳定适配契约内执行租户、数据范围、目标可见性和批量上限校验；动态插件治理方法必须先经动态 host service registry 发布、`plugin.yaml hostServices`声明和运行时授权校验，之后才进入对应领域 owner。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本或`linactl`入口，确认无开发工具跨平台影响。
- 测试策略：本次属于后端 Go、插件宿主服务适配器和动态`WASM host service`变更；采用包级`go test`、宿主启动装配测试、插件子模块测试、`openspec validate`和静态检索作为验证证据。未涉及用户可观察前端行为，未触发新增 E2E。
- DI 来源检查：新增运行期依赖均通过构造函数路径接入，`lina-core`消费者为`capabilityhost`目录和`WASM`运行时状态 handler；owner 分别为`user/capabilityadapter`、`role`、`dict`、`file`、`plugin/internal/hostconfig`、`jobmgmt/capabilityadapter`、`notify`、`plugin/internal/capabilityowner`、`capabilityadapter`和`cachecoord`。创建位置为宿主启动既有`plugin.NewHostServices`调用链，传递路径为启动共享服务实例进入`capabilityhost.New`后装配到`capability.Services`；源码插件新增依赖通过`pluginhost.Services`显式传入，例如`linapro-tenant-core`的`tenant.Service`由插件路由注册期创建并复用同一`tenantplugin.Service`实例。未在请求路径临时`New()`关键服务图，缓存敏感服务仍复用启动期共享实例或共享后端。
- `FB-5`影响判断：仅修改动态 host service catalog 派生的内部可声明方法 lookup 和单元测试，不新增运行期依赖、不修改数据库、SQL、缓存写入、数据读取/写入实现、API DTO、用户可见 UI 或`i18n`资源；预留 capability 不再进入动态插件能力白名单，数据权限和缓存一致性无新增运行时暴露面。
- `FB-6`影响判断：本轮已读取`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`和`backend-go.md`，并使用`goframe-v2`技能；仅重排`pkg/plugin/capability`主契约文件中的接口与支撑类型声明顺序，不修改方法签名、运行期依赖、HTTP API、SQL、缓存写入、数据读取或权限边界。无运行时行为、前端 UI、API 文档源文本、插件清单或语言包资源影响；无开发工具或跨平台脚本影响。
- `FB-7`影响判断：本轮已读取`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`api-contract.md`、`data-permission.md`、`cache-consistency.md`、`i18n.md`和`testing.md`，并使用`goframe-v2`技能；变更将插件领域能力方法上的显式能力上下文参数收敛为`ctx`承载的调用上下文，未新增 HTTP API、DTO、SQL、缓存后端、开发工具脚本、运行时 UI 文案、API 文档源文本、插件清单或语言包资源。数据权限、租户边界和插件授权语义保持不变，仅调整上下文传递载体。
- `FB-8`影响判断：本轮已读取`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`api-contract.md`、`data-permission.md`、`cache-consistency.md`、`i18n.md`和`testing.md`，并使用`goframe-v2`技能；变更删除额外能力调用上下文模型，领域能力仅从标准业务`ctx`读取当前用户、租户、权限、数据权限和系统调用标识。动态插件授权摘要继续由 dispatcher 在进入领域 owner 前校验，不再传入领域方法。未新增 HTTP API、DTO、SQL、缓存后端、开发工具脚本、运行时 UI 文案、API 文档源文本、插件清单或语言包资源；数据权限边界由领域 owner 继续执行，缓存敏感服务仍复用启动期共享实例。
- `FB-9`影响判断：本轮已读取`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`api-contract.md`、`data-permission.md`、`cache-consistency.md`、`i18n.md`、`testing.md`、`database.md`和`.agents/instructions/markdown-format.instructions.md`，并使用`goframe-v2`技能；变更仅调整`hostconfigcap.Service`内部子领域契约、owner adapter、动态不可用替身和测试替身命名，不新增 HTTP API、DTO、SQL、缓存后端、开发工具脚本、运行时 UI 文案、API 文档源文本、插件清单或语言包资源。数据权限仍为`sys_config`key 级可见性校验和租户回退语义；缓存权威源仍为`sys_config`表，写入后继续推进既有共享修订号；DI 来源未新增运行期依赖，`sys_config`adapter 仍由`capabilityhost.New`使用启动期`tenantFilterSvc`构造并注入`HostConfig`能力。
- `FB-10`影响判断：本轮已读取`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`api-contract.md`、`data-permission.md`、`cache-consistency.md`、`i18n.md`和`testing.md`，并使用`goframe-v2`技能；变更只删除`filecap.Service.Download`及其重复 adapter、guest stub 和测试替身，保留`Open`作为唯一受治理文件流读取契约。未新增 HTTP API、DTO、SQL、运行期依赖、缓存后端、开发工具脚本、运行时 UI 文案、API 文档源文本、插件清单或语言包资源。数据权限边界仍由`Open`路径执行`EnsureVisible`和领域 owner 读取校验；动态`host service registry`此前未发布`files.download`，因此无动态授权面收缩之外的运行时影响。
- `FB-11`影响判断：本轮已读取`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`api-contract.md`、`data-permission.md`、`cache-consistency.md`、`i18n.md`、`testing.md`、`database.md`和`.agents/instructions/markdown-format.instructions.md`，并使用`goframe-v2`技能；变更仅补充`HostConfig().SysConfig().Get`单 key 读取方法，生产实现复用`BatchGet`，不新增数据库查询分支、HTTP API、DTO、SQL、缓存后端、开发工具脚本、运行时 UI 文案、API 文档源文本、插件清单或语言包资源。数据权限、租户回退、缺失与不可见统一拒绝语义均沿用`BatchGet`结果；DI 来源无新增运行期依赖。
- `FB-12`影响判断：本轮已读取`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`api-contract.md`、`data-permission.md`、`cache-consistency.md`、`i18n.md`、`testing.md`和`database.md`，并使用`goframe-v2`技能；变更补齐`Files.Upload`和`Files.CreateFromStorage`的文件中心写入能力，不新增 HTTP API、SQL、DAO/DO/Entity、缓存后端、开发工具脚本、运行时 UI 文案、API 文档源文本、插件清单或语言包资源。数据权限影响为新增插件通过宿主发布服务创建`sys_file`记录，创建路径使用标准业务`ctx`中的租户与上传人；本轮同步让`datascope.CurrentTenantID`识别动态 host call 写入的`bizctxcap.CurrentContext`，避免动态插件文件写入落到平台租户或上传人为 0。`CreateFromStorage`只从当前插件作用域`Storage()`复制对象，不暴露 provider key、本地路径或跨插件对象；动态`files.create_from_storage`额外要求源路径已声明`storage.get`授权，避免绕过动态 storage 资源边界；动态直传设置`MaxDirectUploadBytes`上限，大文件仍通过`Storage.Put`分片后调用`CreateFromStorage`。缓存一致性无影响；数据库结构无变化；DI 来源未新增启动期运行依赖，`scopedDirectory.Files()`仅把既有 scoped `Storage()`作为 adapter 运行期输入传入。
- `FB-13`影响判断：本轮已读取`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`api-contract.md`、`data-permission.md`、`cache-consistency.md`、`i18n.md`和`testing.md`，并使用`goframe-v2`技能；变更不新增 HTTP API、SQL、DAO/DO/Entity、缓存后端、开发工具脚本、运行时 UI 文案、API 文档源文本、插件清单或语言包资源。数据权限影响为`Jobs`运行期读取在数据库查询阶段新增`datascope.Service.ApplyUserScopeWithBypass`过滤，写入、删除、状态变更和执行动作改为委托`jobmgmt` owner 执行目标可见性、状态机、审计和 scheduler 副作用；`jobs.register`仅保留在动态插件声明期`pluginbridge.Declarations.Jobs().Register(...)`和 WASM Jobs discovery 路径，不进入运行期`jobcap.Service`。缓存一致性无新增影响；开发工具跨平台无影响；DI 来源为 HTTP 启动期已有共享`jobMgmtSvc`和`scopeSvc`，通过`plugin.NewHostServices`显式传递到`capabilityhost.New`和`jobmgmt/capabilityadapter`，未在请求路径临时`New()`关键服务图。
- `FB-14`影响判断：本轮已读取`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`api-contract.md`、`data-permission.md`、`cache-consistency.md`、`i18n.md`和`testing.md`，并使用`goframe-v2`技能；变更只调整 Job status 的 Go 类型来源，`jobcap.Service`、`jobmgmt` owner 契约、cron/jobmgmt 内部状态参数和测试替身统一复用`api/job/v1.Status`。不新增 HTTP API、路由、SQL、DAO/DO/Entity、运行期依赖、缓存后端、开发工具脚本、运行时 UI 文案、API 文档源文本、插件清单或语言包资源。数据权限、租户边界、状态机、scheduler 副作用和动态 host service 授权路径不变；缓存一致性无影响；开发工具跨平台无影响；未新增 DI 依赖。
- `FB-15`影响判断：本轮已读取`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`i18n.md`和`testing.md`，并使用`goframe-v2`技能；变更仅移动`Manifest`宿主 factory、嵌入资源 resolver 和读取实现的包归属，`manifestcap`继续只暴露插件可见`Service`、DTO 与资源大小上限常量。不新增 HTTP API、路由、SQL、DAO/DO/Entity、缓存后端、开发工具脚本、运行时 UI 文案、API 文档源文本、插件清单或语言包资源；数据权限和租户边界无新增影响，Manifest 读取仍受当前插件作用域、路径归一化、单资源大小和批量总字节数限制约束；`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`已检查，未包含旧 factory 或实现归属描述，无需同步修改。DI 来源：`manifestresource.Factory`由插件服务内部 owner 创建，`plugin.New`为动态`WASM`运行时创建默认 factory，`capabilityhost.New`为源码插件能力目录创建带`sourcePluginEmbeddedFiles`resolver 的 factory，WASM artifact 资源通过既有执行上下文`WithArtifactResources`绑定；无请求路径临时创建关键服务图，无缓存一致性、开发工具跨平台或 E2E 影响。
- `FB-16`影响判断：本轮已读取`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`data-permission.md`、`cache-consistency.md`、`i18n.md`、`testing.md`和`database.md`，并使用`goframe-v2`技能；变更仅将`HostConfig().SysConfig()`宿主适配实现从开放独立`internal/service/runtimeconfig`包迁移到插件服务内部`plugin/internal/hostconfig`组件，不修改`hostconfigcap.Service`、动态 host service wire 协议、HTTP API、路由、SQL、DAO/DO/Entity、缓存后端、开发工具脚本、运行时 UI 文案、API 文档源文本、插件清单或语言包资源。数据权限仍为`sys_config`key 级可见性校验和租户回退语义；缓存权威源仍为`sys_config`表，写入后继续推进既有共享修订号；DI 来源无新增运行期依赖，`capabilityhost.New`继续使用启动期`tenantFilterSvc`构造并注入`HostConfig`能力；开发工具跨平台和 E2E 无影响。
- `FB-18`影响判断：本轮已读取`AGENTS.md`、`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`和`backend-go.md`，并使用`lina-feedback`、`goframe-v2`和`lina-review`技能；变更只删除`authz`、`filecap`、`notifycap`、`sessioncap`、`dictcap`、`hostconfigcap`、`plugincap`、`usercap`和`jobcap`中未实现、未引用的`ScopeService`导出接口，其中`authz.EnsurePermissionsVisible`也未被生产实现、动态 registry、guest client 或内部注入点消费。不新增 HTTP API、DTO、SQL、DAO/DO/Entity、运行期依赖、缓存后端、开发工具脚本、运行时 UI 文案、API 文档源文本、插件清单或语言包资源；数据权限运行逻辑、租户边界、动态 host service 授权和缓存一致性无变化；DI 来源无新增运行期依赖。测试策略为包级 Go 编译测试、静态检索和`openspec validate`，未涉及用户可观察前端行为，未触发 E2E。

### 实现摘要

- 删除插件可见`AdminServices`、领域`AdminService`和`pluginhost.Services.Admin()`入口，将读取、写入、状态变更和治理方法收敛到各领域统一`Service`。
- `pluginhost.Services`已删除`PluginLifecycle()`和`PluginState()`顶层入口；源码插件通过`Services.Plugins().Lifecycle()`和`Services.Plugins().State()`访问插件生命周期与状态子能力。`TenantPluginGovernance()`和`TenantFilter()`仍作为源码插件租户快捷入口保留，并委托到`Tenant()`领域子能力。
- 动态`pluginbridge.Services.Plugins()`和`Tenant()`通过领域能力目录访问普通能力与治理能力；只有已发布到动态 host service registry、已在`plugin.yaml hostServices`声明并通过授权校验的方法才会进入领域 owner。
- 删除`pluginbridge`根包顶层能力 client helper，动态插件统一通过`pluginbridge.Default()`或`pluginbridge.New()`获得`Services`目录后访问`Runtime()`、`Storage()`、`Network()`、`Cache()`、`Lock()`、`HostConfig()`、`Manifest()`和`RecordStore()`；新增包测试阻断根包 facade 回流。
- 将原集中在`internal/service/plugin/internal/capabilityhost`的 DAO-backed 领域实现迁移到 owner 包或 owner-owned adapter 包：`user/capabilityadapter`、`role`、`dict`、`file`、`plugin/internal/hostconfig`、`jobmgmt/capabilityadapter`、`notify`和`plugin/internal/capabilityowner`；纯适配工具归属`capabilityadapter`，事务内缓存修订写入归属`cachecoord`。`capabilityhost`保留目录装配、标准业务上下文桥接和源码插件适配职责；`WASM`host state handler 保留编解码和错误映射，状态持久化委托给插件 owner store。
- 同步更新`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`，删除`AdminService`、`Services.Admin()`和动态插件生命周期暴露描述。
- 审查补充修复`linapro-content-notice`测试替身，补齐统一`usercap.Service`新增的`SetStatus`方法，避免插件测试继续停留在旧窄接口形态。
- 审查补充修复`linapro-demo-dynamic`测试替身，补齐统一`hostconfigcap.Service`的运行时配置方法，并将示例 README 改为通过`pluginbridge.Default()`能力目录获取动态宿主能力。
- 审查补充修复`linapro-org-core`测试替身，补齐统一`usercap.Service`新增写方法，确保 provider 构造测试仍通过公开用户能力契约。
- 审查补充修复`linapro-tenant-core`provider 测试构造，显式传入`tenant.Service`并复用同一`tenantplugin.Service`实例。
- 审查补充修复预留动态`host service`可声明性：`secret.resolve`、`event.publish`和`queue.enqueue`继续保留在 public catalog/README 作为未来预留说明，但内部运行时 capability/resource lookup 只接收`Published=true`的方法，`ValidateHostServiceSpecs`不再接受这些预留方法声明，`AllCapabilities`也不再把预留 capability 视为当前可用能力。
- 插件本地规范检查：本次触及`apps/lina-plugins/linapro-content-notice`、`linapro-monitor-online`、`linapro-tenant-core`、`linapro-org-core`、`linapro-ops-demo-guard`和`linapro-demo-dynamic`，未发现插件根目录`AGENTS.md`。
- `FB-6`重排`pkg/plugin/capability`及其主要子能力主契约文件，将`Service`、子资源服务、provider SPI、scope/runtime 等关键接口放在文件前部，支撑 DTO、输入输出、投影、环境结构、工厂和实现结构放在其后；未更改任一接口方法集合或运行时代码路径。
- `FB-7`将领域能力公开方法收敛为只保留`ctx context.Context`和业务输入参数，源码插件和动态插件调用点不再显式传递第二个能力上下文参数。
- `FB-8`删除额外能力调用上下文类型与 helper，标准业务上下文由 HTTP 中间件既有`internal/model.Context`和动态`contextWithHostCallBizContext`提供；`bizctxcap.CurrentContext`补齐权限、数据权限和超级管理员字段，领域 owner adapter 统一从标准业务`ctx`读取当前用户、租户、权限和数据权限信息。
- `FB-8`同步更新官方源码插件调用点、`WASM host service`dispatcher、动态测试替身和 README；通知能力通过插件作用域服务绑定来源插件，不再依赖调用上下文中的插件来源字段；动态授权结果只在 dispatcher 内部校验，不进入领域 owner。
- `FB-9`将`hostconfigcap.Service`中的`sys_config`能力收敛到`SysConfig()`子领域，方法命名改为`BatchGet`、`List`、`SetValue`、`Reset`和`EnsureVisible`；`SysConfigInfo.Value`直接返回原始`sys_config.value`字符串，不再使用 JSON 字节契约。`plugin/internal/hostconfig`领域 adapter 改为`NewSysConfigCapabilityAdapter`，动态`hostconfig`guest 能力保留未发布语义但通过`SysConfig()`子服务返回不支持错误；测试替身和集成测试同步实现新子领域接口。
- `FB-9`同步收紧`SysConfig().List`查询路径，先在数据库侧按`sys_config.key`分组计数和分页，再批量读取当前页 key 的平台与租户行执行回退合并，避免扫描全量配置行后内存分页。
- `FB-9`同步更新`apps/lina-core/pkg/plugin/README.md`、`README.zh-CN.md`、OpenSpec 任务和增量规范，明确插件作用域配置、静态宿主配置和`sys_config`分别由`Config`、`HostConfig`和`HostConfig().SysConfig()`表达。
- `FB-10`删除`filecap.Service.Download`，同步删除`fileCapabilityAdapter.Download`、动态 guest `filesService.Download`和`testNoopFiles.Download`；`Open`继续复用原`open`内部实现完成目标解析、可见性校验和 owner `OpenByID`流读取。`localdocs/plugin-domain-service-unification-design.md`同步将`Files`内容读取方法矩阵收敛为单一`Open`，并说明 HTTP 下载语义由调用方或控制器决定。
- `FB-11`为`hostconfigcap.SysConfigService`补充`Get(ctx, key)`，`plugin/internal/hostconfig`生产实现委托`BatchGet(ctx, []key)`并在缺失或不可见时返回统一`CodeCapabilityDenied`；静态 host config adapter 和动态未发布 adapter 同步返回原有不可用/未支持语义，测试替身通过`BatchGet`返回空投影。
- `FB-12`为`file.Service`新增`CreateFromReader`，将 HTTP 上传与插件上传统一到同一套文件名清洗、大小限制、临时文件流式 hash、重复 hash 复用、文件中心 storage 写入、`sys_file`落库和失败清理逻辑；`Upload`保留为`ghttp.UploadFile`适配入口。
- `FB-12`补齐动态 host call 标准业务上下文对文件写入的归属桥接：`file.Service`取上传人时回退到`bizCtxSvc.Current(ctx)`，`datascope.CurrentTenantID`在无`model.Context`时读取`bizctxcap.CurrentContext`中的租户。
- `FB-12`实现`fileCapabilityAdapter.Upload`和`CreateFromStorage`，源码插件的`scopedDirectory.Files()`会注入当前插件作用域`Storage()`；`CreateFromStorage`读取插件私有对象后复制到文件中心 storage，不移动、不删除源对象，也不向插件暴露 provider object key、本地路径或`sys_file`内部实体。
- `FB-12`发布动态`files.upload`和`files.create_from_storage`方法，更新 protocol 常量、public catalog、runtime registry、WASM dispatcher 和 guest client；`files.upload`使用有界 JSON 直传，`files.create_from_storage`用于复用动态`Storage.Put`分片后的对象，并在 dispatcher 中校验源路径具备`storage.get`授权。
- `FB-13`新增`jobmgmt/jobcontract`窄契约包，作为插件能力 adapter 与`jobmgmt` owner 之间的稳定运行期接缝，避免`jobmgmt -> jobhandler -> plugin`与`plugin -> capabilityhost -> jobmgmt`形成 import cycle；`jobmgmt.SaveJobInput`和`UpdateJobInput`保留为类型别名。
- `FB-13`从运行期`jobcap.Service`删除`Register`和`RegisterInput`，同步删除动态普通 Jobs guest client 的 runtime `Register`不可用方法和测试替身实现；动态`jobs.register`声明期 catalog、protocol、WASM discovery handler 和示例插件声明保持不变。
- `FB-13`将`jobCapabilityAdapter.Create`、`Update`、`Delete`、`Run`和`SetStatus`改为委托`jobcontract.Owner`，`Create/Update`映射为运行期 shell 任务输入并使用`jobmgmt`默认校验；`BatchGet/List/EnsureVisible`继续使用批量投影查询，但新增租户过滤后的用户数据范围约束。
- `FB-13`同步更新`apps/lina-core/pkg/plugin/README.md`、`README.zh-CN.md`、`localdocs/plugin-domain-service-unification-design.md`和 OpenSpec 增量规范，明确`jobs.register`属于声明期 facade，运行期`jobcap.Service`不暴露`Register`。
- `FB-14`将`jobcap.JobInfo.Status`、`jobcap.ListInput.Status`、`jobcap.SaveInput.Status`和`jobcap.Service.SetStatus`改为直接使用`api/job/v1.Status`；动态 guest 和 WASM dispatcher 在 JSON wire 边界继续显式转换字符串，保证 wire payload 不变。
- `FB-14`删除`internal/service/jobmeta`中重复的`JobStatus`类型和`JobStatus*`常量，`NormalizeJobStatus`返回`api/job/v1.Status`，并将`Status.IsValid()`移动到 API 枚举类型所在包。
- `FB-14`同步将`jobmgmt`、`cron`、scheduler、`jobcontract`和相关测试替身中的 Job status 参数、字段和常量引用改为`api/job/v1.Status`，避免在内部 service 继续使用重复枚举值。
- `FB-15`新增`internal/service/plugin/internal/manifestresource`内部组件，承载`Factory`、`EmbeddedFilesResolver`、artifact 资源绑定和 manifest 读取、批量读取、列表、存在性与 YAML scan 实现；删除`pkg/plugin/capability/manifestcap`中的宿主 factory 和 adapter 实现，使该公开包只保留插件可见契约。
- `FB-15`同步调整`capabilityhost`、`WASM`host service、`plugin.New`、HTTP 启动测试和插件测试工具包的装配路径：`capabilityhost`只持有内部 factory 并返回插件作用域`manifestcap.Service`，`httpstartup`不再直接构造 manifest factory，动态运行时通过插件服务内部 factory 获得 manifest 资源视图。
- `FB-16`新增`internal/service/plugin/internal/hostconfig`内部组件，将组件入口、静态 HostConfig wrapper 和`sys_config`投影读取、写入、重置、可见性校验实现分文件维护；删除开放独立`internal/service/runtimeconfig`包，使该能力实现只在插件服务内部授权边界内被`capabilityhost`装配。
- `FB-16`验证补充修复当前工作区已改用`api/plugin/v1`枚举的插件内部文件缺失`pluginv1`导入问题，仅补齐 import 以恢复`capabilityhost`和`plugin`包编译，不修改业务逻辑。
- `FB-17`删除独立`internal/service/capabilitydomain`和临时`internal/service/capabilityadapter`组件，将可稳定复用的 `PageRequest.Normalize`、ID 解析和`ParseTenantID`收回到`pkg/plugin/capability/capmodel`与`pkg/plugin/capability/tenantcap`，文件专属投影辅助函数保留在`internal/service/file/file_capability.go`。原记录中的事务绑定缓存修订 helper 已在`FB-54`删除，相关写路径改为复用启动期注入的统一`cachecoord.Service`；本次不新增新的耦合中间层，不改变运行期依赖来源。
- `FB-18`删除`authz`、`filecap`、`notifycap`、`sessioncap`、`dictcap`、`hostconfigcap`、`plugincap`、`usercap`和`jobcap`中未被实现或引用的领域`ScopeService`空壳接口，同时保留`orgcap/orgspi.ScopeService`和`tenantcap/tenantspi.ScopeService`这两个具备真实实现、编译断言和宿主内部消费面的 SPI。
- `FB-19`补齐`Notifications`动态插件接口实现，新增`messages.list`、`messages.delete`、`messages.by_source.delete`、`messages.mark_read`和`messages.mark_unread`动态 host-service 方法；同步更新 protocol 常量、public catalog、runtime registry、WASM dispatcher、guest client、README 和中英文 generated host-service 表。
- `FB-19`影响判断：本轮已读取`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`api-contract.md`、`data-permission.md`、`i18n.md`和`testing.md`，并使用`goframe-v2`技能；变更不新增数据库、SQL、DAO、缓存后端、开发工具脚本、前端 UI、菜单、按钮、运行时语言包或 API 文档源文本。数据权限边界保持在启动期共享`notifycap.Service`领域 owner 内执行，动态 dispatcher 仅补齐 registry 授权、编解码和错误映射；`messages.send`继续是唯一需要`resources[].ref`的 notifications 动态方法，新增列表、删除和已读状态方法均按当前 actor、租户和领域可见性治理。
- `FB-21`将`plugincap.ConfigService`和`ConfigServiceFactory`移动到`plugincap.go`主契约入口；将配置实现文件重命名为`config.go`、`config_access.go`和`config_test.go`；将生命周期接口、构造和委托实现合并到`lifecycle.go`；将插件状态接口、构造和启用状态实现合并到`state.go`。
- `FB-21`影响判断：本轮已读取`AGENTS.md`、`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`和`testing.md`，并使用`lina-feedback`、`goframe-v2`和`karpathy-guidelines`技能；变更只调整`apps/lina-core/pkg/plugin/capability/plugincap`包内契约归位、文件命名和实现分布，不修改方法签名、运行期依赖、HTTP API、DTO、SQL、DAO/DO/Entity、缓存写入、数据读取或写入逻辑、插件授权、数据权限、租户边界、运行时 UI 文案、API 文档源文本、插件清单或语言包资源。`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`不含本次文件组织细节，无需同步修改；数据权限、缓存一致性、开发工具跨平台和 E2E 无影响。测试策略为变更包`go test`、静态检索和`openspec validate`。
- `FB-22`将上轮聚合后的通用文件名修正为`plugincap_config.go`、`plugincap_config_access.go`、`plugincap_config_test.go`、`plugincap_lifecycle.go`和`plugincap_state.go`，继续保留`ConfigService`位于`plugincap.go`以及配置、生命周期、状态三类高内聚文件合并结果。
- `FB-22`影响判断：本轮已读取`AGENTS.md`、`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`和`testing.md`，并使用`lina-feedback`、`goframe-v2`和`karpathy-guidelines`技能；变更仅修正源码文件命名和文件头职责描述，不修改接口、构造函数、运行期依赖、HTTP API、DTO、SQL、DAO/DO/Entity、缓存、数据权限、租户边界、运行时 UI 文案、API 文档源文本、插件清单或语言包资源。`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`不含文件名约束，无需同步；开发工具跨平台、缓存一致性、数据权限和 E2E 无影响。测试策略为变更包`go test`、文件列表检查、`git diff --check`和`openspec validate`。
- `FB-24`影响判断：本轮已读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/testing.md`和`.agents/rules/i18n.md`，并使用`lina-feedback`、`goframe-v2`和`karpathy-guidelines`技能；变更仅将插件配置 factory 和读取实现从公开`pkg/plugin/capability/plugincap`边界迁入`apps/lina-core/internal/service/plugin/internal/pluginconfig`，公开包只保留插件可见`ConfigService`契约，同时通过`apps/lina-core/internal/service/plugin/plugin_config.go`提供宿主窄构造入口。未新增 HTTP API、DTO、SQL、DAO/DO/Entity、缓存后端、开发工具脚本、前端 UI、运行时文案、API 文档源文本、插件清单或语言包资源；`i18n`、缓存一致性和数据权限无新增影响。DI 来源仍为启动期共享的`hostConfigReader`和`pluginConfigFactory`，由`internal/cmd/internal/httpstartup`创建并沿`plugin.NewHostServices`、`plugin.New`和 WASM runtime 传递，未在请求路径临时创建独立服务图。插件子模块测试仅使用本地 `plugincap.ConfigService` 替身，不依赖宿主 internal 包。测试策略为宿主插件相关包`go test`、插件子模块全量回归、`openspec validate`和静态检索。

### FB-25 验证

- `cd apps/lina-core && go test ./internal/service/plugin/internal/scoping ./pkg/plugin/capability ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/testutil ./internal/service/plugin ./internal/service/plugin/internal/runtime -count=1`通过；`internal/service/plugin/internal/runtime`的定向 smoke `TestRunReconcilerTickSafelyRecoversPanic` 通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check`通过。
- `rg -n "capability\.ScopedServicesFactory|capability\.ServicesForPlugin" apps/lina-core -g '*.go'`无匹配，确认旧公开绑定辅助已删除。

### 验证记录

- `cd apps/lina-core && go test ./pkg/plugin/capability ./pkg/plugin/pluginhost ./pkg/plugin/pluginbridge ./pkg/plugin/pluginbridge/internal/hostservice ./internal/service/capabilityadapter ./internal/service/cachecoord ./internal/service/user/capabilityadapter ./internal/service/role ./internal/service/dict ./internal/service/file ./internal/service/plugin/internal/hostconfig ./internal/service/jobmgmt/capabilityadapter ./internal/service/notify ./internal/service/plugin/internal/capabilityowner ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin ./internal/cmd -count=1`
- `cd apps/lina-plugins/linapro-content-notice && GOWORK=off go test ./backend/... -count=1`
- `cd apps/lina-plugins/linapro-monitor-online && GOWORK=off go test ./backend/... -count=1`
- `cd apps/lina-plugins/linapro-tenant-core && GOWORK=off go test ./backend/... -count=1`
- `cd apps/lina-plugins/linapro-org-core && GOWORK=off go test ./backend/... -count=1`
- `cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/... -count=1`
- `linapro-ops-demo-guard`未配置独立`lina-core`本地`replace`且根`go.work`未包含该插件，使用临时`go.work`同时纳入`apps/lina-core`和`apps/lina-plugins/linapro-ops-demo-guard`后运行`go test ./backend/... -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`
- `git diff --check`
- `rg -n "\bAdminServices\b|\bAdminService\b|Services\.Admin\(|\.Admin\(\)" apps/lina-core apps/lina-plugins -g '*.go' -g '*.md'`，无匹配。
- `rg -n "\bAdminServices\b|\bAdminService\b|Services\.Admin\(|\.Admin\(\)|HostServiceMethodTenantBatchListUserTenants|HostServiceMethodTenantValidateSwitch|tenant\.membership\.batch_list_by_users|tenant\.membership\.switch\.validate|users\.tenants\.batch_list|tenants\.switch\.validate|users\.dept_assignments\.list|depts\.search|tenants\.search|tenants\.current\b|users\.tenants\.list|HostServiceMethodOrgSearchDepartments|HostServiceMethodTenantSearch" apps/lina-core apps/lina-plugins -g '*.go' -g '*.md' -g '*.yaml' -g '*.yml'`，无匹配。
- `rg -n "lina-core/internal/(dao|model)" apps/lina-core/internal/service/plugin/internal/capabilityhost apps/lina-core/internal/service/plugin/internal/wasm -g '*.go' -g '!**/*_test.go'`，无匹配。
- `rg -n 'plugins\.lifecycle|plugins\.(install|uninstall|enable|disable|upgrade)|tenant_plugin\.(provision|disable|delete)|tenant\.lifecycle|plugin\.Admin\(\)\.Require|cron\.register|CronHostService' apps/lina-core apps/lina-plugins -g '*.go' -g '*.md' -g '*.yaml' -g '*.yml'`仅命中负向测试、源码插件声明期生命周期、源码插件专用治理能力或`tenant_plugin.provision`历史资源字符串；未发现恢复动态插件生命周期 host service 的生产注册。
- `rg -n "pluginbridge\.(Runtime|Storage|Network|Cache|Lock|HostConfig|Manifest|Data|RecordStore|Cron|HostLog|HostState)\(" apps/lina-core apps/lina-plugins -g '*.go' -g '*.md' -g '*.yaml' -g '*.yml'`，无匹配。
- `cd apps/lina-core && go test ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/pluginbridge/protocol/hostservices -count=1`
- `TestValidateHostServiceSpecsRejectsReservedServices`确认`secret.resolve`、`event.publish`和`queue.enqueue`即使携带`resources`也不能通过动态`hostServices`声明；`TestValidateCapabilitiesRejectsReservedCapabilities`确认`host:secret`、`host:event:publish`和`host:queue:enqueue`不进入当前动态 capability 白名单。
- `cd apps/lina-core && go test ./pkg/plugin/capability/... -count=1`
- `openspec validate standardize-plugin-domain-services --strict`
- `git diff --check -- apps/lina-core/pkg/plugin/capability openspec/changes/standardize-plugin-domain-services/tasks.md`
- `cd apps/lina-plugins && git diff --check`
- `cd apps/lina-core && go test ./pkg/plugin/capability ./pkg/plugin/pluginbridge ./internal/service/user/capabilityadapter ./internal/service/role ./internal/service/notify ./internal/service/dict ./internal/service/file ./internal/service/jobmgmt/capabilityadapter ./internal/service/plugin/internal/hostconfig ./internal/service/plugin/internal/capabilityowner ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm -count=1`
- `cd apps/lina-core && go test ./pkg/plugin/capability -count=1`
- `cd apps/lina-core && go test ./internal/cmd -count=1`
- `cd apps/lina-plugins/linapro-content-notice && GOWORK=off go test ./backend/... -count=1`
- `cd apps/lina-plugins/linapro-monitor-loginlog && GOWORK=off go test ./backend/... -count=1`
- `cd apps/lina-plugins/linapro-monitor-operlog && GOWORK=off go test ./backend/... -count=1`
- `cd apps/lina-plugins/linapro-monitor-online && GOWORK=off go test ./backend/... -count=1`
- `cd apps/lina-plugins/linapro-org-core && GOWORK=off go test ./backend/... -count=1`
- `cd apps/lina-plugins/linapro-tenant-core && GOWORK=off go test ./backend/... -count=1`
- `cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/... -count=1`
- `linapro-ops-demo-guard`使用临时`go.work`同时纳入`apps/lina-core`和`apps/lina-plugins/linapro-ops-demo-guard`后运行`go test ./backend/... -count=1`通过。
- 旧能力调用上下文类型、注入 helper、测试替身字段和动态 dispatcher 旧 helper 的静态检索均无匹配，确认 Go 源码、官方插件和当前变更文档已删除旧上下文方案。
- `rg -n "lina-core/internal/(dao|model)" apps/lina-core/internal/service/plugin/internal/capabilityhost apps/lina-core/internal/service/plugin/internal/wasm -g '*.go' -g '!**/*_test.go'`无匹配。
- `FB-8`验证：旧能力调用上下文、旧注入 helper、旧测试替身字段和旧动态 dispatcher helper 静态扫描无匹配。
- `cd apps/lina-core && go test ./pkg/plugin/capability ./pkg/plugin/pluginbridge ./pkg/plugin/pluginbridge/internal/hostservice ./internal/service/user/capabilityadapter ./internal/service/role ./internal/service/notify ./internal/service/dict ./internal/service/file ./internal/service/jobmgmt/capabilityadapter ./internal/service/plugin/internal/hostconfig ./internal/service/plugin/internal/capabilityowner ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm -count=1`
- `cd apps/lina-core && go test ./internal/cmd -count=1`
- `cd apps/lina-plugins/linapro-content-notice && GOWORK=off go test ./backend/... -count=1`
- `cd apps/lina-plugins/linapro-monitor-loginlog && GOWORK=off go test ./backend/... -count=1`
- `cd apps/lina-plugins/linapro-monitor-operlog && GOWORK=off go test ./backend/... -count=1`
- `cd apps/lina-plugins/linapro-monitor-online && GOWORK=off go test ./backend/... -count=1`
- `cd apps/lina-plugins/linapro-org-core && GOWORK=off go test ./backend/... -count=1`
- `cd apps/lina-plugins/linapro-tenant-core && GOWORK=off go test ./backend/... -count=1`
- `cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/... -count=1`
- `openspec validate standardize-plugin-domain-services --strict`
- `git diff --check`
- `git -C apps/lina-plugins diff --check`
- `FB-10`验证：`cd apps/lina-core && go test ./pkg/plugin/capability ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/file ./internal/service/plugin/internal/testutil -count=1`
- `FB-10`验证：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/pluginbridge/protocol/hostservices -count=1`
- `FB-10`验证：`cd apps/lina-core && go test ./internal/cmd -count=1`
- `FB-10`验证：`openspec validate standardize-plugin-domain-services --strict`
- `FB-10`验证：`git diff --check`
- `FB-10`静态检索：`rg -n -e 'files\\.download' -e 'HostServiceMethodFilesDownload' -e 'Download opens a visible file stream' -e 'Download\\(context\\.Context, capabilityfilecap\\.FileID' -e 'Download\\(context\\.Context, filecap\\.FileID' apps/lina-core/pkg/plugin apps/lina-core/internal/service/file apps/lina-core/internal/service/plugin/internal/testutil openspec/changes/standardize-plugin-domain-services/specs localdocs/plugin-domain-service-unification-design.md -S`无匹配。
- `FB-10`补充验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin -count=1`中`capabilityhost`和`plugin`通过；`wasm`包当时受未完成`FB-9`影响编译失败，失败为旧 host config 参数契约和`noopTestSysConfigService`相关未同步项，与本次删除`Files.Download`无关。
- `FB-9`验证：`cd apps/lina-core && go test ./pkg/plugin/capability/hostconfigcap ./internal/service/plugin/internal/hostconfig ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/capabilityhost -count=1`
- `FB-9`验证：`cd apps/lina-core && go test ./pkg/plugin/capability ./pkg/plugin/pluginbridge ./internal/service/plugin ./internal/cmd -count=1`
- `FB-9`验证：`openspec validate standardize-plugin-domain-services --strict`
- `FB-9`验证：`git diff --check -- apps/lina-core/pkg/plugin/capability/hostconfigcap apps/lina-core/internal/service/plugin/internal/hostconfig apps/lina-core/internal/service/plugin/internal/capabilityhost/capabilityhost.go apps/lina-core/pkg/plugin/pluginbridge/internal/domainhostcall/domainhostcall_config.go apps/lina-core/internal/service/plugin/internal/testutil apps/lina-core/internal/service/plugin/internal/integration apps/lina-core/internal/service/plugin/internal/wasm apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md openspec/changes/standardize-plugin-domain-services`
- `FB-9`静态检索：旧 host config 参数方法名、旧投影类型名、旧 adapter 构造名和旧文档措辞均无匹配。
- `FB-11`验证：`cd apps/lina-core && go test ./pkg/plugin/capability/hostconfigcap ./internal/service/plugin/internal/hostconfig ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/capabilityhost -count=1`
- `FB-11`验证：`cd apps/lina-core && go test ./pkg/plugin/capability ./pkg/plugin/pluginbridge ./internal/service/plugin ./internal/cmd -count=1`
- `FB-11`验证：`openspec validate standardize-plugin-domain-services --strict`
- `FB-11`验证：`git diff --check -- apps/lina-core/pkg/plugin/capability/hostconfigcap apps/lina-core/internal/service/plugin/internal/hostconfig apps/lina-core/pkg/plugin/pluginbridge/internal/domainhostcall/domainhostcall_config.go apps/lina-core/internal/service/plugin/internal/testutil apps/lina-core/internal/service/plugin/internal/integration apps/lina-core/internal/service/plugin/internal/wasm openspec/changes/standardize-plugin-domain-services/tasks.md`
- `FB-12`验证：`cd apps/lina-core && go test ./internal/service/file ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm ./pkg/plugin/pluginbridge/internal/domainhostcall ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/pluginbridge/protocol/hostservices ./pkg/plugin/capability/filecap -count=1`
- `FB-12`验证：`cd apps/lina-core && go test ./internal/cmd -count=1`
- `FB-12`验证：`cd apps/lina-core && go test ./pkg/plugin/capability ./pkg/plugin/pluginbridge ./internal/service/plugin -count=1`
- `FB-12`验证：`cd apps/lina-core && go test ./internal/service/file ./internal/service/datascope ./internal/service/plugin/internal/wasm ./internal/cmd -count=1`
- `FB-12`验证：`openspec validate standardize-plugin-domain-services --strict`
- `FB-12`验证：`git diff --check -- apps/lina-core/internal/service/file apps/lina-core/internal/service/plugin/internal/capabilityhost apps/lina-core/internal/service/plugin/internal/wasm apps/lina-core/pkg/plugin/capability/filecap apps/lina-core/pkg/plugin/pluginbridge apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md openspec/changes/standardize-plugin-domain-services/tasks.md`
- `FB-12`补充验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/wasm -run 'TestHandleHostServiceInvokeFiles' -count=1`
- `FB-12`补充验证：`cd apps/lina-core && go test ./internal/service/file ./internal/service/datascope ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm ./pkg/plugin/pluginbridge/internal/domainhostcall ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/pluginbridge/protocol/hostservices ./pkg/plugin/capability/filecap -count=1`
- `FB-12`补充验证：`cd apps/lina-core && go test ./pkg/plugin/capability ./pkg/plugin/pluginbridge ./internal/service/plugin ./internal/cmd -count=1`
- `FB-12`补充验证：`openspec validate standardize-plugin-domain-services --strict`
- `FB-12`补充验证：`git diff --check -- apps/lina-core/internal/service/file apps/lina-core/internal/service/datascope apps/lina-core/internal/service/plugin/internal/capabilityhost apps/lina-core/internal/service/plugin/internal/wasm apps/lina-core/pkg/plugin/capability/filecap apps/lina-core/pkg/plugin/pluginbridge apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md openspec/changes/standardize-plugin-domain-services/tasks.md`
- `FB-13`验证：`cd apps/lina-core && go test ./internal/service/jobmgmt/... ./internal/service/user ./pkg/plugin/capability ./pkg/plugin/pluginbridge ./internal/service/plugin ./internal/service/plugin/internal/wasm ./internal/cmd -count=1`
- `FB-13`验证：`openspec validate standardize-plugin-domain-services --strict`
- `FB-13`验证：`git diff --check -- apps/lina-core/pkg/plugin/capability/jobcap apps/lina-core/internal/service/jobmgmt apps/lina-core/internal/service/plugin/internal/capabilityhost apps/lina-core/internal/service/plugin/plugin_host_services.go apps/lina-core/internal/cmd/internal/httpstartup apps/lina-core/internal/service/user apps/lina-core/pkg/plugin/pluginbridge/internal/domainhostcall apps/lina-core/internal/service/plugin/internal/testutil apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md localdocs/plugin-domain-service-unification-design.md openspec/changes/standardize-plugin-domain-services`
- `FB-13`静态检索：`rg -n "RegisterInput|func \\([^)]*jobs[^)]*\\) Register|func \\([^)]*Jobs[^)]*\\) Register|job-register" apps/lina-core apps/lina-plugins -g '*.go'`仅命中源码插件声明期`RegisterJobs`；`rg -n "InsertAndGetId|\\.Update\\(|\\.Delete\\(|Data\\(do\\.SysJob" apps/lina-core/internal/service/jobmgmt/capabilityadapter -g '*.go'`仅命中测试调用，不存在 adapter 生产代码直写`sys_job`。
- `FB-14`验证：`cd apps/lina-core && go test ./api/job/v1 ./internal/service/jobmeta ./internal/service/jobmgmt/... ./internal/service/cron ./pkg/plugin/capability ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin -count=1`
- `FB-14`验证：`cd apps/lina-core && go test ./internal/cmd -count=1`
- `FB-14`验证：`openspec validate standardize-plugin-domain-services --strict`
- `FB-14`验证：`git diff --check -- apps/lina-core/api/job/v1 apps/lina-core/internal/service/cron apps/lina-core/internal/service/jobmeta apps/lina-core/internal/service/jobmgmt apps/lina-core/internal/service/plugin apps/lina-core/pkg/plugin openspec/changes/standardize-plugin-domain-services/tasks.md`
- `FB-14`静态检索：`rg -n "jobmeta\\.JobStatus|type JobStatus|JobStatusEnabled|JobStatusDisabled|JobStatusPausedByPlugin|SetStatus\\([^\\n]*jobcap\\.JobID[^\\n]*string|SetStatus\\([^\\n]*capabilityjobcap\\.JobID[^\\n]*string" apps/lina-core -g '*.go'`无匹配。
- `FB-15`验证：`cd apps/lina-core && go test ./pkg/plugin/capability/manifestcap ./internal/service/plugin/internal/manifestresource ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm ./internal/service/plugin -count=1`通过；同次尝试覆盖`./internal/cmd`失败，失败点为当前工作区既有`internal/service/cron/cron_managed_jobs.go`中的`jobhandlerv1`未定义和`internal/controller/joblog/joblog_v1_cancel.go`中的`jobv1`未定义，属于未完成的 Job status 相关改动阻断，非 Manifest 迁移新增问题。
- `FB-15`验证：`cd apps/lina-core && go test ./internal/cmd/internal/httpstartup -count=1`同样受上述`jobhandlerv1`和`jobv1`未定义阻断，已记录剩余风险；`openspec validate standardize-plugin-domain-services --strict`通过。
- `FB-15`验证：`rg -n "manifestcap\\.ServiceFactory|capabilitymanifest\\.NewFactory|pluginservicemanifest\\.NewFactory|NewFactory\\(\\\"\\\"\\).*manifest|WithArtifactResources\\([^\\n]+\\) manifestcap\\.ServiceFactory" apps/lina-core -g '*.go'`无匹配；`rg -n "ServiceFactory|NewFactory|EmbeddedFilesResolver|serviceAdapter" apps/lina-core/pkg/plugin/capability/manifestcap -g '*.go'`无匹配；`rg -n "lina-core/internal/service/plugin/internal/manifestresource" apps/lina-core -g '*.go'`仅命中`internal/service/plugin`授权边界内文件。
- `FB-16`验证：`cd apps/lina-core && go test ./pkg/plugin/capability/hostconfigcap ./internal/service/plugin/internal/hostconfig ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm ./internal/service/plugin -count=1`通过。
- `FB-16`验证：`openspec validate standardize-plugin-domain-services --strict`通过。
- `FB-16`验证：`git diff --check -- apps/lina-core/internal/service/plugin/internal/hostconfig apps/lina-core/internal/service/plugin/internal/capabilityhost/capabilityhost.go apps/lina-core/internal/service/plugin/internal/store apps/lina-core/internal/service/plugin/internal/governance apps/lina-core/internal/service/plugin/internal/dependency apps/lina-core/internal/service/plugin/internal/openapi apps/lina-core/internal/service/plugin/internal/frontend apps/lina-core/internal/service/plugin/internal/integration openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `FB-16`静态检索：`rg -n 'internal/service/runtimeconfig|package runtimeconfig|"lina-core/internal/service/runtimeconfig"|runtimeconfig\\.New' apps/lina-core apps/lina-plugins -g'*.go'`无匹配；`rg -n 'lina-core/internal/(dao|model)' apps/lina-core/internal/service/plugin/internal/capabilityhost apps/lina-core/internal/service/plugin/internal/wasm -g'*.go' -g'!*_test.go'`无匹配。
- `FB-16`补充验证：`cd apps/lina-core && go test ./internal/cmd/internal/httpstartup -count=1`仍受当前工作区既有`internal/controller/user/user_v1_create.go`中的`usersvc.StatusNormal`未定义和`internal/controller/joblog/joblog_v1_cancel.go`中的`jobv1`未定义阻断，非本次`HostConfig`迁移新增问题。
- `FB-17`验证：`cd apps/lina-core && go test ./pkg/plugin/capability ./pkg/plugin/capability/capmodel ./pkg/plugin/capability/tenantcap ./internal/service/cachecoord ./internal/service/dict ./internal/service/file ./internal/service/notify ./internal/service/user/capabilityadapter ./internal/service/jobmgmt/capabilityadapter ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/capabilityowner ./internal/service/plugin/internal/hostconfig ./internal/service/role -count=1`通过。
- `FB-17`验证：`openspec validate standardize-plugin-domain-services --strict`通过。
- `FB-17`验证：`git diff --check -- apps/lina-core/internal/service/cachecoord apps/lina-core/internal/service/file apps/lina-core/internal/service/dict apps/lina-core/internal/service/notify apps/lina-core/internal/service/role apps/lina-core/internal/service/user/capabilityadapter apps/lina-core/internal/service/jobmgmt/capabilityadapter apps/lina-core/internal/service/plugin/internal/capabilityhost apps/lina-core/internal/service/plugin/internal/capabilityowner apps/lina-core/internal/service/plugin/internal/hostconfig apps/lina-core/pkg/plugin/capability/capmodel apps/lina-core/pkg/plugin/capability/tenantcap openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `FB-17`静态检索：`rg -n 'capabilitydomain|internal/service/capabilityadapter' apps/lina-core apps/lina-plugins -g '*.go' -g '*.md'`无匹配。
- `FB-18`验证：`rg -n "type ScopeService interface|SysConfigScopeService|jobcap\\.ScopeService|authz\\.ScopeService|filecap\\.ScopeService|notifycap\\.ScopeService|sessioncap\\.ScopeService|dictcap\\.ScopeService|hostconfigcap\\.ScopeService|plugincap\\.ScopeService|usercap\\.ScopeService" apps/lina-core/pkg/plugin/capability apps/lina-core/internal apps/lina-core/pkg/plugin/pluginbridge -g '!**/node_modules/**'`仅命中`orgcap/orgspi.ScopeService`和`tenantcap/tenantspi.ScopeService`。
- `FB-18`验证：`cd apps/lina-core && go test ./pkg/plugin/capability ./pkg/plugin/capability/... ./pkg/plugin/pluginbridge ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/role ./internal/service/dict ./internal/service/file ./internal/service/notify ./internal/service/user/capabilityadapter ./internal/service/plugin/internal/hostconfig ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm -count=1`通过。
- `FB-18`验证：`openspec validate standardize-plugin-domain-services --strict`通过。
- `FB-18`验证：`git diff --check -- apps/lina-core/pkg/plugin/capability/jobcap/jobcap.go apps/lina-core/pkg/plugin/capability/authcap/authz/authz.go apps/lina-core/pkg/plugin/capability/filecap/filecap.go apps/lina-core/pkg/plugin/capability/notifycap/notifycap.go apps/lina-core/pkg/plugin/capability/sessioncap/sessioncap.go apps/lina-core/pkg/plugin/capability/dictcap/dictcap.go apps/lina-core/pkg/plugin/capability/hostconfigcap/hostconfigcap.go apps/lina-core/pkg/plugin/capability/plugincap/plugincap.go apps/lina-core/pkg/plugin/capability/usercap/usercap.go openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `FB-18`验证：`rg -n "ScopeService|SysConfigScopeService|scope service" apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md`无匹配，确认无需同步插件能力 README。
- `FB-19`验证：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/pluginbridge/protocol/hostservices ./pkg/plugin/pluginbridge/internal/domainhostcall ./internal/service/plugin/internal/wasm -run 'TestHostService|TestCatalog|TestNotifications|TestHandleHostServiceInvokeNotifications|TestValidateHostServiceSpecsAcceptsDomainServicesWithoutResources|TestValidateHostServiceSpecsAcceptsCacheLockNotifyResources' -count=1`通过。
- `FB-19`验证：`cd apps/lina-core && go test ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/pluginbridge/protocol/hostservices ./pkg/plugin/pluginbridge/internal/domainhostcall ./pkg/plugin/pluginbridge ./internal/service/plugin/internal/wasm -count=1`通过。
- `FB-19`验证：`cd apps/lina-core && go test ./internal/service/plugin ./internal/cmd -count=1`通过。
- `FB-19`验证：`openspec validate standardize-plugin-domain-services --strict`通过。
- `FB-19`验证：`git diff --check -- apps/lina-core/pkg/plugin/pluginbridge/internal/hostservice apps/lina-core/pkg/plugin/pluginbridge/protocol apps/lina-core/pkg/plugin/pluginbridge/internal/domainhostcall apps/lina-core/internal/service/plugin/internal/wasm apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `FB-21`验证：`cd apps/lina-core && go test ./pkg/plugin/capability/plugincap -count=1`通过。
- `FB-21`验证：`openspec validate standardize-plugin-domain-services --strict`通过。
- `FB-21`验证：`rg -n "plugincap_config_services|plugincap_lifecycle_events|plugincap_lifecycle_services|plugincap_state_services|plugincap_state_enablement" apps/lina-core/pkg/plugin/capability/plugincap openspec/changes/standardize-plugin-domain-services/tasks.md`无匹配。
- `FB-21`验证：`git diff --check -- apps/lina-core/pkg/plugin/capability/plugincap/plugincap.go apps/lina-core/pkg/plugin/capability/plugincap/plugincap_config.go apps/lina-core/pkg/plugin/capability/plugincap/plugincap_config_access.go apps/lina-core/pkg/plugin/capability/plugincap/plugincap_config_services.go apps/lina-core/pkg/plugin/capability/plugincap/plugincap_config_test.go apps/lina-core/pkg/plugin/capability/plugincap/plugincap_lifecycle.go apps/lina-core/pkg/plugin/capability/plugincap/plugincap_lifecycle_events.go apps/lina-core/pkg/plugin/capability/plugincap/plugincap_lifecycle_services.go apps/lina-core/pkg/plugin/capability/plugincap/plugincap_state.go apps/lina-core/pkg/plugin/capability/plugincap/plugincap_state_enablement.go apps/lina-core/pkg/plugin/capability/plugincap/plugincap_state_services.go`通过。
- `FB-21`补充验证：`cd apps/lina-core && go test ./pkg/plugin/capability -count=1`未通过，失败点为当前工作区既有`pkg/plugin/capability/tenantcap/tenantspi/tenantspi_host_impl.go:402`语法错误，阻断能力根包编译；该文件不属于本轮`plugincap`整理改动，变更包自身已通过编译和单元测试。
- `FB-22`验证：`rg --files apps/lina-core/pkg/plugin/capability/plugincap | sort`仅列出`plugincap.go`和`plugincap_*.go`源码文件。
- `FB-22`验证：`cd apps/lina-core && go test ./pkg/plugin/capability/plugincap -count=1`通过。
- `FB-22`验证：`git diff --check -- apps/lina-core/pkg/plugin/capability/plugincap openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `FB-22`验证：`openspec validate standardize-plugin-domain-services --strict`通过。
- `FB-24`验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/pluginconfig ./pkg/plugin/capability/plugincap ./internal/service/plugin/internal/capabilityowner ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin ./internal/service/plugin/internal/testutil ./internal/cmd/internal/httpstartup -count=1`通过。
- `FB-24`验证：`cd apps/lina-plugins/linapro-monitor-server && GOWORK=off go test ./backend/... -count=1`通过。
- `FB-24`验证：`cd apps/lina-plugins/linapro-monitor-server && GOWORK=off go test ./backend/internal/service/config -count=1`通过。
- `FB-24`验证：`openspec validate standardize-plugin-domain-services --strict`通过。
- `FB-24`验证：`git diff --check -- apps/lina-core/internal/service/plugin/internal/pluginconfig apps/lina-core/internal/service/plugin/plugin_config.go apps/lina-core/internal/service/plugin/plugin.go apps/lina-core/internal/service/plugin/plugin_host_services.go apps/lina-core/internal/service/plugin/internal/capabilityhost apps/lina-core/internal/service/plugin/internal/capabilityowner apps/lina-core/internal/service/plugin/internal/wasm apps/lina-core/internal/cmd/internal/httpstartup apps/lina-plugins/linapro-monitor-server/backend/plugin_test.go apps/lina-plugins/linapro-monitor-server/backend/internal/service/config/config_test.go openspec/changes/standardize-plugin-domain-services/tasks.md`通过。

### FB-49 实施记录

- 根因：`FB-44`为了避免普通`tenantcap.FilterService`暴露`*gdb.Model`，将源码插件表过滤放到`pluginhost.Services.TenantTableFilter()`顶层入口；但源码插件同时还能看到`Services.Tenant().Filter()`，形成两个租户过滤入口，违反“治理能力内聚到对应领域组件”的设计目标。
- 将`pluginhost.Services`改为显式源码插件运行期接口，不再内嵌`capability.Services`，并把`Tenant()`返回值改为源码插件租户视图`pluginhost.TenantService`；该视图的`Filter()`静态返回`tenantspi.PluginTableFilterService`。
- 新增`capabilityhost.NewSourceServices`，从启动期共享`capability.Services`构造源码插件服务视图；普通`capability.Services`、动态`pluginbridge.Services`和 WASM host service 仍只通过`tenantcap.Service.Filter()`暴露可序列化租户过滤上下文。
- 源码插件`linapro-content-notice`、`linapro-monitor-operlog`、`linapro-monitor-loginlog`、`linapro-monitor-online`、`linapro-org-core`和`linapro-demo-source`统一改用`services.Tenant().Filter()`获取表过滤器；`OrgProviderEnv`也改为通过源码插件租户视图读取`TenantFilter`。
- 删除生产代码中的顶层`TenantTableFilter()`调用路径，并补充`pkg/plugin/pluginhost`契约测试，防止`pluginhost.Services`重新暴露顶层`TenantTableFilter`。
- 同步更新`apps/lina-core/pkg/plugin/README.md`、`README.zh-CN.md`、当前变更`design.md`和增量规范，明确源码插件表过滤归属`Services.Tenant().Filter()`，普通租户能力和动态插件协议不暴露`*gdb.Model`。
- 插件本地规范检查：本轮修改的`linapro-content-notice`、`linapro-monitor-operlog`、`linapro-monitor-loginlog`、`linapro-monitor-online`、`linapro-org-core`和`linapro-demo-source`根目录均未发现本地`AGENTS.md`。
- 规则读取：本轮已读取`AGENTS.md`、`.agents/rules/openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`data-permission.md`、`testing.md`、`documentation.md`和`.agents/instructions/markdown-format.instructions.md`；未修改 HTTP API、SQL、前端 UI、缓存实现、运行时文案、语言包或开发工具脚本，确认`api-contract`、`database`、`frontend-ui`、`cache-consistency`、`i18n`和`dev-tooling`规则域无影响。
- `i18n`影响：未新增运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源；仅更新 README 和 OpenSpec 文档，确认无运行时`i18n`资源变更。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径、订阅状态或共享修订号策略；源码插件服务视图只包装既有启动期共享能力，确认无缓存一致性影响。
- 数据权限影响：租户表过滤仍由同一个启动期共享`tenantspi.PluginTableFilterService`基于`bizctxcap.CurrentContext`在数据库查询阶段注入`tenant_id`谓词；调用入口从顶层迁移到`Tenant().Filter()`，未新增数据读取、写入或存在性探测。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本、Shell/Node 工具或`linactl`入口，确认无跨平台工具影响。
- DI 来源检查：未新增运行期依赖；`tenantspi.NewPluginTableFilter`仍由`capabilityhost.New`使用启动期共享`bizCtxAdapter`创建一次并传入各领域 adapter，`NewSourceServices`仅将已创建的`capability.Services`包装为源码插件视图，没有请求路径临时`New()`、独立服务图或新缓存后端。
- 测试策略：本次属于后端 Go 插件宿主契约、源码插件运行期接口、官方源码插件注册入口和 README/OpenSpec 治理变更，不涉及用户可观察前端行为，未触发新增 E2E；通过 Go 编译门禁、源码插件包测试、宿主启动装配测试、OpenSpec 校验、静态检索和`git diff --check`验证。

### FB-49 验证

- `cd apps/lina-core && go test ./pkg/plugin/pluginhost ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/integration ./internal/service/plugin -count=1`通过。
- `cd apps/lina-core && go test ./pkg/plugin/capability ./pkg/plugin/capability/tenantcap/tenantspi -count=1`通过。
- `cd apps/lina-core && go test ./internal/cmd -count=1`通过。
- `cd apps/lina-plugins/linapro-content-notice && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-online && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-operlog && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-loginlog && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-org-core && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-demo-source && GOWORK=off go test ./backend/... -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check -- apps/lina-core/pkg/plugin/pluginhost apps/lina-core/internal/service/plugin/plugin_integration.go apps/lina-core/internal/service/plugin/plugin_test.go apps/lina-core/internal/service/plugin/internal/capabilityhost apps/lina-core/internal/service/plugin/internal/testutil/testutil_services.go apps/lina-core/internal/service/plugin/internal/integration/integration_source_services_test.go apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md openspec/changes/standardize-plugin-domain-services`通过。
- `git -C apps/lina-plugins diff --check -- linapro-content-notice/backend/plugin.go linapro-monitor-online/backend/plugin.go linapro-monitor-operlog/backend/plugin.go linapro-monitor-loginlog/backend/plugin.go linapro-org-core/backend/plugin.go linapro-demo-source/backend/plugin.go`通过。
- `rg -n "(\\.|\\b)TenantTableFilter\\(" apps/lina-core apps/lina-plugins -g'*.go' -g'!*_test.go'`无匹配，确认生产代码不再调用旧顶层入口。
- `rg -n "Services\\.TenantTableFilter|pluginhost\\.Services\\.TenantTableFilter|TenantTableFilter\\(\\)" apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md openspec/changes/standardize-plugin-domain-services/design.md openspec/changes/standardize-plugin-domain-services/specs -g'*.md'`仅命中`design.md`中的“不得保留顶层转发”负向描述，确认 README 和当前增量规范不再要求旧入口。

### FB-50 实施记录

- 根因：`FB-49`取消了顶层`TenantTableFilter()`，但通过新增`pluginhost.TenantService`让源码插件看到的`Services.Tenant().Filter()`返回`tenantspi.PluginTableFilterService`。这与普通`tenantcap.Service`形成重复租户服务契约，也让同一个租户过滤能力在`tenantcap`和`pluginhost`两处耦合表达。
- 将`pluginhost.Services`恢复为镜像普通`capability.Services`，删除`pluginhost.TenantService`和`capabilityhost.NewSourceServices`包装层；源码插件继续通过`Services.Tenant().Filter()`获取普通`tenantcap.FilterService`上下文。
- 删除`tenantspi.PluginTableFilterService`服务接口，把`NewPluginTableFilter`返回值收敛为`tenantcap.FilterService`，并新增`tenantspi.ApplyPluginTableFilter(ctx, filter, model, qualifier)`作为同进程 GoFrame 查询 helper。
- 源码插件`linapro-content-notice`、`linapro-monitor-operlog`、`linapro-monitor-loginlog`、`linapro-monitor-online`、`linapro-org-core`、`linapro-demo-source`及宿主 user、dict、file、job、hostconfig、org provider adapter 统一依赖`tenantcap.FilterService`；需要数据库侧租户谓词的调用点显式调用`tenantspi.ApplyPluginTableFilter(...)`。
- 同步更新`apps/lina-core/pkg/plugin/README.md`、`README.zh-CN.md`、当前变更`design.md`和增量规范，明确`pluginhost.Services`不提供独立租户服务镜像，GoFrame 表过滤只存在于`tenantspi`helper。
- 插件本地规范检查：本轮修改的`linapro-content-notice`、`linapro-monitor-operlog`、`linapro-monitor-loginlog`、`linapro-monitor-online`、`linapro-org-core`和`linapro-demo-source`根目录均未发现本地`AGENTS.md`。
- 规则读取：本轮已读取`AGENTS.md`、`.agents/rules/openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`data-permission.md`、`testing.md`、`documentation.md`、`i18n.md`和`.agents/instructions/markdown-format.instructions.md`；未修改 HTTP API、SQL、前端 UI、缓存实现、运行时文案、语言包或开发工具脚本，确认`api-contract`、`database`、`frontend-ui`、`cache-consistency`和`dev-tooling`规则域无影响。
- `i18n`影响：未新增运行时 UI 文案、菜单、按钮、API 文档源文本、插件清单或语言包资源；仅更新 README 和 OpenSpec 文档，确认无运行时`i18n`资源变更。
- 缓存一致性影响：未新增缓存后端、缓存写入、失效路径、订阅状态或共享修订号策略；仅改变租户过滤契约暴露形态，确认无缓存一致性影响。
- 数据权限影响：租户表过滤仍在数据库查询阶段注入`tenant_id`谓词；实现从服务方法改为`tenantspi.ApplyPluginTableFilter(...)`helper，不新增数据读取、写入或存在性探测。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本、Shell/Node 工具或`linactl`入口，确认无跨平台工具影响。
- DI 来源检查：未新增运行期依赖；`tenantspi.NewPluginTableFilter`仍由`capabilityhost.New`使用启动期共享`bizCtxAdapter`创建一次并传入各领域 adapter。删除`NewSourceServices`包装层后，源码插件直接使用同一个 plugin-scoped`capability.Services`实例，没有请求路径临时`New()`、独立服务图或新缓存后端。
- 测试策略：本次属于后端 Go 插件宿主契约、源码插件运行期接口、官方源码插件注册入口和 README/OpenSpec 治理变更，不涉及用户可观察前端行为，未触发新增 E2E；通过 Go 编译门禁、源码插件包测试、宿主启动装配测试、OpenSpec 校验、静态检索和`git diff --check`验证。

### FB-50 验证

- `cd apps/lina-core && go test ./pkg/plugin/pluginhost ./pkg/plugin/capability/tenantcap/tenantspi ./pkg/plugin/capability ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/testutil ./internal/service/plugin/internal/integration ./internal/service/plugin -count=1`通过。
- `cd apps/lina-core && go test ./pkg/plugin/capability/orgcap/orgspi -count=1`通过。
- `cd apps/lina-core && go test ./internal/service/plugin/internal/wasm -count=1`通过。
- `cd apps/lina-core && go test ./internal/service/user/capabilityadapter ./internal/service/file ./internal/service/dict ./internal/service/jobmgmt/capabilityadapter ./internal/service/plugin/internal/hostconfig -count=1`通过。
- `cd apps/lina-core && go test ./internal/cmd -count=1`通过。
- `cd apps/lina-plugins/linapro-content-notice && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-online && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-operlog && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-monitor-loginlog && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-org-core && GOWORK=off go test ./backend/... -count=1`通过。
- `cd apps/lina-plugins/linapro-demo-source && GOWORK=off go test ./backend/... -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check -- apps/lina-core/pkg/plugin/pluginhost apps/lina-core/pkg/plugin/capability/tenantcap/tenantspi apps/lina-core/pkg/plugin/capability/orgcap/orgspi apps/lina-core/internal/service/plugin/plugin_integration.go apps/lina-core/internal/service/plugin/plugin_test.go apps/lina-core/internal/service/plugin/internal/capabilityhost apps/lina-core/internal/service/plugin/internal/testutil/testutil_services.go apps/lina-core/internal/service/plugin/internal/integration/integration_source_services_test.go apps/lina-core/internal/service/plugin/internal/wasm/wasm_host_service_test.go apps/lina-core/internal/service/plugin/internal/hostconfig apps/lina-core/internal/service/dict/dict_capability.go apps/lina-core/internal/service/file/file_capability.go apps/lina-core/internal/service/jobmgmt/capabilityadapter apps/lina-core/internal/service/user/capabilityadapter apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md openspec/changes/standardize-plugin-domain-services`通过。
- `git -C apps/lina-plugins diff --check -- linapro-content-notice/backend linapro-monitor-online/backend linapro-monitor-operlog/backend linapro-monitor-loginlog/backend linapro-org-core/backend linapro-demo-source/backend`通过。
- `rg -n "PluginTableFilterService|NewSourceServices|pluginhost\\.TenantService|type TenantService interface|tenantTableFilter\\(|Services\\.TenantTableFilter|Tenant\\(\\)\\.Filter\\(\\)\\.Apply|Services\\.Tenant\\(\\)\\.Filter\\(\\)\\.Apply" apps/lina-core apps/lina-plugins apps/lina-core/pkg/plugin/README.md apps/lina-core/pkg/plugin/README.zh-CN.md openspec/changes/standardize-plugin-domain-services/design.md openspec/changes/standardize-plugin-domain-services/specs -g'*.go' -g'*.md'`无匹配，确认重复租户服务镜像、旧表过滤 service 和旧顶层入口已清理。

### FB-51 实施记录

- 根因：`apps/lina-core/internal/service/plugin`根包在多轮插件服务拆分后保留了若干小于 50 行的薄 facade 文件和一个空壳文件，增加了服务根包浏览成本，与后端复杂度治理中“组件下源文件过多时尽可能合并薄文件”的要求不一致。
- 将安装 mock-data 错误包装从`plugin_install_mock_data.go`并入`plugin_lifecycle.go`，删除空壳`plugin_lifecycle_source.go`。
- 将插件配置 factory 入口从`plugin_config.go`并入宿主服务适配构造入口`plugin_host_services.go`。
- 将内置插件管理 guard 从`plugin_builtin_guard.go`并入平台治理和启动一致性所在的`plugin_startup_consistency.go`。
- 将 runtime upgrade 预览、执行和缓存发布适配收敛到`plugin_runtime_upgrade.go`，删除`plugin_runtime_upgrade_preview.go`、`plugin_runtime_upgrade_execute.go`和`plugin_upgrade_adapters.go`。
- `apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`已评估：本次仅调整根包内部文件组织，不改变插件公开能力、源码插件服务契约、动态 host service、插件生命周期暴露或 README 描述口径，无需同步更新。
- 规则读取与技能：本轮已读取`lina-feedback`、`lina-review`、`goframe-v2`、`AGENTS.md`、`.agents/rules/openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`testing.md`、`documentation.md`、`cache-consistency.md`、`data-permission.md`和`i18n.md`；未修改 HTTP API、SQL、前端 UI、运行时语言包或开发工具脚本，确认`api-contract`、`database`、`frontend-ui`和`dev-tooling`规则域无影响。
- `i18n`影响：未新增或修改运行时 UI 文案、菜单、按钮、API 文档源文本、错误码、插件清单或语言包资源，确认无运行时`i18n`资源影响。
- 缓存一致性影响：仅移动`upgradeCachePublisher`和`upgradeCacheFreshener`代码位置；仍复用根 facade 的`PublishPluginChange`、`syncEnabledSnapshotAndPublishRuntimeChange`和共享 runtime cache refresh 路径，权威数据源、事务后发布、共享修订号与失效语义不变。
- 数据权限影响：未新增或修改数据读取、写入、存在性探测、租户/组织边界或插件可见数据路径，确认无数据权限行为影响。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本、Shell/Node 工具或`linactl`入口，确认无跨平台工具影响。
- DI 来源检查：未新增运行期依赖、构造参数或服务图；所有被移动的方法仍使用同一个`serviceImpl`已注入的启动期共享服务实例。
- 测试策略：本次为后端 Go 内部文件组织治理，不涉及用户可观察前端行为，未触发新增 E2E；通过 gofmt、静态命名检查、Go 包测试、OpenSpec 严格校验和`git diff --check`验证。

### FB-52 实施记录

- 根因：部分单元测试文件名不再有同名源码文件，主要来自前序文件合并或职责迁移后测试文件未同步归位，违反`backend-go.md`中单元测试文件必须关联源码文件的命名要求。
- 将根包静态边界测试`plugin_boundary_test.go`并入`plugin_test.go`，并将内部引用的 runtime upgrade 文件名更新为`plugin_runtime_upgrade.go`。
- 将内置插件启动测试`plugin_builtin_bootstrap_test.go`并入`plugin_auto_enable_test.go`，复用同一启动 bootstrap 与 mock-data helper。
- 将 runtime upgrade preview/execute 测试合并为`plugin_runtime_upgrade_test.go`，对应新的`plugin_runtime_upgrade.go`。
- 将`internal/capabilityowner/capabilityowner_plugin_config_test.go`重命名为`capabilityowner_plugin_test.go`，对应`capabilityowner_plugin.go`。
- 将`internal/integration/integration_source_services_test.go`重命名为`integration_test.go`，对应`integration.go`。
- 将`internal/runtime/runtime_role_access_test.go`中的同包测试 helper 并入`runtime_wiring_test.go`，保持 helper 对同包 runtime 测试可见。
- 将`internal/wasm/wasm_host_service_dispatch_test.go`并入`wasm_host_service_registry_test.go`，统一覆盖 registry 与 dispatch 注册行为。
- `i18n`、缓存一致性、数据权限、开发工具跨平台和 DI 来源影响：本次仅调整测试文件命名与测试 helper 归位，不修改生产运行时行为、缓存失效、数据边界、工具脚本或服务构造来源，确认无影响。
- 测试策略：本次属于 Go 单元测试治理，不新增业务断言；通过静态文件命名检查、Go 测试和 OpenSpec 校验闭环。

### FB-51 / FB-52 验证

- `find apps/lina-core/internal/service/plugin -type f -name '*_test.go' | sort | while read -r f; do base=${f%_test.go}; if [ ! -f "$base.go" ]; then printf '%s -> missing %s.go\n' "$f" "$base"; fi; done`无输出，确认插件服务目录内测试文件均存在同名源码文件。
- `gofmt -w`已覆盖本轮触碰的 Go 文件。
- `git diff --check -- apps/lina-core/internal/service/plugin/plugin_lifecycle.go apps/lina-core/internal/service/plugin/plugin_host_services.go apps/lina-core/internal/service/plugin/plugin_startup_consistency.go apps/lina-core/internal/service/plugin/plugin_auto_enable_test.go apps/lina-core/internal/service/plugin/plugin_test.go apps/lina-core/internal/service/plugin/internal/runtime/runtime_wiring_test.go apps/lina-core/internal/service/plugin/internal/wasm/wasm_host_service_registry_test.go openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `git diff --cached --check -- openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `cd apps/lina-core && go test ./internal/service/plugin ./internal/service/plugin/internal/capabilityowner ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/wasm -count=1`通过。
- `cd apps/lina-core && go test ./internal/service/plugin/internal/runtime -run 'TestNewRuntimeWiresRequiredDependencies|TestDynamicRouteIdentitySnapshotFiltersRolesByTokenTenant|TestRunReconcilerTickSafelyRecoversPanic|TestRuntimeReconcilerUsesRuntimeUpgradeState' -count=1`通过。
- `cd apps/lina-core && go test ./internal/service/plugin ./internal/service/plugin/internal/capabilityowner ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/wasm -count=1`中`plugin`、`capabilityowner`、`integration`和`wasm`通过；`runtime`全包失败点为动态样例构建依赖`pkg/plugin/pluginbridge/recordstore/recordstore_exec_wasip1.go`中`Tx.invoker`字段不一致（`unknown field invoker in struct literal of type Tx`、`tx.invoker undefined`），属于当前工作区既有 pluginbridge 变更导致的 WASM 构建问题，不是本轮文件合并或测试命名调整引入。

### FB-53 实施记录

- 根因：`cachecoord.BumpSharedRevisionInTx`和多个 capability adapter 在`dao.Transaction`闭包内继续传递并直接使用`gdb.TX`，通过`tx.Model(...)`构造模型；但 GoFrame 事务上下文已经绑定在回调传入的`ctx`中，业务层应通过`dao.Xxx.Ctx(ctx)`继承事务，避免把事务实现泄漏到业务 helper 和跨领域适配器。
- 将`BumpSharedRevisionInTx`签名改为只接收`ctx`和领域参数，内部使用`dao.SysCacheRevision.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error { ... })`进入事务上下文，并在闭包内统一通过`dao.SysCacheRevision.Ctx(ctx)`完成`InsertIgnore`、`LockUpdate`和`Update`；该 helper 已在`FB-54`中删除，相关写路径改为事务提交后调用统一`cachecoord.Service`发布修订号。
- 将用户状态、角色授权、字典刷新、通知删除、运行时配置写入、租户插件启用状态写入和`linapro-org-core`组织关联写入中的`tx.Model(...)`替换为对应`dao.Xxx.Ctx(ctx)`调用；事务闭包参数统一改为`_ gdb.TX`。
- 将`hostconfig.lockVisibleRow`和`capabilityowner.bumpPluginRuntimeCacheRevision`等内部 helper 的`gdb.TX`参数删除，只通过`ctx`传递事务上下文。
- 插件本地规范检查：本轮修改的`apps/lina-plugins/linapro-org-core`根目录未发现本地`AGENTS.md`，按顶层`AGENTS.md`和命中规则执行。
- 规则读取与技能：本轮已读取`lina-feedback`、`lina-review`、`goframe-v2`、`AGENTS.md`、`.agents/rules/openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`database.md`、`cache-consistency.md`、`data-permission.md`、`testing.md`和`i18n.md`；未修改 HTTP API、SQL、前端 UI、运行时文案、语言包或开发工具脚本，确认`api-contract`、`frontend-ui`和`dev-tooling`规则域无影响。
- `i18n`影响：未新增或修改运行时 UI 文案、菜单、按钮、API 文档源文本、错误码、插件清单或语言包资源，确认无运行时`i18n`资源影响。
- 缓存一致性影响：涉及权限、字典、运行时配置和插件运行时缓存修订号写入；权威数据源仍为业务表和`sys_cache_revision`，修订号写入继续与业务写入处于同一个事务`ctx`中，未改变跨实例共享修订号、一致性模型、故障降级或最大可接受陈旧时间。
- 数据权限影响：未新增数据读取、写入或存在性探测路径；既有`EnsureVisible`、租户过滤和目标行过滤条件保持不变，只替换模型构建入口。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本、Shell/Node 工具或`linactl`入口，确认无跨平台工具影响。
- DI 来源检查：未新增运行期依赖、构造参数、服务实例或共享后端；本次仅调整事务上下文传递方式。
- 测试策略：本次属于后端 Go 内部事务治理，不涉及用户可观察前端行为，未触发新增 E2E；通过 Go 包测试、OpenSpec 严格校验、静态检索和`git diff --check`验证。

### FB-53 验证

- `rg -n "tx\\.Model" apps/lina-core/internal/service apps/lina-plugins/linapro-org-core/backend/internal/service apps/lina-plugins/linapro-org-core/backend/internal/provider -g '*.go'`无输出，确认目标业务服务和插件 provider 中不再直接使用`tx.Model`。
- `rg -n "func\\([^\\n]*tx\\s+gdb\\.TX" apps/lina-core/internal/service apps/lina-plugins/linapro-org-core/backend/internal/service apps/lina-plugins/linapro-org-core/backend/internal/provider -g '*.go' -g '!**/datahost/**'`无输出，确认排除底层动态数据事务执行器后，目标业务事务闭包不再命名并传递`tx`参数。
- `cd apps/lina-core && go test ./internal/service/cachecoord ./internal/service/user/capabilityadapter ./internal/service/dict ./internal/service/notify ./internal/service/plugin/internal/hostconfig ./internal/service/plugin/internal/capabilityowner -count=1`通过。
- `cd apps/lina-plugins/linapro-org-core && GOWORK=off go test ./backend/internal/service/dept ./backend/internal/provider/orgcapadapter -count=1`通过。
- `cd apps/lina-core && go test ./internal/service/role -run '^$' -count=1`未通过，失败点为`internal/service/role/role_access_cache_test.go`中的`fakeRoleConfigService`未实现当前`config.Service.GetRaw`方法，属于当前工作区既有测试替身接口缺口；本轮事务改动触及的`role_capability.go`静态扫描已通过，但`role`全包编译门禁仍受该既有问题阻断。
- `git diff --check -- apps/lina-core/internal/service/notify/notify_inbox.go apps/lina-core/internal/service/role/role_impl.go apps/lina-core/internal/service/user/user_impl.go && git -C apps/lina-plugins diff --check -- linapro-org-core/backend/internal/service/dept/dept_impl.go linapro-org-core/backend/internal/provider/orgcapadapter/orgcapadapter.go`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。

### FB-54 实施记录

- 根因：`cachecoord.BumpSharedRevisionInTx`作为导出 helper 在插件领域适配器写路径中直接写`sys_cache_revision`，绕过了`cachecoord.Service`的统一 key 规范、原因归一化、协调后端选择和事件发布；在 Redis/事件后端场景下，SQL 直写无法被其他节点按统一机制观察。同时`cachecoord_impl.go`和`cachecoord_topology.go`均为 50 行以内薄文件，增加了组件浏览成本。
- 删除未跟踪的`apps/lina-core/internal/service/cachecoord/cachecoord_revision_tx.go`，不再保留 SQL 直写修订号入口。
- 合并`cachecoord_impl.go`和`cachecoord_topology.go`为`cachecoord_wiring.go`，保留`cachecoord.go`作为契约入口、`cachecoord_revision.go`作为核心修订逻辑、`cachecoord_code.go`作为错误码约定。
- `plugin.NewHostServices`和`capabilityhost.New`新增显式`cachecoord.Service`参数，由`httpstartup`中已有`cachecoord.Default(clusterSvc)`启动期共享实例传入；该实例已经绑定统一 cluster/coordination 后端。
- `role`、`dict`、`hostconfig.SysConfig`和`capabilityowner`适配器改为持有注入的`cachecoord.Service`，业务事务只写权威数据，事务成功返回后再调用`MarkChanged`或`MarkTenantChanged`发布修订号；缺少`cachecoord`时写路径在进入事务前返回`CodeCapabilityUnavailable`。
- `user`能力的`SetStatus`移除无 owner 时的直接 DAO fallback，不再维护第二套用户状态写入和权限缓存修订逻辑；正常装配继续通过用户 owner 写入并复用既有权限拓扑通知。
- `apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`已审查：文档描述插件能力目录、动态`hostServices`授权和方法级 cache governance，不描述内部`NewHostServices`构造签名或`cachecoord`传参，本次无需同步 README。
- 规则读取与技能：本轮已读取`lina-feedback`、`karpathy-guidelines`、`goframe-v2`、`AGENTS.md`、`.agents/rules/openspec.md`、`backend-go.md`、`architecture.md`、`cache-consistency.md`、`testing.md`、`database.md`、`data-permission.md`、`plugin.md`、`documentation.md`和`i18n.md`；未修改 HTTP API、SQL、前端 UI、运行时语言包或开发工具脚本，确认`api-contract`、`frontend-ui`和`dev-tooling`规则域无影响。
- `i18n`影响：未新增或修改运行时 UI 文案、菜单、按钮、API 文档源文本、错误码、插件清单或语言包资源，确认无运行时`i18n`资源影响。
- 缓存一致性影响：涉及权限、字典、运行时配置和插件运行时缓存修订号发布；权威数据源仍为`sys_role_menu`、`sys_dict_type/sys_dict_data`、`sys_config`和`sys_plugin_state`等业务表；失效触发点改为业务事务提交后；跨实例同步统一由注入的`cachecoord.Service`负责，集群模式复用共享修订号和事件发布，单机模式复用本地修订号；故障时写路径返回可见错误，避免静默丢失失效。
- 数据权限影响：未新增数据读取、写入或存在性探测面；角色、字典、配置、插件和用户写路径保留既有当前用户、租户、可见性和目标校验，缓存发布依赖不扩大数据权限边界。
- 数据库与 SQL 影响：未新增或修改 SQL、DAO、DO、Entity、索引、软删除字段或时间字段；删除的是未跟踪的业务 helper，生产数据写入继续使用既有 DAO/DO。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本、Shell/Node 工具或`linactl`入口，确认无跨平台工具影响。
- DI 来源检查：新增依赖 owner 为启动期共享`cachecoord.Service`；创建位置为`httpstartup`中的`cachecoord.Default(clusterSvc)`；传递路径为`pluginsvc.NewHostServices(..., cacheCoordSvc, ...) -> capabilityhost.New(..., cacheCoordSvc, ...) -> role/dict/hostconfig/capabilityowner adapter`；未在业务路径创建独立服务图，也未使用`cachecoord.Default(...)`作为隐藏 service locator。
- 测试策略：本次为后端 Go 内部缓存发布路径和文件组织治理，不涉及用户可观察前端行为，未触发 E2E；新增轻量单元测试覆盖缺少`cachecoord`时写路径 fail-closed，并通过相关包 Go 测试、OpenSpec 严格校验、静态检索、gofmt 和空白检查验证。

### FB-54 验证

- `cd apps/lina-core && go test ./internal/service/cachecoord ./internal/service/cachecoord/revisionctrl ./internal/service/role ./internal/service/dict ./internal/service/user/capabilityadapter ./internal/service/plugin/internal/hostconfig ./internal/service/plugin/internal/capabilityowner ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin ./internal/cmd/internal/httpstartup ./internal/cmd -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `git diff --check -- apps/lina-core/internal/service/cachecoord apps/lina-core/internal/service/role/role_capability.go apps/lina-core/internal/service/role/role_capability_test.go apps/lina-core/internal/service/role/role_access_cache_test.go apps/lina-core/internal/service/dict/dict_capability.go apps/lina-core/internal/service/dict/dict_capability_test.go apps/lina-core/internal/service/user/capabilityadapter/user_capability.go apps/lina-core/internal/service/plugin/internal/hostconfig apps/lina-core/internal/service/plugin/internal/capabilityowner/capabilityowner_plugin.go apps/lina-core/internal/service/plugin/internal/capabilityowner/capabilityowner_plugin_test.go apps/lina-core/internal/service/plugin/internal/capabilityhost/capabilityhost.go apps/lina-core/internal/service/plugin/internal/capabilityhost/capabilityhost_adapters_test.go apps/lina-core/internal/service/plugin/internal/capabilityhost/capabilityhost_storage_adapter_test.go apps/lina-core/internal/service/plugin/plugin_host_services.go apps/lina-core/internal/cmd/internal/httpstartup/httpstartup_runtime.go apps/lina-core/internal/cmd/internal/httpstartup/httpstartup_routes_test.go openspec/changes/standardize-plugin-domain-services/tasks.md`通过。
- `rg -n "BumpSharedRevisionInTx|cachecoord_revision_tx|cachecoord_impl|cachecoord_topology" apps/lina-core -g'*.go'`无相关生产引用；仅泛型构造名检索命中无关`notify`测试。
- `gofmt -l`覆盖本轮触碰的 Go 文件无输出。
- `rg -n "[ \\t]+$" apps/lina-core/internal/service/cachecoord/cachecoord_wiring.go apps/lina-core/internal/service/dict/dict_capability.go apps/lina-core/internal/service/dict/dict_capability_test.go apps/lina-core/internal/service/role/role_capability.go apps/lina-core/internal/service/role/role_capability_test.go apps/lina-core/internal/service/plugin/internal/capabilityowner/capabilityowner_plugin.go apps/lina-core/internal/service/plugin/internal/capabilityowner/capabilityowner_plugin_test.go apps/lina-core/internal/service/plugin/internal/hostconfig/hostconfig.go apps/lina-core/internal/service/plugin/internal/hostconfig/hostconfig_sysconfig.go apps/lina-core/internal/service/plugin/internal/hostconfig/hostconfig_sysconfig_test.go apps/lina-core/internal/service/user/capabilityadapter/user_capability.go`无输出，确认新增/未跟踪文件无行尾空白。

### FB-55 实施记录

- [x] 收敛`orgspi.Provider`接口方法，删除可由批量资源和用户组织画像派生的单项、可见性、工作台和冗余 host service 方法，并修复源码组织插件 provider、动态插件 guest/dispatcher 与测试覆盖。
- 根因：`orgspi.Provider`同时承载组织资源批量读取、用户组织画像、单项便捷读取、可见性校验和工作台候选接口，导致源码 provider 与动态 host-service 都需要重复实现可由批量接口稳定派生的方法，接口边界偏胖且扩展成本随派生方法线性增加。
- 将`orgspi.Provider`收敛为`DepartmentProvider`、`PostProvider`、`AssignmentProvider`和`ScopeProvider`组合接口，保留`BatchGetDepartments`、`BatchGetPosts`、`BatchGetUserOrgProfiles`、`BuildUserDeptScopeExists`等批量/通用契约，删除`GetUserDeptInfo`、`GetUserDeptIDs`、`GetUserPostIDs`、`ListUserDeptAssignments`、`ApplyUserDeptScope`、`EnsureDepartmentsVisible`、`EnsurePostsVisible`、`GetPost`、`UserDeptTree`、`ListPostOptions`和`ListPostOptionsPage`等 Provider 级冗余方法。
- 在`orgspi`服务实现中保留既有`orgcap.Service`调用面，单项部门/岗位、用户部门/岗位 ID、可见性检查、工作台部门树和岗位候选均改为由批量资源与用户组织画像派生；岗位候选在派生为通用岗位列表时继续默认过滤启用状态，避免把展示型和便捷型方法继续压到 provider SPI。
- 调整`linapro-org-core`源码组织插件 provider：删除不再属于 SPI 的导出方法，将内部复用逻辑降级为私有 helper，provider 只实现收敛后的批量和 scope 契约。
- 调整动态插件 host-service 协议：删除旧的用户部门/岗位单项 wire method，新增并注册`org.department.batch_get`和`org.post.batch_get`，动态 guest client 通过`BatchGetUserProfiles`派生 assignment 便捷读取，并通过真实批量 host service 获取部门和岗位资源。
- 同步更新`apps/lina-core/pkg/plugin/README.md`、`README.zh-CN.md`和`linapro-demo-dynamic/plugin.yaml`中的组织 host-service 声明，避免文档和动态样例继续暴露已删除方法。
- 插件本地规范检查：本轮修改的`apps/lina-plugins/linapro-org-core`和`apps/lina-plugins/linapro-demo-dynamic`根目录均未发现本地`AGENTS.md`，按顶层`AGENTS.md`和命中规则执行。
- 规则读取与技能：本轮已读取`lina-feedback`、`lina-review`、`goframe-v2`、`karpathy-guidelines`、`AGENTS.md`、`.agents/rules/openspec.md`、`architecture.md`、`backend-go.md`、`plugin.md`、`data-permission.md`、`testing.md`、`api-contract.md`、`documentation.md`和`i18n.md`；未修改 HTTP API、SQL、前端 UI、运行时用户文案、语言包或开发工具脚本，确认`database`、`frontend-ui`和`dev-tooling`规则域无影响。
- `i18n`影响：未新增或修改运行时 UI 文案、菜单、按钮、错误消息、语言包或翻译资源；`plugin.yaml`仅调整动态 host-service 方法标识，确认无运行时`i18n`资源影响。
- 数据权限影响：未新增数据读取、写入或存在性探测入口；可见性校验改由`BatchGetDepartments`/`BatchGetPosts`的可见资源结果和缺失 ID 语义派生，用户组织范围继续由 provider 的 scope builder 返回条件表达式，未扩大租户、组织或用户可见性边界。
- 插件文档影响：主框架插件动态 host-service 方法发生变更，已同步审查并更新`apps/lina-core/pkg/plugin`中英 README 的组织能力方法表。
- 测试策略：本次为后端 Go SPI 与动态插件 host-service 契约收敛，不涉及用户可观察前端行为，未触发 E2E；通过 Go 包测试、动态插件测试、OpenSpec 严格校验、静态检索、gofmt 和空白检查验证。

### FB-55 验证

- `cd apps/lina-core && go test ./pkg/plugin/capability/orgcap/orgspi -count=1`通过。
- `cd apps/lina-core && go test ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/pluginbridge/protocol/hostservices ./pkg/plugin/pluginbridge/internal/domainhostcall -count=1`通过。
- `cd apps/lina-core && go test ./internal/service/plugin/internal/wasm -count=1`通过。
- `cd apps/lina-plugins/linapro-org-core && GOWORK=off go test ./backend/internal/provider/orgcapadapter -count=1`通过。
- `cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/internal/service/dynamic ./backend/internal/controller/dynamic ./backend/api/... -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。

### FB-56 实施记录

- 根因：当前 OpenSpec 虽要求动态`host service`顶层 service 与`capability.Services`领域目录一致，但同时允许`authz`作为独立顶层 service；代码也跟随暴露了`protocol.HostServiceAuthz`、`pluginbridge.Services.Authz()`和 WASM registry 中的独立`authz`注册路径，导致源码插件看到`Services.Auth().Authz()`，动态插件却看到`service: authz`，破坏全局领域一致性。
- 将动态顶层 service 收敛为`auth`：token 方法统一改为`token.tenant.select`、`token.tenant.switch`、`token.impersonation_token.issue`和`token.impersonation_token.revoke`；授权方法统一改为`authz.permissions.batch_get`、`authz.permissions.batch_has`、`authz.permissions.has`和`authz.users.platform_admin.check`。
- 授权粒度仍通过派生 capability 区分：token 方法派生`host:auth:token`，授权方法派生`host:auth:authz`；动态插件清单声明必须使用`service: auth`，宿主会拒绝顶层`service: authz`。
- 删除公开动态插件根入口`Services.Authz()`和`directory.Authz()`；动态 guest 授权调用继续通过`Services.Auth().Authz()`进入`authcap.Service`子能力，WASM dispatcher 由`auth`服务分派到`authz.*`方法族。
- 合并`protocol/hostservices`catalog 中的`authz`服务描述到`auth`服务描述，移除`HostServiceAuthz`公开 alias；保留`HostServiceMethodAuthz*`方法常量，作为`auth`服务下的授权方法前缀。
- 更新`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`的动态`hostServices`表，中英文均只保留一个`auth`行，并同步列出`host:auth:token`、`host:auth:authz`与新的方法名。
- 同步更新`standardize-plugin-domain-services`增量规范和设计，明确`authz`只是`auth`领域内的方法前缀/子能力，不得作为顶层动态 service；同时修正`expand-plugin-domain-capabilities`中仍引用旧授权方法名的规范残留，避免活跃变更之间出现矛盾。
- 规则读取与技能：本轮已读取`lina-feedback`、`openspec-apply-change`、`lina-review`、`goframe-v2`、`karpathy-guidelines`、`AGENTS.md`、`.agents/rules/openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`documentation.md`、`testing.md`、`data-permission.md`、`i18n.md`、`cache-consistency.md`、`dev-tooling.md`和`.agents/instructions/markdown-format.instructions.md`；未修改 HTTP API、SQL、前端 UI、运行时用户文案、语言包或开发工具脚本，确认`api-contract`、`database`和`frontend-ui`规则域无影响。
- `i18n`影响：未新增或修改运行时 UI 文案、菜单、按钮、错误消息、API 文档源文本、插件清单文案或语言包资源；仅修改协议标识符、README 和 OpenSpec 文档，确认无运行时`i18n`资源影响。
- 缓存一致性影响：未新增缓存、权限快照失效、配置失效、插件状态快照或跨实例同步路径；本次只调整动态 host-service service/method 命名与授权 capability 派生，确认无缓存一致性影响。
- 数据权限影响：未新增数据读取、写入、聚合、下载或存在性探测；授权方法仍委托既有`Auth().Authz()`领域 owner 和 dispatcher 授权校验，未扩大权限、租户或组织数据边界。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本、Shell/Node 工具或`linactl`入口，确认无跨平台工具影响。
- DI 来源检查：未新增运行期依赖、构造参数、服务实例或共享后端；WASM registry 只把既有`auth`服务注册面覆盖到 token 和 authz 方法族，仍通过启动期共享的`capability.Services.Auth()`获取子能力，没有请求路径临时`New()`或独立服务图。
- 测试策略：本次为后端 Go 动态插件协议、guest SDK、WASM dispatcher 和文档治理变更，不涉及前端用户可观察行为，未触发新增 E2E；通过 Go 单元测试、OpenSpec 严格校验、静态检索、gofmt 和`git diff --check`验证。

### FB-56 验证

- `rg -n 'HostServiceAuthz|host:authz|service:[[:space:]]*authz|Services\.Authz\(' apps/lina-core apps/lina-plugins openspec/changes/standardize-plugin-domain-services openspec/changes/expand-plugin-domain-capabilities openspec/changes/complete-plugin-domain-capability-expansion -g '*.go' -g '*.md' -g '*.yaml'`仅命中拒绝顶层`service: authz`的规范说明和源码插件正常的`Services.Auth().Authz()`子能力命名，未发现旧顶层动态`authz`入口。
- `rg -n 'permissions\.batch_get|permissions\.batch_has|permissions\.has|users\.platform_admin\.check|tenant\.select|tenant\.switch|impersonation_token\.issue|impersonation_token\.revoke' apps/lina-core apps/lina-plugins openspec/changes/standardize-plugin-domain-services openspec/changes/expand-plugin-domain-capabilities openspec/changes/complete-plugin-domain-capability-expansion -g '*.go' -g '*.md' -g '*.yaml'`确认旧方法名已迁移为`token.*`或`authz.*`前缀；唯一`permissions.has`残留位于拒绝顶层`authz`的负向测试输入。
- `cd apps/lina-core && go test ./pkg/plugin/pluginbridge/internal/hostservice ./pkg/plugin/pluginbridge/protocol/hostservices ./pkg/plugin/pluginbridge/internal/domainhostcall -count=1`通过。
- `cd apps/lina-core && go test ./pkg/plugin/pluginbridge ./internal/service/plugin/internal/wasm -count=1`通过。
- `cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/wasm -count=1`通过。
- `openspec validate standardize-plugin-domain-services --strict`通过。
- `openspec validate expand-plugin-domain-capabilities --strict`通过。
- `git diff --check`通过。

### FB-57 实施记录

- 根因：`FB-56`已经移除动态插件顶层`authz` service，但动态 guest client 和 WASM dispatcher 仍分别保留`domainhostcall_authz.go`与`wasm_host_service_authz.go`，文件组织会继续暗示存在独立动态`authz`领域，与`service: auth`的协议边界不一致。
- 将`domainhostcall_authz.go`内容合并进`domainhostcall_auth.go`，让动态 guest 的 token 与授权子能力都集中在`Auth()`领域命名空间实现中；保留`authzService`类型，因为它实现的是`Services.Auth().Authz()`子能力契约。
- 将`wasm_host_service_authz.go`内容合并进`wasm_host_service_auth.go`，并把内部 dispatcher helper 从`dispatchAuthzHostService`改为`dispatchAuthAuthorizationMethods`，避免函数名继续表现为顶层 host service。
- 删除动态插件侧两个独立`authz`实现文件：`apps/lina-core/pkg/plugin/pluginbridge/internal/domainhostcall/domainhostcall_authz.go`和`apps/lina-core/internal/service/plugin/internal/wasm/wasm_host_service_authz.go`。
- 保留`apps/lina-core/pkg/plugin/capability/authcap/authz/authz.go`，因为它是`Auth`领域下的授权子能力公开契约，不属于动态插件顶层 service 拆分残留。
- `README.md`和`README.zh-CN.md`已审查：动态`hostServices`表在`FB-56`已统一为`auth`，本次仅调整 Go 文件组织，无需再次修改 README。
- 规则读取与技能：本轮已读取`lina-feedback`、`lina-review`、`goframe-v2`、`karpathy-guidelines`、`AGENTS.md`、`.agents/rules/openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`testing.md`、`documentation.md`、`data-permission.md`、`i18n.md`、`cache-consistency.md`、`dev-tooling.md`和`.agents/instructions/markdown-format.instructions.md`；未修改 HTTP API、SQL、前端 UI、运行时用户文案、语言包或开发工具脚本，确认`api-contract`、`database`和`frontend-ui`规则域无影响。
- `i18n`影响：未新增或修改运行时 UI 文案、API 文档源文本、错误消息、插件清单文案或语言包资源；仅调整 Go 文件组织和 OpenSpec 任务记录，确认无运行时`i18n`资源影响。
- 缓存一致性影响：未新增缓存、权限快照失效、配置失效、插件状态快照或跨实例同步路径；本次不改变授权方法调用语义，确认无缓存一致性影响。
- 数据权限影响：未新增数据读取、写入、聚合、下载或存在性探测；授权方法仍委托既有`Services.Auth().Authz()`子能力和动态 dispatcher 授权校验，确认无数据权限边界变化。
- 开发工具跨平台影响：未修改`Makefile`、`make.cmd`、CI、构建脚本、代码生成脚本、Shell/Node 工具或`linactl`入口，确认无跨平台工具影响。
- DI 来源检查：未新增运行期依赖、构造参数、服务实例或共享后端；仅移动同包私有函数和类型位置，没有请求路径临时`New()`或独立服务图。
- 测试策略：本次为后端 Go 动态插件 bridge/WASM 文件组织治理，不涉及前端用户可观察行为，未触发新增 E2E；通过 Go 单元测试、静态检索、OpenSpec 严格校验、gofmt 和`git diff --check`验证。

### FB-57 验证

- `find apps/lina-core apps/lina-plugins -name '*authz*.go' -type f | sort`仅输出`apps/lina-core/pkg/plugin/capability/authcap/authz/authz.go`，确认动态插件侧`authz`实现文件已合并。
- `rg -n 'dispatchAuthzHostService|wasm_host_service_authz|domainhostcall_authz' apps/lina-core openspec/changes/standardize-plugin-domain-services -g '*.go' -g '*.md'`无输出。
- `cd apps/lina-core && go test ./pkg/plugin/pluginbridge/internal/domainhostcall ./pkg/plugin/pluginbridge/internal/hostservice ./internal/service/plugin/internal/wasm -count=1`通过。
- `cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/wasm -count=1`通过。
