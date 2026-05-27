## ADDED Requirements

### Requirement:破坏性公共契约收敛不得保留兼容 facade

系统 SHALL 在本迭代中直接删除插件公共包中的历史宽契约、误导性 facade 和兼容转发层。系统 MUST NOT 为已删除的`HostServices`、`HostServicesForPlugin`、`HostServices()`、`contract.ProviderEnv.Services`或`pluginbridge`业务能力 client 保留生产兼容入口。

#### Scenario:旧插件能力入口被生产代码引用

- **WHEN** 生产代码继续引用旧`pluginhost.HostServices`、`HostServicesForPlugin`、`HostServices()`、`contract.ProviderEnv.Services`或`pluginbridge.Runtime()`等业务能力入口
- **THEN** 编译或治理扫描必须失败
- **AND** 调用方必须迁移到`capability.Services`、`pluginhost.Services`的`Services()`访问器、强类型 provider env 或`pkg/plugin/capability/guest`

#### Scenario:插件确实需要新增读能力

- **WHEN** 删除旧宽接口后发现插件仍需要组织、租户或宿主能力读数据
- **THEN** 系统只能新增 DTO 化、批量化、只读能力
- **AND** 不得恢复`*gdb.Model`、`*ghttp.Request`、DAO、DO、Entity、写入接口、数据范围注入或宿主内部治理接口

### Requirement:插件公共契约和宿主内部插件运行时必须分离

系统 SHALL 将`apps/lina-core/pkg/plugin`作为插件公共契约命名空间，将`apps/lina-core/internal/service/plugin`作为宿主内部插件运行时治理实现命名空间。公共契约不得导入宿主内部插件运行时实现，插件代码不得直接调用宿主内部插件运行时包。

#### Scenario:公共插件包不能依赖宿主内部运行时

- **WHEN** 开发者在`apps/lina-core/pkg/plugin/**`下实现或修改公共契约
- **THEN** 代码不得导入`lina-core/internal/service/plugin/**`
- **AND** 公共契约只能通过 DTO、接口、协议 envelope 或 provider-facing 契约表达数据边界

#### Scenario:宿主内部运行时使用公共契约

- **WHEN** `apps/lina-core/internal/service/plugin/**`需要执行插件 catalog、runtime、lifecycle、WASM host service 或管理端投影逻辑
- **THEN** 它可以依赖`apps/lina-core/pkg/plugin/**`中的稳定公共契约
- **AND** 不得要求公共包反向暴露宿主内部缓存快照、私有配置、DAO、DO、Entity 或 runtime 状态结构

#### Scenario:插件代码不能导入宿主内部实现

- **WHEN** 源码插件或动态插件业务代码访问宿主能力
- **THEN** 它必须通过`pkg/plugin/capability`、`pkg/plugin/pluginhost`、`pkg/plugin/pluginbridge`或受治理`hostServices`协议完成
- **AND** 不得导入`lina-core/internal/service/**`、`lina-core/internal/dao/**`或`lina-core/internal/model/**`

### Requirement:宿主内部组织和租户治理能力必须通过窄接口注入

系统 SHALL 将组织、租户、数据范围、成员关系、自动开通和启动一致性等宿主内部治理能力拆分为职责明确的窄接口，并通过构造函数显式注入到需要的宿主 service。

#### Scenario:宿主数据范围服务注入组织范围接口

- **WHEN** 宿主`datascope`或其他核心 service 需要按部门或组织关系在数据库查询阶段过滤数据
- **THEN** 该 service 依赖组织范围治理窄接口
- **AND** 不依赖普通插件消费用的`orgcap.Service`宽接口

#### Scenario:宿主用户和角色服务注入租户成员接口

- **WHEN** 宿主用户、角色或通知 service 需要校验、投影或更新用户租户成员关系
- **THEN** 该 service 依赖租户成员治理窄接口
- **AND** 不通过普通插件消费目录获取写入或底层查询注入能力

#### Scenario:测试替身只实现被测窄接口

- **WHEN** 单元测试为组织或租户能力构造替身
- **THEN** 替身只需要实现当前被测 service 实际依赖的窄接口
- **AND** 不得为了满足过宽接口而实现大量无关空方法
