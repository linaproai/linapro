## ADDED Requirements

### Requirement: 插件运行时列表必须复用批量 manifest 读取

系统 SHALL 在插件 runtime state 列表、启动投影和等价高频插件状态查询路径中批量读取 manifest 数据。列表实现 MUST NOT 在按 registry、release、plugin ID 或返回行循环时逐项调用会重新扫描 source/dynamic manifests 的单项读取方法；动态插件 `.wasm` artifact 解析次数 MUST 与本次查询需要的 manifest 集合有界相关，而不是与 registry 循环中的单项查询次数相乘。

#### Scenario: runtime state 列表一次性读取 manifests

- **WHEN** 系统查询插件 runtime state 列表
- **AND** registry 表中存在多个 source 或 dynamic 插件
- **THEN** 系统在列表调用内复用同一份 manifest map、manifest snapshot 或等价批量读取结果
- **AND** 不在每个 registry 行上重新执行完整 `ScanManifests`

#### Scenario: 动态 artifact 解析次数不随 registry 单项查询放大

- **WHEN** runtime storage 中存在多个动态插件 `.wasm` artifact
- **AND** 系统构建 runtime state 列表
- **THEN** 每个需要参与本次查询的 artifact 在一次列表调用中最多被解析一次
- **AND** 测试或审查证据必须证明不存在 registry 行数乘以 artifact 数量的重复解析

### Requirement: 插件 runtime 必需依赖必须在构造或启动阶段校验

系统 SHALL 将插件 runtime 的必需依赖在构造函数、私有 composition root 或启动校验阶段显式校验。必需的 topology、menu sync、hook dispatch、JWT config、upload size、user context、session store、permission filter、cache change notifier、dependency validator 等能力如果缺失且当前路径不能正确降级，系统 MUST 返回初始化错误或启动错误，而不是通过 nil-safe no-op 静默跳过。

#### Scenario: 缺失必需 menu syncer 时启动校验失败

- **WHEN** 插件 runtime 需要同步动态插件菜单和权限
- **AND** menu syncer 未被构造或启动 wiring 注入
- **THEN** runtime wiring 校验返回明确错误
- **AND** 系统不得在生命周期路径中静默跳过菜单和权限同步

#### Scenario: 可选依赖允许显式降级

- **WHEN** 某个 runtime 依赖被设计为可选能力
- **THEN** 接口注释或 wiring 校验必须说明可选原因和降级语义
- **AND** 对应测试覆盖缺失该依赖时的预期行为
