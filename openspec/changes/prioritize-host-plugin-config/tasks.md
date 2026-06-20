## 1. 配置源解析

- [x] 1.1 在`plugincap`配置工厂中引入宿主静态配置段读取能力，并保持与`internal/service/config`解耦。
- [x] 1.2 实现`plugin.<plugin-id>`配置段级最高优先级解析，配置段不存在时回退到生产文件、开发文件和 artifact 默认配置。
- [x] 1.3 保留插件配置 key 校验，确保空 key、`.`、非法前后点号和跨插件配置读取不会返回宿主配置快照。

## 2. 启动装配和能力一致性

- [x] 2.1 调整 HTTP 启动装配，创建带宿主静态配置 reader 的共享`ConfigServiceFactory`。
- [x] 2.2 调整`NewHostServices()`和`capabilityhost.New()`，让源码插件能力目录接收启动期共享配置工厂。
- [x] 2.3 确认动态插件 WASM host service 继续使用同一个配置工厂，并保留 artifact 默认配置按执行上下文绑定。
- [x] 2.4 记录 DI 来源检查：配置 reader 的 owner、创建位置、传递路径、源码插件和动态插件是否复用同一实例。

## 3. 文档和规范同步

- [x] 3.1 更新`apps/lina-core/pkg/plugin/README.md`说明插件配置来源优先级和`HostConfig()`边界。
- [x] 3.2 更新`apps/lina-core/pkg/plugin/README.zh-CN.md`，确保中英文镜像文档事实一致。
- [x] 3.3 记录影响分析：`i18n`无运行时资源影响，缓存一致性无新增热更新或跨实例失效影响，数据权限、数据库、前端 UI 和开发工具跨平台无影响。

## 4. 测试和验证

- [x] 4.1 增加`plugincap`单元测试，覆盖主静态配置优先、配置段级优先不逐 key 混合、静态配置缺失回退生产文件、生产文件缺失回退开发文件、文件缺失回退 artifact。
- [x] 4.2 增加源码插件能力目录或装配测试，验证源码插件使用启动期共享配置工厂。
- [x] 4.3 增加动态`plugins.config.get`测试，验证动态插件读取主静态配置优先且 artifact 默认配置仍按执行上下文绑定。
- [x] 4.4 运行`cd apps/lina-core && go test ./pkg/plugin/capability/plugincap -count=1`。
- [x] 4.5 运行覆盖启动装配和 WASM host service 的 Go 测试或编译烟测，至少包含`cd apps/lina-core && go test ./internal/cmd -count=1`。
- [x] 4.6 运行`openspec validate prioritize-host-plugin-config --strict`。

## 实施记录

- DI 来源检查：宿主静态配置 reader 的 owner 是`internal/service/config.Service`实现的`GetRaw(ctx, key)`能力；创建位置是 HTTP 启动装配`apps/lina-core/internal/cmd/internal/httpstartup/http_runtime.go`，以`pluginserviceconfig.NewConfigFactoryWithHostStaticConfig("", "", hostConfigReader)`创建启动期共享`ConfigServiceFactory`；传递路径为`newHTTPRuntime`同时传给`pluginsvc.NewHostServices()`和`pluginsvc.New()`，再进入`capabilityhost.New()`、源码插件`Plugins().Config()`和动态 WASM runtime 的`plugins.config.get`。源码插件和动态插件复用同一启动期工厂实例；动态 artifact 默认配置仅通过执行上下文的`WithArtifactConfig()`派生工厂视图，不替换共享基础实例。
- 影响分析：本次变更不新增运行时用户可见文案、菜单、路由、API 文档元数据或翻译资源，`i18n`无运行时资源影响；不新增缓存、热更新、跨实例失效、数据库写路径或配置中心写路径，缓存一致性无新增失效和跨实例同步影响；不修改 HTTP API、数据库、SQL、前端 UI、E2E 资产、开发工具脚本或跨平台执行入口；插件配置读取不访问业务数据，数据权限无影响。
- 验证记录：已运行`cd apps/lina-core && go test ./pkg/plugin/capability/plugincap -count=1`、`cd apps/lina-core && go test ./internal/service/plugin/internal/capabilityhost -count=1`、`cd apps/lina-core && go test ./internal/service/plugin/internal/wasm -count=1`、`cd apps/lina-core && go test ./internal/service/plugin -count=1`、`cd apps/lina-core && go test ./internal/cmd -count=1`、`cd apps/lina-core && go test ./internal/cmd/internal/httpstartup -count=1`、`cd apps/lina-core && go test ./internal/service/user -count=1`和`openspec validate prioritize-host-plugin-config --strict`，均通过。
