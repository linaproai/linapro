## ADDED Requirements

### Requirement: DAO 必须经 tenantcap.Apply
所有读取带 `tenant_id` 的租户敏感 `sys_*` 表、租户作用域运行时状态表与插件业务表的 service 层代码 SHALL 通过 `tenantcap.Apply(ctx, model, "tenant_id")` 注入租户过滤;直接 `dao.Xxx.Ctx(ctx).Where(...)` 而不经接缝的代码视为违反规范。平台控制面表(`sys_locker`、`sys_menu`、`sys_plugin*`、`sys_notify_channel` 等)不经 tenantcap 注入,但只能通过宿主/平台治理 service 访问。

#### Scenario: 直接 DAO 调用被审查阻断
- **WHEN** 代码 `dao.SysFile.Ctx(ctx).Where("status", 1).Scan(&list)` 未先 Apply
- **THEN** `lina-review` 标记为不通过
- **AND** 必须重构为 `model := dao.SysFile.Ctx(ctx).Where("status", 1); model, _, _ := tenantcap.Apply(ctx, model, "tenant_id"); model.Scan(&list)`

### Requirement: DO 写入必须填 tenant_id
所有对应带 `tenant_id` 表的 `do.*` 数据操作对象的 `Insert`/`Update` 调用 SHALL 在 service 层显式填入 `TenantId = bizctx.TenantId(ctx)`;不允许依赖 default 0 隐式写入。无 `tenant_id` 的平台控制面 DO 不适用此规则。

#### Scenario: 显式 TenantId
- **WHEN** service 调用 `dao.SysUser.Ctx(ctx).Data(do.SysUser{...}).Insert()`
- **THEN** DO 字段 `TenantId` 必须明确填值
- **AND** 平台管理员跨租户写入需走 impersonation 或专用平台 API/service,显式指定目标租户并记录审计

### Requirement: 否决钩子 reason 走 i18n key
插件实现 `LifecycleGuard.*` 接口时,reason 字符串 SHALL 是 i18n key,禁止硬编码中文/英文;插件需在自己的 `manifest/i18n/<locale>/plugin.json` 中维护翻译。

#### Scenario: 硬编码 reason 被阻断
- **WHEN** 插件代码 `return false, "尚有租户存在,无法卸载"`
- **THEN** lina-review 标记为不通过
- **AND** 必须改为 `return false, "plugin.<id>.uninstall_blocked.tenants_exist"`

### Requirement: 跨租户操作必须用专用 API/Service
普通业务路径 SHALL 不允许跨租户读写;跨租户操作必须经 `/platform/*` 接口或 `*.PlatformXxx` host service 方法;设计上无中间态。

#### Scenario: 静态扫描验证
- **WHEN** 后端扫描 service 层代码
- **THEN** 不存在 `dao.X.Ctx(ctx).Where("tenant_id", ...)` 这类绕过 tenantcap 的硬编码

### Requirement: 启动期租户上下文构造
后台 worker、定时任务、消息消费者等异步路径在创建新 ctx 时 SHALL 显式构造租户上下文(从任务元数据或消息载荷读取 `tenant_id`);禁止使用空 ctx 操作租户敏感 `sys_*` 业务数据。

#### Scenario: 任务执行租户上下文
- **WHEN** 某定时任务回调被触发
- **THEN** 调度器先构造 `bizctx` with TenantId,再调用 handler
- **AND** handler 内访问租户敏感 sys_* 数据自动按租户过滤
