## ADDED Requirements

### Requirement: 文件中心在支持直连时允许字节路径绕过宿主

系统 SHALL 允许文件中心在 active 对象后端支持客户端直连时，将上传或下载的**文件字节**经客户端与云后端直接传输，而不强制经过 `lina-core` 中转。后端选择规则 MUST 仍遵循：零云→local；一云→该云；多云→冲突失败。文件列表、检索与权限判断 MUST 继续以 `sys_file`（及数据权限）为准，MUST NOT 以云桶列举替代业务目录。

#### Scenario: 唯一云插件下文件中心直传写入该云

- **WHEN** 恰好一个云 storage provider 可服务且支持直连 put
- **AND** 用户完成文件中心直传 init→上传→complete
- **THEN** 对象字节 MUST 位于该云后端（provider key 使用 `files/` 前缀隔离语义）
- **AND** `sys_file` MUST 在 complete 成功后可查询

#### Scenario: 多云冲突时直传 init 失败

- **WHEN** 两个或以上云 storage provider 同时可服务
- **AND** 用户请求文件中心直传 init
- **THEN** init MUST 失败并返回可诊断的存储冲突或存储失败错误
- **AND** MUST NOT 签发任一云的写访问

#### Scenario: 无云插件时文件中心保持中转上传

- **WHEN** 没有任何可服务云 storage provider
- **AND** 用户上传文件
- **THEN** 系统 MUST 使用 local 后端与服务端中转（multipart 或 proxy mode）
- **AND** MUST NOT 要求浏览器直连宿主本地磁盘路径
