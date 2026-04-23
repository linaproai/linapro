## Context

当前宿主开发态工具配置集中在 `apps/lina-core/hack/config.yaml`。该文件同时服务于本地 `init` / `mock` SQL 执行链路、`gf gen dao` 代码生成，以及仓库级升级工具读取的开发态元数据。现状中数据库连接参数在同一文件内重复声明，一处用于 `database.default.link`，另一处用于 `gfcli.gen.dao[].link`；两者一旦出现差异，开发态工具链就会产生隐性漂移。

本地 `init` / `mock` 在执行 SQL 文件时直接将整个文件内容透传给 `g.DB().Exec`。当前之所以能够执行包含多条语句的 SQL 文件，是因为开发态 DSN 额外打开了 `multiStatements=true`。该约束既让配置与 `gf gen dao` 的连接配置不一致，也把开发态命令行为绑定到 MySQL 驱动特性，不利于后续统一开发态数据库配置。

本次变更只处理开发态工具脚本配置与本地 SQL 执行路径，不涉及 `manifest/config/config.yaml` 等运行时配置文件，也不改变交付 SQL 内容或运行时默认资源来源。

## Goals / Non-Goals

**Goals:**
- 消除 `apps/lina-core/hack/config.yaml` 内部重复的数据库连接配置，确保开发态工具脚本共享同一套基础连接参数。
- 统一移除开发态工具链 DSN 中的 `multiStatements=true`，避免本地 SQL 启动命令依赖驱动级多语句能力。
- 保持 `init` / `mock` 的顺序执行、失败即停和错误定位语义。
- 为本次调整补齐可重复运行的单元测试，覆盖 SQL 拆分与执行边界。

**Non-Goals:**
- 不调整运行时配置文件结构或 `manifest/config/` 交付策略。
- 不扩展插件 `backend/hack/config.yaml` 的统一渲染机制。
- 不改写现有交付 SQL 文件的语义，也不引入新的数据库驱动或外部 SQL 解析依赖。

## Decisions

### 1. 在宿主 `hack/config.yaml` 中使用 YAML anchor 复用开发态数据库连接
- 在文件内定义一份共享数据库连接锚点，并由 `database.default.link` 与 `gfcli.gen.dao[].link` 共同引用。
- 共享连接串统一移除 `multiStatements=true`，使本地 SQL 执行与 `gf gen dao` 使用同一套基础连接参数。
- 这样可以在不引入额外脚本或模板渲染步骤的前提下，直接解决“同一文件内重复配置漂移”的问题。

**Alternatives considered:**
- 通过脚本或 Makefile 动态渲染 `hack/config.yaml`：扩展性更强，但对当前“只解决单文件重复”的目标来说过重。
- 继续保留两份 DSN，仅通过注释约束保持一致：无法从结构上防止再次漂移。

### 2. 在命令层显式拆分 SQL 语句，而不是依赖驱动多语句执行
- 在 `apps/lina-core/internal/cmd/` 中新增 SQL 拆分辅助逻辑，将单个 SQL 文件拆成有序语句列表后逐条执行。
- 拆分逻辑需要忽略空白片段，并正确处理常见注释与字符串字面量中的分号，避免误切分。
- `executeSQLAssetsWithExecutor` 的失败即停语义保持不变，但粒度从“整文件一次执行”收敛为“文件内逐语句执行，任一语句失败即终止整个文件与后续文件”。

**Alternatives considered:**
- 继续依赖 `multiStatements=true`：与本次统一开发态连接配置的目标冲突，也不利于后续多数据库适配。
- 引入外部 SQL parser：能力更强，但会增加依赖与维护成本，对当前交付 SQL 的复杂度而言收益不足。

### 3. 以命令层单元测试覆盖配置调整带来的行为变化
- 现有 `apps/lina-core/internal/cmd/cmd_test.go` 已覆盖确认值、资源来源和失败即停语义；本次继续在该层补充语句拆分与逐条执行测试。
- 测试重点放在：多语句文件按顺序执行、空白/注释片段跳过、语句中包含分号的字符串不被误拆、语句失败时立即中止。
- 如需验证 `hack/config.yaml` 的结构调整，则以轻量配置读取测试或 golden-style 断言覆盖 anchor 展开结果，避免依赖真实数据库。

## Risks / Trade-offs

- [Risk] 自定义 SQL 拆分逻辑遗漏边界场景，导致个别 SQL 文件被误切分。 → Mitigation：以当前交付 SQL 样式为基线编写针对性测试，并优先支持字符串、行注释、块注释等最常见语法边界。
- [Risk] 失败粒度从“整文件”变成“文件内语句”后，错误日志若缺少上下文会增加定位成本。 → Mitigation：在保留失败文件名的同时，为日志补充语句序号等上下文信息。
- [Risk] YAML anchor 虽能解决单文件重复，但后续若扩展到跨文件共享配置仍需额外方案。 → Mitigation：在设计中明确本次只解决宿主单文件重复，避免对跨文件复用形成错误预期。

## Migration Plan

1. 先调整 `apps/lina-core/hack/config.yaml`，用 YAML anchor 收敛数据库连接配置并移除 `multiStatements=true`。
2. 实现命令层 SQL 拆分与逐语句执行逻辑，确保在共享 DSN 下本地 `init` / `mock` 仍可运行。
3. 补充并运行 `apps/lina-core/internal/cmd` 相关单元测试，验证拆分、执行顺序与失败中断语义。
4. 如本地数据库环境可用，再补充执行一次开发态命令验证；若环境不可用，则以单元测试作为本次提交的必备验证结果。

## Open Questions

- 当前交付 SQL 是否存在 MySQL 特有的复杂语法边界（例如存储过程定义体内自定义分隔符）需要在拆分器中额外支持？若不存在，本次实现可先以常规 DDL/DML 文件语法为边界。
