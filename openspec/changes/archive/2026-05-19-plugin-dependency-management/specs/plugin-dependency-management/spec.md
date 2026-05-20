## ADDED Requirements

### Requirement: 插件清单必须支持依赖声明

系统 SHALL 允许源码插件和动态插件在 `plugin.yaml` 或等价运行时 manifest 中通过 `dependencies` 声明 LinaPro 框架版本约束和插件依赖约束。未声明 `dependencies` 的插件 SHALL 保持合法，并按无依赖插件处理。

#### Scenario: 解析框架版本约束和插件依赖
- **WHEN** 插件清单包含 `dependencies.framework.version` 和 `dependencies.plugins`
- **THEN** 系统解析框架版本约束
- **AND** 系统解析每个插件依赖的 `id`、`version`、`required` 和 `install`
- **AND** 缺省 `required` 按 `true` 处理
- **AND** 缺省 `install` 按 `manual` 处理

#### Scenario: 未声明依赖的插件保持兼容
- **WHEN** 插件清单未包含 `dependencies`
- **THEN** 系统将该插件视为无依赖插件
- **AND** 清单校验不得因为缺少 `dependencies` 失败

#### Scenario: 动态插件产物携带依赖声明
- **WHEN** 动态插件 WASM 产物的 manifest 自定义段包含 `dependencies`
- **THEN** 系统按源码插件相同语义解析依赖声明
- **AND** 动态插件安装、启用和升级路径使用解析后的依赖约束

### Requirement: 插件依赖声明必须被结构化校验

系统 SHALL 在清单校验阶段验证依赖声明结构。框架版本约束和插件版本约束必须使用受支持的语义化版本范围；插件依赖 ID 必须符合插件 ID 命名规则；插件不得依赖自身；同一清单不得重复声明同一插件依赖。

#### Scenario: 拒绝无效依赖字段
- **WHEN** 插件清单声明空依赖 ID、无效版本范围、未知安装策略或重复依赖
- **THEN** 清单校验失败
- **AND** 错误包含插件 ID、依赖字段路径和无效值

#### Scenario: 拒绝自依赖
- **WHEN** 插件 `content-notice` 在 `dependencies.plugins` 中声明依赖 `content-notice`
- **THEN** 清单校验失败
- **AND** 错误说明插件不得依赖自身

### Requirement: 安装前必须执行依赖检查

系统 SHALL 在执行插件安装生命周期副作用前完成依赖检查。依赖检查必须校验当前 LinaPro 框架版本、依赖插件是否可发现、依赖版本是否满足、硬依赖是否已安装或可自动安装，以及依赖图是否存在循环。

#### Scenario: 框架版本不满足时阻断安装
- **WHEN** 插件声明 `dependencies.framework.version: ">=0.7.0"`
- **AND** 当前 LinaPro 框架版本为 `v0.6.0`
- **THEN** 插件安装请求失败
- **AND** 系统不得执行该插件的 SQL、菜单同步、发布切换或状态写入
- **AND** 错误包含当前框架版本和要求的版本范围

#### Scenario: 缺失硬依赖时阻断安装
- **WHEN** 插件声明硬依赖 `multi-tenant`
- **AND** 插件 catalog 未发现 `multi-tenant`
- **THEN** 插件安装请求失败
- **AND** 错误包含缺失依赖插件 ID 和目标插件 ID

#### Scenario: 依赖版本不满足时阻断安装
- **WHEN** 插件声明依赖 `org-center` 版本范围 `>=0.2.0`
- **AND** 当前可用或已安装的 `org-center` 版本为 `v0.1.0`
- **THEN** 插件安装请求失败
- **AND** 错误包含当前依赖版本和要求的版本范围

#### Scenario: 循环依赖时阻断安装
- **WHEN** 插件依赖图中存在 `a -> b -> c -> a`
- **THEN** 任一参与该循环的插件安装请求失败
- **AND** 错误包含循环依赖链

### Requirement: 自动依赖安装必须按确定性拓扑顺序执行

当插件硬依赖声明 `install: auto` 且依赖插件已被发现、版本满足但尚未安装时，系统 SHALL 在安装目标插件前自动安装该依赖。自动安装计划必须按依赖拓扑排序，依赖插件先于被依赖插件安装；同一拓扑层按插件 ID 稳定排序。

#### Scenario: 自动安装未安装硬依赖
- **WHEN** 管理员安装插件 `content-notice`
- **AND** `content-notice` 声明 `multi-tenant` 为 `required: true` 且 `install: auto`
- **AND** `multi-tenant` 已发现、版本满足且尚未安装
- **THEN** 系统先安装 `multi-tenant`
- **AND** 系统再安装 `content-notice`
- **AND** 安装结果记录自动安装的依赖插件列表

#### Scenario: 自动安装不启用依赖插件
- **WHEN** 系统自动安装依赖插件 `multi-tenant`
- **THEN** 系统只保证依赖插件达到已安装状态
- **AND** 系统不得因为依赖自动安装而默认启用该依赖插件

#### Scenario: 自动安装中途失败时停止后续安装
- **WHEN** 自动安装计划包含 `a`、`b`、`target`
- **AND** 插件 `b` 安装失败
- **THEN** 系统停止安装 `target`
- **AND** 错误包含已安装依赖列表、失败插件 ID 和失败原因

### Requirement: 手动依赖必须阻断目标插件安装并提示操作

当硬依赖声明 `install: manual` 且依赖插件未安装时，系统 SHALL 阻断目标插件安装，并返回需要先手动安装的依赖列表。系统不得在未确认的情况下自动安装 `manual` 依赖。

#### Scenario: 未安装手动硬依赖
- **WHEN** 插件 `x` 声明硬依赖 `org-center` 且 `install: manual`
- **AND** `org-center` 尚未安装
- **THEN** 安装 `x` 失败
- **AND** 错误提示先安装 `org-center`

### Requirement: 软依赖不得阻断插件生命周期

当依赖声明 `required: false` 时，系统 SHALL 将其视为软依赖。软依赖缺失、未安装或版本不满足不得阻断安装、启用或升级，但依赖检查结果必须标记软依赖未满足状态，供 API 和 UI 展示。

#### Scenario: 软依赖缺失时继续安装
- **WHEN** 插件 `x` 声明软依赖 `optional-analytics`
- **AND** `optional-analytics` 未被发现
- **THEN** 插件 `x` 安装继续执行
- **AND** 依赖检查结果标记 `optional-analytics` 为未满足软依赖

### Requirement: 卸载必须保护已安装插件的硬依赖

系统 SHALL 在卸载插件前检查已安装插件的硬依赖声明。如果存在其他已安装插件硬依赖当前插件，卸载请求必须失败，并返回依赖当前插件的下游插件列表。

#### Scenario: 被已安装插件依赖时拒绝卸载
- **WHEN** 插件 `content-notice` 已安装且硬依赖 `multi-tenant`
- **AND** 管理员请求卸载 `multi-tenant`
- **THEN** 卸载请求失败
- **AND** 系统不得执行 `multi-tenant` 的卸载 SQL、菜单清理或状态写入
- **AND** 错误包含下游插件 `content-notice`

#### Scenario: 无下游硬依赖时允许卸载
- **WHEN** 没有已安装插件硬依赖目标插件
- **THEN** 系统允许继续执行既有卸载生命周期

### Requirement: 依赖检查结果必须通过 API 和 UI 可见

系统 SHALL 为插件管理提供依赖检查结果，包含框架版本检查、依赖插件状态、版本匹配结果、自动安装计划、手动处理项、软依赖提示、循环依赖和卸载阻断项。前端 SHALL 使用服务端结果展示安装确认和阻断原因，不得在前端自行决定依赖图语义。

#### Scenario: 安装前展示自动安装计划
- **WHEN** 管理员在插件管理页面点击安装一个存在自动依赖的插件
- **THEN** 前端展示后端返回的自动安装依赖列表
- **AND** 管理员确认后才提交安装请求

#### Scenario: 展示阻断原因
- **WHEN** 后端依赖检查返回框架版本不满足或依赖版本不满足
- **THEN** 插件管理页面展示对应阻断原因
- **AND** 文案使用 i18n 资源而非硬编码文本

#### Scenario: 卸载确认展示下游依赖
- **WHEN** 管理员尝试卸载被其他插件硬依赖的插件
- **THEN** 插件管理页面展示下游插件列表
- **AND** 卸载操作被阻止

### Requirement: 依赖生命周期变化必须保持缓存一致性

系统 SHALL 在依赖自动安装、目标插件安装、卸载阻断解除后的卸载、源码插件升级和动态插件升级成功后，按受影响插件范围发布或刷新插件 runtime revision/event、enabled snapshot、frontend bundle、runtime i18n bundle 和 apidoc i18n 派生缓存。集群模式下不得只刷新当前节点内存状态。

#### Scenario: 集群模式下自动安装多个插件
- **WHEN** 集群模式下主节点自动安装依赖插件并安装目标插件
- **THEN** 主节点为每个受影响插件发布插件运行时修订或等价事件
- **AND** 非主节点观察到事件后刷新本地启用快照和派生缓存

#### Scenario: 只读依赖检查不触发缓存失效
- **WHEN** 管理员只执行安装前依赖检查
- **THEN** 系统不得写入插件 registry、release snapshot 或缓存修订号
- **AND** 系统不得清空所有语言和所有扇区的 i18n 缓存
