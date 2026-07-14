## Requirements

### Requirement: Files 领域必须只表示宿主文件中心投影

系统 SHALL 将 `Files()` 领域能力限定为宿主文件中心资源投影和可见性校验能力。`Files()` 方法 MUST 基于宿主文件领域 owner 的数据权限、租户边界和存在性不泄露策略返回文件投影或执行可见性确认，不得承担插件私有对象存储的内容读写生命周期。

#### Scenario: 插件不得通过 Files 读写私有对象内容

- **WHEN** 插件需要保存或读取插件私有二进制对象
- **THEN** 调用方 MUST 使用 `Storage()` 领域能力
- **AND** MUST NOT 通过 `Files()` 伪装为宿主文件中心资源完成内容生命周期

### Requirement: 插件私有附件不使用 Files

插件需要保存、下载、删除、列出或清理插件私有附件对象时，必须使用 `Storage()` 领域能力，不得通过 `Files()` 领域方法把插件私有对象伪装为宿主文件中心资源。

#### Scenario: 源码插件示例附件走 Storage

- **WHEN** `linapro-demo-source` 保存、下载、替换、删除或卸载清理附件
- **THEN** 实现 MUST 通过 `storagecap.Service` / `pluginhost.Services.Storage()` 完成
- **AND** MUST NOT 直接读取 `upload.path` 拼接宿主本地物理路径
