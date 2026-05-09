# E2E 测试套件

该目录承载 LinaPro 默认管理工作台与宿主-插件集成场景的 Playwright `E2E` 测试套件。

## 目录结构

```text
hack/tests/
  config/        执行清单与套件治理配置
  debug/         与用例树隔离的临时调试脚本
  e2e/           仅存放 TC 测试用例文件
  fixtures/      共享 Playwright fixture 与认证辅助
  pages/         页面对象
  scripts/       套件运行脚本与校验脚本
  support/       共享 helper，例如 API 工具与 UI 等待工具
  temp/          运行时产物，例如生成的 storage state
```

`e2e/` 目录不再沿用历史上的 `system/` 大目录，而是按稳定能力边界组织：

- `auth/`、`dashboard/`、`about/`
- `iam/`
- `settings/`
- `org/`
- `content/`
- `monitor/`
- `scheduler/`
- `extension/`

## 命名规则

- 测试文件必须使用 `TC{NNNN}-{brief-name}.ts`。
- `TC` 编号在整个套件范围内全局唯一。
- `hack/tests/e2e/` 下只允许存放真正的 `TC` 用例文件。
- 共享 helper 必须放在 `fixtures/`、`support/`、`scripts/` 或 `debug/` 中。

## 执行入口

| 命令 | 用途 |
| --- | --- |
| `pnpm test` | 运行完整的分层 `E2E` 套件。 |
| `pnpm test:full` | 显式运行完整的分层 `E2E` 套件。 |
| `pnpm test:smoke` | 运行预定义的高价值 `smoke` 套件。 |
| `pnpm test:module -- <scope>` | 运行 `execution manifest` 中声明的模块范围。 |
| `pnpm test:sqlite` | 准备基于配置文件的 SQLite 数据库，重启应用，并运行 SQLite 专用 E2E 用例。 |
| `pnpm test:sqlite:e2e-smoke` | 只运行 SQLite 浏览器启动/登录 E2E 用例。 |
| `pnpm test:validate` | 校验 `TC` 唯一性、目录归属与 manifest 引用。 |
| `pnpm report` | 打开 Playwright `HTML` 报告。 |

模块范围示例：

- `iam:user`
- `settings:config`
- `monitor:operlog`
- `scheduler:job`
- `extension:plugin`
- `dialect`

`pnpm test:sqlite` 是完整 SQLite 专用通道。脚本会先备份
`apps/lina-core/manifest/config/config.yaml`，再把
`database.default.link=sqlite::@file(./temp/sqlite/linapro.db)` 写入该配置文件，
随后执行 `make init confirm=init rebuild=true`、`make mock confirm=mock`、
启动 `make dev`，运行 `TC0164` 到 `TC0166`，最后恢复原配置。后端运行时仍然只从配置文件解析当前数据库方言。
`pnpm test:sqlite:e2e-smoke` 使用相同准备流程，但只运行 `TC0164`。主 CI 不再为 SQLite 安装前端依赖或 Playwright，而是调用 `hack/tests/scripts/run-sqlite-smoke.sh`，只启动后端并检查 SQLite 启动警告、health 单节点状态和管理员登录接口。

## 执行模型

套件以 `config/execution-manifest.json` 作为单一事实来源，统一维护：

- 历史目录到目标目录的迁移映射
- 模块范围定义
- `smoke` 用例清单
- 共享状态场景的串行执行边界
- 串行隔离类别与有理由的并行例外

`pnpm test`、`pnpm test:full`、`pnpm test:smoke` 与 `pnpm test:module` 都通过 `scripts/run-suite.mjs` 执行。
运行器会把选中的文件拆分为并行池与串行池，使高共享状态场景仍能安全执行。
每次执行都会打印选中文件数、并行文件数、串行文件数、并行 worker 数，以及串行池覆盖的隔离类别摘要。

## 隔离类别

当测试文件或目录会修改或依赖可能影响其他文件的共享状态时，需要在 `config/execution-manifest.json` 的 `serialIsolation` 中声明分类。

| 分类 | 适用场景 |
| --- | --- |
| `authSession` | 验证共享认证浏览器状态的测试，例如登出。 |
| `pluginLifecycle` | 插件同步、安装、启用、禁用、卸载、上传或升级流程。 |
| `runtimeI18nCache` | 运行时语言包版本、ETag 检查与语言缓存重校验。 |
| `systemConfig` | 系统参数与公共前端配置变更。 |
| `dictionaryData` | 字典类型或字典数据新增、导入、编辑、删除与级联场景。 |
| `permissionMatrix` | 菜单、角色、按钮权限与插件生成权限矩阵变更。 |
| `sharedDatabaseSeed` | 依赖 fixture 加载的共享 seed 或 mock 数据的测试。 |
| `filesystemArtifact` | 插件包、运行时插件或其他共享运行时产物变更。 |

只读测试在使用 fixture 管理前置条件且数据局部唯一时，应继续保留在并行池。
如果某个高风险模式确认可以安全并行，需要新增 `parallelIsolationAllowlist`，写明文件、分类和原因。
校验器会拒绝缺失分类的串行文件，以及没有原因说明的并行例外。

## Fixture 前置条件

测试文件必须可以独立运行。
源码插件前置条件应统一走 `fixtures/plugin.ts`，由它负责同步源码插件、按需安装或启用、刷新前端插件投影，并在存在匹配插件 mock SQL 时加载 mock 数据。
创建用户、部门、岗位、通知、文件、插件、导入行或导出产物的测试，应使用唯一名称或稳定测试前缀，并在 `finally`、`afterEach` 或 `afterAll` 中自行清理。

## 缓存重校验

缓存与 ETag 测试应验证协议语义，而不是假设完整回归期间资源版本不会变化。
条件请求必须证明请求携带了预期前置条件。
当 ETag 仍匹配时可以接受 `304 Not Modified`；当资源版本已合法刷新时，只能接受带有新 ETag、且新 ETag 不同于缓存值并包含有效响应体的 `200 OK`。

## 认证态复用

大多数后台已登录测试会复用预生成的管理员 `storageState`。
该文件由 `global-setup.ts` 在每轮执行前重新生成，并写入 `temp/storage-state/admin.json`。
认证主题用例在需要直接验证登录行为时，仍然保留真实登录链路。

## 等待策略

高频页面对象应优先复用 `support/ui.ts` 中的状态型等待 helper，而不是固定睡眠。
优先等待以下状态：

- 路由就绪
- 表格可见且加载遮罩消失
- 弹窗就绪且骨架屏消失
- 下拉面板可见
- 确认弹层出现

只有在确实存在明确业务原因、且无法用确定性 UI 信号表达时，才允许保留固定的 `waitForTimeout`。

## 治理要求

新增、重命名或迁移测试文件后，都应执行 `pnpm test:validate`。
校验脚本会检查：

- 重复 `TC` 编号
- `e2e/` 下混入非 `TC` 文件
- 测试文件落在未允许的模块目录下
- `smoke` 与串行清单中的失效引用
- 缺失串行隔离分类的文件
- 仍处于并行池的高风险共享状态模式
- 没有原因说明的并行隔离例外

新增测试用例后，如果需要把它加入 `smoke` 套件、串行池或新的模块范围，请同步更新 `config/execution-manifest.json`。
