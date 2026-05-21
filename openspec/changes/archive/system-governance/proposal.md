## Why

OpenSpec 归档治理长期依赖人工触发，已完成变更容易滞留在活跃目录中，增加反馈流程噪声和活跃列表维护负担。每月批处理可以降低每次 PR 合并后的文档抖动，同时避免每日自动归档带来的资源浪费。基础归档步骤最初依赖 AI Coding 工具执行，但实践表明 AI 进程可能返回成功却未完成实际归档，因此需要将基础归档改为确定性自动化，让 AI 工具仅用于可选的归档聚合增强。

与此同时，宿主服务、源码插件和 `lina-core/pkg` 公共组件的后端主文件职责尚未完全统一，部分组件将接口契约、类型定义和大量实现逻辑混在同一个主文件中，开发者难以通过主文件快速理解组件边界、能力和使用约束。接口方法注释和文件顶部说明的详细程度不足，不能稳定覆盖输入、输出、错误、权限、缓存、i18n 等使用约束。将主文件职责、接口方法注释、文件顶部说明和 `lina-review` 审查要求固化为项目规范，并按业务模块分批整改，能够系统性提升后端源码的可读性、可维护性和审查可控性。

## What Changes

### OpenSpec 归档自动化

- 新增 GitHub Actions monthly 归档工作流，每月 1 日 00:00 Asia/Shanghai 自动扫描已完成的活跃 OpenSpec 变更。
- 保留 `workflow_dispatch` 手动触发入口，允许维护者在月度周期外按需执行归档；手动触发可从任意暴露该 workflow 的分支 ref 进入，以触发分支作为检测和 PR 目标分支。
- 基础归档步骤改为确定性 GitHub Action，直接运行 `openspec list --json` 和 `openspec archive -y <change>`，不再依赖 AI 工具执行基础归档。
- 主工作流通过 GitHub Variables 中的 `AI_CODING_TOOL` 选择 AI Coding 工具（`codex`、`cc`、`copilot`），仅在确定性归档产生变更后用于归档聚合；未配置时默认使用 `codex`。
- 工具专属 reusable workflow 分别封装 Codex、Claude Code 和 GitHub Copilot CLI 的镜像、认证配置、执行命令与日志上传细节，主工作流只负责检测和路由。
- 工具无关的 runner 准备、归档差异检测、OpenSpec 校验、确定性归档断言、范围保护和 PR 写回治理逻辑抽取到 `.github/actions/` 本地 composite action 复用。
- 自动归档后立即执行 `openspec list --json` 断言无完成状态活跃变更残留；自动归档和归档聚合后分别执行 `openspec validate --all`，任一阶段失败时立即停止后续阶段。
- AI 归档聚合定义为可选增强阶段：失败或产生无效 OpenSpec 时回滚聚合结果，恢复已通过校验的确定性归档状态，不阻塞归档 PR 写回。
- 归档聚合提示词统一维护在 `.github/prompts/` 下，工具专属 workflow 通过文件引用，避免重复维护；基础归档不再使用 AI prompt。
- 运行时从 GitHub Secret/Variable 注入 AI 工具所需的 API key/token、provider `base_url`、模型和推理等级配置，不提交真实认证文件或真实服务 endpoint。
- workflow 采用 GitHub Actions bot 创建或更新维护 PR 的方式写回归档结果，不直接推送到默认分支；归档分支 push 成功但仓库策略阻止 PR 创建时，输出手动 PR 链接并成功结束。

### 后端源码可读性治理

- 强化后端主文件职责规范：宿主与源码插件 `internal/service/<component>/<component>.go`、`lina-core/pkg/<component>/<component>.go` 主文件只保留组件说明、核心类型、接口契约、实现结构体、构造函数和编译期接口断言，具体实现逻辑迁移到同包其他文件。
- 强化接口方法注释规范：后端接口定义中的每个方法注释必须说明功能作用、关键输入参数、输出结果、错误返回、权限/数据权限、缓存、i18n 或调用注意事项中适用的内容。
- 强化文件顶部注释规范：所有后端源文件顶部必须提供该文件职责、主要实现逻辑和注意事项的说明；主文件说明组件整体边界，非主文件说明当前文件承载的实现切片。
- 将上述要求写入 `AGENTS.md` 和 `lina-review` 审查清单，使后续后端 Go 变更必须接受主文件职责、接口注释完整度和文件顶部说明质量检查。
- 按业务模块分批整改，每个任务只覆盖一组职责明确的模块或公共组件，并在任务完成后运行对应 Go 编译门禁和治理验证。
- 固化 `hack/tools/linactl` 命令文件命名规范和子组件组织规范，要求命令实现按 `command_<command>.go` 命名，复杂共享实现迁移到 `internal/<组件名称>/` 子组件。

## Capabilities

### New Capabilities
- `monthly-openspec-archive`: 定义 monthly OpenSpec 确定性自动归档、可选 AI 归档聚合、校验、PR 写回、AI Coding 工具选择与密钥注入的行为要求。

### Modified Capabilities
- `spec-governance`: 补充 OpenSpec 归档治理可由受控 monthly 自动化流程执行的要求，包括确定性归档和归档残留治理。
- `backend-conformance`: 强化后端主文件职责、接口方法注释、文件顶部说明、`lina-review` 审查要求和 `linactl` 工具组织规范，覆盖宿主服务、源码插件后端服务和 `lina-core/pkg` 公共组件。

## Impact

- 影响 `.github/workflows/` 中的 GitHub Actions 工作流配置。
- 影响 `.github/actions/` 中的本地 composite action。
- 影响 `.github/prompts/` 中的 CI 提示词文件。
- 影响 `.github/codex/` 和 `.github/cc/` 下 AI Coding 工具运行配置模板。
- 影响 `.gitignore` 对 AI Coding 工具配置目录的忽略规则。
- 影响 `AGENTS.md` 和 `.agents/skills/lina-review/SKILL.md` 中的后端代码规范与审查标准。
- 影响 `openspec/specs/backend-conformance/spec.md` 和 `openspec/specs/spec-governance/spec.md` 的增量规范。
- 影响 `apps/lina-core/internal/service/**`、`apps/lina-core/pkg/**` 和 `apps/lina-plugins/*/backend/internal/service/**` 的源码组织与注释。
- 影响 `hack/tools/linactl/` 的命令文件组织和子组件结构。
- 不涉及后端运行时业务代码行为变更、前端 UI、数据库 SQL、运行时 i18n 资源或缓存一致性逻辑。
