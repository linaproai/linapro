## Why

社区协作需要统一的自动化技能来处理`GitHub Issue`分类、`Pull Request`审查和版本更新日志生成。当前缺少仓库级`Issue`审查技能来区分疑问、功能需求、`Bug`和无效内容；缺少统一的`PR`审查流程来按项目规范检查代码变更并留下可追踪评论；缺少稳定的手动流程来从`Git`历史和`OpenSpec`内容中整理详尽的双语版本更新日志。

## What Changes

- 新增仓库级`lina-community-issue-review`技能，审查`linaproai/linapro`仓库的`GitHub Issue`，支持按编号或全量遍历，跳过已评论且已打标签的`Issue`，根据项目规范和源码实现分类处理疑问、功能需求、`Bug`和无效内容。
- 新增仓库级`lina-community-pr-review`技能，审查`linaproai/linapro`仓库的开放`Pull Request`，支持按编号或全量遍历，跳过已带`bot-approved`标签或无新提交的`PR`，根据可信项目规范审查代码差异并留下问题评论或通过标签。
- 新增仓库级`lina-community-release-changelog`技能，手动生成版本更新日志，支持默认范围和显式双引用范围，基于`Git`历史、源码差异和`OpenSpec`内容整理双语`Markdown`发布日志并写入`temp/changelog.md`。

## Capabilities

### New Capabilities

- `community-issue-review-skill`：定义`lina-community-issue-review`技能的触发、仓库范围、`Issue`遍历、重复审查跳过、可信上下文加载、已处理核对、分类处理、评论语言、标签和关闭契约。
- `community-pr-review-skill`：定义`lina-community-pr-review`技能的触发、仓库范围、`PR`遍历、跳过规则、规范加载、评论语言、人工升级和通过标签契约。
- `release-changelog-skill`：定义`lina-community-release-changelog`技能的手动触发、比较范围解析、证据收集、双语模板和输出文件契约。

### Modified Capabilities

- 无。

## Impact

- 新增`.agents/skills/lina-community-issue-review/SKILL.md`、`.agents/skills/lina-community-pr-review/SKILL.md`和`.agents/skills/lina-community-release-changelog/SKILL.md`。
- 新增`openspec/changes/archive/community-skills/`下的归档文档。
- 本变更不修改后端、前端、数据库、`HTTP API`、插件运行时、`CI`或`GitHub Actions`流程。
- `i18n`影响：技能生成的`GitHub`评论语言跟随`Issue`或`PR`描述语言，发布日志生成双语内容，但不新增运行时用户可见文案、语言包、接口文档本地化资源或翻译缓存。
- 缓存一致性、数据权限、模块启停、核心宿主接口契约均无影响。
- 开发工具跨平台影响：技能依赖代理环境中的`gh`、`git`和`jq`等命令，不新增长期维护脚本、`Makefile`目标或`linactl`命令。
