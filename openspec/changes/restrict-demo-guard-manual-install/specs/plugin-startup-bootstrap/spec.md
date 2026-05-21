## ADDED Requirements

### Requirement:启动自动启用安装必须向源码插件生命周期暴露可信上下文

系统 SHALL 在 `BootstrapAutoEnable(ctx)` 为 `plugin.autoEnable` 明确列出的目标插件执行安装时，向源码插件 `BeforeInstall` 生命周期回调暴露该安装来自启动自动启用配置的上下文。普通插件管理安装不得携带该上下文；自动依赖预安装也不得继承该上下文，除非该依赖插件自身也被明确列入 `plugin.autoEnable`。

#### Scenario:自动启用目标插件收到启动安装上下文
- **当** `plugin.autoEnable` 包含源码插件 `x`
- **且** 宿主启动引导需要安装 `x`
- **则** `x` 的 `BeforeInstall` 回调可识别当前安装来自启动自动启用配置

#### Scenario:普通管理安装不携带启动安装上下文
- **当** 管理员从插件管理页面或管理 API 安装源码插件 `x`
- **则** `x` 的 `BeforeInstall` 回调不得看到启动自动启用安装上下文

#### Scenario:自动依赖预安装不继承目标插件上下文
- **当** `plugin.autoEnable` 包含插件 `a`
- **且** `a` 自动安装依赖插件 `b`
- **且** `b` 未被明确列入 `plugin.autoEnable`
- **则** `b` 的 `BeforeInstall` 回调不得看到启动自动启用安装上下文
