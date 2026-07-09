## ADDED Requirements

### Requirement: 宿主必须发布文本 AI 抽象能力

系统 SHALL 在 `apps/lina-core` 中发布版本化文本 `AI` 抽象能力 `framework.ai.text.v1`。该能力 MUST 只定义消费契约、状态查询、降级语义和 provider 接入边界，MUST NOT 在宿主中持有渠道、模型、档位或调用日志业务存储。

#### Scenario: 消费方通过文本能力接口调用

- **WHEN** 宿主模块、源码插件或动态插件需要执行文本生成
- **THEN** 调用方 MUST 通过 `framework.ai.text.v1` 的消费接口发起调用
- **AND** 调用方 MUST NOT 直接依赖 `linapro-ai-core` 的 `backend/internal/**`、插件表、渠道密钥结构或 provider adapter

#### Scenario: 官方插件提供文本能力实现

- **WHEN** `linapro-ai-core` 插件处于平台级可用状态并声明 `framework.ai.text.v1` provider
- **THEN** `framework.ai.text.v1` 的消费 service MUST 将文本生成调用委托给该 provider
- **AND** 返回值 MUST 使用该能力自有 DTO、投影和值对象

#### Scenario: 文本生成映射到固定能力方法

- **WHEN** 调用方通过 `framework.ai.text.v1` 执行 `GenerateText`
- **THEN** 宿主契约 MUST 将该调用视为 `capabilityType=text` 与 `capabilityMethod=generate`
- **AND** Go 契约 MUST 使用命名类型和常量表达该能力方法
- **AND** 调用方 MUST NOT 通过请求字段把 `GenerateText` 改写为图片、向量、音频或其他方法

#### Scenario: 渠道存储不进入宿主

- **WHEN** 系统实现 `framework.ai.text.v1`
- **THEN** `apps/lina-core` MUST NOT 新增渠道、模型、档位或调用日志业务表
- **AND** 宿主公开契约 MUST NOT 暴露插件内部 `DAO`、`DO`、`Entity`、缓存快照或密钥明文

### Requirement: 宿主能力目录必须通过 AI 命名空间暴露 AI 能力

系统 SHALL 在根 `capability.Services` 中通过 `AI() ai.Service` 暴露 `AI` 能力族。根能力目录 MUST NOT 直接暴露 `AIText()`、`AIImage()`、`AIEmbedding()` 或其他按 `AI` 子能力展开的根方法。

#### Scenario: 源码插件获取文本 AI 能力

- **WHEN** 源码插件通过宿主发布的能力目录获取文本 `AI` 能力
- **THEN** 插件 MUST 通过 `services.AI().Text()` 获取文本能力
- **AND** 插件 MUST NOT 通过 `services.AIText()` 获取文本能力

#### Scenario: 宿主内部模块获取文本 AI 能力

- **WHEN** 宿主内部模块可选消费文本 `AI` 能力
- **THEN** 模块 MUST 通过显式注入的能力目录调用 `AI().Text()`
- **AND** 模块 MUST NOT 直接依赖 `linapro-ai-core/backend/internal/**` 或 provider adapter

#### Scenario: 根能力目录新增后续 AI 子能力

- **WHEN** 系统后续新增图片、向量、音频或其他 `AI` 子能力
- **THEN** 新子能力 MUST 挂载到 `ai.Service` 下
- **AND** 新子能力 MUST NOT 在根 `capability.Services` 上新增 `AI*()` 方法

### Requirement: AI 聚合服务必须只承担子能力聚合职责

系统 SHALL 使用 `ai.Service` 聚合 `AI` 子能力。`ai.Service` MUST 只暴露类型化子能力入口，例如 `Text() aitext.Service`，MUST NOT 作为弱类型 `AI` 网关执行运行时 method 分发。

#### Scenario: 文本能力通过 Text 入口访问

- **WHEN** 调用方需要执行同步文本生成
- **THEN** 调用方 MUST 使用 `AI().Text().GenerateText(...)`
- **AND** `Text()` 返回的 service MUST 保持 `framework.ai.text.v1` 的状态、降级和错误语义

#### Scenario: 弱类型 AI 网关被拒绝

- **WHEN** 实现 `AI` 能力聚合服务
- **THEN** 系统 MUST NOT 引入 `Generate(ctx, capabilityType, payload)`、`Invoke(ctx, method, payload)` 或等价弱类型业务网关作为普通消费契约
- **AND** 文本、图片、向量等子能力 MUST 维护各自的 DTO、错误和授权边界

### Requirement: 文本 AI 能力必须归属 AI 命名空间

系统 SHALL 将文本 `AI` 能力包归属到 `apps/lina-core/pkg/plugin/capability/ai/aitext`。生产代码 MUST 使用该新路径引用文本能力契约，旧 `apps/lina-core/pkg/plugin/capability/aitext` 路径 MUST 不再作为生产消费入口保留。

#### Scenario: 生产代码引用文本 AI 契约

- **WHEN** 宿主、源码插件或动态插件生产代码引用文本 `AI` 契约
- **THEN** 代码 MUST import `lina-core/pkg/plugin/capability/ai/aitext`
- **AND** 代码 MUST NOT import `lina-core/pkg/plugin/capability/aitext`

#### Scenario: 文本能力行为保持不变

- **WHEN** 文本能力包迁移到 `capability/ai/aitext`
- **THEN** `framework.ai.text.v1` 的 capability ID、`Available(ctx)`、`Status(ctx)`、`GenerateText(ctx, request)` 和 provider factory 语义 MUST 保持不变
- **AND** 迁移 MUST NOT 新增渠道、模型、档位或调用日志宿主存储

### Requirement: 文本 AI 来源身份必须由能力服务注入

系统 SHALL 将文本生成消费请求与 provider 内部请求分离。普通调用方可见的 `GenerateRequest` MUST NOT 要求填写 `SourcePluginID`；插件来源身份 MUST 由 plugin-scoped 能力 service 或动态插件 host-call 上下文注入到 provider 请求。

#### Scenario: 源码插件调用注入插件来源

- **WHEN** 源码插件通过 `ServicesForPlugin(services, pluginID).AI().Text()` 发起文本生成
- **THEN** 文本能力 service MUST 将该 `pluginID` 作为 provider 请求的来源插件标识
- **AND** 普通调用方 MUST NOT 在消费请求中自行填写或伪造 `SourcePluginID`

#### Scenario: 动态插件调用注入插件来源

- **WHEN** 动态插件通过 `ai.text.generate` host service 发起文本生成
- **THEN** `WASM` host service handler MUST 使用 host-call 上下文中的 `pluginID` 作为 provider 请求来源
- **AND** 该来源 MUST 与智能中心调用日志和宿主服务审计中的来源插件保持一致

#### Scenario: 宿主内部调用不伪造插件来源

- **WHEN** 宿主内部模块直接使用 `AI().Text()` 发起文本生成
- **THEN** 文本能力 service MUST 保持来源为空或使用规范定义的宿主来源标识
- **AND** 宿主内部调用 MUST NOT 被记录为任意源码插件或动态插件来源

### Requirement: 文本生成请求响应必须稳定且可扩展

系统 SHALL 为 `framework.ai.text.v1` 定义同步文本生成请求和响应契约。请求 MUST 包含 `purpose`、`tier`、`messages`、可选生成参数和可选 `thinkingEffort`；响应 MUST 返回文本、实际档位、实际渠道模型投影、用量、耗时和 Unix 毫秒时间点。

#### Scenario: 请求使用消息数组

- **WHEN** 调用方构造文本生成请求
- **THEN** 请求 MUST 使用 `messages` 数组表达输入
- **AND** 每条消息首期 MUST 至少包含 `role` 和纯文本 `content`
- **AND** 调用方 MUST NOT 通过 `metadata` 承载大段 prompt、diff、文件内容或业务原文

#### Scenario: 响应返回最小渠道投影

- **WHEN** 文本生成成功
- **THEN** 响应 MUST 包含生成文本、`tier`、`providerName`、`modelName`、`protocol`、`usage.inputTokens`、`usage.outputTokens`、`latencyMs` 和 `generatedAt`
- **AND** `generatedAt` MUST 是 Unix timestamp in milliseconds
- **AND** 响应 MUST NOT 返回 API key、secret 引用解析结果或 provider 内部配置

#### Scenario: 无效档位被拒绝

- **WHEN** 调用方传入不是 `basic`、`standard` 或 `advanced` 的文本档位
- **THEN** 系统 MUST 在执行渠道调用前拒绝请求
- **AND** 错误 MUST 是结构化业务错误，包含可诊断的错误码和可本地化消息键

### Requirement: 文本能力必须支持 thinkingEffort 抽象参数

系统 SHALL 在文本生成请求中预留可选 `thinkingEffort` 参数。`thinkingEffort` MUST 使用平台统一枚举 `low`、`medium`、`high`、`xhigh`、`max`。模型管理 MUST NOT 声明或预先限制支持范围，具体是否可用 MUST 由测试调用、真实运行结果或 provider adapter 的结构化错误反馈判断。

#### Scenario: 请求合法 thinkingEffort

- **WHEN** 调用方请求 `thinkingEffort: high`
- **THEN** 系统 MUST 先校验该值属于平台枚举集合
- **AND** provider adapter SHOULD 将平台枚举映射到目标渠道协议支持的字段或等价参数
- **AND** 调用日志 MUST 记录请求值和实际应用值

#### Scenario: 渠道或协议不支持请求的 thinkingEffort

- **WHEN** 测试调用或真实调用发现目标渠道、协议或模型不支持请求的 `thinkingEffort`
- **THEN** 系统 MUST 返回结构化业务错误
- **AND** 系统 MUST NOT 静默降级到其他 effort
- **AND** 系统 MUST NOT 向渠道发送不受支持的专有 thinking 参数

#### Scenario: 未传 thinkingEffort

- **WHEN** 调用方未传 `thinkingEffort`
- **THEN** 系统 MUST 使用档位默认值或模型默认行为
- **AND** 若档位默认值在实际调用中不受目标渠道、协议或模型支持，系统 MUST 返回结构化错误并指示需要修正档位配置

### Requirement: 文本能力必须提供可用性和降级状态

系统 SHALL 为 `framework.ai.text.v1` 提供 `Available(ctx)` 和 `Status(ctx)` 等状态能力。插件禁用、卸载、provider 冲突、档位未配置、模型禁用或密钥不可用时，系统 MUST 返回明确的不可用状态或业务错误，而不是产生宿主 500。

#### Scenario: Provider 插件不可用

- **WHEN** `linapro-ai-core` 插件被禁用、卸载或启动失败
- **THEN** `Available(ctx)` MUST 返回不可用
- **AND** `Status(ctx)` MUST 返回能力 ID、provider 插件状态和不可用原因
- **AND** 调用 `GenerateText` MUST 返回结构化不可用错误

#### Scenario: 档位未配置

- **WHEN** 调用方使用未配置启用主绑定的档位生成文本
- **THEN** 系统 MUST 拒绝调用
- **AND** 错误 MUST 明确指出该档位未配置可用渠道模型

#### Scenario: 可选消费方降级

- **WHEN** 业务功能可选使用文本 `AI` 能力
- **AND** `Available(ctx)` 返回不可用
- **THEN** 业务功能 MUST 隐藏入口、返回零值或按自身规范提示配置缺失
- **AND** 业务功能 MUST NOT 直接暴露 provider 内部错误或空白页面

### Requirement: 文本能力必须保护敏感输入输出

系统 SHALL 将文本 `AI` 调用视为敏感执行路径。宿主抽象能力和 provider adapter MUST NOT 默认记录完整输入、完整输出、隐藏思考内容、密钥、diff 或业务原文。

#### Scenario: 成功调用不记录完整输入输出

- **WHEN** 文本生成成功
- **THEN** 系统 MAY 记录 `purpose`、档位、渠道模型投影、token 用量、耗时和状态
- **AND** 系统 MUST NOT 在宿主日志、调用日志或审计摘要中保存完整 `messages` 或完整生成正文

#### Scenario: 失败调用脱敏

- **WHEN** 渠道调用失败
- **THEN** 错误摘要 MUST 脱敏 API key、认证头、请求体和响应体中的敏感片段
- **AND** 返回给调用方的错误 MUST 保留足够诊断信息，但不得包含密钥或业务原文

### Requirement: 文本能力版本必须为后续多模态能力保留边界

系统 SHALL 将 `framework.ai.text.v1` 限定为同步文本生成能力。图片、音频、向量、重排、工具调用、多模态消息和流式输出 MUST 通过后续独立能力、独立方法或新版本契约扩展，MUST NOT 破坏 `v1` 文本同步响应语义。

#### Scenario: 能力方法不复用文本契约字段

- **WHEN** 后续新增 `image.generate`、`embedding.create`、`audio.transcribe` 或 `audio.synthesize`
- **THEN** 新能力 MUST 使用独立 capability method、host service method 或 `framework.ai.*` 新契约
- **AND** 新能力 MUST NOT 复用 `thinkingEffort`、`messages` 或文本响应字段作为跨模态通用参数

#### Scenario: 后续新增图片能力

- **WHEN** 系统后续新增图片生成或图片理解能力
- **THEN** 新能力 MUST 使用独立 capability type、host service method 或 `framework.ai.*` 新契约
- **AND** `framework.ai.text.v1` 的请求响应字段和错误语义 MUST 保持兼容

#### Scenario: 后续新增流式文本

- **WHEN** 系统后续需要流式文本输出
- **THEN** 系统 MUST 新增独立 streaming 方法或 `framework.ai.text.v2`
- **AND** 现有 `GenerateText` MUST 继续保持同步调用和单次响应语义

### Requirement: 官方插件必须提供智能中心菜单和页面

系统 SHALL 新增官方源码插件 `linapro-ai-core`，并由该插件贡献"智能中心"菜单。菜单 MUST 包含"渠道""档位管理""调用日志"三个页面，且页面、路由、权限和静态资源 MUST 归属于插件目录。

#### Scenario: 插件启用后显示智能中心

- **WHEN** `linapro-ai-core` 插件已安装、启用且当前用户拥有对应菜单权限
- **THEN** 前端 MUST 显示"智能中心"菜单
- **AND** 菜单下 MUST 显示"渠道""档位管理""调用日志"
- **AND** "智能中心"顶级菜单 MUST 排在"扩展中心"之前

#### Scenario: 插件禁用或无权限时隐藏入口

- **WHEN** `linapro-ai-core` 插件被禁用、卸载或当前用户缺少菜单权限
- **THEN** 前端 MUST 完全隐藏"智能中心"及其子菜单入口
- **AND** 页面 MUST NOT 以置灰入口或空白页方式暴露不可用功能

#### Scenario: 插件资源不回流宿主

- **WHEN** 实现智能中心页面、API、SQL 或多语言资源
- **THEN** 这些资源 MUST 放在 `apps/lina-plugins/linapro-ai-core/` 的统一插件目录结构中
- **AND** 修改插件目录前 MUST 检查并遵守该插件根目录 `AGENTS.md`

### Requirement: 渠道页面必须维护渠道和模型

系统 SHALL 在"渠道"页面和插件 API 中提供渠道及其模型的增删改查能力。渠道 MUST 支持启停、协议地址、密钥引用和备注；模型 MUST 关联渠道，并维护默认端点、协议、名称、来源和启停状态。模型管理 MUST NOT 要求管理员声明、筛选或编辑能力方法、token 上限或 thinking 支持范围。

#### Scenario: 渠道页面按渠道和模型维度分 Tab 展示

- **WHEN** 用户打开"渠道"页面
- **THEN** 页面 MUST 显示"渠道管理"和"模型管理"两个 Tab
- **AND** 两个 Tab 标题 MUST 带有与管理对象匹配的图标
- **AND** 页面内容高度 MUST 稳定，不得因 Tab、表格或内容自适应布局循环增长而导致完整内容不可见
- **AND** "渠道管理"Tab MUST 展示渠道维度列表和渠道操作
- **AND** "模型管理"Tab MUST 按模型维度分页展示所有模型
- **AND** 模型列表 MUST 包含模型名称、渠道、协议、端点、启用状态、更新时间和操作入口
- **AND** 模型管理 Tab MUST 支持新增模型、编辑指定模型和删除指定模型
- **AND** 新增模型入口从渠道行触发时 MUST 默认选中当前渠道

#### Scenario: 查询渠道列表

- **WHEN** 用户打开"渠道"页面
- **THEN** 系统 MUST 分页返回渠道列表
- **AND** 每行 MUST 包含当前页面所需的最小渠道投影、密钥脱敏摘要、端点投影和模型摘要列表
- **AND** 后端 MUST 使用集合化查询或聚合查询装配模型摘要，MUST NOT 诱导前端逐渠道补查详情
- **AND** 页面 MUST 将 OpenAI-compatible 和 Anthropic-compatible 地址合并到"端点"列分行展示，并用标签标识端点类型
- **AND** 多个端点同时存在时，页面 MUST 完整展示每个端点地址，协议标签 MUST NOT 遮挡或挤压地址内容
- **AND** OpenAI-compatible 和 Anthropic-compatible 协议标签 MUST 使用一致宽度并右对齐，使端点 URL 保持左对齐
- **AND** 页面 MUST 在端点列右侧展示"密钥"列，密钥内容必须脱敏展示
- **AND** 密钥脱敏 MUST 基于原有存储值保留可识别前缀和末尾 2 位，中间使用星号隐藏，例如 `sk-1234567890` 展示为 `sk-**********90`
- **AND** 页面 MUST 在更新时间左侧展示"模型"列，MUST NOT 展示"模型数"或"启用模型数"列

#### Scenario: 创建或更新渠道

- **WHEN** 用户创建或更新渠道
- **THEN** 系统 MUST 校验渠道名称、协议地址和启用状态
- **AND** 系统 MUST 只保存密钥引用、加密密文或脱敏摘要
- **AND** API 响应 MUST NOT 返回 API key 明文
- **AND** 页面表单 MUST 使用"名称"作为渠道名称字段的用户可见标签
- **AND** 页面表单 MUST 使用"API 密钥"作为密钥字段名称
- **AND** 修改渠道时，页面 MUST NOT 将脱敏密钥回填到密钥输入框
- **AND** 修改渠道时，密钥输入框 MUST 为空，并通过 placeholder 提示"留空则保持原密钥"
- **AND** 页面表单字段 MUST 单独一行展示，不得把状态字段或其他输入域并排放在同一行
- **AND** 渠道新增和修改表单 MUST NOT 嵌入模型维护内容

#### Scenario: 维护渠道模型

- **WHEN** 用户在渠道下新增、编辑、删除或启停模型
- **THEN** 系统 MUST 校验模型属于目标渠道
- **AND** 模型 MUST 维护 `protocol`、默认 endpoint、模型名称和启用状态
- **AND** 模型 MUST NOT 维护或展示 `capabilityType`、`capabilityMethod`、`supportsThinking`、`supportedEfforts` 或 token 上限作为用户可编辑字段
- **AND** 渠道列表工具栏 MUST 同时提供"新增渠道"和"新增模型"入口，且两个按钮 MUST 使用可区分的样式
- **AND** 渠道列表操作列 MUST 在编辑和删除按钮下方展示"新增模型"按钮
- **AND** 渠道列表操作列 MUST 在"新增模型"按钮下方展示"同步模型"按钮，并保持操作列三行布局
- **AND** 渠道列表模型列 MUST 使用常规字重展示模型名，不得加粗模型名
- **AND** 模型删除入口 MUST 使用图标按钮展示，不得使用文字 `X` 作为删除按钮

#### Scenario: 平台模型名后缀不得传给渠道

- **WHEN** 平台保存的模型名称以 `[...]` 形式携带工具专用后缀，例如 `mimo-v2.5-pro[1m]`
- **THEN** 系统 MUST 允许该模型名称作为平台侧模型身份、页面展示、档位绑定和调用日志投影
- **AND** OpenAI-compatible 与 Anthropic-compatible provider adapter 在真正向渠道发起请求时 MUST 去除末尾 `[...]` 后缀
- **AND** 系统 MUST NOT 因去除后缀而修改已保存模型名称、响应中的平台模型投影或调用日志中的模型名称快照

#### Scenario: 同步模型列表

- **WHEN** 用户触发渠道模型同步
- **THEN** 页面 MUST 在渠道列表操作列中只提供一个"同步模型"入口
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

系统 SHALL 在"档位管理"页面和插件 API 中固定提供 `basic`、`standard`、`advanced` 三个文本能力档位。每个档位 MUST 能绑定一个主渠道模型，配置启用状态、默认 `thinkingEffort` 和测试入口。

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

系统 SHALL 在"档位管理"页面以三档配置为核心展示对象，并提供可发现的保存、测试、校验和状态反馈。页面 MUST 复用现有前端表格、表单、弹窗、抽屉和操作列模式。

#### Scenario: 展示三档配置

- **WHEN** 用户进入"档位管理"页面
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

系统 SHALL 在"调用日志"页面和插件 API 中提供文本 `AI` 调用日志查询。日志 MUST 支持分页、筛选、用量和耗时分析，MUST NOT 保存完整输入、完整输出、隐藏思考内容或密钥。

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

- **WHEN** 用户打开"调用日志"页面
- **THEN** 页面工具栏删除按钮 MUST 使用与系统监控操作日志、登录日志一致的危险主按钮样式
- **AND** 页面筛选区 MUST 使用均衡的响应式网格布局，桌面常规宽度下首行 MUST 展示"调用方法""用途""档位""状态"四个筛选条件，第二行 MUST 展示"创建时间""来源插件"两个筛选条件，避免日期范围控件过宽或右侧字段被挤压
- **AND** 页面筛选区同一网格列上下两行的筛选标题文字 MUST 按统一右边界对齐，避免"调用方法/创建时间"或"用途/来源插件"等标题上下错位
- **AND** 页面筛选区"调用方法"和"创建时间"筛选项左侧 MUST 保留安全间距，英文标题不得越出搜索栏左边界，且不得通过调整其它搜索条件控件宽度来实现
- **AND** 页面筛选区"创建时间"控件 MUST 与同列上方"调用方法"控件保持一致宽度，且两个控件可见宽度 MUST 为 `242px`
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
