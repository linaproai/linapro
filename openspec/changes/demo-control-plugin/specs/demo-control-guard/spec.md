## ADDED Requirements

### Requirement: 宿主必须通过`plugin.autoEnable`控制演示能力开关

系统 MUST 以`plugin.autoEnable`是否包含`demo-control`作为演示能力的唯一开关入口。

#### Scenario: 默认配置保持演示能力关闭
- **WHEN** 宿主使用默认交付配置启动
- **THEN** 默认配置不会强制把`demo-control`加入`plugin.autoEnable`
- **AND** 未启用`demo-control`的实例不会因为该能力默认阻断写操作

#### Scenario: 在自动启用列表中开启演示能力
- **WHEN** 部署环境把`demo-control`加入宿主主配置文件中的`plugin.autoEnable`
- **THEN** 宿主在启动阶段自动安装并启用该插件
- **AND** 演示控制中间件会在`/*`请求链路中生效

### Requirement: 宿主必须随源码树交付演示控制源码插件

系统 MUST 随源码树交付官方源码插件`demo-control`，使部署环境可以通过`plugin.autoEnable`启用该能力。

#### Scenario: 宿主发现演示控制源码插件
- **WHEN** 宿主扫描源码插件目录并同步插件注册表
- **THEN** 宿主能够发现`demo-control`源码插件
- **AND** 运维可以通过`plugin.autoEnable`决定是否在启动阶段启用该插件

### Requirement: 演示控制插件必须在启用时阻断系统写操作

系统 MUST 在`demo-control`启用时，于`/*`作用域下基于系统请求的`HTTP Method`阻断写操作请求，并保留查询型请求能力。

#### Scenario: 演示控制插件未启用时不拦截写操作
- **WHEN** `demo-control` 未被宿主启用
- **THEN** `POST`、`PUT`、`DELETE` 请求不会因为演示控制插件被额外拒绝

#### Scenario: 演示控制插件启用时允许查询型请求
- **WHEN** `demo-control` 已被宿主启用
- **AND** 请求命中系统请求链路且`HTTP Method`为 `GET`、`HEAD` 或 `OPTIONS`
- **THEN** 演示控制插件允许请求继续进入后续处理链路

#### Scenario: 演示控制插件启用时拒绝写操作请求
- **WHEN** `demo-control` 已被宿主启用
- **AND** 请求命中系统请求链路且`HTTP Method`为 `POST`、`PUT` 或 `DELETE`
- **THEN** 演示控制插件拒绝该请求并返回明确的只读演示提示
- **AND** 请求不会继续进入后续业务处理链路

#### Scenario: 演示控制插件覆盖非 API 前缀写请求
- **WHEN** `demo-control` 已被宿主启用
- **AND** 请求路径不在`/api/v1`前缀下
- **AND** 请求的`HTTP Method`为 `POST`、`PUT` 或 `DELETE`
- **THEN** 演示控制插件同样拒绝该请求

### Requirement: 演示控制插件必须保留受控插件治理白名单

系统 MUST 在`demo-control`启用时保留插件管理中的最小治理白名单：除`demo-control`自身外，其他插件仍可执行安装、卸载、启用与禁用操作。

#### Scenario: 演示控制插件启用时允许其他插件安装
- **WHEN** `demo-control` 已被宿主启用
- **AND** 请求为 `POST /api/v1/plugins/{id}/install`
- **AND** `{id}` 不等于 `demo-control`
- **THEN** 演示控制插件允许该请求继续进入插件安装处理链路

#### Scenario: 演示控制插件启用时允许其他插件启用或禁用
- **WHEN** `demo-control` 已被宿主启用
- **AND** 请求为 `PUT /api/v1/plugins/{id}/enable` 或 `PUT /api/v1/plugins/{id}/disable`
- **AND** `{id}` 不等于 `demo-control`
- **THEN** 演示控制插件允许该请求继续进入插件状态更新处理链路

#### Scenario: 演示控制插件启用时允许其他插件卸载
- **WHEN** `demo-control` 已被宿主启用
- **AND** 请求为 `DELETE /api/v1/plugins/{id}`
- **AND** `{id}` 不等于 `demo-control`
- **THEN** 演示控制插件允许该请求继续进入插件卸载处理链路

#### Scenario: 演示控制插件启用时拒绝修改自身治理状态
- **WHEN** `demo-control` 已被宿主启用
- **AND** 请求试图对 `demo-control` 自身执行安装、卸载、启用或禁用
- **THEN** 演示控制插件拒绝该请求并返回明确的只读演示提示

### Requirement: 演示控制插件必须保留最小会话白名单

系统 MUST 在`demo-control`启用时保留登录和登出能力，避免演示环境失去基本可用性。

#### Scenario: 演示控制插件启用时允许登录接口
- **WHEN** `demo-control` 已被宿主启用
- **AND** 请求为 `POST /api/v1/auth/login`
- **THEN** 演示控制插件允许请求继续进入认证处理链路

#### Scenario: 演示控制插件启用时允许登出接口
- **WHEN** `demo-control` 已被宿主启用
- **AND** 请求为 `POST /api/v1/auth/logout`
- **THEN** 演示控制插件允许请求继续进入后续处理链路
