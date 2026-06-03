## ADDED Requirements

### Requirement: 官方插件必须提供智能中心菜单和页面

系统 SHALL 新增官方源码插件 `linapro-ai-core`，并由该插件贡献“智能中心”菜单。菜单 MUST 包含“供应商”“档位管理”“调用日志”三个页面，且页面、路由、权限和静态资源 MUST 归属于插件目录。

#### Scenario: 插件启用后显示智能中心

- **WHEN** `linapro-ai-core` 插件已安装、启用且当前用户拥有对应菜单权限
- **THEN** 前端 MUST 显示“智能中心”菜单
- **AND** 菜单下 MUST 显示“供应商”“档位管理”“调用日志”
- **AND** “智能中心”顶级菜单 MUST 排在“扩展中心”之前

#### Scenario: 插件禁用或无权限时隐藏入口

- **WHEN** `linapro-ai-core` 插件被禁用、卸载或当前用户缺少菜单权限
- **THEN** 前端 MUST 完全隐藏“智能中心”及其子菜单入口
- **AND** 页面 MUST NOT 以置灰入口或空白页方式暴露不可用功能

#### Scenario: 插件资源不回流宿主

- **WHEN** 实现智能中心页面、API、SQL 或多语言资源
- **THEN** 这些资源 MUST 放在 `apps/lina-plugins/linapro-ai-core/` 的统一插件目录结构中
- **AND** 修改插件目录前 MUST 检查并遵守该插件根目录 `AGENTS.md`

### Requirement: 供应商页面必须维护供应商和模型

系统 SHALL 在“供应商”页面和插件 API 中提供供应商及其模型的增删改查能力。供应商 MUST 支持启停、协议地址、密钥引用和备注；模型 MUST 关联供应商，并声明能力类型、协议、名称、启停、token 上限和 thinking 支持范围。

#### Scenario: 查询供应商列表

- **WHEN** 用户打开“供应商”页面
- **THEN** 系统 MUST 分页返回供应商列表
- **AND** 每行 MUST 包含当前页面所需的最小供应商投影、密钥脱敏摘要、端点投影和模型摘要列表
- **AND** 后端 MUST 使用集合化查询或聚合查询装配模型摘要，MUST NOT 诱导前端逐供应商补查详情
- **AND** 页面 MUST 将 OpenAI-compatible 和 Anthropic-compatible 地址合并到“端点”列分行展示，并用标签标识端点类型
- **AND** 多个端点同时存在时，页面 MUST 完整展示每个端点地址，协议标签 MUST NOT 遮挡或挤压地址内容
- **AND** OpenAI-compatible 和 Anthropic-compatible 协议标签 MUST 使用一致宽度并右对齐，使端点 URL 保持左对齐
- **AND** 页面 MUST 在端点列右侧展示“密钥”列，密钥内容必须脱敏展示
- **AND** 密钥脱敏 MUST 基于原有存储值保留可识别前缀和末尾 2 位，中间使用星号隐藏，例如 `sk-1234567890` 展示为 `sk-**********90`
- **AND** 页面 MUST 在更新时间左侧展示“模型”列，MUST NOT 展示“模型数”或“启用模型数”列

#### Scenario: 创建或更新供应商

- **WHEN** 用户创建或更新供应商
- **THEN** 系统 MUST 校验供应商名称、协议地址和启用状态
- **AND** 系统 MUST 只保存密钥引用、加密密文或脱敏摘要
- **AND** API 响应 MUST NOT 返回 API key 明文
- **AND** 页面表单 MUST 使用“名称”作为供应商名称字段的用户可见标签
- **AND** 页面表单 MUST 使用“API 密钥”作为密钥字段名称
- **AND** 修改供应商时，页面 MUST NOT 将脱敏密钥回填到密钥输入框
- **AND** 修改供应商时，密钥输入框 MUST 为空，并通过 placeholder 提示“留空则保持原密钥”
- **AND** 页面表单字段 MUST 单独一行展示，不得把状态字段或其他输入域并排放在同一行
- **AND** 供应商新增和修改表单 MUST NOT 嵌入模型维护内容

#### Scenario: 维护供应商模型

- **WHEN** 用户在供应商下新增、编辑、删除或启停模型
- **THEN** 系统 MUST 校验模型属于目标供应商
- **AND** 模型 MUST 声明 `capabilityType`，首期可用值为 `text`
- **AND** 模型 MUST 声明 `capabilityMethod`，首期 `text` 可用值为 `generate`
- **AND** 文本模型 MUST 声明 `protocol`、`supportsThinking` 和 `supportedEfforts`
- **AND** 供应商列表工具栏 MUST 同时提供“新增供应商”和“新增模型”入口，且两个按钮 MUST 使用可区分的样式
- **AND** 供应商列表模型列 MUST 使用常规字重展示模型名，不得加粗模型名
- **AND** 模型删除入口 MUST 使用图标按钮展示，不得使用文字 `X` 作为删除按钮

#### Scenario: 同步模型列表

- **WHEN** 用户触发供应商模型同步
- **THEN** 系统 MUST 调用供应商协议对应的模型列表接口并写入公开模型元数据
- **AND** 页面 MUST 在供应商列表操作列中提供 OpenAI-compatible 和 Anthropic-compatible 的模型同步入口
- **AND** 同步入口 MUST 直接展示在操作列中，不得折叠进“更多”菜单
- **AND** OpenAI-compatible 和 Anthropic-compatible 同步入口同时存在时 MUST 分两行展示
- **AND** 同步失败 MUST 保留既有手工模型和被引用模型
- **AND** 系统 MUST NOT 自动推断未由供应商或用户明确声明的 `thinkingEffort` 支持范围

### Requirement: 供应商和模型删除必须保护档位引用

系统 SHALL 在删除供应商或模型前检查智能中心档位绑定引用，防止 `framework.ai.text.v1` 出现悬空配置。

#### Scenario: 阻止删除被档位使用的供应商

- **WHEN** 用户删除某个供应商
- **AND** 该供应商下任一模型被 `basic`、`standard` 或 `advanced` 档位绑定引用
- **THEN** 系统 MUST 拒绝删除
- **AND** 错误 MUST 说明该供应商正在被 AI 能力档位使用

#### Scenario: 阻止删除被档位使用的模型

- **WHEN** 用户删除某个模型
- **AND** 该模型被任一 AI 能力档位绑定引用
- **THEN** 系统 MUST 拒绝删除
- **AND** 错误 MUST 说明该模型正在被 AI 能力档位使用

#### Scenario: 删除未引用模型

- **WHEN** 用户删除未被任何档位绑定引用的模型
- **THEN** 系统 MUST 允许删除或软删除该模型
- **AND** 系统 MUST NOT 修改其他供应商、模型或档位配置

### Requirement: 档位管理必须提供三档文本 AI 能力配置

系统 SHALL 在“档位管理”页面和插件 API 中固定提供 `basic`、`standard`、`advanced` 三个文本能力档位。每个档位 MUST 能绑定一个主供应商模型，配置启用状态、默认 `thinkingEffort` 和测试入口。

#### Scenario: 档位身份包含能力方法

- **WHEN** 系统存储或查询能力档位
- **THEN** 档位记录 MUST 使用 `capabilityType + capabilityMethod + tierCode` 作为唯一身份
- **AND** 首期文本生成档位 MUST 使用 `capabilityType=text` 与 `capabilityMethod=generate`
- **AND** 同一能力方法范围下每个 `basic`、`standard`、`advanced` 档位 MUST 只能存在一条记录
- **AND** `capabilityType` 与 `capabilityMethod` MUST NOT 作为档位编辑表单中的可变字段

#### Scenario: 查询档位列表

- **WHEN** 前端调用 `GET /api/ai/tiers`
- **THEN** 系统 MUST 按请求中的 `capabilityType` 与 `capabilityMethod` 返回该能力方法下的 `basic`、`standard`、`advanced` 三个档位
- **AND** 未传 `capabilityMethod` 时首期 MUST 默认使用 `generate`
- **AND** 每个档位 MUST 包含能力类型、能力方法、代码、显示名称、说明、启用状态、默认 `thinkingEffort`、主绑定投影、最近测试摘要和更新时间
- **AND** 所有公开时间点字段 MUST 使用 Unix timestamp in milliseconds

#### Scenario: 更新档位主绑定

- **WHEN** 用户调用 `PUT /api/ai/tiers/{code}` 保存档位
- **THEN** 系统 MUST 校验档位代码属于 `basic`、`standard` 或 `advanced`
- **AND** 系统 MUST 使用请求中的 `capabilityType + capabilityMethod + code` 定位档位，首期默认定位 `text.generate`
- **AND** 系统 MUST 校验供应商和模型真实存在、已启用且模型属于供应商
- **AND** 系统 MUST 校验模型的 `capabilityType` 为 `text`
- **AND** 系统 MUST 校验模型的 `capabilityMethod` 为 `generate`
- **AND** 系统 MUST 将该模型设置为该档位的主绑定

#### Scenario: 默认 thinkingEffort 校验

- **WHEN** 用户为档位配置默认 `thinkingEffort`
- **THEN** 系统 MUST 校验默认值为空或属于 `low`、`medium`、`high`、`xhigh`、`max`
- **AND** 若绑定模型不支持该默认值，系统 MUST 拒绝保存并提示模型不支持该 effort

#### Scenario: 禁用档位

- **WHEN** 用户关闭某个档位的启用状态并保存
- **THEN** 系统 MUST 保留该档位已有绑定
- **AND** 后续文本能力调用 MUST 将该档位视为不可用

### Requirement: 档位管理页面必须方便配置和诊断

系统 SHALL 在“档位管理”页面以三档配置为核心展示对象，并提供可发现的保存、测试、校验和状态反馈。页面 MUST 复用现有前端表格、表单、弹窗、抽屉和操作列模式。

#### Scenario: 展示三档配置

- **WHEN** 用户进入“档位管理”页面
- **THEN** 页面 MUST 按 `basic`、`standard`、`advanced` 稳定顺序展示三档
- **AND** 每档 MUST 展示启用状态、供应商、模型、协议、默认 `thinkingEffort`、模型 thinking 支持范围、最近测试状态和保存入口

#### Scenario: 模型选择按供应商过滤

- **WHEN** 用户在某个档位选择供应商
- **THEN** 模型选择项 MUST 只展示该供应商下已启用且匹配当前 `capabilityType + capabilityMethod` 的模型
- **AND** 模型选项 MUST 展示协议和 thinking 支持信息

#### Scenario: 不支持 effort 时提示

- **WHEN** 用户选择的默认 `thinkingEffort` 不被当前模型支持
- **THEN** 页面 MUST 在保存前给出明确校验提示
- **AND** 页面 MUST NOT 发送会被静默降级的保存请求

### Requirement: 档位必须支持轻量测试

系统 SHALL 允许用户对任一文本能力档位执行轻量连通性测试。测试 MUST 使用已保存绑定或请求中的草稿绑定，MUST NOT 在测试草稿绑定时持久化档位配置。

#### Scenario: 测试已保存档位

- **WHEN** 用户触发某个已启用且已配置主绑定的档位测试
- **THEN** 系统 MUST 使用该档位绑定的供应商模型执行轻量文本生成请求
- **AND** 响应 MUST 返回测试状态、耗时、实际供应商、实际模型、协议、实际 `thinkingEffort` 和测试时间

#### Scenario: 测试草稿绑定

- **WHEN** 用户选择供应商、模型和默认 `thinkingEffort` 但尚未保存并触发测试
- **THEN** 系统 MUST 使用请求中的草稿绑定执行测试
- **AND** 系统 MUST 校验草稿绑定真实存在且模型支持指定 effort
- **AND** 系统 MUST NOT 创建、更新或删除该档位主绑定

#### Scenario: 测试失败保留配置

- **WHEN** 档位测试因密钥、网络、协议、模型、thinking effort 或供应商错误失败
- **THEN** 系统 MUST 返回脱敏且可理解的失败摘要
- **AND** 系统 MUST NOT 删除或修改该档位已有配置

### Requirement: 文本调用日志必须可监控且不泄露敏感内容

系统 SHALL 在“调用日志”页面和插件 API 中提供文本 `AI` 调用日志查询。日志 MUST 支持分页、筛选、用量和耗时分析，MUST NOT 保存完整输入、完整输出、隐藏思考内容或密钥。

#### Scenario: 记录成功调用

- **WHEN** 文本 `AI` 调用成功
- **THEN** 系统 MUST 记录 `requestId`、`capabilityType`、`capabilityMethod`、`purpose`、档位、调用来源、供应商模型投影、协议、`thinkingEffort`、状态、token 用量、耗时和创建时间
- **AND** 系统 MUST NOT 保存完整 `messages`、完整响应正文、隐藏思考内容或 API key

#### Scenario: 记录失败调用

- **WHEN** 文本 `AI` 调用失败
- **THEN** 系统 MUST 记录可解析的供应商模型投影、失败状态、耗时、错误码和脱敏错误摘要
- **AND** 错误摘要 MUST NOT 包含密钥、认证头、完整请求体或完整响应体

#### Scenario: 查询调用日志

- **WHEN** 用户调用 `GET /api/ai/invocations`
- **THEN** 系统 MUST 按创建时间倒序分页返回日志
- **AND** 查询 MUST 支持按 `capabilityType`、`capabilityMethod`、`purpose`、档位、状态、供应商、模型、来源插件和时间范围过滤
- **AND** 后端 MUST 在数据库侧完成过滤、排序和分页

### Requirement: 插件存储必须由 linapro-ai-core 自有并具备性能边界

系统 SHALL 由 `linapro-ai-core` 插件维护供应商、模型、档位、绑定和调用日志表。插件 SQL MUST 使用 PostgreSQL 源语法、幂等 DDL/Seed、软删除语义、合理索引和插件表名前缀。

#### Scenario: 安装插件创建表和固定档位

- **WHEN** 安装或初始化 `linapro-ai-core`
- **THEN** 插件 SQL MUST 创建 `plugin_linapro_ai_provider`、`plugin_linapro_ai_model`、`plugin_linapro_ai_tier`、`plugin_linapro_ai_tier_binding` 和 `plugin_linapro_ai_invocation`
- **AND** SQL MUST seed `text.generate` 能力方法的 `basic`、`standard`、`advanced` 三个固定档位
- **AND** 档位唯一约束 MUST 覆盖 `capability_type`、`capability_method` 和 `code`
- **AND** SQL MUST 满足重复执行结果一致

#### Scenario: 高频查询具备索引

- **WHEN** 系统查询供应商列表、模型列表、档位解析或调用日志
- **THEN** 表结构 MUST 为供应商模型关联、`capabilityType + capabilityMethod` 模型筛选、档位绑定解析、日志时间范围、状态、档位、供应商和模型过滤提供必要索引
- **AND** 设计 MUST 支持集合化读取和当前页批量装配，MUST NOT 依赖动态结果集逐行查询

#### Scenario: 不写入自增主键

- **WHEN** 插件 SQL 写入固定档位或初始化数据
- **THEN** SQL MUST 使用稳定业务键和唯一约束实现幂等
- **AND** SQL MUST NOT 显式写入自增 `id`

### Requirement: 档位解析缓存必须保持一致性

系统 SHALL 在 `linapro-ai-core` provider 内维护文本档位解析缓存，用于将 `capabilityType + capabilityMethod + tier` 解析为可调用供应商模型绑定。缓存权威源 MUST 是插件数据库，失效 MUST 与配置写入成功耦合。

#### Scenario: 配置变更后失效缓存

- **WHEN** 用户创建、更新、启停或删除供应商、模型、档位或绑定
- **THEN** 系统 MUST 在数据库写入成功后失效相关档位解析缓存
- **AND** 后续文本调用 MUST 使用刷新后的绑定或在缓存 miss 时从数据库重建

#### Scenario: 集群模式同步失效

- **WHEN** `cluster.enabled=true` 且某节点修改档位相关配置
- **THEN** 系统 MUST 通过宿主插件运行时修订、共享修订号、事件广播或等价机制让其他节点观察到失效
- **AND** 系统 MUST NOT 在集群模式下只刷新当前节点本地内存

#### Scenario: 缓存故障降级

- **WHEN** 档位解析缓存不可用或条目缺失
- **THEN** provider MUST 直接读取插件数据库重建解析结果
- **AND** 当数据库不可用时 MUST 返回结构化不可用错误并记录脱敏失败摘要

### Requirement: 智能中心管理面必须受平台权限治理

系统 SHALL 将智能中心管理 API 作为平台配置控制面。供应商、模型、档位和日志管理 MUST 要求平台上下文和对应权限；动态插件文本调用 MUST 使用 host service 授权，不得绕过管理面权限。

#### Scenario: 平台管理员访问管理 API

- **WHEN** 平台上下文用户拥有对应权限并访问智能中心管理 API
- **THEN** 系统 MUST 允许执行授权范围内的读取、创建、更新、删除或测试动作
- **AND** 操作 MUST 进入统一权限和审计边界

#### Scenario: 非平台上下文被拒绝

- **WHEN** 租户上下文或代管租户上下文请求创建、更新、删除供应商、模型或档位
- **THEN** 系统 MUST 在执行数据库写入或供应商调用前拒绝请求
- **AND** 错误 MUST 不泄露目标资源是否存在于平台配置中

#### Scenario: 插件调用不获得管理权限

- **WHEN** 动态插件获得 `ai.text.generate` host service 授权
- **THEN** 该授权 MUST 只允许按授权 purpose 调用文本能力
- **AND** 插件 MUST NOT 因该授权访问供应商、档位管理或调用日志管理 API

### Requirement: 插件 i18n 和 API 文档必须按插件配置治理

系统 SHALL 对智能中心菜单、页面文案、按钮、表单、表格、错误消息和 API 文档执行插件维度 `i18n` 影响评估。实现期 MUST 先读取 `linapro-ai-core` 的 `plugin.yaml`，再按插件 `i18n.enabled` 决定资源维护方式。

#### Scenario: 插件启用 i18n

- **WHEN** `linapro-ai-core` 的 `plugin.yaml` 声明 `i18n.enabled: true`
- **THEN** 插件运行时文案和 API 文档翻译资源 MUST 维护在该插件自己的 `manifest/i18n/<locale>/` 和 `manifest/i18n/<locale>/apidoc/`
- **AND** API 文档源文本 MUST 使用清晰英文源内容并维护非英文翻译

#### Scenario: 插件未启用 i18n

- **WHEN** `linapro-ai-core` 未启用插件自身 `i18n`
- **THEN** 实现记录 MUST 明确说明单语言插件判断
- **AND** 系统 MUST NOT 把插件文案翻译键集中写入 `lina-core` 的语言资源中
