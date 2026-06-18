## Why

当前宿主在自定义 `shutdown.timeout` 中维护优雅停止超时，又把该值手动写入 GoFrame HTTP Server 的优雅停止配置，形成两套配置入口和重复实现。项目已经要求 HTTP 信号处理和优雅关闭复用 GoFrame `Server.Run()`，停机预算也应复用 GoFrame Server 已生效的 `server.gracefulShutdownTimeout`，避免主框架重复解析和维护同一类配置。

## What Changes

- 移除 LinaPro 自定义 `shutdown.timeout` 配置入口和 `config.Service.GetShutdown` 读取契约。
- 宿主拥有的运行时资源清理继续在 GoFrame `Server.Run()` 返回后执行，但清理 deadline 从 GoFrame Server 的 `GetGracefulShutdownTimeout()` 派生。
- 配置样例将停机超时迁移到 `server.gracefulShutdownTimeout`，不再声明顶层 `shutdown` 配置段。
- 不新增 HTTP API、前端页面、数据库 schema、权限点、插件清单或用户可见运行时文案。

## Impact

- 影响代码：`apps/lina-core/internal/cmd/internal/httpstartup` 启动和关闭编排、`internal/service/config` 静态配置契约及相关单元测试。
- 影响配置：`apps/lina-core/manifest/config/config*.yaml` 与 `hack/deploy/*.yaml` 中的停机超时配置样例。
- 影响规范：`host-runtime-operations` 中优雅关闭配置契约改为复用 GoFrame Server 配置。
- 无数据库、数据权限、缓存、前端 UI、E2E、插件清单或 `i18n` 资源影响。
