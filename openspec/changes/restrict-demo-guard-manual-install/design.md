## Design

### 安装来源标记

普通管理安装和启动配置自动启用最终都会进入插件 facade 的内部 `install` 路径。为避免插件仅能看到相同的 `BeforeInstall` 输入，本次在宿主内部 `InstallOptions` 中增加仅限启动引导设置的标记，并通过 `pluginhost` 提供的上下文 helper 传入生命周期回调。

该标记只由 `BootstrapAutoEnable` 的目标插件安装设置；自动依赖安装不继承该标记，避免把“被某个自动启用插件依赖”误解释为“该插件自身已由 `plugin.autoEnable` 显式配置”。

### 演示控制插件策略

`linapro-ops-demo-guard` 注册 `BeforeInstall`：

- 当上下文显示当前安装来自启动 `plugin.autoEnable` 引导时，返回允许。
- 其他安装来源返回 veto 原因，阻断安装 SQL、菜单同步和 registry installed 状态写入。

### 风险与边界

- 本次不改变插件启用后的只读中间件行为。
- 本次不新增 REST API 或修改 API 方法语义。
- 集群模式下仍只有主节点执行共享安装副作用，从节点等待 registry 收敛；标记仅在主节点执行启动安装生命周期时生效。
- 如果运维已误安装该插件，本次不会自动卸载；变更只防止后续普通管理安装。
