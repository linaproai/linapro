## Why

`LinaPro`仓库已经统一在`.agents/skills/`下维护`Agent Skills`内容，但不同`AI Coding`工具（如`Claude Code`、`CodeBuddy`、`Windsurf`、`Qoder`等）默认从各自独立的项目路径（如`.claude/skills/`、`.codebuddy/skills/`、`.windsurf/skills/`、`.qoder/skills/`）发现技能。当前仓库需要开发者手工执行`ln -s .agents/skills .<tool>/skills`才能让对应工具加载技能；命令易写错、跨平台行为不一致，并且没有任何可见入口可用于查询和重建链接状态。

`docs/quick/4000-agent-tools.md`已对接的工具集与[`vercel-labs/skills`](https://github.com/vercel-labs/skills#supported-agents)官方支持的项目路径列表是同一类需求。需要一个仓库内置的、跨平台的开发命令统一管理这些项目级软链，避免依赖`npx skills`等外部`CLI`，也避免硬编码`bash`/`PowerShell`脚本。

## What Changes

- 在`hack/tools/linactl`中新增`skills.link`和`skills.unlink`命令，按`vercel-labs/skills`项目路径表为支持的`Agent`在仓库根维护`.<tool>/skills -> .agents/skills`项目级软链。
- 命令仅在**当前仓库根目录内**操作项目路径，不涉及全局`HOME`目录，不引入外部`CLI`依赖，不交互式提问。
- 通过`AGENT=<name|all|csv>`参数选择目标`Agent`；`FORCE=1`允许覆盖指向错误源的旧软链；普通目录或文件冲突永不自动覆盖。
- 内置`Agent`映射表与`vercel-labs/skills`官方表对齐，区分三类：
  - **native**：项目路径已经是`.agents/skills/`的`Agent`（如`amp`、`cursor`、`gemini-cli`、`codex`等），跳过创建；
  - **link**：项目路径是其它路径的`Agent`（如`claude-code`→`.claude/skills`、`codebuddy`→`.codebuddy/skills`、`windsurf`→`.windsurf/skills`），创建软链到`.agents/skills`；
  - **root-collision**：项目路径是仓库根的`skills/`（仅`openclaw`），默认跳过，仅在显式`AGENT=openclaw FORCE=1`下创建。
- 新增`hack/makefiles/skills.mk`，提供`make skills.link`、`make skills.unlink`包装入口，并在根`Makefile`中`include`。
- 新增`hack/tools/linactl/internal/skilllink/`子组件承载映射表、创建/检测/移除逻辑和单元测试。
- 在`.gitignore`中按需新增软链目录的忽略规则，避免`Agent`目录被纳入版本控制。
- 更新`hack/tools/linactl`中英文`README`，说明`skills.link`/`skills.unlink`的语义与`AGENT`/`FORCE`参数。

不在本次范围：

- 不创建`AGENTS.md → CLAUDE.md / GEMINI.md / QWEN.md`这类文件级软链（仓库已通过`CLAUDE.md -> AGENTS.md`等单独维护）。
- 不管理`HOME`目录下的全局技能路径。
- 不下载或同步`vercel-labs/skills`内容；不替代`skills-lock.json`已有的`goframe-v2`从远端拉取流程。
- 不调整`.agents/skills/`下任何`Skill`内容、`SKILL.md`格式或加载顺序。

## Capabilities

### New Capabilities

- `agent-skills-link-cli`：定义仓库内`Agent`项目路径软链的统一管理边界，包括支持的`Agent`列表、软链创建/检测/移除规则、跨平台约束和冲突处理策略。

### Modified Capabilities

- 无。本次新增独立能力，不修改既有`OpenSpec`规范。

## Impact

- 影响`hack/tools/linactl`（新增命令文件与`internal/skilllink`子组件）、根`Makefile`、`hack/makefiles/skills.mk`、`hack/tools/linactl/README.md`、`hack/tools/linactl/README.zh-CN.md`和`.gitignore`。
- 不涉及后端运行时服务、`REST API`、数据库结构、前端页面、用户可见运行时文案、运行时缓存或数据权限逻辑。
- 不涉及`i18n`运行时资源：仅新增开发工具命令与工具文档，不引入用户可见运行时文案；工具文档遵循中英文`README`双语规范。
- 不涉及缓存一致性：命令是开发期一次性`os.Symlink`/`os.Remove`操作，不引入运行时缓存或跨实例协调。
- 不涉及数据权限：命令操作仓库内文件系统，不访问后端数据接口。
- 跨平台影响：完全使用`Go`标准库`os.Symlink`/`os.Readlink`/`os.Lstat`/`os.Remove`实现，不调用`ln`/`mklink`；`Windows`下若`os.Symlink`返回权限错误则给出明确提示（开发者模式或管理员），不退化为平台专属脚本。
