## Requirements

### Requirement: 插件私有文件对象必须归属 Storage 领域

系统 SHALL 将插件自有附件、插件业务二进制对象、导入导出临时对象和插件卸载清理对象归属 `Storage()` 领域能力。源码插件和动态插件 MUST 通过 `storagecap.Service` 管理这些对象，不得通过宿主文件中心 `Files()` 领域或宿主本地物理路径管理。

#### Scenario: 业务插件写入私有对象

- **WHEN** 业务插件保存插件私有附件或临时导出对象
- **THEN** 调用 MUST 走 `storagecap.Service`
- **AND** MUST NOT 写入 `sys_file` 或直接操作宿主上传物理目录

### Requirement: Storage 和 Files 领域边界必须保持独立

系统 SHALL 保持 `Storage()` 和 `Files()` 两个领域能力的职责独立。`Storage()` 拥有插件对象内容生命周期，`Files()` 拥有宿主文件中心资源投影和可见性校验。任一领域的公开契约 MUST NOT 混入另一个领域的内部标识、存储模型或生命周期职责。

#### Scenario: 契约不交叉暴露内部模型

- **WHEN** 插件通过 `Storage()` 或 `Files()` 调用宿主能力
- **THEN** `Storage()` 响应 MUST NOT 暴露 `sys_file` 主键或文件中心内部路径模型作为对象身份
- **AND** `Files()` 响应 MUST NOT 暴露 `storagecap` provider object key 或插件私有 scoped key

### Requirement: 动态插件 Storage Put 必须支持有界内存分片上传

系统 SHALL 允许动态插件通过 `Storage().Put` 写入大文件或未知大小输入，并在 guest SDK 内部按输入大小选择单次 `storage.put` 或分片上传。分片上传 MUST 使用 `put.init`、`put.chunk`、`put.commit` 和 `put.abort` host service 方法完成传输。系统 MUST NOT 对最终对象大小设置动态 host service 固定上限。

#### Scenario: 大输入自动分片

- **WHEN** 动态插件调用 `Storage().Put` 且输入超过 guest 单次上传阈值
- **THEN** guest SDK MUST 使用 `put.init` / `put.chunk` / `put.commit` 完成写入
- **AND** 宿主 commit 时 MUST 通过 `storagecap.Service.Put` 流式写入最终 logical path

#### Scenario: 小输入保持单次 put

- **WHEN** 动态插件调用 `Storage().Put` 且输入适合单次传输
- **THEN** guest SDK MAY 直接使用 `storage.put`
- **AND** 行为与路径授权语义 MUST 与分片路径一致

### Requirement: 动态插件 Storage 分片上传必须保持路径授权和会话绑定

系统 SHALL 对 `put.init`、`put.chunk`、`put.commit` 和 `put.abort` 执行与 `storage.put` 等价的 service、method 和 `storage.resources.paths` 授权校验。授权 path MUST 匹配最终插件 logical path。宿主 MUST 将 upload ID 绑定到当前插件 ID、最终 logical path 和上传会话状态。

#### Scenario: 未授权 path 被拒绝

- **WHEN** 动态插件对未授权 logical path 调用 `put.init` 或 `storage.put`
- **THEN** 宿主 MUST 拒绝该调用
- **AND** MUST NOT 创建可提交的上传会话或写入对象

#### Scenario: 跨会话或跨 path 提交被拒绝

- **WHEN** 调用方使用与会话绑定不一致的 upload ID、插件 ID 或 logical path 执行 `put.chunk` 或 `put.commit`
- **THEN** 宿主 MUST 拒绝该操作

### Requirement: 插件存储必须提供 provider 扩展机制

系统 SHALL 定义 `storagecap.Provider` 和 `storagecap.ProviderFactory`，允许主框架和源码插件提供对象存储后端。Provider MUST 只负责根据 provider object key 执行对象读写、删除、列表和元数据读取，不得接收或解释动态插件 `hostServices` 授权快照。源码插件可以通过 `storagecap.Provide(pluginID, factory)` 注册 COS、OSS、S3、MinIO 兼容端或其他对象存储 provider。主框架内置本地 provider MUST 复用宿主内部 `storage.Service` 执行本地对象读写，不得维护另一套独立的本地文件系统实现。

#### Scenario: 主框架注册默认本地 provider

- **WHEN** 当前没有可服务的 storage provider 插件
- **THEN** 系统使用主框架内置本地磁盘 provider
- **AND** 插件存储能力可在单机模式下正常读写对象
- **AND** 内置本地 provider 通过宿主内部 `storage.Service` 执行本地读写

#### Scenario: 源码插件注册存储 provider

- **WHEN** 源码插件调用 `storagecap.Provide(pluginID, factory)` 注册存储 provider
- **THEN** 系统记录该 provider 工厂
- **AND** 仅当该插件处于平台可服务状态，且为当前唯一可服务 storage provider 插件时，才承接对象存储调用

#### Scenario: Provider 不接收动态授权信息

- **WHEN** 动态插件调用 storage 并通过授权校验
- **THEN** `storagecap.Service` 向 provider 传入 provider object key 和对象操作参数
- **AND** provider 不得接收 `hostServices` 授权快照、授权 path 列表或动态插件原始 envelope

### Requirement: Storage provider 运行时必须以唯一可服务插件自动选中

系统 SHALL 以「唯一可服务已注册 storage provider 插件」作为 active 后端选择策略：0 个可服务插件时使用内置 local；恰好 1 个时使用该插件；2 个及以上时对象操作 MUST 失败并返回 `CodeStorageProviderConflict`（或等价稳定业务码）。系统 MUST NOT 通过宿主主配置 active provider plugin ID 选择后端，MUST NOT 在多插件可服务时静默挑选其一，MUST NOT 在冲突时静默回退 local。

#### Scenario: 未启用任何云插件时回退本地

- **WHEN** 全部云 storage provider 插件均未处于可服务状态
- **THEN** `storagecap.Service` MUST 使用宿主内置 local provider
- **AND** 不得因未安装云插件而拒绝 Storage 调用

#### Scenario: 同时启用多个云插件时冲突

- **WHEN** 两个或以上 storage provider 插件同时可服务
- **THEN** `storagecap.Service` 对象操作 MUST 失败并返回冲突稳定业务码
- **AND** MUST NOT 静默选择其中一个 provider
- **AND** MUST NOT 静默回退 local

### Requirement: 官方云 storage provider 插件必须可交付并接入管理目录

系统 SHALL 提供官方源码插件实现主流云对象存储 provider，并使其管理配置页挂载到宿主 `setting`（系统设置）稳定目录。官方交付范围至少包括腾讯云 COS、阿里云 OSS、华为云 OBS、七牛云 Kodo、AWS S3、Azure Blob 厂商插件，以及 S3 兼容协议插件。插件 MUST 通过 `storagecap.Provide` 注册，MUST NOT 改变插件可见 `storagecap.Service` 契约，MUST NOT 将云 SDK 依赖引入 `lina-core` 宿主模块路径。

#### Scenario: 安装云插件后出现配置入口

- **WHEN** 管理员安装并同步 `linapro-storage-oss`（或 cos / obs / qiniu / aws / azure / s3）
- **THEN** 「系统设置」目录下 MUST 出现对应配置菜单
- **AND** 业务插件调用 `Storage()` 的代码路径 MUST 无需修改

#### Scenario: 唯一云插件启用后承接写入

- **WHEN** 仅一个官方云 storage provider 插件可服务且配置有效
- **AND** 业务插件调用 `Storage().Put`
- **THEN** 对象 MUST 写入该云后端
- **AND** 响应 MUST 仍只暴露 logical path 元数据

### Requirement: 唯一可服务云 provider 配置无效时不得回退本地

当恰好一个 storage provider 插件可服务但其配置缺失或无效时，系统 SHALL 使 Storage 对象操作失败并返回可诊断错误，MUST NOT 静默回退内置 local provider。

#### Scenario: 云插件启用但密钥缺失

- **WHEN** 唯一可服务 provider 为某一云插件
- **AND** 其密钥或 bucket 未配置
- **AND** 调用方执行 `Storage().Put` 或 `Get`
- **THEN** 调用 MUST 失败
- **AND** MUST NOT 将对象写入本地磁盘 provider

### Requirement: 业务上传路径在 Storage provider 冲突时必须 fail-closed

当 `storagecap` 解析结果为 provider 冲突或唯一云 provider 不可用时，依赖插件对象存储的写入路径（含文件管理等内容上传若复用同一 active provider 语义）MUST 将失败暴露给调用方，MUST NOT 因内容哈希复用跳过实际 Put 而表现为上传成功。

#### Scenario: 多云冲突时哈希复用不得伪装成功

- **WHEN** 当前存在 `CodeStorageProviderConflict`
- **AND** 调用方上传内容与已有对象哈希相同
- **THEN** 上传 MUST 失败
- **AND** MUST NOT 仅因跳过 Put 而返回成功
