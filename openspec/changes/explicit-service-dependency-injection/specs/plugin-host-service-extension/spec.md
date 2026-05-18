## ADDED Requirements

### Requirement: 插件宿主服务适配器必须由宿主运行期统一构造
系统 SHALL 由宿主运行期统一构造并发布源码插件和动态插件 host service 适配器。适配器 MUST 复用启动期共享的宿主服务实例或共享后端，MUST 不在插件调用路径中自行构造孤立宿主服务图。

#### Scenario: 源码插件使用宿主服务适配器
- **WHEN** 源码插件调用 `pkg/pluginservice/*` 发布的宿主能力
- **THEN** 该能力适配器由宿主运行期构造并通过 registrar 传递给插件
- **AND** 适配器复用宿主共享的 auth、session、notify、config、i18n、pluginstate 或其他依赖
- **AND** 插件生产路径不得无参创建该适配器

#### Scenario: 动态插件 host service 调用共享宿主能力
- **WHEN** 动态插件通过统一 host service 协议调用 cache、lock、notify、config、runtime、storage 或 data 等宿主能力
- **THEN** host service handler 使用插件 runtime 注入的共享宿主服务或共享后端
- **AND** handler 不得在每次调用中创建独立 cache、lock、notify、config 或 plugin service 实例

#### Scenario: WASM host service 配置入口由启动期注入
- **WHEN** 宿主启动并初始化 WASM host service
- **THEN** 启动路径显式配置 cache、lock、notify、storage、config 和 runtime 等 host service 的共享依赖
- **AND** 包级默认实例不得在生产启动后继续作为实际运行依赖
