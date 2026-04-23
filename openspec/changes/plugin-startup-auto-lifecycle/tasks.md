## 1. 配置与启动链路

- [x] 1.1 扩展宿主主配置文件中的 `plugin.autoEnable` 配置模型、模板示例与配置校验，支持插件 ID 列表式自动启用配置
- [x] 1.2 在宿主启动链路中新增插件 startup bootstrap 阶段，并把执行顺序前移到插件路由注册、插件 cron 接线和动态 bundle 预热之前

## 2. 生命周期 bootstrap 实现

- [x] 2.1 实现基于 `plugin.autoEnable` 的源码插件自动安装与自动启用执行器，并补齐集群主节点共享动作保护
- [x] 2.2 实现基于 `plugin.autoEnable` 的动态插件自动安装与自动启用执行器，复用既有授权快照、`desired_state/current_state` 与 targeted reconcile 机制
- [x] 2.3 实现自动启用插件 fail-fast、等待收敛与 enabled snapshot 刷新逻辑，保证 bootstrap 结束后再进行后续插件接线

## 3. 测试与验证

- [x] 3.1 补充 `plugin.autoEnable` 配置解析与非法配置校验测试
- [x] 3.2 补充源码插件 auto-enable bootstrap 测试，覆盖发现态保持、自动安装与自动启用路径
- [x] 3.3 补充动态插件 auto-enable bootstrap 测试，覆盖既有授权快照、缺少授权快照、单节点与集群主从收敛路径

## 4. 文档与运维说明

- [x] 4.1 更新插件相关技术文档与配置说明，给出宿主主配置文件中的 `plugin.autoEnable` 示例及动态插件授权前置要求

## Feedback

- [x] **FB-1**: 简化插件自动启用配置，统一改为宿主主配置文件中的 `plugin.autoEnable` 插件 ID 列表
- [x] **FB-2**: 将 `catalog/status.go` 中插件安装/启用状态值改为强类型枚举，并清理同文件中不适合枚举化的常量定义方式
- [x] **FB-3**: 在插件管理页明确标识 `plugin.autoEnable` 管理的插件，并在禁用/卸载前提示“本次操作立即生效但重启后会恢复”的风险说明
