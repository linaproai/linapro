## Why

当前 OpenSpec 归档和归档聚合依赖人工触发，已完成变更容易长期停留在活跃目录中，导致活跃变更列表噪声增加，也会让反馈流程误判已有完成变更仍处于待处理状态。将归档治理放到 monthly 批处理可以降低每个 PR 合并后的文档抖动，同时避免每日自动归档带来的资源浪费和构建压力；维护者仍可在必要时通过手动触发执行归档。

## What Changes

- 新增 GitHub Actions `monthly` 归档工作流，每月 1 日 00:00 Asia/Shanghai 自动扫描已完成的活跃 OpenSpec 变更。
- 保留 `workflow_dispatch` 手动触发入口，允许维护者在月度周期外按需执行 OpenSpec 归档。
- 主工作流通过 GitHub Variables 中的 `AI_CODING_TOOL` 选择 AI Coding 工具，支持 `codex` 和 Claude Code 的 `cc` 取值；未配置时默认使用 `codex`。
- 工具专属 reusable workflow 分别封装 Codex 和 Claude Code 的镜像、认证配置、执行命令与日志上传细节，主工作流只负责检测和路由。
- 工具无关的 runner 准备、归档差异检测、OpenSpec 校验、范围保护和 PR 写回治理逻辑抽取到 `.github/actions/` 本地 composite action 复用。
- 工作流先执行 `lina-auto-archive`，仅当本次产生归档变更后再执行 `lina-archive-consolidate`。
- 自动归档和归档聚合提示词统一维护在 `.github/prompts/` 下，工具专属 workflow 通过文件引用，避免重复维护。
- 工作流执行 OpenSpec 校验和变更范围保护，确保自动归档 PR 只包含 OpenSpec 归档治理相关文件。
- 调整 `.github/codex` 与 `.github/cc` 模板治理方式，允许提交无密钥配置模板，同时继续忽略或避免写入真实认证文件。

## Capabilities

### New Capabilities
- `monthly-openspec-archive`: 定义 monthly OpenSpec 自动归档、聚合、校验、PR 写回、AI Coding 工具选择与密钥注入的行为要求。

### Modified Capabilities
- `spec-governance`: 补充 OpenSpec 归档治理可以由受控 monthly 自动化流程执行的要求。

## Impact

- 影响 `.github/workflows/` 中的 GitHub Actions 工作流配置。
- 影响 `.github/actions/` 中的本地 composite action。
- 影响 `.github/codex/` 下 Codex 运行配置模板的版本控制方式。
- 影响 `.github/cc/` 下 Claude Code 运行配置模板的使用方式。
- 影响 `.gitignore` 对 AI Coding 工具配置目录的忽略规则。
- 不涉及后端运行时代码、前端 UI、数据库 SQL、运行时 i18n 资源或缓存一致性逻辑。
