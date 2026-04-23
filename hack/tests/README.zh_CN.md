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
| `pnpm test:validate` | 校验 `TC` 唯一性、目录归属与 manifest 引用。 |
| `pnpm report` | 打开 Playwright `HTML` 报告。 |

模块范围示例：

- `iam:user`
- `settings:config`
- `monitor:operlog`
- `scheduler:job`
- `extension:plugin`

## 执行模型

套件以 `config/execution-manifest.json` 作为单一事实来源，统一维护：

- 历史目录到目标目录的迁移映射
- 模块范围定义
- `smoke` 用例清单
- 共享状态场景的串行执行边界

`pnpm test`、`pnpm test:full`、`pnpm test:smoke` 与 `pnpm test:module` 都通过 `scripts/run-suite.mjs` 执行。
运行器会把选中的文件拆分为并行池与串行池，使高共享状态场景仍能安全执行。

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

新增测试用例后，如果需要把它加入 `smoke` 套件、串行池或新的模块范围，请同步更新 `config/execution-manifest.json`。
