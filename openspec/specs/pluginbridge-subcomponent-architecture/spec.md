# pluginbridge 子组件化架构

## 目的

定义 `apps/lina-core/pkg/pluginbridge` 的子组件化拆分规范，确保职责清晰、依赖方向正确、协议行为不变。

## 需求

### 需求：pluginbridge 必须按职责提供公开子组件

系统 SHALL 将 `apps/lina-core/pkg/pluginbridge` 组织为职责明确的公开子组件包。子组件至少覆盖 bridge 合约、bridge 编解码、WASM 产物辅助、host call 协议、host service 协议和 guest SDK。根目录不得继续承载大量跨职责实现文件；根包只允许保留 facade、包说明和必要的兼容入口。

#### 场景：开发者按职责定位 bridge 能力
- **当** 开发者需要查看动态插件 bridge 合约、编解码、WASM 产物解析、host call、host service 或 guest SDK
- **则** 对应源码位于语义明确的 `pkg/pluginbridge/<subcomponent>/` 子组件目录
- **且** 根包目录下的生产源码文件数量保持在 1 到 3 个范围内

#### 场景：子组件名称表达稳定职责
- **当** 系统完成 pluginbridge 子组件化
- **则** 子组件包名必须使用清晰职责名称
- **且** 不得使用 `common`、`util`、`helper` 等兜底包名承载跨领域逻辑

### 需求：根包 facade 必须保持现有稳定调用路径

系统 SHALL 保留 `lina-core/pkg/pluginbridge` 根包作为兼容 facade。现有公开常量、类型和函数必须继续可以通过根包访问，并委托到对应子组件实现。facade 不得复制协议实现逻辑；协议实现只能存在于一个权威子组件中。

#### 场景：旧 import 路径继续编译
- **当** 宿主内部代码、动态插件样例或用户插件继续 import `lina-core/pkg/pluginbridge`
- **则** 既有公开 API 调用必须继续编译
- **且** 返回行为与重构前保持一致

#### 场景：facade 不重复实现协议逻辑
- **当** 开发者查看根包 facade
- **则** 公开入口通过 type alias、const alias 或 wrapper 调用子组件
- **且** 根包不得维护独立的 protobuf wire 编解码、WASM section 遍历或 host service payload 编解码实现

### 需求：子组件依赖方向必须防止循环依赖

系统 SHALL 明确定义 `pluginbridge` 子组件的依赖方向。底层合约和协议子组件不得依赖根包 facade 或 guest SDK；根包 facade 可以依赖各子组件。任何子组件下沉的 `internal` 实现包必须服务于明确父组件，不得成为跨组件兜底依赖。

#### 场景：子组件构建无 import cycle
- **当** 执行 `go test ./pkg/pluginbridge/...`
- **则** 所有子组件包必须通过编译
- **且** 不得出现 import cycle

#### 场景：底层包不依赖根包
- **当** 检查 `contract`、`codec`、`artifact`、`hostcall`、`hostservice` 子组件 import
- **则** 这些子组件不得 import `lina-core/pkg/pluginbridge`
- **且** 只能依赖职责更底层或同层允许的子组件

### 需求：宿主内部调用必须优先使用精确子组件

系统 SHALL 将项目可控的宿主内部调用逐步迁移到精确子组件 import。动态插件 guest 代码可继续使用根包 facade 兼容路径，但宿主 runtime、WASM host function、artifact 解析、i18n/apidoc 资源加载和 plugindb 应使用能表达职责边界的子组件包。

#### 场景：宿主 runtime 使用精确子组件
- **当** 宿主运行时解析动态插件产物或执行 Wasm bridge 请求
- **则** 代码优先 import `pluginbridge/artifact`、`pluginbridge/codec`、`pluginbridge/hostcall` 或 `pluginbridge/hostservice`
- **且** 不再因为只需要单一协议能力而 import 整个根包 facade

#### 场景：插件侧兼容路径仍可用
- **当** 动态插件 guest 代码继续调用 `pluginbridge.NewGuestRuntime`、`pluginbridge.BindJSON` 或 `pluginbridge.Runtime()`
- **则** 系统继续提供兼容入口
- **且** 这些入口委托到 guest 子组件

### 需求：子组件化不得改变 bridge 协议行为

系统 SHALL 保证子组件化是结构重构，不改变动态插件 bridge 协议行为。ABI 常量、WASM custom section 名称、protobuf 字段编号、host call 状态码、host service service/method 字符串、payload 编解码结果和 guest helper 行为必须保持不变。

#### 场景：bridge envelope 编解码保持不变
- **当** 使用重构后的 API 编码并解码 `BridgeRequestEnvelopeV1` 或 `BridgeResponseEnvelopeV1`
- **则** round trip 结果与重构前等价
- **且** 现有协议测试必须继续通过

#### 场景：host service payload codec 保持不变
- **当** 使用重构后的 API 编码并解码 runtime、storage、network、data、cache、lock、config、notify 或 cron host service payload
- **则** round trip 结果与重构前等价
- **且** 字段编号和默认值语义不得变化

#### 场景：facade 和子组件结果一致
- **当** 同一调用同时通过根包 facade 和目标子组件执行
- **则** 两者返回相同结果或等价错误
- **且** 测试必须覆盖至少 bridge envelope、WASM section 和 host service payload 三类代表性入口

### 需求：子组件化必须有自动化验证

系统 SHALL 为 `pluginbridge` 子组件化提供自动化验证。验证必须覆盖根包兼容、子组件编译、宿主内部调用、动态插件样例和 Wasm guest 构建。

#### 场景：pluginbridge 子组件测试通过
- **当** 执行 `go test ./pkg/pluginbridge/...`
- **则** 根包 facade 和所有子组件测试必须通过

#### 场景：宿主插件运行时测试通过
- **当** 执行插件运行时、WASM host function 和 plugindb 相关 Go 测试
- **则** 测试必须通过
- **且** 不得出现因 import 迁移导致的协议行为回归

#### 场景：动态插件样例可构建
- **当** 对动态插件样例执行普通 Go 测试和 `GOOS=wasip1 GOARCH=wasm` 构建
- **则** 样例必须通过编译
- **且** guest 侧 bridge helper 调用必须可用
