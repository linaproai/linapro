## ADDED Requirements

### Requirement: 安装授权后源码与动态插件同权同信

经宿主安装或升级治理并处于启用状态的动态插件，SHALL 与源码插件适用同一信任级与能力准入模型。系统 MUST NOT 仅因插件 `type` 为 `dynamic` 而永久拒绝发布某一 core-owned 领域能力。

#### Scenario: 能力可用性不由 type 单独决定

- **WHEN** 审查某 core-owned 领域能力是否对插件可用
- **THEN** 判定依据 SHALL 为安装/启用状态、依赖、hostServices（或等价）声明与授权及方法级校验，MUST NOT 仅因 `type=dynamic` 永久拒绝

#### Scenario: 文档不得将动态插件永久标为不可信

- **WHEN** 更新插件公开契约或宿主 README 中关于动态插件信任的表述
- **THEN** 表述 MUST 与「安装授权后同权同信」一致，MUST NOT 声称动态插件因 WASM 形态而永久半可信
