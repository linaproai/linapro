## 1. 公共契约

- [x] 1.1 在 `apps/lina-core/pkg/pluginservice/contract` 新增源码插件可见的 `CacheService` 与 `CacheItem` 契约，TTL 参数统一使用 `time.Duration`
- [x] 1.2 为 `CacheService`、`CacheItem`、值类型常量和方法补齐符合 Go 注释规范的说明，明确有损缓存、插件私有 namespace 和非权威数据源边界
- [x] 1.3 在 `apps/lina-core/pkg/pluginhost.HostServices` 增加 `Cache() contract.CacheService` 方法

## 2. 宿主适配器与作用域绑定

- [x] 2.1 在 `apps/lina-core/internal/service/pluginhostservices` 新增 cache adapter，将 `namespace + key + ttl` 映射到共享 `kvcache.Service`
- [x] 2.2 在 cache adapter 中按绑定的 `pluginID`、当前租户上下文、namespace 和逻辑 key 生成内部 `kvcache` key，并强制使用 `kvcache.OwnerTypePlugin`
- [x] 2.3 在 cache adapter 中映射 `kvcache.Item` 到插件可见 `contract.CacheItem`，并保持 string/int、TTL、未命中和类型错误语义
- [x] 2.4 修改 `pluginhostservices.New`，显式接收启动期共享 `kvcache.Service`，并在依赖缺失时返回 error
- [x] 2.5 实现 plugin-scoped `HostServices` 包装，让 `Cache()` 返回绑定当前插件 ID 的 cache adapter，其余 host services 透传

## 3. 源码插件调用路径接入

- [x] 3.1 修改源码插件 HTTP route 注册路径，向每个插件 registrar 传入绑定该 manifest ID 的 scoped host services
- [x] 3.2 修改源码插件 Cron 注册路径，向每个插件 registrar 传入绑定该 manifest ID 的 scoped host services
- [x] 3.3 修改 managed cron 收集路径，确保收集期间 `HostServices().Cache()` 也绑定当前插件 ID
- [x] 3.4 修改源码插件 hook 分发路径，向 hook payload 传入绑定被调插件 ID 的 scoped host services
- [x] 3.5 修改 HTTP runtime 启动装配，将已创建的共享 `kvCacheSvc` 注入 `pluginhostservices.New`，不得在插件调用路径新增 `kvcache.New()`

## 4. 测试与官方插件编译面

- [x] 4.1 更新所有实现 `pluginhost.HostServices` 的测试替身和官方源码插件测试 fake，补齐 `Cache()` 方法
- [x] 4.2 为 `pluginhostservices` cache adapter 添加单元测试，覆盖 set/get/delete、incr、expire、负 TTL、字符串递增类型错误和 backend 错误透传
- [x] 4.3 添加源码插件作用域隔离测试，覆盖不同 pluginID 使用相同 namespace/key 时不能互相命中
- [x] 4.4 添加租户隔离测试，覆盖同一 pluginID、namespace/key 在不同租户上下文下互不命中
- [x] 4.5 添加启动注入或集成测试，确认源码插件 cache facade 复用传入的共享 `kvCacheSvc`，而不是在调用路径构造新实例

## 5. 验证与治理记录

- [x] 5.1 运行 `cd apps/lina-core && go test ./pkg/pluginhost ./pkg/pluginservice/... ./internal/service/pluginhostservices ./internal/service/plugin -count=1`
  - 2026-05-18: `cd apps/lina-core && go test ./pkg/pluginhost ./pkg/pluginservice/... ./internal/service/pluginhostservices ./internal/service/plugin -count=1` 已执行；`pkg/pluginhost`、`pkg/pluginservice/...`、`internal/service/pluginhostservices` 通过，`internal/service/plugin` 全量测试被当前工作区已有插件依赖治理断言阻断，失败集中为 `Plugin <id> cannot be changed because installed plugins depend on it` 与依赖 blocker 数量断言，不属于本变更新增 cache facade 路径。替代覆盖已通过：`cd apps/lina-core && go test ./internal/service/plugin/internal/integration -count=1`、`cd apps/lina-core && go test ./internal/service/plugin -run '^$' -count=1`。
- [x] 5.2 运行 `cd apps/lina-core && go test ./internal/cmd -count=1`，覆盖 HTTP runtime 构造和 host service 注入编译门禁
  - 2026-05-18: 验证通过。
- [x] 5.3 对受影响官方源码插件运行 Go 编译烟测；若需要临时 workspace，应记录命令和覆盖范围
  - 2026-05-18: `GOWORK=off` 运行 `apps/lina-plugins/monitor-server` 无法解析本地 `lina-core` 模块；随后使用临时 `GOWORK` 文件包含 `apps/lina-core` 与 `apps/lina-plugins/monitor-server`，运行 `cd apps/lina-plugins/monitor-server && GOWORK=<tmp> go test ./backend -count=1`，验证通过，临时文件已清理。
- [x] 5.4 运行 `openspec validate expose-source-plugin-cache-service --strict`
  - 2026-05-18: 验证通过。
- [x] 5.5 运行 `git diff --check`
  - 2026-05-18: 验证通过。
- [x] 5.6 在任务记录中明确 i18n 影响判断：默认无用户可见文案变更；如新增调用端可见错误码则同步维护 error i18n
  - i18n 影响：本轮不新增或修改用户可见 UI 文案、菜单、按钮、表单、manifest i18n 或 apidoc i18n；cache adapter 新增调用端可见 `bizerr.Code` 用于未配置共享 cache service 与缺失 plugin ID 场景，已同步维护 `apps/lina-core/manifest/i18n/en-US/error.json` 与 `apps/lina-core/manifest/i18n/zh-CN/error.json`。
- [x] 5.7 在任务记录中明确缓存一致性判断：单机模式走共享单机后端，集群模式走共享 coordination KV backend，写失败不伪装成功，缓存仍为有损数据
  - 缓存一致性影响：源码插件 cache facade 复用 HTTP 启动期注入的共享 `kvCacheSvc`，生产调用路径未新增 `kvcache.New()`；`cluster.enabled=false` 继续使用共享单机后端，`cluster.enabled=true` 继续使用共享 coordination KV backend。cache key 由宿主按 pluginID、namespace、logical key 和当前租户上下文生成，写入、删除、递增、过期失败均返回错误，不伪装成功；缓存仍为有损数据，不作为权限、配置、插件状态、租户隔离、业务权威数据或关键修订号事实源。
- [x] 5.8 在任务完成后调用 `/lina-review` 进行代码和规范审查
  - 2026-05-18: `/lina-review` 已完成。审查修复了源码插件 cache adapter 的调用端可见错误码治理问题，新增 `bizerr.Code` 与中英文 error i18n；同时修复 `CacheItem.Key` 暴露内部编码 key 的契约边界，改为返回源码插件传入的逻辑 key，并补充单元测试。复核结论：本变更不新增 REST 数据操作接口或前端交互；源码插件 cache facade 复用启动期注入的共享 `kvCacheSvc`，插件 ID 与租户作用域由宿主绑定，写失败不伪装成功。验证命令已重新通过：`cd apps/lina-core && go test ./internal/service/pluginhostservices -count=1`、`cd apps/lina-core && go test ./pkg/pluginhost ./pkg/pluginservice/... ./internal/service/pluginhostservices -count=1`、`cd apps/lina-core && go test ./internal/cmd -count=1`、`cd apps/lina-core && go test ./internal/service/plugin/internal/integration -count=1`、`cd apps/lina-core && go test ./internal/service/plugin -run '^$' -count=1`、`openspec validate expose-source-plugin-cache-service --strict`、`git diff --check`。完整 `cd apps/lina-core && go test ./pkg/pluginhost ./pkg/pluginservice/... ./internal/service/pluginhostservices ./internal/service/plugin -count=1` 仍被当前工作区已有 `internal/service/plugin` 插件依赖治理断言阻断，公共包与 `pluginhostservices` 已通过，阻断点与本变更 cache facade 路径无关。
