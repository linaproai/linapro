## Context

`linactl lint.go`当前流程：

1. 根据`plugins`准备宿主或官方插件完整工作区环境
2. `go list -m`列出工作区全部 module
3. 对每个 module 跑`golangci-lint run ./...`与`staticcheck U1000`（含 guest 敏感矩阵）

`build`/`ctrl`/`dao`等命令已支持`dir=`定向；lint 尚未对齐该语义。

## Goals / Non-Goals

**Goals:**

- 支持`dir=<path>`将 lint 收敛到单个 Go module
- 插件根目录自动落到其 backend module（若存在`backend/go.mod`）
- 失败时给出清晰错误：路径不存在、找不到 go.mod、module 不在当前 workspace
- 日志标明定向范围
- 默认无`dir`行为与 CI 兼容、不变

**Non-Goals:**

- 包级`packages=`筛选（U1000 跨包语义不可靠，留作后续）
- 基于 git diff 自动推断范围
- 改变 CI 默认全量门禁
- 允许多个`dir`一次传入（首期单路径即可）

## Decisions

### Decision 1：定向粒度 = Go module，而非包

- **选择**：`dir`解析到最近的`go.mod`所在目录，只对该 module 执行现有`goLintModulePlan`
- **原因**：module 级仍覆盖`./...`，staticcheck U1000 与 golangci-lint 语义完整；改一个组件通常落在一个 module
- **备选**：包级路径 → 开发更快但死代码检查易误报/漏报，本期不做

### Decision 2：`dir`与`plugins`关系

- **选择**：仍先按`plugins`准备 workspace；再在 workspace modules 中过滤出`dir`对应 module
- **原因**：插件 module 依赖`temp/go.work.plugins`；若跳过插件环境准备，插件定向 lint 可能解析失败
- **规则**：
  - `dir=apps/lina-core`或`hack/tools/linactl`：可用`plugins=0`
  - `dir=apps/lina-plugins/...`：若当前 workspace 不含该 module，失败并提示使用`plugins=1`或初始化插件工作区

### Decision 3：路径解析规则

1. 将`dir`解析为绝对路径（相对仓库根）
2. 必须存在且为目录
3. 若目录含`plugin.yaml`且存在`backend/go.mod`，优先使用`backend`
4. 否则从该目录向上查找最近`go.mod`（不超过仓库根）
5. 将解析到的 module 目录与`go list -m`结果做路径等价匹配（`EvalSymlinks`）
6. 匹配失败则报错，不静默回退全量

### Decision 4：公开入口

```bash
make lint dir=apps/lina-core
make lint.go dir=apps/lina-plugins/<id>/backend plugins=1
linactl lint.go dir=hack/tools/linactl plugins=0
```

`lint.mk`透传`dir`，逻辑只在`linactl`中。

### Decision 5：日志

plan 行增加`scope=workspace|dir`与`dir=<rel>`（定向时），summary 同样标识，避免审查误读。

## Risks / Trade-offs

- **[Risk] 定向 lint 通过但其他 module 仍有问题** → Mitigation：文档与规则写明 CI/PR 仍跑全量；Agent 约定“变更涉及的全部 module 至少各跑一次”
- **[Risk] 插件根路径误解析到 monorepo 上级 go.mod** → Mitigation：插件根优先 backend；向上查找止于 repo root
- **[Risk] 符号链接路径不一致** → Mitigation：复用现有`goLintCanonicalDir`做匹配

## Migration Plan

- 纯增量能力，无迁移
- 不传`dir`行为完全兼容
- 回滚：移除`dir`解析逻辑即可

## Open Questions

- 无。包级筛选明确延后。
