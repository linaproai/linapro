## ADDED Requirements

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

#### Scenario: catalog 与公开 protocol 出口一致

- **WHEN** catalog 声明某个已发布 method 使用公开 payload 类型
- **THEN** `pkg/plugin/pluginbridge/protocol`必须提供对应 DTO 或 codec 入口
- **AND** 治理测试必须在缺少公开协议出口时失败

### Requirement: 普通领域 host service payload 必须优先使用 JSON envelope

系统 SHALL 为普通领域 host service 提供统一 JSON request/response envelope。新增普通领域能力默认 MUST 通过该 JSON envelope 承载领域 DTO 或投影，不得为每个领域默认新增专用`protocol_hostservice_<x>_codec.go`和手写`protowire`codec。只有存在明确性能、资源授权、二进制内容、事务计划或 wire 稳定性需求的服务，才 MAY 保留或新增专用 codec，并且该例外 MUST 在 catalog 中标记 payload kind。

#### Scenario: 新增普通领域能力

- **WHEN** 开发者新增`users`、`dict`、`files`、`sessions`同类的普通领域 host service method
- **THEN** guest client 使用统一 JSON envelope 编码请求和响应
- **AND** protocol 层不得为了该普通领域新增专用 per-domain `protowire`codec 文件
- **AND** 测试必须覆盖 JSON envelope round trip 和 typed client 结果映射

#### Scenario: 特殊服务保留专用 codec

- **WHEN** host service 属于`storage`、`cache`、`lock`、`data`、`recordstore`、`network`或经规范确认的性能和资源敏感服务
- **THEN** 它可以继续使用专用二进制或`protowire`codec
- **AND** catalog 必须将其 payload kind 标记为专用 codec
- **AND** 现有 payload round trip、字段默认值和错误 envelope 语义必须保持不变

#### Scenario: 未说明依据的专用 codec 被拒绝

- **WHEN** 新增普通领域 host service 同时新增 per-domain 专用 codec
- **THEN** 审查必须拒绝该实现
- **AND** 除非设计或任务记录说明性能、资源或 wire 稳定性依据，否则不得通过治理验证

### Requirement: host service catalog 覆盖治理必须校验 guest、codec 和 dispatch

系统 SHALL 基于`protocol/hostservices`catalog 双向校验 host service 同步点。catalog 中声明发布的 guest client、payload codec 和 host dispatcher 必须存在；实现中出现的 guest client selector、专用 codec 或 dispatcher 注册也必须反向存在于 catalog。新增、删除或重命名 host service method 时，自动化验证 MUST 能发现任一同步点遗漏或孤儿实现。

#### Scenario: catalog 声明 guest client 但实现缺失

- **WHEN** catalog 声明某个 method 发布 guest client
- **AND** guest typed client 或目录 getter 没有提供对应入口
- **THEN** 治理测试失败
- **AND** 失败信息指出缺少 guest client 覆盖

#### Scenario: dispatch 注册出现孤儿 method

- **WHEN** host dispatch registry 注册了某个 service/method
- **AND** catalog 中不存在该 service/method 或未声明 dispatcher 发布状态
- **THEN** 治理测试失败
- **AND** 该 method 不得仅靠运行时未知路径暴露

#### Scenario: 普通领域出现专用 codec 漂移

- **WHEN** 静态检索发现普通领域新增`protocol_hostservice_<x>_codec.go`
- **AND** catalog 未将该 service 标记为专用 codec
- **THEN** 治理测试失败
- **AND** 开发者必须改用 JSON envelope 或补充规范级例外依据
