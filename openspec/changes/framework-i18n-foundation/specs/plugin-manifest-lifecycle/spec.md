## ADDED Requirements

### Requirement: 插件必须支持随版本交付国际化资源并纳入生命周期管理
The system SHALL allow plugins to deliver locale resources through a standard plugin resource directory, and the host SHALL manage those resources together with plugin discovery, installation, upgrade, enablement, disablement, and uninstallation.

#### Scenario: 源码插件同步时注册国际化资源
- **WHEN** 宿主同步发现一个带有标准国际化资源目录的源码插件
- **THEN** 宿主注册该插件可用的语言资源
- **AND** 这些资源可参与菜单、插件名称和插件描述的本地化投影

#### Scenario: 动态插件卸载时移除国际化资源
- **WHEN** 管理员卸载一个已经安装的动态插件
- **THEN** 宿主从运行时翻译聚合结果中移除该插件的国际化资源
- **AND** 该插件相关菜单和元数据不再继续暴露其本地化消息

### Requirement: 插件元数据与插件菜单必须支持按当前语言本地化投影
The system SHALL localize plugin name, plugin description, and plugin-declared menu titles according to the current request language while keeping plugin ID, menu key, route path, and permission semantics unchanged.

#### Scenario: 查看插件列表时返回本地化插件名称
- **WHEN** 管理员以 `en-US` 查看插件管理列表或插件详情
- **THEN** 插件名称和插件描述使用英文本地化结果
- **AND** 插件 ID、版本、状态和治理字段保持原有语义

#### Scenario: 插件菜单按语言返回标题
- **WHEN** 已启用插件声明的菜单存在当前语言翻译资源
- **THEN** 左侧导航、菜单管理页和角色授权树中的插件菜单标题使用该语言结果
- **AND** 不需要插件直接改写 `sys_menu` 中的多语言字段结构
