## ADDED Requirements

### Requirement: 领域能力边界分叉必须被治理验证阻断

系统 SHALL 提供静态检索、Go 治理测试或等价验证，阻断领域能力契约、宿主实现、动态`WASM host service`配置和动态 guest 代理再次分叉。治理验证 MUST 区分普通领域能力与资源型或 transport 型 host service client，避免误伤`runtime`、`storage`、`network`、`data`、`cache`、`lock`、`notify`、`config`、`hostconfig`和`manifest`等非普通领域 client。

#### Scenario: 发现领域专用 WASM 配置入口

- **WHEN** 生产 Go 代码新增`ConfigureAITextHostService`、`ConfigureUserHostService`、`ConfigureOrgHostService`、`ConfigureTenantHostService`或同类领域专用`Configure*HostService`入口
- **THEN** 治理验证失败
- **AND** 变更必须改为复用`ConfigureDomainHostServices`

#### Scenario: 发现动态领域专用全局目录

- **WHEN** `WASM`普通领域分发代码新增独立的`capability.Services`包级变量作为某个领域的 fallback 目录
- **THEN** 治理验证失败
- **AND** 该领域必须通过共享领域能力目录解析

#### Scenario: 发现 guest 公共包定义平行领域接口

- **WHEN** `pkg/plugin/pluginbridge`公共包新增与`capability/*cap`平行的普通领域接口
- **THEN** 治理验证失败
- **AND** 接口必须迁移到对应`*cap`契约或作为`domainhostcall`内部实现细节

#### Scenario: 启动层导入内部实现组件

- **WHEN** `apps/lina-core/internal/cmd`生产 Go 代码导入`lina-core/internal/service/plugin/internal/capabilityhost`
- **THEN** 治理验证失败
- **AND** 启动层必须通过`internal/service/plugin`根 facade 获取能力构造入口

### Requirement: 宿主配置管理能力必须归属 HostConfig 管理面

系统 SHALL 将宿主配置相关公开能力收敛到`hostconfigcap`组件。普通插件通过`Services.HostConfig()`读取只读宿主配置；可信源码插件通过`AdminServices.HostConfig()`访问受治理运行时配置管理命令。系统 MUST NOT 继续公开独立`capability/configcap`组件包或根`AdminServices.Config()`管理入口。

#### Scenario: 源码插件管理运行时配置投影

- **WHEN** 可信源码插件需要读取或写入宿主拥有的受治理运行时配置投影
- **THEN** 插件必须通过`pluginhost.Services.Admin().HostConfig()`获取管理能力
- **AND** 管理接口必须归属`pkg/plugin/capability/hostconfigcap.AdminService`
- **AND** 不得通过独立`capability/configcap`组件包或`Admin().Config()`访问

### Requirement: 领域能力边界文档必须随主框架插件能力变更同步

系统 SHALL 在主框架插件能力边界变更时同步审查`apps/lina-core/pkg/plugin`目录下的 README 文档。文档 MUST 明确`capability`、`pluginhost`、`pluginbridge`、`pluginbridge/protocol`和宿主内部`capabilityhost`的职责边界。

#### Scenario: 插件能力边界实现迁移

- **WHEN** 变更迁移宿主领域能力实现目录、动态领域分发入口或 guest 领域代理位置
- **THEN** 任务必须检查`apps/lina-core/pkg/plugin/README.md`和`README.zh-CN.md`是否需要同步
- **AND** 若需要更新，文档必须说明协议目录不是领域契约 owner
