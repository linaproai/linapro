## Context

`linactl`是仓库默认跨平台开发命令入口。当前`linactl ctrl`和`linactl dao`位于`hack/tools/linactl`中，但实际流程仍依赖外部`gf`可执行文件：先检查本机`gf -v`，缺失时从 GitHub release 的`latest`地址下载并安装，再在`apps/lina-core`目录执行`make ctrl`或`make dao`。

这个流程有三个问题：

- 开发者首次执行代码生成时需要依赖外部二进制安装，体验和`linactl`统一工具入口不一致。
- 下载地址使用`latest`，生成器版本可能与仓库锁定的`github.com/gogf/gf/v2`版本不一致。
- GoFrame CLI 生成器本身具有命令行进程入口语义，包含`Fatalf`、`os.Exit`和全局配置搜索等副作用，不适合直接作为普通库函数嵌入`linactl`主进程。

GoFrame CLI 的公开 module 是`github.com/gogf/gf/cmd/gf/v2`，其公开入口为`gfcmd.GetCommand(ctx)`。真正的`genctrl`和`gendao`对象位于 GoFrame CLI 的`internal`包内，`linactl`不能也不应直接 import。

## Goals / Non-Goals

**Goals:**

- 让`make ctrl`、`make dao`、`linactl ctrl`和`linactl dao`在默认开发路径中不再要求本机安装`gf`。
- 通过`hack/tools/linactl/go.mod`锁定 GoFrame CLI module 版本，并与宿主 GoFrame runtime 版本保持一致。
- 复用 GoFrame 官方 CLI 代码生成能力，不重新实现 controller、DAO、DO、Entity 生成器。
- 通过隐藏子进程隔离 GoFrame CLI 的`Fatalf`、`os.Exit`、全局配置和工作目录副作用。
- 保持`apps/lina-core`作为 GoFrame 代码生成工作目录，继续读取`apps/lina-core/hack/config.yaml`中的`gfcli`配置。
- 用单元测试覆盖父命令分发、隐藏命令白名单、工作目录和“不依赖外部`gf`”行为。

**Non-Goals:**

- 不改变 GoFrame 生成文件格式，不修改`dao`、`do`、`entity`或 controller 骨架模板。
- 不新增数据库 schema，不调整`make init`、数据库连接配置或`gfcli.gen.dao`配置语义。
- 不把完整 GoFrame CLI 暴露为公开`linactl`代理命令。
- 不支持插件目录的 GoFrame 代码生成；本次只覆盖当前已有的宿主`ctrl`和`dao`入口。
- 不改变`api/`定义、HTTP API 契约、路由绑定或后端业务逻辑。

## Decisions

### Decision: 使用 GoFrame CLI module 依赖而不是外部二进制

`hack/tools/linactl` SHALL 显式依赖`github.com/gogf/gf/cmd/gf/v2`，版本与`apps/lina-core`使用的`github.com/gogf/gf/v2`保持一致。`ctrl`和`dao`默认路径不得下载、安装或调用`PATH`中的`gf`。

备选方案是继续自动安装外部`gf`，但该方案保留了用户机器状态依赖和`latest`版本漂移，因此不采用。

### Decision: 使用隐藏子进程执行内嵌 GoFrame CLI

父命令`linactl ctrl`和`linactl dao`只负责分发到同一个`linactl`程序的隐藏内部命令，例如`__goframe gen ctrl`和`__goframe gen dao`。隐藏子进程内部 import`gfcmd`，构造 GoFrame CLI root command，并调用`RunWithSpecificArgs(ctx, []string{"gf", "gen", "ctrl"})`或`RunWithSpecificArgs(ctx, []string{"gf", "gen", "dao"})`。

选择子进程隔离的原因：

- GoFrame CLI 失败路径可能直接`os.Exit(1)`，子进程可以把退出限制在代码生成执行层。
- 父`linactl`仍然可以统一收集 stdout、stderr 和 exit code，并保持现有错误包装。
- 测试可以通过注入 runner 验证父命令参数，不需要让单测进程承受 GoFrame CLI 的退出副作用。
- 隐藏命令仍由`linactl`自身提供，不要求用户安装额外工具。

备选方案是在`linactl`主进程里直接调用`gfcmd.Command.Run(ctx)`或`RunWithSpecificArgs`。该方案代码路径更短，但会让 GoFrame CLI 的进程入口语义进入`linactl`主进程，错误和全局状态边界不清晰，因此只作为 spike 参考，不作为最终设计。

### Decision: 隐藏命令只允许代码生成白名单

隐藏 GoFrame 入口 MUST 只接受以下命令：

- `gen ctrl`
- `gen dao`

其他 GoFrame CLI 子命令，例如`install`、`build`、`docker`、`run`、`pack`、`env`和任意未知命令，MUST 返回错误。这样可以避免`linactl`无意中成为完整 GoFrame CLI 代理，扩大维护边界。

### Decision: 保持`apps/lina-core`为生成工作目录

父命令启动隐藏子进程时 SHALL 将工作目录设置为仓库根目录下的`apps/lina-core`，或由隐藏命令在执行 GoFrame CLI 前显式切换到该目录。实现必须保证 GoFrame CLI 查找`hack/config.yaml`、`api/`、`internal/`和`go.mod`时与现有外部`gf`流程一致。

### Decision: 删除或停用外部`gf`安装路径

`linactl ctrl`和`linactl dao`不得再调用`runCLIInstallIfMissing`。原有`linactl cli`和`linactl cli.install`下载 GitHub release 的路径应删除；如果实现阶段选择暂时保留兼容命令，也必须标记为非默认路径，并且不得被`ctrl`/`dao`调用。考虑本项目无历史负担，推荐直接删除。

## Risks / Trade-offs

- [风险] 新增`github.com/gogf/gf/cmd/gf/v2`会增加`linactl`依赖体积和首次`go run`编译耗时。  
  缓解：版本锁定到与 runtime 相同的 GoFrame 版本；实现后用`cd hack/tools/linactl && go test ./... -count=1`和一次`go run . help`/`ctrl` smoke 记录验证结果。

- [风险] GoFrame CLI 内部依赖多个数据库 driver，可能带来间接依赖升级冲突。  
  缓解：在`hack/tools/linactl/go.mod`中显式 pin CLI module；运行`go mod tidy`并审查`go.sum`变化；后续 GoFrame 升级时同步升级 runtime 和 CLI module。

- [风险] 子进程分发如果使用当前可执行文件路径处理不当，`go run . ctrl`和已编译二进制的行为可能不一致。  
  缓解：统一封装当前执行入口解析；必要时支持在`go run`模式下通过`go run . __goframe ...`回调自身，避免依赖平台专属 shell。

- [风险] 隐藏命令被用户直接调用后误以为是公开 API。  
  缓解：命令不出现在默认 help；实现白名单；错误消息说明该入口仅供`linactl ctrl/dao`内部使用。

- [风险] `linactl dao`仍然需要 PostgreSQL 可连接且 schema 已初始化，用户可能把数据库失败误解为 GoFrame CLI 嵌入失败。  
  缓解：保持错误输出中的 GoFrame 原始 stderr，同时在`linactl`文档中说明`dao`仍需先运行`make init confirm=init`或准备等价数据库。

## Migration Plan

1. 在`hack/tools/linactl`中引入 GoFrame CLI module 依赖，并同步 GoFrame runtime 版本。
2. 扩展`internal/goframecli`，提供父进程分发和隐藏子进程执行两个边界。
3. 新增隐藏`__goframe`命令并实现`gen ctrl`/`gen dao`白名单。
4. 改造`command_ctrl.go`和`command_dao.go`，移除外部`gf`安装和调用路径。
5. 删除或停用`cli`、`cli.install`下载外部`gf`的命令路径。
6. 更新`hack/tools/linactl`中英文说明文档。
7. 补充单元测试和 smoke 验证。

回滚策略：如果内嵌 GoFrame CLI 在目标平台存在不可接受的依赖或启动问题，可以恢复`ctrl`/`dao`调用外部`gf`的实现，但必须避免`latest`下载，改为仓库锁定版本的安装方式。

## Open Questions

- 隐藏命令是否应完全不出现在`help --all`中？当前倾向新增`Hidden`字段并完全隐藏。
- 外部`cli`/`cli.install`命令是否直接删除？当前项目无历史负担，推荐删除。
