## Why

当前`linactl ctrl`和`linactl dao`需要在本机安装或自动下载外部`gf`可执行文件，导致代码生成路径依赖第三方工具的本机状态和`latest`发布版本，用户体验割裂且生成结果存在版本漂移风险。

本次变更将 GoFrame CLI 代码生成能力内嵌到`linactl`工具链中，由仓库`go.mod`锁定版本，并通过隐藏子进程隔离 GoFrame CLI 的命令行副作用。

## What Changes

- `linactl ctrl`和`linactl dao`不再检测、下载、安装或调用本机`gf`可执行文件。
- `hack/tools/linactl`显式依赖`github.com/gogf/gf/cmd/gf/v2`，并与宿主使用的`github.com/gogf/gf/v2`版本保持一致。
- 新增仅供内部使用的隐藏 GoFrame 执行入口，用于执行白名单内的`gen ctrl`和`gen dao`命令。
- `linactl ctrl`和`linactl dao`通过隐藏子进程调用内嵌 GoFrame CLI，保持`apps/lina-core`作为生成工作目录。
- 废弃或移除外部`gf`下载安装流程，避免默认开发路径继续依赖 GitHub release 的`latest`二进制。
- 补充`linactl`单元测试和代码生成 smoke 验证，确保生成命令不依赖`PATH`中的`gf`。

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `linactl-build-tool-consolidation`：扩展`linactl`统一承载仓库开发工具实现的要求，新增 GoFrame controller 和 DAO/DO/Entity 代码生成入口的内嵌执行约束。

## Impact

- 影响代码：`hack/tools/linactl`命令注册、`ctrl`/`dao`命令实现、`internal/goframecli`组件、`go.mod`和相关测试。
- 影响文档：`hack/tools/linactl`说明文档需要更新，明确开发者不再需要单独安装`gf`。
- 影响依赖：新增`github.com/gogf/gf/cmd/gf/v2`作为`linactl`工具依赖，并要求与 GoFrame runtime 版本同步。
- 不影响`apps/lina-core`核心领域契约、HTTP API、数据库 schema、数据权限、缓存一致性和运行时用户界面。
- `linactl dao`仍然依赖数据库可连接且 schema 已初始化；本变更只移除`gf`可执行文件安装依赖，不改变数据库前置条件。
