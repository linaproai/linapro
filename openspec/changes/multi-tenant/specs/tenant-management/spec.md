## ADDED Requirements

### Requirement: 租户主体表结构
`multi-tenant` 插件 SHALL 维护 `plugin_multi_tenant_tenant` 表,字段至少包含 `id`(主键)、`code`(全局唯一,只允许 ASCII 小写字母/数字/连字符且符合 `[a-z0-9-]{2,32}`)、`name`(显示名,可包含中文或其他 Unicode 字符)、`status`(active/suspended/deleted)、`remark`、`created_at`、`updated_at`、`deleted_at`(软删除);租户主体表、API DTO 与管理页面 SHALL NOT 暴露套餐/plan 字段。

#### Scenario: 创建租户
- **WHEN** 平台管理员调用 `POST /platform/tenants` 创建租户 `code=acme, name=ACME 集团`
- **THEN** 系统校验 code 不在保留子域名列表,符合 ASCII 字符集与长度,且不允许中文或其他 Unicode 字符
- **AND** 写入 `plugin_multi_tenant_tenant` 一行 status=active
- **AND** 直接执行新租户插件默认启用策略的领域服务

#### Scenario: 中文租户编码被拒
- **WHEN** 平台管理员尝试创建租户 `code=研发部, name=研发部`
- **THEN** 返回 `bizerr.CodeTenantCodeInvalid`
- **AND** 提示租户 code 只能使用 `[a-z0-9-]{2,32}`,显示名称可使用中文

#### Scenario: 租户编码冲突
- **WHEN** 创建租户 `code=acme` 时已存在同 code 租户
- **THEN** 返回 `bizerr.CodeTenantCodeDuplicated`
- **AND** 不写入数据

### Requirement: 租户生命周期状态机
租户 SHALL 在 `active`、`suspended`、`deleted`(软删除)三个状态间流转;只允许以下迁移:
- active → suspended(暂停)
- suspended → active(恢复)
- active/suspended → deleted(经平台管理员二次确认 + LifecycleGuard 通过)

#### Scenario: 暂停租户
- **WHEN** 平台管理员暂停租户 T
- **THEN** T 的成员登录被拒绝(返回 `TENANT_SUSPENDED`)
- **AND** T 的现有 token 立即作废
- **AND** T 的数据保留,租户成员不可读不可写
- **AND** 平台管理员可通过 `/platform/*` 执行只读排障与恢复操作;业务写入必须通过 impersonation 或专用平台 API 并记录审计

#### Scenario: 删除租户
- **WHEN** 平台管理员删除 active 或 suspended 租户 T
- **THEN** 系统执行所有插件的 `CanTenantDelete` 钩子
- **AND** 全部通过后 `plugin_multi_tenant_tenant` 中该行 `deleted_at` 置为当前时间
- **AND** 不写入 outbox 或触发未实现的跨插件事件总线

### Requirement: 平台管理员权限要求
租户 CRUD 操作 SHALL 要求请求用户为平台管理员(`bizctx.TenantId = 0`、有效数据权限为全部数据权限且持有对应 `system:tenant:*` 功能权限);非平台管理员请求 `/platform/tenants/*` 必须返回 403。

#### Scenario: 租户管理员尝试访问平台 API
- **WHEN** 租户管理员请求 `GET /platform/tenants`
- **THEN** 返回 403 `bizerr.CodePlatformPermissionRequired`
- **AND** 操作日志记录一条权限拒绝事件

### Requirement: 租户列表查询与筛选
`GET /platform/tenants` SHALL 支持按 `code`、`name`、`status` 过滤与分页;查询结果不暴露 `deleted_at != NULL` 的租户,并在列表项中返回 `created_at` 对应的 `createdAt` 字段用于管理页面展示。

#### Scenario: 默认查询排除已删除
- **WHEN** 平台管理员调用 `GET /platform/tenants`
- **THEN** 仅返回 `deleted_at IS NULL` 的租户
- **AND** 默认按 `created_at desc` 排序

### Requirement: 租户管理页面行内运营入口
租户管理页面 SHALL 在每个租户行内提供代操作入口;代操作按钮 SHALL 在鼠标悬浮时展示简短说明。租户管理页面 SHALL NOT 提供成员管理按钮、成员抽屉、成员弹窗或跳转系统用户管理的行内入口;租户成员关系、租户筛选用户列表、用户租户归属调整与租户内角色维护 SHALL 统一由系统用户管理页面承载。

#### Scenario: 查看代操作说明
- **WHEN** 平台管理员将鼠标悬浮到租户列表某行的代操作按钮
- **THEN** 前端展示简短提示,说明该操作会以平台身份进入该租户视角

#### Scenario: 租户管理页不承载成员管理
- **WHEN** 平台管理员查看租户管理列表
- **THEN** 行内操作不展示成员管理入口
- **AND** 前端不打开租户成员抽屉或弹窗
- **AND** 租户成员关系维护通过系统用户管理页面的租户筛选与用户租户归属字段完成

### Requirement: 租户配额暂不建模
`multi-tenant` 插件 SHALL NOT 在本轮创建租户配额表、写入配额 mock 数据或执行配额检查;租户级配额/计费能力 SHALL 由后续迭代重新设计数据模型、执行点和一致性策略。

#### Scenario: 插件安装不创建配额占位表
- **WHEN** 多租户插件安装
- **THEN** 不创建 `plugin_multi_tenant_quota` 表
- **AND** 插件不在任何业务路径上读取、写入或校验租户配额

### Requirement: 租户编码不可改且不可被复用
租户 `code` 一旦创建 SHALL 不可修改;租户被删除后,其 `code` SHALL 在被复用前保留 30 天 tombstone(`plugin_multi_tenant_tenant` 中保留软删除行,新建同 code 时拒绝)。

#### Scenario: 尝试复用已删除租户的 code
- **WHEN** 租户 `acme` 被删除 10 天后
- **AND** 平台管理员尝试创建新租户 `code=acme`
- **THEN** 返回 `bizerr.CodeTenantCodeReserved`(原因 i18n)
- **AND** 创建被拒
