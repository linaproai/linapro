# object-multipart-upload Specification

## Purpose
TBD - created by archiving change storage-multipart-auto-upload. Update Purpose after archive.
## Requirements
### Requirement: 系统必须提供可探测的对象 Multipart 上传契约

系统 SHALL 在对象存储扩展契约上提供可选的 Multipart 上传能力探测与会话操作。实现 MUST 以中立方式表达 CreateMultipart、UploadPart、CompleteMultipart、AbortMultipart，MUST NOT 要求调用方按云厂商 ID 分支业务逻辑。未实现该可选能力的 provider（含内置 local）MUST 报告不支持云 Multipart，且 MUST NOT 假装已完成云端 Multipart 会话。

#### Scenario: 支持 Multipart 的 provider 可创建会话

- **WHEN** 当前 active 对象后端实现 Multipart 可选接口且配置可用
- **AND** 调用方请求 CreateMultipart
- **THEN** 系统 MUST 返回宿主侧可绑定的 upload 会话标识（或 provider upload id 的封装）
- **AND** 后续 UploadPart / Complete / Abort MUST 仅针对该会话与宿主分配的 object key 生效

#### Scenario: 不支持 Multipart 时探测为 false

- **WHEN** 当前后端为 local，或 active provider 未实现 Multipart
- **AND** 调用方探测 SupportsMultipart
- **THEN** 系统 MUST 返回不支持
- **AND** MUST NOT 签发可用的云 Multipart 会话

### Requirement: 文件中心必须支持服务端中转分片上传

系统 SHALL 为文件中心提供服务端中转分片上传生命周期（init、上传分片、complete、abort）。中转分片 MUST 在 complete 成功后写入与整包上传语义一致的 `sys_file` 记录。当 active provider 支持云 Multipart 时，宿主 MUST 优先将各分片 UploadPart 到云并 CompleteMultipart；否则 MUST 将分片拼装后以单次对象写入落盘，且 MUST NOT 在 complete 前将对象视为已登记业务文件。

#### Scenario: 中转分片成功落库

- **WHEN** 已授权用户对合法 scene 与大小内的文件启动中转分片会话
- **AND** 客户端按会话依次提交全部分片
- **AND** 客户端 complete 且校验通过
- **THEN** 系统 MUST 创建与 `/file/upload` 整包上传语义一致的 `sys_file` 记录
- **AND** 记录的 path 规则 MUST 与统一对象后端及租户分区路径策略一致

#### Scenario: 中转分片未 complete 不产生业务记录

- **WHEN** 用户已上传部分或全部分片但未 complete，或 complete 失败
- **THEN** 系统 MUST NOT 创建成功状态的业务文件元数据供列表当作已上传文件

#### Scenario: 中转分片 abort 清理会话

- **WHEN** 客户端或系统对中转分片会话执行 abort
- **THEN** 系统 MUST 丢弃会话状态
- **AND** 若已创建云 Multipart 会话 MUST 尽最大努力 AbortMultipart
- **AND** 缺失会话 MUST 视为成功 no-op

### Requirement: 文件中心必须支持客户端直传 Multipart

当 active 对象后端同时支持客户端直连 put 与云 Multipart 时，系统 SHALL 支持文件中心 direct multipart：init 创建宿主会话与云 Multipart，客户端通过宿主签发的 **单 part** 短时访问描述上传各分片，complete 时提交 parts 列表并由宿主 CompleteMultipart 后写入 `sys_file`。长期密钥 MUST NOT 下发；object key MUST 仍由宿主分配。

#### Scenario: 直传 Multipart 成功落库

- **WHEN** init 策略为 direct + multipart
- **AND** 客户端为每个 part 获取短时访问并完成上传
- **AND** complete 携带完整有效的 partNumber 与 etag 列表且云 Complete 成功
- **THEN** 系统 MUST 写入与整包/中转上传语义一致的 `sys_file` 记录

#### Scenario: part 访问必须绑定会话与 partNumber

- **WHEN** 客户端请求某一 part 的直传访问
- **THEN** 签发的访问 MUST 仅能写入该会话绑定的 Multipart upload 与指定 partNumber
- **AND** 客户端 MUST NOT 通过参数改写为任意桶内 key

#### Scenario: 直传 Multipart complete 校验失败

- **WHEN** parts 不完整、etag 无效、会话过期或云 Complete 失败
- **THEN** complete MUST 失败
- **AND** MUST NOT 写入成功态业务文件记录

### Requirement: 系统必须按阈值自动选择分片上传

系统 SHALL 根据运行时配置的分片阈值、上传大小上限与后端能力，在上传规划（init）阶段决定 `channel`（direct 或 proxy）与 `encoding`（single 或 multipart）。当文件声明大小 **大于等于** `sys.upload.multipartThresholdMB` 对应字节数，且存在可用分片路径时，系统 MUST 选择 multipart encoding；小于阈值时 MUST 使用 single encoding。自动决策结果 MUST 通过中立字段返回客户端，业务页面 MUST NOT 被要求按厂商分支。

#### Scenario: 超过阈值且支持云 Multipart 时自动直传分片

- **WHEN** 文件大小 ≥ multipart 阈值且 ≤ maxSize
- **AND** active 后端支持 direct put 与云 Multipart
- **THEN** init 响应的 strategy MUST 为 channel=direct 且 encoding=multipart

#### Scenario: 超过阈值但不支持云 Multipart 时自动中转分片

- **WHEN** 文件大小 ≥ multipart 阈值且 ≤ maxSize
- **AND** active 后端不支持云 Multipart（含 local）
- **THEN** init 响应的 strategy MUST 为 channel=proxy 且 encoding=multipart
- **AND** 客户端 MUST 使用中转分片 API 而非单次整包上传（在本能力交付后的客户端实现中）

#### Scenario: 低于阈值保持单对象上传

- **WHEN** 文件大小 < multipart 阈值
- **THEN** encoding MUST 为 single
- **AND** 仍可按既有规则在 direct 与 proxy 之间选择 channel

#### Scenario: 超过 maxSize 拒绝规划

- **WHEN** 文件大小超过有效 `sys.upload.maxSize`
- **THEN** init/规划 MUST 失败
- **AND** MUST NOT 创建上传会话或签发写访问

### Requirement: 上传大小与分片相关配置必须可运行时治理

系统 SHALL 提供并校验以下运行时配置（键名可等价，语义必须具备）：

- `sys.upload.maxSize`：单文件上限（MB），**默认 200**
- `sys.upload.multipartThresholdMB`：自动分片阈值（MB），**默认 100**
- `sys.upload.multipartPartSizeMB`：建议/强制分片大小（MB），默认正整数且不得小于云兼容最小中间分片要求（至少 5）
- `sys.upload.multipartMaxConcurrency`：客户端建议并行 part 数，默认正整数

系统 MUST 拒绝使 `multipartThresholdMB >= maxSize` 的无效组合（或在生效时明确使自动分片不可用并记录可诊断原因）。修改配置后新上传规划 MUST 使用新值。

#### Scenario: 默认 maxSize 为 200MB

- **WHEN** 部署使用默认 seed 且无运行时覆盖
- **THEN** 有效上传上限 MUST 为 200MB

#### Scenario: 默认阈值 100MB 触发自动分片评估

- **WHEN** 使用默认 threshold 与 maxSize
- **AND** 用户上传 150MB 文件且后端支持相应分片路径
- **THEN** 系统 MUST 选择 multipart encoding

#### Scenario: 非法 threshold 配置被拒绝

- **WHEN** 管理员将 multipartThresholdMB 设置为大于或等于 maxSize 的值
- **THEN** 系统 MUST 拒绝保存或拒绝将该组合作为有效自动分片配置生效

### Requirement: 通用上传客户端必须自动执行分片策略

工作台通用上传入口（`uploadApi` 及 FileUpload/ImageUpload）MUST 消费 init 返回的 strategy，并在 encoding=multipart 时自动切分文件、上传分片并 complete；业务组件 props MUST NOT 因分片能力而强制增加厂商或手动分片配置。直传 multipart 执行失败时 MUST NOT 静默改为中转分片或伪装成功。

#### Scenario: 自动直传分片无业务改造

- **WHEN** 业务页面使用默认 FileUpload 上传超过阈值的文件
- **AND** init 返回 direct + multipart
- **THEN** 上传 MUST 自动完成分片直传与 complete
- **AND** 业务 props 无需声明 partSize 或云厂商

#### Scenario: 自动中转分片无业务改造

- **WHEN** init 返回 proxy + multipart
- **THEN** 上传 MUST 自动走中转分片 API
- **AND** 成功后 v-model 绑定的文件 ID 语义与整包上传一致

#### Scenario: 直传分片失败不静默降级

- **WHEN** 直传 multipart 因网络或云错误失败
- **THEN** 上传 UI MUST 展示失败
- **AND** MUST NOT 自动改走 proxy 并宣称成功
- **AND** MUST 尽最大努力 abort 会话

