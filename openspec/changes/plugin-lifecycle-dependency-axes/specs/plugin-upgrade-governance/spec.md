## MODIFIED Requirements

### Requirement: 插件升级必须校验下游 Provider 插件依赖

插件升级 SHALL 校验升级后的 provider 插件状态不会破坏其他已安装插件通过既有`dependencies.plugins`声明的硬依赖。升级与发布切换 MUST 按安装轴评估：只要存在已安装下游硬依赖且候选版本不满足其版本范围，系统 MUST 拒绝该操作或进入规范明确的阻断状态，无论下游当前是否启用。

禁用 provider 插件时 MUST 按运行轴评估：仅当存在已启用下游硬依赖时阻断禁用；已安装但已禁用的下游 MUST NOT 单独阻断禁用。pluginservice capability 的可选消费仍通过运行时可用性降级表达，不引入独立 capability 依赖阻断模型。

#### Scenario: Provider 升级后不满足下游插件依赖版本

- **WHEN** 已安装插件`consumer`在`dependencies.plugins`中硬依赖`linapro-org-core`版本范围`>=1.0.0 <2.0.0`
- **AND** 管理员尝试将`linapro-org-core`升级为不满足该范围的新版本
- **THEN** 升级请求失败或要求先处理下游依赖
- **AND** 错误包含下游插件 ID、provider 插件 ID 和所需版本范围
- **AND** 即使`consumer`当前已禁用，升级仍 MUST 被阻断

#### Scenario: 禁用唯一 Provider 时保护已启用下游硬依赖

- **WHEN** 插件`consumer`已启用且通过`dependencies.plugins`硬依赖唯一 tenant provider 插件
- **AND** 管理员尝试禁用唯一 tenant provider 插件
- **THEN** 禁用请求失败
- **AND** 错误包含依赖该 provider 插件的已启用下游插件列表

#### Scenario: 下游仅禁用时允许禁用唯一 Provider

- **WHEN** 插件`consumer`已安装但已禁用，且通过`dependencies.plugins`硬依赖唯一 tenant provider 插件
- **AND** 管理员尝试禁用唯一 tenant provider 插件
- **THEN** 禁用请求 MUST 成功（在无其他已启用下游硬依赖时）
