## ADDED Requirements

### Requirement: 插件依赖阻断必须区分安装轴与运行轴

系统 SHALL 将插件硬依赖生命周期阻断拆分为安装轴与运行轴，并按操作选择对应轴：

- **安装轴**：安装、卸载，以及升级时的版本契约与反向版本保护
- **运行轴**：启用、禁用

安装轴正向检查 MUST 以依赖插件是否已安装及版本是否满足为准，不得要求依赖插件已启用。运行轴正向检查 MUST 要求依赖插件已安装、已启用且版本满足。卸载反向检查 MUST 以已安装下游硬依赖为准。禁用反向检查 MUST 仅以已启用下游硬依赖为准；已安装但已禁用的下游 MUST NOT 阻断禁用。系统 MUST NOT 因依赖不满足而自动安装、卸载、启用或禁用相关插件。

#### Scenario: 安装不要求依赖已启用

- **WHEN** 插件 `consumer` 在 `dependencies.plugins` 中硬依赖 `core`
- **AND** `core` 已安装但未启用
- **AND** 管理员请求安装 `consumer`
- **THEN** 安装依赖检查 MUST 将 `core` 视为结构满足（版本满足时）
- **AND** 不得因 `core` 未启用阻断安装

#### Scenario: 启用要求依赖已启用

- **WHEN** 插件 `consumer` 在 `dependencies.plugins` 中硬依赖 `core`
- **AND** `core` 已安装但未启用
- **AND** 管理员请求启用 `consumer`
- **THEN** 启用请求失败
- **AND** 依赖检查结果包含 `core` 未启用的阻断信息
- **AND** 系统不得执行启用后的路由发布、cron 注册或运行时状态切换

#### Scenario: 下游全部禁用后允许禁用依赖插件

- **WHEN** 已安装插件 `oidc-a`、`oidc-b` 均在 `dependencies.plugins` 中硬依赖 `extlogin-core`
- **AND** `oidc-a` 与 `oidc-b` 当前均为已禁用
- **AND** 管理员请求禁用 `extlogin-core`
- **THEN** 禁用请求 MUST 成功（在无其他已启用下游硬依赖时）
- **AND** `oidc-a`、`oidc-b` 可继续保持已安装且已禁用

#### Scenario: 存在已启用下游时阻断禁用

- **WHEN** 已安装插件 `consumer` 在 `dependencies.plugins` 中硬依赖 `core` 且当前已启用
- **AND** 管理员请求禁用 `core`
- **THEN** 禁用请求失败
- **AND** 错误指出存在已启用下游插件 `consumer`

#### Scenario: 下游仅禁用仍阻断卸载

- **WHEN** 已安装插件 `consumer` 在 `dependencies.plugins` 中硬依赖 `core` 且当前已禁用
- **AND** 管理员请求卸载 `core`
- **THEN** 卸载请求失败
- **AND** 错误指出存在已安装下游插件 `consumer`
- **AND** 系统不得执行 `core` 的卸载 SQL、菜单清理或状态写入

#### Scenario: 不自动级联启停装卸

- **WHEN** 任何安装、卸载、启用或禁用请求因依赖轴检查失败
- **THEN** 系统 MUST 仅返回阻断原因
- **AND** 不得自动安装、卸载、启用或禁用依赖链上的其他插件

## MODIFIED Requirements

### Requirement: 硬 Provider 依赖必须使用既有插件依赖声明

当消费方插件必须保证某个 provider 插件存在、已安装、版本满足或生命周期顺序受保护时，系统 SHALL 要求消费方使用既有`dependencies.plugins`声明具体 provider 插件依赖和版本范围。插件安装、启用、卸载、升级和发布切换路径 MUST 继续按既有插件依赖语义保护这些硬依赖，并按安装轴与运行轴分别评估：安装与升级正向检查以已安装和版本满足为准；启用正向检查额外要求依赖插件已启用；卸载反向检查保护已安装下游；禁用反向检查仅保护已启用下游。不得引入第二套 capability 依赖阻断模型。

#### Scenario: 缺失硬 Provider 插件阻断启用

- **WHEN** 插件`consumer`在`dependencies.plugins`中硬依赖`linapro-tenant-core`
- **AND** `linapro-tenant-core`未安装、未启用或版本不满足
- **THEN** 启用`consumer`失败
- **AND** 系统不得执行该插件启用后的路由发布、cron 注册或运行时状态切换

#### Scenario: Provider 升级受下游插件依赖保护

- **WHEN** 已安装插件`consumer`在`dependencies.plugins`中硬依赖`linapro-org-core`版本范围`>=1.0.0 <2.0.0`
- **AND** 管理员尝试将`linapro-org-core`升级为不满足该范围的版本
- **THEN** 升级请求失败
- **AND** 错误包含下游插件 ID、provider 插件 ID 和版本要求

#### Scenario: 禁用已启用下游的 Provider 时保护硬依赖

- **WHEN** 插件`consumer`已启用且通过`dependencies.plugins`硬依赖唯一 tenant provider 插件
- **AND** 管理员尝试禁用唯一 tenant provider 插件
- **THEN** 禁用请求失败
- **AND** 错误包含依赖该 provider 插件的已启用下游插件列表

#### Scenario: 下游已全部禁用时允许禁用 Provider

- **WHEN** 插件`consumer`已安装且通过`dependencies.plugins`硬依赖唯一 tenant provider 插件
- **AND** `consumer`当前已禁用
- **AND** 管理员尝试禁用唯一 tenant provider 插件
- **THEN** 禁用请求 MUST 成功（在无其他已启用下游硬依赖时）

### Requirement: owner 插件生命周期必须保护下游能力消费方

系统 SHALL 在 owner 插件禁用、卸载、升级和版本切换前检查下游插件的`dependencies.plugins`和 owner-aware`hostServices`声明，并按操作轴评估：

- 禁用：仅当存在**已启用**下游硬依赖，且操作会导致 owner 未启用或缺失时，MUST 阻断
- 卸载：当存在**已安装**下游硬依赖时 MUST 阻断
- 升级/版本切换：当候选版本不满足**已安装**下游声明的版本范围时 MUST 阻断

反向阻断结果 MUST 包含下游插件 ID、owner 插件 ID、要求版本、候选版本（若适用）和触发的 owner 能力方法摘要。

#### Scenario: 禁用被已启用动态插件依赖的 owner

- **WHEN** 动态插件依赖`linapro-ai-core`并已授权调用`owner: linapro-ai-core service: ai`
- **AND** 该动态插件当前已启用
- **AND** 管理员尝试禁用`linapro-ai-core`
- **THEN** 禁用请求 MUST 被依赖检查阻断
- **AND** 阻断结果 MUST 列出依赖该 owner 的动态插件和对应 owner 能力声明

#### Scenario: 下游动态插件已禁用时允许禁用 owner

- **WHEN** 动态插件依赖`linapro-ai-core`并已授权调用`owner: linapro-ai-core service: ai`
- **AND** 该动态插件当前已安装但已禁用
- **AND** 管理员尝试禁用`linapro-ai-core`
- **THEN** 禁用请求 MUST 成功（在无其他已启用下游硬依赖时）

#### Scenario: 升级 owner 后版本不满足

- **WHEN** 已安装下游插件声明`linapro-ai-core`版本范围`>=0.1.0 <0.2.0`
- **AND** 管理员尝试升级`linapro-ai-core`到`v0.2.0`
- **THEN** 升级预检查 MUST 阻断或要求下游插件先调整依赖范围
- **AND** 系统不得执行会破坏下游插件运行期 owner 能力授权的发布切换
