## ADDED Requirements

### Requirement: 动态插件运行时派生缓存必须跨节点失效

系统 SHALL 在动态插件安装、启用、禁用、卸载、升级或同版本刷新后，通过统一缓存协调机制使所有节点的插件运行时派生缓存失效或刷新。

#### Scenario: 非主节点观察到插件运行时修订号变化
- **WHEN** 集群模式下主节点完成动态插件运行时状态切换并发布插件运行时缓存修订号
- **THEN** 非主节点在下一次请求路径或 watcher 路径观察到新修订号
- **AND** 非主节点刷新插件启用快照
- **AND** 非主节点失效插件前端包、运行时 i18n 包和 Wasm 编译缓存

#### Scenario: 插件禁用后非主节点不继续暴露能力
- **WHEN** 动态插件在主节点被禁用或卸载
- **THEN** 非主节点不得继续基于旧本地缓存暴露该插件菜单、前端资产或动态路由能力超过插件运行时缓存域允许的陈旧窗口

### Requirement: Wasm 编译缓存必须绑定 artifact checksum 或 generation

系统 SHALL 将动态插件 Wasm 编译缓存绑定到当前 active release 的 artifact checksum 或 generation，不能只按可变 artifact 路径判断缓存是否可复用。

#### Scenario: 同版本动态插件刷新后重新编译
- **WHEN** 动态插件以相同版本刷新但 artifact checksum 发生变化
- **THEN** 所有节点在观察到插件运行时修订号变化后不得继续命中旧 checksum 的 Wasm 编译缓存
- **AND** 下一次动态路由或动态任务执行必须基于新 artifact 编译或加载

#### Scenario: artifact 路径相同但 checksum 不同
- **WHEN** active release 的 artifact 路径与旧版本缓存路径相同但 checksum 不同
- **THEN** 系统将其视为不同的编译缓存条目
- **AND** 旧条目必须被失效或自然清理

### Requirement: 动态插件 artifact 归档必须支持同版本刷新一致性

系统 SHALL 保证同版本刷新后的 active release 指向可验证的新 artifact 内容，且其他节点能够基于共享发布状态判断本地缓存是否过期。

#### Scenario: 同版本刷新写入新 artifact
- **WHEN** 插件同版本刷新提交新的 artifact 内容
- **THEN** 系统更新 active release 的 checksum 或 generation
- **AND** 系统发布插件运行时缓存修订号
- **AND** 其他节点可以通过 active release 的 checksum 或 generation 判断本地缓存是否需要重建

#### Scenario: 旧 artifact 清理不影响当前 active release
- **WHEN** 系统清理旧动态插件 artifact
- **THEN** 当前 active release 引用的 artifact 不得被删除
- **AND** 仍被本地缓存引用但已非 active 的 artifact 可以按保留策略延迟清理
