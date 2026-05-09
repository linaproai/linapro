## ADDED Requirements

### Requirement: 租户主体表结构
`multi-tenant` 插件 SHALL 维护 `plugin_multi_tenant_tenant` 表,字段至少包含 `id`(主键)、`code`(全局唯一,只允许 ASCII 小写字母/数字/连字符且符合 `[a-z0-9-]{2,32}`)、`name`(显示名,可包含中文或其他 Unicode 字符)、`status`(active/suspended/archived/deleted)、`plan`(套餐占位)、`created_at`、`updated_at`、`deleted_at`(软删除)。

#### Scenario: 创建租户
- **WHEN** 平台管理员调用 `POST /platform/tenants` 创建租户 `code=acme, name=ACME 集团`
- **THEN** 系统校验 code 不在保留子域名列表,符合 ASCII 字符集与长度,且不允许中文或其他 Unicode 字符
- **AND** 写入 `plugin_multi_tenant_tenant` 一行 status=active
- **AND** 触发 `tenant.created` 事件

#### Scenario: 中文租户编码被拒
- **WHEN** 平台管理员尝试创建租户 `code=研发部, name=研发部`
- **THEN** 返回 `bizerr.CodeTenantCodeInvalid`
- **AND** 提示租户 code 只能使用 `[a-z0-9-]{2,32}`,显示名称可使用中文

#### Scenario: 租户编码冲突
- **WHEN** 创建租户 `code=acme` 时已存在同 code 租户
- **THEN** 返回 `bizerr.CodeTenantCodeDuplicated`
- **AND** 不写入数据

### Requirement: 租户生命周期状态机
租户 SHALL 在 `active`、`suspended`、`archived`、`deleted`(软删除)四个状态间流转;只允许以下迁移:
- active → suspended(暂停)
- suspended → active(恢复)
- active/suspended → archived(归档,只读)
- archived → deleted(经平台管理员二次确认 + LifecycleGuard 通过)

#### Scenario: 暂停租户
- **WHEN** 平台管理员暂停租户 T
- **THEN** T 的成员登录被拒绝(返回 `TENANT_SUSPENDED`)
- **AND** T 的现有 token 立即作废
- **AND** T 的数据保留,租户成员不可读不可写
- **AND** 平台管理员可通过 `/platform/*` 执行只读排障与恢复操作;业务写入必须通过 impersonation 或专用平台 API 并记录审计

#### Scenario: 删除租户
- **WHEN** 平台管理员删除已归档的租户 T
- **THEN** 系统执行所有插件的 `CanTenantDelete` 钩子
- **AND** 全部通过后,触发 `tenant.deleted` 事件
- **AND** 各插件级联清理本租户数据(membership、dept、post、log 等)
- **AND** `plugin_multi_tenant_tenant` 中该行 `deleted_at` 置为当前时间

### Requirement: 平台管理员权限要求
租户 CRUD 操作 SHALL 要求请求用户为平台管理员(`bizctx.TenantId = 0` 且持有 `platform:tenant:*` 权限);非平台管理员请求 `/platform/tenants/*` 必须返回 403。

#### Scenario: 租户管理员尝试访问平台 API
- **WHEN** 租户管理员请求 `GET /platform/tenants`
- **THEN** 返回 403 `bizerr.CodePlatformPermissionRequired`
- **AND** 操作日志记录一条权限拒绝事件

### Requirement: 租户列表查询与筛选
`GET /platform/tenants` SHALL 支持按 `code`、`name`、`status` 过滤与分页;查询结果不暴露 `deleted_at != NULL` 的租户(除非显式查询 archived)。

#### Scenario: 默认查询排除已删除
- **WHEN** 平台管理员调用 `GET /platform/tenants`
- **THEN** 仅返回 `deleted_at IS NULL` 的租户
- **AND** 默认按 `created_at desc` 排序

### Requirement: 租户配额扩展位
`plugin_multi_tenant_quota` 表 SHALL 预留为 `(tenant_id, quota_key, quota_value, unit)` 结构;首版不强制任何配额检查,仅作为后续配额功能的承载结构。

#### Scenario: 配额表存在但不参与执行
- **WHEN** 多租户插件安装
- **THEN** `plugin_multi_tenant_quota` 表被创建
- **AND** 插件不在任何业务路径上读取或校验该表
- **AND** 该表上的 INSERT 是允许的,但不影响业务行为

### Requirement: 租户编码不可改且不可被复用
租户 `code` 一旦创建 SHALL 不可修改;租户被删除后,其 `code` SHALL 在被复用前保留 30 天 tombstone(`plugin_multi_tenant_tenant` 中保留软删除行,新建同 code 时拒绝)。

#### Scenario: 尝试复用已删除租户的 code
- **WHEN** 租户 `acme` 被删除 10 天后
- **AND** 平台管理员尝试创建新租户 `code=acme`
- **THEN** 返回 `bizerr.CodeTenantCodeReserved`(原因 i18n)
- **AND** 创建被拒
