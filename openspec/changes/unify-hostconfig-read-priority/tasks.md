## 1. 配置读取链路

- [x] 1.1 调整`GetRaw(ctx, key)`的非 root key 读取顺序为`sys_config`有效快照、`config.yaml`、系统默认值、`nil`。
- [x] 1.2 移除`GetRaw()`中的具体 key 分支和`IsManagedSysConfigKey()`读取顺序判断，保留空 key 与`.`的完整静态配置快照语义。
- [x] 1.3 建立配置包内部的系统默认值元数据或等价 resolver，聚合已有运行时参数、公开前端设置和静态宿主配置默认值。
- [x] 1.4 调整`sys.log.retentionDays`通用读取语义，使`GetRaw()`不再对该 key 执行缺失即错误的特殊分支。

## 2. 专用 getter 与插件边界

- [x] 2.1 调整`GetJwtExpire()`、`GetSessionTimeout()`、`GetUploadMaxSize()`、`GetCronLogRetention()`等专用 getter，确认来源优先级与`GetRaw()`一致且保留类型校验。
- [x] 2.2 确认源码插件`HostConfig`适配器继续复用宿主配置服务，不新增插件侧 key 白名单。
- [x] 2.3 确认动态插件`hostconfig.get`仍在读取前执行`hostServices.resources.keys`授权校验，授权后使用统一读取优先级。
- [x] 2.4 检查`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`是否需要同步说明 HostConfig 来源优先级；如需修改，保持中英文内容一致。

## 3. 测试与治理验证

- [x] 3.1 补充`internal/service/config`单元测试，覆盖`sys_config`优先于静态配置、静态配置优先于系统默认值、系统默认值 fallback、全部缺失返回`nil`。
- [x] 3.2 补充默认值元数据测试或静态检索，确认新增默认值不需要修改`GetRaw()`读取分支。
- [x] 3.3 补充`sys_config`freshness 错误测试，确认错误不会被静态配置或系统默认值 fallback 掩盖。
- [x] 3.4 补充或更新动态 WASM HostConfig 测试，覆盖未授权 key 拒绝和授权 key 使用统一读取优先级。
- [x] 3.5 运行`cd apps/lina-core && go test ./internal/service/config ./internal/service/plugin/internal/hostconfig ./internal/service/plugin/internal/wasm -count=1`。
- [x] 3.6 运行`openspec validate unify-hostconfig-read-priority --strict`，并记录`i18n`、缓存一致性、数据权限、开发工具跨平台和测试策略影响判断。

## 执行记录

- DI 来源：未新增运行期依赖、构造函数或启动装配；`HostConfig`源码插件适配器继续复用注入的宿主配置服务，动态 WASM host service 继续使用既有 runtime 注入的`hostConfigService`。
- 缓存一致性：继续复用`runtime-config`共享 revision、本地`gcache`快照、租户作用域 cache key 和 freshness 可见错误策略；未新增缓存层，已补充 freshness 错误不 fallback 的单元测试。
- 数据权限与租户边界：未新增 HTTP API、列表、详情、写入或下载接口；`sys_config`读取仍使用当前上下文可见快照，动态`hostconfig.get`仍先校验`hostServices.resources.keys`授权。
- `i18n`影响：未新增或修改运行时用户可见文案、API 文档源文本、菜单、语言包或翻译资源；仅同步插件公开契约 README 对 HostConfig 来源优先级的说明。
- 开发工具跨平台：未修改`Makefile`、脚本、CI、构建工具或跨平台执行入口，无跨平台工具影响。
- 测试策略：已运行`cd apps/lina-core && go test ./internal/service/config ./internal/service/plugin/internal/hostconfig ./internal/service/plugin/internal/wasm -count=1`、`openspec validate unify-hostconfig-read-priority --strict`和`git diff --check`。

## Feedback

- [x] **FB-1**: 移除运行时参数校验和快照解析中的配置键硬编码分支，确保通用配置读取链路通过元数据调度而不是框架层 key 判断。

### FB-1 执行记录

- 根因：前一轮仅移除了`GetRaw()`读取路径中的具体 key 分支，但`validateRuntimeParamValue()`和`loadRuntimeParamSnapshot()`仍通过`RuntimeParamKey*`集中分支调度校验与 typed snapshot 解析，新增或修改运行时配置键仍需要改通用配置读取层。
- 实现：将受保护配置校验器和运行时快照解析器挂载到配置元数据；`validateRuntimeParamValue()`、`ValidateProtectedConfigValue()`、`validatePublicFrontendSettingValue()`和`loadRuntimeParamSnapshot()`只做元数据查找与函数调度，不再在框架层按具体 key 分支。
- 影响分析：未新增运行期依赖、构造函数、启动装配、HTTP API、DTO、SQL、开发工具脚本或前端页面；动态`hostconfig.get`授权边界不变。
- 缓存一致性：继续复用`runtime-config`共享 revision、本地`gcache`快照和租户作用域 cache key；仅调整快照内 typed value 的元数据解析入口，无新增缓存层或失效策略。
- 数据权限与租户边界：未新增数据操作接口；`sys_config`读取仍按当前上下文可见快照和租户行覆盖平台行规则执行，动态插件仍先通过 manifest 授权快照校验 key。
- `i18n`影响：未新增或修改运行时用户可见文案、API 文档源文本、菜单、语言包、翻译资源或插件清单。
- 开发工具跨平台：未修改`Makefile`、脚本、CI、构建工具或跨平台执行入口，无跨平台工具影响。
- 测试策略：已运行`cd apps/lina-core && go test ./internal/service/config ./internal/service/plugin/internal/hostconfig ./internal/service/plugin/internal/wasm -count=1`、`openspec validate unify-hostconfig-read-priority --strict`、`git diff --check`，并通过静态检索确认`config_raw.go`和相关测试中不存在`case RuntimeParamKey`或按 key switch。
