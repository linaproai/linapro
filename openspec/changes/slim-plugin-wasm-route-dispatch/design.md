## Context

`localdocs/plugin-service-complexity-refactor-plan.md`的方案 E 定位为可选优化，前置的 host service registry 与插件升级治理变更已经归档。归档后的基线显示：

- `apps/lina-core/internal/service/plugin/internal/runtime/route.go`为`995`行，仍同时承载动态路由匹配、JWT/session 鉴权、角色菜单权限查询、请求 envelope、header/cookie 适配和响应写回。
- `apps/lina-core/internal/service/plugin/internal/wasm`已有`hostservicedispatch`registry 和`wasm_host_service_registry.go`，但`capabilityContextForHostCall`仍位于`wasm_host_service_users.go`，公共 helper 与具体领域文件耦合。
- 工作区存在前序方案未提交变更，本变更只在新 OpenSpec 目录和方案 E 目标文件上追加最小改动，不回退或重排前序变更。

## Goals / Non-Goals

**Goals:**

- 让 WASM host service 公共 helper 位于公共层，避免领域文件承载跨领域语义。
- 让新增普通领域 host service 的维护路径保持为“领域文件 + 显式 registry 注册”，不新增第二套分发抽象。
- 将动态路由分发代码按职责拆分，使`route.go`保留入口和核心编排，文件行数不超过`400`。
- 用静态测试固化边界，避免后续重构回退。

**Non-Goals:**

- 不改变动态插件公开路由路径、内部 route contract、HTTP API 契约或权限标签。
- 不改变 WASM guest ABI、host service service/method 字符串、请求响应 envelope 或错误 envelope。
- 不迁移所有领域 handler 到新子包，不重写现有 host service 领域逻辑。
- 不处理方案 B 中 WASM runtime 包级实例字段化议题。
- 不修改 SQL、前端页面、业务插件目录或`pkg/plugin`公开契约；仅审查`pkg/plugin`README 是否需要同步。

## Decisions

### 1. 公共 helper 迁回`wasm`公共层

将`capabilityContextForHostCall`放到`wasm_host_service.go`或同包公共 host service 文件中，继续保持小写未导出函数。这样可避免为一个 helper 新增导出面，也不会为了迁移到`hostservicedispatch`而扩大`hostCallContext`、插件快照或运行态结构的公开范围。

备选方案是迁入`hostservicedispatch`子包并导出窄上下文接口。当前 helper 直接依赖`wasm`包内部`hostCallContext`，为了单个 helper 抽接口会增加额外适配层，因此不采用。

### 2. registry 模式只补边界，不新增分发框架

继续使用现有`hostservicedispatch.Registry`和显式注册函数。方案 E 只清理公共 helper 归属，并通过测试确认入口文件不恢复 service 级大 switch。仍保留各领域文件内的 method switch，除非后续某个领域存在明确重复样板和测试覆盖，才做小步迁移。

备选方案是一次性把全部领域 method 改成 method registry。该方案风险高、触达面大，并会与现有已经可工作的 registry 结构产生重复抽象，因此不纳入本次。

### 3. `route.go`拆为同包职责文件

保持`runtime`包不新增子包，避免内部类型和测试夹层扩散。拆分方向如下：

| 文件 | 职责 |
|------|------|
| `route.go` | 动态路由入口、核心 dispatcher 编排、与 executor 交互 |
| `route_match.go` | route contract 匹配、公开路径解析、path pattern 编译与参数提取 |
| `route_auth.go` | JWT/session 解析、访问级别判断、角色菜单权限查询、access context 构造 |
| `route_envelope.go` | bridge request envelope、header/cookie/query/body 适配 |
| `route_response.go` | bridge response 写回、header/cookie/status/body 处理 |

备选方案是建立`runtime/internal/route`子包。当前代码大量使用`runtime`内部类型，立即下沉会迫使导出更多中间结构；同包拆分即可满足复杂度治理和行数约束。

## Risks / Trade-offs

- 路由拆分可能遗漏原有鉴权或响应细节 → 通过现有`route_dispatch_test.go`和新增静态测试覆盖，并运行`go test ./internal/service/plugin/internal/runtime -count=1`。
- 文件拆分不改变行为但可能影响编译依赖 → 不新增运行期依赖，不改构造函数；用变更包 Go 测试验证。
- WASM helper 迁移可能造成测试包引用不一致 → 保持同包未导出函数名称不变，只移动定义位置，并运行`go test ./internal/service/plugin/internal/wasm ./internal/service/plugin/internal/wasm/hostservicedispatch -count=1`。
- `route.go`行数治理可能被注释或空行影响 → 使用测试读取文件并按实际行数断言，验收口径透明。

## Migration Plan

1. 创建增量规范和任务，运行`openspec validate slim-plugin-wasm-route-dispatch --strict`。
2. 移动 WASM 公共 helper，补充静态测试，运行 WASM 相关包测试。
3. 拆分`runtime/route.go`，保持函数名和调用链尽量不变，补充行数与职责边界测试，运行 runtime 包测试。
4. 运行`git diff --check`和 OpenSpec 校验，完成`lina-review`审查。

回滚策略为恢复本变更新增文件和移动的函数定义；由于不改变存储、协议或公开 API，不需要数据迁移或兼容开关。

## Open Questions

无。若实现过程中发现某个领域 host service method switch 存在高风险重复样板，本变更只记录为后续任务，不在方案 E 当前范围中扩大迁移。
