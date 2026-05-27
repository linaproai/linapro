## 1. 实现

- [x] 1.1 将`linactl env.setup`的 Playwright 安装命令改为下载 Chromium headless shell，并保留`--with-deps`。
- [x] 1.2 更新`linactl`单元测试，断言`env.setup`执行新的 Playwright 安装参数。
- [x] 1.3 同步更新`hack/tests/README.md`和`hack/tests/README.zh-CN.md`的前置依赖说明。

## 2. 验证与审查

- [x] 2.1 运行`cd hack/tools/linactl && go test ./... -count=1`。
- [x] 2.2 运行`openspec validate speed-up-playwright-env-setup --strict`和目标文件空白检查。
- [x] 2.3 记录影响分析，并按`lina-review`口径完成审查。

## 3. 执行记录

- i18n 影响评估：本变更只修改开发工具命令、E2E 说明文档和 OpenSpec 文档，不新增或修改运行时前端文案、菜单、路由、按钮、表单、表格、提示、宿主或插件语言包、`manifest/i18n`、`apidoc i18n JSON`或错误消息资源。
- 缓存一致性影响评估：本变更不涉及运行时缓存、分布式缓存、跨实例失效、共享修订号、翻译缓存或业务状态缓存。
- 数据权限影响评估：本变更不新增或修改列表、详情、导出、下载、聚合统计、批量信息、下拉候选、写操作、授权关系或插件数据访问路径。
- 开发工具跨平台影响评估：本变更修改`hack/tools/linactl`默认 Playwright 安装参数；仍通过 Go 工具链调用`pnpm exec playwright`，不新增平台专属脚本或 Shell 依赖。验证使用`cd hack/tools/linactl && go test ./... -count=1`。
- 测试策略：本变更属于开发工具初始化行为调整，使用`linactl`单元测试覆盖命令参数；不涉及用户可观察业务页面或端到端业务流程，不新增 E2E。
- 规则加载：已读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`，并使用`lina-feedback`、`openspec-propose`、`openspec-apply-change`、`lina-review`、`goframe-v2`和`karpathy-guidelines`技能。
- 验证：已运行`openspec validate speed-up-playwright-env-setup --strict`，通过。
- 验证：已运行`git diff --check -- hack/tools/linactl/command_env.setup.go hack/tools/linactl/main_test.go hack/tests/README.md hack/tests/README.zh-CN.md openspec/changes/speed-up-playwright-env-setup`，通过。
- 验证阻断：已运行`cd hack/tools/linactl && go test . -run TestRunEnvSetupInstallsFrontendAndPlaywright -count=1`，失败原因是当前工作区已有未完成的`apps/lina-core/pkg/plugin/pluginbridge/internal/hostservice`改动缺少`HostServiceFramework`、`CapabilityFramework`等符号，导致`linactl`包编译失败；该阻断不由本变更修改的`env.setup`代码引入。
- 验证阻断：已运行`cd hack/tools/linactl && go test ./... -count=1`，同样因当前工作区已有`pluginbridge`/`guest`符号缺失导致编译失败；待相关无关改动补齐或回退后需重新运行。
- 验证：为隔离当前主工作区无关未完成重构，已创建临时干净 worktree，应用本变更中`hack/tools/linactl/command_env.setup.go`和`hack/tools/linactl/main_test.go`的补丁后运行`cd hack/tools/linactl && go test ./... -count=1`，通过；临时 worktree 已删除。
- Review：已按`lina-review`口径完成本变更范围审查。审查范围限定为`linactl env.setup`命令参数、对应单元测试断言、E2E 中英文 README 和`speed-up-playwright-env-setup` OpenSpec 文档；未发现本变更范围内阻塞问题。剩余风险为当前工作区无关`pluginbridge`重构导致`linactl`测试无法完成，需在该无关编译问题解决后重跑完整`linactl`测试。
