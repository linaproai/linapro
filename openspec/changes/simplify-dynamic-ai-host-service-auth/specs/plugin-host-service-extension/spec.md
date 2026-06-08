## ADDED Requirements

### Requirement: AI host service 调用必须受 service、method 和 DTO 能力边界约束

系统 SHALL 对每一次`ai`host service 调用校验`service: ai`、声明的`method`、调用来源和请求 DTO 边界。`ai`host service MUST NOT 使用`resources`、`resourceRef`或`purpose:<name>`作为运行时授权条件；`purpose`、`tier`、`maxOutputTokens`、资产引用和其他方法参数 MUST 由请求 DTO 承载，并由对应`AI`子能力服务或`linapro-ai-core`治理。

#### Scenario: 方法授权后调用文本能力

- **WHEN** 动态插件已获`ai.text.generate`方法授权
- **AND** 请求 DTO 中提交`purpose`、`tier`、`messages`和`maxOutputTokens`
- **THEN** host service handler MUST 将请求转换为`AI().Text().GenerateText(...)`调用
- **AND** 宿主 MUST 使用 host-call 上下文中的`pluginID`注入来源插件身份
- **AND** 宿主 MUST NOT 按`purpose:<name>`资源授权或`resources.attributes`限制该请求

#### Scenario: 未授权方法被拒绝

- **WHEN** 动态插件未声明或未获确认`ai.document.cite`对应方法授权
- **THEN** 宿主 MUST 在执行`AI().Document().Cite(...)`或任何渠道调用前拒绝
- **AND** 宿主 MUST 返回结构化授权错误

#### Scenario: 请求参数由 AI 能力服务校验

- **WHEN** 动态插件已获`ai.text.generate`方法授权但请求 DTO 中`tier`非法
- **THEN** host service handler MUST 将请求交给类型化`AI`文本能力边界处理
- **AND** 错误 MUST 保持`AI`能力服务的 DTO 校验和 provider 可用性语义

## MODIFIED Requirements

### Requirement: 宿主服务访问同时受宿主服务声明推导的能力分类和资源授权约束

系统 SHALL 对每一次宿主服务调用执行由`hostServices`声明自动推导的粗粒度 capability 校验。对于资源型 host service，系统还 SHALL 执行细粒度资源授权校验；对于`ai`这类方法授权型 host service，系统 MUST 只按`service + method`授权快照校验，不得要求插件作者声明或确认额外`resources`。只读读取型服务的`methods` MUST 表达真实 host service 调用动作，SDK typed helper 不得作为独立授权方法进入声明或运行时快照。

#### Scenario: 插件声明宿主服务策略

- **WHEN** 开发者在动态插件清单中声明`hostServices`
- **THEN** 构建器校验 service、method、资源声明（如`storage.paths`、URL 模式、`data.tables`、宿主公开配置 key 或 manifest 资源路径）和策略参数是否合法
- **AND** 对`service: ai`只校验声明的`methods`是否合法，并拒绝`resources`、`paths`、`tables`或`keys`
- **AND** 宿主根据这些 methods 自动推导内部 capability 分类快照
- **AND** 将归一化后的宿主服务策略写入运行时产物
- **AND** 宿主装载产物后恢复为当前 release 的服务授权快照

#### Scenario: 缺少授权的宿主服务调用被拒绝

- **WHEN** 插件调用未声明的 service、method 或资源型服务的未授权资源标识
- **THEN** 宿主返回显式拒绝错误
- **AND** 宿主不执行任何目标宿主服务逻辑

#### Scenario: typed helper 不作为授权方法声明

- **WHEN** 动态插件需要读取插件配置并使用 guest SDK 的`String`、`Bool`、`Int`、`Duration`或`Scan`helper
- **THEN** `plugin.yaml`只声明`service: config`和`methods: [get]`
- **AND** 宿主授权快照只记录`config.get`
- **AND** guest SDK helper 在插件侧或共享适配层基于`get`结果完成类型转换

### Requirement: 动态插件必须通过 ai.text.generate 调用文本 AI

系统 SHALL 在动态插件宿主服务体系中提供`ai`service family，并开放`text.generate`方法。动态插件 MUST 通过`hostServices`声明`service: ai`和对应`methods`申请文本`AI`调用能力，并由宿主授权快照确认后才能调用。

#### Scenario: 插件声明文本 AI 宿主服务

- **WHEN** 动态插件在`plugin.yaml`中声明`service: ai`和`methods: [text.generate]`
- **THEN** 构建器或宿主清单校验 MUST 识别该声明为文本`AI`调用权限申请
- **AND** 声明 MUST NOT 包含`resources`、`paths`、`tables`或`keys`
- **AND** 声明本身 MUST NOT 自动授予运行时调用权限，运行时只认宿主确认后的授权快照
- **AND** 运行时 MUST 将该方法映射为`capabilityType=text`与`capabilityMethod=generate`

#### Scenario: 未声明插件调用被拒绝

- **WHEN** 动态插件未声明或未获确认`ai.text.generate`授权却发起文本`AI`调用
- **THEN** 宿主 MUST 在执行`framework.ai.text.v1`或渠道调用前拒绝
- **AND** 宿主 MUST 返回结构化授权错误

### Requirement: 动态插件 AI host service 必须支持多模态方法声明

系统 SHALL 扩展动态插件`service: ai`host service，使插件可以声明图片、向量、音频、视觉、文档、安全审核和视频方法。每个方法 MUST 映射到明确的`capabilityType + capabilityMethod`和独立授权分类；这些方法授权 MUST 仅通过`methods`表达，不得通过`resources`表达`purpose`或策略属性。

#### Scenario: 插件声明图片生成能力

- **WHEN** 动态插件在`plugin.yaml`中声明`service: ai`和`methods: [image.generate]`
- **THEN** 构建器或宿主清单校验 MUST 识别该声明为`image.generate`权限申请
- **AND** 运行时 MUST 将方法映射为`capabilityType=image`和`capabilityMethod=generate`
- **AND** 声明本身 MUST NOT 自动授予运行时调用权限

#### Scenario: 插件声明音频能力

- **WHEN** 动态插件声明`audio.transcribe`或`audio.synthesize`
- **THEN** 宿主 MUST 分别识别为不同 host service 方法
- **AND** 两个方法 MUST 使用独立 payload 契约和授权分类
- **AND** 两个方法 MUST NOT 要求插件声明`purpose`资源策略

#### Scenario: computer act 声明被拒绝

- **WHEN** 动态插件声明`computer.act`、`ui.operate`或等价 UI 控制方法
- **THEN** 清单校验或运行时 MUST 拒绝该声明
- **AND** 错误 MUST 表明该方法不属于本轮`ai`host service 支持范围

### Requirement: 动态插件 guest SDK 必须通过 AI 命名空间调用文本 AI

系统 SHALL 在动态插件 guest 侧通过`AI().Text()`暴露文本`AI`能力。guest SDK 的`AI().Text().GenerateText(...)`MUST 继续使用既有`ai.text.generate`host service 协议，并保持`host:ai:text`能力分类、类型化 DTO 和脱敏审计语义；请求资源 MUST NOT 再使用`purpose:<name>`表达授权用途。

#### Scenario: 动态插件通过 AI 命名空间生成文本

- **WHEN** 动态插件需要调用文本`AI`生成能力
- **THEN** guest 代码 MUST 通过`guest.Default().AI().Text().GenerateText(...)`或等价能力目录入口发起调用
- **AND** guest SDK MUST NOT 继续要求调用方使用根目录`AIText()`方法

#### Scenario: guest AI Text 调用进入既有 host service

- **WHEN** guest SDK 执行`AI().Text().GenerateText(...)`
- **THEN** SDK MUST 构造既有`service: ai`、`method: text.generate`host service 调用
- **AND** 请求 envelope 的`resourceRef` MUST 为空
- **AND** `purpose` MUST 仅作为请求 DTO 字段传递
- **AND** 宿主 MUST 在执行文本能力或渠道调用前完成 service、method 和来源身份校验

#### Scenario: 动态插件协议不因 Go 入口重构改变

- **WHEN** 系统将 guest 侧调用入口从`AIText()`重构为`AI().Text()`
- **THEN** 动态插件`plugin.yaml`中的`hostServices`声明格式 MUST 保持`service: ai`和`methods: [text.generate]`
- **AND** `host:ai:text`的能力分类和脱敏审计语义 MUST 保持不变

## REMOVED Requirements

### Requirement: ai.text.generate 必须受 service、method 和资源授权约束

**Reason**: `purpose`资源授权和`resources.attributes`策略使动态插件清单同时承担方法授权和业务参数策略，增加复杂度，并与源码插件通过类型化`AI`能力 DTO 调用的模型不一致。

**Migration**: 使用新的`AI host service 调用必须受 service、method 和 DTO 能力边界约束`要求。动态插件只声明`service: ai`和`methods`；`purpose`、`tier`、`maxOutputTokens`等参数由请求 DTO 提交，并由对应`AI`子能力服务或`linapro-ai-core`治理。

### Requirement: 多模态 AI host service 必须按方法和资源授权

**Reason**: 多模态`AI`方法同样应采用方法授权模型，不能继续要求`purpose`资源或资源属性策略。

**Migration**: 使用新的`AI host service 调用必须受 service、method 和 DTO 能力边界约束`要求。多模态请求中的`purpose`、资产引用、输出数量和其他参数由请求 DTO 承载，并由对应`AI`子能力服务或智能中心策略治理。
