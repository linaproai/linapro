## Why

`linapro-ops-demo-guard` 是用于演示环境的全局只读保护插件。一旦误从插件管理页面安装并启用，会拦截系统写操作和插件治理写操作，给普通管理环境带来恢复成本。

该插件应只通过运维显式配置 `plugin.autoEnable` 在宿主启动期间自动安装和启用，避免页面误安装。

## What Changes

- 为源码插件 `BeforeInstall` 生命周期提供启动自动启用安装上下文。
- `linapro-ops-demo-guard` 实现 `BeforeInstall` 回调：仅允许 `plugin.autoEnable` 启动引导安装，拒绝普通插件管理安装。
- 保持已通过配置启用后的只读演示保护行为不变。
- 补充单元测试覆盖普通安装拒绝、启动自动启用允许和演示控制插件自身回调判断。

## Capabilities

### Modified Capabilities

- `demo-control-guard`：演示控制插件不得通过插件管理页面安装，只能通过 `plugin.autoEnable` 启动自动启用安装。
- `plugin-startup-bootstrap`：启动自动启用安装向源码插件 `BeforeInstall` 生命周期暴露可信上下文，供插件区分配置引导和普通管理操作。

## Impact

- 后端：`pluginhost` 生命周期上下文、插件安装选项、启动自动启用路径、`linapro-ops-demo-guard` 源码插件注册。
- 前端：不新增页面交互；页面安装将收到现有生命周期前置条件错误。
- i18n：新增演示控制插件安装拒绝原因的中英文运行时错误资源。
- 缓存一致性：不新增缓存；启动自动启用仍复用既有插件 registry 状态收敛、启用快照刷新和运行时缓存修订通知。
- 数据权限：不新增或扩大数据操作接口；插件安装仍属于平台插件治理写操作，由既有平台治理权限控制。
