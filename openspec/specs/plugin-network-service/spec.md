# plugin-network-service Specification

## Purpose
TBD - created by archiving change dynamic-plugin-host-service-extension. Update Purpose after archive.
## Requirements
### Requirement: 动态插件通过宿主确认授权的 URL 模式发起出站 HTTP 请求

系统 SHALL 为动态插件提供受控的出站 HTTP 能力，插件只能访问宿主确认授权的 URL 模式，不能直接访问未授权 URL 或底层 socket。

#### Scenario: 插件调用授权 URL

- **WHEN** 插件通过网络服务发起一次出站 HTTP 请求
- **THEN** 请求目标 URL 必须命中当前插件已授权的 URL 模式
- **AND** 宿主允许插件直接向该 URL 发送 HTTP 方法、请求头和请求体
- **AND** 宿主仍保留最小必要的协议安全保护，例如禁止覆盖受保护的 hop-by-hop 头部

#### Scenario: 插件尝试访问未授权目标

- **WHEN** 插件请求访问未命中授权 URL 模式的目标地址
- **THEN** 宿主拒绝该调用
- **AND** 宿主不向 guest 暴露原始网络连接能力

#### Scenario: URL 模式支持模糊匹配

- **WHEN** 开发者在插件清单中声明类似`http://*.example.com/api`的 URL 模式
- **THEN** 宿主允许其匹配同 scheme、同主机通配规则且命中路径前缀的目标 URL
- **AND** 插件无需为同一模式下的每个 API 单独声明一条资源

#### Scenario: URL 模式按多维规则判定是否命中

- **WHEN** 宿主判定一个目标 URL 是否命中某条已授权 URL 模式
- **THEN** 宿主必须至少同时校验 scheme、host、可选 port 和 path 前缀
- **AND** scheme 必须精确匹配
- **AND** host 按不区分大小写的通配规则匹配，其中`*`按 hostname 字符串的 glob 方式匹配
- **AND** 当 URL 模式声明了 port 时，目标 URL 必须匹配同一 port；未声明 port 时视为不限制 port
- **AND** path 必须在归一化后按前缀匹配，`/api`只匹配`/api`与`/api/...`，不匹配`/api-v2`

#### Scenario: Query 与 Fragment 不参与授权匹配

- **WHEN** 插件访问的目标 URL 仅在 query string 或 fragment 上发生变化
- **THEN** 宿主不得仅因 query string 不同而拒绝访问
- **AND** URL fragment 不参与授权匹配，也不作为授权边界的一部分

#### Scenario: URL 模式未命中时默认拒绝

- **WHEN** 目标 URL 在 scheme、host、port 或 path 任一维度上未命中已授权 URL 模式
- **THEN** 宿主必须默认拒绝该请求
- **AND** 插件不能通过只调整 query、fragment 或大小写差异绕过未授权边界

### Requirement: 宿主网络服务在简化声明模型下保持失败隔离

系统 SHALL 在简化 URL 授权模型下仍保持请求失败隔离，并以平台默认保护控制宿主风险。

#### Scenario: 上游调用失败

- **WHEN** 上游超时、拒绝连接或返回超限响应
- **THEN** 宿主向插件返回结构化失败结果
- **AND** 该失败不得影响宿主进程和其他插件请求

