## Context

根插件服务当前同时包含公共契约、子组件 wiring、source/dynamic lifecycle facade、runtime upgrade facade、management list projection、startup auto-enable、host service 配置和大量测试 fixture。已有`internal/lifecycle`、`internal/management`、`internal/runtimeupgrade`、`internal/sourceupgrade`、`internal/testutil`等子组件，本次优先复用这些边界，不引入新的多层转发抽象。

## Goals / Non-Goals

**Goals:**

- 让根包文件数量和单文件体量更可控，尤其是测试文件。
- 合并同一职责下低行数且只有配置入口、状态转发或同类型方法职责的小文件。
- 将重复测试 helper 收敛到职责明确的位置，减少跨测试文件复制。
- 保持公开契约、插件运行语义和缓存一致性机制不变。

**Non-Goals:**

- 不重写插件生命周期、runtime reconciler 或 host service 协议。
- 不新增 HTTP API、前端页面、SQL、DAO 或生成代码。
- 不为了减少文件数把职责不同的子组件合并成大文件。
- 不为了测试便利扩大生产代码导出面。

## Decisions

### 决策 1：先做机械收敛，再做职责拆分

低行数文件合并只在同包同职责范围内执行，例如 host service 配置入口、runtime cache revision 和 tenant governance policy。大文件拆分以测试文件为主，生产大流程只做不改变控制流的文件级整理。

### 决策 2：根包保持 facade，子组件保持内部边界

根插件`Service`仍作为控制器、启动装配、cron、中间件和插件生命周期调用方的稳定入口。实现细节如果已经有清晰内部组件，则优先移动到已有`internal/<subcomponent>`；如果移动会引入循环依赖或新增仅透传抽象，则暂不移动。

### 决策 3：测试按被测职责拆文件，helper 单独收敛

测试文件命名必须与被测源码或主题职责关联。大测试文件拆分后，每个测试仍自行构造数据、依赖替身和清理逻辑。共享 helper 只放在`*_test.go`测试支撑文件或`internal/testutil`，并由当前测试显式调用。

## Risks / Trade-offs

- 大规模移动代码可能掩盖行为回归，因此本次避免改变生命周期和 runtime upgrade 的控制流。
- 将 helper 下沉到`internal/testutil`可能扩大测试支撑面，因此本轮只把依赖根包私有类型的 helper 收敛到根包测试支撑文件。
- 文件拆分会产生较大的 diff，但验证重点仍是`go test`和 OpenSpec 严格校验。

## Validation

- `cd apps/lina-core && go test ./internal/service/plugin -count=1`
- `cd apps/lina-core && go test ./internal/service/plugin/internal/... ./internal/service/plugin/runtimecache -count=1`
- `openspec validate refactor-plugin-service-layout --strict`
- `git diff --check`
