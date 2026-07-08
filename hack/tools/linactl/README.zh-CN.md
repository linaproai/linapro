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
go run . db.init confirm=init
go run . db.upgrade confirm=upgrade
go run . db.mock confirm=mock
go run . tidy
go run . lint.go plugins=0
go run . lint.go plugins=1
go run . lint.go plugins=0 fix=true
go run . build platforms=linux/amd64,linux/arm64
go run . build dir=apps/lina-plugins/john-ai-agentbox
go run . dev dir=tools/custom-builder
go run . stop dir=tools/custom-builder
go run . status dir=tools/custom-builder
go run . image tag=v0.2.0 push=0
go run . version to=v0.2.0
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
make.cmd db.init confirm=init
make.cmd db.upgrade confirm=upgrade
make.cmd db.mock confirm=mock
make.cmd tidy
make.cmd lint.go plugins=0
make.cmd lint.go plugins=1
make.cmd version to=v0.2.0
make.cmd release.tag.check tag=v0.2.0
```

在`PowerShell`中，需要显式添加当前目录前缀：

```powershell
.\make.cmd help
.\make.cmd status
.\make.cmd pack.assets
.\make.cmd i18n.check
.\make.cmd lint.go plugins=0
.\make.cmd version to=v0.2.0
.\make.cmd release.tag.check tag=v0.2.0
```

## 参数

`linactl`支持现有`make`风格的`key=value`参数，降低命令迁移成本。

| 参数 | 示例 | 用途 |
|------|------|------|
| `confirm` | `confirm=upgrade` | 确认高风险数据库维护命令。 |
| `rebuild` | `rebuild=true` | 在`db.init`时重建配置中的数据库。 |
| `dir` | `dir=tools/custom-builder` | 为`build`、`dev`、`stop`或`status`选择单个定向命令目录。省略时执行对应命令的默认完整流程。 |
| `platforms` | `platforms=linux/amd64,linux/arm64` | 指定构建目标平台。 |
| `plugins` | `plugins=0` | 覆盖构建、开发、镜像、`Go`测试和`Go`静态检查命令的自动插件完整模式探测。 |
| `fix` | `fix=true` | 允许`lint.go`向`golangci-lint`传入`--fix`；默认不启用，避免检查路径改写文件。 |
| `to` | `to=v0.2.0` | 指定`version`写入的框架版本号。 |
| `tag` | `tag=v0.2.0` | 指定`release.tag.check`校验的 release tag。 |
| `print-version` | `print-version=1` | 输出已校验的`framework.version`，供发布自动化使用。 |
| `p` | `p=linapro-tenant-core` | 为 Wasm 构建或插件工作区管理命令选择单个插件。 |
| `plugin-dir` | `plugin_dir=/path/to/plugin` | 从显式源码目录构建单个动态插件产物。 |
| `out` | `out=temp/output` | 指定动态插件产物输出目录；相对路径按仓库根目录解析。 |
| `source` | `source=official` | 为插件工作区管理命令选择单个已配置来源。 |
| `force` | `force=1` | 允许插件安装或更新命令覆盖已存在或存在本地改动的插件目录。 |
| `verbose` | `verbose=1` | 构建任务展示子命令输出。 |

未传入`plugins`时，构建和开发命令会在`apps/lina-plugins`存在插件清单时启用插件完整模式。插件完整模式会基于宿主专用的根目录`go.work`生成或刷新已忽略的`temp/go.work.plugins`，并通过`GOWORK`解析源码插件`Go`模块。

未传入`dir`时，`linactl build`会构建宿主框架后端、默认管理工作台前端、宿主`manifest`资源和所有已启用官方插件。需要从仓库根目录或通过`make.cmd`跨平台定向构建时，使用`dir=<path>`，例如`dir=apps/lina-vben`、`dir=apps/lina-core`、`dir=apps/lina-plugins/<plugin-id>`，或任意拥有`hack/config.yaml`的目录。

目标目录可以在自身`hack/config.yaml`的对应命令分段下维护自定义指令。`linactl build dir=<path>`和`linactl dev dir=<path>`执行`build.commands`；`linactl stop dir=<path>`执行`stop.commands`；`linactl status dir=<path>`执行`status.commands`。指令会在所选目录执行。`$(TARGET_DIR)`和`$(BUILD_DIR)`都会展开为所选目录，`$(PLUGIN_ROOT)`作为插件兼容别名继续指向同一目录，`$(REPO_ROOT)`展开为仓库根目录：

```yaml
build:
  commands:
    - pnpm --dir "$(BUILD_DIR)/frontend" run build
stop:
  commands:
    - node scripts/stop.mjs --root "$(TARGET_DIR)"
status:
  commands:
    - node scripts/status.mjs --root "$(TARGET_DIR)"
```

传入`dir=apps/lina-plugins/<plugin-id>`时，官方插件模式仍作用于该插件。源码插件会使用官方插件构建环境，动态插件会在配置指令完成后继续生成自身`WASM`产物。非插件目录存在`hack/config.yaml`时以该配置为准；没有该配置的目录继续回退到本地`package.json`的`build`脚本。

传入`dir`给`linactl dev`时，命令会执行与`linactl build dir=<path>`一致的定向构建路径，不启动或重启开发服务。传入`dir`给`linactl stop`或`linactl status`时，目录配置指令会替代默认宿主服务停止或状态查询流程。

## Go 静态检查

`linactl lint.go`通过`golangci-lint`运行仓库`Go`静态检查门禁。主检查工具版本由仓库根目录`.golangci-lint-version`锁定，规则配置位于仓库根目录`.golangci.yml`，死代码检查使用仓库根目录`.staticcheck-version`锁定的`staticcheck`版本。

如果`PATH`中缺少`golangci-lint`或`staticcheck`，或其版本与锁定版本不一致，`linactl lint.go`会通过`go install`安装锁定版本，并使用安装后的精确二进制继续执行。`linactl env.setup`会在前端和浏览器环境安装前执行同一套锁定工具安装流程，使新开发环境能够提前准备`Go`静态检查工具。安装流程使用`GOWORK=off`并移除构建标签或交叉编译变量，避免插件完整模式的 lint 设置影响外部工具构建。首次安装需要正常的`Go`模块网络访问。

```bash
make lint.go plugins=0
make lint.go plugins=1
make lint.go plugins=0 fix=true
go run . lint.go plugins=0
```

使用`plugins=0`检查宿主工作区，覆盖`apps/lina-core`和`hack/tools/linactl`。官方插件源码已初始化时，使用`plugins=1`；该模式会准备已忽略的`temp/go.work.plugins`工作区，并检查宿主、工具和官方插件`Go`模块。未传入`plugins`时，`linactl`沿用构建和测试命令的自动探测行为。

`golangci-lint`不启用独立`unused` linter。`linactl lint.go`会对所有包运行`staticcheck U1000`作为死代码检查；非测试文件包含`//go:build wasip1`或`//go:build !wasip1`的包使用宿主目标和`GOOS=wasip1 GOARCH=wasm`矩阵，避免 guest 专属桥接代码在默认宿主构建下被误报为死代码。

`fix=true`是显式开发者操作。它允许`golangci-lint`在支持时改写导入和格式；`CI`不会启用该参数。

## 构建工具命令

`linactl`统一承载仓库镜像构建和动态插件`Wasm`打包实现。公开入口仍然是根目录`make`目标和对应的`linactl`命令：

```bash
make image tag=v0.2.0 push=0
make image.build tag=v0.2.0
make wasm p=linapro-demo-dynamic
```

当测试或本地夹具需要打包`apps/lina-plugins`之外的动态插件目录时，可以使用`plugin_dir=<path>`。

## GoFrame 代码生成

`linactl ctrl`和`linactl dao`会直接运行内嵌在`linactl`中的`GoFrame CLI`模块；开发者不再需要单独安装`gf`，也不需要在`PATH`中提供`gf`可执行文件。

```bash
go run . ctrl
go run . dao
go run . ctrl dir=apps/lina-plugins/linapro-content-notice/backend
go run . dao dir=apps/lina-plugins/linapro-content-notice/backend
```

未传入目标参数时，生成流程使用`apps/lina-core`的`GoFrame`项目布局，并读取`apps/lina-core/hack/config.yaml`。使用`dir=<backend-dir>`可以定向其他后端。标准插件后端目标会保持`GoFrame`工作目录为`apps/lina-plugins/<plugin-id>/backend`，并从插件根`apps/lina-plugins/<plugin-id>/hack/config.yaml`读取代码生成配置；非插件目标继续读取`<backend-dir>/hack/config.yaml`。`dao`生成仍要求配置的数据库可连接且已初始化，因此执行前需要先运行仓库初始化流程或准备等价数据库。

## 运行时 I18n 检查

`linactl i18n.check`统一承载运行时`i18n`治理检查。该命令会扫描高风险运行时可见硬编码文案，并校验宿主和插件运行时消息`key`覆盖：

```bash
make i18n.check
go run . i18n.check
```

默认扫描`allowlist`维护在`hack/tools/linactl/internal/runtimei18n/allowlist.json`。

## 插件治理检查

`linactl plugins.check`会扫描`apps/lina-plugins`下所有包含`plugin.yaml`的插件目录。它检查插件生产路径中的宿主核心表生成、`sys_*`表直接访问、旧`pluginbridge host-service`使用，以及动态`data host-service`中不属于当前插件的表授权。

```bash
make plugins.check
go run . plugins.check
go run . plugins.check format=json
```

## Agent 软链管理（agents.* 命令树）

`linactl agents.<resource>.<action>` 用于管理仓库内三类资源的本地软链，把 `.agents/`（以及 `AGENTS.md`）下的标准源映射到各 AI Coding 工具的私有项目路径：

- **skills**：目录级软链，`.<tool>/skills` → `.agents/skills`。受支持的 Agent 列表与 [vercel-labs/skills](https://github.com/vercel-labs/skills#supported-agents) 官方项目路径表一致。
- **prompts**：目录级软链，各 Agent 的 commands/prompts 根目录（例如 `.claude/commands`）→ `.agents/prompts`。
- **md**：单文件软链，`.<tool>.md`（或其他私有规范文件）→ 仓库根 `AGENTS.md`。

命令只在仓库根目录范围内操作，不会修改`HOME`或任何系统全局路径，也不会自动删除真实目录或非 Git 降级伪文件（`force=1`同样不会）。

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

`agent`必须是单个受支持 Agent 名称：聚合命令显式拒绝`agent=all`与逗号列表（批量场景请走下方子命令）。Agent 名称会归一化为标准`kebab-case`，所以`ClaudeCode`、`Claude Code`、`claude_code`和`claude-code`都会解析为`claude-code`。`action`默认为`link`。未传`agent`时，非终端环境会打印用法指引而不会阻塞等待输入。参数键区分大小写，并统一使用小写的`linactl` `key=value`名称。

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

任何情况下命令都不会自动删除已存在的真实目录或文件，**除非**该文件是 Git 在`core.symlinks=false`时创建的降级软链伪文件。Git 在禁用符号链接时会将预期的链接目标路径作为纯文本写入文件内容；`force=1`能自动检测这类伪文件（普通文件、≤512 字节、内容与预期的相对源路径匹配）并将其替换为真正的符号链接。具有任意内容的正常文件仍然不会被触碰。`force=1`同时也会重建"已是软链但指向其它位置"的情况。所有 skills 与 prompts 受管软链目录已在`.gitignore`中忽略，本地创建不会污染仓库。

### 从 `make skills.*` 迁移

旧的 `make skills` / `make skills.link` / `make skills.unlink` 目标，以及对应的 `linactl skills*` 子命令均已**删除**，被 `agents.*` 命令树取代。**没有保留任何别名**；现有脚本与文档必须更新：

| 已删除（不再生效） | 新命令 |
| --- | --- |
| `make skills` | `make agents` |
| `make skills.link` | `make agents.skills.link` |
| `make skills.link agent=<name>` | `make agents.skills.link agent=<name>` |
| `make skills.link agent=all force=1` | `make agents.skills.link agent=all force=1` |
| `make skills.unlink` | `make agents.skills.unlink` |
| `make skills.unlink agent=<name>` | `make agents.skills.unlink agent=<name>` |
| `linactl skills` | `linactl agents` |
| `linactl skills.link` | `linactl agents.skills.link` |
| `linactl skills.unlink` | `linactl agents.skills.unlink` |

`agents.skills.*` 子命令的行为与原 `skills.*` 完全一致（同一注册表、同一状态机、同一 TTY/CI 行为），仅命令名称变化。

## 版本元数据

`version`会更新`apps/lina-core/manifest/config/metadata.yaml`中的`framework.version`，并为根目录`README`图片地址刷新`v=<version>`缓存键。

```bash
make.cmd version to=v0.2.0
make version to=v0.2.0
```

## Release Tag 校验

`release.tag.check`会读取`apps/lina-core/manifest/config/metadata.yaml`，并校验 release tag 与`framework.version`完全一致。

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
go run . lint.go plugins=0
go run . release.tag.check tag=v0.2.0
```
