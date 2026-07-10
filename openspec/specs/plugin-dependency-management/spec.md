# plugin-dependency-management Specification

## Purpose
TBD - created by archiving change plugin-dependency-management. Update Purpose after archive.
## Requirements
### Requirement: 插件清单必须支持依赖声明

系统 SHALL 允许源码插件和动态插件在`plugin.yaml`或等价运行时 manifest 中通过`dependencies`声明 LinaPro 框架版本约束和插件依赖约束。未声明`dependencies`的插件 SHALL 保持合法，并按无依赖插件处理。插件依赖项 MUST 只包含`id`和可选`version`字段；声明在`dependencies.plugins`中的插件依赖一律视为硬依赖。系统 MUST NOT 支持`required`、`install`或等价软依赖、自动安装策略字段。

#### Scenario: 解析框架版本约束和插件依赖

- **WHEN** 插件清单包含`dependencies.framework.version`和`dependencies.plugins`
- **THEN** 系统解析框架版本约束
- **AND** 系统解析每个插件依赖的`id`和可选`version`
- **AND** 每个插件依赖均按硬依赖处理

#### Scenario: 未声明依赖的插件保持合法

- **WHEN** 插件清单未包含`dependencies`
- **THEN** 系统将该插件视为无依赖插件
- **AND** 清单校验不得因为缺少`dependencies`失败

#### Scenario: 动态插件产物携带依赖声明

- **WHEN** 动态插件 WASM 产物的 manifest 自定义段包含`dependencies`
- **THEN** 系统按源码插件相同语义解析依赖声明
- **AND** 动态插件安装、启用和升级路径使用解析后的依赖约束

#### Scenario: 拒绝软依赖和自动安装字段

- **WHEN** 插件清单在`dependencies.plugins`中声明`required`、`install`或等价字段
- **THEN** 清单校验失败
- **AND** 错误包含插件 ID、依赖字段路径和不支持的字段名

### Requirement: 插件依赖声明必须被结构化校验

系统 SHALL 在清单校验阶段验证依赖声明结构。框架版本约束和插件版本约束必须使用受支持的语义化版本范围；插件依赖 ID 必须符合插件 ID 命名规则；插件不得依赖自身；同一清单不得重复声明同一插件依赖；插件依赖项不得包含除`id`、`version`之外的策略字段。

#### Scenario: 拒绝无效依赖字段

- **WHEN** 插件清单声明空依赖 ID、无效版本范围、重复依赖或不支持的依赖策略字段
- **THEN** 清单校验失败
- **AND** 错误包含插件 ID、依赖字段路径和无效值

#### Scenario: 拒绝自依赖

- **WHEN** 插件`content-notice`在`dependencies.plugins`中声明依赖`content-notice`
- **THEN** 清单校验失败
- **AND** 错误说明插件不得依赖自身

### Requirement: 安装前必须执行依赖检查

系统 SHALL 在执行插件安装生命周期副作用前完成依赖检查。依赖检查必须校验当前 LinaPro 框架版本、依赖插件是否可发现、依赖版本是否满足、依赖插件是否已安装，以及依赖图是否存在循环。声明在`dependencies.plugins`中的依赖不满足时 MUST 阻断目标插件安装；系统 MUST NOT 根据插件清单自动安装依赖插件。

#### Scenario: 框架版本不满足时阻断安装

- **WHEN** 插件声明`dependencies.framework.version: ">=0.7.0"`
- **AND** 当前 LinaPro 框架版本为`v0.6.0`
- **THEN** 插件安装请求失败
- **AND** 系统不得执行该插件的 SQL、菜单同步、发布切换或状态写入
- **AND** 错误包含当前框架版本和要求的版本范围

#### Scenario: 缺失插件依赖时阻断安装

- **WHEN** 插件在`dependencies.plugins`中声明依赖`multi-tenant`
- **AND** 插件 catalog 未发现或未安装`multi-tenant`
- **THEN** 插件安装请求失败
- **AND** 错误包含缺失依赖插件 ID 和目标插件 ID
- **AND** 系统不得尝试自动安装`multi-tenant`

#### Scenario: 依赖版本不满足时阻断安装

- **WHEN** 插件声明依赖`org-center`版本范围`>=0.2.0`
- **AND** 当前已安装的`org-center`版本为`v0.1.0`
- **THEN** 插件安装请求失败
- **AND** 错误包含当前依赖版本和要求的版本范围

#### Scenario: 循环依赖时阻断安装

- **WHEN** 插件依赖图中存在`a -> b -> c -> a`
- **THEN** 任一参与该循环的插件安装请求失败
- **AND** 错误包含循环依赖链

### Requirement: 卸载必须保护已安装插件的硬依赖

系统 SHALL 在卸载插件前检查已安装插件的依赖声明。由于`dependencies.plugins`中的依赖均为硬依赖，如果存在其他已安装插件依赖当前插件，卸载请求必须失败，并返回依赖当前插件的下游插件列表。

#### Scenario: 被已安装插件依赖时拒绝卸载

- **WHEN** 插件`content-notice`已安装且在`dependencies.plugins`中依赖`multi-tenant`
- **AND** 管理员请求卸载`multi-tenant`
- **THEN** 卸载请求失败
- **AND** 系统不得执行`multi-tenant`的卸载 SQL、菜单清理或状态写入
- **AND** 错误包含下游插件`content-notice`

#### Scenario: 无下游依赖时允许卸载

- **WHEN** 没有已安装插件依赖目标插件
- **THEN** 系统允许继续执行既有卸载生命周期

### Requirement: 依赖检查结果必须通过 API 和 UI 可见

系统 SHALL 为插件管理的详情和治理动作提供依赖检查结果，包含框架版本检查、依赖插件状态、版本匹配结果、循环依赖和卸载阻断项。前端 SHALL 使用服务端结果展示阻断原因，不得在前端自行决定依赖图语义。依赖检查结果 MUST NOT 包含自动安装计划、自动安装结果、手动安装策略或软依赖提示。首屏插件列表 SHALL NOT 随列表项返回完整依赖检查结果；列表只负责展示摘要治理状态，依赖决策 MUST 在详情、安装、升级或卸载工作流中按需获取。

#### Scenario: 展示阻断原因
- **WHEN** 后端依赖检查返回框架版本不满足、依赖缺失或依赖版本不满足
- **THEN** 插件管理页面在详情或治理动作弹窗中展示对应阻断原因
- **AND** 文案使用 i18n 资源而非硬编码文本

#### Scenario: 不展示自动安装计划
- **WHEN** 管理员在插件管理页面点击安装一个依赖其他插件的插件
- **THEN** 后端依赖检查结果只返回已满足或阻断项
- **AND** 前端不得展示由插件清单驱动的自动安装计划

#### Scenario: 卸载确认展示下游依赖
- **WHEN** 管理员尝试卸载被其他插件依赖的插件
- **THEN** 插件管理页面在卸载确认工作流中展示下游插件列表
- **AND** 卸载操作被阻止

#### Scenario: 首屏列表不返回完整依赖检查结果
- **WHEN** 管理员请求插件管理首屏列表
- **THEN** `GET /plugins`列表项不得包含 `dependencyCheck`
- **AND** 后端不得为了首屏列表逐插件执行依赖检查
- **AND** 管理员打开详情、安装、升级或卸载工作流时，系统再通过服务端接口返回该工作流需要的依赖检查结果

### Requirement: 依赖生命周期变化必须保持缓存一致性

系统 SHALL 在目标插件安装、卸载阻断解除后的卸载、源码插件升级和动态插件升级成功后，按受影响插件范围发布或刷新插件 runtime revision/event、enabled snapshot、frontend bundle、runtime i18n bundle 和 apidoc i18n 派生缓存。集群模式下不得只刷新当前节点内存状态。只读依赖检查不得写入插件 registry、release snapshot 或缓存修订号。

#### Scenario: 集群模式下安装目标插件

- **WHEN** 集群模式下主节点完成目标插件安装
- **THEN** 主节点为受影响插件发布插件运行时修订或等价事件
- **AND** 非主节点观察到事件后刷新本地启用快照和派生缓存

#### Scenario: 只读依赖检查不触发缓存失效

- **WHEN** 管理员只执行安装前依赖检查
- **THEN** 系统不得写入插件 registry、release snapshot 或缓存修订号
- **AND** 系统不得清空所有语言和所有扇区的 i18n 缓存

### Requirement: Pluginservice Capability 消费不得新增依赖配置块

系统 SHALL 复用既有`dependencies`模型表达插件间依赖和 LinaPro 框架版本约束。Pluginservice capability 消费 MUST NOT 定义、读取或消费与`dependencies`并列的`capabilities`顶层配置块，也 MUST NOT 在`dependencies`下定义、读取或消费`capabilities`或等价 capability 依赖子字段。能力是否可用 SHALL 由对应能力组件的 provider 激活状态和消费 service 的`Available(ctx)`或等价状态表达。

#### Scenario: 能力依赖不进入插件清单模型

- **WHEN** 系统解析插件清单或动态插件构建器生成运行时产物
- **THEN** 清单模型不得包含顶层`capabilities`、`dependencies.capabilities`或等价字段
- **AND** 构建器不得把 capability 依赖字段写入运行时产物或运行时依赖快照

#### Scenario: 未声明能力依赖的插件仍可消费可用能力

- **WHEN** 插件未声明任何 capability 依赖字段
- **AND** 运行时存在 active provider 或 fallback
- **THEN** 插件可通过`pluginservice.Services.Org()`或等价注入的`orgcap.Service`使用该能力
- **AND** 依赖检查结果不得要求存在 capability 依赖声明

### Requirement: 硬 Provider 依赖必须使用既有插件依赖声明

当消费方插件必须保证某个 provider 插件存在、已安装、版本满足或生命周期顺序受保护时，系统 SHALL 要求消费方使用既有`dependencies.plugins`声明具体 provider 插件依赖和版本范围。插件安装、启用、卸载、升级和发布切换路径 MUST 继续按既有插件依赖语义保护这些硬依赖，不得引入第二套 capability 依赖阻断模型。

#### Scenario: 缺失硬 Provider 插件阻断启用

- **WHEN** 插件`consumer`在`dependencies.plugins`中硬依赖`linapro-tenant-core`
- **AND** `linapro-tenant-core`未安装、未启用或版本不满足
- **THEN** 启用`consumer`失败
- **AND** 系统不得执行该插件启用后的路由发布、cron 注册或运行时状态切换

#### Scenario: Provider 升级受下游插件依赖保护

- **WHEN** 已启用插件`consumer`在`dependencies.plugins`中硬依赖`linapro-org-core`版本范围`>=1.0.0 <2.0.0`
- **AND** 管理员尝试将`linapro-org-core`升级为不满足该范围的版本或禁用该插件
- **THEN** 升级或禁用请求失败
- **AND** 错误包含下游插件 ID、provider 插件 ID 和版本要求

### Requirement: 可选 Pluginservice Capability 必须通过运行时可用性降级

当插件只是可选使用 pluginservice capability 时，系统 SHALL 允许插件不声明 provider 插件硬依赖。消费方 MUST 通过`orgcap.Service`、`tenantcap.Service`等能力服务的`Available(ctx)`或等价能力状态判断能力是否可用，并在不可用时执行规范定义的隐藏、零值、跳过或降级行为；能力不可用不得由独立 manifest capability 依赖字段表达。

#### Scenario: 可选组织能力缺失时继续启用

- **WHEN** 插件只可选展示组织树、用户组织投影或组织数据范围增强信息
- **AND** 插件未在`dependencies.plugins`中硬依赖 provider 插件
- **AND** 当前环境没有可用`framework.org.v1`provider 或 fallback
- **THEN** 插件安装和启用继续执行
- **AND** 插件运行时通过`pluginservice.Services.Org().Available(ctx)`、注入的`orgcap.Service.Available(ctx)`或等价状态执行降级

#### Scenario: 动态插件消费能力

- **WHEN** 动态插件通过 guest SDK 消费`framework.tenant.v1`
- **THEN** 插件必须在`hostServices`中声明对应宿主服务、方法和资源边界
- **AND** 若该动态插件需要硬依赖具体 provider 插件，则继续使用`dependencies.plugins`声明该 provider 插件和版本约束
- **AND** 系统不得要求或解析`dependencies.capabilities`

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

