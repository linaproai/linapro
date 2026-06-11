# pluginbridge-subcomponent-architecture 规范增量

## ADDED Requirements

### Requirement: pluginbridge 必须以公开子组件承载动态插件专属 record store SDK

系统 SHALL 将动态插件 record store guest SDK 作为`pkg/plugin/pluginbridge/recordstore`公开子组件发布。该子组件承载 ORM-style record store facade、typed query plan 契约及其`internal/plan`私有实现；它 MAY import `pluginbridge/protocol`完成 host-service data 协议的请求构造与响应解码，该方向属于`pluginbridge`同层内聚的合法依赖。宿主侧 data governance 适配入口消费的 typed plan 契约 MUST 通过该公开子组件访问，不得直接 import 其`internal/plan`实现。

#### Scenario: 动态插件构造受治理记录操作

- **WHEN** 动态插件 guest 代码需要构造受治理表记录查询、变更或事务
- **THEN** 插件 import `lina-core/pkg/plugin/pluginbridge/recordstore`
- **AND** guest 能力目录`Services.RecordStore()`返回该子组件的 facade

#### Scenario: recordstore 依赖 protocol 属于同层合法方向

- **WHEN** 检查`pluginbridge/recordstore`的 import
- **THEN** 它可以 import `pluginbridge/protocol`与`capability`公共原语
- **AND** 它不得 import `pluginbridge`根包、guest runtime 或 route helper
- **AND** `capability/**`非测试代码不得反向 import 该子组件

#### Scenario: 宿主 datahost 消费 typed plan 契约

- **WHEN** 宿主 datahost 需要解码 typed query plan 并执行受治理表操作
- **THEN** 宿主 import `lina-core/pkg/plugin/pluginbridge/recordstore`公开契约
- **AND** 不得直接 import `pluginbridge/recordstore/internal/plan`

### Requirement: record store SDK 迁移必须保持协议与计划语义不变

系统 SHALL 将 record store SDK 从`pkg/plugin/capability/recordstore`迁移到`pkg/plugin/pluginbridge/recordstore`视为纯包路径重构。typed query plan 的 JSON 编码、host service data 协议的 service/method 字符串、payload 编解码结果、过滤与排序校验语义、事务操作语义 MUST 保持不变；旧路径 MUST 删除且不保留兼容入口。

#### Scenario: 迁移后 plan 编解码等价

- **WHEN** 使用迁移后的`pluginbridge/recordstore`编码并解码 typed query plan
- **THEN** round trip 结果与迁移前等价
- **AND** 既有 recordstore 测试随包迁移后继续通过

#### Scenario: 旧 capability 路径不可用

- **WHEN** 静态检索生产 Go 代码、动态插件样例或测试替身的 import
- **THEN** 不存在`lina-core/pkg/plugin/capability/recordstore`引用
- **AND** 调用方全部使用`lina-core/pkg/plugin/pluginbridge/recordstore`

#### Scenario: 动态插件样例迁移后可构建

- **WHEN** 对动态插件样例执行普通 Go 测试和`GOOS=wasip1 GOARCH=wasm`构建
- **THEN** 样例必须通过编译
- **AND** record store guest 执行路径在 wasip1 构建中可用
