# file-upload-storage-path Specification

## Purpose
TBD - created by archiving change simplify-upload-storage-path. Update Purpose after archive.
## Requirements
### Requirement: 新上传文件路径必须省略 tenant 缩写目录

普通文件上传写入新的物理文件时，系统 SHALL 使用 `<tenantId>/<yyyy>/<MM>/<generated-file-name>` 作为文件中心相对存储 key，并通过宿主内部 `storage.Service` 写入对象内容，不得再在租户 ID 前增加 `t` 目录层。文件中心 SHALL 继续将该相对 key 写入 `sys_file.path`，并继续拥有公开访问 URL 生成、hash 复用、业务场景和数据权限语义。

#### Scenario: 新租户文件写入简化路径

- **WHEN** 租户 `42` 上传一个此前不存在 hash 的文件
- **THEN** 系统写入的 `sys_file.path` 必须匹配 `42/<yyyy>/<MM>/<generated-file-name>`
- **AND** 路径不得以 `t/42/` 开头
- **AND** 物理对象内容通过宿主内部 `storage.Service` 写入

### Requirement: 历史上传路径必须继续兼容访问

系统 SHALL 保留已写入 `sys_file.path` 的历史相对路径语义，不得因为新上传路径规则或底层存储服务收敛而要求迁移历史 `t/<tenantId>/...` 文件。

#### Scenario: 历史 t 前缀路径继续通过元数据访问

- **WHEN** `sys_file.path` 已保存为 `t/42/2026/05/demo.png`
- **THEN** 文件访问流程必须继续按该记录路径读取存储后端
- **AND** 不得在读取时强制改写为 `42/2026/05/demo.png`
- **AND** 底层读取仍必须限制在文件中心存储 namespace 内

### Requirement: 重复 hash 上传必须保持物理文件复用语义

普通文件上传 SHALL 继续按当前租户和文件 hash 复用已存在的物理文件记录；底层存储服务收敛不得导致相同内容在同一租户内被重复写入新物理文件。

#### Scenario: 重复内容复用历史路径

- **WHEN** 租户内已有相同 hash 的历史文件记录，且其路径为 `t/<tenantId>/...`
- **AND** 用户再次上传相同内容
- **THEN** 系统必须创建新的元数据记录并复用已有物理路径
- **AND** 不得为了生成新格式路径而重复写入同一文件内容
- **AND** 不得因为底层存储服务变更而改变 `sys_file.path` 的复用值

### Requirement: 直传 init 生成的文件中心对象 key 必须遵守租户分区路径规则

当文件中心通过客户端直传写入**新的**物理对象时，系统 SHALL 使用与 multipart 上传相同的相对存储 key 规则：`<tenantId>/<yyyy>/<MM>/<generated-file-name>`（不再增加历史 `t/` 目录层），并将该相对 key 在 complete 成功后写入 `sys_file.path`。历史记录路径兼容策略 MUST 保持不变。

#### Scenario: 直传新文件 path 格式

- **WHEN** 租户用户通过直传完成一个新文件上传
- **THEN** `sys_file.path` MUST 匹配租户分区与年/月/生成文件名规则
- **AND** MUST NOT 引入新的随意 key 格式导致与中转上传分裂

#### Scenario: 秒传复用历史 path

- **WHEN** 直传 init 因 content hash 命中可复用对象而秒传
- **THEN** 系统 MUST 复用已有 `sys_file` 或已有物理 path 语义（与中转秒传一致）
- **AND** MUST NOT 再生成冲突的新 path 指向不同内容

