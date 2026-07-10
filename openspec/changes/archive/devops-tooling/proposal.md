## Why

LinaPro 的开发与运维工具链需要支撑跨平台协作、可持续交付、自动化治理和可审计发布。早期命令路径分散在 GNU Make、POSIX Shell、独立 Go 工具、GitHub Actions 内联脚本和手工流程中，导致 Windows 开发门槛高、工具版本漂移、构建职责分散、OpenSpec 归档依赖人工触发、发布门禁不统一、升级流程与源码插件治理边界不清晰。

作为面向可持续交付的 AI 原生全栈框架，LinaPro 还需要把 Agent 资源、OpenSpec 月度归档、性能审计、安装脚本、镜像发布、版本一致性和 GoFrame 代码生成纳入同一套开发工具治理中。工具链应降低本地环境差异，保持运行时系统边界稳定，并让自动化流程在失败时给出明确、可复现、可审查的证据。

## What Changes

- 以`hack/tools/linactl`作为跨平台开发命令主入口，`Makefile`和 Windows`make.cmd`仅作为薄包装入口，统一`dev`、`stop`、`status`、`build`、`wasm`、`init`、`mock`、`test`、`env.check`、`env.setup`等命令。
- 将镜像构建、动态插件 Wasm 打包、运行时 i18n 扫描和 GoFrame controller/DAO 生成整合到`linactl/internal/`，移除默认开发路径中的独立工具模块和外部`gf`二进制依赖。
- 扩展`linactl ctrl`和`linactl dao`，支持显式指定生成目标目录（插件 ID 或后端目录），默认继续指向`apps/lina-core`；根`Makefile`、宿主`Makefile`和插件根目录`Makefile`提供一致的代码生成入口，插件目录通过共享`hack/makefiles/plugin.codegen.mk`统一维护，不再硬编码插件 ID 或后端路径。
- 将插件代码生成配置从`backend/hack/config.yaml`迁移到插件根`hack/config.yaml`，解耦 GoFrame 工作目录和配置目录；`linactl ctrl`和`linactl dao`只保留`dir=`目标参数，删除旧`p=`、`plugin=`和`target=`参数。
- 将插件自定义构建指令从插件`Makefile`变量收敛到插件根`hack/config.yaml`的`build.commands`；删除`apps/lina-plugins`根`go.mod`、`go.sum`和`lina-plugins.go`，改为由`linactl`自动生成源码插件聚合模块。
- 移除 LinaPro 自定义`shutdown.timeout`配置入口，改用 GoFrame Server 的`server.gracefulShutdownTimeout`作为停机超时唯一来源。
- 提供`agents`多资源命令树，统一管理 Agent 的`skills`、`prompts`和`AGENTS.md`桥接软链，并保持跨平台、安全、不删除真实目录或文件。
- 建立月度 OpenSpec 自动归档和归档聚合 workflow，支持 Codex、Claude Code 和 GitHub Copilot CLI，使用共享 prompt、运行时凭据注入、阶段性 fail-fast、OpenSpec 校验和 PR 写回。
- 加强 release 与 nightly 镜像发布治理：release 复用共享测试验证套件，校验 tag 与`framework.version`一致，成功后创建 GitHub Release；manual nightly 可显式跳过测试门禁用于维护重发。
- 提供受控 release tag 创建入口、跨平台 tag 校验、GitHub App token 规则集绕过说明和 Docker tag 兼容版本格式约束。
- 记录 framework/source-plugin 开发期升级入口的治理演进，后续插件有效版本、发现版本、运行时升级和发布切换契约归插件框架分组统一承载。
- 提供跨平台安装脚本、内存态 demo Compose、开发容器 Compose、`lina-perf-audit`手动触发性能审计技能和持久 issue-card 机制。
- 将非工具链 owner 的项目初始化、数据库启动、E2E 组织和定时任务清理内容迁移为交叉影响摘要，避免在本分组重复保存完整能力规范。
- 建立仓库级`Go`静态检查门禁：固定`golangci-lint`与`staticcheck`版本，通过`linactl lint.go`提供跨平台入口，支持宿主模式与插件完整模式，并接入主`CI`与发布验证。
- 将动态插件 builder 专用的 hook、resource 与 lifecycle timeout 配置统一收敛到插件根`hack/config.yaml`的`wasm.*`配置，不再从`backend/*/*.yaml`读取。
- 收敛`linactl`公开契约：删除旧构建变量、`package.json`构建回退、`plugins=auto`、snake_case 参数映射、宽松布尔别名、环境变量标签/registry 覆盖与镜像调试入口；`wasm`单插件入口仅`dir=`。

## Capabilities

### New Capabilities

- `cross-platform-dev-commands`
- `linactl-build-tool-consolidation`
- `agents-multi-resource`
- `framework-bootstrap-installer`
- `lina-perf-audit-skill`
- `monthly-openspec-archive`
- `release-version-governance`
- `go-static-lint-governance`：仓库级`Go`静态检查配置、`linactl lint.go`入口、多目标死代码检查、插件工作区覆盖、`CI`门禁与治理文档。
- `plugin-runtime-builder-config`：动态插件 builder 配置统一到插件根`hack/config.yaml`的`wasm`配置。

### Modified Capabilities

- `upgrade-governance`
- `release-image-build`
- `runtime-upgrade-governance`
- `agent-skills-link-cli`

## Impact

- 影响仓库开发命令、工具模块组织、GitHub Actions、安装脚本、Agent 资源桥接、OpenSpec 月度治理、release/nightly 发布、GoFrame 代码生成入口和开发期升级流程的历史追溯。
- 不改变当前运行时代码、HTTP API、数据库 schema、业务权限、数据权限、插件运行时契约、前端 UI 或运行时缓存行为；`plugin-upgrade-governance`等运行时插件契约由`archive/plugin-framework`和主规范承载。
- 插件目录影响仅为新增根目录薄`Makefile`并引入共享`hack/makefiles/plugin.codegen.mk`，不修改插件业务源码、插件清单、生命周期资源或运行时授权边界。
- `i18n`历史影响主要来自工具文档、workflow 注释、OpenSpec 任务记录和少量运行时实现阶段的独立 owner 变更；本次压缩不修改语言包、`manifest/i18n`或`apidoc i18n JSON`。
- 验证以工具单元测试、命令 smoke、GitHub Actions YAML/shell 检查、OpenSpec 校验、release/nightly workflow 验证、性能审计 dry-run 和文档镜像同步为主；本次压缩只做 OpenSpec 文档治理验证。
