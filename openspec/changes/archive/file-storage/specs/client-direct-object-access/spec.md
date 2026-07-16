# client-direct-object-access Specification

## Purpose

定义客户端直连对象访问的通用契约与生命周期：能力探测、中立访问模式、init/complete/abort、安全 scope、local 降级，以及与领域元数据确认的边界。适用于文件中心直传与插件 Storage 直连编排。

## Requirements

### Requirement: 系统必须提供可探测的客户端直连对象访问契约

系统 SHALL 在对象存储扩展契约上提供客户端直连访问能力探测与中立访问描述。访问描述 MUST 使用通用 mode（至少包括 `presigned_url`、`form_post`、`temporary_credentials`、以及降级用的 `proxy`），MUST NOT 要求调用方按云厂商 ID 分支。长期访问密钥（永久 AK/SK 或等价）MUST NOT 下发给浏览器或不受信任客户端。

#### Scenario: 支持直连的 provider 返回中立访问描述

- **WHEN** 当前 active 对象后端支持对某操作（put 或 get）的客户端直连
- **AND** 调用方请求创建直连访问
- **THEN** 系统 MUST 返回包含 mode、过期时间与执行上传/下载所需字段的中立描述
- **AND** 响应 MUST NOT 包含永久密钥或可无界访问整个桶的凭证

#### Scenario: 不支持直连时返回 proxy 降级

- **WHEN** 当前后端为 local，或 active provider 不支持该操作的直连
- **AND** 调用方请求创建直连访问
- **THEN** 系统 MUST 指示 `mode=proxy`（或等价“需服务端中转”语义）
- **AND** MUST NOT 假装已签发可用的云端直连 URL

### Requirement: 直连上传必须采用 init-complete 生命周期

对需要领域元数据确认的直连上传（至少包括文件中心），系统 SHALL 提供 init、客户端传输、complete 三阶段；MAY 提供 abort。Init MUST 在签发访问前完成鉴权与策略校验。Complete MUST 在写入或确认领域状态前校验对象已存在且满足会话约束。仅 init 成功但未 complete 的对象 MUST NOT 被当作已登记业务文件。

#### Scenario: 文件中心直传成功落库

- **WHEN** 已授权用户对合法 scene 与大小内的文件执行 init
- **AND** 客户端按返回的 direct 描述完成上传
- **AND** 客户端调用 complete 且对象校验通过
- **THEN** 系统 MUST 创建与 multipart 上传语义一致的 `sys_file` 记录
- **AND** 记录的存储 path/engine 规则 MUST 与统一对象后端策略一致

#### Scenario: 未 complete 不产生业务文件记录

- **WHEN** 用户 init 成功并可能已向云端写入对象
- **AND** 未调用 complete 或 complete 校验失败
- **THEN** 系统 MUST NOT 创建成功状态的业务文件元数据记录供列表当作已上传文件
- **AND** 校验失败时 MUST 返回可诊断错误

#### Scenario: init 阶段拒绝超限或未授权

- **WHEN** 文件大小超过有效上传上限，或用户缺少上传权限，或 scene/类型不被允许
- **THEN** init MUST 失败
- **AND** MUST NOT 签发可用的直连写访问

### Requirement: 直连会话必须绑定安全 scope

系统 SHALL 将每次直连上传会话绑定到至少：租户上下文、操作者、目标 scoped object key（或可映射到其的领域标识）、允许的最大 size、可选 content-type 约束与过期时间。客户端 MUST NOT 自行指定可逃逸隔离边界的云端 object key。会话过期后 complete MUST 失败。

#### Scenario: 会话过期后 complete 失败

- **WHEN** init 签发的访问与会话已过期
- **AND** 客户端调用 complete
- **THEN** 系统 MUST 拒绝 complete
- **AND** MUST NOT 基于过期会话创建业务文件记录

#### Scenario: 对象 key 由宿主分配

- **WHEN** 文件中心或插件 Storage 创建直连 put 访问
- **THEN** 写入目标 key MUST 由宿主按既有隔离与路径规则生成
- **AND** 调用方不得通过请求参数覆盖为任意桶内路径

### Requirement: 提供内容 hash 时必须支持秒传

当客户端在 init 时提供内容 SHA-256，且文件中心在相同租户作用域下已存在可复用的相同内容对象时，系统 SHALL 直接返回已有文件元数据而不签发新的写访问（秒传）。

#### Scenario: 相同 hash 秒传

- **WHEN** init 请求携带已存在且可复用的 content hash
- **THEN** 响应 MUST 指示无需上传（或直接返回文件信息）
- **AND** MUST NOT 要求客户端再次传输文件字节

### Requirement: 直连下载必须可签发短时读访问

当对象内容位于支持直连 get 的 active 云后端时，系统 SHALL 能签发短时只读访问描述。Local 或不支持时 MUST 保持服务端中转读取。公开或业务访问 URL 的兼容策略 MUST 避免在未鉴权场景下泄露私有对象。

#### Scenario: 云对象下载可直链

- **WHEN** 已授权用户请求下载位于云后端的文件
- **AND** provider 支持 get 直连
- **AND** 调用方选择直链偏好
- **THEN** 系统 MUST 返回未过期的只读直连访问描述或等价跳转
- **AND** 该访问 MUST 在过期后失效

#### Scenario: local 对象下载保持中转

- **WHEN** 文件内容仅存在于 local 后端
- **AND** 用户请求下载
- **THEN** 系统 MUST 通过宿主中转提供内容
- **AND** MUST NOT 返回无效云直链

### Requirement: 前端与 API 调用方不得绑定厂商

工作台通用上传组件与文件中心直传 API 的调用方 MUST 仅依赖宿主通用字段（mode 与执行载荷）。系统 MUST NOT 要求业务页面引入特定云厂商 SDK 作为默认上传路径。

#### Scenario: 通用上传组件双模式

- **WHEN** 工作台使用通用 FileUpload/ImageUpload 上传
- **AND** init 返回 direct mode
- **THEN** 组件 MUST 按中立 mode 执行直传并 complete
- **WHEN** init 返回 proxy 或不支持直传
- **THEN** 组件 MUST 回退现有 multipart 上传且业务 props 不变

### Requirement: 直传执行失败不得伪装成功

当客户端在 direct 模式下传输失败（含 CORS、网络、云端 4xx/5xx）时，系统与前端 MUST NOT 将上传标记为成功，MUST NOT 调用成功态 complete 写入业务记录。

#### Scenario: 直传网络失败

- **WHEN** 浏览器直传因 CORS 或网络错误失败
- **THEN** 上传 UI MUST 展示失败状态
- **AND** MUST NOT 创建成功的文件中心记录
