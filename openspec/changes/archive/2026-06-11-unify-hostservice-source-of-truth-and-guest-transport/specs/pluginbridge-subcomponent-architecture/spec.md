# pluginbridge 子组件化架构规范增量

## ADDED Requirements

### Requirement: host service README 表格必须由 descriptor 生成

系统 SHALL 使用`pluginbridge/internal/hostservice`的 descriptor 作为`pkg/plugin`双语`README`中 host service 表格的单一事实源。`README.md`与`README.zh-CN.md`中的 host service 表格 MUST 位于稳定生成标记之间，并由 descriptor 渲染器维护。自动化测试 MUST 比对 descriptor 渲染结果与当前文档内容，发现漂移时失败。

#### Scenario: 维护双语 host service 表格

- **WHEN** 开发者维护 host service 文档表格
- **THEN** 系统基于 descriptor 渲染结果更新`apps/lina-core/pkg/plugin/README.md`与`README.zh-CN.md`中的生成区块
- **AND** 两份文档保持相同 service/method/capability/resource 事实，仅语言不同

#### Scenario: descriptor 变更但 README 未刷新

- **WHEN** descriptor 中新增、删除或修改 host service 声明
- **AND** 对应`README`生成区块未同步刷新
- **THEN** README 漂移测试失败
- **AND** 失败信息提示从 descriptor 渲染器更新 host service 文档生成区块

#### Scenario: 不保留独立生成入口

- **WHEN** 开发者维护 host service 文档表格
- **THEN** 仓库不需要提供独立`go run`生成入口、平台脚本或默认开发命令
- **AND** README 漂移治理不得依赖 Unix-only shell 管道或平台专属命令

### Requirement: host service descriptor 覆盖治理必须双向校验

系统 SHALL 通过自动化测试双向校验 host service descriptor、guest client selector 和宿主 dispatcher 绑定。descriptor 中声明发布的 guest client 或 dispatcher method MUST 有对应实现；实现中出现的 service/method 也 MUST 反向存在于 descriptor 并声明对应发布位。宿主 service 级 switch 与 dispatcher 文件集合 MUST 与 descriptor 中启用 dispatcher 的 service 集合保持一致。

#### Scenario: descriptor method 缺少宿主 dispatcher

- **WHEN** descriptor 声明某个 method 的`Dispatcher=true`
- **AND** 宿主 dispatcher selector 未处理该 service/method
- **THEN** 覆盖治理测试失败
- **AND** 失败信息指出缺少 dispatcher 覆盖的 service/method

#### Scenario: 宿主 dispatcher 存在孤儿 method

- **WHEN** 宿主 dispatcher selector 处理某个 service/method
- **AND** descriptor 中不存在该 method 或未声明`Dispatcher=true`
- **THEN** 覆盖治理测试失败
- **AND** 失败信息指出孤儿 dispatcher service/method

#### Scenario: guest client selector 与 descriptor 不一致

- **WHEN** guest client selector 中出现某个 host service method
- **AND** descriptor 中不存在该 method 或未声明`GuestClient=true`
- **THEN** 覆盖治理测试失败
- **AND** 动态插件 guest client 不得发布未纳入 descriptor 治理的方法

#### Scenario: dispatcher 文件集合与 service 集合不一致

- **WHEN** 宿主存在`dispatchXxxHostService`文件但 descriptor 没有对应 dispatcher service，或 descriptor 有 dispatcher service 但缺少对应文件
- **THEN** 覆盖治理测试失败
- **AND** 失败信息指出多余或缺失的 dispatcher 文件

### Requirement: guest host service client 必须使用注入式传输单轨

系统 SHALL 将动态插件 guest host service client 统一为 invoker 注入式结构。除`recordstore`等承载领域执行逻辑的独立 SDK 外，`pluginbridge`根目录 MUST NOT 保留逐域`pluginbridge_hostcall_*_wasip1.go`单例客户端、逐域 adapter 或逐域非 WASI 镜像 stub。`pluginbridge_directory.go`MUST 通过统一 invoker 装配基础能力和领域能力 guest client，wire 格式和 getter 签名 MUST 保持不变。

#### Scenario: 基础能力通过注入式 client 调用

- **WHEN** 动态插件 guest 通过能力目录访问 runtime、storage、cache、lock、host config、manifest 或 plugins config 能力
- **THEN** 对应 client 由`internal/domainhostcall`或等价内部子组件通过 invoker 构造
- **AND** 不依赖`pluginbridge`根目录包级 WASI 单例

#### Scenario: 根目录无逐域 host call 镜像残留

- **WHEN** 执行静态检索或治理测试检查`pluginbridge`根目录
- **THEN** 不存在逐域`pluginbridge_hostcall_*_wasip1.go`客户端文件
- **AND** 不存在逐域非 WASI mirror stub 或 adapter 文件
- **AND** 非 WASI 不可用行为收敛到传输层`InvokeHostService`stub

#### Scenario: wire 行为保持不变

- **WHEN** 迁移后的基础能力 client 发起 host service 调用
- **THEN** service/method 字符串、payload codec、字段编号、默认值和错误 envelope 与迁移前保持一致
- **AND** 现有`pluginbridge`协议测试和动态插件`wasip1`构建继续通过

#### Scenario: recordstore 执行文件不作为逐域镜像残留

- **WHEN** 检查`pluginbridge/recordstore`的 WASI 或 stub 执行文件
- **THEN** 这些文件可以保留为 record store 查询计划执行逻辑
- **AND** 它们不得重新引入`pluginbridge`根目录逐域客户端单例模式
