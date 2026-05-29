## MODIFIED Requirements

### Requirement:动态插件运行时资源视图必须包含配置和通用 manifest 资源

系统 SHALL 在动态插件运行时加载时从 active release artifact 构建插件资源视图。该视图 MUST 包含 artifact 实际携带的`manifest/`目录下所有可投影文件，并保持与源码插件目录一致的路径语义；`manifest.get`返回路径 MUST 相对`manifest/`根目录。该视图 MUST 包含`manifest/config/`、`manifest/sql/`、`manifest/i18n/`和插件自定义资源目录中的文件原文。未提供`manifest/metadata.yaml`的插件不得被要求提交占位文件。

#### Scenario:动态插件加载 metadata 普通资源

- **WHEN** 动态插件 active release artifact 携带`manifest/metadata.yaml`
- **THEN** 运行时资源视图包含相对路径`metadata.yaml`
- **AND** 插件可通过`manifest.get`或`HostServices.Manifest()`读取该资源
- **AND** 系统不需要独立的 `Metadata` 资源视图

#### Scenario:动态插件加载默认配置资源

- **WHEN** 动态插件 active release artifact 携带`manifest/config/config.yaml`
- **THEN** 运行时资源视图包含相对路径`config/config.yaml`
- **AND** 插件配置 resolver 可在不存在生产外部配置时读取该默认配置
- **AND** 已授权的插件也可通过`manifest.get`读取该文件原文

#### Scenario:动态插件加载 SQL 和 i18n 资源原文

- **WHEN** 动态插件 active release artifact 携带`manifest/sql/001-schema.sql`和`manifest/i18n/zh-CN/plugin.json`
- **THEN** 运行时资源视图包含相对路径`sql/001-schema.sql`和`i18n/zh-CN/plugin.json`
- **AND** 已授权的插件可通过`manifest.get`读取这些文件原文
- **AND** 该资源视图不执行 SQL、不加载翻译资源

#### Scenario:资源视图不暴露宿主路径

- **WHEN** 动态插件读取 artifact 中的配置、SQL、i18n 或其他 manifest 资源
- **THEN** 系统使用 artifact 资源视图返回内容
- **AND** 响应不得暴露宿主本地 artifact 存储绝对路径作为插件可用资源路径

#### Scenario:同版本刷新后使用最新 manifest 资源

- **WHEN** 动态插件以相同版本刷新 artifact
- **AND** 新 artifact 修改了`manifest/i18n/zh-CN/plugin.json`或`manifest/sql/001-schema.sql`
- **THEN** 系统按 active release checksum 或 generation 重建 manifest 资源视图
- **AND** 后续`manifest.get`读取返回新 active release 中的资源内容
