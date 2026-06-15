## ADDED Requirements

### Requirement: 插件治理读投影必须共享清单快照和依赖索引

系统 SHALL 为插件管理列表、详情、依赖检查、OpenAPI 文档投影和 hook 分发复用同一批 manifest、store、runtime 和 dependency 快照。读取路径 MUST 避免为每个插件、每个依赖检查或每个文档请求重复全量扫描和重复 artifact 解析。

#### Scenario: 管理列表和详情复用统一投影入口

- **WHEN** 系统构建插件管理摘要列表或单插件详情
- **THEN** 系统通过统一投影构建入口读取清单和治理投影
- **AND** 单插件详情不得通过全量扫描全部插件来定位目标插件

#### Scenario: 反向依赖检查使用索引

- **WHEN** 系统为插件列表或卸载校验计算反向依赖
- **THEN** 系统基于当前快照集一次性构建`pluginID`到下游依赖方的索引
- **AND** 单次反向依赖查询不得遍历全部插件快照

#### Scenario: DependencyCheck 请求共享一次快照

- **WHEN** 调用方请求插件依赖检查
- **THEN** 依赖解析、安装快照投影和反向依赖投影共享同一个 request 级 manifest 快照
- **AND** 单次请求不得触发多次完整`ScanManifests`

### Requirement: OpenAPI 插件投影缓存必须绑定运行时和语言版本

系统 SHALL 将动态插件 OpenAPI 文档投影视为插件运行时展示派生缓存。缓存键 MUST 包含`plugin-runtime`修订号、当前 locale 和运行时翻译包版本。插件启停、升级、动态 artifact 刷新或翻译包变化后，后续 OpenAPI 请求 MUST 重建受影响投影。

#### Scenario: Runtime revision 变化后重建 OpenAPI 投影

- **WHEN** 插件`P`升级或禁用并发布新的`plugin-runtime`修订号
- **THEN** 后续 OpenAPI 文档请求不得继续返回旧修订号下的插件`P`路由投影
- **AND** 系统重建或失效对应缓存条目

#### Scenario: Locale 隔离 OpenAPI 投影

- **WHEN** 两个请求使用不同 locale 访问 OpenAPI 文档
- **THEN** 系统使用不同缓存键读取或构建文档投影
- **AND** 不得把旧语言版本的插件文档投影返回给当前请求
