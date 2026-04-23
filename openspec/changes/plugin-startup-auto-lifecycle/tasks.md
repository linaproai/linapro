## 1. 配置与启动链路

- [ ] 1.1 扩展 `plugin.startup` 配置模型、模板示例与配置校验，支持 `desiredState`、`required`、`blockUntilReady`、`readyTimeout` 和动态插件 `authorization`
- [ ] 1.2 在宿主启动链路中新增插件 startup bootstrap 阶段，并把执行顺序前移到插件路由注册、插件 cron 接线和动态 bundle 预热之前

## 2. 生命周期 bootstrap 实现

- [ ] 2.1 实现源码插件启动策略执行器，支持 `manual` / `installed` / `enabled` 的最低目标状态语义与集群主节点共享动作保护
- [ ] 2.2 实现动态插件启动策略执行器，复用授权快照、`desired_state/current_state` 与 targeted reconcile 机制推进到 `installed` / `enabled`
- [ ] 2.3 实现 `required` / 可选插件的失败处理、等待超时与 enabled snapshot 刷新逻辑，保证 bootstrap 结束后再进行后续插件接线

## 3. 测试与验证

- [ ] 3.1 补充 `plugin.startup` 配置解析与非法配置校验测试
- [ ] 3.2 补充源码插件 startup bootstrap 测试，覆盖发现态保持、自动安装、自动启用与非破坏性最低目标状态语义
- [ ] 3.3 补充动态插件 startup bootstrap 测试，覆盖授权存在/缺失、`required=true/false`、单节点与集群主从收敛路径

## 4. 文档与运维说明

- [ ] 4.1 更新插件相关技术文档与配置说明，给出源码插件和动态插件的启动策略示例及运维注意事项
