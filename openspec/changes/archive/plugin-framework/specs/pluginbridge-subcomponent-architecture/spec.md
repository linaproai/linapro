# pluginbridge 子组件化架构

## Purpose

定义 `apps/lina-core/pkg/pluginbridge` 的子组件化拆分规范，确保职责清晰、依赖方向正确、协议行为不变。
## Requirements
### Requirement:pluginbridge 必须按职责提供公开子组件

系统 SHALL 将`apps/lina-core/pkg/pluginbridge`组织为动态插件 ABI、WASM transport 和协议 facade 组件。根包必须保留稳定 facade、包说明和必要的兼容入口；`contract`和`guest`等确需被插件作者直接依赖的契约或 guest SDK 可以保持公开。bridge 编解码、WASM 产物解析、host call dispatcher、host service payload wire 等低层实现 MUST 默认放入`pluginbridge/internal/<subcomponent>/`，不得作为插件业务开发公共 API 暴露。

#### Scenario:开发者按职责定位 bridge 能力
- **当** 开发者需要查看动态插件 bridge 合约或 guest SDK
- **则** 稳定入口位于`pkg/pluginbridge`根 facade、`pkg/pluginbridge/contract`或`pkg/pluginbridge/guest`
- **且** 低层 codec、artifact、host call 和 host service wire 实现位于`pkg/pluginbridge/internal/**`

#### Scenario:子组件名称表达稳定职责
- **当** 系统完成 pluginbridge 职责收敛
- **则** 公开子组件和内部子组件包名必须使用清晰职责名称
- **且** 不得使用`common`、`util`、`helper`等兜底包名承载跨领域逻辑

### Requirement:子组件依赖方向必须防止循环依赖

系统 SHALL 明确定义`pluginbridge`子组件的依赖方向。底层合约和协议子组件不得依赖根包、guest runtime 或 route helper；`protocol`可以依赖底层 contract、artifact、codec、hostcall 和 hostservice 子组件并作为公开聚合出口。任何子组件下沉的`internal`实现包必须服务于明确父组件，不得成为跨组件兜底依赖。

#### Scenario:子组件构建无 import cycle
- **WHEN** 执行`go test ./pkg/plugin/pluginbridge/...`
- **THEN** 所有子组件包必须通过编译
- **AND** 不得出现 import cycle

#### Scenario:底层包不依赖根包
- **WHEN** 检查`contract`、`codec`、`artifact`、`hostcall`、`hostservice`子组件 import
- **THEN** 这些子组件不得 import `lina-core/pkg/plugin/pluginbridge`
- **AND** 只能依赖职责更底层或同层允许的子组件

### Requirement:宿主内部调用必须优先使用精确子组件

系统 SHALL 将项目可控的宿主内部调用迁移到能表达职责边界的入口。插件侧 guest 代码可继续使用`pluginbridge`根包 facade 或`pluginbridge/guest`；宿主 runtime、WASM host function、artifact 解析、i18n/apidoc 资源加载和 data host 等内部实现应通过根 facade、同一`internal`授权边界内的精确内部子组件或更上层`pluginservice`适配调用，不得把低层实现子包重新公开为公共 API。

#### Scenario:宿主 runtime 使用受控内部入口
- **当** 宿主运行时解析动态插件产物或执行 Wasm bridge 请求
- **则** 代码使用`pluginbridge`根 facade 或`pluginbridge/internal/artifact`、`pluginbridge/internal/codec`等授权内部入口
- **且** 不为了单一协议能力重新公开`pluginbridge/artifact`、`pluginbridge/codec`、`pluginbridge/hostcall`或`pluginbridge/hostservice`

#### Scenario:插件侧兼容路径仍可用
- **当** 动态插件 guest 代码继续调用`pluginbridge.NewGuestRuntime`、`pluginbridge.BindJSON`或`pluginbridge.Runtime()`
- **则** 系统继续提供兼容入口
- **且** 这些入口委托到 guest 子组件或内部 transport 实现

### Requirement:子组件化不得改变 bridge 协议行为

系统 SHALL 保证子组件化和路径迁移是结构重构，不改变动态插件 bridge 协议行为。ABI 常量、WASM custom section 名称、protobuf 字段编号、host call 状态码、host service service/method 字符串、payload 编解码结果和 guest helper 行为必须保持不变。

#### Scenario:bridge envelope 编解码保持不变
- **WHEN** 使用重构后的 API 编码并解码`BridgeRequestEnvelopeV1`或`BridgeResponseEnvelopeV1`
- **THEN** round trip 结果与重构前等价
- **AND** 现有协议测试必须继续通过

#### Scenario:host service payload codec 保持不变
- **WHEN** 使用重构后的 API 编码并解码 runtime、storage、network、data、cache、lock、config、notify 或 cron host service payload
- **THEN** round trip 结果与重构前等价
- **AND** 字段编号和默认值语义不得变化

#### Scenario:protocol 和底层子组件结果一致
- **WHEN** 同一协议调用同时通过`pluginbridge/protocol`和底层目标子组件执行
- **THEN** 两者返回相同结果或等价错误
- **AND** 测试必须覆盖至少 bridge envelope、WASM section 和 host service payload 三类代表性入口

### Requirement:子组件化必须有自动化验证

系统 SHALL 为`pluginbridge`子组件化和路径迁移提供自动化验证。验证必须覆盖公开`protocol`出口、子组件编译、宿主内部调用、动态插件样例和 Wasm guest 构建。

#### Scenario:pluginbridge 子组件测试通过
- **WHEN** 执行`go test ./pkg/plugin/pluginbridge/...`
- **THEN** `protocol`出口和所有子组件测试必须通过

#### Scenario:宿主插件运行时测试通过
- **WHEN** 执行插件运行时、WASM host function 和 data host 相关 Go 测试
- **THEN** 测试必须通过
- **AND** 不得出现因 import 迁移导致的协议行为回归

#### Scenario:动态插件样例可构建
- **WHEN** 对动态插件样例执行普通 Go 测试和`GOOS=wasip1 GOARCH=wasm`构建
- **THEN** 样例必须通过编译
- **AND** guest 侧 bridge runtime helper 与`capability/guest`能力 client 调用必须可用

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

### Requirement:公开协议出口必须唯一

系统 SHALL 以`lina-core/pkg/plugin/pluginbridge/protocol`作为 dynamic plugin bridge 的唯一公开协议出口。公开协议 DTO、ABI 常量、WASM section 名称、host call payload、host service payload 和协议 codec 必须通过`protocol`包访问；`pluginbridge`根包不得作为 facade 重新导出这些类型、常量或函数。协议实现只能存在于一个权威子组件中。

#### Scenario:目标协议 import 路径可以编译
- **WHEN** 宿主内部代码、动态插件样例或用户插件需要访问 bridge envelope、ABI 常量、WASM section、host call、host service 或 codec
- **THEN** 代码 import `lina-core/pkg/plugin/pluginbridge/protocol`
- **AND** 返回行为与迁移前目标协议语义保持一致

#### Scenario:根包不重复公开协议逻辑
- **WHEN** 开发者查看`pkg/plugin/pluginbridge`根包
- **THEN** 根包只提供包说明
- **AND** 根包不得维护 type alias、const alias、codec wrapper、protobuf wire 编解码、WASM section 遍历或 host service payload 编解码入口

### Requirement: pluginbridge 不得拥有宿主业务能力语义

系统 SHALL 将`pluginbridge`限定为动态插件 transport 和 ABI 层。宿主业务能力的契约、授权资源语义、降级行为和消费 service MUST 归属于`pluginservice`及其能力组件；`pluginbridge`不得新增与源码插件平行的业务能力目录。

#### Scenario: 动态插件调用框架能力

- **WHEN** 动态插件调用`framework.org.v1`能力
- **THEN** `pluginbridge`只负责编码、传输和解码 host service envelope
- **AND** 能力授权、provider 激活、消费 DTO 和降级语义由`pluginservice`、`orgcap`或`tenantcap`处理

#### Scenario: 新增宿主能力时不修改 ABI 语义 owner

- **WHEN** 系统新增一个插件可消费宿主能力
- **THEN** 该能力首先定义在`pluginservice`或其独立能力组件
- **AND** 动态插件仅在需要 transport 支持时扩展 bridge payload 或 handler 适配

### Requirement: host service 协议同步点必须由单一描述源覆盖

系统 SHALL 为动态插件 host service 协议维护单一描述源或等价的集中元数据表。该描述源 MUST 覆盖 service、method、capability、资源类型、请求响应 payload、guest client 方法、非 WASI stub、public protocol alias 和 host dispatcher 绑定。新增、删除或重命名 host service method 时，自动化验证 MUST 能发现任一同步点遗漏。

#### Scenario: 新增 host service method 缺少 guest client

- **WHEN** 描述源声明了新的 host service method
- **AND** guest 侧没有提供对应 guest client 或明确标记为不发布 guest helper
- **THEN** 协议覆盖测试失败

#### Scenario: dispatcher 未覆盖声明的方法

- **WHEN** 描述源声明了 host service method
- **AND** host-side dispatcher 没有处理该 service/method
- **THEN** 协议覆盖测试失败

### Requirement: 生成代码不得改变 dynamic plugin bridge wire 行为

系统 MAY 使用生成代码维护 host service DTO、codec、alias、guest client、stub 或 dispatcher 绑定。生成代码 MUST 保持现有 service/method 字符串、protobuf wire 字段编号、默认值、错误状态和 guest helper 行为不变，并且 MUST 带有 `Code generated` 标记。生成流程必须有跨平台入口或明确的 Go 测试覆盖。

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

### Requirement: host service catalog 必须作为公开协议描述源

系统 SHALL 在`pkg/plugin/pluginbridge/protocol/hostservices`维护动态插件 host service 的公开协议 catalog。该 catalog MUST 覆盖 service、method、capability、资源类型、payload 形态、请求响应 payload 名称、guest client 发布状态和 host dispatcher 发布状态。`pluginbridge/internal/hostservice`可以继续提供 manifest validation、capability derivation 和治理测试辅助，但其 descriptor MUST 从公开 catalog 派生，不得维护第二份手写 service/method 表。

#### Scenario: 新增普通领域 host service

- **WHEN** 开发者新增一个普通领域 host service
- **THEN** service、method、capability、资源类型、payload 形态、guest client 发布状态和 host dispatcher 发布状态必须在`protocol/hostservices`catalog 中声明
- **AND** `pluginbridge/internal/hostservice`descriptor 从该 catalog 读取或转换元数据
- **AND** 不得在`pluginbridge/internal/hostservice`新增独立的手写镜像 descriptor 数据

#### Scenario: 宿主 dispatch 需要读取 host service 元数据

- **WHEN** `internal/service/plugin/internal/wasm`需要校验 dispatcher 注册覆盖
- **THEN** 它可以导入`pkg/plugin/pluginbridge/protocol/hostservices`
- **AND** 不得导入`pkg/plugin/pluginbridge/internal/hostservice`
- **AND** Go internal import 边界必须通过编译和静态检索验证

### Requirement: 普通领域 host service payload 必须优先使用 JSON envelope

系统 SHALL 为普通领域 host service 提供统一 JSON request/response envelope。新增普通领域能力默认 MUST 通过该 JSON envelope 承载领域 DTO 或投影，不得为每个领域默认新增专用`protocol_hostservice_<x>_codec.go`和手写`protowire`codec。只有存在明确性能、资源授权、二进制内容、事务计划或 wire 稳定性需求的服务，才 MAY 保留或新增专用 codec，并且该例外 MUST 在 catalog 中标记 payload kind。

### Requirement: host service catalog 覆盖治理必须校验 guest、codec 和 dispatch

系统 SHALL 基于`protocol/hostservices`catalog 双向校验 host service 同步点。catalog 中声明发布的 guest client、payload codec 和 host dispatcher 必须存在；实现中出现的 guest client selector、专用 codec 或 dispatcher 注册也必须反向存在于 catalog。新增、删除或重命名 host service method 时，自动化验证 MUST 能发现任一同步点遗漏或孤儿实现。

### Requirement: host service README 表格必须由 descriptor 生成

系统 SHALL 使用`pluginbridge/internal/hostservice`的 descriptor 作为`pkg/plugin`双语`README`中 host service 表格的单一事实源。`README.md`与`README.zh-CN.md`中的 host service 表格 MUST 位于稳定生成标记之间，并由 descriptor 渲染器维护。自动化测试 MUST 比对 descriptor 渲染结果与当前文档内容，发现漂移时失败。

### Requirement: host service descriptor 覆盖治理必须双向校验

系统 SHALL 通过自动化测试双向校验 host service descriptor、guest client selector 和宿主 dispatcher 绑定。descriptor 中声明发布的 guest client 或 dispatcher method MUST 有对应实现；实现中出现的 service/method 也 MUST 反向存在于 descriptor 并声明对应发布位。宿主 service 级 switch 与 dispatcher 文件集合 MUST 与 descriptor 中启用 dispatcher 的 service 集合保持一致。

### Requirement: guest host service client 必须使用注入式传输单轨

系统 SHALL 将动态插件 guest host service client 统一为 invoker 注入式结构。除`recordstore`等承载领域执行逻辑的独立 SDK 外，`pluginbridge`根目录 MUST NOT 保留逐域`pluginbridge_hostcall_*_wasip1.go`单例客户端、逐域 adapter 或逐域非 WASI 镜像 stub。`pluginbridge_directory.go`MUST 通过统一 invoker 装配基础能力和领域能力 guest client，wire 格式和 getter 签名 MUST 保持不变。
