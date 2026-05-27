# pluginbridge 子组件化架构

## Purpose

定义 `apps/lina-core/pkg/pluginbridge` 的子组件化拆分规范，确保职责清晰、依赖方向正确、协议行为不变。
## Requirements
### Requirement:pluginbridge 必须按职责提供公开子组件

系统 SHALL 将 `apps/lina-core/pkg/pluginbridge` 组织为职责明确的公开子组件包。子组件至少覆盖 bridge 合约、bridge 编解码、WASM 产物辅助、host call 协议、host service 协议和 guest SDK。根目录不得继续承载大量跨职责实现文件；根包只允许保留 facade、包说明和必要的兼容入口。

#### Scenario:开发者按职责定位 bridge 能力
- **当** 开发者需要查看动态插件 bridge 合约、编解码、WASM 产物解析、host call、host service 或 guest SDK
- **则** 对应源码位于语义明确的 `pkg/pluginbridge/<subcomponent>/` 子组件目录
- **且** 根包目录下的生产源码文件数量保持在 1 到 3 个范围内

#### Scenario:子组件名称表达稳定职责
- **当** 系统完成 pluginbridge 子组件化
- **则** 子组件包名必须使用清晰职责名称
- **且** 不得使用 `common`、`util`、`helper` 等兜底包名承载跨领域逻辑

### Requirement:子组件依赖方向必须防止循环依赖

系统 SHALL 明确定义 `pluginbridge` 子组件的依赖方向。底层合约和协议子组件不得依赖根包 facade 或 guest SDK；根包 facade 可以依赖各子组件。任何子组件下沉的 `internal` 实现包必须服务于明确父组件，不得成为跨组件兜底依赖。

#### Scenario:子组件构建无 import cycle
- **当** 执行 `go test ./pkg/pluginbridge/...`
- **则** 所有子组件包必须通过编译
- **且** 不得出现 import cycle

#### Scenario:底层包不依赖根包
- **当** 检查 `contract`、`codec`、`artifact`、`hostcall`、`hostservice` 子组件 import
- **则** 这些子组件不得 import `lina-core/pkg/pluginbridge`
- **且** 只能依赖职责更底层或同层允许的子组件

### Requirement:宿主内部调用必须优先使用精确子组件

系统 SHALL 将项目可控的宿主内部调用逐步迁移到精确子组件 import。动态插件 guest 代码可继续使用根包 facade 兼容路径，但宿主 runtime、WASM host function、artifact 解析、i18n/apidoc 资源加载和 plugindb 应使用能表达职责边界的子组件包。

#### Scenario:宿主 runtime 使用精确子组件
- **当** 宿主运行时解析动态插件产物或执行 Wasm bridge 请求
- **则** 代码优先 import `pluginbridge/artifact`、`pluginbridge/codec`、`pluginbridge/hostcall` 或 `pluginbridge/hostservice`
- **且** 不再因为只需要单一协议能力而 import 整个根包 facade

#### Scenario:插件侧兼容路径仍可用
- **当** 动态插件 guest 代码继续调用 `pluginbridge.NewGuestRuntime`、`pluginbridge.BindJSON` 或 `pluginbridge.Runtime()`
- **则** 系统继续提供兼容入口
- **且** 这些入口委托到 guest 子组件

### Requirement:子组件化不得改变 bridge 协议行为

系统 SHALL 保证子组件化是结构重构，不改变动态插件 bridge 协议行为。ABI 常量、WASM custom section 名称、protobuf 字段编号、host call 状态码、host service service/method 字符串、payload 编解码结果和 guest helper 行为必须保持不变。

#### Scenario:bridge envelope 编解码保持不变
- **当** 使用重构后的 API 编码并解码 `BridgeRequestEnvelopeV1` 或 `BridgeResponseEnvelopeV1`
- **则** round trip 结果与重构前等价
- **且** 现有协议测试必须继续通过

#### Scenario:host service payload codec 保持不变
- **当** 使用重构后的 API 编码并解码 runtime、storage、network、data、cache、lock、config、notify 或 cron host service payload
- **则** round trip 结果与重构前等价
- **且** 字段编号和默认值语义不得变化

#### Scenario:facade 和子组件结果一致
- **当** 同一调用同时通过根包 facade 和目标子组件执行
- **则** 两者返回相同结果或等价错误
- **且** 测试必须覆盖至少 bridge envelope、WASM section 和 host service payload 三类代表性入口

### Requirement:子组件化必须有自动化验证

系统 SHALL 为 `pluginbridge` 子组件化提供自动化验证。验证必须覆盖根包兼容、子组件编译、宿主内部调用、动态插件样例和 Wasm guest 构建。

#### Scenario:pluginbridge 子组件测试通过
- **当** 执行 `go test ./pkg/pluginbridge/...`
- **则** 根包 facade 和所有子组件测试必须通过

#### Scenario:宿主插件运行时测试通过
- **当** 执行插件运行时、WASM host function 和 plugindb 相关 Go 测试
- **则** 测试必须通过
- **且** 不得出现因 import 迁移导致的协议行为回归

#### Scenario:动态插件样例可构建
- **当** 对动态插件样例执行普通 Go 测试和 `GOOS=wasip1 GOARCH=wasm` 构建
- **则** 样例必须通过编译
- **且** guest 侧 bridge helper 调用必须可用

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

