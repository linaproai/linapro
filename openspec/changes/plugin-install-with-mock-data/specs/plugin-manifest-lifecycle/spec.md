## ADDED Requirements

### Requirement: 插件 manifest SQL 资源必须能识别 mock-data 目录作为独立的资产分类

宿主在扫描插件 `manifest/sql/` 目录时 SHALL 识别 install / uninstall / mock-data 三类 SQL 资产，互不重叠：

- `manifest/sql/*.sql`（不递归）属于 install 方向，命名匹配 `^\d{3}-[a-z0-9-]+\.sql$`。
- `manifest/sql/uninstall/*.sql` 属于 uninstall 方向，命名规则同上。
- `manifest/sql/mock-data/*.sql` 属于 mock 方向，命名规则同上；这些文件 MUST NOT 进入 install 或 uninstall 方向的资产清单。

`pluginfs.DiscoverSQLPathsFromFS` 与 `catalog.serviceImpl.ResolvePluginSQLAssets` SHALL 提供独立的 mock-data 扫描入口（如 `MigrationDirectionMock`），不与 install/uninstall 入口共享文件清单。源码插件（embed FS）与动态插件（制品 FS）必须使用相同的扫描逻辑。

#### Scenario: 安装方向的资产清单不包含 mock-data 文件
- **WHEN** 宿主对包含 `manifest/sql/001-schema.sql` 与 `manifest/sql/mock-data/001-mock.sql` 的插件调用 `ResolvePluginSQLAssets(manifest, MigrationDirectionInstall)`
- **THEN** 返回的资产清单只包含 `001-schema.sql`
- **AND** 不包含 `mock-data/001-mock.sql` 或其任何变体

#### Scenario: Mock 方向独立扫描返回 mock-data 文件
- **WHEN** 宿主对相同的插件调用 `ResolvePluginSQLAssets(manifest, MigrationDirectionMock)`
- **THEN** 返回的资产清单只包含 `mock-data/001-mock.sql`
- **AND** 文件按文件名升序排列

### Requirement: 动态插件制品打包必须沿用 mock-data 目录约定

动态插件在打包阶段 SHALL 将 `manifest/sql/mock-data/` 目录原样保留到制品 FS 中，与源码插件 `embed FS` 的目录布局保持一致；动态插件运行时加载链路 MUST 通过共享的扫描方法读取该目录。打包工具与制品 schema MUST NOT 引入与源码插件不同的 mock-data 路径或额外的清单字段。

#### Scenario: 动态插件升级保留 mock-data 资源可见性
- **WHEN** 一个动态插件在新版本中新增或修改 `manifest/sql/mock-data/*.sql`
- **AND** 宿主完成对该插件的升级与制品替换
- **THEN** 升级后宿主对该插件的 mock-data 扫描结果反映新版本内容
- **AND** 该 mock-data 目录在制品的 FS 视图中始终可见
