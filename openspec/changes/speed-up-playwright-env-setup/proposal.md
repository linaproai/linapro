## Why

当前`make env.setup`会为默认`headless` E2E 套件下载完整 Chromium 浏览器，首次初始化和 CI 缓存预热成本偏高。仓库默认测试只使用 Chromium `headless` 模式，应优先下载 Playwright 的 Chromium headless shell，同时保留 Linux 系统依赖安装能力。

## What Changes

- 将`make env.setup`触发的 Playwright 安装命令从完整 Chromium 下载调整为 Chromium headless shell 下载。
- 保留`--with-deps`参数，确保 Linux 环境仍安装 Playwright 运行所需系统依赖。
- 更新`linactl env.setup`单元测试，锁定新的命令参数。
- 同步更新 E2E 测试说明文档，明确默认安装的是 headless shell，以及 headed/debug 场景需要完整浏览器。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `project-setup`：开发环境初始化入口默认安装 Playwright Chromium headless shell，而不是完整 Chromium 浏览器。

## Impact

- 影响`hack/tools/linactl`中的`env.setup`命令实现和对应单元测试。
- 影响`hack/tests/README.md`与`hack/tests/README.zh-CN.md`中的 E2E 前置依赖说明。
- 不改变公开 HTTP API、生产后端服务、数据库结构、前端运行时页面、插件资源、缓存策略或数据权限边界。
