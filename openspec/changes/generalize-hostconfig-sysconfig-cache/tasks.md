## 1. 规范与设计

- [x] 1.1 编写 OpenSpec proposal、design、tasks 和增量规范，明确源码插件与动态插件的 HostConfig 边界。

## 2. 后端实现

- [x] 2.1 将`config`运行时配置快照从硬编码受管 key 泛化为按租户上下文加载`sys_config`有效行。
- [x] 2.2 调整`GetRaw()`读取顺序，使源码插件可通过`HostConfig()`读取自定义`sys_config`有效 key，同时保留静态配置 fallback 和内置默认语义。
- [x] 2.3 调整`sysconfig`写路径，使创建、更新、导入和删除任意配置记录后按需 bump runtime-config revision。

## 3. 验证

- [x] 3.1 补充单元测试覆盖自定义`sys_config`读取、租户 fallback、revision 刷新和静态配置 fallback。
- [x] 3.2 补充或确认动态插件`hostconfig.get`未授权 key 仍被拒绝，授权 key 可读取自定义`sys_config`。
- [x] 3.3 运行 OpenSpec 严格校验和相关 Go 测试门禁，记录缓存一致性、数据权限、i18n、开发工具和 DI 影响。

## Feedback

- [x] **FB-1**: 自定义`sys_config`只修改 key、不修改 value 时也必须推进 runtime-config revision，避免旧 HostConfig 快照继续返回旧 key。
- [x] **FB-2**: 按反馈将`config_runtime_params.go`、`config_runtime_params_cache.go`和`config_runtime_params_revision.go`合并到`config_raw.go`，并将相关单元测试合并到合适的测试文件。
- [x] **FB-3**: 审查并修正`apps/lina-core/internal/service/config`目录下不符合源码关联命名规范的单元测试文件。

## 验证记录

- OpenSpec：`openspec validate generalize-hostconfig-sysconfig-cache --strict`通过。
- Go 测试：`cd apps/lina-core && go test ./internal/service/config ./internal/service/sysconfig ./internal/service/plugin/internal/wasm ./internal/cmd -count=1`通过。
- 反馈验证：`cd apps/lina-core && go test ./internal/service/sysconfig -count=1`通过，覆盖自定义`sys_config`改 key 后旧 key 失效、新 key 生效的快照刷新场景。
- 文件合并验证：`cd apps/lina-core && go test ./internal/service/config -count=1`通过，覆盖合并后的`config_raw.go`与`config_raw_test.go`。
- 测试文件命名治理：`for f in apps/lina-core/internal/service/config/*_test.go; do base=${f%_test.go}.go; [ -f "$base" ] || echo "$f -> missing ${base##*/}"; done`无输出；`cd apps/lina-core && go test ./internal/service/config -count=1`通过；`openspec validate generalize-hostconfig-sysconfig-cache --strict`通过。
- FB-3 影响分析：仅调整`apps/lina-core/internal/service/config`目录下测试文件组织和 OpenSpec 任务记录，不改变运行时代码、API、SQL、插件协议或缓存实现；治理类反馈不新增运行时测试场景，使用文件存在性检查、包级 Go 测试和 OpenSpec 严格校验闭环。
- FB-3 i18n：无运行时用户可见文案、API 文档源文本、语言包或翻译缓存变更。
- FB-3 缓存一致性：无生产缓存逻辑变更；仅移动既有缓存相关单元测试到源码关联文件。
- FB-3 数据权限：无数据读写边界、租户边界或插件授权边界变更。
- FB-3 开发工具跨平台：无 Makefile、脚本、CI、代码生成入口或跨平台工具变更。
- FB-3 DI 来源检查：无新增运行期依赖、构造函数参数、服务装配或独立服务图。
- 缓存一致性：权威数据源为`sys_config`；快照复用 runtime-config revision、本地`gcache`和`cachecoord`单机/集群分支；缓存 key 增加 tenant scope；创建、更新、导入和删除任意`sys_config`有效值后推进 revision；写入节点立即清理当前 scope 快照，其他节点沿用共享 revision 和 watcher 刷新。
- 数据权限/租户边界：读取按当前 tenant scope 加载平台行或平台+租户行，租户行覆盖平台行；动态插件仍由`hostconfig.get`的`resources.keys`授权拦截；源码插件通过稳定`HostConfig()`读取，不直接访问 DAO/Entity。
- i18n：无运行时用户可见文案、API 文档源文本、语言包或翻译缓存变更。
- 开发工具跨平台：无 Makefile、脚本、CI、代码生成入口或跨平台工具变更。
- DI 来源检查：无新增运行期依赖；继续复用启动期注入的`config.Service`实例、`hostconfigcap.Service`适配器和 WASM runtime 的`hostConfigService`。未新增独立服务图或中间路径`New()`构造关键服务。
