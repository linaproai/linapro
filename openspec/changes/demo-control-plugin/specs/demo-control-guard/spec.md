## ADDED Requirements

### Requirement: 宿主必须提供默认关闭的演示控制静态配置

系统 MUST 在宿主主配置文件中提供演示控制开关，并默认保持关闭状态。

#### Scenario: 默认配置关闭演示控制
- **WHEN** 宿主使用默认交付配置启动
- **THEN** `demo.control.enabled` 的默认值为 `false`
- **AND** 未显式开启演示控制的实例不会因为该能力默认阻断写操作

#### Scenario: 显式覆盖配置启用演示控制
- **WHEN** 部署环境把宿主主配置文件中的 `demo.control.enabled` 设置为 `true`
- **THEN** 宿主能够读取到演示控制开启状态
- **AND** 已启用的演示控制插件会按开启状态对请求链路生效

### Requirement: 宿主默认交付必须自动启用演示控制源码插件

系统 MUST 随源码树交付官方源码插件`demo-control`，并把它加入宿主默认启动自动启用列表。

#### Scenario: 宿主启动时自动启用演示控制插件
- **WHEN** 宿主使用默认交付配置启动并完成插件启动期 bootstrap
- **THEN** `demo-control` 会被自动安装并启用
- **AND** 该插件的全局 HTTP 中间件在宿主对外服务前已经接入统一请求链路

### Requirement: 演示控制插件必须在开启时阻断系统写操作

系统 MUST 在演示控制开关开启时，基于系统 API 的`HTTP Method`阻断写操作请求，并保留查询型请求能力。

#### Scenario: 演示控制关闭时不拦截写操作
- **WHEN** `demo.control.enabled` 为 `false`
- **THEN** `POST`、`PUT`、`DELETE` 请求不会因为演示控制插件被额外拒绝

#### Scenario: 演示控制开启时允许查询型请求
- **WHEN** `demo.control.enabled` 为 `true`
- **AND** 请求命中系统 API 链路且`HTTP Method`为 `GET`、`HEAD` 或 `OPTIONS`
- **THEN** 演示控制插件允许请求继续进入后续处理链路

#### Scenario: 演示控制开启时拒绝写操作请求
- **WHEN** `demo.control.enabled` 为 `true`
- **AND** 请求命中系统 API 链路且`HTTP Method`为 `POST`、`PUT` 或 `DELETE`
- **THEN** 演示控制插件拒绝该请求并返回明确的只读演示提示
- **AND** 请求不会继续进入后续业务处理链路

### Requirement: 演示控制插件必须保留最小会话白名单

系统 MUST 在演示控制开关开启时保留登录和登出能力，避免演示环境失去基本可用性。

#### Scenario: 演示控制开启时允许登录接口
- **WHEN** `demo.control.enabled` 为 `true`
- **AND** 请求为 `POST /api/v1/auth/login`
- **THEN** 演示控制插件允许请求继续进入认证处理链路

#### Scenario: 演示控制开启时允许登出接口
- **WHEN** `demo.control.enabled` 为 `true`
- **AND** 请求为 `POST /api/v1/auth/logout`
- **THEN** 演示控制插件允许请求继续进入后续处理链路
