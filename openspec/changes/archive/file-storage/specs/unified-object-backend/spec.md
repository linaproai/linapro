# unified-object-backend Specification

## Purpose

定义全站对象字节后端选择规则，以及宿主文件中心内容读写对统一后端的接入要求：与插件 `Storage()` 共享 `ResolveProvider`，使用 `files/` 键前缀隔离，并在过渡期对历史本地对象提供云未命中回退。

## Requirements

### Requirement: 文件中心内容操作必须使用统一对象后端选择

系统 SHALL 使宿主文件中心的对象内容写入、读取与删除通过与插件 `Storage()` 相同的后端选择规则完成：零个可服务云 storage provider 插件时使用内置 local；恰好一个可服务云插件时使用该插件；两个及以上可服务云插件时 MUST 返回冲突错误且 MUST NOT 静默选择或回退。

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

### Requirement: 文件中心对象键必须使用 files 前缀

系统 SHALL 将文件中心相对路径 `P` 映射为 provider key `files/P`，以与插件私有对象键空间隔离。本地 provider 对 `files/...` 键 MUST 路由到 `NamespaceFiles` 并去掉前缀后的相对 key，MUST NOT 附加 `.capability-storage` 前缀。

#### Scenario: 上传内容使用 files 前缀键

- **WHEN** 文件中心上传写入相对路径 `42/2026/07/demo.png`
- **THEN** 对象后端使用的 provider key MUST 为 `files/42/2026/07/demo.png`

#### Scenario: 本地 provider 路由 files 键到 NamespaceFiles

- **WHEN** 活动后端为 local
- **AND** provider key 以 `files/` 开头
- **THEN** 本地 provider MUST 将该键写入 `NamespaceFiles` 下的去前缀相对路径
- **AND** MUST NOT 使用插件 `.capability-storage` 命名空间布局

### Requirement: 新写入必须记录生效存储引擎

系统 SHALL 在文件中心新写入成功时，将 `sys_file.engine` 设置为当前生效 provider id（`local` 或云插件 id）。列表与下载路由 MUST NOT 仅依赖 `engine` 字段；读取 MUST 以 `ResolveProvider` 结果为准，并在适用时执行本地回退。

#### Scenario: 本地写入记录 engine 为 local

- **WHEN** 当前 Resolve 结果为 local
- **AND** 文件中心成功上传新文件
- **THEN** `sys_file.engine` MUST 为 `local`

#### Scenario: 云写入记录 engine 为插件 id

- **WHEN** 当前 Resolve 结果为某可服务云 storage provider 插件
- **AND** 文件中心成功上传新文件
- **THEN** `sys_file.engine` MUST 为该插件 id

### Requirement: 删除必须优先活动后端并尽力清理本地残留

系统 SHALL 在文件中心物理删除时先对活动后端执行删除；若活动后端非 local，系统 SHOULD best-effort 再删除 local 上同 key 对象，以避免切换后端后的本地残留。

#### Scenario: 云活动后端删除后清理本地同键

- **WHEN** 当前活动后端为云 provider
- **AND** 用户删除文件中心记录并触发物理删除
- **THEN** 系统 MUST 删除云端对应 `files/` 键对象
- **AND** 系统 SHOULD 尝试删除本地同 key 对象（若存在）
