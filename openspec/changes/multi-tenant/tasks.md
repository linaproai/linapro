## 1. Schema 与初始化

- [x] 1.1 在各宿主表对应的源建表 SQL 中直接定义 `tenant_id INT NOT NULL DEFAULT 0`,索引升级为 `(tenant_id, ...)` 联合索引(幂等);`016` 仅保留说明,不承载迁移式修补或 seed
- [x] 1.2 在 `008-menu-role-management.sql` 的 `sys_role` 建表中直接定义 `tenant_id` 与四级 `data_scope`(`1=全部,2=租户,3=部门,4=个人`),不维护平台角色布尔字段
- [x] 1.3 在 `008-menu-role-management.sql` 的 `sys_user_role` 建表中直接定义 `tenant_id INT NOT NULL DEFAULT 0`,主键改为 `(user_id, role_id, tenant_id)`
- [x] 1.4 在 `002-dict-type-data.sql` 的 `sys_dict_type` 建表中直接定义 `allow_tenant_override BOOL NOT NULL DEFAULT FALSE`
- [x] 1.5 在 `011-plugin-framework.sql` 的 `sys_plugin` 建表中直接定义 `scope_nature VARCHAR(32) NOT NULL DEFAULT 'tenant_aware'`、`install_mode VARCHAR(32) NOT NULL DEFAULT 'global'` 与 `auto_enable_for_new_tenants BOOL NOT NULL DEFAULT FALSE`
- [x] 1.6 在 `012-plugin-host-call.sql` 的 `sys_plugin_state` 建表中直接定义 `tenant_id INT NOT NULL DEFAULT 0`,保留 `id` 自增技术主键,使用 `(plugin_id, tenant_id, state_key)` 唯一索引表达业务唯一性,插件启用状态使用 `state_key='__tenant_enabled__'`
- [x] 1.7 在 `monitor-loginlog`、`monitor-operlog` 插件 schema 增加 `acting_user_id`、`on_behalf_of_tenant_id`、`is_impersonation` 字段(改其 manifest/sql/001-*)
- [x] 1.8 在所有插件业务表 (`plugin_*`) 中加 `tenant_id INT NOT NULL DEFAULT 0`,索引升级
- [x] 1.9 seed 数据:平台超级管理员角色(`tenant_id=0, data_scope=1`)、admin 用户绑定平台超管角色;现有 seed 字典/配置统一标记 `tenant_id=0`
- [x] 1.10 修改现有 mock-data SQL,默认写入 `tenant_id=0`,确保 `make mock` 单租户开箱体验不变
- [x] 1.11 执行 `make init` 重建数据库,验证所有表结构与索引正确
- [x] 1.12 执行 `make mock` 验证 mock 数据加载成功且单租户行为不破坏

## 2. 宿主稳定接缝(pkg/tenantcap)

- [x] 2.1 创建 `apps/lina-core/pkg/tenantcap/` 包,定义 `Provider` 接口、`TenantID` 类型、`PLATFORM` 常量、`RegisterProvider`/`CurrentProvider`/`HasProvider`(参考 pkg/orgcap 形态)
- [x] 2.2 已评估并移除未完整实现的租户生命周期事件接口;`pkg/tenantcap` 仅保留稳定 Provider、Resolver 与租户投影契约
- [x] 2.3 在 `pkg/tenantcap/` 内定义 `TenantInfo` 结构(id, code, name, status)与 `Resolver` 子接口

## 3. 宿主侧 tenantcap 服务

- [x] 3.1 创建 `apps/lina-core/internal/service/tenantcap/tenantcap.go` 主文件:`Service` 接口、`serviceImpl`、`New()`,实现 `Enabled`、`Apply`、`Current`、`PlatformBypass`、`EnsureTenantVisible`
- [x] 3.2 在 `apps/lina-core/pkg/tenantcap/tenantcap_code.go` 中集中定义共享 `bizerr.Code`(`CodeTenantRequired`、`CodeTenantForbidden`、`CodeCrossTenantNotAllowed`、`CodePlatformPermissionRequired`、`CodeTenantSuspended` 等),宿主内部服务直接引用共享契约错误码
- [x] 3.3 在 `tenantcap_resolver_chain.go` 中实现责任链调度器,支持注册多个 `Resolver` 与按配置顺序遍历
- [x] 3.4 在 `tenantcap_fallback.go` 中实现 `ReadWithPlatformFallback` helper(应用层两次查询合并),供字典/配置等"租户覆盖"资源使用
- [x] 3.5 增补单元测试(自包含、顺序无关):覆盖 `Apply` 注入、`PlatformBypass`、责任链顺序、fallback 语义

## 4. bizctx 与中间件

- [x] 4.1 在 `internal/service/bizctx/bizctx.go` 增加 `TenantId int`、`ActingAsTenant bool`、`ActingUserId int`、`IsImpersonation bool` 字段;增加 `SetTenant`、`SetImpersonation` 方法
- [x] 4.2 创建 `internal/service/middleware/middleware_tenancy.go`,实现 tenancy 中间件:在 auth 之后、权限校验之前调用 `tenantcap.Provider.ResolveTenant` 并写入 bizctx
- [x] 4.3 在 `cmd` 启动期注册 tenancy 中间件;multi-tenant 未启用时短路注入 `TenantId=0`
- [x] 4.4 单元测试:覆盖中间件注入路径、短路路径、resolver 链选用

## 5. 启动期一致性校验

- [x] 5.1 在 `framework-bootstrap-installer` 启动流程中增加一致性校验:`scope_nature` ↔ `install_mode`、租户角色不得使用 `data_scope=1`、平台用户无 membership、multi-tenant enabled ↔ Provider 注册
- [x] 5.2 检查失败时打印明确日志并 panic 阻止启动
- [x] 5.3 集成测试:故意构造非法状态验证启动失败

## 6. 多租户插件骨架(lina-plugin-multi-tenant)

- [x] 6.1 创建 `apps/lina-plugins/multi-tenant/` 目录,包含 `plugin.yaml`(scope_nature=platform_only、install_mode=global)、`plugin_embed.go`、`backend/`、`frontend/`、`manifest/`
- [x] 6.2 在 `apps/lina-plugins/lina-plugins.go` 中注册 `_ "lina-plugin-multi-tenant/backend"`
- [x] 6.3 在 `backend/` 下建立 `api/`、`internal/{controller, service, dao, model/{do, entity}}`、`hack/config.yaml`、`plugin.go`
- [x] 6.4 创建 `manifest/sql/001-multi-tenant-schema.sql`(所有插件持有的表)与 `manifest/sql/uninstall/`
- [x] 6.5 创建 `manifest/i18n/zh-CN/plugin.json` 与 `manifest/i18n/en-US/plugin.json` 的占位骨架
- [x] 6.6 plugin.yaml 列出菜单(平台管理员侧:tenants)与隐藏 permission 点(统一 `system:tenant:*`、`system:tenant:member:*`、`system:tenant:plugin:*` 等);租户成员关系由用户管理页面承载,不提供独立租户工作台目录

## 7. 多租户插件 schema

- [x] 7.1 `plugin_multi_tenant_tenant`:id、code(unique)、name、status、remark、created_at、updated_at、deleted_at(软删)
- [x] 7.2 `plugin_multi_tenant_user_membership`:id、user_id、tenant_id、status、joined_at,UNIQUE (user_id, tenant_id)
- [x] 7.3 `plugin_multi_tenant_config_override`:占位扩展,首版可不创建或仅创建表骨架
- [x] 7.4 本轮不创建租户配额表或占位执行逻辑,后续迭代重新设计配额/计费模型
- [x] 7.5 删除 `plugin_multi_tenant_resolver_config`:解析链、保留子域和 ambiguous 行为固定在代码中
- [x] 7.6 删除 `plugin_multi_tenant_event_outbox`:不保留缺少订阅、分发、重试和 per-subscriber 状态的 outbox 占位表
- [x] 7.7 `manifest/sql/uninstall/001-cleanup.sql`:卸载时清理插件表(前置条件由 LifecycleGuard 保证)
- [x] 7.8 `dao/`、`model/{do,entity}/` 通过 `cd apps/lina-plugins/multi-tenant/backend && make dao` 生成

## 8. 多租户插件 service 层

- [x] 8.1 `service/tenant/`:`Service` 接口、`serviceImpl`,实现租户 CRUD、状态机迁移、code 唯一性校验、tombstone 30 天保留
- [x] 8.2 `service/membership/`:1:N 成员关系 CRUD、预留 `single` 策略校验、平台管理员不允许加 membership
- [x] 8.3 `service/resolver/`:实现 6 个 Resolver(override/jwt/session/header/subdomain/default),支持配置驱动的链;正式 JWT 优先于 header/subdomain hint
- [x] 8.4 删除未完整实现的 `service/lifecycle/` 事件 outbox 链路,租户创建副作用改为显式领域服务调用
- [x] 8.5 `service/provider/`:`tenantcap.Provider` 实现,聚合 tenant + membership service,在插件 enable 时调 `tenantcap.RegisterProvider`
- [x] 8.6 `service/lifecycleguard/`:实现 `CanUninstall`(active 租户存在则 false)、`CanDisable`、`CanTenantDelete`(订阅其他插件钩子聚合)
- [x] 8.7 `service/impersonate/`:平台管理员 impersonation token 签发、结束 impersonation、双轨日志写入
- [x] 8.8 `service/resolverconfig/`:提供代码内置解析策略查询与 no-op 校验,拒绝运行时策略变更
- [x] 8.9 service 层单元测试自包含覆盖

## 9. 多租户插件 controller 层

- [x] 9.1 `controller/platform/`:`/platform/tenants/*`(CRUD)、`/platform/tenants/{id}/impersonate`、`/platform/tenant/resolver-config` 策略查询/校验、`/platform/users`
- [x] 9.2 `controller/tenant/`:`/tenant/members/*`(列表、邀请、移除、调整角色)、`/tenant/members/me`
- [x] 9.3 `controller/auth/`:`/auth/login-tenants`、`/auth/select-tenant`、`/auth/switch-tenant`(覆盖宿主原 login)
- [x] 9.4 controller 通过 `cd apps/lina-plugins/multi-tenant/backend && make ctrl` 生成骨架,业务逻辑写在生成文件内
- [x] 9.5 所有 controller 字段持有依赖 service,在 NewV1 中初始化(禁止方法内 service.New())

## 10. JWT / 会话 / 鉴权改造

- [x] 10.1 在 `auth.Claims` 增加 `TenantId int`、`IsImpersonation bool`、`ActingUserId int` 字段;调整签发与解析逻辑
- [x] 10.2 修改 `auth.Service.Login` 为两阶段:返回 pre_token + 租户列表(单租户用户走兼容直接签发)
- [x] 10.3 实现 pre_token 短期单次使用机制(60s TTL,放在 redis/session store)
- [x] 10.4 实现 `/auth/select-tenant`:校验 pre_token + membership → 签发正式 JWT
- [x] 10.5 实现 `/auth/switch-tenant`:校验 membership → 旧 token 加入 revoke 列表 → 重签
- [x] 10.6 实现 token revoke 列表(本地内存 + cluster 广播)
- [x] 10.7 修改 `session.Store`:主键为全局唯一 `token_id`,保留 `tenant_id` 作为会话归属与校验维度,index 覆盖 `(tenant_id, user_id)` 与 `(tenant_id, login_time)`
- [x] 10.8 修改 `auth.Logout`:仅撤销当前 (tenant, token) 行
- [x] 10.9 单元测试覆盖 Login → SelectTenant → SwitchTenant → Logout 全链路

## 11. DAO 注入纪律落地

- [x] 11.1 在 `internal/service/user/` 改造用户 service:多租户启用时列表/详情以 membership join 作为可见性权威边界,`sys_user.tenant_id` 仅作为主租户/默认登录租户;写时填主租户并创建 membership
- [x] 11.2 在 `internal/service/role/` 改造所有读 `sys_role` 与 `sys_user_role`;角色查询按当前租户上下文与 `tenant_id` 过滤,并用 `data_scope` 表达数据边界
- [x] 11.3 在 `internal/service/menu/` 改造,菜单仍为平台全局,但解析时按 (tenant, plugin) 启用状态过滤插件菜单
- [x] 11.4 在 `internal/service/dict/` 改造为 `ReadWithPlatformFallback` 模式;写入按当前租户;`allow_tenant_override` 校验
- [x] 11.5 在 `internal/service/sysconfig/` 同上
- [x] 11.6 在 `internal/service/file/` 改造存储路径前缀 `/storage/t/{tid}/...`;读取校验
- [x] 11.7 在 `internal/service/notify/` 改造通知发送/查询,跨租户广播仅平台
- [x] 11.8 在 `internal/service/usermsg/` 改造消息 inbox 按租户过滤
- [x] 11.9 在 `internal/service/jobmgmt/` `jobmeta/` `jobhandler/` 改造,任务执行前构造租户 bizctx
- [x] 11.10 在 `internal/service/session/` 与 `internal/service/cron/` 改造,session 清理与定时任务上下文绑定 tenant
- [x] 11.11 在 `internal/service/datascope/` 中调整 Apply 顺序:先 tenantcap 后 datascope
- [x] 11.12 全宿主代码 grep `dao.Sys*\.Ctx(ctx)` 确保无散点遗漏

## 12. 缓存与一致性

- [x] 12.1 在 `internal/service/kvcache/`、`internal/service/pluginruntimecache/`、`internal/service/cachecoord/` 改造 cache key,引入 `(tenant_id, scope, key)` 形态
- [x] 12.2 创建 `tenantcap.CacheKey(tenant, scope, key)` helper,所有缓存使用方统一调用
- [x] 12.3 失效消息 schema 增加 `tenant_id` 字段;支持 `cascade_to_tenants` 标志
- [x] 12.4 `cluster.Service` 集群广播链路透传 tenant 维度
- [x] 12.5 翻译缓存(framework-i18n-runtime)key 与失效作用域改造
- [x] 12.6 字典缓存、配置缓存、角色缓存、权限缓存、菜单缓存全部按租户分桶
- [x] 12.7 单元测试:跨租户缓存隔离反例、平台默认级联失效

## 13. 插件治理改造

- [x] 13.1 在 `internal/service/plugin/` 中增加 `scope_nature` / `install_mode` 解析与持久化,plugin.yaml 解析期校验
- [x] 13.2 改造 `IsEnabled(ctx, pluginID)` 为租户感知(读 `(plugin_id, tenant_id)`)
- [x] 13.3 实现 install_mode 切换流程(global ↔ tenant_scoped)与 force 通道
- [x] 13.4 实现 `LifecycleGuard` 接口族(`pkg/pluginhost/lifecycle_guard.go`)与并发调用 / 超时 / panic recover 框架
- [x] 13.5 在 `plugin.Uninstall` 流程中调用 `CanUninstall` 钩子并聚合否决
- [x] 13.6 在 `plugin.Disable` 流程中调用 `CanDisable` 钩子(global)或 `CanTenantDisable`(tenant_scoped)
- [x] 13.7 在 `tenant.Delete` 流程中调用 `CanTenantDelete` 钩子
- [x] 13.8 实现 `--force` 通道(配置开关 `plugin.allow_force_uninstall`)与平台审计
- [x] 13.9 单元测试:钩子聚合、超时 fail-safe、panic recover、force 通道

## 14. 路由与权限改造

- [x] 14.1 改造 `pluginruntimecache` / `plugin.routing` 的请求时过滤逻辑:按 (tenant, plugin) 启用状态返回 404
- [x] 14.2 改造 `permission` 中间件:权限解析按当前租户过滤未启用插件
- [x] 14.3 移除权限点平台/租户前缀约束,统一使用 `system:*`;平台/租户边界由路由平面、租户上下文、数据权限与插件启用状态约束
- [x] 14.4 单元测试:覆盖路由 404 / 菜单隐藏 / 权限点过滤

## 15. org-center 插件租户化改造

- [x] 15.1 修改 `apps/lina-plugins/org-center/manifest/sql/001-org-center-schema.sql`,所有表加 `tenant_id`,索引升级
- [x] 15.2 修改 mock-data/*.sql,默认写入 `tenant_id=0`(单租户场景体验不变)
- [x] 15.3 修改 `org-center` plugin.yaml:`scope_nature: tenant_aware`、`default_install_mode: global`
- [x] 15.4 改造 service/dao 实现:所有查询走 `tenantcap.Apply`,所有写入填 `tenant_id`
- [x] 15.5 移除未接入的 `tenantcap.LifecycleSubscriber` 占位契约,不承诺跨插件租户事件订阅
- [x] 15.6 改造 `orgcap.Provider` 实现按 `bizctx.TenantId` 过滤
- [x] 15.7 单元测试自包含,覆盖租户隔离、事件订阅幂等、级联清理

## 16. 其他既有插件租户化改造

- [x] 16.1 `monitor-loginlog`:表加 `tenant_id`、`acting_user_id`、`on_behalf_of_tenant_id`、`is_impersonation`;dao/service/controller 改造;plugin.yaml `scope_nature: tenant_aware, default_install_mode: tenant_scoped`
- [x] 16.2 `monitor-online`:表加 `tenant_id`,会话查询/踢人接口加租户校验;plugin.yaml 同上
- [x] 16.3 `monitor-operlog`:表加 tenant 与 impersonation 字段;controller 支持 `operType` 筛选;plugin.yaml 同上
- [x] 16.4 `monitor-server`:plugin.yaml `scope_nature: platform_only`;租户管理员视图不可见
- [x] 16.5 `content-notice`:表加 `tenant_id`,通知按租户隔离;plugin.yaml `scope_nature: tenant_aware, default_install_mode: tenant_scoped`
- [x] 16.6 `demo-control`、`plugin-demo-source`、`plugin-demo-dynamic`:加 tenant 字段(若有持久化);plugin.yaml `scope_nature: tenant_aware`
- [x] 16.7 各插件 i18n 资源新增 `plugin.<id>.uninstall_blocked.*` 等 reason key 翻译

## 17. 后端 API 与 DTO

- [x] 17.1 所有平台 API DTO 在 `apps/lina-plugins/multi-tenant/backend/api/platform/v1/` 定义,英文 dc + eg
- [x] 17.2 所有租户 API DTO 在 `apps/lina-plugins/multi-tenant/backend/api/tenant/v1/` 定义
- [x] 17.3 在 g.Meta 上声明 `permission` 标签,平台 API 与租户 API 均使用统一 `system:*` 权限标识
- [x] 17.4 apidoc i18n:多租户插件维护自己的 `manifest/i18n/<locale>/apidoc/**/*.json`,英文置空,zh-CN 提供翻译
- [x] 17.5 错误码 `bizerr.Code*` 在 `internal/service/<module>/<module>_code.go` 集中定义
- [x] 17.6 单元测试:DTO 字段验证、permission 校验、i18n 完整性

## 18. 前端 工作台改造

- [x] 18.1 在 `apps/lina-vben/apps/web-antd/src/api/` 增加平台/租户/auth 新接口客户端
- [x] 18.2 在 `src/views/` 下增加 `platform/tenants/`、`platform/users/` 等页面,不提供解析策略运行时配置页面
- [x] 18.3 租户管理员入口回收到用户管理页面,不再暴露 `tenant/members/`、`tenant/plugins/` 左侧菜单入口
- [x] 18.4 改造登录页 `views/login/` 支持租户挑选器(基于解析策略呈现)
- [x] 18.5 工作台头部增加租户标识 + 切换器(平台 vs 租户视觉区分)
- [x] 18.6 改造路由守卫,根据 multi-tenant 启用状态联动隐藏租户 UI
- [x] 18.7 增加 impersonation 头部红条提示与"退出 impersonation"按钮
- [x] 18.8 增加 `views/platform/plugins/install-mode-selector.vue` 安装时选择 install_mode 的对话框
- [x] 18.9 增加 LifecycleGuard 否决理由展示对话框(支持多 reason 聚合 + i18n 渲染 + force 二次确认)
- [x] 18.10 路由模块 `src/router/routes/modules/platform.ts`、`tenant.ts` 注册
- [x] 18.11 全局 Pinia store 增加 `useTenantStore` 管理当前租户、可选租户列表、是否 impersonation

## 19. 前端 i18n 资源

- [x] 19.1 在 `apps/lina-vben/apps/web-antd/src/locales/zh-CN.json` 与 `en-US.json` 增加多租户相关 key(菜单、表单、错误、否决理由、impersonation 提示)
- [x] 19.2 平台多租户菜单 i18n key 命名空间保留 `menu.platform.*`,并移除独立租户工作台 `menu.tenant.*` 顶级菜单命名空间
- [x] 19.3 i18n 完整性校验脚本(CI)阻断遗漏

## 20. 配置文件与文档

- [x] 20.1 租户默认策略在代码中集中定义;宿主配置模板不提供 `tenant.*`,不创建或读取运行时解析配置表
- [x] 20.2 `plugin.allow_force_uninstall: true` 增加配置项与文档
- [x] 20.3 更新根 `README.md` 与 `README.zh-CN.md` 增加多租户章节(启用步骤、解析策略、典型场景、限制)
- [x] 20.4 在 `docs/` 或对应位置增加插件作者指南:scope_nature、install_mode、LifecycleGuard、tenant 事件订阅
- [x] 20.5 在 `apps/lina-plugins/multi-tenant/README.md` 与 `README.zh-CN.md` 双语镜像

## 21. 单元测试

- [x] 21.1 tenantcap 单元测试自包含
- [x] 21.2 bizctx 单元测试
- [x] 21.3 tenancy 中间件单元测试
- [x] 21.4 解析责任链单元测试(每种 resolver 一组用例)
- [x] 21.5 LifecycleGuard 钩子框架单元测试(聚合、超时、panic、force)
- [x] 21.6 多租户插件 service 层单元测试自包含,覆盖各分支
- [x] 21.7 org-center 改造的单元测试自包含
- [x] 21.8 缓存隔离单元测试(集群模式 mock)

## 22. E2E 测试 — 多租户基础

- [x] 22.1 创建 `hack/tests/e2e/multi-tenant/` 目录与 fixture(`multi-tenant-disabled` 与 `multi-tenant-enabled`)
- [x] 22.2 TC0178 多租户启用:平台管理员安装 multi-tenant 插件,启用后系统行为正确
- [x] 22.3 TC0179 平台管理员创建租户:基础 CRUD、code 校验、tombstone
- [x] 22.4 TC0180 租户暂停/恢复:暂停后用户登录被拒,恢复后正常
- [x] 22.5 TC0181 租户管理页面不暴露归档入口
- [x] 22.6 TC0182 租户删除受 LifecycleGuard 保护:guard 拒绝时阻断,通过后允许 active/suspended 租户直接删除
- [x] 22.7 TC0183 多租户禁用:卸载 multi-tenant 后系统退化为单租户行为
- [x] 22.8 TC0184 1:N 用户登录挑选租户:返回 pre_token + 列表 + select-tenant 流程
- [x] 22.9 TC0185 切换租户重签 token:旧 token 立即作废
- [x] 22.10 TC0186 平台管理员 impersonation 启用与退出

## 23. E2E 测试 — 跨租户隔离矩阵

- [x] 23.1 TC0187 用户 跨租户隔离:租户 A 不可见 B 用户
- [x] 23.2 TC0188 角色 跨租户隔离 + 平台角色仅平台管理员可见
- [x] 23.3 TC0189 字典 跨租户隔离 + platform fallback 正确
- [x] 23.4 TC0190 配置 跨租户隔离 + platform fallback 正确
- [x] 23.5 TC0191 文件 跨租户隔离 + 平台共享路径正确
- [x] 23.6 TC0192 通知 跨租户隔离 + 平台广播正确
- [x] 23.7 TC0193 任务 跨租户隔离 + 内置系统任务平台级
- [x] 23.8 TC0194 在线会话 跨租户隔离 + 踢人按租户校验
- [x] 23.9 TC0195 登录日志/操作日志 跨租户隔离 + impersonation 双轨标记
- [x] 23.10 TC0196 部门(org-center)跨租户隔离 + 租户事件触发部门初始化
- [x] 23.11 TC0197 岗位(org-center)跨租户隔离 + 同 code 跨租户允许

## 24. E2E 测试 — 解析策略与登录流程

- [x] 24.1 TC0198 header 解析器:登录前 X-Tenant-Code hint 命中,已登录业务请求不得覆盖 JWT TenantId
- [x] 24.2 TC0199 subdomain 解析器:登录前子域名 hint 命中 + 保留子域名忽略
- [x] 24.3 TC0200 jwt 解析器:Claims 命中
- [x] 24.4 TC0201 session 解析器:登录后挑选 + 持续命中
- [x] 24.5 TC0202 default 解析器 + ambiguous prompt 模式
- [x] 24.6 TC0203 固定 prompt 歧义策略,拒绝运行时切换 reject
- [x] 24.7 TC0204 override:平台管理员 X-Tenant-Override 合法 + 普通用户被忽略
- [x] 24.8 TC0205 解析链固定策略,内置策略 no-op 写入不改变运行时状态

## 25. E2E 测试 — 插件治理两层化

- [x] 25.1 TC0206 平台管理员安装 tenant_aware 插件:install_mode 选择 global / tenant_scoped
- [x] 25.2 TC0207 安装 platform_only 插件:install_mode 强制 global,租户管理员看不到
- [x] 25.3 TC0208 租户管理员启用/禁用 tenant_scoped 插件:菜单/路由/权限联动
- [x] 25.4 TC0209 install_mode 切换 global ↔ tenant_scoped:状态正确迁移
- [x] 25.5 TC0210 LifecycleGuard 否决卸载:active 租户时 multi-tenant 拒绝卸载,聚合 reason
- [x] 25.6 TC0211 force 通道:平台管理员强制卸载 + 双重确认 + 强审计
- [x] 25.7 TC0212 钩子超时 / panic fail-safe
- [x] 25.8 TC0213 平台新租户自动启用策略在新租户创建时写入租户插件状态

## 26. E2E 测试 — 平台管理员场景

- [x] 26.1 TC0214 平台管理员可跨租户读全量数据
- [x] 26.2 TC0215 平台管理员 impersonation 后操作日志双轨记录
- [x] 26.3 TC0216 平台管理员强制操作单独审计类
- [x] 26.4 TC0217 平台管理员视图与租户管理员视图 UI 差异化

## 27. 集群一致性 E2E

- [x] 27.1 TC0218 拒绝解析策略运行时变更,不创建 tenant-resolution shared revision
- [x] 27.2 TC0219 跨节点缓存按租户失效隔离
- [x] 27.3 TC0220 token revoke 跨节点广播

## 28. 性能与回归

- [x] 28.1 联合索引性能验证:用 PG `EXPLAIN` 验证关键查询(用户列表、角色列表、菜单解析、字典 fallback)
- [x] 28.2 现有单租户 e2e 套件在 multi-tenant 未启用 fixture 下全部通过
- [x] 28.3 启动期一致性校验通过基线测试

## 29. 审查与归档准备

- [x] 29.1 调用 `/lina-review` 进行 spec 与代码审查,处理审查项
- [x] 29.2 i18n 完整性校验(zh-CN / en-US)无缺失
- [x] 29.3 文档双语镜像同步(README、AGENTS、SKILL 等)
- [x] 29.4 `openspec validate multi-tenant --strict` 通过
- [ ] 29.5 用户验收确认本迭代完成

## Verification

- [x] 2026-05-10: `make init confirm=init rebuild=true` 通过,宿主表结构由各源建表 SQL 初始化;`016-multi-tenant-and-plugin-governance.sql` 为说明文件且不包含可执行 SQL。
- [x] 2026-05-10: `make mock confirm=mock` 通过,宿主 mock 数据加载完成。
- [x] 2026-05-10: 后端单元测试通过:`cd apps/lina-core && go test ./... -count=1`;插件单元测试通过:`for mod in apps/lina-plugins/*/go.mod; do dir=$(dirname "$mod"); (cd "$dir" && go test ./... -count=1) || exit 1; done`。
- [x] 2026-05-10: 前端验证通过:`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`pnpm -F @lina/web-antd i18n:check`;`pnpm run build`。
- [x] 2026-05-10: E2E 静态验证通过:`cd hack/tests && pnpm test:validate && pnpm exec tsc --noEmit --pretty false`。
- [x] 2026-05-10: 完整 E2E 通过:`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ E2E_PARALLEL_WORKERS=2 pnpm test`;结果为并行段 `110 passed, 2 skipped`,串行段 `408 passed, 7 skipped`。
- [x] 2026-05-10: PG `EXPLAIN` 验证关键查询,用户 membership 查询命中 `idx_plugin_multi_tenant_membership_tenant`,菜单插件启用子查询命中 `idx_sys_plugin_state_tenant_enabled`;角色/字典在当前 mock 小表下成本低。
- [x] 2026-05-10: `/lina-review` 只读审查通过:未发现违规直接 `g.Log`、插件引用宿主 `internal`、multi-tenant 非英文 apidoc 空占位或 SQL 裸自增 id 写入;`Data(g.Map)` 命中仅为既有 locker/cluster 测试。
- [x] 2026-05-10: `openspec validate multi-tenant --strict` 通过。
- [x] 2026-05-10: 本次 multi-tenant 相关文件 `git diff --check` 通过;全局检查曾受无关既有文件 EOF 空行影响。
- [x] 2026-05-10: FB-7 验证通过:`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts --project=chromium`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/tenant-lifecycle/TC0178-multi-tenant-enabled.ts --project=chromium`;`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate && pnpm exec tsc --noEmit --pretty false`;`openspec validate multi-tenant --strict`。
- [x] 2026-05-10: FB-8 验证通过:`rg -n "func \\(c \\*ControllerV1\\) [A-Z]" apps/lina-core/internal/controller apps/lina-plugins -g '*.go' | awk -F: '{count[$1]++} END {for (file in count) if (count[file] > 1) print count[file], file}' | sort -nr` 无输出;`cd apps/lina-core && go test ./internal/controller/auth ./internal/controller/dict -count=1`;`cd apps/lina-plugins/multi-tenant && go test ./backend/internal/controller/... -count=1`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-core/internal/controller/auth apps/lina-core/internal/controller/dict apps/lina-plugins/multi-tenant/backend/internal/controller openspec/changes/multi-tenant/tasks.md`。
- [x] 2026-05-10: FB-9~FB-15 验证通过:`cd apps/lina-core && go test ./internal/service/user ./internal/controller/user ./api/user ./api/auth ./internal/controller/auth ./internal/cmd ./internal/service/middleware ./pkg/pluginservice/auth -count=1`;`cd apps/lina-plugins/multi-tenant && go test ./... -count=1`;`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate && pnpm exec tsc --noEmit --pretty false`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- <本次反馈相关文件>`。全局 `git diff --check` 仍受无关已修改文件 `openspec/specs/project-setup/spec.md` 与 `openspec/specs/readme-localization-governance/spec.md` EOF 空行影响,本次未修改。
- [x] 2026-05-10: FB-12 追加验证通过:租户切换器移至 `header-right-40`,位于全局搜索(index 50)左侧;`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts --project=chromium`。
- [x] 2026-05-10: FB-10 数据边界追加验证通过:`cd apps/lina-core && go test ./internal/service/user -run 'TestListUserTenantMemberships(AggregatesTenantNames|RespectsTenantContext)' -count=3`,确认平台上下文显示用户全部租户,租户上下文只显示当前租户归属。
- [x] 2026-05-10: FB-16 验证通过:`make init confirm=init rebuild=true`;`make mock confirm=mock`;`cd apps/lina-core && go test ./internal/cmd -count=1`;`cd apps/lina-core && go test ./pkg/dialect -count=1`;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-16 touched files>`;静态扫描确认宿主 SQL 不再包含 `ALTER TABLE sys_*`/`DROP INDEX`/`DROP COLUMN`/`ADD COLUMN`/`ADD CONSTRAINT`/`UPDATE sys_*` 迁移式修补,源 `manifest/sql` 与 `internal/packed/manifest/sql` 目录一致,`016` 非注释可执行 SQL 扫描无命中,源 SQL 中 `sys_role_menu` 主键为 `(role_id, menu_id, tenant_id)` 且 `platform`/`tenant` 新目录已回收到 `008-menu-role-management.sql`。
- [x] 2026-05-10: FB-17/FB-18 验证通过:`cd apps/lina-core && go test ./internal/service/auth -count=1`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts --project=chromium`;`cd hack/tests && pnpm test:validate && pnpm exec tsc --noEmit --pretty false`;`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`rg` 确认无旧 `tenant-switcher-trigger`/`tenant-switch-option-*` 定位器和裸 `gerror.Wrap(err, "error.auth.login.updateLastLoginFailed")`。
- [x] 2026-05-10: FB-19 验证通过:`make init confirm=init rebuild=true`;`make mock confirm=mock`;`cd apps/lina-core && go test ./internal/service/plugin/... ./internal/service/notify ./internal/service/locker ./internal/service/cluster ./internal/service/role ./pkg/dialect ./internal/packed -count=1`;`cd apps/lina-plugins/multi-tenant && go test ./backend/internal/service/tenantplugin ./backend/internal/service/resolverconfig -count=1`;`openspec validate multi-tenant --strict`;静态扫描确认 `sys_locker`、`sys_menu`、`sys_plugin`、`sys_plugin_release`、`sys_plugin_migration`、`sys_plugin_resource_ref`、`sys_plugin_node_state`、`sys_notify_channel` 的源 SQL、packed SQL 与生成模型不再包含 `tenant_id`/`TenantId`;本次变更文件 `git diff --check` 通过。
- [x] 2026-05-10: FB-20 验证通过:`make init confirm=init rebuild=true`;`make mock confirm=mock`;`cd apps/lina-core && go test ./internal/service/plugin/internal/catalog ./internal/service/plugin/internal/integration ./internal/service/menu -count=1`;`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate && pnpm exec tsc --noEmit --pretty false`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts --project=chromium`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/tenant-lifecycle/TC0178-multi-tenant-enabled.ts e2e/multi-tenant/plugin-governance/TC0209-install-mode-migration.ts --project=chromium`;`openspec validate multi-tenant --strict`;本次变更文件 `git diff --check` 通过。静态扫描确认宿主与插件菜单不再暴露 `tenant` 顶级目录、`/platform/tenant-members`、`/tenant/members`、`/tenant/plugins`,平台管理目录排序位于权限管理与组织管理之间;保留的 `/tenant/*` API 仅作为权限/API 能力存在,不再投影为左侧菜单。
- [x] 2026-05-10: FB-21 验证通过:`make init confirm=init rebuild=true`;`cd apps/lina-core && make dao`;`cd apps/lina-core && make prepare-packed-assets`;`cd apps/lina-core && go test ./internal/service/role ./internal/service/datascope ./internal/service/tenantcap ./internal/service/middleware ./internal/service/i18n ./pkg/pluginservice/bizctx -count=1`;`cd apps/lina-core && go test ./internal/service/plugin ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/datahost -count=1`;`cd apps/lina-plugins/multi-tenant && go test ./... -count=1`;`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate`;`cd hack/tests && pnpm exec tsc --noEmit --pretty false`;`openspec validate multi-tenant --strict`;静态扫描确认无旧平台角色字段残留、无旧三档数据权限校验或说明、无平台/租户前缀权限标识,源 `manifest/sql` 与 `internal/packed/manifest/sql`、源 `manifest/i18n` 与 `internal/packed/manifest/i18n` 目录一致;前端角色页数据权限选项改为消费 `sys_data_scope` 字典,不再维护并行翻译键。
- [x] 2026-05-10: FB-21 `/lina-review` 审查通过:审查发现角色抽屉仍沿用旧数据权限值降级逻辑且前端维护数据权限镜像翻译键,已修复为预加载并消费后端 `sys_data_scope` 字典,组织能力关闭时仅过滤部门数据权限并将既有部门值降级为个人值;复跑上述 FB-21 验证命令均通过。
- [x] 2026-05-10: FB-22 验证通过:`make init confirm=init rebuild=true`;`cd apps/lina-core && make dao`;`cd apps/lina-core && make prepare-packed-assets`;`cd apps/lina-core && go test ./pkg/dialect -run 'TestPluginStateKeepsTechnicalPrimaryKeyAndBusinessUniqueIndex|TestOnConflictTargetsHaveDeclaredIdempotencyBasis|TestSQLColumnIdentifiersAreQuoted' -count=1`;`cd apps/lina-core && go test ./internal/service/plugin/internal/integration ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/wasm -count=1`;`openspec validate multi-tenant --strict`;本次变更文件 `git diff --check` 通过。静态扫描确认 `sys_plugin_state` 不再使用 `(plugin_id, tenant_id, state_key)` 复合主键,而是保留 `id` 自增技术主键并使用 `uk_sys_plugin_state_plugin_tenant_key` 唯一索引表达业务唯一性。尝试运行 `go test ./internal/service/plugin -count=1` 时,既有 `TestValidateStartupConsistencyRejectsPlatformUserMembership` 因当前只初始化宿主 SQL、未安装 multi-tenant 插件表而失败,缺少 `plugin_multi_tenant_user_membership` 表;该失败与本次 `sys_plugin_state` 变更无关。
- [x] 2026-05-10: FB-23 验证通过：`cd hack/tests && pnpm test:validate`；`cd hack/tests && pnpm exec tsc --noEmit --pretty false`；`openspec validate multi-tenant --strict`；`git diff --check -- hack/tests/playwright.config.ts hack/tests/README.md hack/tests/README.zh-CN.md openspec/changes/multi-tenant/tasks.md`。本次为测试治理改进，不涉及业务行为、数据权限、缓存或`i18n`运行时资源变更；验证方式采用 E2E 治理校验、测试 TypeScript 编译、OpenSpec 校验与本次文件格式检查。
- [x] 2026-05-10: FB-24 验证通过:`make init confirm=init rebuild=true`;`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate && pnpm exec tsc --noEmit --pretty false`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/iam/role/TC0223-role-data-scope-dict-select.ts e2e/iam/role/TC0064-role-form-defaults.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-24 touched files>`。确认角色列表 `data_scope=4` 渲染为“本人数据”而非数字 `4`,角色新增抽屉数据权限使用下拉框并默认“本租户数据”;组织模块未启用时按模块解耦规则继续隐藏“本部门数据”。
- [x] 2026-05-10: FB-24 `/lina-review` 审查通过:角色页列表与表单统一消费后端 `sys_data_scope` 字典,不再维护前端镜像枚举翻译;新增/编辑改为 Select 下拉框并保留组织模块未启用时隐藏“本部门数据”的降级规则;本次不修改 Go 接口、数据操作边界或缓存逻辑;已同步源 manifest 与 packed manifest 字典资源、前端 placeholder、OpenSpec 提案/设计/规范和 E2E 覆盖。
- [x] 2026-05-10: FB-25 验证通过:`cd apps/lina-vben && pnpm exec vitest run apps/web-antd/src/router/backend-menu-normalizer.test.ts --dom`;`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-vben/apps/web-antd/src/router/access.ts apps/lina-vben/apps/web-antd/src/router/backend-menu-normalizer.ts apps/lina-vben/apps/web-antd/src/router/backend-menu-normalizer.test.ts openspec/changes/multi-tenant/tasks.md`。真实浏览器登录 admin 后访问 `/platform/tenants` 可打开宿主租户管理页并展示租户列表;本次修复不新增用户可见文案或 i18n 资源,不涉及数据权限或缓存变更。
- [x] 2026-05-10: FB-26/FB-27 验证通过:`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate && pnpm exec tsc --noEmit --pretty false`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-26/FB-27 touched files>`。确认租户管理列表删除重复“更多”按钮,新增/编辑租户弹窗使用正确 `code/name/remark` 表单字段,编辑时正确填充当前记录并不提交 `code`;本次不新增用户可见文案,仅复用已有 `remark` 翻译键,不涉及数据权限或缓存逻辑。
- [x] 2026-05-10: FB-28 验证通过:`cd apps/lina-core && go test ./pkg/dialect -run 'TestSeedDictDataTagStylesAreUniquePerType|TestOnConflictTargetsHaveDeclaredIdempotencyBasis|TestSQLColumnIdentifiersAreQuoted' -count=1`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-core/manifest/sql/014-scheduled-job-management.sql apps/lina-core/internal/packed/manifest/sql/014-scheduled-job-management.sql apps/lina-core/pkg/dialect/dialect_sql_idempotency_test.go openspec/changes/multi-tenant/tasks.md`;`diff -u apps/lina-core/manifest/sql/014-scheduled-job-management.sql apps/lina-core/internal/packed/manifest/sql/014-scheduled-job-management.sql`。确认 `cron_job_log_status` 的 8 个 seed 字典值使用不同 `tag_style`;新增 SQL 审计覆盖 seed 字典数据空颜色与同类型重复颜色两类回归;本次为 seed 数据治理与 SQL 资产审计,不新增用户可见文案,不涉及 i18n 资源、数据权限或运行时缓存逻辑变更。
- [x] 2026-05-10: FB-30 验证通过:`make init confirm=init rebuild=true`;手动执行 `apps/lina-plugins/multi-tenant/manifest/sql/001-multi-tenant-schema.sql` 后 `cd apps/lina-plugins/multi-tenant/backend && gf gen dao`;`cd apps/lina-plugins/multi-tenant && go test ./backend/internal/service/lifecycleguard ./backend/internal/service/tenant ./backend/internal/service/membership ./backend/internal/service/resolver ./backend/internal/service/impersonate ./backend/internal/controller/platform ./backend/api/platform/v1 -count=1`;`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate && pnpm exec tsc --noEmit --pretty false`;`make dev && cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts e2e/multi-tenant/tenant-lifecycle/TC0179-tenant-crud-code-tombstone.ts e2e/multi-tenant/tenant-lifecycle/TC0180-tenant-suspend-resume.ts e2e/multi-tenant/tenant-lifecycle/TC0181-tenant-no-archive-lifecycle.ts e2e/multi-tenant/tenant-lifecycle/TC0182-tenant-delete-lifecycle-guard.ts e2e/multi-tenant/lifecycle-guard/TC0212-hook-timeout-panic-failsafe.ts --project=chromium` 中 5 个生命周期用例通过、平台页面用例因测试定位器严格模式命中两个“暂停”文案失败;修正定位器后复跑 `TC0222-management-workbench-routes.ts` 通过;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-30 touched files>`。确认租户管理页面不再展示套餐列/输入与归档按钮,创建/编辑请求不提交 `plan`,状态文案从“已暂停”改为“暂停”,列表 `createdAt` 由 API 返回并显示;后端租户主体 schema/API/service/DAO 移除套餐字段,状态机移除归档,删除直接经过 LifecycleGuard 后软删,并新增 LifecycleGuard 单测覆盖 suspended tenant 仍阻止插件禁用/卸载、软删后解除阻断;已同步插件 apidoc i18n、前端运行时语言包与嵌入资源。本次不新增缓存逻辑;租户列表/状态/删除仍受平台管理员权限与 LifecycleGuard 边界约束。
- [x] 2026-05-10: FB-30 `/lina-review` 审查通过:未发现阻塞问题。套餐字段已从租户表 schema、生成 DAO/DO/Entity、API DTO、service 输入输出、前端表格/弹窗与测试请求中移除;归档状态与归档入口已从状态机、事件映射、前端操作和 i18n 文案中移除,仅保留 archived 状态被拒绝的回归断言;`createdAt` 从 `created_at` 查询投影到 API DTO 并由列表列展示。i18n 影响已同步前端三语运行时语言包与插件 zh-CN/zh-TW apidoc 资源,成员状态“已暂停”属于成员域未改动;本次不新增 API 路由或扩大数据操作面,平台租户管理仍沿用 `system:tenant:*` 权限边界与 LifecycleGuard 前置校验;本次不新增缓存逻辑。
- [x] 2026-05-10: FB-29 验证通过:`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate && pnpm exec tsc --noEmit --pretty false`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-vben/apps/web-antd/src/views/platform/tenants/tenant-modal.vue hack/tests/pages/MultiTenantPage.ts openspec/changes/multi-tenant/tasks.md`。确认新增/编辑租户弹窗改为参考参数设置弹窗的 Vben 横向表单,字段标题与输入框保持同一行;本次不新增用户可见文案或 i18n key,不涉及后端接口、数据权限或缓存逻辑变更。
- [x] 2026-05-10: FB-32 验证通过:`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate`;`cd hack/tests && pnpm exec tsc --noEmit --pretty false`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/i18n/TC0112-english-layout-regression.ts --project=chromium -g "TC-112b"`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-vben/apps/web-antd/src/views/system/dict/data/data.ts hack/tests/e2e/i18n/TC0112-english-layout-regression.ts openspec/changes/multi-tenant/tasks.md`。确认字典数据列表 `Dictionary Label` 列宽显著大于 `Order` 列;本次不新增用户可见文案,不涉及 i18n 资源、数据权限、接口或缓存逻辑变更。完整运行 `TC0112` 文件时,既有 `TC-112a` 在 `Organization` 菜单等待处超时,与本次字典数据列宽调整无关。
- [x] 2026-05-10: FB-31 验证通过:`cd apps/lina-core && go test ./internal/service/role -run 'TestImpersonationAccessUsesPlatformRoles|TestTenantRoleAccessFiltersRoleMenuByTenant' -count=1`;`cd apps/lina-core && go test ./internal/service/datascope ./internal/service/role ./internal/service/tenantcap ./internal/service/middleware -count=1`;`cd apps/lina-plugins/multi-tenant && go test ./backend/internal/service/impersonate ./backend/internal/service/resolver -count=1`;`cd hack/tests && pnpm test:validate`;`cd hack/tests && pnpm exec tsc --noEmit --pretty false`;`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0186-impersonation-start-exit.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-core/internal/service/datascope/tenant_scope.go apps/lina-core/internal/service/role/role_user_access.go apps/lina-core/internal/service/role/role.go apps/lina-core/internal/service/role/role_tenant_boundary_test.go apps/lina-plugins/multi-tenant/backend/internal/service/resolver/resolver.go apps/lina-plugins/multi-tenant/backend/internal/service/resolver/resolver_test.go hack/tests/support/multi-tenant-scenarios.ts openspec/changes/multi-tenant/tasks.md`。确认代操作/租户覆盖上下文构建权限快照时读取平台角色,租户解析阶段跳过代操作 membership 校验,请求租户仍保持目标租户用于数据过滤;TC0186 已补充代操作 token 访问 `/user/info` 与 `/menus/all` 的菜单/权限非空断言;本次不新增用户可见文案或 i18n 资源,缓存仍复用现有租户分桶权限快照与权限拓扑修订号。
- [x] 2026-05-10: FB-33 验证通过:`cd apps/lina-core && LINA_TEST_PGSQL_LINK='pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable' go test ./internal/service/plugin -run '^TestInstallMultiTenantWithMockDataOnPostgreSQL$' -count=1 -v`;`cd apps/lina-core && LINA_TEST_PGSQL_LINK='pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable' go test ./pkg/dialect -run '^TestSQLiteTranslateDDLProjectSQLAssetsSmoke$|^TestPostgreSQLProjectSQLAssetsSmoke$' -count=1 -v`。确认当前源 SQL 在全新 PostgreSQL schema 上可以通过真实源码插件安装链路加载 `multi-tenant` mock 数据,并写入租户、用户、角色、配额、配置覆盖与 mock 迁移账本;解决用户环境 SQL 执行错误的操作是先执行 `make init confirm=init rebuild=true` 重建到当前源 schema,再重新安装租户管理插件并勾选 mock 数据。本次仅新增后端回归测试和反馈记录,不新增用户可见文案/i18n 资源,不涉及运行时缓存逻辑变更。
- [x] 2026-05-10: FB-34 验证通过:`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate`;`cd hack/tests && pnpm exec tsc --noEmit --pretty false`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/iam/role/TC0064-role-form-defaults.ts e2e/iam/role/TC0223-role-data-scope-dict-select.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-34 touched files>`。确认角色管理列表、新增抽屉和数据权限下拉框仅在 `multi-tenant` 插件启用时展示/选择“本租户数据”;插件未启用时默认值回退为“全部数据”。本次未新增用户可见文案或 i18n key,仅复用 `sys_data_scope` 字典标签;不涉及后端数据操作接口或缓存逻辑变更。
- [x] 2026-05-10: FB-34 `/lina-review` 审查通过:角色管理列表、创建抽屉与编辑抽屉均通过 `plugins/dynamic` 能力状态过滤 `data_scope=2`;前端仍以 `sys_data_scope` 字典为唯一标签来源,未新增运行时 i18n 文案;测试通过 API 路由 mock 覆盖多租户启用/未启用两种状态,不修改真实插件生命周期。未发现数据权限接口、缓存一致性或后端 REST 合规影响。
- [x] 2026-05-10: FB-35 验证通过:`cd apps/lina-core && go test ./internal/service/plugin -run '^TestMultiTenantMockDataContainsExpectedTenantNames$' -count=1`;`openspec validate multi-tenant --strict`;静态扫描确认 `multi-tenant` mock SQL 不再包含旧 `Alpha Retail`/`Beta Manufacturing`/`Gamma Sandbox` 展示名,并包含 17 个指定中文租户名称。新增单测直接读取 mock SQL 资产并断言每个指定名称作为租户值出现一次;真实 PostgreSQL 安装链路回归测试的租户总数期望同步更新为 17。本次仅调整 mock 数据与回归测试,不新增运行时 UI 文案或 i18n key,不涉及数据权限接口或缓存逻辑变更。
- [x] 2026-05-10: FB-35 `/lina-review` 审查通过:`multi-tenant` mock SQL 仍位于插件 mock-data 目录,未写入自增 `id`,使用真实唯一键 `code` 的 `ON CONFLICT DO NOTHING` 保持可重复加载;字段标识符使用 PostgreSQL 双引号。新增 Go 单测自包含、可单独运行,只读取源码 SQL 资产并校验指定租户名称列表;本次不新增 API、数据操作接口、运行时 UI 文案、i18n key 或缓存逻辑。
- [x] 2026-05-10: FB-40 验证通过:`rg -n "plugin_multi_tenant_quota|multi_tenant_quota|tenant_quota|quota_key|quota_value|users\\.max|storage\\.max|multiTenantMockExpectedQuotaCount|dao\\.Quota|do\\.Quota|entity\\.Quota" . -S` 仅剩 OpenSpec 中"不创建该表"的规范描述与 FB-40 任务记录;`rg --files apps/lina-plugins/multi-tenant/backend/internal/dao apps/lina-plugins/multi-tenant/backend/internal/model/do apps/lina-plugins/multi-tenant/backend/internal/model/entity | rg '/quota\\.go$|/quota_'` 无输出;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-40 touched files>`;`cd apps/lina-plugins/multi-tenant && go test ./backend/internal/dao ./backend/internal/dao/internal ./backend/internal/model/do ./backend/internal/model/entity -count=1`;`cd apps/lina-plugins/multi-tenant && go test ./backend/internal/service/resolverconfig ./backend/internal/service/tenant ./backend/internal/service/membership ./backend/internal/service/tenantplugin -count=1`;`cd apps/lina-core && go test ./internal/service/plugin -run '^TestMultiTenantMockDataContainsExpectedTenantNames$' -count=1`;`cd apps/lina-core && go test ./pkg/dialect -run '^TestSQLiteTranslateDDLProjectSQLAssetsSmoke$' -count=1`。已删除 `plugin_multi_tenant_quota` schema、卸载语句、mock 写入、DAO/DO/Entity 生成物、mock 安装计数断言和 OpenSpec 占位表要求;本次不新增用户可见文案或 i18n 资源,不新增数据操作接口,不涉及运行时缓存逻辑。尝试运行 `cd apps/lina-plugins/multi-tenant && go test ./backend/... -count=1` 时,既有 `TestGuardRejectsSuspendedTenantBeforePluginRemoval` 失败于软删除租户仍阻止卸载,该失败与本次 quota 表删除无关。
- [x] 2026-05-10: FB-36 验证通过:`cd apps/lina-core && go test ./internal/service/user ./internal/controller/user -count=1`;`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate && pnpm exec tsc --noEmit --pretty false`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-36 touched files>`。确认用户管理列表在多租户启用时展示“所属租户”列,平台上下文顶部展示租户筛选条件并向 `GET /user` 传递 `tenantId`,后端按 active membership 过滤;租户上下文隐藏租户筛选并继续由当前租户上下文限定。已同步前端三语运行时语言包、宿主 zh-CN/zh-TW apidoc 资源与 OpenSpec 用户管理规范。本次无新增缓存逻辑;用户列表仍以 membership 为租户权威边界并继续叠加角色数据权限。
- [x] 2026-05-10: FB-36 `/lina-review` 审查通过:新增 `GET /user?tenantId=` 仍符合 REST 只读语义;平台筛选在数据库查询阶段通过 active membership 子查询完成,租户上下文忽略跨租户筛选提示并继续由当前 tenant membership 与角色数据权限共同限定,未发现先查全量再内存过滤或数据权限绕过。前端仅在平台上下文显示租户筛选,租户上下文隐藏筛选避免暗示跨租户查询;运行时 i18n 三语与 apidoc zh-CN/zh-TW 已同步,未新增后端缓存或缓存失效路径。
- [x] 2026-05-10: FB-37 验证通过:`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate`;`cd hack/tests && pnpm exec tsc --noEmit --pretty false`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-vben/apps/web-antd/src/api/tenant/index.ts apps/lina-vben/apps/web-antd/src/api/tenant/model.ts apps/lina-vben/apps/web-antd/src/store/tenant.ts apps/lina-vben/apps/web-antd/src/layouts/basic.vue hack/tests/pages/MultiTenantPage.ts openspec/changes/multi-tenant/tasks.md`。确认顶部租户切换器在平台管理员本地租户列表为空时会主动调用 `/platform/tenants` 加载 active 租户选项,下拉框可选择租户并按平台管理员规范进入 impersonation;租户成员上下文保留 `/auth/login-tenants?userId=` 候选回补。本次不新增用户可见文案或 i18n key,不新增后端数据操作接口,前端只消费既有平台租户列表与认证租户候选接口,未新增运行时缓存逻辑。
- [x] 2026-05-10: FB-37 `/lina-review` 审查通过:本次修复仅调整前端租户候选加载与既有 E2E 页面对象断言,未新增后端 API、SQL、缓存或数据写入路径;平台管理员顶部选择租户复用既有 `/platform/tenants/{id}/impersonate` 进入租户视角,不绕过"平台管理员无 membership、租户内操作走 impersonation"规范;租户成员候选回补复用既有 `/auth/login-tenants?userId=` 查询接口。未新增用户可见文案或 i18n key,数据权限与缓存一致性无新增影响。
- [x] 2026-05-10: FB-38/FB-39 验证通过:`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate`;`cd hack/tests && pnpm exec tsc --noEmit --pretty false`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-vben/apps/web-antd/src/layouts/basic.vue apps/lina-vben/apps/web-antd/src/store/tenant.ts apps/lina-vben/apps/web-antd/src/api/platform/tenant/index.ts apps/lina-vben/apps/web-antd/src/api/tenant/model.ts hack/tests/pages/MultiTenantPage.ts hack/tests/e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts openspec/changes/multi-tenant/specs/dashboard-workbench/spec.md openspec/changes/multi-tenant/tasks.md`。确认代操作提示框位于顶部租户切换器左侧;退出代操作先调用 `/platform/tenants/{id}/end-impersonate`,随后恢复平台管理员原 token、清空当前租户上下文并刷新权限/路由;用户管理页重新查询时不再携带目标租户 `X-Tenant-Code`,可见平台全量用户。已同步 dashboard-workbench 规范;本次不新增用户可见文案或 i18n key,不新增后端数据操作接口或缓存逻辑。
- [x] 2026-05-10: FB-38/FB-39 `/lina-review` 审查通过:代操作提示框与租户切换器同处 `header-right-40`,排序在全局搜索左侧且提示框位于切换器左侧;退出代操作复用既有 `POST /platform/tenants/{id}/end-impersonate` 端点,前端在撤销成功后恢复原平台 token 并清理备份 token 与当前租户状态。用户管理列表未新增查询接口或数据操作边界,恢复平台 token 后继续由既有 `GET /user` 平台上下文、membership 查询和角色数据权限控制可见性;未新增用户可见文案、i18n key、后端缓存或缓存失效逻辑。
- [x] 2026-05-10: FB-42 验证通过:`cd apps/lina-core && go test ./internal/service/plugin -run '^TestMultiTenantMockData(ContainsExpectedTenantNames|DocumentsBlocksAndNicknames)$' -count=1`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-core/internal/service/plugin/plugin_multi_tenant_mock_test.go apps/lina-core/manifest/sql/mock-data/015-mock-multi-tenant-platform.sql apps/lina-plugins/multi-tenant/manifest/sql/mock-data/001-multi-tenant-demo-data.sql openspec/changes/multi-tenant/tasks.md`。确认多租户 mock 数据全部收敛到 `multi-tenant` 插件 `manifest/sql/mock-data`,宿主 `manifest/sql/mock-data` 与 packed manifest 均不再包含 `015-mock-multi-tenant-platform.sql`;插件 mock SQL 每个数据块都有中英文用途注释,用户 nickname 明确体现 PLATFORM 或所属租户与账号用途。本次仅调整 mock 数据与静态回归测试,不新增用户可见运行时文案或 i18n key,不新增数据操作接口,不涉及运行时缓存逻辑。
- [x] 2026-05-10: FB-42 `/lina-review` 审查通过:多租户平台侧演示账号、角色与绑定已迁移到 `apps/lina-plugins/multi-tenant/manifest/sql/mock-data/001-multi-tenant-demo-data.sql`,不再回流宿主 mock SQL;写入不包含自增 `id`,继续使用 `ON CONFLICT DO NOTHING` 与真实唯一键保持幂等;新增 Go 静态测试自包含、可单独运行,覆盖宿主 mock 边界、每段中英文注释和用户 nickname 语义。未新增 API、运行时用户可见文案、i18n key、数据操作接口或缓存逻辑。
- [x] 2026-05-10: FB-43 验证通过:`cd apps/lina-core && make prepare-packed-assets`;`cd apps/lina-core && go test ./pkg/dialect -run 'TestSQLCreateTablesHaveBilingualPurposeComments|TestSQLColumnIdentifiersAreQuoted' -count=1 -v`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-core/manifest/sql apps/lina-core/internal/packed/manifest/sql apps/lina-core/pkg/dialect/dialect_sql_idempotency_test.go apps/lina-plugins/content-notice/manifest/sql apps/lina-plugins/monitor-loginlog/manifest/sql apps/lina-plugins/monitor-operlog/manifest/sql apps/lina-plugins/monitor-server/manifest/sql apps/lina-plugins/multi-tenant/manifest/sql apps/lina-plugins/org-center/manifest/sql apps/lina-plugins/plugin-demo-dynamic/manifest/sql apps/lina-plugins/plugin-demo-source/manifest/sql openspec/changes/multi-tenant/tasks.md`。确认所有源码建表 SQL 的每个 `CREATE TABLE` 上方均有英文 `Purpose` 与中文“用途”注释,重点覆盖 `apps/lina-plugins/multi-tenant/manifest/sql/001-multi-tenant-schema.sql`;新增 SQL 静态审计防止后续建表漏注释,并修正列名审计在建表前存在注释块时的解析。本次为 SQL 资产治理,不新增用户可见运行时文案或 i18n key,不新增数据操作接口,不涉及运行时缓存逻辑。
- [x] 2026-05-10: FB-44 验证通过:`cd apps/lina-core && go test ./internal/service/plugin -run '^TestMultiTenantMockDataDocumentsBlocksAndNicknames$' -count=1`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-core/internal/service/plugin/plugin_multi_tenant_mock_test.go apps/lina-plugins/multi-tenant/manifest/sql/mock-data/001-multi-tenant-demo-data.sql openspec/changes/multi-tenant/tasks.md`。确认多租户插件 mock SQL 中 7 个 `sys_user` 演示账号 nickname 均使用中文,并新增静态断言防止平台用户和租户用户 nickname 回退为英文。本次仅调整 mock 数据与静态回归测试,不新增运行时 UI 文案或 i18n key,不新增数据操作接口,不涉及运行时缓存逻辑。
- [x] 2026-05-10: FB-44 `/lina-review` 审查通过:mock 数据写入仍位于多租户插件 `manifest/sql/mock-data`,不包含自增 `id`,继续依赖 `sys_user.username` 唯一键和 `ON CONFLICT ("username") DO NOTHING` 保持幂等;新增 Go 静态测试自包含、可单独运行,校验所有 `tenant_`/`platform_` 演示用户 nickname 均含中文字符。本次不新增 API、运行时用户可见文案、i18n key、数据操作接口或缓存逻辑。
- [x] 2026-05-10: FB-45 验证通过:`cd apps/lina-core && go test ./internal/service/tenantcap ./internal/service/user ./internal/service/role ./internal/service/notify ./internal/service/plugin ./internal/service/auth ./internal/service/middleware -count=1`;`cd apps/lina-core && go test ./pkg/pluginservice/notify -count=1`;`cd apps/lina-plugins/multi-tenant && go test ./backend/internal/service/membership ./backend/internal/service/provider -count=1`;`openspec validate multi-tenant --strict`;`rg -n "plugin_multi_tenant_user_membership|tenantMembershipTable|notifyTenantMembershipTable|roleTenantMembershipTableExists|pluginMembershipTableExists|validatePlatformUserMembershipStartupConsistency" apps/lina-core -g '*.go' -g '!*_test.go'` 无输出;`git diff --check -- <FB-45 touched files>`。确认宿主用户管理、角色授权、通知扇出和插件启动一致性校验均通过 `tenantcap.UserMembershipProvider` 能力接缝访问租户成员关系,宿主生产代码不再写死 multi-tenant 插件 membership 表名、状态值或查询实现;插件侧集中持有 membership 表、join、状态过滤、替换与启动校验逻辑。新增/更新单元测试覆盖用户列表/创建/更新/租户投影、角色租户授权边界、通知租户扇出、启动校验委托、provider 清空平台用户 membership 与单租户替换边界。本次不新增 API 路由、前端 UI 文案、运行时 i18n key 或缓存逻辑;数据权限仍在数据库查询阶段叠加 tenant membership 与角色数据范围过滤。
- [x] 2026-05-10: FB-41 验证通过:`cd apps/lina-core && go test ./internal/service/user ./internal/controller/user ./internal/service/i18n -count=1`;`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate`;`cd hack/tests && pnpm exec tsc --noEmit --pretty false`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts --project=chromium`;`jq empty <FB-41 i18n/apidoc JSON files>`;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-41 touched files>`。确认用户新增/编辑抽屉在多租户启用时展示"所属租户",平台上下文可维护 `tenantIds` 并提交到 `POST /user`/`PUT /user/{id}`,编辑详情通过 `GET /user/{id}` 回显 `tenantIds/tenantNames`;租户上下文锁定当前租户且不能跨租户写入。已同步前端三语运行时语言包、宿主 zh-CN/zh-TW apidoc、三语错误资源及 packed manifest;新增后端单测覆盖平台创建、平台编辑替换 membership、详情回显和错误本地化。
- [x] 2026-05-10: FB-41 `/lina-review` 审查通过:新增 `tenantIds` 字段复用既有 `POST /user`、`PUT /user/{id}` 与 `GET /user/{id}` 资源语义,未新增非 REST 查询或动作接口;平台写入在事务内更新 `sys_user.tenant_id` 与 membership,租户上下文显式拒绝跨租户写入,详情/写操作仍先执行数据权限可见性校验。前端仅在多租户启用时显示所属租户字段,平台可多选,租户上下文只回显并锁定当前租户;无新增运行时缓存或缓存失效路径,权威数据源仍为数据库中的用户租户 membership。
- [x] 2026-05-10: FB-46 验证通过:`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate && pnpm exec tsc --noEmit --pretty false`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/iam/role/TC0223-role-data-scope-dict-select.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-vben/apps/web-antd/src/views/system/role/index.vue apps/lina-vben/apps/web-antd/src/views/system/role/role-drawer.vue hack/tests/e2e/iam/role/TC0223-role-data-scope-dict-select.ts openspec/changes/multi-tenant/tasks.md`。确认角色新增和编辑抽屉在运行时插件状态列表缺失源码 `multi-tenant` 状态、但租户工作台状态已启用时仍展示“本租户数据”;当状态列表明确返回 `multi-tenant` 禁用时继续隐藏该选项。本次不新增用户可见文案或 i18n key,不新增后端数据操作接口,不涉及运行时缓存逻辑。
- [x] 2026-05-10: FB-47/FB-48 验证通过:`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate`;`cd hack/tests && pnpm exec tsc --noEmit --pretty false`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-47/FB-48 touched files>`。确认租户管理列表代操作按钮悬浮展示简短说明,新增用户管理按钮跳转 `/system/user?tenantId=<id>`,用户管理页租户筛选下拉框自动选中当前租户名称并通过 `GET /user?tenantId=<id>` 查询。已同步前端 zh-CN/en-US/zh-TW 运行时语言包与租户管理增量规范;本次不新增后端 API、数据写入接口或运行时缓存逻辑,用户列表数据权限仍复用既有 `tenantId` 查询的数据库级 active membership 过滤。
- [x] 2026-05-10: FB-47/FB-48 `/lina-review` 审查通过:租户管理页新增的行内按钮仅做前端路由跳转,不新增后端 API 或数据操作边界;用户管理页仍复用既有 `GET /user?tenantId=` 只读查询和数据库级 active membership 过滤。代操作提示与用户管理按钮文案均通过前端三语运行时语言包维护,未发现硬编码 UI 文案;E2E 覆盖已在 TC0222 中验证 tooltip、跳转 URL、下拉框回显和查询参数。未新增缓存或缓存失效逻辑。
- [x] 2026-05-10: FB-49 验证通过:`cd apps/lina-core && go test ./internal/service/user -run 'Test(TenantBoundOperator(CreateAllowsOwnedTenantAssignments|CreateRejectsForeignTenantAssignments|CreateRejectsEmptyTenantAssignments|UpdateRejectsForeignTenantAssignments|ListRejectsForeignTenantFilter)|TenantBoundAllScopeOperatorListRejectsForeignTenantFilter|PlatformUser(CreateWritesSelectedTenantMemberships|UpdateReplacesTenantMemberships)|UserListTenantFilterUsesMembershipForPlatformContext)' -count=1`;`cd apps/lina-core && go test ./internal/service/user ./internal/service/tenantcap -count=1`;`cd apps/lina-plugins/multi-tenant && go test ./backend/internal/service/provider ./backend/internal/service/membership -count=1`;`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm exec tsc --noEmit --pretty false`;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-49 touched files>`。确认用户管理页租户筛选与新增/编辑抽屉优先从 `/auth/login-tenants?userId=` 加载当前操作用户 active membership,有 membership 时不再展示全量租户;后端创建、编辑和 `GET /user?tenantId=` 均拒绝超出操作用户 active membership 的租户,即使操作用户同时具备全量数据权限也不能越权筛选外部租户。完整运行 `TC0222-management-workbench-routes.ts` 时已通过 FB-49 相关用户管理断言,但后续租户切换后仍停留 `/platform/tenants` 的既有 FB-51 问题导致完整用例失败,该阻塞由 FB-51 单独跟踪。本次不新增用户可见文案或 i18n key,不新增运行时缓存逻辑;数据权限边界在数据库查询前的租户候选/筛选校验和写入前 membership 校验处收敛。
- [x] 2026-05-10: FB-50 验证通过:`cd apps/lina-core && go test ./internal/service/plugin -run '^TestMultiTenantMockDataDocumentsBlocksAndNicknames$' -count=1`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-core/internal/service/plugin/plugin_multi_tenant_mock_test.go apps/lina-plugins/multi-tenant/manifest/sql/mock-data/001-multi-tenant-demo-data.sql openspec/changes/multi-tenant/tasks.md`。确认 `apps/lina-plugins/multi-tenant/manifest/sql/mock-data/001-multi-tenant-demo-data.sql` 的平台 mock 用户与租户范围 mock 用户注释均写明演示登录密码 `admin123`,且 `tenant_alpha_ops` 同时拥有 `alpha-retail` 与 `beta-manufacturing` 两个租户 membership,用于租户切换、跨租户列表和演示测试。本次仅调整 mock SQL 注释与静态回归测试,不新增运行时 UI 文案或 i18n key,不新增数据操作接口,不涉及数据权限边界或运行时缓存逻辑。
- [x] 2026-05-10: FB-50 `/lina-review` 审查通过:mock 用户密码说明仅作为 SQL 注释补充,未改变哈希、seed 数据结构、数据写入或运行时行为;跨租户演示用户仍为 `tenant_alpha_ops`,通过 `alpha-retail` 与 `beta-manufacturing` 两条 membership 行表达。新增静态测试自包含并可单独运行,不依赖测试顺序;未新增 API、运行时 i18n key、数据权限边界或缓存失效路径。
- [x] 2026-05-10: FB-52 验证通过:`cd apps/lina-core && go test ./internal/service/plugin -run '^TestMultiTenantMockData(ContainsExpectedTenantNames|DocumentsBlocksAndNicknames)$' -count=1`;`cd apps/lina-core && LINA_TEST_PGSQL_LINK='pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable' go test ./internal/service/plugin -run '^TestInstallMultiTenantWithMockDataOnPostgreSQL$' -count=1 -v`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-core/internal/service/plugin/plugin_multi_tenant_mock_test.go apps/lina-plugins/multi-tenant/manifest/sql/mock-data/001-multi-tenant-demo-data.sql openspec/changes/multi-tenant/tasks.md`。确认 16 个 active 演示租户均至少有一个 active mock 用户、active membership、租户内角色与同租户 `sys_user_role` 绑定;`tenant_alpha_ops` 同时属于 `alpha-retail` 与 `beta-manufacturing`,并分别绑定 `tenant-alpha-ops` 与 `tenant-beta-auditor`,避免切换到 Beta 后只有 membership 没有租户角色。已新增静态 SQL 审计和真实 PostgreSQL 插件 mock 安装链路断言。本次仅调整 mock 数据与测试,不新增 API、运行时 UI 文案或 i18n key,不新增运行时缓存逻辑;用户列表数据权限仍由数据库级 active membership 与角色数据范围过滤控制。
- [x] 2026-05-10: FB-52 `/lina-review` 审查通过:mock 数据仍位于 `multi-tenant` 源码插件 `manifest/sql/mock-data`,未写入自增 `id`,继续依赖 `sys_user.username`、`sys_role(tenant_id,key)`、`plugin_multi_tenant_user_membership(user_id,tenant_id)` 等真实唯一约束和 `ON CONFLICT DO NOTHING` 保持幂等。新增用户均为演示租户内用户且昵称为中文,角色为租户内 `data_scope=4` 演示角色,只授予用户列表/成员只读权限;不扩大后端接口或数据权限边界,不新增缓存失效路径。
- [x] 2026-05-10: FB-53 验证通过:`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd hack/tests && pnpm test:validate && pnpm exec tsc --noEmit --pretty false`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0225-tenant-search-inline-layout.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-vben/apps/web-antd/src/views/platform/tenants/index.vue hack/tests/pages/MultiTenantPage.ts hack/tests/e2e/multi-tenant/platform-admin/TC0225-tenant-search-inline-layout.ts openspec/changes/multi-tenant/tasks.md`。确认租户管理搜索区在 1440px 桌面视口下,租户编码、租户名称、状态、重置、搜索保持一行展示;本次仅调整前端表单栅格与测试定位器,不新增用户可见文案或 i18n key,不新增后端接口、数据权限或缓存逻辑。
- [x] 2026-05-10: FB-53 `/lina-review` 审查通过:租户管理页复用现有 Vben Grid 搜索表单,仅将桌面端搜索栅格调整为与角色管理一致的 `xl:grid-cols-4`,并新增 `TC0225` 聚焦验证搜索条件和按钮同排展示;未发现硬编码 UI 文案、缺失 i18n、后端 REST/API、数据权限或缓存一致性影响。
- [x] 2026-05-10: FB-51 验证通过:`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate && pnpm exec tsc --noEmit --pretty false`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0224-tenant-context-default-route-refresh.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-51 touched files>`。确认租户管理行内代操作、顶部租户下拉代操作、普通租户切换和退出代操作均重新拉取 `/user/info` 与 `/menus/all`,并强制进入当前上下文默认页面;`tenantStore` 不再在 store 初始化时依赖 `useRouter()` 注入,由调用组件显式传入 `router`。本次不新增用户可见文案或 i18n key,不新增后端 API、数据操作接口或运行时缓存逻辑;权限刷新仍以新 token 下重新拉取菜单/用户信息为权威。
- [x] 2026-05-10: FB-51 `/lina-review` 审查通过:租户上下文变化统一经 `tenantStore.switchTenant/exitImpersonation` 刷新权限状态并强制进入默认页,租户管理行内代操作与顶部租户切换入口复用同一 store 流程;store 不再直接调用 `useRouter()` 依赖组件注入上下文,避免在 Pinia 初始化路径下出现 router undefined。新增 `TC0224` 聚焦覆盖行内代操作、顶部代操作、普通租户切换与退出代操作的默认页刷新,并断言新 token 下重新拉取用户信息和菜单;未发现 i18n、后端 REST/API、数据权限或缓存一致性影响。
- [x] 2026-05-10: FB-55 验证通过:`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate`;`cd hack/tests && pnpm exec tsc --noEmit --pretty false`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-vben/apps/web-antd/src/views/platform/tenants/index.vue hack/tests/pages/MultiTenantPage.ts openspec/changes/multi-tenant/tasks.md`。确认租户管理页面工具栏新增按钮复用通用 `pages.common.add` 文案,真实按钮文本为“新 增”,不再显示“新增租户”;新增/编辑租户弹窗标题仍使用 `pages.multiTenant.tenant.actions.create` 显示“新增租户”。本次复用既有通用前端运行时语言包 key,不新增或删除 i18n 资源,不新增后端接口、数据权限或缓存逻辑。
- [x] 2026-05-10: FB-55 `/lina-review` 审查通过:租户管理工具栏新增按钮复用 `pages.common.add`,与用户、角色、岗位等业务模块保持一致;新增/编辑弹窗标题保留租户领域文案。E2E 页面对象在 `TC0222` 中断言按钮文本为“新增”(兼容 Ant Design Vue 中文自动间距),可防止回归为“新增租户”。本次仅调整前端模板、E2E 断言和反馈记录,未新增 API、数据操作、SQL、缓存或数据权限边界;未新增或删除 i18n 资源。
- [x] 2026-05-10: FB-57 验证通过:`cd apps/lina-core && go test ./internal/service/auth ./pkg/pluginservice/auth ./pkg/pluginservice/session ./internal/service/middleware ./internal/controller/auth ./internal/service/user -count=1`;`cd apps/lina-plugins/multi-tenant && go test ./backend/internal/controller/auth -count=1`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-core/internal/service/auth/auth.go apps/lina-core/internal/service/auth/auth_contract_test.go apps/lina-core/internal/service/auth/auth_tenant_flow_test.go apps/lina-core/internal/service/middleware/middleware_tenancy_test.go apps/lina-core/pkg/pluginservice/auth/auth.go apps/lina-core/pkg/pluginservice/auth/auth_test.go openspec/changes/multi-tenant/tasks.md`。确认宿主 `auth.Service` 不再暴露 `SelectTenant`、`SwitchTenant` 或 `SwitchTenantToken`,新增 `TenantTokenIssuer` 窄接缝承载 pre-token 换租户 token 与 bearer token 重签,并增加接口契约单测防止租户流程方法回流到核心 auth service;`pkg/pluginservice/auth` 继续向 multi-tenant 插件暴露租户认证流程但只依赖该窄接缝;本次不改变 HTTP API、前端文案、i18n 资源、数据库读写边界或运行时缓存逻辑。
- [x] 2026-05-10: FB-57 `/lina-review` 审查通过:宿主核心 `auth.Service` 仅保留会话、登录、JWT 解析、密码哈希、登出和会话撤销等认证内核能力;租户选择/切换的 token 签发与重签通过 `TenantTokenIssuer` 窄接口提供给 `pkg/pluginservice/auth` adapter,源码插件仍通过插件发布契约编排 `/auth/select-tenant` 与 `/auth/switch-tenant`。新增 `auth_contract_test.go` 与 adapter 单测均自包含,无测试顺序依赖。调用端可见错误继续使用既有 `bizerr.CodeAuth*`;未新增 HTTP API、SQL、前端文案、i18n 资源、数据操作边界或缓存失效路径。
- [x] 2026-05-10: FB-54 验证通过:`make init confirm=init rebuild=true`;`docker exec -i linapro-postgres-readme psql "postgresql://postgres:postgres@host.docker.internal:5432/linapro?sslmode=disable" -v ON_ERROR_STOP=1 < apps/lina-plugins/multi-tenant/manifest/sql/001-multi-tenant-schema.sql`;`cd apps/lina-plugins/multi-tenant/backend && gf gen dao -c hack/config.yaml`;`cd apps/lina-plugins/multi-tenant && go test ./backend/internal/service/membership ./backend/internal/service/provider ./backend/internal/controller/tenant ./backend/internal/controller/platform ./backend/api/tenant/v1 ./backend/api/platform/v1 -count=1`;`cd apps/lina-core && go test ./internal/service/user ./internal/service/notify ./internal/service/plugin -count=1`;`cd apps/lina-core && go test ./pkg/dialect -run 'TestOnConflictTargetsHaveDeclaredIdempotencyBasis|TestSQLColumnIdentifiersAreQuoted' -count=1`;`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate && pnpm exec tsc --noEmit --pretty false`;`openspec validate multi-tenant --strict`;`rg -n "tenantAdmin|isTenantAdmin|lastActiveAt|is_tenant_admin|last_active_at|IsTenantAdmin|LastActiveAt" <相关实现路径>` 仅剩 FB-54 任务描述命中;`git diff --check -- <FB-54 touched files>`。确认 `plugin_multi_tenant_user_membership` 仅保留 `user_id`、`tenant_id`、`status`、`joined_at` 等成员关系字段,租户管理能力统一由租户上下文、角色和 `system:*` 权限表达。
- [x] 2026-05-10: FB-54 `/lina-review` 审查通过:已从多租户插件 schema、mock SQL、DAO/DO/Entity、membership 服务、tenant/platform API DTO、控制器、前端模型、租户成员页面、平台成员页面、运行时 i18n/apidoc 资源、E2E fixtures/page object/用例与宿主相关静态测试中移除 `is_tenant_admin` 和 `last_active_at` 字段。API 表面删除 `isTenantAdmin`/`lastActiveAt` 输入输出字段,未新增 REST 端点或数据操作面;membership 仍以 `user_id`、`tenant_id`、`status` 作为租户可见性权威边界,不扩大数据权限。i18n 影响已处理:删除 apidoc 字段翻译和前端废弃运行时 key,无硬编码文案遗留。缓存影响:未新增缓存、缓存键或失效路径。
- [x] 2026-05-10: FB-56 验证通过:`cd apps/lina-core && go test ./internal/service/plugin -run '^TestMultiTenantManifestTenantManagementButtonsMatchWorkbench$' -count=1`;`cd hack/tests && pnpm test:validate`;`cd hack/tests && pnpm exec tsc --noEmit --pretty false`;`jq empty apps/lina-plugins/multi-tenant/manifest/i18n/zh-CN/menu.json apps/lina-plugins/multi-tenant/manifest/i18n/en-US/menu.json apps/lina-plugins/multi-tenant/manifest/i18n/zh-TW/menu.json`;`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/tenant-lifecycle/TC0178-multi-tenant-enabled.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-plugins/multi-tenant/plugin.yaml apps/lina-plugins/multi-tenant/manifest/i18n/zh-CN/menu.json apps/lina-plugins/multi-tenant/manifest/i18n/en-US/menu.json apps/lina-plugins/multi-tenant/manifest/i18n/zh-TW/menu.json apps/lina-core/internal/service/plugin/plugin_multi_tenant_mock_test.go hack/tests/support/api/job.ts hack/tests/support/multi-tenant-scenarios.ts openspec/changes/multi-tenant/specs/plugin-permission-governance/spec.md openspec/changes/multi-tenant/tasks.md`。确认菜单管理中租户管理按钮权限仅保留租户查询、新增、修改、删除、代操作和跳转用户管理入口,删除租户解析、租户成员、租户插件等非页面按钮权限;`system:tenant:edit` 继续覆盖状态切换,用户管理入口复用既有 `system:user:query`。已同步插件菜单三语 i18n 和权限治理规范;本次不新增后端 API、数据操作接口、数据权限边界或运行时缓存逻辑。
- [x] 2026-05-10: FB-56 `/lina-review` 审查通过:租户管理菜单按钮权限已与真实页面按钮和入口对齐,不再把租户解析、租户成员或租户插件 API 能力投影为该页面按钮;新增用户管理按钮复用既有系统用户查询权限,未新增权限字符串或后端 API。插件菜单 zh-CN/en-US/zh-TW 资源已同步,新增 Go 静态测试与 E2E 断言覆盖菜单清单和运行时菜单树;本次不涉及数据权限扩大、缓存键或缓存失效路径。
- [x] 2026-05-10: FB-58 验证通过:`cd apps/lina-core && go test ./internal/service/tenantcap ./internal/service/middleware ./internal/service/user ./internal/service/role -count=1`;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-58 touched files>`;静态扫描确认 `internal/service/tenantcap/tenantcap_code.go` 已删除,且不存在 `tenantcapsvc.Code*` 调用或 `CodeTenant* = pkgtenantcap.CodeTenant*` 包内别名。本次为 Go 错误码引用治理,不新增 API、SQL、数据操作、前端文案、i18n 资源或运行时缓存逻辑。
- [x] 2026-05-10: FB-59 验证通过:`cd apps/lina-core && go test ./internal/service/tenantcap -count=1`;`cd apps/lina-core && go test ./internal/service/plugin -run 'TestValidateStartupConsistency(RejectsPlatformUserMembership|RejectsEnabledTenantPluginWithoutProvider|AllowsEnabledTenantPluginWithProvider)' -count=1`;`rg -n "tenantcapsvc\\.ProviderPluginID|const ProviderPluginID = \"multi-tenant\"" apps/lina-core apps/lina-plugins`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-core/pkg/tenantcap/tenantcap.go apps/lina-core/internal/service/tenantcap/tenantcap.go apps/lina-core/internal/service/tenantcap/tenantcap_test.go apps/lina-core/internal/service/plugin/plugin_startup_consistency.go apps/lina-core/internal/service/plugin/plugin_startup_consistency_test.go apps/lina-core/internal/service/notify/notify_send_tenant_test.go apps/lina-core/internal/service/user/user_tenant_membership_test.go apps/lina-core/internal/service/role/role_tenant_boundary_test.go openspec/changes/multi-tenant/tasks.md`。确认多租户官方插件 ID 与 `orgcap` 一致收敛到 `pkg/tenantcap.ProviderPluginID`,宿主 service、启动一致性校验与测试调用方不再引用 `internal/service/tenantcap` 的包内常量;本次不修改插件 manifest ID、API、SQL、用户可见文案、i18n 资源、数据权限或缓存逻辑。
- [x] 2026-05-10: FB-59 `/lina-review` 审查通过:多租户官方插件 ID 现在只在 `pkg/tenantcap.ProviderPluginID` 定义,与 `pkg/orgcap.ProviderPluginID` 的稳定能力契约模式一致;`internal/service/tenantcap`、插件启动一致性校验以及相关单元测试均消费共享契约常量,未发现继续从 `internal/service/tenantcap` 暴露或引用插件 ID 的路径。该治理不修改插件 manifest `id`、REST API、SQL、数据权限边界、前端文案、i18n 资源、日志路径或缓存失效逻辑;相关 Go 单测、OpenSpec 校验与静态扫描已覆盖回归风险。
- [x] 2026-05-10: FB-58 `/lina-review` 审查通过:共享错误码只在 `pkg/tenantcap` 定义,宿主 `tenantcap` service、middleware、user、role 相关调用方直接引用 `pkgtenantcap.Code*`,不再通过 `internal/service/tenantcap` 包内变量转发;调用端可见错误仍使用 `bizerr.NewCode`/`bizerr.Is` 与稳定错误码契约。该治理不新增或修改 REST API、SQL、数据权限边界、前端文案、i18n 资源、日志路径或缓存失效路径;静态扫描和相关 Go 包测试已覆盖本次风险。
- [x] 2026-05-10: FB-61 验证通过:`openspec validate multi-tenant --strict`;`git diff --check -- openspec/changes/multi-tenant/tasks.md`;`rg -n "Platform Operations|Platform Tenant Auditor|Alpha Tenant Admin|Alpha Operations|Beta Tenant Admin|Beta Auditor|Gamma Suspended Admin|Demo User|Eye Catching|One Click|Cyber Wellness|Clock In|Crazy Thursday|Deal Hunter|Stay Calm|Anti Involution|Drink Hot Water|Juejuezi|Sudden Fortune|Breakdown Repair|Yep Yep|Read Random Reply" apps/lina-plugins/multi-tenant/manifest/sql/mock-data/001-multi-tenant-demo-data.sql` 无输出;`rg -n "[[:blank:]]$" apps/lina-plugins/multi-tenant/manifest/sql/mock-data/001-multi-tenant-demo-data.sql openspec/changes/multi-tenant/tasks.md` 无输出。确认多租户 mock SQL 中 `sys_role.name` 已改为中文展示名,角色 `key`、权限绑定和用户角色绑定不变;本次为 mock seed 展示数据治理,不新增业务行为、API、数据权限边界、运行时缓存逻辑或 i18n 资源。
- [x] 2026-05-10: FB-61 `/lina-review` 审查通过:本次只修改 `apps/lina-plugins/multi-tenant/manifest/sql/mock-data/001-multi-tenant-demo-data.sql` 中平台角色与租户角色的 `sys_role.name` mock 展示值,保留 `sys_role.key`、`sys_role_menu` 与 `sys_user_role` 绑定不变,不会改变授权语义或租户隔离边界。未发现硬编码 UI 文案、缺失 i18n、REST/API、Go 代码、数据权限或缓存一致性影响;治理验证方式采用 OpenSpec 严格校验、格式/空白扫描和旧英文角色展示名静态扫描。
- [x] 2026-05-10: FB-62 验证通过:`cd apps/lina-core && go test ./pkg/pluginhost -count=1`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-core/pkg/pluginhost/lifecycle_guard.go openspec/changes/multi-tenant/tasks.md`。确认 `pkg/pluginhost/lifecycle_guard.go` 文件顶部已补充中英文维护注释,覆盖生命周期保护契约、Guard 可选接口、i18n reason、并发聚合、超时、panic 恢复与失败关闭策略。本次仅修改源码注释和 OpenSpec 任务记录,不新增运行时行为、REST/API、数据操作、前端文案、i18n 资源或缓存逻辑。
- [x] 2026-05-10: FB-62 `/lina-review` 审查通过:该变更仅增强 `apps/lina-core/pkg/pluginhost/lifecycle_guard.go` 文件顶部维护说明,未修改 Go 类型、函数、接口、常量或调用链。注释遵循 Go 文件顶部注释规范,保留 `package pluginhost` 前的文件职责说明,且不会引入可执行行为变化;无需新增单元测试或 E2E,以目标包测试、OpenSpec 严格校验和 diff 空白检查作为治理验证。未发现 i18n、数据权限、REST/API、日志上下文或分布式缓存一致性影响。
- [x] 2026-05-10: FB-60 验证通过:`cd apps/lina-plugins/multi-tenant && go test ./backend/internal/service/resolverconfig ./backend/internal/service/resolver ./backend/internal/service/membership ./backend/api/platform/v1 ./backend/internal/controller/platform -count=1`;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-60 touched files>`;`jq empty apps/lina-plugins/multi-tenant/manifest/i18n/zh-CN/apidoc/platform/resolver-config.json apps/lina-plugins/multi-tenant/manifest/i18n/zh-TW/apidoc/platform/resolver-config.json apps/lina-plugins/multi-tenant/manifest/i18n/en-US/apidoc/platform/resolver-config.json`;静态扫描确认宿主源配置模板与 packed 配置模板均不含 `tenant:` 段、`tenant.isolation` 或 `tenant.resolution`,后端不再通过 `g.Cfg().Get("tenant.*")` 读取租户配置。确认租户默认策略集中在代码常量中,`rootDomain` 写入/加载/转发均保持空值,当前不支持设置。本次不新增数据操作接口或运行时缓存路径。
- [x] 2026-05-10: FB-63 验证通过:`cd apps/lina-core && go test ./pkg/pluginservice/bizctx -count=1`;`openspec validate multi-tenant --strict`;`git diff --check -- apps/lina-core/pkg/pluginservice/bizctx/bizctx.go apps/lina-core/pkg/pluginservice/bizctx/bizctx_test.go openspec/changes/multi-tenant/specs/plugin-host-service-extension/spec.md openspec/changes/multi-tenant/tasks.md`。确认 `pkg/pluginservice/bizctx.Service` 新增只读 `Current(ctx)` 快照,一次返回插件可见的用户、租户与 impersonation 元数据;新增单元测试覆盖宿主真实上下文与插件本地上下文两种路径。本次不新增 REST API、SQL、前端文案、i18n 资源、数据操作接口或缓存逻辑。
- [x] 2026-05-10: FB-63 `/lina-review` 审查通过:`pkg/pluginservice/bizctx.Service` 现在通过 `Current(ctx)` 暴露只读 `CurrentContext` 快照,未暴露宿主 `internal/model.Context` 可变指针。审查中补充了内部 service 缺失时仍读取插件本地上下文的回归单测,避免测试替身绕过 context fallback。该变更不新增 REST/API、数据库读写、用户可见文案、i18n 资源、数据权限边界或运行时缓存路径。
- [x] 2026-05-10: FB-60 `/lina-review` 审查通过:宿主源配置模板与 packed 配置模板已移除 `tenant.*` 段,多租户默认策略集中到插件共享常量并带注释说明;membership 不再读取宿主 `g.Cfg().Get("tenant.cardinality")`,解析器默认链与保留子域名通过拷贝函数返回,避免调用方修改全局默认值。`rootDomain` 当前保持空值:非空更新被 `bizerr.CodeResolverConfigInvalid` 拒绝,子域名解析在空根域名下返回空。API 英文源标签与 zh-CN/zh-TW apidoc 资源已同步说明 rootDomain 暂不支持设置,en-US apidoc 继续为空占位。本次没有新增 REST 路由、扩大数据操作面或新增运行时缓存路径。
- [x] 2026-05-10: FB-66 验证通过:`cd apps/lina-core && go test ./pkg/pluginservice/bizctx -count=1`;`cd apps/lina-plugins/multi-tenant && go test ./backend/internal/service/membership ./backend/internal/service/resolver ./backend/internal/service/impersonate ./backend/internal/service/resolverconfig ./backend/internal/service/tenantplugin ./backend/internal/service/tenant -count=1`;`cd apps/lina-plugins/content-notice && go test ./backend/internal/service/notice -count=1`;`cd apps/lina-plugins/monitor-operlog && go test ./backend/internal/service/middleware -count=1`;`openspec validate multi-tenant --strict`;`rg -n "Current(UserID|Username|TenantID|ActingUserID|ActingAsTenant|Impersonation)|PlatformBypass\\(" apps/lina-core/pkg/pluginservice/bizctx apps/lina-plugins -g'*.go'` 无输出;`git diff --check -- <FB-66 touched files>`。确认 `pkg/pluginservice/bizctx.Service` 删除单字段 getter,源码插件统一通过 `Current(ctx)` 快照读取用户、租户与 impersonation 元数据。本次不新增 REST API、SQL、前端文案、i18n 资源、数据操作接口或缓存逻辑。
- [x] 2026-05-10: FB-66 `/lina-review` 审查通过:`pkg/pluginservice/bizctx.Service` 公共契约现在只保留 `Current(ctx)` 只读快照,未继续暴露 `CurrentUserID`、`CurrentTenantID`、`CurrentUsername` 等单字段 getter;源码插件调用点已统一先读取 `CurrentContext` 再访问字段。该变更不新增 REST/API、数据库读写、用户可见文案、i18n 资源、数据权限边界或运行时缓存路径;相关插件服务包测试、OpenSpec 严格校验、旧 API 静态扫描和空白检查均已通过。
- [x] 2026-05-10: FB-64 验证通过:`cd apps/lina-core && go test ./pkg/pluginservice/tenantfilter ./pkg/pluginservice/bizctx -count=1`;`cd apps/lina-plugins/content-notice && go test ./backend/internal/service/notice -count=1`;`cd apps/lina-plugins/monitor-loginlog && go test ./backend/internal/service/loginlog -count=1`;`cd apps/lina-plugins/monitor-operlog && go test ./backend/internal/service/operlog -count=1`;`cd apps/lina-plugins/org-center && go test ./backend/internal/service/dept ./backend/internal/service/post ./backend/provider/orgcapadapter ./backend/internal/service/tenantlifecycle -count=1`;`cd apps/lina-plugins/plugin-demo-source && go test ./backend/internal/service/demo -count=1`;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-64 touched files>`。静态扫描确认源码插件中不再存在本地 `backend/internal/service/tenantfilter` 包或直接反射读取 `BizCtx` 的重复实现。新增公共 `pkg/pluginservice/tenantfilter` 复用 `pkg/pluginservice/bizctx.Current`,统一 `tenant_id` 查询注入、join 显式列过滤和审计字段派生;普通租户请求 `OnBehalfOfTenantID=0`,impersonation/代操作请求 `OnBehalfOfTenantID=TenantID`。尝试运行 `cd apps/lina-plugins/monitor-online && go test ./backend/internal/service/monitor -count=1` 时被既有 `apps/lina-core/internal/service/session/session.go:338:9: no new variables on left side of :=` 编译错误阻断,该包无 `tenantfilter` 调用残留。本次不新增 REST API、SQL、前端文案、i18n 资源或运行时缓存逻辑;数据隔离仍在数据库查询阶段注入租户条件。
- [x] 2026-05-10: FB-64 `/lina-review` 审查通过:`pkg/pluginservice/tenantfilter` 位于宿主稳定插件服务契约层,只依赖 `pkg/pluginservice/bizctx` 与 GoFrame `gdb`,未引入宿主 `internal` 依赖给源码插件;各源码插件已删除重复本地 `tenantfilter` 包并统一 import 公共组件。审计字段派生符合多租户规范:普通租户不写 `on_behalf_of_tenant_id`,impersonation/代操作写目标租户;登录日志和操作日志均有单元测试覆盖普通与 impersonation 两种上下文。本次不新增 REST/API、SQL、用户可见文案、i18n 资源、缓存键或缓存失效路径;未发现数据权限绕过,租户隔离仍由查询模型注入 `tenant_id` 条件。
- [x] 2026-05-10: FB-65 验证通过:`make init confirm=init rebuild=true`;`make mock confirm=mock`;`cd apps/lina-core && make dao`;`docker exec -i linapro-postgres-readme psql "postgresql://postgres:postgres@host.docker.internal:5432/linapro?sslmode=disable" -Atc "SELECT conname || ':' || pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid='sys_online_session'::regclass AND contype='p' UNION ALL SELECT indexname || ':' || indexdef FROM pg_indexes WHERE tablename='sys_online_session' ORDER BY 1;"`;`cd apps/lina-core && go test ./internal/service/session ./internal/service/auth ./pkg/pluginservice/session ./internal/service/middleware ./internal/service/cron -count=1`;`cd apps/lina-core && go test ./pkg/pluginservice/session -run TestSessionListPageAndRevokeApplyDataScope -count=3`;`cd apps/lina-core && go test ./internal/service/session -run 'TestTouchOrValidateRejectsTenantMismatch|TestSessionRecordSurvivesStoreRecreation' -count=3`;`cd apps/lina-plugins/monitor-online && go test ./backend/internal/service/monitor -count=1`;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-65 touched files>`。确认 `sys_online_session` 主键为全局唯一 `token_id`,并存在 `(tenant_id,user_id)`、`(tenant_id,login_time)` 等索引;`auth.RevokeSession`、`session.Store.Get/Delete` 与插件发布服务撤销均按 token 定位,请求时 `TouchOrValidate` 仍校验 claim tenant 与 session tenant 一致;在线会话列表在数据库查询阶段先注入 tenantcap 过滤再叠加 datascope 用户范围,撤销前同时校验目标 session 的 `tenant_id` 和 `user_id` 可见性。静态扫描确认旧 `Get(ctx, 0, tokenID)`、`RevokeSession(ctx, 0, tokenID)` 与旧 `(tenant_id, token_id)` 主键无残留。本次不新增 REST API、前端文案、i18n 资源或运行时缓存键;缓存仍按 token 失效角色访问上下文。
- [x] 2026-05-10: FB-65 `/lina-review` 审查通过:在线会话身份模型已与 JWT、撤销列表和角色访问缓存统一为全局唯一 `token_id`;`tenant_id` 保留为归属、列表过滤、数据权限和请求校验维度。审查中补充了列表侧 tenantcap 数据库过滤和撤销侧 `EnsureTenantVisible` 校验,避免同一用户跨租户会话被误列出或误踢。SQL 源与 packed SQL 一致,生成 DAO/DO/Entity 无列结构差异;新增/更新单测自包含且可按 `-run` 重复执行。本次无用户可见文案和 i18n 资源影响,无新增缓存路径或分布式失效逻辑。
- [x] 2026-05-10: FB-67 验证通过:`cd apps/lina-core && go test ./internal/service/bizctx ./internal/service/datascope ./pkg/pluginservice/bizctx ./pkg/pluginservice/tenantfilter -count=1`;`openspec validate multi-tenant --strict`;`rg -n "StrKey\\(\\\"BizCtx\\\"\\)|\\\"BizCtx\\\"|\\\"bizCtx\\\"|contextKey" apps/lina-core/internal/service/bizctx apps/lina-core/pkg/pluginservice/bizctx apps/lina-core/pkg/pluginservice/tenantfilter apps/lina-core/internal/service/datascope -g'*.go'` 仅命中 `ContextKey` 常量定义;`git diff --check -- <FB-67 touched files>`。确认宿主写入和插件/数据权限读取均复用 `internal/service/bizctx.ContextKey`,不再在读取路径维护多个硬编码字符串 key。本次不新增 REST API、SQL、前端文案、i18n 资源、数据操作接口或缓存逻辑。
- [x] 2026-05-10: FB-68 验证通过:`cd apps/lina-core && go test ./internal/service/bizctx ./pkg/pluginservice/bizctx ./pkg/pluginservice/tenantfilter ./internal/service/datascope -count=1`;`openspec validate multi-tenant --strict`;`rg -n "\\\"UserId\\\"|\\\"Username\\\"|\\\"TenantId\\\"|\\\"ActingUserId\\\"|\\\"ActingAsTenant\\\"|\\\"IsImpersonation\\\"|reflect|FieldByName" apps/lina-core/pkg/pluginservice/bizctx apps/lina-core/pkg/pluginservice/tenantfilter` 无输出;`git diff --check -- apps/lina-core/pkg/pluginservice/bizctx/bizctx.go apps/lina-core/pkg/pluginservice/bizctx/bizctx_test.go apps/lina-core/pkg/pluginservice/tenantfilter/tenantfilter_test.go openspec/changes/multi-tenant/tasks.md`。确认 `pkg/pluginservice/bizctx` 删除反射字段名兜底,仅读取宿主通过统一 `ContextKey` 写入的 `*model.Context`,并转换为插件可见的只读 `CurrentContext`;`tenantfilter` 测试同步改为使用宿主统一上下文模型。本次不新增 REST API、SQL、前端文案、i18n 资源、数据操作接口或缓存逻辑。
- [x] 2026-05-10: FB-68 `/lina-review` 审查通过:`pkg/pluginservice/bizctx.Current` 不再维护 `"UserId"`、`"TenantId"` 等反射字段名字符串,读取路径由 `internal/service/bizctx.ContextKey` 和 `internal/model.Context` 的 Go 类型约束保证与写入侧一致;对外仍只返回只读快照,未暴露可变内部上下文指针。新增单测覆盖内部 service 缺失时仍可读取宿主 `*model.Context`,并覆盖类宿主字段的非预期结构体会被拒绝。本次无用户可见文案、i18n、REST/API、SQL、数据权限、日志或分布式缓存影响。
- [x] 2026-05-10: FB-69 验证通过:`make init confirm=init rebuild=true`;`cd apps/lina-core && go test ./internal/service/plugin -run 'TestUpdateTenantProvisioningPolicy|TestTenantProvisioning' -count=1`;`cd apps/lina-core && go test ./internal/service/plugin ./internal/controller/plugin -count=1`;`cd apps/lina-plugins/multi-tenant && go test ./backend/internal/service/tenantplugin -count=1`;`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate`;`cd hack/tests && pnpm exec tsc --noEmit --pretty false`;`jq empty <FB-69 i18n JSON files>`;`openspec validate multi-tenant --strict`;旧字段静态扫描无输出;`git diff --check -- <FB-69 touched files>`。确认插件清单不再声明旧的新租户默认启用字段,平台策略改由 `sys_plugin.auto_enable_for_new_tenants` 和 `PUT /plugins/{id}/tenant-provisioning-policy` 维护;manifest 同步不会覆盖平台策略;新租户创建只为已安装、已启用、`tenant_aware` 且 `tenant_scoped` 并开启平台策略的插件写入租户启用行,不会安装或启用宿主禁用插件。完整宿主插件服务包测试在 `make init` 后需先加载 `apps/lina-plugins/multi-tenant/manifest/sql/001-multi-tenant-schema.sql`,否则既有启动一致性测试缺少插件 membership 表;补齐插件 schema 后全包测试通过。
- [x] 2026-05-10: FB-69 `/lina-review` 审查通过:新 API 使用 `PUT /plugins/{id}/tenant-provisioning-policy` 和 `plugin:edit` 权限,符合更新资源策略语义;调用方可见校验错误通过 `bizerr.CodePluginTenantProvisioningPolicyInvalid` 封装。`auto_enable_for_new_tenants` 为平台拥有字段,不再从 `plugin.yaml` 读取或同步,并在源码插件 manifest、宿主/插件生成模型、E2E 契约和 OpenSpec 规范中清理旧字段。前端仅在 `tenant_aware + tenant_scoped` 插件上显示可操作开关,并同步三语运行时语言包、菜单权限 i18n、zh-CN/zh-TW apidoc 与错误资源;本次新增的数据写入面仅限平台插件策略列,不涉及租户业务数据权限过滤。缓存影响边界为既有插件启动/运行时 registry 快照,策略更新后同步更新启动 registry 投影,新租户自动启用路径以数据库为权威源读取策略;未新增分布式缓存或跨实例失效路径。
- [x] 2026-05-10: FB-70 验证通过:`cd apps/lina-plugins/monitor-operlog && go test ./backend/api/operlog/v1 -run TestOperLogFilterRequestsExposeOnlyOperType -count=1`;`jq empty apps/lina-plugins/monitor-operlog/manifest/i18n/zh-CN/apidoc/plugin-api-main.json apps/lina-plugins/monitor-operlog/manifest/i18n/zh-TW/apidoc/plugin-api-main.json`;`openspec validate multi-tenant --strict`;重复动作别名字段静态扫描无输出;`git diff --check -- <FB-70 touched files>`。确认操作日志列表和导出接口仅保留 `operType` 查询参数,控制器和服务层不再接收重复动作别名字段,apidoc i18n 与活跃 OpenSpec 规范统一到 `operType/oper_type`。尝试运行 `cd apps/lina-plugins/monitor-operlog && go test ./backend/internal/service/operlog ./backend/internal/controller/operlog -count=1` 时被既有 `apps/lina-core/internal/service/plugin/internal/catalog` 中 `AutoEnableForNewTenants` 生成模型字段不一致阻断,该编译错误与本次操作日志接口清理无关。本次不新增前端 UI 文案、SQL、缓存逻辑或新的数据操作接口;操作日志读写仍在数据库查询阶段复用既有租户过滤。
- [x] 2026-05-10: FB-70 `/lina-review` 审查通过:操作日志列表与导出 API 继续使用 `GET` 读语义和现有 `operType` 查询参数,未新增非 REST 接口;新增 DTO 契约测试自包含、可单独运行,防止重复动作别名字段回流。插件 apidoc zh-CN/zh-TW 资源已删除过期字段翻译,en-US 源文档继续直接使用英文 DTO 标签;活跃 OpenSpec 中审计分类统一以 `oper_type='other'` 加摘要、route metadata 或载荷表达。未新增 SQL、前端 UI 文案、数据操作面、权限边界或缓存失效路径;现有列表、详情、删除和导出仍在数据库查询阶段复用租户过滤。
- [x] 2026-05-11: FB-71 验证通过:`docker exec -i linapro-postgres-readme psql -U postgres -d linapro -v ON_ERROR_STOP=1 < apps/lina-plugins/multi-tenant/manifest/sql/001-multi-tenant-schema.sql`;`cd apps/lina-plugins/multi-tenant/backend && gf gen dao -a`;`cd apps/lina-plugins/multi-tenant && go test ./backend/internal/service/tenant -run 'TestCreateRollsBackTenantWhenProvisioningFails|TestDeleteRunsLifecycleGuardBeforeSoftDelete' -count=1`;`cd apps/lina-plugins/multi-tenant && go test ./backend/internal/service/resolverconfig ./backend/internal/service/tenantplugin ./backend/internal/service/tenant ./backend/internal/service/resolver ./backend/api/platform/v1 ./backend/internal/controller/platform -count=1`;`cd apps/lina-core && go test ./pkg/tenantcap -count=1`;`cd apps/lina-plugins/org-center && go test ./backend/internal/service/dept ./backend/internal/service/post ./backend/provider/orgcapadapter -count=1`;`cd hack/tests && pnpm test:validate`;`cd hack/tests && pnpm exec tsc --noEmit --pretty false`;`jq empty <FB-71 changed JSON files>`;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-71 touched files>`;静态扫描确认 `plugin_multi_tenant_resolver_config` 与 `plugin_multi_tenant_event_outbox` 在实现代码、SQL 配置和生成 DAO/DO/Entity 中无正向引用,仅保留 OpenSpec 禁止性描述和 E2E 表不存在断言。
- [x] 2026-05-11: FB-71 `/lina-review` 审查通过:`plugin_multi_tenant_resolver_config` 与 `plugin_multi_tenant_event_outbox` 已从插件 schema、卸载 SQL、DAO 生成配置、生成代码和事件 outbox 服务中移除;解析规则由 `shared`/`resolverconfig` 代码常量固定为 `override,jwt,session,header,subdomain,default`,保留子域为 `www/api/admin/static/docs`,`rootDomain` 固定空值且 ambiguous 固定 `prompt`,API 只保留内置策略查询与精确 no-op 校验。租户创建副作用改为 `tenant.Create` 同步调用 `tenantplugin.ProvisionForTenant`,并已纳入同一事务,避免插件默认开通失败后留下半创建租户。i18n 影响已处理:更新插件说明与 zh-CN/zh-TW resolver-config apidoc,en-US apidoc 继续为空占位;本次删除 resolver 配置缓存、shared revision 和 outbox 事件路径,不新增分布式缓存机制;数据权限影响限于平台租户创建路径的既有平台权限边界,未新增普通租户可访问的数据操作接口。
- [x] 2026-05-11: FB-72 验证通过:`ruby -ryaml -e 'ARGV.each { |f| YAML.load_file(f) }' apps/lina-plugins/*/plugin.yaml`;`ruby -ryaml -rjson -rdigest <semantic-hash-check>` 确认 10 个插件 `plugin.yaml` 解析后的结构哈希与改动前一致;`ruby <comment-scan> apps/lina-plugins/*/plugin.yaml` 确认所有非注释配置行前均存在中英文注释;`git diff --check -- apps/lina-plugins/*/plugin.yaml openspec/changes/multi-tenant/tasks.md`;`openspec validate multi-tenant --strict`。确认本次仅补充插件清单注释,不改变插件 id、版本、类型、多租户作用域、安装模式、菜单、权限、hostServices 或资源边界等运行时语义。
- [x] 2026-05-11: FB-72 `/lina-review` 审查通过:审查范围限定为 10 个源码插件 `plugin.yaml` 与反馈任务记录,未涉及 Go、前端、SQL、API、manifest/i18n 资源或缓存实现。插件清单中已有用户可见值保持不变,新增内容均为维护注释,不需要新增运行时语言包或 apidoc i18n;未新增数据操作接口、权限边界或缓存失效路径。
- [x] 2026-05-11: FB-73 验证通过:`ruby -ryaml -rjson -rdigest <semantic-hash-check>` 确认 10 个插件 `plugin.yaml` 解析后的结构哈希与改动前一致;`ruby <comment-order-scan> apps/lina-plugins/*/plugin.yaml` 确认所有配置行注释均为英文在前、中文在后;`ruby <enum-comment-scan> apps/lina-plugins/*/plugin.yaml` 确认 `type`、`scope_nature`、`default_install_mode`、菜单 `type`、`visible`、`hostServices.service`、`pluginAccessMode` 等枚举语义字段均说明可选值;`ruby -ryaml -e 'ARGV.each { |f| YAML.load_file(f) }' apps/lina-plugins/*/plugin.yaml`;`git diff --check -- apps/lina-plugins/*/plugin.yaml openspec/changes/multi-tenant/tasks.md`;`openspec validate multi-tenant --strict`。确认本次仅调整维护注释顺序和枚举说明,不改变插件运行时清单语义。
- [x] 2026-05-11: FB-73 `/lina-review` 审查通过:审查范围仍限定为 10 个源码插件 `plugin.yaml` 与反馈任务记录。插件清单结构哈希不变,说明 id、版本、类型、多租户作用域、安装模式、菜单、权限、hostServices、资源边界和查询参数均未改变;本次不涉及 Go、前端、SQL、API、manifest/i18n 资源、数据权限或缓存实现。
- [x] 2026-05-11: FB-74 验证通过:`cd apps/lina-vben && pnpm -F @lina/web-antd typecheck`;`cd apps/lina-vben && pnpm -F @lina/web-antd i18n:check`;`cd hack/tests && pnpm test:validate`;`cd hack/tests && pnpm exec tsc --noEmit --pretty false`;`cd apps/lina-core && go test ./internal/service/plugin/internal/integration -run TestSyncMultiTenantPluginMenusResolveAllowedHostParents -count=1`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0222-management-workbench-routes.ts --project=chromium`;`cd hack/tests && E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ pnpm exec playwright test e2e/multi-tenant/platform-admin/TC0225-tenant-search-inline-layout.ts --project=chromium`;`openspec validate multi-tenant --strict`;`git diff --check -- <FB-74 touched files>`。确认租户管理页面迁入 `apps/lina-plugins/multi-tenant/frontend/pages/tenant-management.vue`,弹窗迁入 `frontend/pages/components/tenant-modal.vue`,页面本地 CRUD API 迁入 `frontend/pages/tenant-client.ts`;`plugin.yaml` 改为 `system/plugin/dynamic-page`,前端移除 `/platform/tenants` 宿主静态页映射与旧 normalizer 特例。路径 `/platform/tenants`、菜单 key、权限点、运行时文案和用户可见交互保持不变;宿主侧仅保留顶部租户切换、用户管理租户选项和 impersonation 仍需的运行时集成 API。本次不新增后端数据操作接口、SQL 或缓存路径;无新增 i18n key,只更新插件前端 README 双语说明。
- [x] 2026-05-11: FB-74 `/lina-review` 审查通过:多租户租户管理页面已按源码插件页面惯例通过 `pluginPageMeta.routePath='platform/tenants'` 注册,后端菜单继续使用同一路径、同一菜单 key 和 `system:tenant:*` 权限点,组件指向 `system/plugin/dynamic-page`;删除宿主 `backend-menu-normalizer` 特例和宿主静态租户管理页面后,路由仍由后端插件菜单和前端插件页注册共同驱动。宿主 `#/api/platform/tenant` 已收缩为顶部租户切换、用户管理租户选项、平台 impersonation 等运行时集成所需接口,页面 CRUD 收敛到插件本地 `tenant-client.ts`。本次无 Go 业务逻辑、REST DTO、SQL、数据权限规则、运行时缓存或分布式失效变更;前端未新增用户可见文案和 i18n key,仅补充双语 README 说明页面边界。

## Feedback

- [x] **FB-1**: 固化多租户需求澄清:租户 code ASCII 约束、TenantId 优先级、平台 bypass/impersonation、暂停租户访问边界、tenant_scoped → global 强制启用、用户列表 membership 权威边界
- [x] **FB-2**: 收尾缓存一致性:角色访问缓存 key 统一 tenantcap 分桶,解析器配置缓存使用可验证跨实例修订号失效
- [x] **FB-3**: 修正宿主 auth 路由注册依赖具体控制器类型断言,改为通过生成接口注册租户选择与切换接口
- [x] **FB-4**: 保留启动期插件/租户治理一致性校验,并在插件安装启用、角色变更、租户 membership 写入链路提前拒绝会制造不一致状态的请求
- [x] **FB-5**: 拆分宿主 API DTO 文件,确保每个带路由 path 的 API 接口定义独立源码文件
- [x] **FB-6**: 拆分多租户源码插件 API DTO 文件,确保插件每个带路由 path 的 API 接口定义独立源码文件
- [x] **FB-7**: 平台管理与租户管理工作台菜单误指向插件动态页壳,导致页面打开显示插件未找到
- [x] **FB-8**: 拆分宿主与多租户插件聚合控制器源码文件,确保每个接口独立维护
- [x] **FB-9**: 多租户选择与切换 API 不应由宿主 auth API 模块声明,应收敛到 multi-tenant 源码插件并仅复用宿主认证能力接缝
- [x] **FB-10**: 多租户启用时用户管理列表应展示用户所属租户,平台用户管理功能合并到用户管理页面
- [x] **FB-11**: 租户管理页面交互体验参考 RuoYi Plus Vben5,补齐更适合平台运营的表格操作与状态展示
- [x] **FB-12**: 工作台右上角应在搜索条左侧提供租户切换下拉框,参考 RuoYi Plus Vben5 的租户切换入口
- [x] **FB-13**: 移除租户解析设置页面与平台用户管理页面,同步清理菜单、路由、API 客户端与 i18n 资源
- [x] **FB-14**: 平台管理下平台用户与系统用户管理重复,统一以系统用户管理承载跨租户用户可见性与租户归属展示
- [x] **FB-15**: 增加多租户 mock 数据,覆盖多个租户、成员关系、租户管理员和跨租户展示场景
- [x] **FB-16**: 多租户宿主表结构改动应合并回对应建表 SQL,避免在 `016` 中保留迁移式 `ALTER TABLE` 修补
- [x] **FB-17**: 租户切换 E2E 页面对象应匹配当前真实下拉框 DOM,覆盖真实切换交互
- [x] **FB-18**: 登录链路更新最后登录时间失败应使用 `bizerr` 封装调用端可见错误
- [x] **FB-19**: 严格收敛租户字段边界,移除平台控制面表中无业务归属的 `tenant_id` 并补齐字段必要性规范
- [x] **FB-20**: 删除租户工作台目录、平台租户成员菜单,并将平台管理目录移动到权限管理与组织管理之间
- [x] **FB-21**: 去除角色平台布尔标识,统一使用 `system:*` 权限标识,以租户上下文、数据权限和功能权限组合表达平台/租户管理员能力
- [x] **FB-22**: `sys_plugin_state` 保留 `id` 自增技术主键,同时使用插件、租户、状态键唯一索引表达业务唯一性
- [x] **FB-23**: E2E 测试应保留失败诊断`trace`、`video`、`screenshot`，并明确`CRUD`用例按真实页面元素与用户流程编写
- [x] **FB-24**: 角色管理列表数据权限不得显示原始数字,新增/编辑角色数据权限改为下拉框并包含全部数据、本租户数据、本部门数据、本人数据四类选项
- [x] **FB-25**: 租户管理菜单打开后应始终进入宿主租户管理页面,避免旧插件动态页组件残留导致找不到页面
- [x] **FB-26**: 租户管理列表行内操作已覆盖状态切换与进入租户,应删除重复的更多按钮
- [x] **FB-27**: 新增租户和编辑租户弹窗应展示正确租户表单字段并在打开时填充当前记录
- [x] **FB-28**: 字典 seed 数据同一字典类型下不得复用相同 tag 颜色,并增加 SQL 资产审计防止回归
- [x] **FB-29**: 新增租户和编辑租户弹窗表单应参考参数设置弹窗,将字段标题与输入框保持同一行
- [x] **FB-30**: 租户管理页面移除套餐列/字段和归档功能,暂停状态文案改为“暂停”,并修复创建时间列展示
- [x] **FB-32**: 字典管理页面中字典数据列表的字典标签列应增加宽度,排序列应减少宽度
- [x] **FB-31**: 平台管理员点击租户管理代操作后访问业务页面不应因权限快照按目标租户读取而全部提示无权限
- [x] **FB-33**: 租户管理插件安装时勾选同时安装 mock 数据会触发 SQL 执行错误
- [x] **FB-34**: 角色管理页面仅在多租户插件启用时展示和选择"本租户数据"
- [x] **FB-35**: 租户管理 mock 数据应使用指定中文公司名称列表
- [x] **FB-36**: 用户管理页面列表缺少所属租户列和租户筛选条件,并需审查多租户相关能力缺口
- [x] **FB-37**: 管理页面顶部“切换租户”下拉框应主动加载可切换租户数据
- [x] **FB-38**: 租户代操作提示框应显示在顶部“切换租户”框左侧
- [x] **FB-39**: 退出租户代操作后用户管理列表仍按先前租户过滤,应恢复平台上下文并查询正确用户数据
- [x] **FB-40**: 去掉无业务调用方的 `plugin_multi_tenant_quota` 占位表及相关 SQL、生成代码、mock 数据、测试断言和 OpenSpec 描述
- [x] **FB-41**: 用户管理新增和编辑抽屉缺少所属租户字段,平台上下文无法维护用户租户归属
- [x] **FB-42**: 租户相关 mock SQL 需补齐每段中英文用途注释,用户 nickname 需体现所属租户与用途
- [x] **FB-43**: 所有建表 SQL 在 `CREATE TABLE` 前补齐中英文表用途注释,重点覆盖多租户插件 schema
- [x] **FB-44**: 多租户插件 mock 数据中的用户昵称应使用中文
- [x] **FB-45**: 宿主用户管理、角色、通知和启动校验不应直接耦合 multi-tenant 插件 membership 表名与查询实现
- [x] **FB-46**: 角色新增和编辑抽屉在多租户插件已启用时仍可能缺少“本租户数据”选项
- [x] **FB-47**: 租户管理列表代操作按钮缺少简短说明,用户无法快速理解代操作含义
- [x] **FB-48**: 租户管理列表缺少跳转用户管理并按当前租户自动筛选的入口
- [x] **FB-49**: 用户管理页租户筛选与新增/编辑租户选择不能越过当前操作用户的所属租户范围
- [x] **FB-50**: 多租户 mock 用户注释应写明演示密码,并确认至少存在一个用户同时属于多个租户用于演示和测试
- [x] **FB-51**: 代操作租户、租户切换与退出代操作后应强制刷新权限并进入当前上下文默认页面
- [x] **FB-52**: 用户列表页多租户 mock 数据缺少部分租户用户,且跨租户 mock 用户存在 membership 与租户角色绑定不一致
- [x] **FB-53**: 租户管理页面顶部搜索框过长,应参考角色管理页面让搜索条件与按钮在桌面端一行展示
- [x] **FB-54**: 删除 membership 表中 `is_tenant_admin` 和 `last_active_at` 字段,统一以角色权限表达租户管理能力并清理前后端字段依赖
- [x] **FB-55**: 租户管理页面新增按钮名称应由“新增租户”改为“新增”
- [x] **FB-56**: 菜单管理页面中租户管理的按钮权限应按租户管理页面实际按钮更新
- [x] **FB-57**: 宿主 `auth.Service` 不应直接暴露租户选择与切换流程方法,租户登录编排应通过窄接缝供 multi-tenant 插件复用
- [x] **FB-58**: 删除 `internal/service/tenantcap` 中对 `pkg/tenantcap` 错误码的包内别名变量,调用方直接引用共享契约错误码
- [x] **FB-59**: 多租户官方插件 ID 应参考 `orgcap` 收敛到 `pkg/tenantcap` 稳定契约包维护
- [x] **FB-60**: 从宿主配置模板移除租户配置项,将租户默认策略硬编码到代码并注明含义,且 `rootDomain` 当前不支持设置
- [x] **FB-61**: 多租户插件 mock 数据中的角色名称应使用中文名称
- [x] **FB-62**: `pkg/pluginhost/lifecycle_guard.go` 文件顶部需补充中英文维护注释,说明生命周期保护契约、失败关闭策略和长期维护边界
- [x] **FB-63**: `pkg/pluginservice/bizctx.Service` 应提供插件可见的只读上下文快照,避免插件为一次读取多个上下文字段反复调用单字段 getter
- [x] **FB-64**: 各源码插件重复维护 `tenantfilter` 且审计代操作租户字段语义不一致,应抽象为公共插件租户过滤组件
- [x] **FB-65**: 在线会话撤销不应依赖 `tenant_id=0` 魔法值,应以全局唯一 `token_id` 定位会话并将 `tenant_id` 仅作为归属与校验维度
- [x] **FB-66**: 删除 `pkg/pluginservice/bizctx.Service` 上的单字段 getter,源码插件统一通过 `Current(ctx)` 只读快照读取上下文字段
- [x] **FB-67**: `bizctx` 上下文读写应复用同一个 `ContextKey` 常量,禁止读取路径硬编码多个字符串 key
- [x] **FB-68**: `pkg/pluginservice/bizctx` 不应通过反射字段名字符串读取上下文,应只读取宿主写入的统一 `model.Context` 类型
- [x] **FB-69**: 新租户自动启用插件策略不应由 `plugin.yaml` 声明,应由平台插件系统维护并避免插件同步覆盖后台策略
- [x] **FB-70**: 操作日志导出与列表接口不应同时暴露重复动作别名字段和 `operType`,应统一使用现有 `operType`
- [x] **FB-71**: 删除多租户插件 `resolver_config` 与 `event_outbox` 占位表,解析规则固定在代码中,租户创建副作用改为显式领域编排
- [x] **FB-72**: 所有源码插件 `plugin.yaml` 中的配置项应补齐中英文注释说明
- [x] **FB-73**: `plugin.yaml` 注释应统一英文在前中文在后,并为枚举配置项说明可选值
- [x] **FB-74**: 多租户租户管理前端页面应收敛到 `multi-tenant` 插件 `frontend/pages` 目录维护
