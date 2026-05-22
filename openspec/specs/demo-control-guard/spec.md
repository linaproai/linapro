# 演示控制守卫

## Purpose

定义官方 `linapro-ops-demo-guard` 源码插件在启用演示模式时如何启用只读演示保护，同时保留必要的会话访问并阻断插件治理写操作。
## Requirements
### Requirement:演示只读模式由插件的启用状态控制系统

系统 SHALL 将 `linapro-ops-demo-guard` 的已安装并启用状态视为演示保护的运行时开关。`plugin.autoEnable` 控制启动安装和启用；启动后不得将其视为单独的运行时开关。`linapro-ops-demo-guard` 本身 SHALL 拒绝普通插件管理安装，只允许通过 `plugin.autoEnable` 启动自动启用完成安装。

#### Scenario:默认配置下演示保护保持禁用
- **当** 宿主以默认交付配置启动且 `plugin.autoEnable` 不包含 `linapro-ops-demo-guard` 时
- **则** 宿主不安装或启用 `linapro-ops-demo-guard`
- **且** 从未启用该插件的部署默认不阻断写操作

#### Scenario:配置自动启用激活演示保护
- **当** 运维人员在 `plugin.autoEnable` 中配置 `linapro-ops-demo-guard`
- **且** 宿主启动引导安装并启用该插件
- **则** linapro-ops-demo-guard 中间件对后续请求生效
- **且** 写请求被只读演示规则阻断

#### Scenario:拒绝插件管理页面安装
- **当** 管理员从插件管理页面或管理 API 对 `linapro-ops-demo-guard` 发起普通安装
- **则** `linapro-ops-demo-guard` 的 `BeforeInstall` 生命周期回调拒绝该安装
- **且** 宿主不得执行该插件安装 SQL、菜单同步或 installed 状态写入

### Requirement:宿主必须随源码树交付 linapro-ops-demo-guard 源码插件

系统 SHALL 交付名为 `linapro-ops-demo-guard` 的官方源码插件，使部署可通过启动配置启用该能力。该插件 SHALL 继续被源码插件发现和注册表同步识别，但普通插件治理安装入口不得安装该插件。

#### Scenario:宿主发现 linapro-ops-demo-guard 源码插件
- **当** 宿主扫描源码插件并同步注册表数据时
- **则** 发现 `linapro-ops-demo-guard`
- **且** 运维人员可通过 `plugin.autoEnable` 决定是否在启动时安装并启用
- **且** 管理页面普通安装该插件时被生命周期前置条件拒绝

### Requirement:linapro-ops-demo-guard 插件启用时必须阻断系统写操作

启用时，linapro-ops-demo-guard SHALL 按 HTTP 方法语义阻断系统写请求，同时允许读式请求。

#### Scenario:禁用时无写拦截
- **当** `linapro-ops-demo-guard` 未启用时
- **则** `POST`、`PUT` 和 `DELETE` 请求不被 linapro-ops-demo-guard 拒绝

#### Scenario:查询式请求保持允许
- **当** `linapro-ops-demo-guard` 已启用
- **且** 请求使用 `GET`、`HEAD` 或 `OPTIONS` 时
- **则** linapro-ops-demo-guard 允许请求继续

#### Scenario:写请求被拒绝
- **当** `linapro-ops-demo-guard` 已启用
- **且** 请求使用 `POST`、`PUT` 或 `DELETE` 时
- **则** linapro-ops-demo-guard 以清晰的只读演示消息拒绝请求
- **且** 请求不继续进入业务处理

### Requirement:linapro-ops-demo-guard 插件必须保留最小会话白名单

系统 SHALL 在 linapro-ops-demo-guard 启用时保留登录和退出行为，使演示环境保持可用。

#### Scenario:登录保持允许
- **当** `linapro-ops-demo-guard` 已启用
- **且** 请求为 `POST /api/v1/auth/login` 时
- **则** linapro-ops-demo-guard 允许请求继续

#### Scenario:退出保持允许
- **当** `linapro-ops-demo-guard` 已启用
- **且** 请求为 `POST /api/v1/auth/logout` 时
- **则** linapro-ops-demo-guard 允许请求继续

### Requirement:linapro-ops-demo-guard 插件启用时必须拒绝插件治理写操作

`linapro-ops-demo-guard` 启用时，系统 SHALL 拒绝插件治理写操作，包括插件同步、动态包上传、安装、卸载、启用和禁用。插件管理的 `GET`、`HEAD` 和 `OPTIONS` 请求作为只读操作保持允许。

#### Scenario:启用 linapro-ops-demo-guard 时拒绝插件安装
- **当** `linapro-ops-demo-guard` 已启用
- **且** 请求为 `POST /api/v1/plugins/{id}/install` 时
- **则** linapro-ops-demo-guard 以只读演示消息拒绝请求

#### Scenario:拒绝插件启用和禁用请求
- **当** `linapro-ops-demo-guard` 已启用
- **且** 请求为 `PUT /api/v1/plugins/{id}/enable` 或 `PUT /api/v1/plugins/{id}/disable` 时
- **则** linapro-ops-demo-guard 以只读演示消息拒绝请求

#### Scenario:拒绝插件卸载
- **当** `linapro-ops-demo-guard` 已启用
- **且** 请求为 `DELETE /api/v1/plugins/{id}` 时
- **则** linapro-ops-demo-guard 以只读演示消息拒绝请求

#### Scenario:拒绝插件同步和上传写操作
- **当** `linapro-ops-demo-guard` 已启用
- **且** 请求为 `POST /api/v1/plugins/sync` 或 `POST /api/v1/plugins/dynamic/package` 时
- **则** linapro-ops-demo-guard 以只读演示消息拒绝请求

#### Scenario:插件管理读取保持允许
- **当** `linapro-ops-demo-guard` 已启用
- **且** 请求为使用 `GET`、`HEAD` 或 `OPTIONS` 的插件管理查询时
- **则** linapro-ops-demo-guard 允许请求继续

