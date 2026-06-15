## ADDED Requirements

### Requirement: 插件清单读取必须复用有界读模型缓存

系统 SHALL 为插件清单事实源提供内部读模型缓存。源码插件 manifest、动态插件 desired manifest、release manifest 与 release YAML 快照 MUST 按各自权威数据源和不可变性分开缓存。缓存读取路径 MUST 不写入插件治理表，不触发菜单、权限、资源引用、hook 或运行时状态副作用。

#### Scenario: 动态插件稳态扫描复用文件状态守卫

- **WHEN** 系统扫描动态插件 artifact 目录且某 artifact 的路径、大小和修改时间均未变化
- **THEN** 系统复用已缓存的 manifest 解析结果
- **AND** 不重新整文件读取、哈希和解析该 artifact

#### Scenario: 单插件读取不全量扫描

- **WHEN** 调用方按`pluginID`读取单个插件清单且缓存索引已包含该插件 artifact 路径
- **THEN** 系统只读取或校验该插件对应 artifact 的缓存条目
- **AND** 不为了单插件详情查询扫描全部插件 artifact

#### Scenario: Release 快照按校验和复用

- **WHEN** 调用方读取同一`pluginID`、`releaseID`和`checksum`对应的 release manifest
- **THEN** 系统复用已有解析快照
- **AND** 不重复反序列化同一 release YAML 内容

#### Scenario: 清单读取保持无副作用

- **WHEN** 插件列表、详情、依赖检查、OpenAPI 投影或 hook 分发读取清单缓存
- **THEN** 读取过程不得写入`sys_plugin`、`sys_plugin_release`、`sys_plugin_resource_ref`、`sys_menu`或`sys_role_menu`
- **AND** 所有治理同步仍由显式同步、安装、升级、启用、禁用或启动同步编排入口触发

### Requirement: 插件清单缓存必须支持按插件显式失效

系统 SHALL 在插件同步、动态包上传、安装、卸载、启用、禁用、升级、active release 切换或源码插件同步成功后，按`pluginID`或 artifact 路径失效对应清单缓存。全局失效只允许用于全局配置或无法确定影响范围的治理事件。

#### Scenario: 单插件升级只失效目标插件清单

- **WHEN** 动态插件`P`升级到新的 active release
- **THEN** 系统失效插件`P`的 desired manifest、release manifest 和相关索引条目
- **AND** 其他插件清单缓存保持可用

#### Scenario: 卸载插件清理 release 快照

- **WHEN** 插件`P`被卸载且其 release 不再作为当前权威发布使用
- **THEN** 系统清理或淘汰插件`P`相关 release manifest 缓存
- **AND** 后续读取插件`P`必须回源确认其治理状态
