## 1. 依赖解析器两轴语义

- [x] 1.1 在 `PluginSnapshot` 增加 `Enabled`，并在 registry 快照装配中写入全局启用态
- [x] 1.2 `InstallCheckInput` 支持 `RequireEnabled`；启用路径要求依赖 installed+enabled+version
- [x] 1.3 增加 `DependencyStatusNotEnabled` 与 `BlockerCodeDependencyNotEnabled`（API 枚举）
- [x] 1.4 `ReverseCheckInput` 支持 `OnlyEnabledDependents`；禁用仅阻断已启用下游
- [x] 1.5 补充 resolver 单测：安装不要求 enabled、启用要求 enabled、禁用忽略 disabled 下游、卸载仍挡 installed 下游

## 2. 生命周期接线与错误文案

- [x] 2.1 启用路径改用 `RequireEnabled=true` 的依赖检查
- [x] 2.2 禁用路径使用 `OnlyEnabledDependents=true` 的反向检查；卸载保持 installed 反向
- [x] 2.3 新增禁用反向错误码与 zh-CN/en-US 文案分流
- [x] 2.4 更新 lifecycle / plugin 依赖相关集成测试（含 owner disable、extlogin 类场景等价用例）

## 3. 验证

- [x] 3.1 运行变更相关 Go 测试包并通过
- [x] 3.2 `openspec validate plugin-lifecycle-dependency-axes --strict`
- [x] 3.3 执行 lina-review 并处理必要问题
