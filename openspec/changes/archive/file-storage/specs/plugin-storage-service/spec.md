# plugin-storage-service Specification

## Purpose

定义插件 `storagecap` 领域契约在直连与 Multipart 扩展下的服务面要求：与文件中心共享后端选择语义，并在逻辑路径隔离下支持可选 Multipart 会话与大对象 Put 优化。

## Requirements

### Requirement: storagecap 必须支持可选 Multipart 会话操作

系统 SHALL 在 `storagecap` 领域契约中定义可选 Multipart 能力（探测 + Create/UploadPart/Complete/Abort，以及可选的 part 级 DirectAccess 签发）。公共插件调用面 MUST 继续以 logical path 表达对象，MUST NOT 暴露 provider object key、云 upload id 原始细节给不受信任方（源码插件若需要可使用 Service 封装后的会话句柄，但不得要求业务按厂商 ID 分支）。

#### Scenario: Service 探测 Multipart

- **WHEN** 当前 active provider 实现 Multipart 可选接口
- **THEN** `storagecap.Service`（或等价探测 API）MUST 报告支持
- **WHEN** local 或未实现 provider
- **THEN** MUST 报告不支持

#### Scenario: Multipart 写入受路径隔离约束

- **WHEN** 插件通过 Multipart 会话写入 logical path `reports/demo.bin`
- **THEN** 宿主 MUST 将会话绑定到当前插件与租户作用域下的 provider key
- **AND** 其他插件或租户 MUST NOT 向该会话 UploadPart 或 Complete

### Requirement: 大对象 Put 在支持时可以使用 Multipart 写后端

当调用方使用 `storagecap.Service.Put` 写入大对象且 active provider 支持 Multipart 时，实现 MAY 在内部将流按 part 上传并 Complete，对外仍表现为单次 Put 成功或失败。动态插件既有 `put.init/chunk/commit` 传输分片 MUST 保持可用；commit 阶段 MAY 改为 Provider Multipart 而非整文件单次 Put。

#### Scenario: 动态插件大文件 commit 可走 Multipart

- **WHEN** 动态插件通过分片传输完成大对象并 commit
- **AND** active provider 支持 Multipart
- **THEN** 宿主 MUST 将对象写入最终 logical path
- **AND** 允许通过 Multipart 而非单次 PutObject 完成后端写入

#### Scenario: 不支持 Multipart 时 Put 行为不变

- **WHEN** provider 不支持 Multipart
- **AND** 调用方 Put 或动态插件 commit
- **THEN** 系统 MUST 继续使用既有单流 Put 路径
