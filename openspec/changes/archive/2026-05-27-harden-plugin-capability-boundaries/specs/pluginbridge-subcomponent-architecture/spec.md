## ADDED Requirements

### Requirement:pluginbridge 根包不得发布业务能力 client

系统 SHALL 将`pkg/plugin/pluginbridge`根包收敛为动态插件 ABI、协议 facade 和 transport 边界。根包 MUST NOT 发布 runtime、storage、data、cache、lock、config、notify、manifest、org 或 tenant 等业务能力 client；这些动态插件业务能力 client MUST 由`pkg/plugin/capability/guest`发布。

#### Scenario:动态插件访问业务能力

- **WHEN** 动态插件 guest 代码需要访问 runtime、storage、data、cache、lock、config、notify、manifest、org 或 tenant 能力
- **THEN** 它从`pkg/plugin/capability/guest`导入对应 client
- **AND** 不通过`pluginbridge.Runtime()`、`pluginbridge.Data()`或同类根包入口访问业务能力

#### Scenario:pluginbridge 根包暴露协议能力

- **WHEN** 调用方需要 ABI 常量、bridge envelope、WASM section、host call、host service wire 校验或动态路由 dispatcher
- **THEN** 它可以使用`pluginbridge`根包或更精确子包提供的协议入口
- **AND** 根包不得因此重新拥有 capability 业务语义
- **AND** 根包不得提供对`capability/guest`业务 client 的兼容转发方法或类型别名

#### Scenario:新增动态宿主能力 client

- **WHEN** 开发者新增一个动态插件宿主能力 client
- **THEN** client 首先定义在`pkg/plugin/capability/guest`
- **AND** `pluginbridge`只维护必要的 service/method wire 常量、payload 编解码和授权校验

## REMOVED Requirements

### Requirement:根包 facade 必须保持现有稳定调用路径

**Reason**: 本项目没有历史兼容负担，继续要求根包 facade 保持全部现有调用路径会阻止删除旧业务能力 client，也会让`pluginbridge`继续承担宿主能力目录语义。

**Migration**: 只保留动态插件协议和 transport 必需入口；动态插件业务能力调用迁移到`pkg/plugin/capability/guest`；宿主运行时使用精确协议子包或根包最小协议 facade。

#### Scenario:旧业务能力入口被删除

- **WHEN** 代码继续调用`pluginbridge.Runtime()`、`pluginbridge.Storage()`、`pluginbridge.Network()`、`pluginbridge.Data()`、`pluginbridge.Cache()`、`pluginbridge.Config()`或同类业务能力入口
- **THEN** 编译必须失败
- **AND** 调用方迁移到`pkg/plugin/capability/guest`

#### Scenario:协议入口仍可定位

- **WHEN** 代码需要`hostServices`声明校验、WASM section 读取、host call envelope 或 bridge request/response 编解码
- **THEN** 它使用`pluginbridge`协议 facade 或职责精确的协议子包
- **AND** 行为与删除业务能力入口前保持等价
