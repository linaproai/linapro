## MODIFIED Requirements

### Requirement: core 插件公共命名空间必须集中到 pkg/plugin

系统 SHALL 将`lina-core`拥有的插件相关公共 Go 组件集中到`apps/lina-core/pkg/plugin/`命名空间下。该命名空间下的公开顶层组件必须按职责拆分为源码插件贡献入口、动态插件桥接协议和 core-owned 插件消费宿主能力目录，不得继续在`apps/lina-core/pkg/`根层新增语义模糊的插件公共组件。plugin-owned 非核心领域能力的公开契约不属于 core 公共命名空间，MUST 位于 owner 插件`apps/lina-plugins/<plugin-id>/backend/cap/<domain>cap`，并受跨插件 import 边界治理。

#### Scenario: 开发者定位 core 插件公共入口

- **WHEN** 开发者需要查找 LinaPro 插件内核、源码插件贡献 API、动态桥接协议或 core-owned 宿主能力契约
- **THEN** 系统在`apps/lina-core/pkg/plugin/`下提供对应公共组件
- **AND** 不要求开发者在`pkg/pluginservice`、`pkg/pluginhost`、`pkg/pluginbridge`和其他顶层插件包之间猜测职责边界

#### Scenario: 开发者定位 plugin-owned 领域契约

- **WHEN** 开发者需要查找`linapro-ai-core`拥有的`AI`领域公开契约
- **THEN** 系统在`apps/lina-plugins/linapro-ai-core/backend/cap/aicap`下提供对应契约和 SDK
- **AND** 开发者不得在`lina-core/pkg/plugin/capability/aicap`继续寻找生产 owner 契约

#### Scenario: 新增 core 插件公共组件

- **WHEN** 系统需要新增插件开发者或宿主插件运行时共享的 core-owned 公共契约
- **THEN** 该契约必须优先放入`apps/lina-core/pkg/plugin/`下职责明确的公开子包
- **AND** 不得新增`pluginservice`、`plugincommon`、`pluginutil`等语义模糊的顶层公共包

#### Scenario: 新增 plugin-owned 领域公共契约

- **WHEN** 系统需要新增非核心领域 owner 插件公开契约
- **THEN** 该契约必须放入 owner 插件`backend/cap/<domain>cap`
- **AND** 不得为了方便跨插件 import 而放入 owner 插件`backend/internal`、`backend/pkg`或 core `pkg/plugin/capability`

### Requirement: `pluginhost`、`pluginbridge`、`capability`和 owner cap 必须职责分离

系统 SHALL 在 core `pkg/plugin`命名空间下保持三类核心公共组件职责分离：`pluginhost`只负责源码插件贡献 API、源码插件获取统一服务目录和通用 capability descriptor 接收；`pluginbridge`只负责动态插件 ABI、WASM transport、公开协议出口、core-owned 动态插件公共入口和 owner-aware 通用 host service envelope；`capability`只负责 core-owned 插件消费宿主能力的稳定目录、公共原语和`*cap`能力契约。plugin-owned 非核心领域能力的公开契约、动态 guest SDK 和 provider SPI MUST 归属 owner 插件`backend/cap/<domain>cap`及其子包。仅服务动态插件的 core guest SDK（如 record store）MUST 归属`pluginbridge`，不得放入 owner 插件 cap；仅服务 owner 领域的 guest SDK MUST 归属 owner 插件 cap。

#### Scenario: 源码插件注册贡献

- **WHEN** 源码插件需要注册路由、hook、cron、生命周期回调或 core-owned provider factory
- **THEN** 插件使用`pkg/plugin/pluginhost`
- **AND** `pluginhost`不得拥有宿主能力业务实现
- **AND** `pluginhost`不得暴露`Admin()`或`AdminServices`管理目录
- **AND** `pluginhost.Services`顶层不得成为新增治理能力的事实 owner，治理能力必须委托到对应领域`Service`

#### Scenario: 源码插件注册 plugin-owned provider

- **WHEN** 源码插件需要声明`AI`或其他 plugin-owned 非核心领域 provider
- **THEN** 插件使用 owner 插件`backend/cap/<domain>cap/spi`或等价 helper 构造通用 descriptor
- **AND** `pluginhost`只接收通用 descriptor
- **AND** `pluginhost`不得 import owner 插件 cap 包来定义领域专属 facade

#### Scenario: 动态插件执行 bridge 调用

- **WHEN** 动态插件需要声明 route、处理 WASM request envelope 或调用 host call transport
- **THEN** 插件使用`pkg/plugin/pluginbridge/guest`和`pkg/plugin/pluginbridge/protocol`
- **AND** `pluginbridge`不得定义业务能力可用性、provider 激活、数据权限降级或配置读取语义
- **AND** `pluginbridge`根包不得重新导出协议 DTO、常量、codec、guest helper 或`Runtime()`、`Data()`、`RecordStore()`、`Cron()`等能力 client facade

#### Scenario: 动态插件消费 plugin-owned 能力

- **WHEN** 动态插件需要调用 owner 插件发布的`AI`能力
- **THEN** 插件使用 owner 插件`backend/cap/aicap/bridge`或等价公开 guest SDK
- **AND** SDK 只负责编码、声明 helper 和调用通用 host call，不得绕过`hostServices`授权、owner 依赖和宿主审计

#### Scenario: 插件消费 core-owned 宿主能力

- **WHEN** 源码插件或动态插件需要访问配置、manifest、缓存、通知、组织、租户或业务上下文等 core-owned 宿主能力
- **THEN** 插件使用`pkg/plugin/capability`、对应`pkg/plugin/capability/<domain>cap`能力组件或 core guest client
- **AND** 能力目录不得被命名为`pluginservice`或动态`hostServices`
- **AND** 能力 guest client 方法需要的 bridge DTO、常量和 codec 必须直接使用`pkg/plugin/pluginbridge/protocol`，不得在`capability/guest`重复定义公开别名

#### Scenario: 动态插件访问受治理 record store 能力

- **WHEN** 动态插件需要使用 ORM-style record store facade、typed record store plan 或宿主 data governance 适配入口
- **THEN** 插件使用`pkg/plugin/pluginbridge/recordstore`
- **AND** Go guest 能力目录通过`RecordStore()`返回该 facade
- **AND** 不得继续通过`pkg/plugin/capability/recordstore`、owner 插件 cap 或顶层`pkg/plugindb`暴露该能力

## ADDED Requirements

### Requirement: 跨插件 import 边界必须由治理扫描覆盖

系统 SHALL 提供静态治理扫描或等价验证，确保跨插件生产 Go import 只允许依赖目标插件的`backend/cap/...`公开契约。扫描 MUST 阻断跨插件 import `backend/internal/...`、`backend/internal/dao`、`backend/internal/model`、`backend/api`、controller、service 实现、私有 provider adapter 和`backend/pkg`领域能力入口。测试 fixture、代码生成输入和受控开发工具例外 MUST 在扫描规则中显式限定目录和用途。

#### Scenario: 生产代码 import owner internal

- **WHEN** 插件生产代码 import `lina-plugin-linapro-ai-core/backend/internal/service/ai`
- **THEN** 治理扫描 MUST 失败
- **AND** 调用方必须改为依赖`lina-plugin-linapro-ai-core/backend/cap/aicap/...`

#### Scenario: 测试例外受限

- **WHEN** 同插件内部测试需要访问`backend/internal`实现
- **THEN** 该 import MAY 被允许
- **AND** 例外不得允许其他插件生产代码跨插件访问 internal 实现
