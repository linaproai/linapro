## ADDED Requirements

### Requirement: 动态插件请求热路径不得重复解析同一产物

系统 SHALL 在动态插件 HTTP 请求热路径中复用同一次请求已经解析或命中的 manifest 快照。路由匹配、运行时准备和升级状态构建 MUST 不在同一请求内重复执行完整 artifact 解析或重复反序列化同一 release manifest。

#### Scenario: 路由匹配与运行时准备复用 manifest

- **WHEN** 动态插件请求先完成路由匹配再进入运行时准备阶段
- **THEN** 运行时准备阶段复用路由匹配阶段已获取的 manifest 快照
- **AND** 缓存命中时不再次读取和解析同一 artifact

#### Scenario: Manifest 缓存不替代执行 freshness 检查

- **WHEN** 动态插件请求即将进入 guest 执行
- **THEN** 系统必须确认`plugin-runtime`freshness、enabled snapshot 和 active release 状态
- **AND** freshness 不可确认且超过允许陈旧窗口时必须拒绝执行或 conservative-hide

### Requirement: Wasm 编译缓存必须精细失效且避免全局编译阻塞

系统 SHALL 将动态插件`WASM`编译缓存绑定当前 active release 的 artifact 身份、校验和或 generation，并在已知单插件变更时按插件或 artifact 路径失效。系统 MUST 不因单插件启用、禁用、升级或刷新无条件丢弃全部插件的编译缓存。编译过程 MUST 不持有阻塞其他 artifact 读取的全局缓存写锁。

#### Scenario: 单插件变更只失效目标 artifact

- **WHEN** 插件`P`的 active release 变化且系统可以解析目标 artifact 路径或校验和
- **THEN** 系统只失效插件`P`旧 active artifact 对应的编译缓存
- **AND** 其他插件已编译模块保持可用

#### Scenario: 同 artifact 并发只编译一次

- **WHEN** 多个请求同时命中同一个未编译 artifact
- **THEN** 系统通过 per-artifact single-flight 只执行一次文件读取和编译
- **AND** 其他请求等待同一编译结果或收到同一编译错误

#### Scenario: 不同 artifact 编译互不阻塞

- **WHEN** artifact`A`正在编译且 artifact`B`已有可用缓存或需要独立编译
- **THEN** artifact`B`的请求不得被 artifact`A`的全局缓存写锁阻塞
- **AND** 缓存 map 锁只保护条目读写和状态切换

### Requirement: 插件运行时派生缓存对账必须有界

系统 SHALL 在集群 peer 观察到`plugin-runtime`修订号变化时执行有界差异对账。对账 MUST 基于当前 active release 集合、artifact 身份和文件状态判断需要失效的 manifest、OpenAPI 投影和`WASM`编译缓存。系统 MUST 不在普通单插件变更回调中无条件清空所有插件所有派生缓存。

#### Scenario: Peer 通过差异对账回收旧 artifact

- **WHEN** 集群节点观察到`plugin-runtime`修订号前进
- **THEN** 节点读取当前 active artifact 集合并与本地缓存条目对比
- **AND** 仅失效不再活跃、校验和变化或文件状态变化的缓存条目

#### Scenario: 全局变更保留全量失效语义

- **WHEN** 变更事件确认为全局运行时配置变化且无法定位受影响插件
- **THEN** 系统可以执行全量运行时派生缓存失效
- **AND** 审查记录必须说明该全量失效的权威原因和触发频率边界
