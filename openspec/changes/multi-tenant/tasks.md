## 1. Schema 与初始化

- [ ] 1.1 新增 `apps/lina-core/manifest/sql/016-multi-tenant-and-plugin-governance.sql`,在所有 `sys_*` 表上 `ALTER TABLE ADD COLUMN IF NOT EXISTS tenant_id INT NOT NULL DEFAULT 0`,索引升级为 `(tenant_id, ...)` 联合索引(幂等)
- [ ] 1.2 在 `sys_role` 表加 `is_platform_role BOOL NOT NULL DEFAULT FALSE`,加约束 `CHECK (is_platform_role = false OR tenant_id = 0)`
- [ ] 1.3 在 `sys_user_role` 表加 `tenant_id INT NOT NULL DEFAULT 0`,UNIQUE 约束改为 `(user_id, role_id, tenant_id)`
- [ ] 1.4 在 `sys_dict_type` 表加 `allow_tenant_override BOOL NOT NULL DEFAULT FALSE`
- [ ] 1.5 在 `sys_plugin` 表加 `scope_nature VARCHAR(32) NOT NULL DEFAULT 'tenant_aware'` 与 `install_mode VARCHAR(32) NOT NULL DEFAULT 'global'`
- [ ] 1.6 在 `sys_plugin_state` 表加 `tenant_id INT NOT NULL DEFAULT 0`,主键改为 `(plugin_id, tenant_id)`
- [ ] 1.7 在 `monitor-loginlog`、`monitor-operlog` 插件 schema 增加 `acting_user_id`、`on_behalf_of_tenant_id`、`is_impersonation` 字段(改其 manifest/sql/001-*)
- [ ] 1.8 在所有插件业务表 (`plugin_*`) 中加 `tenant_id INT NOT NULL DEFAULT 0`,索引升级
- [ ] 1.9 seed 数据:平台超级管理员角色(`is_platform_role=true, tenant_id=0`)、admin 用户绑定平台超管角色;现有 seed 字典/配置统一标记 `tenant_id=0`
- [ ] 1.10 修改现有 mock-data SQL,默认写入 `tenant_id=0`,确保 `make mock` 单租户开箱体验不变
- [ ] 1.11 执行 `make init` 重建数据库,验证所有表结构与索引正确
- [ ] 1.12 执行 `make mock` 验证 mock 数据加载成功且单租户行为不破坏

## 2. 宿主稳定接缝(pkg/tenantcap)

- [ ] 2.1 创建 `apps/lina-core/pkg/tenantcap/` 包,定义 `Provider` 接口、`TenantID` 类型、`PLATFORM` 常量、`RegisterProvider`/`CurrentProvider`/`HasProvider`(参考 pkg/orgcap 形态)
- [ ] 2.2 在 `pkg/tenantcap/` 内定义 `LifecycleEvent`、`LifecycleSubscriber` 接口与事件类型枚举
- [ ] 2.3 在 `pkg/tenantcap/` 内定义 `TenantInfo` 结构(id, code, name, status)与 `Resolver` 子接口

## 3. 宿主侧 tenantcap 服务

- [ ] 3.1 创建 `apps/lina-core/internal/service/tenantcap/tenantcap.go` 主文件:`Service` 接口、`serviceImpl`、`New()`,实现 `Enabled`、`Apply`、`Current`、`PlatformBypass`、`EnsureTenantVisible`
- [ ] 3.2 在 `tenantcap_code.go` 中集中定义 `bizerr.Code`(`CodeTenantRequired`、`CodeTenantForbidden`、`CodeCrossTenantNotAllowed`、`CodePlatformPermissionRequired`、`CodeTenantSuspended` 等)
- [ ] 3.3 在 `tenantcap_resolver_chain.go` 中实现责任链调度器,支持注册多个 `Resolver` 与按配置顺序遍历
- [ ] 3.4 在 `tenantcap_fallback.go` 中实现 `ReadWithPlatformFallback` helper(应用层两次查询合并),供字典/配置等"租户覆盖"资源使用
- [ ] 3.5 增补单元测试(自包含、顺序无关):覆盖 `Apply` 注入、`PlatformBypass`、责任链顺序、fallback 语义

## 4. bizctx 与中间件

- [ ] 4.1 在 `internal/service/bizctx/bizctx.go` 增加 `TenantId int`、`ActingAsTenant bool`、`ActingUserId int`、`IsImpersonation bool` 字段;增加 `SetTenant`、`SetImpersonation` 方法
- [ ] 4.2 创建 `internal/service/middleware/middleware_tenancy.go`,实现 tenancy 中间件:在 auth 之后、权限校验之前调用 `tenantcap.Provider.ResolveTenant` 并写入 bizctx
- [ ] 4.3 在 `cmd` 启动期注册 tenancy 中间件;multi-tenant 未启用时短路注入 `TenantId=0`
- [ ] 4.4 单元测试:覆盖中间件注入路径、短路路径、resolver 链选用

## 5. 启动期一致性校验

- [ ] 5.1 在 `framework-bootstrap-installer` 启动流程中增加一致性校验:`scope_nature` ↔ `install_mode`、`is_platform_role` ↔ `tenant_id=0`、平台用户无 membership、multi-tenant enabled ↔ Provider 注册
- [ ] 5.2 检查失败时打印明确日志并 panic 阻止启动
- [ ] 5.3 集成测试:故意构造非法状态验证启动失败

## 6. 多租户插件骨架(lina-plugin-multi-tenant)

- [ ] 6.1 创建 `apps/lina-plugins/multi-tenant/` 目录,包含 `plugin.yaml`(scope_nature=platform_only、install_mode=global)、`plugin_embed.go`、`backend/`、`frontend/`、`manifest/`
- [ ] 6.2 在 `apps/lina-plugins/lina-plugins.go` 中注册 `_ "lina-plugin-multi-tenant/backend"`
- [ ] 6.3 在 `backend/` 下建立 `api/`、`internal/{controller, service, dao, model/{do, entity}}`、`hack/config.yaml`、`plugin.go`
- [ ] 6.4 创建 `manifest/sql/001-multi-tenant-schema.sql`(所有插件持有的表)与 `manifest/sql/uninstall/`
- [ ] 6.5 创建 `manifest/i18n/zh-CN/plugin.json` 与 `manifest/i18n/en-US/plugin.json` 的占位骨架
- [ ] 6.6 plugin.yaml 列出菜单(平台管理员侧:tenants/members/resolver-config;租户管理员侧:members/plugin-management)、permission 点(`platform:tenant:*`、`tenant:member:*` 等)

## 7. 多租户插件 schema

- [ ] 7.1 `plugin_multi_tenant_tenant`:id、code(unique)、name、status、plan、deleted_at(软删)
- [ ] 7.2 `plugin_multi_tenant_user_membership`:id、user_id、tenant_id、is_tenant_admin、status、joined_at、last_active_at,UNIQUE (user_id, tenant_id)
- [ ] 7.3 `plugin_multi_tenant_config_override`:占位扩展,首版可不创建或仅创建表骨架
- [ ] 7.4 `plugin_multi_tenant_quota`:占位扩展(tenant_id, quota_key, quota_value, unit)
- [ ] 7.5 `plugin_multi_tenant_resolver_config`:存放后台改动的解析链/保留子域/ambiguous 行为
- [ ] 7.6 `plugin_multi_tenant_event_outbox`:事件 outbox 表占位(首版同步分发)
- [ ] 7.7 `manifest/sql/uninstall/001-cleanup.sql`:卸载时清理插件表(前置条件由 LifecycleGuard 保证)
- [ ] 7.8 `dao/`、`model/{do,entity}/` 通过 `cd apps/lina-plugins/multi-tenant/backend && make dao` 生成

## 8. 多租户插件 service 层

- [ ] 8.1 `service/tenant/`:`Service` 接口、`serviceImpl`,实现租户 CRUD、状态机迁移、code 唯一性校验、tombstone 30 天保留
- [ ] 8.2 `service/membership/`:1:N 成员关系 CRUD、`tenant.cardinality=single` 校验、平台管理员不允许加 membership
- [ ] 8.3 `service/resolver/`:实现 6 个 Resolver(override/jwt/session/header/subdomain/default),支持配置驱动的链;正式 JWT 优先于 header/subdomain hint
- [ ] 8.4 `service/lifecycle/`:发布 `tenant.created/suspended/resumed/archived/deleted` 事件,outbox 表写入与同步分发
- [ ] 8.5 `service/provider/`:`tenantcap.Provider` 实现,聚合 tenant + membership service,在插件 enable 时调 `tenantcap.RegisterProvider`
- [ ] 8.6 `service/lifecycleguard/`:实现 `CanUninstall`(active 租户存在则 false)、`CanDisable`、`CanTenantDelete`(订阅其他插件钩子聚合)
- [ ] 8.7 `service/impersonate/`:平台管理员 impersonation token 签发、结束 impersonation、双轨日志写入
- [ ] 8.8 `service/resolverconfig/`:解析器配置后台读写、配置缓存与失效广播
- [ ] 8.9 service 层单元测试自包含覆盖

## 9. 多租户插件 controller 层

- [ ] 9.1 `controller/platform/`:`/platform/tenants/*`(CRUD)、`/platform/tenants/{id}/impersonate`、`/platform/tenant/resolver-config/*`、`/platform/users`
- [ ] 9.2 `controller/tenant/`:`/tenant/members/*`(列表、邀请、移除、调整角色)、`/tenant/members/me`
- [ ] 9.3 `controller/auth/`:`/auth/login-tenants`、`/auth/select-tenant`、`/auth/switch-tenant`(覆盖宿主原 login)
- [ ] 9.4 controller 通过 `cd apps/lina-plugins/multi-tenant/backend && make ctrl` 生成骨架,业务逻辑写在生成文件内
- [ ] 9.5 所有 controller 字段持有依赖 service,在 NewV1 中初始化(禁止方法内 service.New())

## 10. JWT / 会话 / 鉴权改造

- [ ] 10.1 在 `auth.Claims` 增加 `TenantId int`、`IsImpersonation bool`、`ActingUserId int` 字段;调整签发与解析逻辑
- [ ] 10.2 修改 `auth.Service.Login` 为两阶段:返回 pre_token + 租户列表(单租户用户走兼容直接签发)
- [ ] 10.3 实现 pre_token 短期单次使用机制(60s TTL,放在 redis/session store)
- [ ] 10.4 实现 `/auth/select-tenant`:校验 pre_token + membership → 签发正式 JWT
- [ ] 10.5 实现 `/auth/switch-tenant`:校验 membership → 旧 token 加入 revoke 列表 → 重签
- [ ] 10.6 实现 token revoke 列表(本地内存 + cluster 广播)
- [ ] 10.7 修改 `session.Store`:主键 `(tenant_id, token_id)`,index `(user_id, tenant_id)`,所有方法签名增加 tenant_id
- [ ] 10.8 修改 `auth.Logout`:仅撤销当前 (tenant, token) 行
- [ ] 10.9 单元测试覆盖 Login → SelectTenant → SwitchTenant → Logout 全链路

## 11. DAO 注入纪律落地

- [ ] 11.1 在 `internal/service/user/` 改造用户 service:多租户启用时列表/详情以 membership join 作为可见性权威边界,`sys_user.tenant_id` 仅作为主租户/默认登录租户;写时填主租户并创建 membership
- [ ] 11.2 在 `internal/service/role/` 改造所有读 `sys_role` 与 `sys_user_role`;角色查询联合考虑 `is_platform_role`
- [ ] 11.3 在 `internal/service/menu/` 改造,菜单仍为平台全局,但解析时按 (tenant, plugin) 启用状态过滤插件菜单
- [ ] 11.4 在 `internal/service/dict/` 改造为 `ReadWithPlatformFallback` 模式;写入按当前租户;`allow_tenant_override` 校验
- [ ] 11.5 在 `internal/service/sysconfig/` 同上
- [ ] 11.6 在 `internal/service/file/` 改造存储路径前缀 `/storage/t/{tid}/...`;读取校验
- [ ] 11.7 在 `internal/service/notify/` 改造通知发送/查询,跨租户广播仅平台
- [ ] 11.8 在 `internal/service/usermsg/` 改造消息 inbox 按租户过滤
- [ ] 11.9 在 `internal/service/jobmgmt/` `jobmeta/` `jobhandler/` 改造,任务执行前构造租户 bizctx
- [ ] 11.10 在 `internal/service/session/` 与 `internal/service/cron/` 改造,session 清理与定时任务上下文绑定 tenant
- [ ] 11.11 在 `internal/service/datascope/` 中调整 Apply 顺序:先 tenantcap 后 datascope
- [ ] 11.12 全宿主代码 grep `dao.Sys*\.Ctx(ctx)` 确保无散点遗漏

## 12. 缓存与一致性

- [ ] 12.1 在 `internal/service/kvcache/`、`internal/service/pluginruntimecache/`、`internal/service/cachecoord/` 改造 cache key,引入 `(tenant_id, scope, key)` 形态
- [ ] 12.2 创建 `tenantcap.CacheKey(tenant, scope, key)` helper,所有缓存使用方统一调用
- [ ] 12.3 失效消息 schema 增加 `tenant_id` 字段;支持 `cascade_to_tenants` 标志
- [ ] 12.4 `cluster.Service` 集群广播链路透传 tenant 维度
- [ ] 12.5 翻译缓存(framework-i18n-runtime)key 与失效作用域改造
- [ ] 12.6 字典缓存、配置缓存、角色缓存、权限缓存、菜单缓存全部按租户分桶
- [ ] 12.7 单元测试:跨租户缓存隔离反例、平台默认级联失效

## 13. 插件治理改造

- [ ] 13.1 在 `internal/service/plugin/` 中增加 `scope_nature` / `install_mode` 解析与持久化,plugin.yaml 解析期校验
- [ ] 13.2 改造 `IsEnabled(ctx, pluginID)` 为租户感知(读 `(plugin_id, tenant_id)`)
- [ ] 13.3 实现 install_mode 切换流程(global ↔ tenant_scoped)与 force 通道
- [ ] 13.4 实现 `LifecycleGuard` 接口族(`pkg/pluginhost/lifecycle_guard.go`)与并发调用 / 超时 / panic recover 框架
- [ ] 13.5 在 `plugin.Uninstall` 流程中调用 `CanUninstall` 钩子并聚合否决
- [ ] 13.6 在 `plugin.Disable` 流程中调用 `CanDisable` 钩子(global)或 `CanTenantDisable`(tenant_scoped)
- [ ] 13.7 在 `tenant.Delete` 流程中调用 `CanTenantDelete` 钩子
- [ ] 13.8 实现 `--force` 通道(配置开关 `plugin.allow_force_uninstall`)与平台审计
- [ ] 13.9 单元测试:钩子聚合、超时 fail-safe、panic recover、force 通道

## 14. 路由与权限改造

- [ ] 14.1 改造 `pluginruntimecache` / `plugin.routing` 的请求时过滤逻辑:按 (tenant, plugin) 启用状态返回 404
- [ ] 14.2 改造 `permission` 中间件:权限解析按当前租户过滤未启用插件
- [ ] 14.3 引入"平台权限点 `platform:*` 前缀"约束:角色保存时校验前缀与 `is_platform_role` 一致性
- [ ] 14.4 单元测试:覆盖路由 404 / 菜单隐藏 / 权限点过滤

## 15. org-center 插件租户化改造

- [ ] 15.1 修改 `apps/lina-plugins/org-center/manifest/sql/001-org-center-schema.sql`,所有表加 `tenant_id`,索引升级
- [ ] 15.2 修改 mock-data/*.sql,默认写入 `tenant_id=0`(单租户场景体验不变)
- [ ] 15.3 修改 `org-center` plugin.yaml:`scope_nature: tenant_aware`、`default_install_mode: global`、`default_for_new_tenants: true`
- [ ] 15.4 改造 service/dao 实现:所有查询走 `tenantcap.Apply`,所有写入填 `tenant_id`
- [ ] 15.5 实现 `tenantcap.LifecycleSubscriber`,订阅 `tenant.created` 创建租户根部门、`tenant.deleted` 级联清理
- [ ] 15.6 改造 `orgcap.Provider` 实现按 `bizctx.TenantId` 过滤
- [ ] 15.7 单元测试自包含,覆盖租户隔离、事件订阅幂等、级联清理

## 16. 其他既有插件租户化改造

- [ ] 16.1 `monitor-loginlog`:表加 `tenant_id`、`acting_user_id`、`on_behalf_of_tenant_id`、`is_impersonation`;dao/service/controller 改造;plugin.yaml `scope_nature: tenant_aware, default_install_mode: tenant_scoped`
- [ ] 16.2 `monitor-online`:表加 `tenant_id`,会话查询/踢人接口加租户校验;plugin.yaml 同上
- [ ] 16.3 `monitor-operlog`:表加 tenant 与 impersonation 字段;controller 支持 `action_kind` 筛选;plugin.yaml 同上
- [ ] 16.4 `monitor-server`:plugin.yaml `scope_nature: platform_only`;租户管理员视图不可见
- [ ] 16.5 `content-notice`:表加 `tenant_id`,通知按租户隔离;plugin.yaml `scope_nature: tenant_aware, default_install_mode: tenant_scoped`
- [ ] 16.6 `demo-control`、`plugin-demo-source`、`plugin-demo-dynamic`:加 tenant 字段(若有持久化);plugin.yaml `scope_nature: tenant_aware`
- [ ] 16.7 各插件 i18n 资源新增 `plugin.<id>.uninstall_blocked.*` 等 reason key 翻译

## 17. 后端 API 与 DTO

- [ ] 17.1 所有平台 API DTO 在 `apps/lina-plugins/multi-tenant/backend/api/platform/v1/` 定义,英文 dc + eg
- [ ] 17.2 所有租户 API DTO 在 `apps/lina-plugins/multi-tenant/backend/api/tenant/v1/` 定义
- [ ] 17.3 在 g.Meta 上声明 `permission` 标签(平台 API:`platform:*`;租户 API:`tenant:*`)
- [ ] 17.4 apidoc i18n:多租户插件维护自己的 `manifest/i18n/<locale>/apidoc/**/*.json`,英文置空,zh-CN 提供翻译
- [ ] 17.5 错误码 `bizerr.Code*` 在 `internal/service/<module>/<module>_code.go` 集中定义
- [ ] 17.6 单元测试:DTO 字段验证、permission 校验、i18n 完整性

## 18. 前端 工作台改造

- [ ] 18.1 在 `apps/lina-vben/apps/web-antd/src/api/` 增加平台/租户/auth 新接口客户端
- [ ] 18.2 在 `src/views/` 下增加 `platform/tenants/`、`platform/users/`、`platform/resolver-config/` 等页面
- [ ] 18.3 增加 `tenant/members/`、`tenant/plugins/` 页面(租户管理员入口)
- [ ] 18.4 改造登录页 `views/login/` 支持租户挑选器(基于解析策略呈现)
- [ ] 18.5 工作台头部增加租户标识 + 切换器(平台 vs 租户视觉区分)
- [ ] 18.6 改造路由守卫,根据 multi-tenant 启用状态联动隐藏租户 UI
- [ ] 18.7 增加 impersonation 头部红条提示与"退出 impersonation"按钮
- [ ] 18.8 增加 `views/platform/plugins/install-mode-selector.vue` 安装时选择 install_mode 的对话框
- [ ] 18.9 增加 LifecycleGuard 否决理由展示对话框(支持多 reason 聚合 + i18n 渲染 + force 二次确认)
- [ ] 18.10 路由模块 `src/router/routes/modules/platform.ts`、`tenant.ts` 注册
- [ ] 18.11 全局 Pinia store 增加 `useTenantStore` 管理当前租户、可选租户列表、是否 impersonation

## 19. 前端 i18n 资源

- [ ] 19.1 在 `apps/lina-vben/apps/web-antd/src/locales/zh-CN.json` 与 `en-US.json` 增加多租户相关 key(菜单、表单、错误、否决理由、impersonation 提示)
- [ ] 19.2 平台/租户两层菜单分组 i18n key 命名空间 `menu.platform.*` / `menu.tenant.*`
- [ ] 19.3 i18n 完整性校验脚本(CI)阻断遗漏

## 20. 配置文件与文档

- [ ] 20.1 `apps/lina-core/manifest/config/config.yaml` 增加 `tenant.*` 段(isolation.mode、cardinality、resolution.chain、subdomain.root_domain、reserved、on_ambiguous)
- [ ] 20.2 `plugin.allow_force_uninstall: true` 增加配置项与文档
- [ ] 20.3 更新根 `README.md` 与 `README.zh-CN.md` 增加多租户章节(启用步骤、解析策略、典型场景、限制)
- [ ] 20.4 在 `docs/` 或对应位置增加插件作者指南:scope_nature、install_mode、LifecycleGuard、tenant 事件订阅
- [ ] 20.5 在 `apps/lina-plugins/multi-tenant/README.md` 与 `README.zh-CN.md` 双语镜像

## 21. 单元测试

- [ ] 21.1 tenantcap 单元测试自包含
- [ ] 21.2 bizctx 单元测试
- [ ] 21.3 tenancy 中间件单元测试
- [ ] 21.4 解析责任链单元测试(每种 resolver 一组用例)
- [ ] 21.5 LifecycleGuard 钩子框架单元测试(聚合、超时、panic、force)
- [ ] 21.6 多租户插件 service 层单元测试自包含,覆盖各分支
- [ ] 21.7 org-center 改造的单元测试自包含
- [ ] 21.8 缓存隔离单元测试(集群模式 mock)

## 22. E2E 测试 — 多租户基础

- [ ] 22.1 创建 `hack/tests/e2e/multi-tenant/` 目录与 fixture(`multi-tenant-disabled` 与 `multi-tenant-enabled`)
- [ ] 22.2 TC0178 多租户启用:平台管理员安装 multi-tenant 插件,启用后系统行为正确
- [ ] 22.3 TC0179 平台管理员创建租户:基础 CRUD、code 校验、tombstone
- [ ] 22.4 TC0180 租户暂停/恢复:暂停后用户登录被拒,恢复后正常
- [ ] 22.5 TC0181 租户归档:archived 状态只读
- [ ] 22.6 TC0182 租户删除受 LifecycleGuard 保护:active 时拒绝、归档后允许
- [ ] 22.7 TC0183 多租户禁用:卸载 multi-tenant 后系统退化为单租户行为
- [ ] 22.8 TC0184 1:N 用户登录挑选租户:返回 pre_token + 列表 + select-tenant 流程
- [ ] 22.9 TC0185 切换租户重签 token:旧 token 立即作废
- [ ] 22.10 TC0186 平台管理员 impersonation 启用与退出

## 23. E2E 测试 — 跨租户隔离矩阵

- [ ] 23.1 TC0187 用户 跨租户隔离:租户 A 不可见 B 用户
- [ ] 23.2 TC0188 角色 跨租户隔离 + 平台角色仅平台管理员可见
- [ ] 23.3 TC0189 字典 跨租户隔离 + platform fallback 正确
- [ ] 23.4 TC0190 配置 跨租户隔离 + platform fallback 正确
- [ ] 23.5 TC0191 文件 跨租户隔离 + 平台共享路径正确
- [ ] 23.6 TC0192 通知 跨租户隔离 + 平台广播正确
- [ ] 23.7 TC0193 任务 跨租户隔离 + 内置系统任务平台级
- [ ] 23.8 TC0194 在线会话 跨租户隔离 + 踢人按租户校验
- [ ] 23.9 TC0195 登录日志/操作日志 跨租户隔离 + impersonation 双轨标记
- [ ] 23.10 TC0196 部门(org-center)跨租户隔离 + 租户事件触发部门初始化
- [ ] 23.11 TC0197 岗位(org-center)跨租户隔离 + 同 code 跨租户允许

## 24. E2E 测试 — 解析策略与登录流程

- [ ] 24.1 TC0198 header 解析器:登录前 X-Tenant-Code hint 命中,已登录业务请求不得覆盖 JWT TenantId
- [ ] 24.2 TC0199 subdomain 解析器:登录前子域名 hint 命中 + 保留子域名忽略
- [ ] 24.3 TC0200 jwt 解析器:Claims 命中
- [ ] 24.4 TC0201 session 解析器:登录后挑选 + 持续命中
- [ ] 24.5 TC0202 default 解析器 + ambiguous prompt 模式
- [ ] 24.6 TC0203 ambiguous reject 模式
- [ ] 24.7 TC0204 override:平台管理员 X-Tenant-Override 合法 + 普通用户被忽略
- [ ] 24.8 TC0205 解析链配置改动后台立即生效

## 25. E2E 测试 — 插件治理两层化

- [ ] 25.1 TC0206 平台管理员安装 tenant_aware 插件:install_mode 选择 global / tenant_scoped
- [ ] 25.2 TC0207 安装 platform_only 插件:install_mode 强制 global,租户管理员看不到
- [ ] 25.3 TC0208 租户管理员启用/禁用 tenant_scoped 插件:菜单/路由/权限联动
- [ ] 25.4 TC0209 install_mode 切换 global ↔ tenant_scoped:状态正确迁移
- [ ] 25.5 TC0210 LifecycleGuard 否决卸载:active 租户时 multi-tenant 拒绝卸载,聚合 reason
- [ ] 25.6 TC0211 force 通道:平台管理员强制卸载 + 双重确认 + 强审计
- [ ] 25.7 TC0212 钩子超时 / panic fail-safe
- [ ] 25.8 TC0213 default_for_new_tenants:org-center 在新租户自动启用

## 26. E2E 测试 — 平台管理员场景

- [ ] 26.1 TC0214 平台管理员可跨租户读全量数据
- [ ] 26.2 TC0215 平台管理员 impersonation 后操作日志双轨记录
- [ ] 26.3 TC0216 平台管理员强制操作单独审计类
- [ ] 26.4 TC0217 平台管理员视图与租户管理员视图 UI 差异化

## 27. 集群一致性 E2E

- [ ] 27.1 TC0218 集群模式下租户解析配置变更广播 + 立即生效
- [ ] 27.2 TC0219 跨节点缓存按租户失效隔离
- [ ] 27.3 TC0220 token revoke 跨节点广播

## 28. 性能与回归

- [ ] 28.1 联合索引性能验证:用 PG `EXPLAIN` 验证关键查询(用户列表、角色列表、菜单解析、字典 fallback)
- [ ] 28.2 现有单租户 e2e 套件在 multi-tenant 未启用 fixture 下全部通过
- [ ] 28.3 启动期一致性校验通过基线测试

## 29. 审查与归档准备

- [ ] 29.1 调用 `/lina-review` 进行 spec 与代码审查,处理审查项
- [ ] 29.2 i18n 完整性校验(zh-CN / en-US)无缺失
- [ ] 29.3 文档双语镜像同步(README、AGENTS、SKILL 等)
- [ ] 29.4 `openspec validate multi-tenant --strict` 通过
- [ ] 29.5 用户验收确认本迭代完成

## Feedback

- [x] **FB-1**: 固化多租户需求澄清:租户 code ASCII 约束、TenantId 优先级、平台 bypass/impersonation、暂停租户访问边界、tenant_scoped → global 强制启用、用户列表 membership 权威边界
- [x] **FB-2**: 同步源码插件 E2E 项目标准为 `apps/lina-plugins/<plugin-id>/hack/tests/{e2e,pages,support}/`,并更新 `AGENTS.md`
- [x] **FB-3**: 迁移框架内置源码插件既有 E2E 到各插件 `hack/tests/e2e` 所有权目录,并修正测试运行清单与校验规则
