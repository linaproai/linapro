## Why

LinaPro 后端曾同时存在 API 合同漂移、响应边界过宽、GoFrame v2 用法不一致、运行期依赖隐式构造、事务错误吞掉、启动期重复 SQL、文件职责混杂和单元测试耗时过高等系统性质量问题。这些问题会让接口对外暴露内部字段、让缓存敏感服务出现实例分裂、让归档和审查难以判断真实契约，也会在项目持续交付过程中放大维护成本。

项目没有历史兼容包袱，因此历史整改选择直接把 API 合同、后端分层、显式依赖注入、宿主运行能力、启动 SQL 效率、i18n 相关前端状态刷新和 Go 单元测试效率统一纳入代码质量基线，并通过 OpenSpec、规则文件、静态扫描、单元测试、E2E 和`lina-review`形成可持续治理闭环。

此外，OpenSpec 归档文档在初始聚合后存在跨分组重复承载最终能力契约的问题，归档体量约`3.5M`、归档`spec.md`为`277`个、跨分组重复能力为`61`个，导致维护者和 AI 难以快速判断当前契约与历史信息的边界。同时，`lina-review` 对 E2E 的审查重点偏向文件组织和编号等合规项，缺少对覆盖有效性、断言有效性和测试可信度的结果级质量判断。

## What Changes

- 统一 API REST 语义、路径参数绑定、DTO 文档标签、批量删除、响应 DTO、瞬时时间字段毫秒时间戳和跨模块稳定枚举契约。
- 收敛 GoFrame v2 ORM、软删除、事务、`panic`、主文件职责、接口注释、接口方法唯一性、文件顶部注释和公共组件主文件治理。
- 将 Controller、Middleware、Service、源码插件、插件宿主适配器和 WASM host service 的运行期依赖改为构造函数逐项显式注入，禁止聚合依赖结构体和隐式关键服务图。
- 建立匿名健康检查、GoFrame 优雅停机复用、上传文件受保护访问、调度默认时区配置、事务一致性、关键索引和菜单子树内存遍历等宿主运行质量基线。
- 引入启动编排内共享 startup snapshot、差异驱动插件同步、无差异零写入路径、内置任务投影复用和结构化启动摘要，降低启动 SQL 噪声和重复装配成本。
- 优化语言切换、监控轮询、消息轮询和路由缓存等前端运行路径，避免语言切换重拉权限菜单、隐藏页签继续轮询或长会话缓存无界增长。
- 建立 Go 单元测试效率规则：保留`-race`主路径，收敛真实 dynamic Wasm 执行为 smoke，复用轻量 fixture，减少无测试包完整执行并输出耗时摘要。
- 建立 OpenSpec 归档文档治理规则：信息分层、能力 owner 映射、重复规范裁剪、分阶段压缩、语义覆盖和验证报告要求，将归档体量从约`3.5M`压缩至约`1.5M`。
- 增强`.agents/rules/testing.md`，新增 E2E 质量审查的结果级要求，覆盖有效性、断言有效性、稳定性、异常路径和审查输出，不绑定具体测试实现方式。
- 解耦官方源码插件`init`注册 fail-fast 与宿主 panic allowlist：`apps/lina-core`继续精确白名单，`apps/lina-plugins/*/backend/plugin.go`的约定 fail-fast 按 AST 模式自动归类，非常规插件 panic 仍失败。
- 将宿主能力约束类错误文案改为能力语义（如 multi-tenant governance），错误码与`messageKey`保持稳定，用户可见消息不得绑定官方插件品牌 ID。

## Capabilities

### New Capabilities

- `host-runtime-operations`：匿名健康检查、GoFrame 优雅停机复用、受保护上传访问、调度默认时区和 HTTP 入口职责拆分。
- `startup-sql-efficiency`：启动快照复用、差异驱动同步、无差异零写入、内置任务投影复用和启动摘要日志。
- `service-dependency-injection-governance`：显式依赖注入、共享实例、初始化错误返回、聚合依赖结构体禁止和静态治理扫描。
- `go-unit-test-execution-efficiency`：Go 单测分层、真实 Wasm smoke 边界、fixture 复用、`-race`保留、无测试包规划和耗时摘要。
- `openspec-archive-document-governance`：OpenSpec 归档文档的信息分层、能力 owner 映射、重复规范裁剪、分阶段压缩、语义覆盖和验证报告要求。
- `runtime-message-i18n-governance`：宿主运行时错误文案能力化，禁止官方插件品牌 ID 进入用户可见约束类消息。

### Modified Capabilities

- `api-contract-consistency`：REST 语义、参数绑定、文档标签、响应 DTO 加固、毫秒时间戳和公共枚举契约。
- `backend-conformance`：GoFrame ORM/软删除、事务、`panic`治理（含官方插件`init`模式自动放行）、主文件职责、接口注释、接口方法唯一性、文件顶部注释和`lina-review`后端检查。
- `framework-i18n-runtime-performance`：语言切换只刷新强语言相关本地状态，菜单和路由标题通过本地 i18n key 响应式更新。
- `e2e-suite-organization`：新增 E2E 质量审查的结果级要求和审查输出证据要求。

## Impact

- 历史实现曾影响宿主后端、源码插件后端、SQL、前端、E2E、OpenSpec 主规范和审查规则；当前归档压缩只保留这些质量治理的历史原因、决策和验证摘要。
- 当前能力契约以`openspec/specs/<capability>/spec.md`为准；本分组只保留代码质量 owner 能力的历史规范入口。
- 插件、用户/角色/菜单、调度、分布式缓存、监控、消息、模块禁用和 OpenSpec 流程治理等非 owner 能力只在`design.md`中保留交叉影响摘要，不再重复保存完整规范全文。
- 本次压缩不修改运行时代码、HTTP API、数据库、前端 UI、插件源码、语言包、`manifest/i18n`、`apidoc i18n JSON`、缓存实现或生产构建入口。
