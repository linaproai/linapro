## Why

LinaPro 当前是单租户架构:`sys_*` 表无租户维度、用户角色全局共享、插件启用对所有用户一刀切。要把它进化为面向"公司内部多业务单元(BU)+ 平台管理员可穿透"的多租户全栈框架,必须从 schema、bizctx、解析中间件、缓存、JWT/会话、插件治理一路改造。同时,既有插件治理模型("装即启用,影响所有用户")在多租户下风险面过大,需要拆分为"平台管理员负责安装、租户管理员负责启用"的两层治理,并由插件自身通过否决型钩子保护卸载/禁用前置条件。

本次迭代选择**一次完整落地**(不分期),原因是:项目仍在初期、无历史负担,Pool 模型的列改动一次到位反而比分期改造代价更小;插件治理模型的两层化必须与多租户能力同时上线才自洽。

## What Changes

- **新增多租户能力接缝(host)**:在 `pkg/tenantcap` 与 `internal/service/tenantcap` 中建立稳定的 Provider 接缝,所有租户敏感业务表与租户作用域运行时状态表的 DAO 必须经过 `tenantcap.Apply` 注入 `tenant_id` 过滤,与既有 `datascope` 叠加。
- **新增 `multi-tenant` 源码插件**:owns 租户主表、用户-租户成员表、租户配置覆盖表;实现 `tenantcap.Provider`、代码固定解析责任链(override/JWT/session/header/subdomain/default)、LifecycleGuard 前置校验与平台管理入口;租户成员关系统一在用户管理页面承载。
- **改造 `org-center` 插件为租户感知**:dept/post/user 关联表加 `tenant_id`,DAO 接入租户过滤,监听租户生命周期事件以初始化默认部门树并清理租户数据。
- **schema 按职责加 `tenant_id` 列**:所有租户敏感业务表、租户作用域运行时状态表与现有插件业务表新增 `tenant_id INT NOT NULL DEFAULT 0`(0 = PLATFORM),并将对应查询索引升级为 `(tenant_id, ...)` 联合索引;平台控制面表保持全局唯一,不机械增加 `tenant_id`。
- **bizctx 增加 `TenantId` 字段**:从中间件层注入,沿请求链路传递,所有日志、缓存、审计统一携带。
- **JWT/会话 携带租户**:Claims 增加 `TenantId`,会话 store 以全局唯一 `token_id` 定位会话并保留 `tenant_id` 作为归属与校验维度,登录后或切换租户时重签 token,旧 token 立即作废。
- **角色与用户-角色关联租户化**:`sys_role` 增加 `tenant_id`,并将 `data_scope` 扩展为 `1=全部数据`、`2=本租户数据`、`3=本部门数据`、`4=本人数据`;`sys_user_role` 增加 `tenant_id`;权限/菜单解析按当前租户过滤。
- **平台/租户管理员能力组合模型**:不在角色表维护平台布尔字段,也不通过 `platform:*` / `tenant:*` 权限前缀表达边界。平台管理员由 `TenantId=0`、`data_scope=1` 与统一的 `system:*` 功能权限组合推导;租户管理员由目标租户上下文、租户数据权限与租户相关 `system:*` 功能权限组合推导。impersonation 某租户时按目标租户过滤,仅审计标记代操作。
- **字典/配置 实现"平台默认 + 租户覆盖"**:`sys_dict_*` 与 `sys_config` 读路径走 `(tenant_id=current) → fallback (tenant_id=0)`;写路径默认写当前租户,平台管理员可显式写平台层。
- **菜单保持平台全局**:`sys_menu` 不接入租户过滤;按租户隐藏功能通过"按租户启用插件"或角色分配实现。
- **文件存储租户隔离**:本地与对象存储路径前缀按租户分隔(`/storage/t/{tenant_id}/...`)。
- **缓存 key 与失效 携带租户维度**:`kvcache` / 运行时缓存 key 加 tenant 维度,`cluster` 失效广播携带 tenant scope。
- **插件治理两层化(BREAKING)**:
  - `plugin.yaml` 新增 `scope_nature`(`platform_only` / `tenant_aware`)。
  - `sys_plugin` 新增 `scope_nature` / `install_mode`(`global` / `tenant_scoped`)。
  - `sys_plugin_state` 新增 `tenant_id`,保留 `id` 自增技术主键,并以 `(plugin_id, tenant_id, state_key)` 唯一索引表达业务唯一性,保留插件运行时多 key 状态能力。
  - 平台管理员负责安装/卸载并选择 `install_mode`;租户管理员对 `tenant_scoped` 插件负责启用/禁用。
  - `IsEnabled(ctx, pluginID)` 改造为租户感知。
- **插件生命周期否决钩子(LifecycleGuard)**:新增 `CanUninstall` / `CanDisable` / `CanTenantDisable` / `CanTenantDelete` 接口族,插件自检卸载/禁用前置条件,宿主聚合多否决统一展示;支持平台管理员紧急 `--force` 通道并强制审计。
- **租户解析中间件固定策略**:责任链顺序、保留子域名清单、未识别请求行为(`prompt`)使用代码默认值,不提供解析配置表或后台运行时调整;`rootDomain` 后续开放设置,当前固定为空并禁用子域名解析;正式 JWT 是普通业务请求的权威租户身份,header/subdomain 仅作为登录前 hint。
- **审计日志全面租户化**:`monitor-operlog` 与 `monitor-loginlog` 表加 `tenant_id` 与 `acting_on_behalf_of_tenant_id`,平台管理员代为操作时双轨记录。
- **i18n 资源治理**:运行时翻译缓存保持按交付资源维度(`locale` / `sector` / `plugin`)精细失效;字典/配置等租户覆盖缓存按租户维度隔离;所有新增否决理由、错误信息均通过 i18n key 维护。

**重要边界**:
- 隔离模型采用 Pool(单库 + tenant_id 列),代码默认值固定为 `pool`,宿主配置模板不暴露租户隔离配置项。
- 用户-租户关系采用 1:N membership 模型(全局身份 + 多租户绑定),代码默认值为 `multi`;`single` 策略作为后续受控管理设置保留,当前不通过宿主配置文件开放。
- 租户 code 固定为 ASCII `[a-z0-9-]{2,32}`,不允许中文或其他 Unicode 字符;展示名 `name` 可本地化。
- 默认登录解析策略为代码固定责任链兜底 `default`(从用户所属租户列表中挑选),解析链不支持插件运行时配置调整;`rootDomain` 当前不支持设置。
- 本次不实现"按租户启用插件 UI 的批量操作面板"高级形态(单租户管理员单租户内启用/禁用即可),但 schema 与接口预留扩展点。

## Capabilities

### New Capabilities

- `multi-tenancy-foundation`: 宿主侧多租户能力接缝(`tenantcap` 接口、Service、bizctx 集成、DAO 注入纪律、Pool 模型 schema 总则)。
- `tenant-resolution`: 租户解析责任链(override/JWT/session/header/subdomain/default)、代码固定策略与未识别请求行为。
- `tenant-management`: 平台管理员侧的租户主体与生命周期(创建、暂停、删除)。
- `tenant-membership`: 用户-租户 1:N 绑定模型(成员关系、租户内角色、状态、平台/租户管理员区分)。
- `tenant-aware-authentication`: 多租户登录、租户选择、JWT 租户 claim、切换租户重签、平台管理员 impersonation。
- `tenant-config-override`: 字典/配置的"平台默认 + 租户覆盖"读取与写入语义。
- `tenant-lifecycle-events`: 租户创建/暂停/恢复/删除的显式领域编排边界,并删除未完整实现的 outbox/事件总线占位。
- `plugin-scope-nature`: `plugin.yaml` 中 `scope_nature` 字段语义、安装期校验、不可变契约。
- `plugin-install-mode`: 平台管理员安装期选择 `global`/`tenant_scoped`、模式切换规则、新租户加入策略。
- `plugin-tenant-enablement`: 租户管理员对 `tenant_scoped` 插件的启用/禁用、租户级状态存储、缓存失效。
- `plugin-lifecycle-guard`: 插件否决型钩子族(`CanUninstall`/`CanDisable`/`CanTenantDisable`/`CanTenantDelete`)、否决聚合、超时容错、`--force` 通道、审计要求。
- `tenant-data-isolation`: 文件存储路径、缓存 key、审计日志、跨租户操作日志的隔离与可观察性规范。

### Modified Capabilities

仅列出"规范级行为契约发生变化"的能力(下文每项均产出 delta spec)。租户敏感业务表的 `tenant_id` 字段、查询过滤与平台控制面表豁免清单在 `tasks.md` 中执行并由规范边界约束:

- `user-auth`: 登录流程增加租户解析与挑选;Claims 增加 TenantId;新增切换租户重签接口与 token 作废规则。
- `user-management`: 用户身份与租户成员关系分层;查询/创建/导入按租户隔离。
- `user-role-association`: 用户-角色绑定按租户隔离;平台上下文角色仅可绑定平台用户,租户上下文角色仅可绑定同租户 active membership 用户。
- `role-management`: 角色以 `tenant_id` 归属和 `data_scope` 表达数据边界;不维护平台角色布尔字段,可见性与可分配性按当前租户上下文隔离。
- `dict-management`: 读路径实现 platform fallback;字典类型可声明是否允许租户覆盖。
- `dictionary-management`: 平台字典与租户字典的可见性、写入权限规则。
- `config-management`: 配置读路径 platform fallback;租户管理员仅可写本租户配置项。
- `dept-management`: org-center 插件部门树按租户隔离;不依赖未实现的租户事件总线。
- `post-management`: org-center 插件岗位选项按租户隔离。
- `online-user`: 会话存储按 (tenant, token) 组合;会话查询/踢除按租户过滤。
- `login-log`: 登录日志增加租户与 impersonation 双轨字段。
- `oper-log`: 操作日志增加租户与 impersonation 双轨字段;强制操作单独审计类。
- `notice-management`: 通知按租户隔离;跨租户通知需平台权限。
- `user-message`: 消息按租户隔离。
- `cron-job-management`: 任务按租户隔离;内置系统任务为平台级。
- `plugin-manifest-lifecycle`: plugin.yaml 增加 `scope_nature`;sys_plugin 增加 `scope_nature` / `install_mode`;安装期一致性校验。
- `plugin-runtime-loading`: 路由全局挂载 + 请求时按 (tenant, plugin) 启用状态过滤。
- `plugin-startup-bootstrap`: 启动期按 (plugin, tenant) 维度装配状态缓存与一致性校验。
- `plugin-permission-governance`: 权限点统一使用 `system:*` 功能命名空间,平台/租户边界由路由平面、租户上下文、数据权限与插件启用状态共同约束。
- `plugin-cache-service`: 缓存 key 默认携带租户维度;失效广播按租户精细化。
- `plugin-storage-service`: 文件存储路径按租户前缀;跨租户访问需显式平台 API 或 impersonation。
- `plugin-host-service-extension`: 暴露给插件的 host service 自动透传 `bizctx.TenantId` 并校验租户可见性。
- `core-host-boundary-governance`: 把"租户能力接缝"列为宿主稳定接缝之一(与 orgcap 并列)。
- `module-decoupling`: 新增 multi-tenant 插件禁用/启用时的联动隐藏规范。
- `distributed-cache-coordination`: 失效消息与作用域 scope 增加 `tenant_id` 维度。
- `framework-i18n-runtime-performance`: 翻译包缓存保持全局交付资源维度;字典/配置租户覆盖缓存承担租户维度。
- `database-bootstrap-commands`: `make init` / `make mock` 默认写入 PLATFORM 并支持指定租户。
- `framework-bootstrap-installer`: 启动期装配 tenantcap.Provider 并校验 install_mode 与 scope_nature 一致性。
- `dashboard-workbench`: 工作台头部增加当前租户标识与切换器。
- `login-page-presentation`: 登录页根据解析策略呈现租户输入/挑选 UI。
- `management-workbench-i18n`: i18n 文案覆盖租户切换、平台管理员、租户管理员等新增术语。
- `e2e-suite-organization`: e2e 用例新增多租户分组与跨租户隔离矩阵。
- `spec-governance`: 多租户相关增量规范必须显式声明 tenant_id 在读/写/缓存/审计四类行为上的契约。
- `backend-conformance`: 增加"DAO 必须经 tenantcap.Apply"硬性条款。

## Impact

- **schema**:宿主租户敏感 `sys_*` 表结构按项目无历史负担策略直接合并回对应源建表 SQL;`sys_locker`、`sys_menu`、`sys_plugin`、`sys_plugin_release`、`sys_plugin_migration`、`sys_plugin_resource_ref`、`sys_plugin_node_state`、`sys_notify_channel` 等平台控制面或全局配置表不携带 `tenant_id`;`manifest/sql/016-multi-tenant-and-plugin-governance.sql` 仅保留本迭代新增的多租户 seed。
- **新插件**:`apps/lina-plugins/multi-tenant/`,owns 租户主体、成员、配置覆盖与平台租户管理 UI;租户成员关系由宿主用户管理页面展示与管理。
- **现有插件改造**:`org-center` 全表加 `tenant_id` 并接入租户过滤;`monitor-loginlog` / `monitor-online` / `monitor-operlog` / `content-notice` / `demo-control` / `plugin-demo-source` / `plugin-demo-dynamic` 全部加 `tenant_id` 列并接入租户过滤;`monitor-server` 保持 `platform_only`。
- **宿主代码**:新增 `pkg/tenantcap` 与 `internal/service/tenantcap`;`bizctx` 增字段;`auth` / `session` / `role` / `dict` / `sysconfig` / `file` / `notify` 消息与投递 / `usermsg` / `jobmgmt` / `jobhandler` / `jobmeta` / `online` / `cachecoord` / `kvcache` / `pluginruntimecache` / `plugin` 启用状态全面接入租户上下文;菜单目录、插件注册目录、插件发布迁移记录、分布式锁、通知通道配置保持平台全局。
- **API**:新增 `/auth/login-tenants`(登录后挑选租户)、`/auth/switch-tenant`、`/platform/tenants/*`、`/platform/plugins/*` install_mode 选项、`/tenant/plugins/*` 启用禁用接口、`/tenant/members/*` 成员管理接口。所有现有接口的请求/响应保持兼容(租户从 ctx 注入,不暴露在 DTO 中)。
- **前端**:登录页增加租户挑选器;工作台头部增加租户切换器;平台管理下新增"租户管理"页面;租户管理员通过"用户管理"维护本租户成员关系;不再提供独立"租户工作台"目录、平台"租户成员"菜单或租户侧成员/插件管理菜单。
- **配置**:宿主 `config.template.yaml` 不包含 `tenant.*` 段;租户默认策略在代码中集中定义,不创建或读取运行时解析配置表。
- **审计与可观察性**:操作日志、登录日志全面租户化;平台管理员 impersonation 双轨记录;否决钩子结果与 `--force` 操作单独审计。
- **i18n**:新增租户管理、平台管理员、否决理由等 i18n key,zh-CN/en-US 双语完整覆盖。
- **测试**:新增 e2e 多租户分组,覆盖租户隔离、平台 impersonation、插件 tenant_scoped 启用、否决钩子、租户生命周期、登录解析策略矩阵。
- **文档**:`README.md` 与 `README.zh-CN.md` 增加多租户章节;插件开发指南增加 `scope_nature` 与 `LifecycleGuard` 章节。
- **依赖**:无新增第三方依赖。
