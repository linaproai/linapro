## Why

当前 OpenSpec 归档和归档聚合依赖人工触发，已完成变更容易长期停留在活跃目录中，导致活跃变更列表噪声增加，也会让反馈流程误判已有完成变更仍处于待处理状态。将归档治理放到 nightly 批处理可以降低每个 PR 合并后的文档抖动，同时保持归档资产按日自动收敛。

## What Changes

- 新增 GitHub Actions `nightly` 归档工作流，每天凌晨低峰期自动扫描已完成的活跃 OpenSpec 变更。
- 工作流通过 `loads/codex:latest` 运行 Codex，并在运行时从 GitHub Secret 注入 `OPENAI_API_KEY`。
- 工作流先执行 `lina-auto-archive`，仅当本次产生归档变更后再执行 `lina-archive-consolidate`。
- 工作流执行 OpenSpec 校验和变更范围保护，确保自动提交只包含 OpenSpec 归档治理相关文件。
- 调整 `.github/codex` 模板治理方式，允许提交无密钥配置模板，同时继续忽略真实认证文件。

## Capabilities

### New Capabilities
- `nightly-openspec-archive`: 定义 nightly OpenSpec 自动归档、聚合、校验、提交与密钥注入的行为要求。

### Modified Capabilities
- `spec-governance`: 补充 OpenSpec 归档治理可以由受控 nightly 自动化流程执行的要求。

## Impact

- 影响 `.github/workflows/` 中的 GitHub Actions 工作流配置。
- 影响 `.github/codex/` 下 Codex 运行配置模板的版本控制方式。
- 影响 `.gitignore` 对 Codex 配置目录的忽略规则。
- 不涉及后端运行时代码、前端 UI、数据库 SQL、运行时 i18n 资源或缓存一致性逻辑。
