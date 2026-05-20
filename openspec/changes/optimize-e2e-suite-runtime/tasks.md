## 1. CI 分片与基础治理

- [x] 1.1 修正 browser E2E workflow 的 PostgreSQL health check，显式使用 `pg_isready -U postgres -d linapro`。
- [x] 1.2 将 plugin-full E2E job 改为基于通用入口的分片执行，分片覆盖 `extension:plugin` 与 `plugins`。
- [x] 1.3 为 plugin-full 分片生成唯一 artifact 名称，确保 Playwright report、test-results、后端日志和前端日志不会互相覆盖。
- [x] 1.4 验证分片失败会让完整 verification suite 失败，并阻止依赖验证成功的后续 job。

## 2. Plugin-full 覆盖范围收敛

- [x] 2.1 梳理 plugin-full 需要保留的插件框架通用用例，明确菜单、权限、路由、i18n、任务、运行时资源的通用覆盖文件。
- [x] 2.2 收敛根 E2E manifest，使 plugin-full 不再选择依赖具体官方插件的根测试文件集合。
- [x] 2.3 确认 host-only 仍覆盖宿主全量能力，plugin-full 仍覆盖官方插件自有用例和插件生命周期。

## 3. 认证 fixture 与宿主慢用例优化

- [x] 3.1 在 `hack/tests/fixtures/auth.ts` 中新增不自动导航到 dashboard 的轻量认证页面 fixture，并保留现有 `adminPage` 行为。
- [x] 3.2 优先迁移菜单 CRUD 中适合直接进入业务路由的用例，减少重复 dashboard 加载。
- [x] 3.3 优先迁移文件管理中适合直接进入业务路由的用例，减少重复 dashboard 加载。
- [x] 3.4 评估并迁移角色 CRUD、参数导入、字典导入等相同模式的慢文件。

## 4. 插件 baseline 与普通插件用例优化

- [x] 4.1 在插件 E2E fixture/support 中新增幂等 baseline 辅助，支持一次性同步、安装、启用插件、加载可用 mock 数据并刷新插件投影。
- [x] 4.2 将普通插件页面测试中的重复 `ensureSourcePluginEnabled` 迁移为 suite 或 shard 级 baseline。
- [x] 4.3 确认插件生命周期测试仍显式控制安装、启用、禁用、卸载、上传、同步和清理状态，不被普通 baseline 干扰。

## 5. 生命周期大户重构

- [x] 5.1 重构官方源码插件生命周期用例，保留一个代表官方插件的完整 UI 生命周期，其余官方插件改为 API/contract smoke 加页面可访问性验证。
- [x] 5.2 重构 dynamic runtime 生命周期用例，区分 runtime 核心 UI 生命周期与 dynamic demo API/功能验证，合并重复装配并保留关键 UI 覆盖。
- [x] 5.3 复查源码插件生命周期用例，消除可合并或可 API 化的重复生命周期步骤。

## 6. 验证与验收记录

- [x] 6.1 运行 `openspec validate optimize-e2e-suite-runtime --strict` 并修复所有规范问题。
- [x] 6.2 运行受影响的 module scope E2E smoke，至少覆盖 `extension:plugin`、一个官方插件功能 scope、以及迁移后的宿主慢文件。
- [x] 6.3 记录 host-only 优化前后 wall clock、测试总耗时、最慢文件和最慢用例对比。
- [x] 6.4 记录 plugin-full 优化前后 wall clock、各分片耗时、最长分片和 runner minutes 变化。
- [x] 6.5 明确记录本变更不影响生产 API、数据库 schema、运行时缓存语义和 i18n 资源；若实现中新增可见文案或脚本入口，再同步补充对应治理说明。
- [x] 6.6 完成任务后执行 `/lina-review`，审查 CI 分片、fixture、baseline、慢用例重构和验证记录。

## 验证记录

- 已通过 `openspec validate optimize-e2e-suite-runtime --strict`。
- 已通过 `pnpm -C hack/tests exec tsc --noEmit`。
- 已通过 `pnpm -C hack/tests test:validate`。
- 已通过 `git diff --check`。
- 已验证本地服务 `http://127.0.0.1:5666` 与 `http://127.0.0.1:8080/api/v1/health` 可访问。
- 初次尝试运行 `pnpm -C hack/tests test:module -- settings:file --grep TC001a` 与 `pnpm -C hack/tests test:module -- monitor:operlog --grep TC001a` 时，Playwright global setup 因本机缺少 `chromium_headless_shell-1217` 浏览器二进制失败；随后尝试 `pnpm -C hack/tests exec playwright install chromium`，下载长时间无进展并被终止。
- 已使用本机 Google Chrome 通道完成最小浏览器 smoke：`E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests test:module -- settings:file --grep "TC001a"`，结果 `1 passed (8.0s)`。
- 已使用本机 Google Chrome 通道完成普通插件 baseline smoke：`E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests test:module -- monitor:operlog --grep "TC001a"`，结果 `1 passed (9.9s)`。
- 已使用本机 Google Chrome 通道完成代表性插件生命周期 smoke：`E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests test:module -- extension:plugin --grep "TC001a"`，结果 `1 passed (29.9s)`。
- host-only 优化前基线来自用户提供日志：job 约 36 分钟，Playwright 报告 `197 passed (25.1m)`；本次已迁移菜单 CRUD、文件管理、角色 CRUD、参数导入、字典导入等慢用例使用不预加载 dashboard 的 `authenticatedPage`，预期减少重复 dashboard 首屏加载成本，最终耗时需以 CI artifact 复核。
- plugin-full 优化前基线来自用户提供日志：job 约 2 小时，`pnpm test` 约 112 分钟；本次改为 `extension:plugin` 与 `plugins` 两个通用分片，最长分片预计由插件自有用例集合决定，最终 wall clock 与 runner minutes 需以 CI artifact 复核。
- 本变更只调整 CI workflow、E2E runner manifest、Playwright fixture 和测试代码，不修改生产 API、数据库 schema、运行时缓存语义或用户可见功能；未新增或修改前端运行时文案、插件 manifest i18n、apidoc i18n JSON，确认不需要同步 i18n 资源。

## Feedback

- [x] **FB-1**: 区分无 `apps/lina-plugins` 的宿主模块 scope 与需要官方插件工作区的 plugin-full 接缝 scope。
- [x] **FB-2**: 收敛 plugin-full scope，只保留 `plugins` 与 `plugin:<plugin-id>` 作为源码插件自有用例的通用选择入口。
- [x] **FB-3**: 根 `hack/tests` E2E 代码与配置不得耦合任何具体官方源码插件 ID，插件相关用例必须闭环到对应插件目录。
- [x] **FB-4**: 根路径 E2E 测试文件、配置、测试数据和 baseline 不得耦合任何具体插件信息，插件相关测试资产必须放在对应插件目录。
- [x] **FB-5**: E2E 测试文件名前缀不再全局递增，改为按当前模块目录从 `TC001` 开始递增。
- [x] **FB-6**: `extension:plugin` 分片中动态插件资源“仅本人”数据权限和插件管理动作权限夹具在 plugin-full 环境失败。
- [x] **FB-7**: GitHub Actions host-only E2E 仍运行部分 plugin-full 或插件依赖用例，导致共享种子和宿主断言失败。
- [x] **FB-8**: GitHub Actions plugin-full E2E 中动态插件示例记录和英文布局回归用例存在跨用例状态泄漏。
- [x] **FB-9**: 完整 E2E 中角色新增/编辑抽屉会在异步初始化完成后覆盖已填写字段，导致提交未发出角色保存请求。
- [x] **FB-10**: Nightly plugin-full `plugins` scope 仍在单个 CI job 中串行执行全部源码插件自有 E2E，导致最长分片接近 1 小时。
- [x] **FB-11**: GitHub Actions `plugins-1-of-5` 分片中动态插件禁用后侧边栏菜单隐藏断言存在菜单投影刷新竞态。

## Feedback 验证记录

- 已移除需要具体官方插件环境的根目录接缝文件集合，根 `extension:plugin` 只覆盖宿主插件框架、动态测试插件与通用插件治理能力。
- 已新增 `pnpm test:host:module -- <scope>`，用于在未安装 `apps/lina-plugins` 时只运行指定宿主模块中可在 host-only 环境执行的用例。
- 已让 `moduleRequiresPluginWorkspace` 根据 scope entry 自动识别 `plugins/`、`apps/lina-plugins/` 与 `plugin:<plugin-id>`，避免通用插件入口在缺少 submodule 时被误判为可运行。
- 已在 `pnpm test:validate` 中增加治理校验：`host:` 前缀的 module scope 不允许依赖 `apps/lina-plugins`。
- 已通过 `pnpm -C hack/tests test:host:module -- settings:file --list`，确认 host-only 单模块入口可列出宿主文件且不执行真实浏览器用例。
- 已通过 `pnpm -C hack/tests test:host:module -- iam:user --list`，确认混合宿主 scope 会过滤插件敏感文件并保留可在主框架环境运行的宿主用例。
- 已通过 `pnpm -C hack/tests test:host:module -- plugin:<plugin-id> --list` 的预期失败验证，确认源码插件通用入口在 host-only module 模式下会被拒绝。
- 已通过 `pnpm -C hack/tests test:module -- <removed-plugin-alias> --list` 的预期失败验证，确认已移除的源码插件业务别名 scope 不再可用。
- 已移除按官方插件业务模块命名的长期别名 scope，源码插件自有用例统一使用 `plugins` 或 `plugin:<plugin-id>`。
- CI plugin-full 分片改用 `extension:plugin` 与 `plugins` 两个通用入口，不在根 workflow matrix 中枚举具体官方插件 ID。
- 已将具体官方插件的 E2E 用例、页面对象与服务依赖 baseline 迁移到对应 `apps/lina-plugins/<plugin-id>/hack/tests/` 目录；根 `hack/tests` 只保留宿主通用测试资产和参数化 helper。
- 已将根 E2E 的插件 ID 防回归校验改为从可选 `apps/lina-plugins/*/plugin.yaml` 动态发现插件标识，不在 `hack/tests` 治理脚本中维护官方插件 ID denylist；当未安装插件工作区时，该校验不要求插件工作区存在。
- 已通过基于 `apps/lina-plugins/*/plugin.yaml` 动态发现插件标识的根 E2E 治理扫描，确认根 E2E 资产和 CI workflow 不再包含具体官方源码插件 ID 或插件业务名称。
- 已通过旧接缝 scope 与具体源码插件入口静态扫描，确认旧 `plugin-host-seams` scope 和具体 `plugin:<concrete-id>` 入口无残留。
- 已通过 `rg -n "@playwright/test" apps/lina-plugins -g '*.ts'`，确认插件测试代码改用宿主封装的 Playwright 支持入口。
- 已通过 `pnpm -C hack/tests exec tsc --noEmit`。
- 已通过 `pnpm -C hack/tests test:validate`，结果 `Validated 233 E2E test files across 17 scopes. Smoke files: 6. Serial files: 216.`。
- 已通过 `pnpm -C hack/tests test:service-deps`，结果 `Service dependency governance passed: 28 files, 113 baseline critical constructor calls.`。
- 已通过 `openspec validate optimize-e2e-suite-runtime --strict`。
- 已通过 `git diff --check`。
- 本次反馈属于 E2E runner 和项目治理修正，不改变生产 API、数据库 schema、运行时缓存语义或用户可见业务功能；仅新增 README 中的命令入口说明，未新增前端运行时文案、插件 manifest i18n 或 apidoc i18n JSON，确认不需要同步 i18n 资源。
- 已将 `hack/tests` 根目录的插件耦合治理改为动态发现具体源码插件标识，仅拒绝具体源码插件 ID、根 manifest 中的 `plugins/<concrete-id>` / `apps/lina-plugins/<concrete-id>` 条目、以及根 service dependency baseline 中的插件路径；`plugins/sync`、`plugins/dynamic` 等宿主通用插件框架 API 保持允许。
- 已将所有根目录和源码插件目录下 E2E 测试文件重命名为模块本地 `TC{NNN}`，每个模块目录从 `TC001` 连续递增；`pnpm -C hack/tests test:validate` 会拒绝旧四位全局编号、目录内重复编号和目录内不连续编号。
- 已将根路径 E2E 不耦合具体源码插件信息、插件测试资产闭环在插件目录、以及 E2E 测试文件按模块目录本地 `TC001` 递增的规则同步到 `AGENTS.md` 项目规范和 `.agents/skills/lina-review/SKILL.md` 审查清单。
- 已通过本次规则同步后的治理验证：`openspec validate optimize-e2e-suite-runtime --strict`、`git diff --check -- AGENTS.md .agents/skills/lina-review/SKILL.md openspec/changes/optimize-e2e-suite-runtime/tasks.md`、`rg -n '项目根路径下的 `E2E`|模块本地递增|TC\{NNN\}|具体源码插件 `ID`|根路径 E2E 不耦合' AGENTS.md .agents/skills/lina-review/SKILL.md openspec/changes/optimize-e2e-suite-runtime/tasks.md`。
- 已通过 `pnpm -C hack/tests test:module -- extension:plugin --list`，确认宿主插件框架 scope 只选择 11 个根目录通用插件框架用例文件。
- 已通过 `pnpm -C hack/tests test:module -- plugins --list`，确认源码插件全量入口选择 138 个插件自有用例文件，文件均来自 `apps/lina-plugins/<plugin-id>/hack/tests/e2e/`。
- 已通过 `pnpm -C hack/tests test:host:module -- iam:user --list`，确认 host-only 单模块入口选择 9 个宿主用户模块用例文件。
- 已通过 `pnpm -C hack/tests test:host:module -- 'plugin:<plugin-id>' --list` 的预期失败验证，确认 host-only 模块入口拒绝需要 `apps/lina-plugins` 的插件 scope。
- FB-6 修复了 generic plugin resource 查询层与宿主 `sys_role.data_scope` 枚举不一致的问题：`dataScope=4` 现在按“仅本人”过滤，`dataScope=3` 仍按部门过滤；同时将插件管理动作权限 E2E 的查询角色改为不依赖组织插件的全量数据权限，因为该用例只验证插件动作权限。
- FB-6 已补充后端单测 `TestResolvePluginResourceDataScopeModeUsesHostScopeValues`，防止插件资源数据权限映射再次从宿主角色枚举漂移。
- FB-6 已通过 `cd apps/lina-core && go test ./internal/service/plugin/internal/integration -count=1`。
- FB-6 已通过宿主启动/路由绑定包测试：`cd apps/lina-core && go test ./internal/cmd -count=1`。
- FB-6 已通过 E2E 精确回归：`E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests exec playwright test e2e/extension/plugin/TC002-plugin-permission-governance.ts e2e/extension/plugin/TC007-plugin-management-action-permissions.ts --config playwright.config.ts --project=chromium --workers=1`，结果 `4 passed (55.4s)`。
- FB-6 已通过完整分片回归：`E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests test:module -- extension:plugin`，结果 `26 passed, 1 skipped (2.4m)`。
- FB-6 已通过 `pnpm -C hack/tests exec tsc --noEmit`、`pnpm -C hack/tests test:validate`、`openspec validate optimize-e2e-suite-runtime --strict` 和 `git diff --check`。
- FB-6 i18n 影响：本次只调整后端数据权限映射、后端单测和 E2E 测试夹具，不新增或修改前端运行时文案、插件 manifest i18n 或 apidoc i18n JSON。
- FB-6 缓存一致性影响：本次不新增缓存；插件资源查询继续按请求上下文中的数据权限快照即时应用数据库过滤，不改变插件 runtime freshness、启用状态快照或跨实例失效语义。
- FB-7 已让 `pnpm test:host` 的 Playwright 子进程显式注入 `E2E_HOST_ONLY_PLUGINS=1`，普通 plugin-full 入口仍保留外部传入值或默认 `0`，使测试代码可区分 host-only 与 plugin-full 环境。
- FB-7 已修复动态插件热升级用例的宿主菜单投影刷新：启用动态插件后通过访问工作台并刷新页面重新加载宿主投影，不再依赖 focus 事件，也不直接改写角色菜单授权表绕过权限缓存。
- FB-7 已将英文内置治理数据本地化用例按 host-only 环境收敛断言：host-only 下任务日志只断言宿主内置任务，并跳过由源码插件提供的登录日志和操作日志接口断言。
- FB-7 已将调度任务用例按 host-only 环境区分宿主内置任务集合，host-only 只断言 `任务日志清理` 和 `在线会话清理`，plugin-full 继续覆盖 `服务监控采集`。
- FB-7 已将字典数据面板用例从插件拥有的 `sys_oper_type` 切换到宿主内置的 `sys_menu_status`，避免 host-only 环境缺少插件字典种子导致失败。
- FB-7 已修复用户角色 POM 断言：运行时根据可见表头 `角色` / `Roles` 解析 VXE `colid`，再定位同一行对应单元格，避免组织/租户列显示顺序变化或 VXE 运行时列 ID 导致读错列。
- FB-7 已通过 host-only 文件选择验证：`pnpm -C hack/tests test:host -- --list`，结果选择 `95` 个文件，其中 `17` 个并行、`78` 个串行。
- FB-7 已通过 host-only 精确回归：`E2E_HOST_ONLY_PLUGINS=1 E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests exec playwright test e2e/extension/plugin/TC003-plugin-hot-upgrade.ts e2e/i18n/TC005-english-built-in-governance-data-localization.ts e2e/iam/user/TC008-user-role.ts e2e/scheduler/job/TC002-job-handler-crud.ts e2e/settings/dict/TC006-dict-data-export.ts --config playwright.config.ts --project=chromium --workers=1`，结果 `14 passed, 1 skipped (1.6m)`；跳过项为 host-only 环境下由源码插件提供的日志接口断言。
- FB-7 已通过用户角色单文件回归：`E2E_HOST_ONLY_PLUGINS=1 E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests exec playwright test e2e/iam/user/TC008-user-role.ts --config playwright.config.ts --project=chromium --workers=1`，结果 `4 passed (37.2s)`。
- FB-7 已通过 `pnpm -C hack/tests exec tsc --noEmit`、`pnpm -C hack/tests test:validate`、`openspec validate optimize-e2e-suite-runtime --strict` 和 `git diff --check`。
- FB-7 i18n 影响：本次只调整 E2E 环境分支、断言和跳过逻辑，不新增或修改前端运行时文案、插件 manifest i18n 或 apidoc i18n JSON。
- FB-7 缓存一致性影响：本次不修改生产缓存逻辑；动态插件用例改为通过 UI 重新加载宿主投影观察实际缓存刷新效果，避免测试侧直接写权限表造成缓存状态与真实运行路径不一致。
- FB-8 已为 Vite 开发服务补充 `/x` 代理到后端运行时，避免动态插件页面在 dev origin 下请求 `/x/linapro-demo-dynamic/demo-records` 时命中 Vite 404；同时补齐该配置的 node 级 TypeScript 校验保护。
- FB-8 已让动态插件运行时 E2E 清理 `sys_plugin_migration` 与示例记录表状态，分页种子显式写入 `tenant_id=0`，并在页面断言前通过后端 `/x/linapro-demo-dynamic/demo-records` API 等待示例数据可见。
- FB-8 已让英文运行时页面巡检在每个用例前卸载并清理动态插件数据，防止跨用例迁移状态泄漏；源码示例英文布局回归在前置步骤中显式启用 `linapro-org-core`，确保英文侧栏断言自包含。
- FB-8 已通过动态插件与源码示例精确回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:8080 pnpm -C hack/tests exec playwright test ../apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts ../apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC003-english-runtime-page-audit.ts ../apps/lina-plugins/linapro-demo-source/hack/tests/e2e/host-integration/TC006-english-layout-regression.ts --config playwright.config.ts --project=chromium --workers=1`，结果 `15 passed (4.4m)`。
- FB-8 已通过 plugin-full 分片回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:8080 pnpm -C hack/tests test:module -- extension:plugin`，结果 `26 passed, 1 skipped (2.2m)`；`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:8080 pnpm -C hack/tests test:module -- plugins`，结果 `272 passed, 7 skipped (25.7m)`。
- FB-8 i18n 影响：本次只调整 E2E 状态清理、测试前置条件、后端可见性等待和 Vite dev proxy，不新增或修改前端运行时文案、插件 manifest i18n 或 apidoc i18n JSON；现有英文布局和英文动态页面 E2E 已覆盖英文展示不回退。
- FB-8 缓存一致性影响：本次不修改生产缓存逻辑；动态插件 E2E 通过卸载、清理迁移记录、重新安装启用和 API 可见性等待验证真实运行路径，避免测试环境残留的本地或数据库状态影响跨用例一致性。
- FB-9 已修复角色页面 POM 在新增/编辑角色抽屉中与权限树异步初始化的竞态：打开抽屉后等待权限工具栏和首行菜单树渲染完成，普通创建/编辑路径预先标记权限导览已读，提交时等待 `POST /api/v1/role` 或 `PUT /api/v1/role/{id}` 响应并确认抽屉关闭。
- FB-9 已通过角色精确回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:8080 pnpm -C hack/tests exec playwright test e2e/iam/role/TC001-role-crud.ts e2e/iam/role/TC004-role-permission-drawer-close.ts --config playwright.config.ts --project=chromium --workers=1`，结果 `13 passed (52.6s)`。
- FB-9 i18n 影响：本次只调整 E2E 页面对象的等待与导览处理，不新增或修改用户可见文本、语言包、插件 manifest i18n 或 apidoc i18n JSON。
- FB-9 缓存一致性影响：本次不修改生产缓存逻辑；角色 POM 仅等待前端权限树初始化和角色保存响应，不改变权限缓存、角色授权缓存或跨实例失效语义。
- FB-10 日志分析结论：用户提供的 GitHub Actions 日志显示 `plugin-full / plugins` job 从 `2026-05-19T23:55:31` 开始执行 Playwright，到 `2026-05-20T00:56:21` 结束，`138` 个文件全部进入 serial pool，`272 passed, 7 skipped (1.0h)`；最慢文件为 `linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts`，耗时 `7.4m`。按日志聚合，`linapro-org-core` 约 `12.1m`、`linapro-content-notice` 约 `10.1m`、`linapro-demo-dynamic` 约 `9.3m`，主瓶颈是单 job 串行执行所有源码插件自有用例。
- FB-10 已将 plugin-full CI matrix 中的 `plugins` 单分片拆为 `plugins-1-of-5` 至 `plugins-5-of-5`，每个分片继续使用通用 `plugins` scope，并通过 Playwright 原生 `--shard=N/5` 在独立 runner、独立 PostgreSQL 和独立 plugin-full 服务实例中执行，避免在同一数据库内并发污染；`extension:plugin` 分片保持不变。
- FB-10 已通过用户日志耗时映射估算 `plugins --shard=N/5` 分布：`1/5` 约 `17.9m`、`2/5` 约 `11.7m`、`3/5` 约 `11.1m`、`4/5` 约 `13.4m`、`5/5` 约 `4.2m` 的测试用例耗时；相比原单 job 约 `58.3m` 的已解析测试用例耗时，最长测试体量预计下降约 `69%`。选择 5 分片是因为 6 分片估算最长仍约 `17.5m`，收益很小；7 分片最长约 `13.9m` 但会额外增加两个 CI job。
- FB-10 已通过 `pnpm -C hack/tests test:module -- plugins -- --shard=1/5 --list`、`--shard=2/5 --list`、`--shard=3/5 --list`、`--shard=4/5 --list`、`--shard=5/5 --list` 验证每个分片可解析并列出测试。
- FB-10 已修复本次验证中暴露的既有 E2E 治理问题：将 `hack/tests/e2e/extension/plugin/TC0243-plugin-status-switch-feedback.ts` 重命名为模块本地连续编号 `TC012-plugin-status-switch-feedback.ts`，并同步测试标题为 `TC-12`，使 `pnpm -C hack/tests test:validate` 可通过。
- FB-10 i18n 影响：本次只调整 GitHub Actions E2E job matrix 和 OpenSpec 任务记录，不新增或修改前端运行时文案、插件 manifest i18n 或 apidoc i18n JSON。
- FB-10 缓存一致性影响：本次不修改生产缓存逻辑；CI 层把 `plugins` scope 分散到独立 runner 与独立服务实例中执行，不改变运行时缓存、插件启用状态快照、i18n 缓存或跨实例失效语义。
- Review(FB-10): 已完成 `lina-review` 审查。审查范围来源：`git status --short --ignore-submodules=none`、`git ls-files --others --exclude-standard`、`openspec status --change optimize-e2e-suite-runtime --json`、`git diff -- .github/workflows/reusable-test-verification-suite.yml openspec/changes/optimize-e2e-suite-runtime/tasks.md`。确认本次只将 plugin-full `plugins` scope 从单个 CI job 拆为 5 个 Playwright shard，仍使用通用 `plugins` 入口和 plugin-full 服务启动语义；artifact 名称继续包含 `matrix.shard.name`，不会覆盖其他分片证据。未修改生产 Go/前端代码、业务 API、数据库 schema、运行时缓存逻辑、数据权限逻辑或 i18n 资源。严重问题 0；警告 0。
- 本次完整单元测试已通过 host-only Go 单元测试：`cp apps/lina-core/manifest/config/config.template.yaml apps/lina-core/manifest/config/config.yaml && make init confirm=init rebuild=true && make pack.assets && LINA_TEST_PGSQL_LINK='pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable' make test.go plugins=0 race=true verbose=true`。
- 本次完整单元测试已通过 plugin-full Go 单元测试：`cp apps/lina-core/manifest/config/config.template.yaml apps/lina-core/manifest/config/config.yaml && make init confirm=init rebuild=true && make pack.assets && LINA_TEST_PGSQL_LINK='pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable' make test.go plugins=1 race=true verbose=true`。
- 本次完整前端单元测试已通过：`pnpm -C apps/lina-vben test:unit`，结果 `Test Files 42 passed (42)`、`Tests 347 passed (347)`。
- 本次完整 host-only E2E 已通过：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:8080 pnpm -C hack/tests test:host`，结果 `244 passed, 1 skipped (14.6m)`。
- 本次完整 plugin-full E2E 已通过：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:8080/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:8080 pnpm -C hack/tests test`，结果 `516 passed, 8 skipped (42.8m)`。
- 本次最终静态与治理验证已通过：`pnpm -C apps/lina-vben exec tsc -p apps/web-antd/tsconfig.node.json --noEmit`、`pnpm -C hack/tests exec tsc --noEmit`、`pnpm -C hack/tests test:validate`、`openspec validate optimize-e2e-suite-runtime --strict` 和 `git diff --check`。
- FB-11 日志分析结论：用户提供的 GitHub Actions 日志显示 `plugin-full / plugins-1-of-5` 于 `2026-05-20T01:54:31Z` 和 retry `2026-05-20T01:55:28Z` 均失败在 `linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts` 的 `TC-1k`，错误为禁用插件后 `expectSidebarMenuHidden("动态插件示例")` 仍观察到侧边栏菜单可见；同一分片其余 `55` 个测试通过，根因是 UI 开关状态已切换但当前页面的菜单投影刷新存在 CI 时序竞态。
- FB-11 已在 `TC001-runtime-wasm-lifecycle.ts` 中新增 `expectCurrentUserRouteVisible`，在禁用 `linapro-demo-dynamic` 后轮询 `menus/all`，确认当前用户路由投影已不包含 `动态插件示例`，再刷新页面并断言侧边栏隐藏，避免断言抢在菜单投影刷新前执行。
- FB-11 已通过精确回归：`E2E_BROWSER_CHANNEL=chrome E2E_BASE_URL=http://127.0.0.1:5666 E2E_API_BASE_URL=http://127.0.0.1:9120/api/v1/ E2E_PUBLIC_BASE_URL=http://127.0.0.1:9120 pnpm -C hack/tests exec playwright test ../apps/lina-plugins/linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts --config playwright.config.ts --project=chromium --workers=1 --grep "TC-1k"`，结果 `1 passed (1.3m)`。
- FB-11 已通过 `pnpm -C hack/tests exec tsc --noEmit`、`pnpm -C hack/tests test:validate`、`openspec validate optimize-e2e-suite-runtime --strict`、`git diff --check` 和 `git -C apps/lina-plugins diff --check`。
- FB-11 i18n 影响：本次只调整 E2E 等待与断言顺序，不新增或修改前端运行时文案、插件 manifest i18n 或 apidoc i18n JSON。
- FB-11 缓存一致性影响：本次不修改生产缓存逻辑；E2E 通过后端 `menus/all` 轮询确认插件禁用后的菜单投影失效已完成，再刷新当前页面观察真实侧栏结果。
- Review(FB-11): 已完成 `lina-review` 审查。审查范围来源：`git status --short --ignore-submodules=none`、`git -C apps/lina-plugins status --short --untracked-files=all`、`openspec status --change optimize-e2e-suite-runtime --json`、`git diff -- openspec/changes/optimize-e2e-suite-runtime/tasks.md` 和 `git -C apps/lina-plugins diff -- linapro-demo-dynamic/hack/tests/e2e/runtime/TC001-runtime-wasm-lifecycle.ts`。确认本次只调整动态插件 E2E 对菜单投影刷新的等待与 OpenSpec 反馈记录；未修改生产 Go/前端代码、业务 API、数据库 schema、运行时缓存逻辑、数据权限逻辑或 i18n 资源。严重问题 0；警告 0。
