## REMOVED Requirements

### Requirement: 宿主必须提供匿名健康探测端点

宿主 SHALL 在公共路由组下提供 `GET /api/v1/health`，无需登录状态即可访问。该端点必须通过标准 API DTO 和控制器流程暴露，返回服务自检结果，并包含一个轻量级数据库探测。数据库不可用或探测超时时，必须返回 HTTP `503` 和稳定的脱敏不可用原因。健康时，必须返回 HTTP `200` 和 `{"status":"ok","mode":"<single|master|slave>"}`。探测超时 SHALL 由配置键 `health.timeout` 控制，默认 `5s`，解析为 `time.Duration`。

#### Scenario: 数据库健康时健康探测返回 200

- **当** 调用方匿名访问 `GET /api/v1/health`
- **且** 数据库可达且探测在超时内完成
- **则** 端点返回 `200`，响应体包含 `status="ok"` 加当前部署 `mode`
- **且** 端点不要求 `Authorization` 头

#### Scenario: 数据库不可用时健康探测返回 503

- **当** 调用方匿名访问 `GET /api/v1/health`
- **且** 数据库探测未在 `health.timeout` 内返回
- **则** 端点返回 `503`，响应体包含 `status="unavailable"` 加稳定的脱敏原因
- **且** 原始数据库、网络或模式错误必须仅记录日志，不得返回给匿名调用方

#### Scenario: 健康探测不挂载在受保护路由下

- **当** 服务启动且路由注册完成时
- **则** `/api/v1/health` 必须注册在公共路由组下
- **且** Auth 和 Permission 中间件不得拦截该路由

## ADDED Requirements

### Requirement: 宿主不得内建业务健康检查接口

宿主 SHALL NOT 提供内建匿名业务健康检查 HTTP API、健康检查 DTO、健康检查控制器或`health.timeout`等专用配置入口。具体业务应用、交付镜像或业务插件需要健康检查时，必须在自身边界定义业务健康语义和路由，避免`lina-core`将数据库探测、集群角色或业务可用性固化为框架级契约。

#### Scenario: 主框架不注册健康检查端点

- **当** LinaPro 宿主启动并完成静态 API 路由注册时
- **则** 主框架不得注册`GET /api/v1/health`
- **且** 主框架不得保留`api/health`或`internal/controller/health`生产代码

#### Scenario: 主框架不读取健康检查配置

- **当** 宿主配置服务初始化时
- **则** 配置服务不得暴露`GetHealth`或`HealthConfig`
- **且** 交付配置模板不得声明`health.timeout`

#### Scenario: 业务自行定义健康语义

- **当** 业务应用需要对外提供健康检查时
- **则** 业务必须在自身应用、交付层或插件边界实现健康检查
- **且** 业务健康接口不得依赖主框架内建的`/api/v1/health`

#### Scenario: 保留已认证运行诊断

- **当** 需要验证集群 coordination 状态时
- **则** 已认证调用方可以继续通过系统信息能力读取 coordination 诊断
- **且** 该诊断不得退化为匿名业务健康检查 API
