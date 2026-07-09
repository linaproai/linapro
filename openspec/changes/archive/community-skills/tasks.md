## Summary

- [x] 新建`.agents/skills/lina-community-issue-review/SKILL.md`，覆盖默认仓库、指定`Issue`、开放`Issue`遍历、重复审查跳过、已处理核对、评论语言、`question`/`feature`/`bug`标签、`resolved`关闭和不可信输入边界。
- [x] 新建`.agents/skills/lina-community-pr-review/SKILL.md`，覆盖默认仓库、指定`PR`、开放`PR`遍历、`bot-approved`跳过、隐藏标记幂等、可信规范读取、评论语言、问题评论、阻断评论、通过标签和历史维护成员`@`升级。
- [x] 新建`.agents/skills/lina-community-release-changelog/SKILL.md`，覆盖手动触发、默认范围、显式双引用范围、证据收集、双语模板和`temp/changelog.md`输出。
- [x] **FB-1**（`issue-review`）：新增已处理核对流程，功能或`Bug`已在当前项目中处理时评论说明并关闭`Issue`。
- [x] **FB-2**（`pr-review`）：将`SKILL.md`改为中文描述，保留英文评论模板。
- [x] **FB-3**（`release-changelog`）：修复`plugin-full / plugins-2-of-5`中源码插件卸载用例受动态插件反向依赖状态污染导致的`GitHub Actions`失败，在`beforeEach`中同步重置`linapro-demo-dynamic`注册行。
- [x] 运行`openspec validate`三个变更均通过`--strict`验证。
- [x] 静态检索确认技能覆盖默认仓库、指定编号、全量遍历、跳过规则、语言跟随、标签/关闭/通过契约和不可信输入边界。
- [x] `git diff --name-only -- .github apps manifest hack Makefile make.cmd`无输出，确认未修改运行时代码、数据库、前端、插件或`CI`。
- [x] 影响分析：`i18n`仅影响技能生成评论和发布日志语言，缓存一致性无影响，数据权限无影响，模块启停无影响，核心宿主接口契约无影响，开发工具跨平台无新增长期脚本。
- [x] `lina-review`审查通过，未发现阻塞问题。
