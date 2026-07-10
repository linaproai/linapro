# linapro-ai-core-plugin Specification

## Purpose
TBD - created by archiving change add-ai-text-capability-center. Update Purpose after archive.
## Requirements
### Requirement: 官方插件必须提供智能中心菜单和页面

系统 SHALL 新增官方源码插件 `linapro-ai-core`，并由该插件贡献“智能中心”菜单。菜单 MUST 包含“渠道”“档位管理”“调用日志”三个页面，且页面、路由、权限和静态资源 MUST 归属于插件目录。

#### Scenario: 插件启用后显示智能中心

- **WHEN** `linapro-ai-core` 插件已安装、启用且当前用户拥有对应菜单权限
- **THEN** 前端 MUST 显示“智能中心”菜单
- **AND** 菜单下 MUST 显示“渠道”“档位管理”“调用日志”
- **AND** “智能中心”顶级菜单 MUST 排在“扩展中心”之前

#### Scenario: 插件禁用或无权限时隐藏入口

- **WHEN** `linapro-ai-core` 插件被禁用、卸载或当前用户缺少菜单权限
- **THEN** 前端 MUST 完全隐藏“智能中心”及其子菜单入口
- **AND** 页面 MUST NOT 以置灰入口或空白页方式暴露不可用功能

#### Scenario: 插件资源不回流宿主

- **WHEN** 实现智能中心页面、API、SQL 或多语言资源
- **THEN** 这些资源 MUST 放在 `apps/lina-plugins/linapro-ai-core/` 的统一插件目录结构中
- **AND** 修改插件目录前 MUST 检查并遵守该插件根目录 `AGENTS.md`

### Requirement: 渠道页面必须维护渠道和模型

系统 SHALL 在“渠道”页面和插件 API 中提供渠道及其模型的增删改查能力。渠道 MUST 支持启停、协议地址、密钥引用和备注；模型 MUST 关联渠道，并维护默认端点、协议、名称、来源和启停状态。模型管理 MUST NOT 要求管理员声明、筛选或编辑能力方法、token 上限或 thinking 支持范围。

#### Scenario: 渠道页面按渠道和模型维度分 Tab 展示

- **WHEN** 用户打开“渠道”页面
- **THEN** 页面 MUST 显示“渠道管理”和“模型管理”两个 Tab
- **AND** 两个 Tab 标题 MUST 带有与管理对象匹配的图标
- **AND** 页面内容高度 MUST 稳定，不得因 Tab、表格或内容自适应布局循环增长而导致完整内容不可见
- **AND** “渠道管理”Tab MUST 展示渠道维度列表和渠道操作
- **AND** “模型管理”Tab MUST 按模型维度分页展示所有模型
- **AND** 模型列表 MUST 包含模型名称、渠道、协议、端点、启用状态、更新时间和操作入口
- **AND** 模型管理 Tab MUST 支持新增模型、编辑指定模型和删除指定模型
- **AND** 新增模型入口从渠道行触发时 MUST 默认选中当前渠道

#### Scenario: 查询渠道列表

- **WHEN** 用户打开“渠道”页面
- **THEN** 系统 MUST 分页返回渠道列表
- **AND** 每行 MUST 包含当前页面所需的最小渠道投影、密钥脱敏摘要、端点投影和模型摘要列表
- **AND** 后端 MUST 使用集合化查询或聚合查询装配模型摘要，MUST NOT 诱导前端逐渠道补查详情
- **AND** 页面 MUST 将 OpenAI-compatible 和 Anthropic-compatible 地址合并到“端点”列分行展示，并用标签标识端点类型
- **AND** 多个端点同时存在时，页面 MUST 完整展示每个端点地址，协议标签 MUST NOT 遮挡或挤压地址内容
- **AND** OpenAI-compatible 和 Anthropic-compatible 协议标签 MUST 使用一致宽度并右对齐，使端点 URL 保持左对齐
- **AND** 页面 MUST 在端点列右侧展示“密钥”列，密钥内容必须脱敏展示
- **AND** 密钥脱敏 MUST 基于原有存储值保留可识别前缀和末尾 2 位，中间使用星号隐藏，例如 `sk-1234567890` 展示为 `sk-**********90`
- **AND** 页面 MUST 在更新时间左侧展示“模型”列，MUST NOT 展示“模型数”或“启用模型数”列

#### Scenario: 创建或更新渠道

- **WHEN** 用户创建或更新渠道
- **THEN** 系统 MUST 校验渠道名称、协议地址和启用状态
- **AND** 系统 MUST 只保存密钥引用、加密密文或脱敏摘要
- **AND** API 响应 MUST NOT 返回 API key 明文
- **AND** 页面表单 MUST 使用“名称”作为渠道名称字段的用户可见标签
- **AND** 页面表单 MUST 使用“API 密钥”作为密钥字段名称
- **AND** 修改渠道时，页面 MUST NOT 将脱敏密钥回填到密钥输入框
- **AND** 修改渠道时，密钥输入框 MUST 为空，并通过 placeholder 提示“留空则保持原密钥”
- **AND** 页面表单字段 MUST 单独一行展示，不得把状态字段或其他输入域并排放在同一行
- **AND** 渠道新增和修改表单 MUST NOT 嵌入模型维护内容

#### Scenario: 维护渠道模型

- **WHEN** 用户在渠道下新增、编辑、删除或启停模型
- **THEN** 系统 MUST 校验模型属于目标渠道
- **AND** 模型 MUST 维护 `protocol`、默认 endpoint、模型名称和启用状态
- **AND** 模型 MUST NOT 维护或展示 `capabilityType`、`capabilityMethod`、`supportsThinking`、`supportedEfforts` 或 token 上限作为用户可编辑字段
- **AND** 渠道列表工具栏 MUST 同时提供“新增渠道”和“新增模型”入口，且两个按钮 MUST 使用可区分的样式
- **AND** 渠道列表操作列 MUST 在编辑和删除按钮下方展示“新增模型”按钮
- **AND** 渠道列表操作列 MUST 在“新增模型”按钮下方展示“同步模型”按钮，并保持操作列三行布局
- **AND** 渠道列表模型列 MUST 使用常规字重展示模型名，不得加粗模型名
- **AND** 模型删除入口 MUST 使用图标按钮展示，不得使用文字 `X` 作为删除按钮

#### Scenario: 平台模型名后缀不得传给渠道

- **WHEN** 平台保存的模型名称以 `[...]` 形式携带工具专用后缀，例如 `mimo-v2.5-pro[1m]`
- **THEN** 系统 MUST 允许该模型名称作为平台侧模型身份、页面展示、档位绑定和调用日志投影
- **AND** OpenAI-compatible 与 Anthropic-compatible provider adapter 在真正向渠道发起请求时 MUST 去除末尾 `[...]` 后缀
- **AND** 系统 MUST NOT 因去除后缀而修改已保存模型名称、响应中的平台模型投影或调用日志中的模型名称快照

#### Scenario: 同步模型列表

- **WHEN** 用户触发渠道模型同步
- **THEN** 页面 MUST 在渠道列表操作列中只提供一个“同步模型”入口
- **AND** 系统 MUST 按该渠道已启用端点的协议分别调用对应模型列表接口，并汇总写入公开模型元数据
- **AND** 单个协议端点不支持查询模型或查询失败时 MUST 保留其他协议端点的成功同步结果
- **AND** 仅当该渠道所有可同步端点查询模型都失败时，系统 MUST 向页面返回同步失败错误
- **AND** 同步失败 MUST 保留既有手工模型和被引用模型
- **AND** 系统 MUST NOT 自动推断未由渠道或用户明确声明的 `thinkingEffort` 支持范围

### Requirement: 渠道和模型删除必须保护档位引用

系统 SHALL 在删除渠道或模型前检查智能中心档位绑定引用，防止 `framework.ai.text.v1` 出现悬空配置。

#### Scenario: 阻止删除被档位使用的渠道

- **WHEN** 用户删除某个渠道
- **AND** 该渠道下任一模型被 `basic`、`standard` 或 `advanced` 档位绑定引用
- **THEN** 系统 MUST 拒绝删除
- **AND** 错误 MUST 说明该渠道正在被 AI 能力档位使用

#### Scenario: 阻止删除被档位使用的模型

- **WHEN** 用户删除某个模型
- **AND** 该模型被任一 AI 能力档位绑定引用
- **THEN** 系统 MUST 拒绝删除
- **AND** 错误 MUST 说明该模型正在被 AI 能力档位使用

#### Scenario: 删除未引用模型

- **WHEN** 用户删除未被任何档位绑定引用的模型
- **THEN** 系统 MUST 允许删除或软删除该模型
- **AND** 系统 MUST NOT 修改其他渠道、模型或档位配置

### Requirement: 档位管理必须提供三档文本 AI 能力配置

系统 SHALL 在“档位管理”页面和插件 API 中固定提供 `basic`、`standard`、`advanced` 三个文本能力档位。每个档位 MUST 能绑定一个主渠道模型，配置启用状态、默认 `thinkingEffort` 和测试入口。

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
- **AND** 系统 MUST 校验渠道和模型真实存在、已启用且模型属于渠道
- **AND** 系统 MUST NOT 要求模型存在匹配的能力方法声明
- **AND** 是否适配当前能力方法由管理员通过选择、保存和测试结果判断
- **AND** 系统 MUST 将该模型设置为该档位的主绑定

#### Scenario: 默认 thinkingEffort 校验

- **WHEN** 用户为档位配置默认 `thinkingEffort`
- **THEN** 系统 MUST 校验默认值为空或属于 `low`、`medium`、`high`、`xhigh`、`max`
- **AND** 系统 MUST NOT 依赖模型能力声明预先拒绝保存
- **AND** 若所选渠道协议在实际调用时不支持该 effort，测试或运行时调用 MUST 返回结构化错误

#### Scenario: 禁用档位

- **WHEN** 用户关闭某个档位的启用状态并保存
- **THEN** 系统 MUST 保留该档位已有绑定
- **AND** 后续文本能力调用 MUST 将该档位视为不可用

### Requirement: 档位管理页面必须方便配置和诊断

系统 SHALL 在“档位管理”页面以三档配置为核心展示对象，并提供可发现的保存、测试、校验和状态反馈。页面 MUST 复用现有前端表格、表单、弹窗、抽屉和操作列模式。

#### Scenario: 展示三档配置

- **WHEN** 用户进入“档位管理”页面
- **THEN** 页面 MUST 按 `basic`、`standard`、`advanced` 稳定顺序展示三档
- **AND** 每档 MUST 展示启用状态、渠道、模型、协议、默认 `thinkingEffort`、最近测试状态和保存入口

#### Scenario: 模型选择按渠道过滤

- **WHEN** 用户在某个档位选择渠道
- **THEN** 模型选择项 MUST 展示该渠道下所有已启用模型
- **AND** 模型选项 MUST 展示模型名称和协议分组
- **AND** 模型选择 MUST NOT 按能力方法声明、thinking 支持范围或 token 上限过滤

#### Scenario: 不支持 effort 时提示

- **WHEN** 用户配置默认 `thinkingEffort`
- **THEN** 页面 MUST 允许保存枚举范围内的值
- **AND** 页面 MUST 通过测试结果或运行时错误提示模型/协议不支持情况

### Requirement: 档位必须支持轻量测试

系统 SHALL 允许用户对任一文本能力档位执行轻量连通性测试。测试 MUST 使用已保存绑定或请求中的草稿绑定，MUST NOT 在测试草稿绑定时持久化档位配置。

#### Scenario: 测试已保存档位

- **WHEN** 用户触发某个已启用且已配置主绑定的档位测试
- **THEN** 系统 MUST 使用该档位绑定的渠道模型执行轻量文本生成请求
- **AND** 响应 MUST 返回测试状态、耗时、实际渠道、实际模型、协议、实际 `thinkingEffort` 和测试时间

#### Scenario: 测试草稿绑定

- **WHEN** 用户选择渠道、模型和默认 `thinkingEffort` 但尚未保存并触发测试
- **THEN** 系统 MUST 使用请求中的草稿绑定执行测试
- **AND** 系统 MUST 校验草稿绑定真实存在
- **AND** 系统 MUST NOT 依赖模型能力声明预先拒绝草稿绑定
- **AND** 系统 MUST NOT 创建、更新或删除该档位主绑定

#### Scenario: 测试失败保留配置

- **WHEN** 档位测试因密钥、网络、协议、模型、thinking effort 或渠道错误失败
- **THEN** 系统 MUST 返回脱敏且可理解的失败摘要
- **AND** 系统 MUST NOT 删除或修改该档位已有配置

### Requirement: 文本调用日志必须可监控且不泄露敏感内容

系统 SHALL 在“调用日志”页面和插件 API 中提供文本 `AI` 调用日志查询。日志 MUST 支持分页、筛选、用量和耗时分析，MUST NOT 保存完整输入、完整输出、隐藏思考内容或密钥。

#### Scenario: 记录成功调用

- **WHEN** 文本 `AI` 调用成功
- **THEN** 系统 MUST 记录 `requestId`、`capabilityType`、`capabilityMethod`、`purpose`、档位、调用来源、渠道模型投影、协议、`thinkingEffort`、状态、token 用量、耗时和创建时间
- **AND** 系统 MUST NOT 保存完整 `messages`、完整响应正文、隐藏思考内容或 API key

#### Scenario: 记录失败调用

- **WHEN** 文本 `AI` 调用失败
- **THEN** 系统 MUST 记录可解析的渠道模型投影、失败状态、耗时、错误码和脱敏错误摘要
- **AND** 错误摘要 MUST NOT 包含密钥、认证头、完整请求体或完整响应体

#### Scenario: 查询调用日志

- **WHEN** 用户调用 `GET /api/ai/invocations`
- **THEN** 系统 MUST 按创建时间倒序分页返回日志
- **AND** 查询 MUST 支持按 `capabilityType`、`capabilityMethod`、`purpose`、档位、状态、渠道、模型、来源插件和时间范围过滤
- **AND** 页面展示调用协议时 MUST 使用统一展示名，例如 `OpenAI` 和 `Anthropic`，不得直接展示数据库中的大小写不统一原始协议值
- **AND** 后端 MUST 在数据库侧完成过滤、排序和分页

#### Scenario: 清理调用日志

- **WHEN** 用户调用 `DELETE /api/ai/invocations/clean`
- **THEN** 系统 MUST 清理调用日志
- **AND** 未传时间范围时 MUST 清理全部调用日志
- **AND** 传入 `startedAt` 或 `endedAt` 时 MUST 仅清理创建时间落入该时间范围的调用日志
- **AND** 清理操作 MUST 在数据库侧执行范围删除并返回实际删除数量
- **AND** 系统 MUST 继续要求平台上下文和独立的调用日志清理权限

#### Scenario: 调用日志页面清理入口

- **WHEN** 用户打开“调用日志”页面
- **THEN** 页面工具栏删除按钮 MUST 使用与系统监控操作日志、登录日志一致的危险主按钮样式
- **AND** 页面筛选区 MUST 使用均衡的响应式网格布局，桌面常规宽度下首行 MUST 展示“调用方法”“用途”“档位”“状态”四个筛选条件，第二行 MUST 展示“创建时间”“来源插件”两个筛选条件，避免日期范围控件过宽或右侧字段被挤压
- **AND** 页面筛选区同一网格列上下两行的筛选标题文字 MUST 按统一右边界对齐，避免“调用方法/创建时间”或“用途/来源插件”等标题上下错位
- **AND** 页面筛选区“调用方法”和“创建时间”筛选项左侧 MUST 保留安全间距，英文标题不得越出搜索栏左边界，且不得通过调整其它搜索条件控件宽度来实现
- **AND** 页面筛选区“创建时间”控件 MUST 与同列上方“调用方法”控件保持一致宽度，且两个控件可见宽度 MUST 为 `242px`
- **AND** 页面筛选区创建时间组件 MUST 使用日期范围选择和标准输入样式，日期文本 MUST 完整可读，且不得通过修改系统监控操作日志、登录日志的原有筛选宽度或额外宽度补偿来实现对齐
- **AND** 页面筛选区调用方法、用途、档位、创建时间、状态和来源插件控件 MUST 保持可读宽度，其它列控件可不与创建时间等宽但 MUST 留出足够输入宽度，不得出现控件重叠、明显截断或与相邻标题互相侵占
- **AND** 页面筛选区各筛选字段组之间的水平间距 MUST 保持一致，不得依赖某一列的额外内边距或剩余列宽形成不均衡间隔
- **WHEN** 用户点击删除按钮
- **THEN** 页面 MUST 弹出与系统监控操作日志、登录日志清理弹窗一致的对话框
- **AND** 弹窗 MUST 包含警告提示、删除所有调用日志选项和日期时间范围选择区
- **AND** 日期时间范围选择区 MUST 与上方提示和选项区域保持清晰间隔
- **AND** 默认未选择删除所有调用日志时，确认前 MUST 要求选择完整日期时间范围
- **AND** 选择删除所有调用日志后，日期时间范围组件 MUST 禁用，确认时 MUST 请求 `DELETE /api/ai/invocations/clean` 且不携带 `startedAt` 或 `endedAt`

### Requirement: 插件存储必须由 linapro-ai-core 自有并具备性能边界

系统 SHALL 由 `linapro-ai-core` 插件维护渠道、模型、档位、绑定和调用日志表。插件 SQL MUST 使用 PostgreSQL 源语法、幂等 DDL/Seed、软删除语义、合理索引和插件表名前缀。

#### Scenario: 安装插件创建表和固定档位

- **WHEN** 安装或初始化 `linapro-ai-core`
- **THEN** 插件 SQL MUST 创建 `plugin_linapro_ai_provider`、`plugin_linapro_ai_model`、`plugin_linapro_ai_tier`、`plugin_linapro_ai_tier_binding` 和 `plugin_linapro_ai_invocation`
- **AND** SQL MUST seed `text.generate` 能力方法的 `basic`、`standard`、`advanced` 三个固定档位
- **AND** 档位唯一约束 MUST 覆盖 `capability_type`、`capability_method` 和 `code`
- **AND** SQL MUST 满足重复执行结果一致

#### Scenario: 高频查询具备索引

- **WHEN** 系统查询渠道列表、模型列表、档位解析或调用日志
- **THEN** 表结构 MUST 为渠道模型关联、档位绑定解析、日志时间范围、状态、档位、渠道和模型过滤提供必要索引
- **AND** 设计 MUST 支持集合化读取和当前页批量装配，MUST NOT 依赖动态结果集逐行查询

#### Scenario: 不写入自增主键

- **WHEN** 插件 SQL 写入固定档位或初始化数据
- **THEN** SQL MUST 使用稳定业务键和唯一约束实现幂等
- **AND** SQL MUST NOT 显式写入自增 `id`

### Requirement: 档位解析缓存必须保持一致性

系统 SHALL 在 `linapro-ai-core` provider 内维护文本档位解析缓存，用于将 `capabilityType + capabilityMethod + tier` 解析为可调用渠道模型绑定。缓存权威源 MUST 是插件数据库，失效 MUST 与配置写入成功耦合。

#### Scenario: 配置变更后失效缓存

- **WHEN** 用户创建、更新、启停或删除渠道、模型、档位或绑定
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

系统 SHALL 将智能中心管理 API 作为平台配置控制面。渠道、模型、档位和日志管理 MUST 要求平台上下文和对应权限；动态插件文本调用 MUST 使用 host service 授权，不得绕过管理面权限。

#### Scenario: 平台管理员访问管理 API

- **WHEN** 平台上下文用户拥有对应权限并访问智能中心管理 API
- **THEN** 系统 MUST 允许执行授权范围内的读取、创建、更新、删除或测试动作
- **AND** 操作 MUST 进入统一权限和审计边界

#### Scenario: 非平台上下文被拒绝

- **WHEN** 租户上下文或代管租户上下文请求创建、更新、删除渠道、模型或档位
- **THEN** 系统 MUST 在执行数据库写入或渠道调用前拒绝请求
- **AND** 错误 MUST 不泄露目标资源是否存在于平台配置中

#### Scenario: 插件调用不获得管理权限

- **WHEN** 动态插件获得 `ai.text.generate` host service 授权
- **THEN** 该授权 MUST 只允许按授权 purpose 调用文本能力
- **AND** 插件 MUST NOT 因该授权访问渠道、档位管理或调用日志管理 API

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

### Requirement: 动态插件文本 AI 授权不得授予管理权限

系统 SHALL 确保动态插件获得`ai.text.generate`或其他`ai`host service 方法授权后，只能通过类型化`AI`能力发起对应方法调用。该授权 MUST NOT 授予渠道、模型、档位管理或调用日志管理 API 的访问权限。

#### Scenario: 插件调用不获得管理权限

- **WHEN** 动态插件获得`ai.text.generate`host service 方法授权
- **THEN** 该授权 MUST 只允许调用文本生成能力
- **AND** 插件 MUST NOT 因该授权访问渠道、档位管理或调用日志管理 API
- **AND** 插件请求中的`purpose`、`tier`和其他参数 MUST NOT 被解释为管理权限授权来源

### Requirement: linapro-ai-core 必须拥有 AI 领域公开契约

系统 SHALL 扩展`linapro-ai-core`插件职责，使其除管理页面、provider、模型、档位和调用日志外，还拥有`AI`领域公开契约。插件 MUST 在`backend/cap/aicap`发布普通消费接口、DTO、命名类型、错误语义、方法状态、动态 guest SDK、provider SPI、descriptor helper 和版本策略。`backend/internal`继续承载 provider、模型路由、分层、调用日志、外部协议和业务实现。

#### Scenario: 创建 AI cap 目录

- **WHEN** 实现本变更
- **THEN** `apps/lina-plugins/linapro-ai-core/backend/cap/aicap` MUST 存在并作为`AI`领域公开契约入口
- **AND** 该目录 MUST NOT import `linapro-ai-core/backend/internal/**`
- **AND** 该目录 MUST NOT 暴露`DAO`、`DO`、`Entity`、密钥、模型路由内部配置或调用日志内部结构

#### Scenario: 插件 ID 不再依赖 core aicap

- **WHEN** `linapro-ai-core`声明自身插件 ID、provider ID 或能力 owner ID
- **THEN** 这些稳定标识 MUST 由插件自身公开契约或 manifest 维护
- **AND** `backend/plugin.go`不得通过 core `aicap/aitext`常量间接获得 owner 身份

### Requirement: linapro-ai-core 必须发布 AI capability descriptor

系统 SHALL 要求`linapro-ai-core`在源码插件注册阶段发布`AI`能力 descriptor。descriptor MUST 描述 owner 插件 ID、`ai`能力键、`v1`协议版本、当前已发布方法、风险等级、资源形态、源码契约、动态契约、provider factory、运行依赖和启用策略。descriptor 注册 MUST 通过 owner 插件提供的类型安全 helper 完成。已发布方法 MUST 与真实 invoker 路径一致，不得把仅存在 DTO 的多模态方法提前写入授权 catalog。

#### Scenario: 注册文本 provider

- **WHEN** `linapro-ai-core`注册文本生成 provider
- **THEN** 插件 MUST 使用`aicap.ProviderDescriptor`或等价 helper 将 typed factory 包装为带 invoker 的通用 descriptor
- **AND** invoker MUST 通过`aicap.Service`分发，不得在 SPI 包内复制一套并行业务 switch
- **AND** 不得调用 core `plugin.Providers().ProvideAIText`

#### Scenario: 发布动态方法目录

- **WHEN** `linapro-ai-core`发布`ai.v1`descriptor
- **THEN** descriptor MUST 发布当前可运行的文本和方法状态 methods
- **AND** 管理端和动态授权展示 MUST 能显示这些方法来自`linapro-ai-core`
- **AND** 尚未接线的多模态 methods MUST NOT 出现在 descriptor 授权目录中

### Requirement: linapro-ai-core bridge SDK 必须复用 owner DTO

系统 SHALL 要求`linapro-ai-core/backend/cap/aicap/bridge`或等价公开包提供动态插件 guest SDK。该 SDK MUST 复用`backend/cap/aicap`下普通消费 DTO 或 owner 契约定义的投影，MUST 只负责编码 owner-aware host call、声明 helper 和错误映射，不得包含 provider SPI、源码插件注册 API、宿主 dispatcher 或内部业务实现。

#### Scenario: 动态 SDK 复用文本 DTO

- **WHEN** 动态插件使用 bridge SDK 调用文本生成
- **THEN** 请求响应类型 MUST 与源码插件消费契约保持同一 owner
- **AND** 不得在 core protocol 中维护并行的文本生成 DTO owner

#### Scenario: bridge SDK 不绕过授权

- **WHEN** 动态插件通过 SDK 调用`AI`方法
- **THEN** SDK MUST 通过通用 host call 进入宿主授权和审计
- **AND** SDK 不得直接调用`linapro-ai-core/backend/internal/service`

### Requirement: linapro-ai-core i18n 和文档必须覆盖 owner 契约

系统 SHALL 按`linapro-ai-core/plugin.yaml`中的`i18n.enabled: true`治理插件 owner 契约的用户可见文案、错误 fallback、API 文档源文本、动态授权展示名称和 README 文档。插件自身运行时语言包和`apidoc`翻译 MUST 维护在该插件`manifest/i18n/<locale>/`和`manifest/i18n/<locale>/apidoc/`下，不得集中写入`lina-core`语言资源。

#### Scenario: 新增 owner 错误码

- **WHEN** `linapro-ai-core`新增`CodeCapabilityDenied`、`CodeCapabilityUnavailable`或等价 owner 错误码
- **THEN** 错误必须具备稳定 errorCode、messageKey、messageParams 和英文 fallback
- **AND** 插件必须维护对应目标语言资源

#### Scenario: 更新插件 README

- **WHEN** 完成`AI`契约 owner 迁移
- **THEN** `apps/lina-plugins/linapro-ai-core/README.md`和`README.zh-CN.md` MUST 同步说明`backend/cap/aicap`、源码插件消费、动态 SDK、provider SPI、依赖声明和版本策略
- **AND** 两个 README 的事实内容必须一致

