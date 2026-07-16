## ADDED Requirements

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
