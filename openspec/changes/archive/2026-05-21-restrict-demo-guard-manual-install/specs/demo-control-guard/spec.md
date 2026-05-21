## MODIFIED Requirements

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
