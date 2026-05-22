# linactl

`linactl`是`LinaPro`的跨平台开发命令入口。它将仓库长期维护的任务编排放在`Go`工具中，确保`Windows`、`Linux`和`macOS`可以运行同一套命令，而不依赖`GNU Make`或`POSIX Shell`工具。

## 使用方式

```bash
cd hack/tools/linactl
go run . help
go run . status
go run . pack.assets
go run . wasm p=linapro-demo-dynamic
go run . wasm plugin_dir=/path/to/plugin out=temp/output
go run . plugins.status
go run . i18n.check
go run . init confirm=init
go run . tidy
go run . build platforms=linux/amd64,linux/arm64
go run . image tag=v0.2.0 push=0
go run . release.tag.check tag=v0.2.0
go run . release.tag.check print-version=1
```

## Windows 入口

仓库根目录提供`make.cmd`作为`Windows`薄包装入口：

```cmd
make.cmd help
make.cmd status
make.cmd pack.assets
make.cmd plugins.status
make.cmd i18n.check
make.cmd init confirm=init
make.cmd tidy
make.cmd release.tag.check tag=v0.2.0
```

在`PowerShell`中，需要显式添加当前目录前缀：

```powershell
.\make.cmd help
.\make.cmd status
.\make.cmd pack.assets
.\make.cmd i18n.check
.\make.cmd release.tag.check tag=v0.2.0
```

## 参数

`linactl`支持现有`make`风格的`key=value`参数，降低命令迁移成本。

| 参数 | 示例 | 用途 |
|------|------|------|
| `confirm` | `confirm=init` | 确认高风险初始化命令。 |
| `rebuild` | `rebuild=true` | 在`init`时重建配置中的数据库。 |
| `platforms` | `platforms=linux/amd64,linux/arm64` | 指定构建目标平台。 |
| `plugins` | `plugins=0` | 覆盖构建、开发、镜像和 Go 测试命令的自动插件完整模式探测。 |
| `tag` | `tag=v0.2.0` | 指定 `release.tag.check` 校验的 release tag。 |
| `print-version` | `print-version=1` | 输出已校验的 `framework.version`，供发布自动化使用。 |
| `p` | `p=linapro-tenant-core` | 为 Wasm 构建或插件工作区管理命令选择单个插件。 |
| `plugin-dir` | `plugin_dir=/path/to/plugin` | 从显式源码目录构建单个动态插件产物。 |
| `out` | `out=temp/output` | 指定动态插件产物输出目录。 |
| `source` | `source=official` | 为插件工作区管理命令选择单个已配置来源。 |
| `force` | `force=1` | 允许插件安装或更新命令覆盖已存在或存在本地改动的插件目录。 |
| `verbose` | `verbose=1` | 构建任务展示子命令输出。 |

未传入`plugins`时，构建和开发命令会在`apps/lina-plugins`存在插件清单时启用插件完整模式。插件完整模式会基于宿主专用的根目录`go.work`生成或刷新已忽略的`temp/go.work.plugins`，并通过`GOWORK`解析源码插件`Go`模块。

## 构建工具命令

`linactl`统一承载仓库镜像构建和动态插件`Wasm`打包实现。公开入口仍然是根目录`make`目标和对应的`linactl`命令：

```bash
make image tag=v0.2.0 push=0
make image.build tag=v0.2.0
make wasm p=linapro-demo-dynamic
```

当测试或本地夹具需要打包`apps/lina-plugins`之外的动态插件目录时，可以使用`plugin_dir=<path>`。

## 运行时 I18n 检查

`linactl i18n.check`统一承载运行时`i18n`治理检查。该命令会扫描高风险运行时可见硬编码文案，并校验宿主和插件运行时消息`key`覆盖：

```bash
make i18n.check
go run . i18n.check
```

默认扫描`allowlist`维护在`hack/tools/linactl/internal/runtimei18n/allowlist.json`。

## Agent 软链管理（agents.* 命令树）

`linactl agents.<resource>.<action>` 用于管理仓库内三类资源的本地软链，把 `.agents/`（以及 `AGENTS.md`）下的标准源映射到各 AI Coding 工具的私有项目路径：

- **skills**：目录级软链，`.<tool>/skills` → `.agents/skills`。受支持的 Agent 列表与 [vercel-labs/skills](https://github.com/vercel-labs/skills#supported-agents) 官方项目路径表一致。
- **prompts**：目录级软链，各 Agent 的 commands/prompts 根目录（例如 `.claude/commands`）→ `.agents/prompts`。
- **md**：单文件软链，`.<tool>.md`（或其他私有规范文件）→ 仓库根 `AGENTS.md`。

命令只在仓库根目录范围内操作，不会修改`HOME`或任何系统全局路径，也不会自动删除真实目录或文件（`force=1`同样不会）。

### 聚合入口（推荐用法）

聚合命令 `agents` 采用 **Agent 优先** 设计：选定一个 Agent，所选动作会自动作用到该 Agent 在 skills/prompts/md 三类资源中所有适用的绑定；对该 Agent 而言为 `native` 或未注册的资源会在最终摘要中显式列出跳过原因。

```bash
# 交互模式（终端）：
#   第 1 步：方向键选择 Agent（可输入字符过滤）。
#   第 2 步：方向键选择 `link` 或 `unlink`。
make agents

# 一键模式（CI/管道也可用）：
make agents agent=claude-code                    # 一次为 claude-code 在 skills + prompts + md 建立软链
make agents agent=ClaudeCode                     # 等价于 agent=claude-code
make agents agent=claude-code force=1            # 同时重建指向错误源的旧软链
make agents agent=claude-code action=unlink      # 移除 claude-code 的所有受管软链
```

`agent`必须是单个受支持 Agent 名称：聚合命令显式拒绝`agent=all`与逗号列表（批量场景请走下方子命令）。Agent 名称会归一化为标准`kebab-case`，所以`ClaudeCode`、`Claude Code`、`claude_code`和`claude-code`都会解析为`claude-code`。`action`默认为`link`。未传`agent`时，非终端环境会打印用法指引而不会阻塞等待输入。大写 Make 变量`AGENT`、`ACTION`和`FORCE`仍作为兼容别名保留，但新示例应优先使用小写名称，因为它们与`linactl`的`key=value`参数保持一致。

### 各资源子命令（高级用法）

推荐入口为聚合命令`make agents`。下列各资源子命令保留用于聚合命令显式不支持的批量场景，特别是`agent=all`与逗号列表。

```bash
# skills
make agents.skills.link                              # 终端下交互式选择；CI/管道下只读列表
make agents.skills.link agent=claude-code            # 非交互：为单个 Agent 创建软链
make agents.skills.link agent=claude-code,qoder      # 为多个 Agent 创建软链
make agents.skills.link agent=all                    # 为所有 link 类 Agent 创建软链
make agents.skills.link agent=all force=1            # 强制重建指向错误源的旧软链
make agents.skills.unlink                            # 终端下交互式选择（仅列出受管软链）
make agents.skills.unlink agent=claude-code          # 移除单个 Agent 的受管软链
make agents.skills.unlink agent=all                  # 移除所有受管软链

# prompts
make agents.prompts.link agent=claude-code           # 链接 .claude/commands -> .agents/prompts
make agents.prompts.link agent=all                   # 为所有受支持 Agent 创建 prompts 软链
make agents.prompts.unlink agent=claude-code         # 移除 prompts 软链

# md
make agents.md.link agent=claude-code                # 链接 CLAUDE.md -> AGENTS.md
make agents.md.link agent=all                        # 为所有 link 类 Agent 创建私有规范文件软链
make agents.md.unlink agent=claude-code              # 移除 AGENTS.md 软链
```

### 交互模式

所有交互入口（聚合命令`agents`与各`agents.<resource>.<action>`子命令）统一基于 [charmbracelet/huh](https://github.com/charmbracelet/huh) 的方向键交互：使用**方向键**移动、**空格**切换多选行、**回车**确认、**直接输入字符**快速过滤、**Esc** / **Ctrl+C**取消。CI 与管道环境保持非交互：`agents`打印用法指引，`agents.<resource>.link`退化为只读列表，`agents.<resource>.unlink`必须显式传入`agent=`。

候选项标题根据交互场景采用不同约定：

- 聚合命令 `agents` 的"选 Agent"是跨资源单选。每个选项只展示人类可读的 Agent 名称（例如 `Claude Code`、`Codex`、`Cursor`），保持选择列表简洁；确认后输出的结果表会列出每类资源是已应用还是已跳过。
- 各资源子命令 `agents.<resource>.<action>` 仅作用于单个资源，标题嵌入 **单字符状态符号** 与简短状态说明（形如 `[~] claude-code  (mismatch)`），便于直接看清当前绑定状态。

各资源子命令选项标题中嵌入的状态符号：

- `[+]` linked — 软链存在且指向标准源
- `[~]` mismatch — 软链存在但指向其他位置
- `[.]` absent — 尚未建立软链（或 `native`，无需操作）
- `[!]` conflict — 真实目录或文件阻止建立软链
- `[*]` root-collision — Agent 使用仓库根冲突路径（仅 skills 资源中的 `openclaw`）
- `[?]` error — 检测失败，详情请运行非交互列表

### 分类

- `native`：Agent 直接读取标准源，无需软链（例如 skills 中的 `cursor`、`gemini-cli`、`codex`；md 中所有原生读取 `AGENTS.md` 的 Agent）。
- `link`：Agent 使用其它项目路径，按需创建相对软链指向标准源。
- `rootCollision`：项目路径为仓库根的裸名（仅 skills 中的 `skills/`，由 `openclaw` 使用）。默认跳过；显式`agent=openclaw force=1`才创建。prompts 与 md 资源中不存在该分类。

> **md 资源的 fallback 行为说明**：部分 Agent 在私有规范文件（如 `CODEBUDDY.md`、`CLAUDE.md`）不存在时，会自动 fallback 读取 `AGENTS.md`。`CodeBuddy` 就是这样一个 Agent——根据腾讯官方文档，CodeBuddy 优先读取 `CODEBUDDY.md`，但当 `CODEBUDDY.md` 不存在时会自动加载 `AGENTS.md`。这类有官方文档支持的自动 fallback 机制的 Agent，在 md 注册表中按 `native` 注册，这样仓库 clone 即可用，无需建链；只有当 Agent 仅读取私有规范文件、不存在 fallback 路径时，才注册为 `link` 以便用户显式建链接入。每条 Agent 的证据来源都记录在 `internal/agents/md/md_agents.go` 的行内注释中。

任何情况下命令都不会自动删除已存在的真实目录或文件，包含`force=1`时也不会。`force=1`仅作用于"已是软链但指向其它位置"的情况。所有 skills 与 prompts 受管软链目录已在`.gitignore`中忽略，本地创建不会污染仓库。

### 从 `make skills.*` 迁移

旧的 `make skills` / `make skills.link` / `make skills.unlink` 目标，以及对应的 `linactl skills*` 子命令均已**删除**，被 `agents.*` 命令树取代。**没有保留任何别名**；现有脚本与文档必须更新：

| 已删除（不再生效） | 新命令 |
| --- | --- |
| `make skills` | `make agents` |
| `make skills.link` | `make agents.skills.link` |
| `make skills.link AGENT=<name>` | `make agents.skills.link agent=<name>` |
| `make skills.link AGENT=all FORCE=1` | `make agents.skills.link agent=all force=1` |
| `make skills.unlink` | `make agents.skills.unlink` |
| `make skills.unlink AGENT=<name>` | `make agents.skills.unlink agent=<name>` |
| `linactl skills` | `linactl agents` |
| `linactl skills.link` | `linactl agents.skills.link` |
| `linactl skills.unlink` | `linactl agents.skills.unlink` |

`agents.skills.*` 子命令的行为与原 `skills.*` 完全一致（同一注册表、同一状态机、同一 TTY/CI 行为），仅命令名称变化。

## Release Tag 校验

`release.tag.check` 会读取 `apps/lina-core/manifest/config/metadata.yaml`，并校验 release tag 与 `framework.version` 完全一致。

```bash
make.cmd release.tag.check tag=v0.2.0
make release.tag.check tag=v0.2.0
make release.tag.check metadata=apps/lina-core/manifest/config/metadata.yaml tag=v0.2.0
```

在 GitHub Actions 中，如果未传入 `tag`，该命令也会使用 `GITHUB_REF_NAME` 作为待校验标签。

## 插件工作区命令

插件工作区管理始终使用固定目录 `apps/lina-plugins`。在 `hack/config.yaml` 中配置来源：

```yaml
plugins:
  sources:
    official:
      repo: "https://github.com/linaproai/official-plugins.git"
      root: "."
      ref: "main"
      items:
        - "linapro-tenant-core"
        - "linapro-org-core"
```

`items` 只接受插件 ID 字符串。使用带引号的 `"*"` 可安装 source `root` 下一层的全部插件目录；不要写裸的 `- *`，因为 YAML 会把它当作 alias 语法。如果同一仓库中的插件需要不同 `ref`，应拆成多个 source。

常用命令：

```bash
make plugins.init
make plugins.install
make plugins.install p=linapro-tenant-core
make plugins.update source=official
make plugins.update force=1
make plugins.status
```

`plugins.init` 会将 `apps/lina-plugins` 从 `submodule` 转成普通目录并保留文件。`plugins.install`、`plugins.update` 和 `plugins.status` 会在需要时自动执行同等工作区初始化，因此用户可以直接执行实际需要的命令。`plugins.install` 和 `plugins.update` 会复用 `temp/plugin-sources/<source>` 下的配置来源缓存，首次 clone 后通过 fetch 更新，再复制插件目录到 `apps/lina-plugins/<plugin-id>`，并更新工具生成的 `apps/lina-plugins/.linapro-plugins.lock.yaml` 锁文件。

## 验证

```bash
cd hack/tools/linactl
go test ./...
go run . help
go run . wasm dry-run=true
go run . plugins.status
go run . i18n.check
go run . release.tag.check tag=v0.2.0
```
