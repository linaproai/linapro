## ADDED Requirements

### Requirement: make init 默认 PLATFORM
`make init` 执行 seed SQL 时,凡目标表携带 `tenant_id` 的 seed 数据 SHALL 全部写入 `tenant_id = 0`(PLATFORM);包括 admin 用户、平台超管角色、内置字典/配置等。平台控制面或全局配置表无 `tenant_id` 时 SHALL 保持全局唯一写入。

#### Scenario: 初始化数据归属 PLATFORM
- **WHEN** 执行 `make init`
- **THEN** 所有租户敏感 seed 行 `tenant_id = 0`
- **AND** 平台控制面 seed 行不携带 `tenant_id`
- **AND** 所有租户敏感索引 `(tenant_id, ...)` 联合形态前导列均为 0

### Requirement: make mock 支持指定租户
`make mock` 默认 SHALL 写入 PLATFORM(等同今天);可选参数 `make mock tenant=acme` MUST 将 mock 数据写入指定租户(自动 ensure 该租户存在)。

#### Scenario: 默认 mock 行为
- **WHEN** 执行 `make mock`(无参数)
- **THEN** 所有租户敏感 mock 数据 `tenant_id = 0`
- **AND** 平台控制面 mock 数据不携带 `tenant_id`

#### Scenario: 指定租户 mock
- **WHEN** 执行 `make mock tenant=acme`
- **THEN** 系统先 ensure 存在 `code=acme` 的 active 租户(若无则创建)
- **AND** 所有租户敏感 mock 数据 `tenant_id = acme.id`
- **AND** 用户自动添加到该租户 membership

### Requirement: 多租户能力开关感知
`make init` 完成后,系统启动时 SHALL 检测 `multi-tenant` 插件是否已安装并 enabled;未启用时短路 tenancy 中间件,启用时正常装配。

#### Scenario: 默认 init 后未启用 multi-tenant
- **WHEN** 执行 `make init` 后启动
- **THEN** multi-tenant 插件未自动启用
- **AND** 系统行为等价于单租户

#### Scenario: 启用后切换
- **WHEN** 平台管理员安装并启用 multi-tenant 后,所有现有 PLATFORM 数据保持可见
- **AND** 新创建的租户在自己空间独立运行
