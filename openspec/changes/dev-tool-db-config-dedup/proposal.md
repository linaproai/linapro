## Why

当前 `apps/lina-core/hack/config.yaml` 在开发态工具链中重复维护数据库连接配置，`database.default.link` 与 `gfcli.gen.dao[].link` 容易产生漂移，也让后续为多数据库场景收敛开发态连接参数变得困难。与此同时，本地 `init` / `mock` SQL 执行路径当前依赖 MySQL DSN 中的 `multiStatements=true`，这会把命令行为绑定到特定驱动能力，不利于统一开发态配置和后续扩展数据库适配。

## What Changes

- 收敛宿主开发态工具配置，将 `apps/lina-core/hack/config.yaml` 中重复的数据库连接参数统一为 YAML anchor 复用形式。
- 移除宿主开发态工具配置中的 `multiStatements=true`，统一 `database.default.link` 与 `gfcli.gen.dao[].link` 的数据库连接方式。
- 调整本地 `init` / `mock` SQL 执行逻辑，使其在不依赖驱动多语句执行能力的前提下仍能按文件顺序、失败即停地执行交付 SQL。
- 补充对应的单元测试，覆盖 SQL 拆分执行、空白/注释处理以及错误中断语义，确保开发态工具脚本调整后行为稳定。

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `database-bootstrap-commands`: 调整本地 SQL 启动命令的执行约束，使开发态 `local` SQL 源不再依赖连接串中的 `multiStatements` 能力，同时保持顺序执行与失败即停语义。

## Impact

- Affected code: `apps/lina-core/hack/config.yaml`, `apps/lina-core/internal/cmd/`, related command unit tests.
- Affected systems: host development-only bootstrap/tooling flow (`make init`, `make mock`, `gf gen dao`).
- Dependencies: GoFrame command/database access path, local SQL parsing/execution helper logic.
