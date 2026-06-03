## ADDED Requirements

### Requirement: 智能中心必须只管理 AI 能力元数据和 provider 实现

系统 SHALL 让`linapro-ai-core`作为智能中心管理多模态`AI`能力元数据、供应商端点、模型能力、档位、调用日志和 provider adapter。`linapro-ai-core` MUST NOT 承载具体业务场景任务实现。

#### Scenario: 视频业务任务不进入智能中心

- **WHEN** 业务模块需要实现“生成视频”“批量转写”或等价长耗时业务流程
- **THEN** 业务模块 MUST 自行管理业务任务、队列、进度、重试、通知和资产归属
- **AND** `linapro-ai-core` MUST NOT 新增`/api/ai/video-jobs`、视频业务任务表或用户进度页面

#### Scenario: provider operation 只用于协议适配

- **WHEN** 供应商返回异步 operation ID
- **THEN** `linapro-ai-core`如需保存或暴露 operation 状态，MUST 只保存最小 provider operation 投影用于后续查询和诊断
- **AND** 该投影 MUST NOT 包含业务对象 ID、业务状态机或业务通知配置

### Requirement: 智能中心必须使用可扩展 provider endpoint 模型

系统 SHALL 将供应商协议端点建模为 provider 下的可扩展 endpoint 记录。系统 MUST NOT 通过在 provider 主表追加按协议命名的固定基础地址列、密钥引用列或等价字段来支持新协议。

#### Scenario: 创建 provider endpoint

- **WHEN** 管理员为供应商新增协议端点
- **THEN** 系统 MUST 保存`providerId`、`protocol`、`baseUrl`、`secretRef`或等价密钥引用、启用状态和必要元数据
- **AND** API 响应 MUST NOT 返回 API key 明文

#### Scenario: 一个 provider 支持多个协议

- **WHEN** 同一供应商同时配置 OpenAI-compatible、Anthropic-compatible 或 Voyage-compatible 端点
- **THEN** 系统 MUST 允许多个 endpoint 关联同一个 provider
- **AND** 供应商列表 MUST 使用当前页 provider ID 集合化装配 endpoint 摘要
- **AND** 前端 MUST NOT 对每个 provider 逐项请求 endpoint 详情

#### Scenario: 删除被引用 endpoint

- **WHEN** 管理员删除或禁用某个 endpoint
- **AND** 该 endpoint 被启用模型能力或档位绑定引用
- **THEN** 系统 MUST 在破坏绑定前拒绝操作或要求先解除引用
- **AND** 错误 MUST 是结构化且可本地化的业务错误

### Requirement: 模型必须声明多模态能力方法

系统 SHALL 为模型维护能力方法声明。模型是否支持`image.generate`、`embedding.create`、`audio.transcribe`、`vision.analyze`、`document.analyze`、`safety.moderate`或`video.generate`等方法 MUST 由模型能力记录显式表达。模型基础记录 MUST NOT 再保存`capabilityType`、`capabilityMethod`、token 上限、`thinkingEffort`支持范围或其他方法级能力字段。

#### Scenario: 查询模型能力

- **WHEN** 前端查询供应商模型列表或档位可选模型
- **THEN** API MUST 返回模型支持的`capabilityType`、`capabilityMethod`、输入模态、输出模态、token 上限、资产数量上限、资产大小上限和 operation 支持状态
- **AND** 后端 MUST 使用批量查询或集合化投影装配当前页模型能力
- **AND** 模型方法筛选、候选模型查询和档位绑定校验 MUST 以模型能力记录为唯一事实来源

#### Scenario: 模型同步不自动推断能力

- **WHEN** 管理员从供应商同步模型列表
- **THEN** 系统 MUST 只写入供应商明确返回且平台可确认的公开模型元数据
- **AND** 系统 MUST NOT 自动推断未声明的多模态能力、`thinkingEffort`支持范围或视频 operation 支持

#### Scenario: 档位绑定校验模型能力

- **WHEN** 管理员把模型绑定到某个能力方法档位
- **THEN** 系统 MUST 校验该模型启用且显式声明支持目标`capabilityType + capabilityMethod`
- **AND** 不支持目标方法时 MUST 拒绝保存

#### Scenario: 模型基础表不重复保存能力方法字段

- **WHEN** 系统保存、同步或更新供应商模型基础信息
- **THEN** 模型基础记录 MUST 只保存供应商、默认 endpoint、模型名称、协议、来源和启用状态等模型身份字段
- **AND** 方法级能力、输入输出约束、`thinkingEffort`支持和默认参数 MUST 保存到模型能力记录

### Requirement: 档位和默认参数必须按能力方法管理

系统 SHALL 继续使用`capabilityType + capabilityMethod + tierCode`作为档位身份。系统 SHALL 允许每个能力方法拥有`basic`、`standard`、`advanced`三档及其默认参数，但默认参数 MUST 按方法语义分别定义。

#### Scenario: 查询能力方法档位

- **WHEN** 前端按`capabilityType=image`和`capabilityMethod=generate`查询档位
- **THEN** 系统 MUST 返回该能力方法下的`basic`、`standard`、`advanced`档位投影
- **AND** 响应 MUST 包含主绑定、默认参数摘要、最近测试摘要和 Unix 毫秒更新时间

#### Scenario: 默认参数不跨模态复用

- **WHEN** 管理员配置`audio.transcribe`档位
- **THEN** 系统 MUST 使用音频转写相关默认参数
- **AND** 系统 MUST NOT 暴露`thinkingEffort`、图片尺寸或视频时长等不属于该方法的默认参数

#### Scenario: 固定档位 seed

- **WHEN** 插件安装或初始化多模态档位
- **THEN** SQL MUST 使用稳定业务键和唯一约束为目标能力方法 seed 固定档位
- **AND** SQL MUST NOT 显式写入自增`id`

### Requirement: 调用日志必须覆盖多模态方法且保持脱敏

系统 SHALL 扩展智能中心调用日志，使其覆盖多模态`AI`方法。日志 MUST 支持分页筛选、最小审计和用量诊断，MUST NOT 保存完整输入输出或大对象内容。

#### Scenario: 记录多模态成功调用

- **WHEN** `image.generate`、`audio.transcribe`、`vision.analyze`、`document.analyze`、`safety.moderate`或`video.generate`调用成功
- **THEN** 系统 MUST 记录 request ID、能力方法、purpose、来源插件、租户和用户投影、供应商模型投影、状态、耗时、用量和资产引用摘要
- **AND** 系统 MUST NOT 保存完整图片、音频、视频、文档、prompt 或供应商响应原文

#### Scenario: 查询调用日志

- **WHEN** 管理员查询调用日志
- **THEN** API MUST 支持按`capabilityType`、`capabilityMethod`、`purpose`、`tier`、`status`、`providerId`、`modelId`、`sourcePluginId`和时间范围过滤
- **AND** 后端 MUST 在数据库侧完成过滤、排序和分页

#### Scenario: 日志首期平台可见

- **WHEN** 用户不是平台管理员或缺少`ai:invocation:list`权限
- **THEN** 系统 MUST 拒绝查询智能中心调用日志
- **AND** 系统 MUST NOT 通过日志接口泄露租户外业务输入存在性

### Requirement: 智能中心缓存必须按能力方法保持一致

系统 SHALL 为多模态能力方法解析 provider、endpoint、model、tier 和默认参数维护受控缓存。缓存权威源 MUST 是`linapro-ai-core`插件数据库，失效 MUST 与配置写入成功耦合。

#### Scenario: 配置变更后失效方法缓存

- **WHEN** 管理员创建、更新、启停或删除 provider、endpoint、model、model capability、tier、binding 或默认参数
- **THEN** 系统 MUST 在数据库写入成功后失效相关`capabilityType + capabilityMethod`解析缓存
- **AND** 后续调用 MUST 使用刷新后的配置或在 cache miss 时从数据库重建

#### Scenario: 集群模式同步失效

- **WHEN** `cluster.enabled=true`且某节点修改多模态能力配置
- **THEN** 系统 MUST 通过宿主统一集群协调、共享修订号、事件广播或等价机制让其他节点观察到失效
- **AND** 系统 MUST NOT 只刷新当前节点本地内存

#### Scenario: 缓存故障降级

- **WHEN** 多模态能力解析缓存不可用或条目缺失
- **THEN** provider adapter MUST 从插件数据库重建解析结果
- **AND** 数据库不可用时 MUST 返回结构化不可用错误并记录脱敏失败摘要

### Requirement: 智能中心多模态页面必须按能力类型组织交互

系统 SHALL 在智能中心页面中按能力类型组织多模态配置入口。档位管理页面当前版本 MUST 使用能力类型`Tab`降低操作复杂度，后端和内部请求仍 MUST 使用`capabilityType + capabilityMethod + tierCode`作为档位身份。页面 MUST 复用现有`Vben`、`vxe-table`、表单、抽屉、弹窗和操作列模式，并遵守插件`i18n`治理。

#### Scenario: 供应商页面展示 endpoint 和模型能力

- **WHEN** 管理员打开供应商页面
- **THEN** 页面 MUST 分页展示 provider、endpoint 摘要、密钥脱敏摘要和模型能力摘要
- **AND** 后端 MUST 一次性返回当前页所需投影，MUST NOT 诱导前端逐行补查

#### Scenario: 档位页面按能力类型 Tab 切换

- **WHEN** 管理员在档位管理页选择文档能力`Tab`
- **THEN** 页面 MUST 展示该能力类型当前默认方法下的三档配置、绑定模型、默认参数和最近测试结果
- **AND** 页面 MUST 不再要求管理员通过顶部搜索表单选择`document.analyze`等具体能力方法
- **AND** `Tab`标题 MUST 通过插件运行时`i18n`资源渲染，英文标题使用首字母大写，中文标题使用专业能力名称
- **AND** 页面内部请求和保存仍 MUST 带上目标`capabilityType`和默认`capabilityMethod`

#### Scenario: 不展示业务任务页面

- **WHEN** 智能中心扩展视频能力
- **THEN** 页面如需展示 operation 状态，MUST 只展示 provider operation 诊断摘要
- **AND** 页面 MUST NOT 提供业务视频任务管理、业务进度通知或业务资产归属管理
