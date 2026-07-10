## ADDED Requirements

### Requirement: owner 能力消费必须声明 owner 插件硬依赖

系统 SHALL 要求插件消费 plugin-owned 领域能力时，通过既有`dependencies.plugins`声明 owner 插件硬依赖和版本范围。源码插件只要生产代码 import 其他插件`backend/cap/...`，就 MUST 在`plugin.yaml dependencies.plugins`中声明该 owner 插件。动态插件只要在`hostServices`中声明带`owner`字段的 owner 能力，就 MUST 在`dependencies.plugins`中声明同一个 owner 插件。系统 MUST NOT 引入`dependencies.capabilities`、软依赖或自动安装策略表达 owner 能力依赖。

#### Scenario: 源码插件 import owner backend/cap

- **WHEN** 源码插件生产代码 import `lina-plugin-linapro-ai-core/backend/cap/aicap/aitext`
- **THEN** 该插件`plugin.yaml` MUST 在`dependencies.plugins`中声明`id: linapro-ai-core`
- **AND** 治理扫描 MUST 在缺少依赖声明时失败

#### Scenario: 动态插件声明 owner host service

- **WHEN** 动态插件`hostServices`声明`owner: linapro-ai-core`
- **THEN** 该插件`dependencies.plugins` MUST 声明`linapro-ai-core`和兼容版本范围
- **AND** 清单校验、artifact 校验、安装或启用路径 MUST 在依赖缺失时阻断

#### Scenario: 不引入 capability 依赖块

- **WHEN** 插件需要硬依赖`AI`或其他 owner 能力
- **THEN** 插件只能使用`dependencies.plugins`
- **AND** 清单模型不得新增或接受`dependencies.capabilities`、顶层`capabilities`或自动安装字段

### Requirement: owner 插件生命周期必须保护下游能力消费方

系统 SHALL 在 owner 插件禁用、卸载、升级和版本切换前检查已安装插件的`dependencies.plugins`和 owner-aware`hostServices`声明。若下游插件硬依赖该 owner 插件，且目标操作会导致 owner 缺失、未启用或版本不满足，下游依赖 MUST 阻断该操作。反向阻断结果 MUST 包含下游插件 ID、owner 插件 ID、要求版本、候选版本和触发的 owner 能力方法摘要。

#### Scenario: 禁用被动态插件依赖的 owner

- **WHEN** 动态插件依赖`linapro-ai-core`并已授权调用`owner: linapro-ai-core service: ai`
- **AND** 管理员尝试禁用`linapro-ai-core`
- **THEN** 禁用请求 MUST 被依赖检查阻断
- **AND** 阻断结果 MUST 列出依赖该 owner 的动态插件和对应 owner 能力声明

#### Scenario: 升级 owner 后版本不满足

- **WHEN** 下游插件声明`linapro-ai-core`版本范围`>=0.1.0 <0.2.0`
- **AND** 管理员尝试升级`linapro-ai-core`到`v0.2.0`
- **THEN** 升级预检查 MUST 阻断或要求下游插件先调整依赖范围
- **AND** 系统不得执行会破坏下游插件运行期 owner 能力授权的发布切换

### Requirement: owner 能力依赖检查不得产生 N+1 查询

系统 SHALL 在插件列表、详情、安装、启用、升级和卸载工作流中集合化装配 owner 插件依赖与 owner-aware host service 诊断。首屏插件列表不得为了每个插件逐项执行完整依赖检查；详情和治理动作可以按目标插件批量读取依赖、反向依赖、host service 声明和 owner descriptor 状态。

#### Scenario: 插件详情展示 owner 能力依赖

- **WHEN** 管理员打开声明 owner 能力的动态插件详情
- **THEN** 后端 MUST 通过集合化查询或缓存快照返回依赖插件状态、版本匹配结果和 owner host service 授权摘要
- **AND** 不得为每个 method 单独查询 owner 插件详情或 descriptor

#### Scenario: 首屏列表不做完整依赖展开

- **WHEN** 管理员请求插件管理首屏列表
- **THEN** 列表项不得包含完整 owner 能力依赖检查结果
- **AND** 后端不得为了首屏列表对每个插件执行完整反向依赖遍历
