## Why

发布人员当前缺少一个稳定、可复用的手动流程来从`Git`历史、源码变更和`OpenSpec`内容中整理详尽的版本更新日志。新增`lina-community-release-changelog`技能可以在接入`CI`前先通过人工执行验证生成质量，降低发布说明遗漏关键变更或格式不一致的风险。

## What Changes

- 新增仓库级`lina-community-release-changelog`技能，用于手动生成`LinaPro`版本更新日志。
- 技能根据指定或默认的`Git`比较范围分析提交历史、源码差异和`OpenSpec`文档，生成详尽的双语`Markdown`更新日志。
- 生成结果固定写入项目根目录`temp/changelog.md`，不创建`GitHub Release`，不修改`CI`或发布工作流。
- 技能支持用户显式指定两个比较版本、标签或提交，例如比较`v0.1.0`与`v0.2.0`之间的历史变更；未指定时默认使用最近可达发布标签到当前`HEAD`。
- 输出使用固定模板：英文内容在上、中文内容在下，中间使用模板要求的分割线，章节仅包含`Highlights`、`Improvements`、`Bug Fixes`、`Tooling and Experience`及对应中文章节。
- 技能不要求在`PR`中新增关键标识、标签或发布说明字段，必须仅依赖仓库内可见证据整理内容。

## Capabilities

### New Capabilities

- `release-changelog-skill`: 定义`lina-community-release-changelog`技能的手动触发、比较范围解析、证据收集、双语模板和输出文件契约。

### Modified Capabilities

- 无。

## Impact

- 新增`.agents/skills/lina-community-release-changelog/SKILL.md`。
- 可能新增技能评估或示例材料，用于验证默认范围、显式版本范围和内容完整性。
- 生成产物位于已忽略的`temp/changelog.md`，不会作为版本库交付文件提交。
- 本变更不修改后端、前端、数据库、`HTTP API`、插件运行时、`CI`或`GitHub Actions`发布流程。
- `i18n`影响：技能本身会生成英文和中文双语发布日志，但不新增运行时用户可见文案、语言包或接口文档本地化资源。
- 缓存一致性、数据权限、模块启停、核心宿主接口契约均无影响。
- 开发工具跨平台影响：技能依赖代理可执行`git`与文件读取命令，当前阶段不新增长期维护脚本或`linactl`命令；后续如引入确定性辅助工具，应优先采用跨平台`Go`工具。
