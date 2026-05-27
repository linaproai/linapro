## MODIFIED Requirements

### Requirement:动态插件运行时派生缓存必须跨节点失效

动态插件安装、启用、禁用、卸载、升级或同版本刷新后，系统 SHALL 使用统一缓存协调机制使所有节点上的插件运行时派生缓存失效或刷新。派生缓存范围 SHALL 包含插件前端包、运行时 i18n 包、Wasm 编译缓存、插件 manifest 资源视图和动态 artifact 中的插件默认配置视图。

#### Scenario:非主节点观察到插件运行时修订号变更

- **WHEN** 集群模式下主节点完成动态插件运行时状态转换并发布插件运行时缓存修订号
- **THEN** 非主节点在下一个请求路径或监听路径上观察到新修订号
- **AND** 非主节点刷新插件启用快照
- **AND** 非主节点使插件前端包、运行时 i18n 包、Wasm 编译缓存、manifest 资源视图和 artifact 默认配置视图失效

#### Scenario:插件禁用后非主节点不再暴露能力

- **WHEN** 主节点上动态插件被禁用或卸载
- **THEN** 非主节点不得在插件运行时缓存域允许的陈旧窗口之外继续从过期本地缓存暴露该插件的菜单、前端资产、manifest 资源、默认配置或动态路由能力

### Requirement:动态插件产物归档必须支持同版本刷新一致性

系统 SHALL 确保同版本刷新后的活跃发布指向可验证的新产物内容，并且其他节点可使用共享发布状态判断本地缓存是否过期。动态插件 manifest 资源视图和 artifact 默认配置视图 MUST 绑定当前活跃发布的产物校验和或 generation，不得仅通过可变产物路径决定复用。

#### Scenario:同版本刷新写入新产物

- **WHEN** 插件同版本刷新提交新产物内容
- **THEN** 系统更新活跃发布的校验和或 generation
- **AND** 发布插件运行时缓存修订号
- **AND** 其他节点可使用活跃发布的校验和或 generation 判断本地 manifest 资源视图、artifact 默认配置视图和 Wasm 编译缓存是否需要重建

#### Scenario:旧产物清理不影响当前活跃发布

- **WHEN** 系统清理旧动态插件产物
- **THEN** 当前活跃发布引用的产物不得被删除
- **AND** 仍被本地缓存引用但不再活跃的产物可根据保留策略稍后清理

### Requirement: 插件运行时变更必须发布 Redis coordination event

系统 SHALL 在集群模式下为插件安装、启用、禁用、卸载、升级、active release 切换、动态插件 artifact 变化发布`plugin-runtime`Redis revision 和 event。该 revision/event MUST 覆盖动态插件 manifest 资源视图和 artifact 默认配置视图的失效。

#### Scenario: 动态插件启用后其他节点刷新

- **WHEN** 主节点启用动态插件 P
- **THEN** 系统发布`plugin-runtime`Redis revision
- **AND** 其他节点收到 event 后刷新 enabled snapshot
- **AND** 其他节点可路由到插件 P 的 active release
- **AND** 其他节点后续读取插件 P 的 manifest 资源或 artifact 默认配置时使用 active release 的最新资源视图

#### Scenario: 动态插件禁用后其他节点隐藏

- **WHEN** 主节点禁用动态插件 P
- **THEN** 系统发布`plugin-runtime`Redis revision
- **AND** 其他节点失效 frontend bundle、runtime i18n、Wasm、manifest 资源视图和 artifact 默认配置视图派生缓存
- **AND** 后续访问插件 P 路由或插件 P 的 host service 资源读取返回不可用或拒绝

### Requirement: 插件派生缓存失效必须按 scope 精细化

系统 SHALL 在插件运行时变更时按插件 ID、sector、locale 或 global scope 精细失效 frontend bundle、runtime i18n、Wasm 缓存、manifest 资源视图和 artifact 默认配置视图。普通路径不得无理由清空所有插件所有派生缓存。

#### Scenario: 单插件 frontend bundle 失效

- **WHEN** 动态插件 P 上传新 frontend bundle
- **THEN** 系统仅失效插件 P 相关 frontend bundle cache
- **AND** 其他插件 bundle cache 保持可用

#### Scenario: 单插件 manifest 资源失效

- **WHEN** 动态插件 P 上传或刷新包含新`manifest/metadata.yaml`的 artifact
- **THEN** 系统仅失效插件 P 相关 manifest 资源视图和 artifact 默认配置视图
- **AND** 其他插件的 manifest 资源缓存保持可用

## ADDED Requirements

### Requirement:动态插件运行时资源视图必须包含配置和通用 manifest 资源

系统 SHALL 在动态插件运行时加载时从 active release artifact 构建插件资源视图。该视图 MUST 包含 artifact 实际携带的`manifest/config/config.yaml`、`manifest/config/config.example.yaml`以及`manifest/**/*.yaml`中属于插件声明型资源的文件，并保持与源码插件目录一致的路径语义；未提供`manifest/metadata.yaml`的插件不得被要求提交占位文件。

#### Scenario:动态插件加载 metadata 资源

- **WHEN** 动态插件 active release artifact 携带`manifest/metadata.yaml`
- **THEN** 运行时资源视图包含相对路径`metadata.yaml`
- **AND** 插件可通过`manifest.get`或`HostServices.Manifest()`读取该资源

#### Scenario:动态插件加载默认配置资源

- **WHEN** 动态插件 active release artifact 携带`manifest/config/config.yaml`
- **THEN** 运行时资源视图记录该 artifact 默认配置来源
- **AND** 插件配置 resolver 可在不存在生产外部配置时读取该默认配置

#### Scenario:资源视图不暴露宿主路径

- **WHEN** 动态插件读取 artifact 中的配置或 manifest 资源
- **THEN** 系统使用 artifact 资源视图返回内容
- **AND** 响应不得暴露宿主本地 artifact 存储绝对路径作为插件可用资源路径
