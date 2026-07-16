## ADDED Requirements

### Requirement: 分片上传必须使用与整包上传相同的租户分区路径

当文件中心通过中转分片或直传 Multipart 写入**新的**物理对象时，系统 SHALL 使用与整包 `/file/upload` 及既有直传 single 相同的相对存储 key 规则：`<tenantId>/<yyyy>/<MM>/<generated-file-name>`（无历史 `t/` 目录层），并在 complete 成功后将该相对 key 写入 `sys_file.path`。历史记录路径兼容策略 MUST 保持不变。

#### Scenario: 中转分片 complete 后 path 规则一致

- **WHEN** 用户通过 proxy multipart 完成上传
- **THEN** `sys_file.path` MUST 符合租户分区相对 key 规则
- **AND** MUST 与同等条件下整包上传生成规则一致

#### Scenario: 直传 Multipart complete 后 path 规则一致

- **WHEN** 用户通过 direct multipart 完成上传
- **THEN** `sys_file.path` MUST 与 init 时宿主分配的 key 一致
- **AND** MUST 符合相同租户分区规则
