## 为什么

LinaPro 需要从单租户后台基座演进为支持多业务单元、平台管理员穿透、租户管理员自治和未来 SaaS 形态的全栈框架。原有`sys_*`业务表没有租户维度，用户角色全局共享，JWT、会话、缓存、文件、审计和插件启用状态都默认以全局视角运行，无法保证租户之间的数据隔离、权限隔离和运行时状态隔离。

多租户能力必须与认证会话和插件治理同时落地。认证负责确定“当前用户以哪个租户身份访问系统”，插件治理负责确定“某插件在平台层还是租户层生效”，数据权限、缓存、文件和审计负责保证这个身份在后续链路中不被绕过。项目无历史负担，因此采用一次性完整落地的方式，避免分阶段兼容带来的模型撕裂。

压缩后的归档将`user-auth`作为多租户、租户感知认证、租户成员关系、租户配置覆盖、租户数据隔离、会话热状态和多租户插件治理的历史 owner。组织、日志、通知、调度、插件运行时、工作台、数据库引导和`E2E`等能力只保留交叉影响摘要，当前契约以`openspec/specs/<capability>/spec.md`为准。

## 变更内容

- 建立宿主多租户能力接缝：`tenantcap`提供 no-op 默认实现、Provider 注册、当前租户读取、租户过滤、平台 bypass 和可见性校验，具体租户主体由`multi-tenant`源码插件提供。
- 采用 Pool 隔离模型：租户敏感业务表、运行时状态表和插件业务表通过`tenant_id`隔离，平台控制面表保持全局；查询通过`tenantcap.Apply`注入租户过滤，写入由服务层显式填充租户。
- 引入`bizctx.TenantId`和固定租户解析链：`override`、JWT、session、header、subdomain、default 按代码固定顺序执行，正式 JWT 是业务请求的权威租户身份，header/subdomain 只作为登录前 hint。
- 重构认证流程：登录支持`pre_token`和租户选择，单租户用户可直接签发正式 JWT；切换租户必须重签 token 并撤销旧 token；平台管理员通过 impersonation token 进入目标租户视角并双轨审计。
- 建立 1:N 用户-租户成员模型：全局用户身份可以拥有多个 active membership；平台管理员不持有 membership；租户内管理权限由当前租户、角色数据范围和`system:*`功能权限组合推导。
- 将`clientType`作为会话事实源：登录请求必须显式传入`web`、`mobile`、`desktop`或`cli`，并写入 JWT、`pre_token`、`bizctx`、在线会话、Redis session hot state 和认证生命周期事件。
- 定义 Redis 会话热状态：集群模式下认证热路径以 Redis session hot key 和 revoke 状态为事实源，PostgreSQL `sys_online_session`保留为在线用户管理投影。
- 定义租户配置覆盖、文件路径、缓存 key、审计日志、用户消息、通知、任务和日志的租户隔离语义，跨租户访问只能通过显式平台 API 或 impersonation。
- 两层化插件治理：`scope_nature`表达插件作用域天性，`install_mode`表达平台全局或租户级启用模式，租户管理员只能治理当前租户的`tenant_scoped`插件启用状态。
- 引入`LifecycleGuard`否决型钩子：插件可在卸载、禁用、租户禁用、租户删除和安装模式切换前返回本地化 reason；平台管理员可按配置使用强制通道，但必须二次确认和审计。
- 修复角色权限树在“独立选择/父子联动”切换时重算并扩大授权提交的问题：仅按钮授权不得被前端额外提交父级或同级菜单。
- 明确不做的内容：不开放 schema-per-tenant 或 DB-per-tenant，不创建运行时租户解析配置表，不实现租户级配额/计费，不创建不完整的生命周期 outbox，不落地租户级 i18n override。

## Capabilities

### New Capabilities

- `multi-tenancy-foundation`
- `tenant-resolution`
- `tenant-aware-authentication`
- `tenant-management`
- `tenant-membership`
- `tenant-config-override`
- `tenant-data-isolation`
- `tenant-lifecycle-events`
- `session-hot-state`
- `plugin-scope-nature`
- `plugin-install-mode`
- `plugin-tenant-enablement`
- `plugin-lifecycle-guard`

### Modified Capabilities

- `user-auth`：登录、登出、租户选择、租户切换、token 撤销、impersonation 和认证事件全部携带租户与`clientType`上下文。
- `user-management`、`user-role-association`和`role-management`：用户、角色和用户角色关系按租户上下文隔离，平台角色和租户角色使用`tenant_id`及数据范围区分。
- `dict-management`、`dictionary-management`和`config-management`：支持“平台默认 + 租户覆盖”的读取、写入、动作元数据和缓存失效语义。
- `online-user`、`login-log`和`oper-log`：会话、登录日志和操作日志投影携带租户、impersonation 和`clientType`字段。
- `cron-job-management`、`notice-management`和`user-message`：任务、通知公告和用户消息按当前租户隔离。
- `plugin-manifest-lifecycle`、`plugin-runtime-loading`、`plugin-permission-governance`、`plugin-cache-service`、`plugin-host-service-extension`、`plugin-storage-service`和`plugin-startup-bootstrap`：只保留多租户治理交叉影响，长期插件框架契约由`archive/plugin-framework`和主规范承载。

## Impact

- 数据影响：宿主租户敏感表、插件业务表和运行时状态表增加`tenant_id`与联合索引；平台控制面表保持全局；初始化与 mock 数据默认写入`tenant_id=0`。
- 后端影响：新增`tenantcap`、tenancy 中间件、多租户插件、租户感知认证、session hot state、插件租户治理、LifecycleGuard、缓存与审计隔离。
- API 影响：新增登录租户选择、切换租户、平台租户管理、impersonation、租户插件启用和租户成员管理等接口；既有业务接口主要从上下文读取租户，不把租户字段暴露为普通 DTO 参数。
- 前端影响：登录页租户选择、工作台租户切换、平台租户管理、impersonation 提示、插件安装模式选择和 LifecycleGuard reason 展示。
- `i18n`影响：新增租户、平台管理员、租户管理员、否决理由、impersonation 和登录租户选择等中英文资源；运行时翻译包缓存不增加租户分桶。
- 测试影响：历史实现覆盖单元测试、多租户 E2E、跨租户隔离矩阵、解析链、插件治理、平台管理员场景、集群一致性、`EXPLAIN`性能验证、`i18n`完整性和 OpenSpec 校验。
- 本归档压缩不修改运行时代码、数据库、HTTP API、前端页面、插件源码或生产构建；仅保留历史设计、语义覆盖和验证证据。
