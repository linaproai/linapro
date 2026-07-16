## ADDED Requirements

### Requirement: 直连上传规划必须返回 channel 与 encoding 策略

文件中心直连上传 init（或等价规划入口）MUST 在响应中返回中立的上传策略，至少包括 `channel`（`direct` 或 `proxy`）与 `encoding`（`single` 或 `multipart`）。当 encoding 为 `multipart` 时，响应 MUST 包含分片执行参数（至少 partSize，以及可选 minPartSize、maxParts、maxConcurrency）。客户端 MUST 以策略字段为准执行上传路径，MUST NOT 仅根据历史 `access.mode=proxy` 假设「只能整包 form 上传」。

#### Scenario: init 返回 direct multipart 策略

- **WHEN** 后端支持直连 put 与云 Multipart，且文件大小达到自动分片阈值
- **THEN** init 响应 MUST 包含 channel=direct、encoding=multipart 与 uploadSessionId
- **AND** MUST 创建云 Multipart 会话（或等价 provider 会话）供后续 part 签发与 complete

#### Scenario: init 返回 proxy multipart 策略

- **WHEN** 文件大小达到自动分片阈值且不能使用 direct multipart
- **THEN** init 响应 MUST 包含 channel=proxy、encoding=multipart 与可用于中转分片的 uploadSessionId（或明确指示客户端调用中转分片 init）

#### Scenario: 旧字段兼容

- **WHEN** encoding=single 且 channel=direct
- **THEN** 响应 MUST 仍包含既有 DirectAccess 字段供单次传输
- **WHEN** encoding=single 且 channel=proxy
- **THEN** access.mode 可为 proxy，客户端 MUST 可继续使用 `/file/upload` 整包上传

### Requirement: 直连 Multipart 必须支持按 part 签发访问

系统 SHALL 提供按 `uploadSessionId` 与 `partNumber` 签发短时 part 写访问的 API。签发结果 MUST 使用既有中立 DirectAccess 模式（推荐 `presigned_url`），MUST NOT 下发永久密钥。会话过期后签发 MUST 失败。

#### Scenario: 获取第 N 个 part 的短时 URL

- **WHEN** 合法用户对未过期的 direct multipart 会话请求 partNumber=N 的访问
- **THEN** 系统 MUST 返回仅对应该 part 的短时写访问描述
- **AND** 过期时间 MUST 受 directUrlTTL 或更短约束限制

### Requirement: 直连 complete 必须支持 Multipart parts 清单

当会话 encoding 为 multipart 时，complete 请求 MUST 携带 parts 列表（partNumber 与 etag）。系统 MUST 在写 `sys_file` 前完成云 CompleteMultipart（或等价提交）。当 encoding 为 single 时，既有「对象已存在 + Stat 校验」语义 MUST 保持可用。

#### Scenario: multipart complete 提交 parts

- **WHEN** 客户端在 direct multipart 会话上调用 complete 并提交完整 parts
- **AND** 云 Complete 成功
- **THEN** 系统 MUST 创建业务文件记录并返回与整包上传一致的文件元数据

#### Scenario: single complete 行为不变

- **WHEN** 会话为 direct single
- **AND** 客户端 complete 且对象已存在
- **THEN** 系统 MUST 按既有直连 complete 语义落库
