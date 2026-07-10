# Design

## 多租户接缝和隔离模型

多租户能力通过宿主稳定接缝承载，而不是把具体租户主体硬编码进`lina-core`。宿主保留`tenantcap`接口、no-op 默认实现、租户过滤、当前租户读取和平台 bypass 判断；`multi-tenant`源码插件提供租户表、成员关系、解析器、Provider 和平台租户管理入口。这样未启用多租户时系统仍以`tenant_id=0`的 PLATFORM 语义开箱运行，启用后同一接缝变为租户隔离事实源。

隔离模型选择 Pool，即单库单 schema 加`tenant_id`列。租户敏感`sys_*`业务表、租户作用域运行时状态表和插件业务表必须带`tenant_id`，原查询索引升级为`(tenant_id, ...)`联合索引；`sys_locker`、`sys_menu`、`sys_plugin`、插件发布与迁移记录、插件资源引用和通知通道等平台控制面表保持全局。替代方案 schema-per-tenant 和 DB-per-tenant 会把连接路由和资源治理推入宿主，违背插件化目标，因此首版不做。

查询路径以`tenantcap.Apply(ctx, model, "tenant_id")`统一注入租户条件，并先于数据权限过滤执行。写路径由 service 根据`bizctx.TenantId`写入租户字段；平台管理员跨租户写必须使用 impersonation 或专用平台 API，不能在普通业务 API 中隐式跨租户写入。平台管理员只有在`TenantId=0`且具备全部数据权限的管理平台上下文才 bypass 租户过滤；impersonation 某租户时仍按目标租户过滤，只在审计中记录真实操作人。

## 租户主体、成员和权限模型

租户主体由`plugin_multi_tenant_tenant`维护，`code`全局唯一、不可修改、只允许`[a-z0-9-]{2,32}`，显示名可本地化。租户状态为`active`、`suspended`和软删除`deleted`；删除保留 tombstone，暂停只阻断成员登录和业务读写，不清理数据。

用户模型选择全局身份 + 1:N membership。`sys_user.tenant_id`只表示主租户或默认登录租户，真正的租户可见性由`plugin_multi_tenant_user_membership`决定。平台管理员是`tenant_id=0`平台上下文用户，不持有 membership；如果需要进入租户视角，必须使用 impersonation。

角色模型使用`sys_role.tenant_id`和四级数据范围表达平台/租户边界，不增加平台角色布尔字段，也不使用`platform:*`和`tenant:*`两套权限前缀。平台角色属于`tenant_id=0`上下文，可以配合全部数据权限和统一`system:*`功能权限执行平台治理；租户角色只能属于具体租户，禁止配置全部数据权限。用户角色关系也带`tenant_id`，角色授权树必须按当前上下文过滤可分配菜单和按钮，租户上下文不得分配平台租户管理、平台插件治理或全局菜单治理写权限。

## 租户解析和认证会话

租户解析链固定为`override`、JWT、session、header、subdomain、default。正式 JWT 的`TenantId`是普通业务请求的权威租户身份，header 和 subdomain 只作为登录前或`pre_token`阶段 hint，不得覆盖已认证 JWT。`X-Tenant-Override`只允许平台管理员在平台上下文使用，并进入 impersonation 审计语义。`rootDomain`首版固定为空，子域名解析保持禁用；解析链、保留子域和歧义策略都固定在代码中，不创建运行时解析配置表或共享 revision。

登录被拆成凭证校验和租户选择两个步骤。多 membership 用户先获得 60 秒、单次使用、仅限`select-tenant`的`pre_token`和候选租户列表，选择租户后才签发正式 JWT；单 membership 用户可自动选择。切换租户必须调用`/auth/switch-tenant`，服务端校验 membership、撤销旧 token、删除旧 session、签发新 token，并返回新菜单和权限。

平台管理员 impersonation 通过专用平台接口签发 token，token 中记录目标`tenant_id`、`is_impersonation`和`acting_user_id`。该 token 在业务链路中按目标租户过滤，不继承平台全局 bypass；登录日志、操作日志和审计载荷同时记录真实操作人和代操作租户。

`clientType`被确认为会话事实源。登录请求必须显式传入`web`、`mobile`、`desktop`或`cli`，后续登录、退出、租户 token 签发、租户切换、刷新和 impersonation 流程都复用该值。JWT、`pre_token`、`bizctx`、`sys_online_session`、Redis session hot state 和认证生命周期 Hook 必须保持一致，避免未来移动端、桌面端或 CLI 客户端在审计和插件 Hook 中被误判为默认 Web。

## 会话热状态和集群安全

集群模式下，请求热路径的会话有效性存储在 Redis，PostgreSQL `sys_online_session`保留为在线用户列表、数据权限过滤、登录信息展示和清理投影。Redis session key 至少包含租户和 token 维度，payload 包含用户、租户、`clientType`、登录时间、最后活跃时间和必要客户端上下文。

认证链读取 Redis session hot state、token revoke 状态和权限/配置 freshness 时遵循 fail-closed。Redis 不可读且超过允许窗口时，不得仅凭 JWT 签名放行业务请求。强制下线必须删除 Redis session hot key、写入 revoke 状态并删除或标记 PostgreSQL 投影；只删除投影不视为完成强退。

活跃时间写回 PostgreSQL 必须节流。Redis 维护请求热路径 last active 和 TTL，超过写回窗口后再更新投影，避免每个受保护请求都写数据库。投影清理任务保留，用于删除 Redis 已过期但 PostgreSQL 尚未清理的孤儿会话。

## 租户配置覆盖和数据隔离

字典和配置支持“平台默认 + 租户覆盖”。读取先查当前租户，未覆盖时 fallback 到`tenant_id=0`平台默认；写入默认作用于当前租户，平台默认只能由平台 API 或平台 service 显式写入。字典类型通过`allow_tenant_override`控制是否允许租户覆盖，fallback 行必须返回`sourceTenantId`、`isFallback`、`canEdit`、`canOverride`和`overrideMode`等动作元数据，前端不得把平台行伪装成可直接编辑的本租户记录。

文件存储按租户前缀隔离，本地路径使用`/storage/t/{tenant_id}/...`，对象存储 key 使用`t/{tenant_id}/...`。文件下载和预览接口校验请求租户与文件租户一致；平台管理员的跨租户只读访问必须走显式`/platform/*`接口，impersonation 模式不因真实身份是平台管理员而绕过目标租户过滤。

缓存 key 必须携带租户维度。租户覆盖变更只失效本租户对应缓存；平台默认变更按 fallback 影响范围级联失效。运行时翻译包缓存不增加租户分桶，因为本轮不落地租户级 i18n override；字典、配置、权限、角色、菜单和插件启用状态等业务缓存承担租户维度。

审计日志全面双轨化。登录日志和操作日志增加`tenant_id`、`acting_user_id`、`on_behalf_of_tenant_id`和`is_impersonation`；force-uninstall、impersonation 启用、install mode 切换和 LifecycleGuard 调用使用`oper_type='other'`并写入可检索摘要。

## 插件租户治理

插件治理从“平台安装即全局启用”演进为“平台管理员安装，租户管理员启用”的两层模型。`plugin.yaml`必须声明`scope_nature`：`platform_only`只能以`global`方式运行，租户管理员不可见；`tenant_aware`可由平台管理员选择`global`或`tenant_scoped`安装模式。`scope_nature`安装后不可运行时修改，只能随插件升级和迁移脚本变更。

`install_mode=global`时，插件启用状态由`sys_plugin_state(tenant_id=0, state_key='__tenant_enabled__')`表示，对所有租户生效；`tenant_scoped`时，每个租户持有独立启用状态，不存在状态行视为未启用。平台管理员维护`auto_enable_for_new_tenants`，该策略不来自`plugin.yaml`，只在插件支持多租户、已安装、已启用且为`tenant_scoped`时用于新租户开通。

租户管理员只能通过租户插件接口启用或禁用当前租户的`tenant_scoped`插件。插件路由可在启动期全局挂载，但请求时必须根据`(tenant_id, plugin_id)`启用状态过滤；未启用返回受控 404，不进入插件 handler。菜单、权限和运行时缓存按租户启用状态投影，启用/禁用后按插件和租户精细失效。

LifecycleGuard 让插件在卸载、禁用、租户禁用、租户删除和安装模式切换前执行自检。钩子并发调用、单钩子超时 5 秒、总超时 10 秒；任何否决都会阻断动作，但系统仍收集所有 reason。reason 必须是 i18n key，前端按当前语言渲染。平台管理员可按配置使用 force 通道，必须二次输入插件 ID，审计记录被绕过的 reason 和触发用户。

## 生命周期、启动和协同插件

租户生命周期不通过不完整的 outbox 或未实现事件总线表达。创建租户后，`multi-tenant`插件直接调用插件治理服务执行新租户默认启用策略；删除租户前先调用所有插件的`CanTenantDelete`，全部通过后才软删除租户主体。暂停和恢复只改变租户状态，不触发数据清理或异步事件分发。

启动期装配需要保证 Provider 和治理状态一致。启用`multi-tenant`插件时，插件必须注册`tenantcap.Provider`；未启用时装配 no-op service。启动一致性校验覆盖`scope_nature`与`install_mode`组合、租户角色数据范围、平台用户 membership 和 Provider 注册状态，失败时拒绝启动并输出明确诊断。

`org-center`协同改造为租户感知插件，部门、岗位和用户组织关系表带`tenant_id`并通过`orgcap.Provider`按当前租户过滤。当前版本不承诺通过`tenant.created`或`tenant.deleted`事件自动初始化或清理组织数据；需要跨插件清理时必须先设计可靠生命周期编排。

历史验证覆盖多租户基础、跨租户隔离矩阵、解析策略、租户切换、平台 impersonation、插件治理两层化、LifecycleGuard、集群缓存失效、token revoke 广播、联合索引`EXPLAIN`和单租户回归。`i18n`完整性、双语文档镜像、OpenSpec 严格校验和`lina-review`审查作为归档前治理证据保留。

## Cross-Domain Impacts

- `plugin-runtime-loading`、`plugin-manifest-lifecycle`、`plugin-cache-service`、`plugin-host-service-extension`、`plugin-storage-service`、`plugin-startup-bootstrap`和`plugin-permission-governance`受多租户启用状态、租户缓存 key、host service 租户透传和平台/租户插件治理影响；当前插件框架契约由`openspec/specs/<capability>/spec.md`承载，历史 owner 为`archive/plugin-framework`。
- `distributed-cache-coordination`为租户缓存失效、Redis revision/event、fail-closed 和 conservative-hide 提供基础；当前契约由`openspec/specs/distributed-cache-coordination/spec.md`承载，历史 owner 为`archive/distributed-infra`。`session-hot-state`在本分组保留认证会话历史 owner，分布式归档中的同名副本后续可裁剪。
- `config-management`、`dict-management`和`dictionary-management`只保留租户 fallback、动作元数据和租户缓存失效影响；当前契约由系统配置、字典和主规范承载，长期 owner 分别为`archive/system-config`或`archive/org-structure`。
- `user-management`、`user-role-association`和`role-management`受 membership、租户角色、数据范围和可分配权限集合影响；当前通用用户角色契约由`archive/user-management`及主规范承载，本分组只保留多租户身份边界。
- `dept-management`和`post-management`由组织源码插件交付，本分组只保留 org-center 租户感知、部门岗位过滤和未实现租户事件清理的决策；当前契约由`archive/org-structure`和主规范承载。
- `login-log`、`oper-log`和`online-user`受租户审计字段、impersonation 双轨、`clientType`投影和 Redis 会话热状态影响；当前管理能力契约由系统监控相关归档和主规范承载，认证会话事实源仍由本分组设计说明。
- `cron-job-management`受任务和任务分组租户隔离、系统任务平台级和执行时构造租户`bizctx`影响；当前契约由调度相关主规范和`archive/scheduled-jobs`承载。
- `notice-management`和`user-message`受消息、公告、投递和未读数的当前租户/当前用户边界影响；当前契约由`archive/notification`及主规范承载。
- `framework-i18n-runtime-performance`、`management-workbench-i18n`、`login-page-presentation`和`dashboard-workbench`只保留租户术语、登录挑选器、工作台租户切换和运行时翻译缓存边界；长期 owner 为`archive/i18n`和工作台主规范。
- `database-bootstrap-commands`、`project-setup`和`framework-bootstrap-installer`受`make db.init`/`make db.mock`写入 PLATFORM、开发入口、多租户启动校验和 PostgreSQL-only 初始化影响；当前契约由`archive/database-engine`、`archive/foundation`或开发工具主规范承载。
- `e2e-suite-organization`只保留多租户测试矩阵和插件自有测试目录归属影响；当前契约由`archive/e2e-testing`和测试主规范承载。
- `backend-conformance`、`core-host-boundary-governance`、`module-decoupling`和`spec-governance`提供 DAO 注入纪律、宿主接缝、可选插件降级和 OpenSpec 规则加载要求；当前契约由主规范和对应治理分组承载，本分组通过设计记录多租户相关约束。

## 角色权限树选择模式

角色抽屉维护独立的真实待提交选择状态；模式切换仅重绘展示勾选，不回写`menuIds`。打开编辑时可根据是否缺少祖先节点推断初始选择模式。提交时只包含真实选择的菜单/按钮 ID，不得因父子联动展示而扩大授权范围。
