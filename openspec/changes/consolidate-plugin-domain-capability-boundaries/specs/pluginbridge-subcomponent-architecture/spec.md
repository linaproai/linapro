## ADDED Requirements

### Requirement: 动态 guest 普通领域代理必须收敛到 domainhostcall

系统 SHALL 将动态插件 guest 侧普通领域能力代理固定在`pkg/plugin/pluginbridge/internal/domainhostcall`或等价 internal 子组件中。`pkg/plugin/pluginbridge`公共包 SHALL 只暴露动态插件开发者使用的能力目录、声明期启动 facade、资源型 host service client 和必要 transport facade；公共包不得为普通领域能力长期维护与`pkg/plugin/capability/*cap`平行的接口集合。

#### Scenario: guest 读取用户领域能力

- **WHEN** 动态插件通过`pluginbridge.Default().Users()`读取用户投影
- **THEN** 公共方法返回`usercap.Service`
- **AND** transport 实现位于`pluginbridge/internal/domainhostcall`
- **AND** 公共`guest`包不重新定义用户领域接口

#### Scenario: guest 调用 AI 能力

- **WHEN** 动态插件通过`pluginbridge.Default().AI()`调用`AI`能力
- **THEN** 返回值必须复用或实现`aicap.Service`
- **AND** 文本、图片、向量、音频、视觉、文档、安全和视频子能力代理必须对齐`capability/aicap`子接口
- **AND** 不得继续维护与`aicap`平行的`pluginbridge.AITextService`、`pluginbridge.AIImageService`或同类领域接口作为长期公共契约

#### Scenario: 协议目录描述动态 host service

- **WHEN** 新增或修改动态插件普通领域`host service method`
- **THEN** `pluginbridge/internal/hostservice`描述源只维护 service、method、capability、资源类型、payload、guest client 和 dispatcher 同步元数据
- **AND** 领域业务接口、输入输出语义和降级行为必须继续归属`pkg/plugin/capability`

### Requirement: pluginbridge 协议目录不得成为领域能力实现入口

系统 SHALL 将`pkg/plugin/pluginbridge/protocol`和`pkg/plugin/pluginbridge/internal/hostservice`限定为动态插件协议与授权目录。它们 MUST NOT 构造宿主领域能力实现、保存运行期领域服务实例或定义与源码插件不同的领域业务规则。

#### Scenario: 宿主分发动态领域方法

- **WHEN** `WASM`host dispatcher 收到普通领域`host service`调用
- **THEN** `pluginbridge`协议目录只提供 service/method 常量、payload 编解码和授权元数据
- **AND** 实际业务调用必须进入`capability.Services`或其领域`*cap.Service`

#### Scenario: 新增领域能力状态语义

- **WHEN** 领域能力需要新增可用性、状态或降级语义
- **THEN** 语义必须定义在对应`*cap`组件包
- **AND** `pluginbridge`只能增加必要的 transport payload 或 descriptor 同步点

### Requirement: 动态 host service payload codec 必须归属 protocol 公共协议目录

系统 SHALL 将动态插件`host service`的 payload DTO、wire payload struct、marshal 和 unmarshal 实现归属到`pkg/plugin/pluginbridge/protocol`公共协议目录。`pkg/plugin/pluginbridge/internal/hostservice` SHALL 只维护 descriptor、capability、资源类型、授权校验、清单规范化和治理同步点，不得作为 payload codec 的实际 owner。

#### Scenario: 新增动态 host service payload

- **WHEN** 开发者新增或修改动态插件`host service`请求或响应 payload
- **THEN** payload struct 和编解码函数必须定义在`pluginbridge/protocol`
- **AND** `pluginbridge/internal/hostservice`只能引用`protocol`中的公开协议类型

#### Scenario: 公共协议目录复用 codec

- **WHEN** 动态 guest、`WASM`host dispatcher 或 descriptor 测试需要编解码`host service`payload
- **THEN** 调用方必须通过`pluginbridge/protocol`获取 codec
- **AND** `protocol`不得通过类型别名或函数别名重新导出`internal/hostservice`中的 codec 实现

#### Scenario: 内部 hostservice 维护授权治理

- **WHEN** 动态插件清单声明`hostServices`
- **THEN** `pluginbridge/internal/hostservice`负责校验 service、method、resource kind、capability 和 manifest 规范化规则
- **AND** 该内部包不得反向拥有或导出 payload wire 格式实现
