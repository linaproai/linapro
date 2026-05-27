## Context

`make env.setup`通过`linactl env.setup`统一安装前端依赖和 Playwright 浏览器资源。当前 E2E 配置默认只运行`chromium`项目，且`use.headless`固定为`true`，没有默认浏览器`channel`。在这种约束下，Playwright 支持通过`--only-shell chromium`只下载 Chromium headless shell，避免完整 Chromium 下载成本。

## Goals / Non-Goals

**Goals:**

- 降低首次执行`make env.setup`时的 Playwright 浏览器下载体积和耗时。
- 保持 Linux 上系统依赖安装能力，避免 CI 或新开发机缺少运行依赖。
- 通过单元测试锁定`linactl env.setup`实际执行的 Playwright 命令。

**Non-Goals:**

- 不新增可配置参数或新的 Makefile 入口。
- 不改变 E2E 默认项目、浏览器类型、`headless`设置或运行脚本。
- 不支持 headed/debug 场景的完整浏览器自动安装；需要完整浏览器时仍由开发者按 Playwright 命令手动安装。

## Decisions

1. `env.setup`默认使用`pnpm exec playwright install --with-deps --only-shell chromium`。
   - 理由：该命令符合当前 headless E2E 默认运行方式，并保留`--with-deps`对 Linux 系统依赖的安装语义。
   - 替代方案：使用`PLAYWRIGHT_DOWNLOAD_HOST`镜像加速完整 Chromium 下载。该方式依赖外部镜像可用性，不能从根本上减少下载内容。

2. 不新增环境变量或`make`参数切换完整 Chromium。
   - 理由：仓库默认入口应服务最常见的 headless E2E 初始化路径；headed/debug 属于显式本地调试需求，可以使用 Playwright 原生命令安装完整浏览器。
   - 替代方案：添加`ENV_SETUP_PLAYWRIGHT_FULL=1`之类的开关。当前需求不需要扩展入口，增加开关会扩大维护和测试面。

3. 同步更新中英文 E2E README。
   - 理由：`hack/tests`已有目录级中英文说明文档，安装语义变化必须保持镜像文档一致。

## Risks / Trade-offs

- 使用`pnpm test:headed`、`pnpm test:ui`或`pnpm test:debug`时，headless shell 可能不足以覆盖完整浏览器需求。
  - 缓解：README 明确这些场景需要额外执行完整 Chromium 安装命令。
- 如果未来 E2E 默认切换到指定`E2E_BROWSER_CHANNEL`或完整 Chromium 新 headless 模式，默认安装内容需要重新评估。
  - 缓解：当前测试通过单元测试锁定命令，未来变更需要同步更新测试和文档。
