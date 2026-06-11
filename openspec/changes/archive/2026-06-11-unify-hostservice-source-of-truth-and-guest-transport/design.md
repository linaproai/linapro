## Context

`pluginbridge/internal/hostservice/hostservice_descriptor.go`已经以结构化表记录 service、method、资源类型、capability、请求响应 payload、guest client 和 dispatcher 发布状态。现有测试已经覆盖 descriptor 到部分 protocol、guest 和 dispatcher 的正向关系，但仍存在三个缺口：

- `apps/lina-core/pkg/plugin/README.md`与`README.zh-CN.md`中的 host service 表格靠手写维护，容易与 descriptor 漂移。
- 治理测试还没有完全反向校验宿主 service 级 switch、dispatcher 文件集合和 guest selector 集合，孤儿 dispatcher 或多注册可能漏过。
- guest 侧存在两种客户端装配模式：大部分领域使用`internal/domainhostcall`注入式 invoker，部分基础能力仍保留根目录`pluginbridge_hostcall_*_wasip1.go`单例客户端、adapter 或镜像 stub。

本变更是纯结构重构和治理增强。项目无兼容性负担，不保留旧内部文件布局；但动态插件作者可见的`pluginbridge.Services`getter、host service wire、payload codec 和宿主运行时行为必须保持不变。

## Goals / Non-Goals

**Goals:**

- 将 descriptor 作为 host service 文档表格、guest 覆盖和 dispatcher 覆盖的单一事实源。
- 为双语`README`的 host service 表格建立可重复生成和漂移测试。
- 将仍残留的根目录逐域 host call WASI 客户端迁移到`internal/domainhostcall`注入式构造。
- 删除根目录逐域 WASI 单例、adapter 和镜像 stub 残留，保持传输层 stub 单点。
- 通过`go test`、静态检索和动态插件 wasip1 构建验证 wire 与 guest 编译闭包不回归。

**Non-Goals:**

- 不引入`.proto`或新的 codec 生成链，不替换现有手写`protowire`payload codec。
- 不修改 host service 的 service/method 字符串、字段编号、默认值、错误 envelope 或授权快照语义。
- 不重写宿主`internal/service/plugin/internal/wasm`dispatcher 的业务处理逻辑；本轮只增强覆盖治理。
- 不迁移`recordstore`执行文件；它已经是 invoker 注入式结构，WASI/stub 文件承载查询计划执行领域逻辑，不属于逐域客户端镜像。
- 不新增 HTTP API、数据库、前端页面或运行时`i18n`资源。

## Decisions

### D-C1：README host service 表格由 descriptor 生成

新增`pluginbridge/internal/hostservice`README 渲染器，直接复用 descriptor 渲染中英文 host service 表格。两份`README`只保留固定手写章节和带标记的生成区块：

- `<!-- BEGIN generated:host-services -->`
- `<!-- END generated:host-services -->`

渲染器负责将`ResourceKind`、reserved capability 和混合资源说明映射为稳定的英文/中文表格文本。漂移测试调用同一渲染逻辑，比对当前`README`内容；若 descriptor 或映射变化后未刷新文档，测试失败并提示从 descriptor 渲染器更新生成区块。

不保留独立`go run`生成入口、shell 脚本或默认开发命令；漂移治理由 Go 测试内的 descriptor 渲染器完成，避免引入额外长期维护入口。

**备选方案：**继续手写双语表格，只依赖 review。该方案无法解决 drift 问题，因此拒绝。

### D-C2：descriptor 覆盖治理改为双向

在现有 AST 扫描测试基础上补齐反向校验：

- descriptor 中`Dispatcher=true`的方法必须存在于宿主 dispatcher selector。
- 宿主 dispatcher selector 中出现的 service/method 必须存在于 descriptor 且声明`Dispatcher=true`。
- 宿主 service 级 switch 的 service 集合和`dispatchXxxHostService`文件集合必须与 descriptor 的 dispatcher service 集合一致。
- descriptor 中`GuestClient=true`的方法必须存在于 guest client selector；guest selector 中出现的方法也必须存在于 descriptor 且声明`GuestClient=true`。

这使新增、删除或重命名 host service method 时，遗漏任一同步点都会在`go test`阶段失败。

**备选方案：**生成宿主 dispatcher 或 guest client 全部代码。本轮只需要收敛事实源和测试治理；全量生成会扩大改造面并提高 wire 行为回归风险，因此拒绝。

### D-C3：guest 客户端统一为注入式 domainhostcall

根目录基础能力客户端迁入`internal/domainhostcall`，形态与现有领域客户端一致：客户端构造函数接收 invoker，通过统一 host service envelope 发起调用。`pluginbridge_directory.go`只负责把同一 invoker 注入能力目录，不再依赖包级 WASI 单例。

热路径能力继续复用现有`protocol`payload codec 和 wire 格式；低频能力继续使用 JSON 信封。双编码是协议层历史取舍，本变更不改变；本轮只消灭双客户端装配模式。

删除根目录逐域`pluginbridge_hostcall_*_wasip1.go`、adapter 和镜像 stub 后，非 WASI 构建只通过传输层`InvokeHostService`stub 暴露不可用行为。测试从硬编码期待逐域 stub 文件，改为断言根目录无逐域 host call WASI/stub 残留。

**备选方案：**一次性把所有 payload codec 改为 JSON 或生成式 protobuf。该方案会改变性能边界或引入新工具链，不符合本轮纯结构重构目标，因此拒绝。

## Risks / Trade-offs

- `Risk`：README 生成器映射文本与现有文档事实不一致。`Mitigation`：首次生成后人工检查中英文表格，漂移测试固定输出；双语 README 保持同一结构和事实。
- `Risk`：迁移基础能力客户端时误改 wire payload 或错误 envelope。`Mitigation`：复用现有`protocol`codec 和 invoker；运行`pluginbridge`协议测试、动态插件样例普通构建和`wasip1`构建。
- `Risk`：双向 AST 测试过度依赖文件命名。`Mitigation`：只约束宿主 dispatcher 当前已稳定的`dispatchXxxHostService`结构，并在失败信息中指出具体 service/method 或文件。
- `Risk`：README 渲染器被误当作独立开发工具。`Mitigation`：不保留独立`go run`入口、脚本或默认开发命令；验证记录只依赖 Go drift 测试和静态检索。

## Migration Plan

1. 新增 descriptor README 渲染器、生成区块和漂移测试，不保留独立生成入口。
2. 扩展 descriptor 覆盖测试为正反向治理。
3. 按基础能力逐步迁移根目录 host call 客户端到`internal/domainhostcall`，保持 getter 签名和 wire 格式不变。
4. 删除根目录逐域 WASI 单例、adapter 和镜像 stub 残留，更新测试期望。
5. 运行`go test ./pkg/plugin/pluginbridge/... -count=1`、`go test ./pkg/plugin/... -count=1`、动态插件样例构建、静态检索和`openspec validate`。

回滚策略为整体`git revert`；不涉及数据迁移、配置迁移或运行时状态迁移。

## Open Questions

无需要用户澄清的问题。实现阶段只需以当前代码为准确认仍残留的根目录逐域 host call 文件清单，并避免触碰`recordstore`领域执行文件。
