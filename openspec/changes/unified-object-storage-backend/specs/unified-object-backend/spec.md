## ADDED Requirements

### Requirement: 文件中心内容操作必须使用统一对象后端选择

系统 SHALL 使宿主文件中心的对象内容写入、读取与删除通过与插件 `Storage()` 相同的后端选择规则完成：零个可服务云 storage provider 插件时使用内置 local；恰好一个时可服务云插件时使用该插件；两个及以上可服务云插件时 MUST 返回冲突错误且 MUST NOT 静默选择或回退。

文件中心的列表、检索与元数据浏览 MUST 继续基于 `sys_file`（及数据权限），MUST NOT 以云桶全量 List 作为业务目录来源。

#### Scenario: 无云插件时文件上传写入本地

- **WHEN** 没有任何可服务云 storage provider 插件
- **AND** 用户在文件中心上传文件
- **THEN** 对象内容 MUST 写入本地 files 存储
- **AND** `sys_file` 元数据 MUST 正常创建

#### Scenario: 唯一云插件启用时文件上传写入云

- **WHEN** 恰好一个云 storage provider 插件可服务且配置可用于写入
- **AND** 用户在文件中心上传文件
- **THEN** 对象内容 MUST 写入该云 provider（provider key 使用 `files/` 前缀隔离）
- **AND** 列表仍从 `sys_file` 返回

#### Scenario: 多云插件冲突时文件上传失败

- **WHEN** 两个或以上云 storage provider 插件同时可服务
- **AND** 用户在文件中心上传文件
- **THEN** 上传 MUST 失败并返回可诊断的存储冲突或存储失败错误
- **AND** MUST NOT 静默写入本地或任一云后端

#### Scenario: 历史本地文件在启用云后仍可下载

- **WHEN** 文件中心记录对应对象仅存在于本地 files 存储
- **AND** 当前活动后端为云 provider
- **AND** 用户下载该文件
- **THEN** 系统 MUST 在云端未命中后回退读取本地对象（过渡兼容）
