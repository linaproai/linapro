# host-object-storage Specification

## Purpose
TBD - created by archiving change unify-host-object-storage. Update Purpose after archive.
## Requirements
### Requirement: 宿主必须提供中立对象存储服务

系统 SHALL 提供宿主内部 `storage.Service` 作为对象内容读写的唯一通用 owner。该服务 MUST 通过 namespace 和相对 key 表达存储对象位置，并提供 `Put`、`Get`、`Delete`、`Stat` 和有界 `List` 能力。该服务不得依赖文件中心元数据、插件 logical path、动态授权快照、HTTP 请求对象或插件公开 DTO。

#### Scenario: 文件中心和插件存储复用同一存储 owner

- **WHEN** 文件中心写入上传对象
- **AND** 插件本地 provider 写入插件对象
- **THEN** 两者都通过宿主内部 `storage.Service` 执行底层对象内容读写
- **AND** 文件中心不得依赖 `storagecap.Service`
- **AND** 插件公开能力不得暴露内部 `storage.Service`

### Requirement: 对象 key 必须被限制在命名空间内

系统 SHALL 要求 `storage.Service` 在访问本地 provider 前规范化 namespace 和 key。空 namespace、绝对路径、目录穿越、混用反斜杠绕过、空对象 key 和解析后逃逸存储根的路径 MUST 被拒绝。

#### Scenario: 拒绝目录穿越 key

- **WHEN** 调用方使用 `../secret.txt`、`/secret.txt` 或 `a/../../secret.txt` 作为对象 key
- **THEN** `storage.Service` 拒绝该请求
- **AND** 本地 provider 不得访问存储根之外的文件

#### Scenario: 合法相对 key 被锚定到命名空间

- **WHEN** 调用方在 namespace `files` 下写入 key `42/2026/06/demo.png`
- **THEN** 本地 provider 只在 `files/42/2026/06/demo.png` 对应的存储根子路径内写入对象
- **AND** 返回给调用方的对象 key 不包含本地绝对路径

### Requirement: 对象列表必须有有界性能契约

系统 SHALL 要求 `storage.Service.List` 接收 prefix 和 limit，并限制返回数量。未提供 limit 时 MUST 使用默认上限，超过最大上限时 MUST 截断到最大上限或返回明确错误。本地 provider 不得提供无界遍历整个存储根的生产调用路径。

#### Scenario: 按前缀有界列出对象

- **WHEN** 调用方按 namespace `plugins` 和 prefix `tenant/1001/plugin-a/reports/` 列出对象
- **THEN** `storage.Service` 只返回该 namespace 和 prefix 下的对象
- **AND** 返回数量不得超过系统定义的有效 limit

### Requirement: 中立存储不得拥有领域 URL 或业务元数据

系统 SHALL 要求 `storage.Service` 只返回对象 key、大小、content type、etag、更新时间和读取流等对象存储元数据。文件中心 URL、`sys_file` 主键、插件 logical path、插件 ID、租户 ID 解释、业务场景和数据权限判断 MUST 由调用方所属领域维护。

#### Scenario: 文件中心生成下载 URL

- **WHEN** 文件中心通过 `storage.Service` 写入对象并创建 `sys_file` 记录
- **THEN** `storage.Service` 不返回 `/api/v1/uploads/...` URL
- **AND** 文件中心根据自身领域规则生成公开访问 URL

#### Scenario: 插件存储隐藏 provider key

- **WHEN** 插件通过 `storagecap.Service.Stat` 读取对象元数据
- **THEN** 插件只看到 logical path 和对象元数据
- **AND** 插件不得看到内部 `storage.Service` namespace、provider key 或本地绝对路径

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

