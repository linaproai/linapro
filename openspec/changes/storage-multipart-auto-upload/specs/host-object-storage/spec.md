## ADDED Requirements

### Requirement: 宿主 storage.Service 必须透传文件命名空间的 Multipart 能力

系统 SHALL 使宿主内部 `storage.Service` 在 `NamespaceFiles` 上能够透传当前 Resolve 后的 active provider 的 Multipart 与 part 级访问签发能力。文件中心 MUST 通过该宿主服务完成对象字节的 Multipart 写入，不得在 file 服务内直接 import 云 SDK。当 active provider 不支持 Multipart 时，对应方法 MUST 返回明确不支持错误或由上层回退拼装 Put，MUST NOT 静默选择其他 provider。

#### Scenario: 文件中心 Multipart 走宿主 storage

- **WHEN** 文件中心对 NamespaceFiles 执行 CreateMultipart / UploadPart / Complete
- **THEN** 调用 MUST 进入宿主 `storage.Service` 再委派 active provider
- **AND** 多云冲突时 MUST 与普通 Put 相同 fail-closed

#### Scenario: 非文件命名空间首期可不支持 Multipart

- **WHEN** 调用方对非 `NamespaceFiles` 请求 Multipart
- **THEN** 系统可返回不支持
- **AND** MUST NOT 破坏既有 Put/Get/Delete 行为
