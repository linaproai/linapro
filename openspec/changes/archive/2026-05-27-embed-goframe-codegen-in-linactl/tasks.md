## 1. 依赖与边界确认

- [x] 1.1 读取并记录本次实现命中的规则文件：`openspec.md`、`dev-tooling.md`、`backend-go.md`、`api-contract.md`、`database.md`、`testing.md`、`documentation.md`。
- [x] 1.2 在`hack/tools/linactl/go.mod`中引入`github.com/gogf/gf/cmd/gf/v2`，并确认其版本与宿主`github.com/gogf/gf/v2`一致。
- [x] 1.3 运行`go mod tidy`或等价依赖整理，审查`go.mod`和`go.sum`变化只服务内嵌 GoFrame CLI。

## 2. 内嵌 GoFrame 执行组件

- [x] 2.1 重构`hack/tools/linactl/internal/goframecli`，拆分父进程分发和隐藏子进程执行边界。
- [x] 2.2 实现隐藏入口参数白名单，只允许`gen ctrl`和`gen dao`。
- [x] 2.3 在隐藏入口中通过`gfcmd.GetCommand(ctx)`和`RunWithSpecificArgs`执行等价`make ctrl`或`make dao`参数。
- [x] 2.4 确保隐藏入口执行 GoFrame CLI 前使用`apps/lina-core`作为工作目录。
- [x] 2.5 确保隐藏入口返回清晰错误，拒绝除`gen ctrl`和`gen dao`以外的 GoFrame CLI 子命令。

## 3. 命令改造

- [x] 3.1 改造`linactl ctrl`，移除`runCLIInstallIfMissing`和外部`gf`调用，改为启动隐藏 GoFrame 入口执行`gen ctrl`。
- [x] 3.2 改造`linactl dao`，移除`runCLIInstallIfMissing`和外部`gf`调用，改为启动隐藏 GoFrame 入口执行`gen dao`。
- [x] 3.3 新增或调整命令注册能力，使隐藏 GoFrame 入口不出现在默认`help`中；若新增`Hidden`字段，补齐 help 行为。
- [x] 3.4 删除或停用`linactl cli`和`linactl cli.install`的外部`gf`下载路径，确保`ctrl`/`dao`默认路径不再引用 GitHub release。

## 4. 测试与验证

- [x] 4.1 新增或更新`linactl`单元测试，验证`ctrl`分发到隐藏 GoFrame 入口且不调用外部`gf`。
- [x] 4.2 新增或更新`linactl`单元测试，验证`dao`分发到隐藏 GoFrame 入口且不调用外部`gf`。
- [x] 4.3 新增隐藏入口白名单测试，覆盖允许`gen ctrl`、允许`gen dao`、拒绝非生成命令。
- [x] 4.4 新增 help 行为测试，确认隐藏入口不会出现在默认用户可见命令列表中。
- [x] 4.5 增加最小 controller 生成 smoke 或等价验证，确认无`gf`可执行文件时内嵌路径仍可生成 controller 骨架。
- [x] 4.6 评估`dao`生成验证策略；若不运行真实数据库生成，必须记录数据库依赖原因，并至少验证分发路径和错误边界。
- [x] 4.7 运行`cd hack/tools/linactl && go test ./... -count=1`。
- [x] 4.8 运行`openspec validate embed-goframe-codegen-in-linactl --strict`。

## 5. 文档与影响记录

- [x] 5.1 更新`hack/tools/linactl/README.md`和`README.zh-CN.md`，说明`ctrl`/`dao`不再要求独立安装`gf`，并保留数据库前置条件说明。
- [x] 5.2 在任务记录或审查结论中记录影响分析：`i18n`无运行时影响、缓存无影响、数据权限无影响、数据库 schema 无影响、开发工具跨平台有影响、测试策略已覆盖。
- [x] 5.3 记录跨平台验证方式，确认默认开发路径不新增平台专属 shell、PowerShell 或系统命令依赖。

## 实施记录

- 规则读取：本次实现已读取`openspec.md`、`dev-tooling.md`、`backend-go.md`、`api-contract.md`、`database.md`、`testing.md`、`documentation.md`和`i18n.md`。主要命中开发工具与文档规则；后端、接口和数据库规则仅作为代码生成流程边界校验。
- 依赖版本：`hack/tools/linactl`显式依赖`github.com/gogf/gf/cmd/gf/v2 v2.10.1`和`github.com/gogf/gf/v2 v2.10.1`，与`apps/lina-core`的`github.com/gogf/gf/v2 v2.10.1`一致。已执行`go mod tidy`。
- 测试策略：`ctrl`和`dao`通过单元测试验证父进程只分发到当前`linactl`隐藏入口，不解析或执行外部`gf`；隐藏入口白名单覆盖允许`gen ctrl`、允许`gen dao`和拒绝其他命令；`help`测试确认`__goframe`不在默认和`--all`列表中，且`help __goframe`不可见；controller smoke 在清空`PATH`的子进程中运行隐藏入口并生成最小 controller 骨架。
- `dao`验证策略：真实`dao`生成依赖`apps/lina-core/hack/config.yaml`中配置的 PostgreSQL 可连接且 schema 已初始化。本次未在单测中启动真实数据库，改为覆盖父分发路径、白名单和错误边界，并在文档中保留数据库前置条件。
- 影响分析：`i18n`无运行时资源影响；缓存一致性无影响；数据权限无影响；数据库 schema 无影响；HTTP API 契约无影响；开发工具跨平台路径有影响，默认入口仍为 Go 标准库进程调用，不新增 shell、PowerShell 或系统命令依赖。
- 验证记录：已运行`cd hack/tools/linactl && go test ./... -count=1`、`cd hack/tools/linactl && go run . help`和`openspec validate embed-goframe-codegen-in-linactl --strict`，均通过。
