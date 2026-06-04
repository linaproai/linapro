## ADDED Requirements

### Requirement: host service 协议同步点必须由单一描述源覆盖

系统 SHALL 为动态插件 host service 协议维护单一描述源或等价的集中元数据表。该描述源 MUST 覆盖 service、method、capability、资源类型、请求响应 payload、guest client 方法、非 WASI stub、public protocol alias 和 host dispatcher 绑定。新增、删除或重命名 host service method 时，自动化验证 MUST 能发现任一同步点遗漏。

#### Scenario: 新增 host service method 缺少 guest client

- **WHEN** 描述源声明了新的 host service method
- **AND** `pkg/plugin/capability/guest` 没有提供对应 guest client 或明确标记为不发布 guest helper
- **THEN** 协议覆盖测试失败
- **AND** 失败信息指出缺少 guest client 覆盖

#### Scenario: 新增 payload codec 缺少 public protocol alias

- **WHEN** 描述源声明了新的请求或响应 payload codec
- **AND** `pkg/plugin/pluginbridge/protocol` 未公开对应类型或 marshal/unmarshal 入口
- **THEN** 协议覆盖测试失败
- **AND** 调用方不得只能通过 internal hostservice 包访问公开协议 payload

#### Scenario: dispatcher 未覆盖声明的方法

- **WHEN** 描述源声明了 host service method
- **AND** host-side dispatcher 没有处理该 service/method
- **THEN** 协议覆盖测试失败
- **AND** 未知方法不能在运行时才被偶然发现

### Requirement: 生成代码不得改变 dynamic plugin bridge wire 行为

系统 MAY 使用生成代码维护 host service DTO、codec、alias、guest client、stub 或 dispatcher 绑定。生成代码 MUST 保持现有 service/method 字符串、protobuf wire 字段编号、默认值、错误状态和 guest helper 行为不变，并且 MUST 带有 `Code generated` 标记。生成流程必须有跨平台入口或明确的 Go 测试覆盖。

#### Scenario: 生成后 payload round trip 等价

- **WHEN** host service payload codec 由生成代码维护
- **THEN** 现有 runtime、storage、data、cache、lock、config、notify、manifest、org、tenant 和 AI payload round trip 测试继续通过
- **AND** 字段编号和默认值语义不得变化

#### Scenario: 生成入口跨平台可执行

- **WHEN** 开发者需要刷新 host service 协议生成文件
- **THEN** 仓库提供 Go 命令、`linactl` 子命令、`make` 目标或等价跨平台入口
- **AND** 不依赖仅 Unix shell 可用的脚本作为唯一入口
