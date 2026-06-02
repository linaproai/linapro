## 1. 技能实现

- [x] 1.1 新建`.agents/skills/lina-community-release-changelog/`目录和`SKILL.md`，声明手动触发、输出路径、依赖工具和触发场景。
- [x] 1.2 在`SKILL.md`中定义比较范围解析规则，覆盖默认范围、显式双引用范围、反向输入规范化、空范围和无法安全排序的失败处理。
- [x] 1.3 在`SKILL.md`中定义证据收集流程，要求读取`Git`历史、文件差异统计、关键源码变更和相关`OpenSpec`文档。
- [x] 1.4 在`SKILL.md`中固定`temp/changelog.md`的双语`Markdown`模板，确保章节与规范一致且不新增额外章节。
- [x] 1.5 在`SKILL.md`中明确详尽性、证据约束和双语一致性要求，避免只输出提交摘要或生成无证据内容。

## 2. 质量验证

- [x] 2.1 使用默认范围手动试运行技能，确认能够生成`temp/changelog.md`且不修改`CI`、标签、提交或发布状态。
- [x] 2.2 使用显式历史范围试运行技能，例如`v0.1.0..v0.2.0`，确认来源范围、标题版本和章节内容正确。
- [x] 2.3 检查生成的`temp/changelog.md`是否使用固定模板，英文和中文内容分区清晰且语义一致。
- [x] 2.4 检查生成内容是否覆盖关键源码、工具链、`OpenSpec`和修复类变更；如发现遗漏，更新技能指令后重试。

## 3. 治理与门禁

- [x] 3.1 运行`openspec validate add-release-changelog-skill --strict`并记录结果。
- [x] 3.2 运行文件存在性或静态检索检查，确认未修改`.github/workflows/`且技能输出路径仍为`temp/changelog.md`。
- [x] 3.3 在任务完成记录中说明影响分析：`i18n`仅影响生成文档内容、缓存一致性无影响、数据权限无影响、模块启停无影响、核心宿主接口契约无影响、开发工具跨平台无新增长期脚本。
- [x] 3.4 完成实现后调用`lina-review`进行规范和实现审查。

## 4. 执行记录

- 默认范围试运行：已按`v0.2.0..HEAD`生成`temp/changelog.md`，标题版本为`v0.3.0`。证据来源包括`git log --first-parent`、完整`git log`、`git diff --stat`、`git diff --name-status`、关键提交`git show --stat`、`apps/lina-core/manifest/config/metadata.yaml`和归档`OpenSpec`内容。
- 显式范围试运行：已临时生成并检查`v0.1.0..v0.2.0`样例，确认来源范围和标题版本为`v0.2.0`；随后已恢复默认范围输出。
- 模板检查：已验证`temp/changelog.md`仅包含固定模板章节，英文在上、中文在下，中间保留模板分割线，来源范围为`v0.2.0..HEAD`。
- 内容覆盖检查：默认范围输出覆盖插件框架、动态插件运行时、分布式协调、国际化治理、发布元数据、`linactl`工具链、月度`OpenSpec`自动化、`E2E`治理、`Go`单测效率和修复类变更。
- 验证命令：`openspec validate add-release-changelog-skill --strict`通过；文件存在性检查通过；静态检索确认技能输出路径仍为`temp/changelog.md`；`git status --short .github/workflows`、`git diff --name-only -- .github/workflows`和`git ls-files --others --exclude-standard -- .github/workflows`均无输出。
- 技能结构检查：使用`Ruby YAML`等价检查验证`SKILL.md`frontmatter只包含允许字段且`name`和`description`有效；官方`quick_validate.py`因本机缺少`PyYAML`未能执行。
- 影响分析：`i18n`仅影响生成的双语发布文档内容，不修改运行时语言包、接口文档翻译或翻译缓存；缓存一致性无影响；数据权限无影响；模块启停无影响；核心宿主接口契约无影响；开发工具跨平台无新增长期脚本或`linactl`命令；未新增运行期依赖，`DI`来源无影响。
- `lina-review`结果：审查范围为`.agents/skills/lina-community-release-changelog/SKILL.md`和`openspec/changes/add-release-changelog-skill/`下的`OpenSpec`文件；已按`git status --short`和`git ls-files --others --exclude-standard`展开未跟踪目录，并将忽略的`temp/changelog.md`作为生成产物验证证据；未发现阻塞问题。

## Feedback

- [x] **FB-1**: 修复`plugin-full / plugins-2-of-5`中源码插件卸载用例受动态插件反向依赖状态污染导致的`GitHub Actions`失败

### FB-1 执行记录

- 根因：`linapro-demo-dynamic`在`plugin.yaml`中声明依赖`linapro-demo-source`。`plugins-2-of-5`分片先运行动态插件用例后会留下已安装的动态依赖方，源码插件生命周期用例只重置`linapro-demo-source`和`linapro-ops-demo-guard`状态，未清理`linapro-demo-dynamic`。因此`TC-1j`、`TC-1p`、`TC-1q`中的源码插件卸载被宿主反向依赖保护正确拦截，确认按钮保持禁用且不会发出`DELETE /api/v1/plugins/linapro-demo-source`请求，最终超时。
- 修复：在`apps/lina-plugins/linapro-demo-source/hack/tests/e2e/host-integration/TC001-source-plugin-lifecycle.ts`的`beforeEach`中同步重置`linapro-demo-dynamic`注册行，确保源码插件卸载场景独立于前序动态插件状态；同时增强`hack/tests/support/api/job.ts`的`expectSuccess`失败信息，输出请求`URL`、`errorCode`、`messageKey`和`messageParams`，便于后续定位类似`payload.code=55`问题。
- 影响分析：`i18n`无运行时文案、语言包、插件清单或接口文档源文本变更；缓存一致性无生产缓存、快照或失效策略变更；数据权限无产品接口、服务数据读写或租户边界变更；开发工具跨平台无长期脚本、`Makefile`或`linactl`入口变更；插件影响仅限`linapro-demo-source`插件自有`E2E`用例状态隔离，未修改插件运行时代码、清单、`SQL`或宿主插件契约；前端用户可观察产品行为无变更，测试可观察行为通过既有`E2E`覆盖。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/testing.md`、`.agents/rules/plugin.md`、`.agents/rules/i18n.md`、`.agents/rules/documentation.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/frontend-ui.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/data-permission.md`；插件根目录`apps/lina-plugins/linapro-demo-source/AGENTS.md`和`apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`均不存在。
- 验证命令：`openspec validate add-release-changelog-skill --strict`通过；`pnpm -C hack/tests run test:validate`通过；`pnpm -C hack/tests exec tsc --noEmit`通过；`git diff --check && git -C apps/lina-plugins diff --check`通过；使用隔离数据库`linapro_e2e_ci_fix`启动`make dev plugins=1`后，`E2E_DB_NAME=linapro_e2e_ci_fix E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 E2E_BASE_URL=http://127.0.0.1:5666 E2E_FRONTEND_PROXY_BACKEND_ORIGIN=http://localhost:9120 pnpm -C hack/tests test:module -- plugin:linapro-demo-source -- --grep "TC-1j|TC-1p|TC-1q"`通过，`TC-1j`、`TC-1p`、`TC-1q`共`3`条通过；同环境下`pnpm -C hack/tests test:module -- plugin:linapro-demo-dynamic -- --grep "TC-4c"`通过；同环境下`pnpm -C hack/tests test:module -- plugin:linapro-demo-source -- --grep "TC-3a"`通过；重建隔离库后先执行`pnpm -C hack/tests test:module -- plugin:linapro-demo-dynamic -- --grep "TC-4a"`制造动态插件依赖方已安装启用状态，再执行`pnpm -C hack/tests test:module -- plugin:linapro-demo-source -- --grep "TC-1j|TC-1p|TC-1q"`仍通过。验证结束后已执行`make stop`并恢复本地忽略的`apps/lina-core/manifest/config/config.yaml`，恢复后`SHA-256`为`386c520352ed1683cb2841e65327c8a4da93c9942ea04e4e561c19cdf3696632`。
- `lina-review`结果：审查范围为`hack/tests/support/api/job.ts`、`openspec/changes/add-release-changelog-skill/tasks.md`和`apps/lina-plugins/linapro-demo-source/hack/tests/e2e/host-integration/TC001-source-plugin-lifecycle.ts`；已按`git status --short`、`git ls-files --others --exclude-standard`和子模块状态检查工作区，未发现阻塞问题。剩余发布注意事项：插件测试修复位于`apps/lina-plugins`子模块，`GitHub Actions`只有在提交并推送该子模块修复、再更新父仓库子模块指针后才能使用该改动。
