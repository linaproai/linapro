## Context

HTTP 启动会通过 `WithStartupDataSnapshot` 创建插件治理表的启动期快照，并在 `BootstrapAutoEnable`、插件路由绑定和后续预热阶段复用。源码插件自动启用路径先执行 `Install`，再执行 `Enable`；安装阶段调用 `applySourcePluginStableState` 将 `sys_plugin.installed` 更新为已安装，但该 helper 只写数据库，没有同步当前启动快照。

因此，在携带启动快照的 HTTP 启动上下文中，`Enable` 内部的 `CheckIsInstalled` 仍从旧快照读取到未安装状态，触发 `CodePluginNotInstalled`，最终包装为启动自动启用失败。

## Goals / Non-Goals

**Goals:**

- 源码插件自动安装后，同一启动上下文中的后续启用检查必须读取到已安装状态。
- 保持源码插件手动安装、卸载、启用和禁用的现有行为。
- 增加能够覆盖共享启动快照路径的回归测试。
- 明确 i18n 和缓存一致性影响。

**Non-Goals:**

- 不修改 `plugin.autoEnable` 配置格式。
- 不改动态插件自动启用和授权快照语义。
- 不引入进程级插件状态缓存或新的分布式协调机制。
- 不修改数据库 schema、SQL seed 或插件清单结构。

## Decisions

### 决策一：在源码插件稳定状态写入后刷新启动快照

`applySourcePluginStableState` 是源码插件安装、卸载和回滚路径更新 registry 稳定状态的集中点。修复应在该方法完成数据库更新后调用 catalog 的启动快照刷新能力，使携带启动快照的上下文立刻看到最新 registry 行。

理由：

- 根因是写后快照陈旧，集中修复比在自动启用调用点绕过快照更直接。
- 该方法也覆盖卸载和回滚路径，可避免同类启动上下文读到旧生命周期状态。
- 不改变无启动快照上下文的行为；没有快照时刷新只是普通数据库回读。

替代方案：在 `Enable` 前重新构造启动快照。该方案会破坏启动 SQL 优化目标，并且只修复自动启用调用点，不能覆盖其他源码插件生命周期写入后读路径。

### 决策二：用携带启动快照的单元测试复现问题

新增测试应先调用 `WithStartupDataSnapshot` 构造与 HTTP 启动一致的上下文，再执行 `BootstrapAutoEnable`。断言插件最终为已安装且已启用。

理由：

- 现有测试直接使用普通 context，无法暴露启动快照陈旧问题。
- 该测试聚焦后端生命周期行为，失败信号明确，不需要启动真实 HTTP 服务。

## Risks / Trade-offs

- [Risk] 稳定状态写入后刷新 registry 会增加一次回读。
  Mitigation：该路径只发生在真实生命周期状态变更时，不影响 no-op 清单同步；相比启动失败，成本可接受。
- [Risk] 启动快照刷新遗漏 release 或菜单投影。
  Mitigation：本次失败只涉及 registry 安装状态；release、菜单和资源引用仍由现有同步路径维护。
- [Risk] 集群模式下从节点读到本节点快照旧状态。
  Mitigation：只有主节点执行共享生命周期写入；其他节点继续通过现有轮询等待数据库状态收敛。快照仅是单次启动上下文内的本地投影，不承担跨实例一致性。

## Migration Plan

无需数据迁移。部署新版本后，使用 `plugin.autoEnable` 自动启用源码插件的启动流程会在安装后继续完成启用。若需回滚，恢复本次代码改动即可；数据库中已安装/已启用的插件状态仍可由现有治理页面或启动配置管理。
